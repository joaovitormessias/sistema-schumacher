package routes

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrRouteHasLinkedTrips = errors.New("route has linked non-cancelled trips")
	ErrUnsupported         = errors.New("operation not supported in production ticketing schema")
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

const routeSelectQuery = `select
    r.route_id,
    r.route_name,
    coalesce(st.first_city, '') as origin_city,
    coalesce(st.last_city, '') as destination_city,
    r.is_active,
    null::text as duplicated_from_route_id,
    r.created_at,
    coalesce(st.stop_count, 0)::int as stop_count,
    1::int as min_stop_order,
    coalesce(st.stop_count, 0)::int as max_stop_order,
    coalesce(st.first_city, '') as first_city,
    coalesce(st.last_city, '') as last_city,
    false as has_negative_eta,
    false as has_eta_backtrack,
    exists(select 1 from trips t where t.route_id = r.route_id) as has_linked_trips
  from routes r
  left join lateral (
    with latest_trip as (
      select t.trip_id
      from trips t
      where t.route_id = r.route_id
      order by t.trip_date desc, t.created_at desc
      limit 1
    ),
    ordered as (
      select ts.stop_sequence, s.display_name
      from trip_stops ts
      join stops s on s.stop_id = ts.stop_id
      join latest_trip lt on lt.trip_id = ts.trip_id
      where ts.is_active = true
      order by ts.stop_sequence asc
    )
    select
      count(*)::int as stop_count,
      (select display_name from ordered order by stop_sequence asc limit 1) as first_city,
      (select display_name from ordered order by stop_sequence desc limit 1) as last_city
    from ordered
  ) st on true`

func (r *Repository) List(ctx context.Context, filter ListFilter) ([]Route, error) {
	limit := filter.Limit
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	query := routeSelectQuery
	args := []interface{}{}
	clauses := []string{}

	if filter.Search != "" {
		args = append(args, "%"+filter.Search+"%")
		clauses = append(clauses, fmt.Sprintf(`(
      r.route_name ilike $%d
      or r.route_id ilike $%d
      or coalesce(st.first_city, '') ilike $%d
      or coalesce(st.last_city, '') ilike $%d
    )`, len(args), len(args), len(args), len(args)))
	}

	if filter.Status == "active" {
		clauses = append(clauses, "r.is_active = true")
	}
	if filter.Status == "inactive" {
		clauses = append(clauses, "r.is_active = false")
	}

	if len(clauses) > 0 {
		query += " where " + strings.Join(clauses, " and ")
	}

	query += " order by r.created_at desc"
	args = append(args, limit)
	query += fmt.Sprintf(" limit $%d", len(args))
	args = append(args, offset)
	query += fmt.Sprintf(" offset $%d", len(args))

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []Route{}
	for rows.Next() {
		item, snapshot, err := scanRouteWithSnapshot(rows)
		if err != nil {
			return nil, err
		}
		applyDerivedRouteState(&item, snapshot)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) Get(ctx context.Context, id string) (Route, error) {
	query := routeSelectQuery + " where r.route_id=$1"
	row := r.pool.QueryRow(ctx, query, id)
	item, snapshot, err := scanRouteWithSnapshot(row)
	if err != nil {
		return Route{}, err
	}
	applyDerivedRouteState(&item, snapshot)
	return item, nil
}

func (r *Repository) Create(ctx context.Context, input CreateRouteInput) (Route, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Route{}, err
	}
	defer tx.Rollback(ctx)

	routeName := strings.TrimSpace(input.Name)
	originCity := strings.TrimSpace(input.OriginCity)
	destinationCity := strings.TrimSpace(input.DestinationCity)
	if routeName == "" || originCity == "" || destinationCity == "" {
		return Route{}, errors.New("name, origin_city and destination_city are required")
	}

	routeID, err := r.allocateRouteIDTx(ctx, tx, originCity, destinationCity)
	if err != nil {
		return Route{}, err
	}

	isActive := false
	if input.IsActive != nil {
		isActive = *input.IsActive
	}

	if _, err := tx.Exec(ctx, `
    insert into routes(route_id, route_name, is_active, created_at)
    values ($1, $2, $3, now())
  `, routeID, routeName, isActive); err != nil {
		return Route{}, err
	}

	templateTripID, err := r.allocateTripIDTx(ctx, tx, routeID)
	if err != nil {
		return Route{}, err
	}

	if _, err := tx.Exec(ctx, `
    insert into trips(
      trip_id,
      route_id,
      trip_date,
      bus_id,
      default_price,
      seats_total,
      seats_available,
      duration_hours,
      status,
      package_name,
      created_at
    ) values ($1, $2, date '1900-01-01', null, 0, 0, 0, null, 'TEMPLATE', null, now())
  `, templateTripID, routeID); err != nil {
		return Route{}, err
	}

	originStopID, _, err := r.findOrCreateStopTx(ctx, tx, originCity, input.OriginLatitude, input.OriginLongitude)
	if err != nil {
		return Route{}, err
	}
	destinationStopID, _, err := r.findOrCreateStopTx(ctx, tx, destinationCity, input.DestinationLatitude, input.DestinationLongitude)
	if err != nil {
		return Route{}, err
	}
	if originStopID == destinationStopID {
		return Route{}, errors.New("origin and destination must be different stops")
	}

	if err := r.insertTripStopTx(ctx, tx, templateTripID, originStopID, 1, nil); err != nil {
		return Route{}, err
	}
	if err := r.insertTripStopTx(ctx, tx, templateTripID, destinationStopID, 2, nil); err != nil {
		return Route{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return Route{}, err
	}

	return r.Get(ctx, routeID)
}

func (r *Repository) Update(ctx context.Context, id string, input UpdateRouteInput) (Route, error) {
	sets := make([]string, 0, 2)
	args := make([]interface{}, 0, 3)

	if input.Name != nil {
		args = append(args, strings.TrimSpace(*input.Name))
		sets = append(sets, fmt.Sprintf("route_name = $%d", len(args)))
	}
	if input.IsActive != nil {
		args = append(args, *input.IsActive)
		sets = append(sets, fmt.Sprintf("is_active = $%d", len(args)))
	}

	if len(sets) == 0 {
		return r.Get(ctx, id)
	}

	args = append(args, id)
	commandTag, err := r.pool.Exec(ctx, fmt.Sprintf("update routes set %s where route_id = $%d", strings.Join(sets, ", "), len(args)), args...)
	if err != nil {
		return Route{}, err
	}
	if commandTag.RowsAffected() == 0 {
		return Route{}, pgx.ErrNoRows
	}

	return r.Get(ctx, id)
}

func (r *Repository) Publish(ctx context.Context, id string) (Route, error) {
	commandTag, err := r.pool.Exec(ctx, `update routes set is_active = true where route_id = $1`, id)
	if err != nil {
		return Route{}, err
	}
	if commandTag.RowsAffected() == 0 {
		return Route{}, pgx.ErrNoRows
	}
	return r.Get(ctx, id)
}

func (r *Repository) Duplicate(ctx context.Context, id string) (Route, error) {
	return Route{}, ErrUnsupported
}

func (r *Repository) HasLinkedTrips(ctx context.Context, routeID string) (bool, error) {
	var hasLinkedTrips bool
	if err := r.pool.QueryRow(ctx, `select exists(select 1 from trips where route_id = $1)`, routeID).Scan(&hasLinkedTrips); err != nil {
		return false, err
	}
	return hasLinkedTrips, nil
}

func (r *Repository) ListStops(ctx context.Context, routeID string) ([]RouteStop, error) {
	rows, err := r.pool.Query(ctx, `
    with latest_trip as (
      select t.trip_id
      from trips t
      where t.route_id = $1
      order by t.trip_date desc, t.created_at desc
      limit 1
    )
    select
      ts.trip_stop_id as id,
      $1 as route_id,
      s.display_name as city,
      nullif(to_jsonb(s)->>'latitude', '')::double precision as latitude,
      nullif(to_jsonb(s)->>'longitude', '')::double precision as longitude,
      ts.stop_sequence as stop_order,
      null::int as eta_offset_minutes,
      null::text as notes,
      ts.created_at
    from trip_stops ts
    join stops s on s.stop_id = ts.stop_id
    join latest_trip lt on lt.trip_id = ts.trip_id
    where ts.is_active = true
    order by ts.stop_sequence asc`, routeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []RouteStop{}
	for rows.Next() {
		var item RouteStop
		if err := rows.Scan(
			&item.ID,
			&item.RouteID,
			&item.City,
			&item.Latitude,
			&item.Longitude,
			&item.StopOrder,
			&item.ETAOffsetMinutes,
			&item.Notes,
			&item.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) CreateStop(ctx context.Context, routeID string, input CreateRouteStopInput) (RouteStop, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return RouteStop{}, err
	}
	defer tx.Rollback(ctx)

	if err := r.ensureRouteExistsTx(ctx, tx, routeID); err != nil {
		return RouteStop{}, err
	}

	tripIDs, err := r.listRouteTripIDsTx(ctx, tx, routeID)
	if err != nil {
		return RouteStop{}, err
	}
	if len(tripIDs) == 0 {
		return RouteStop{}, errors.New("route has no trips to propagate stops")
	}

	stopID, _, err := r.findOrCreateStopTx(ctx, tx, input.City, input.Latitude, input.Longitude)
	if err != nil {
		return RouteStop{}, err
	}

	var alreadyExists bool
	if err := tx.QueryRow(ctx, `
    select exists(
      select 1
      from trip_stops
      where trip_id = $1
        and stop_id = $2
        and is_active = true
    )
  `, tripIDs[0], stopID).Scan(&alreadyExists); err != nil {
		return RouteStop{}, err
	}
	if alreadyExists {
		return RouteStop{}, errors.New("stop already exists in route")
	}

	order, err := r.normalizeStopOrderTx(ctx, tx, tripIDs[0], input.StopOrder, true)
	if err != nil {
		return RouteStop{}, err
	}

	for _, tripID := range tripIDs {
		if _, err := tx.Exec(ctx, `
      update trip_stops
      set stop_sequence = stop_sequence + 1
      where trip_id = $1
        and stop_sequence >= $2
    `, tripID, order); err != nil {
			return RouteStop{}, err
		}
		if err := r.insertTripStopTx(ctx, tx, tripID, stopID, order, nil); err != nil {
			return RouteStop{}, err
		}
	}

	if err := r.refreshAvailableSegmentsForRouteTx(ctx, tx, routeID); err != nil {
		return RouteStop{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return RouteStop{}, err
	}

	return r.getRouteStopByOrder(ctx, routeID, order)
}

func (r *Repository) UpdateStop(ctx context.Context, routeID string, stopID string, input UpdateRouteStopInput) (RouteStop, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return RouteStop{}, err
	}
	defer tx.Rollback(ctx)

	if err := r.ensureRouteExistsTx(ctx, tx, routeID); err != nil {
		return RouteStop{}, err
	}

	tripIDs, err := r.listRouteTripIDsTx(ctx, tx, routeID)
	if err != nil {
		return RouteStop{}, err
	}
	if len(tripIDs) == 0 {
		return RouteStop{}, errors.New("route has no trips to propagate stops")
	}

	ref, err := r.getRouteStopReferenceTx(ctx, tx, routeID, stopID)
	if err != nil {
		return RouteStop{}, err
	}

	targetOrder, err := r.normalizeStopOrderTx(ctx, tx, tripIDs[0], valueOrDefaultInt(input.StopOrder, ref.StopOrder), false)
	if err != nil {
		return RouteStop{}, err
	}

	targetStopID := ref.StopID
	if input.City != nil && strings.TrimSpace(*input.City) != "" {
		targetStopID, _, err = r.findOrCreateStopTx(ctx, tx, *input.City, input.Latitude, input.Longitude)
		if err != nil {
			return RouteStop{}, err
		}
	} else if input.Latitude != nil && input.Longitude != nil {
		if err := r.updateStopCoordinatesTx(ctx, tx, targetStopID, input.Latitude, input.Longitude); err != nil {
			return RouteStop{}, err
		}
	}

	if targetStopID != ref.StopID {
		var existsConflict bool
		if err := tx.QueryRow(ctx, `
      select exists(
        select 1
        from trip_stops
        where trip_id = $1
          and stop_id = $2
          and stop_sequence <> $3
          and is_active = true
      )
    `, tripIDs[0], targetStopID, ref.StopOrder).Scan(&existsConflict); err != nil {
			return RouteStop{}, err
		}
		if existsConflict {
			return RouteStop{}, errors.New("target stop already exists in route")
		}
	}

	for _, tripID := range tripIDs {
		if targetOrder != ref.StopOrder {
			if _, err := tx.Exec(ctx, `
        update trip_stops
        set stop_sequence = 0
        where trip_id = $1
          and stop_sequence = $2
      `, tripID, ref.StopOrder); err != nil {
				return RouteStop{}, err
			}

			if targetOrder > ref.StopOrder {
				if _, err := tx.Exec(ctx, `
          update trip_stops
          set stop_sequence = stop_sequence - 1
          where trip_id = $1
            and stop_sequence > $2
            and stop_sequence <= $3
        `, tripID, ref.StopOrder, targetOrder); err != nil {
					return RouteStop{}, err
				}
			} else {
				if _, err := tx.Exec(ctx, `
          update trip_stops
          set stop_sequence = stop_sequence + 1
          where trip_id = $1
            and stop_sequence >= $2
            and stop_sequence < $3
        `, tripID, targetOrder, ref.StopOrder); err != nil {
					return RouteStop{}, err
				}
			}

			if _, err := tx.Exec(ctx, `
        update trip_stops
        set stop_sequence = $3
        where trip_id = $1
          and stop_sequence = 0
      `, tripID, ref.StopOrder, targetOrder); err != nil {
				return RouteStop{}, err
			}
		}

		if targetStopID != ref.StopID {
			if _, err := tx.Exec(ctx, `
        update trip_stops
        set stop_id = $1
        where trip_id = $2
          and stop_sequence = $3
      `, targetStopID, tripID, targetOrder); err != nil {
				return RouteStop{}, err
			}
		}
	}

	if targetStopID != ref.StopID {
		if _, err := tx.Exec(ctx, `
      delete from route_segment_prices
      where route_id = $1
        and (origin_stop_id = $2 or destination_stop_id = $2)
    `, routeID, ref.StopID); err != nil {
			return RouteStop{}, err
		}
	}

	if err := r.refreshAvailableSegmentsForRouteTx(ctx, tx, routeID); err != nil {
		return RouteStop{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return RouteStop{}, err
	}

	return r.getRouteStopByOrder(ctx, routeID, targetOrder)
}

func (r *Repository) DeleteStop(ctx context.Context, routeID string, stopID string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := r.ensureRouteExistsTx(ctx, tx, routeID); err != nil {
		return err
	}

	tripIDs, err := r.listRouteTripIDsTx(ctx, tx, routeID)
	if err != nil {
		return err
	}
	if len(tripIDs) == 0 {
		return errors.New("route has no trips to propagate stops")
	}

	ref, err := r.getRouteStopReferenceTx(ctx, tx, routeID, stopID)
	if err != nil {
		return err
	}

	for _, tripID := range tripIDs {
		if _, err := tx.Exec(ctx, `
      delete from trip_stops
      where trip_id = $1
        and stop_sequence = $2
    `, tripID, ref.StopOrder); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `
      update trip_stops
      set stop_sequence = stop_sequence - 1
      where trip_id = $1
        and stop_sequence > $2
    `, tripID, ref.StopOrder); err != nil {
			return err
		}
	}

	if _, err := tx.Exec(ctx, `
    delete from route_segment_prices
    where route_id = $1
      and (origin_stop_id = $2 or destination_stop_id = $2)
  `, routeID, ref.StopID); err != nil {
		return err
	}

	if err := r.refreshAvailableSegmentsForRouteTx(ctx, tx, routeID); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *Repository) GetStop(ctx context.Context, routeID string, stopID string) (RouteStop, error) {
	ref, err := r.getRouteStopReference(ctx, routeID, stopID)
	if err != nil {
		return RouteStop{}, err
	}
	return r.getRouteStopByOrder(ctx, routeID, ref.StopOrder)
}

func (r *Repository) ListSegmentPrices(ctx context.Context, routeID string) (RouteSegmentPriceMatrix, error) {
	if err := r.ensureRouteExists(ctx, routeID); err != nil {
		return RouteSegmentPriceMatrix{}, err
	}

	stops, err := r.listRouteOrderedStops(ctx, routeID)
	if err != nil {
		return RouteSegmentPriceMatrix{}, err
	}

	rows, err := r.pool.Query(ctx, `
    select
      origin_stop_id,
      destination_stop_id,
      price,
      status,
      created_at,
      updated_at
    from route_segment_prices
    where route_id = $1
  `, routeID)
	if err != nil {
		return RouteSegmentPriceMatrix{}, err
	}
	defer rows.Close()

	type configuredSegment struct {
		Price     *float64
		Status    string
		CreatedAt *time.Time
		UpdatedAt *time.Time
	}
	configuredByPair := make(map[string]configuredSegment)
	for rows.Next() {
		var origin string
		var destination string
		var item configuredSegment
		var rawStatus *string
		if err := rows.Scan(&origin, &destination, &item.Price, &rawStatus, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return RouteSegmentPriceMatrix{}, err
		}
		if rawStatus == nil {
			item.Status = "ACTIVE"
		} else {
			item.Status = strings.ToUpper(strings.TrimSpace(*rawStatus))
			if item.Status == "" {
				item.Status = "ACTIVE"
			}
		}
		configuredByPair[origin+"->"+destination] = item
	}
	if err := rows.Err(); err != nil {
		return RouteSegmentPriceMatrix{}, err
	}

	items := make([]RouteSegmentPriceItem, 0)
	for i := 0; i < len(stops); i++ {
		for j := i + 1; j < len(stops); j++ {
			item := RouteSegmentPriceItem{
				OriginStopID:           stops[i].StopID,
				OriginDisplayName:      stops[i].DisplayName,
				OriginStopOrder:        stops[i].StopOrder,
				DestinationStopID:      stops[j].StopID,
				DestinationDisplayName: stops[j].DisplayName,
				DestinationStopOrder:   stops[j].StopOrder,
				Configured:             false,
				Status:                 "MISSING",
			}
			key := item.OriginStopID + "->" + item.DestinationStopID
			if configured, ok := configuredByPair[key]; ok {
				item.Price = configured.Price
				item.Status = configured.Status
				item.CreatedAt = configured.CreatedAt
				item.UpdatedAt = configured.UpdatedAt
				item.Configured = true
			}
			items = append(items, item)
		}
	}

	return RouteSegmentPriceMatrix{
		RouteID: routeID,
		Stops:   stops,
		Items:   items,
	}, nil
}

func (r *Repository) UpsertSegmentPrices(ctx context.Context, routeID string, input UpsertRouteSegmentPricesInput) (RouteSegmentPriceMatrix, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return RouteSegmentPriceMatrix{}, err
	}
	defer tx.Rollback(ctx)

	if err := r.ensureRouteExistsTx(ctx, tx, routeID); err != nil {
		return RouteSegmentPriceMatrix{}, err
	}

	stopOrderByStopID, err := r.getRouteStopOrderMapTx(ctx, tx, routeID)
	if err != nil {
		return RouteSegmentPriceMatrix{}, err
	}

	type normalizedSegmentUpdate struct {
		OriginStopID      string
		DestinationStopID string
		Price             *float64
		Status            string
	}

	updates := make([]normalizedSegmentUpdate, 0, len(input.Items))
	seen := make(map[string]int, len(input.Items))

	for _, item := range input.Items {
		origin := strings.TrimSpace(item.OriginStopID)
		destination := strings.TrimSpace(item.DestinationStopID)
		originOrder, originOk := stopOrderByStopID[origin]
		destinationOrder, destinationOk := stopOrderByStopID[destination]
		if !originOk || !destinationOk || originOrder >= destinationOrder {
			return RouteSegmentPriceMatrix{}, ErrInvalidSegmentPair
		}

		status := "ACTIVE"
		if item.Status != nil {
			status = strings.ToUpper(strings.TrimSpace(*item.Status))
		}
		if status == "" {
			status = "ACTIVE"
		}

		key := origin + "->" + destination
		normalized := normalizedSegmentUpdate{
			OriginStopID:      origin,
			DestinationStopID: destination,
			Price:             item.Price,
			Status:            status,
		}

		if idx, ok := seen[key]; ok {
			updates[idx] = normalized
			continue
		}
		seen[key] = len(updates)
		updates = append(updates, normalized)
	}

	for _, item := range updates {
		if item.Price == nil {
			if _, err := tx.Exec(ctx, `
        delete from route_segment_prices
        where route_id = $1
          and origin_stop_id = $2
          and destination_stop_id = $3
      `, routeID, item.OriginStopID, item.DestinationStopID); err != nil {
				return RouteSegmentPriceMatrix{}, err
			}
			continue
		}

		if _, err := tx.Exec(ctx, `
      insert into route_segment_prices (
        route_id,
        origin_stop_id,
        destination_stop_id,
        price,
        status,
        created_at,
        updated_at
      ) values ($1, $2, $3, $4, $5, now(), now())
      on conflict (route_id, origin_stop_id, destination_stop_id)
      do update set
        price = excluded.price,
        status = excluded.status,
        updated_at = now()
    `, routeID, item.OriginStopID, item.DestinationStopID, *item.Price, item.Status); err != nil {
			return RouteSegmentPriceMatrix{}, err
		}
	}

	if err := r.refreshAvailableSegmentsForRouteTx(ctx, tx, routeID); err != nil {
		return RouteSegmentPriceMatrix{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return RouteSegmentPriceMatrix{}, err
	}

	return r.ListSegmentPrices(ctx, routeID)
}

func IsNotFound(err error) bool {
	return err == pgx.ErrNoRows
}

type routeValidationSnapshot struct {
	StopCount       int
	MinStopOrder    int
	MaxStopOrder    int
	FirstCity       string
	LastCity        string
	HasNegativeETA  bool
	HasETABacktrack bool
	HasLinkedTrips  bool
}

type routeStopReference struct {
	StopID    string
	StopOrder int
}

func scanRouteWithSnapshot(scanner interface {
	Scan(dest ...interface{}) error
}) (Route, routeValidationSnapshot, error) {
	item := Route{}
	snapshot := routeValidationSnapshot{}
	if err := scanner.Scan(
		&item.ID,
		&item.Name,
		&item.OriginCity,
		&item.DestinationCity,
		&item.IsActive,
		&item.DuplicatedFromRouteID,
		&item.CreatedAt,
		&snapshot.StopCount,
		&snapshot.MinStopOrder,
		&snapshot.MaxStopOrder,
		&snapshot.FirstCity,
		&snapshot.LastCity,
		&snapshot.HasNegativeETA,
		&snapshot.HasETABacktrack,
		&snapshot.HasLinkedTrips,
	); err != nil {
		return Route{}, routeValidationSnapshot{}, err
	}
	return item, snapshot, nil
}

func applyDerivedRouteState(item *Route, snapshot routeValidationSnapshot) {
	item.StopCount = snapshot.StopCount
	item.HasLinkedTrips = snapshot.HasLinkedTrips
	item.MissingRequirements = []string{}
	if item.IsActive {
		item.ConfigurationStatus = "ACTIVE"
		return
	}
	item.ConfigurationStatus = "READY"
}

func (r *Repository) ensureRouteExists(ctx context.Context, routeID string) error {
	return r.ensureRouteExistsTx(ctx, r.pool, routeID)
}

func (r *Repository) ensureRouteExistsTx(ctx context.Context, q interface {
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
}, routeID string) error {
	var exists bool
	if err := q.QueryRow(ctx, `select exists(select 1 from routes where route_id = $1)`, routeID).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *Repository) listRouteTripIDsTx(ctx context.Context, tx pgx.Tx, routeID string) ([]string, error) {
	rows, err := tx.Query(ctx, `
    select trip_id
    from trips
    where route_id = $1
    order by trip_date desc, created_at desc
  `, routeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]string, 0)
	for rows.Next() {
		var tripID string
		if err := rows.Scan(&tripID); err != nil {
			return nil, err
		}
		items = append(items, tripID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *Repository) normalizeStopOrderTx(ctx context.Context, tx pgx.Tx, tripID string, requested int, allowAppend bool) (int, error) {
	var count int
	if err := tx.QueryRow(ctx, `
    select count(*)
    from trip_stops
    where trip_id = $1
      and is_active = true
  `, tripID).Scan(&count); err != nil {
		return 0, err
	}

	if requested <= 0 {
		return 1, nil
	}
	maxOrder := count
	if allowAppend {
		maxOrder = count + 1
	}
	if requested > maxOrder {
		return maxOrder, nil
	}
	return requested, nil
}

func (r *Repository) findOrCreateStopTx(
	ctx context.Context,
	tx pgx.Tx,
	rawCity string,
	latitude *float64,
	longitude *float64,
) (string, string, error) {
	displayName, stopName, state := normalizeStopInput(rawCity)
	if displayName == "" {
		return "", "", errors.New("city is required")
	}

	var existingStopID string
	var existingDisplay string
	var existingLatitude *float64
	var existingLongitude *float64
	err := tx.QueryRow(ctx, `
    select
      stop_id,
      display_name,
      nullif(to_jsonb(stops)->>'latitude', '')::double precision as latitude,
      nullif(to_jsonb(stops)->>'longitude', '')::double precision as longitude
    from stops
    where lower(display_name) = lower($1)
    limit 1
  `, displayName).Scan(&existingStopID, &existingDisplay, &existingLatitude, &existingLongitude)
	if err == nil {
		if latitude != nil && longitude != nil {
			if err := r.updateStopCoordinatesTx(ctx, tx, existingStopID, latitude, longitude); err != nil {
				return "", "", err
			}
		}
		return existingStopID, existingDisplay, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return "", "", err
	}

	baseID := sanitizeStopToken(stopName)
	if baseID == "" {
		baseID = "STOP"
	}
	statePrefix := sanitizeStopToken(state)
	if statePrefix == "" {
		statePrefix = "ST"
	}
	candidateBase := statePrefix + "_" + baseID

	for attempt := 0; attempt < 64; attempt++ {
		candidateID := candidateBase
		if attempt > 0 {
			candidateID = fmt.Sprintf("%s_%d", candidateBase, attempt)
		}

		var candidateExists bool
		if err := tx.QueryRow(ctx, `
      select exists(select 1 from stops where stop_id = $1)
    `, candidateID).Scan(&candidateExists); err != nil {
			return "", "", err
		}
		if candidateExists {
			continue
		}

		_, insertErr := tx.Exec(ctx, `
      insert into stops (
        stop_id,
        stop_name,
        state,
        display_name,
        latitude,
        longitude,
        is_active,
        created_at
      ) values ($1, $2, nullif($3, ''), $4, $5, $6, true, now())
    `, candidateID, stopName, state, displayName, nullableFloat64(latitude), nullableFloat64(longitude))
		if insertErr != nil {
			var pgErr *pgconn.PgError
			if errors.As(insertErr, &pgErr) && pgErr.Code == "42703" {
				if latitude != nil || longitude != nil {
					return "", "", errors.New("latitude and longitude require database migration 0017_stops_geolocation")
				}
				_, insertErr = tx.Exec(ctx, `
          insert into stops (
            stop_id,
            stop_name,
            state,
            display_name,
            is_active,
            created_at
          ) values ($1, $2, nullif($3, ''), $4, true, now())
        `, candidateID, stopName, state, displayName)
			}
		}
		if insertErr == nil {
			return candidateID, displayName, nil
		}

		var pgErr *pgconn.PgError
		if errors.As(insertErr, &pgErr) && pgErr.Code == "23505" {
			return "", "", errors.New("stop insert conflict; retry the operation")
		}
		return "", "", insertErr
	}

	return "", "", errors.New("could not allocate stop id")
}

func (r *Repository) updateStopCoordinatesTx(
	ctx context.Context,
	tx pgx.Tx,
	stopID string,
	latitude *float64,
	longitude *float64,
) error {
	if latitude == nil || longitude == nil {
		return nil
	}
	_, err := tx.Exec(ctx, `
    update stops
    set latitude = $2,
        longitude = $3
    where stop_id = $1
  `, stopID, *latitude, *longitude)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "42703" {
		return errors.New("latitude and longitude require database migration 0017_stops_geolocation")
	}
	return err
}

func (r *Repository) insertTripStopTx(ctx context.Context, tx pgx.Tx, tripID string, stopID string, stopOrder int, departAt *time.Time) error {
	for attempt := 0; attempt < 16; attempt++ {
		tripStopID := fmt.Sprintf("%s_%s_%d", tripID, strings.ToLower(stopID), time.Now().UnixNano()+int64(attempt))
		_, err := tx.Exec(ctx, `
      insert into trip_stops (
        trip_stop_id,
        trip_id,
        stop_id,
        stop_sequence,
        depart_time,
        is_active,
        created_at
      ) values ($1, $2, $3, $4, $5, true, now())
    `, tripStopID, tripID, stopID, stopOrder, departAt)
		if err == nil {
			return nil
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			continue
		}
		return err
	}
	return errors.New("could not insert trip stop after retries")
}

func (r *Repository) getRouteStopReference(ctx context.Context, routeID string, tripStopID string) (routeStopReference, error) {
	row := r.pool.QueryRow(ctx, `
    with latest_trip as (
      select trip_id
      from trips
      where route_id = $1
      order by trip_date desc, created_at desc
      limit 1
    )
    select ts.stop_id, ts.stop_sequence
    from latest_trip lt
    join trip_stops ts on ts.trip_id = lt.trip_id
    where ts.trip_stop_id = $2
      and ts.is_active = true
  `, routeID, tripStopID)

	var ref routeStopReference
	if err := row.Scan(&ref.StopID, &ref.StopOrder); err != nil {
		return routeStopReference{}, err
	}
	return ref, nil
}

func (r *Repository) getRouteStopReferenceTx(ctx context.Context, tx pgx.Tx, routeID string, tripStopID string) (routeStopReference, error) {
	row := tx.QueryRow(ctx, `
    with latest_trip as (
      select trip_id
      from trips
      where route_id = $1
      order by trip_date desc, created_at desc
      limit 1
    )
    select ts.stop_id, ts.stop_sequence
    from latest_trip lt
    join trip_stops ts on ts.trip_id = lt.trip_id
    where ts.trip_stop_id = $2
      and ts.is_active = true
  `, routeID, tripStopID)

	var ref routeStopReference
	if err := row.Scan(&ref.StopID, &ref.StopOrder); err != nil {
		return routeStopReference{}, err
	}
	return ref, nil
}

func (r *Repository) getRouteStopByOrder(ctx context.Context, routeID string, order int) (RouteStop, error) {
	row := r.pool.QueryRow(ctx, `
    with latest_trip as (
      select trip_id
      from trips
      where route_id = $1
      order by trip_date desc, created_at desc
      limit 1
    )
    select
      ts.trip_stop_id as id,
      $1 as route_id,
      s.display_name as city,
      nullif(to_jsonb(s)->>'latitude', '')::double precision as latitude,
      nullif(to_jsonb(s)->>'longitude', '')::double precision as longitude,
      ts.stop_sequence as stop_order,
      null::int as eta_offset_minutes,
      null::text as notes,
      ts.created_at
    from latest_trip lt
    join trip_stops ts on ts.trip_id = lt.trip_id
    join stops s on s.stop_id = ts.stop_id
    where ts.is_active = true
      and ts.stop_sequence = $2
    limit 1
  `, routeID, order)

	var item RouteStop
	if err := row.Scan(
		&item.ID,
		&item.RouteID,
		&item.City,
		&item.Latitude,
		&item.Longitude,
		&item.StopOrder,
		&item.ETAOffsetMinutes,
		&item.Notes,
		&item.CreatedAt,
	); err != nil {
		return RouteStop{}, err
	}
	return item, nil
}

func (r *Repository) listRouteOrderedStops(ctx context.Context, routeID string) ([]RouteSegmentPriceStop, error) {
	rows, err := r.pool.Query(ctx, `
    with latest_trip as (
      select trip_id
      from trips
      where route_id = $1
      order by trip_date desc, created_at desc
      limit 1
    )
    select
      ts.stop_id,
      s.display_name,
      ts.stop_sequence
    from latest_trip lt
    join trip_stops ts on ts.trip_id = lt.trip_id
    join stops s on s.stop_id = ts.stop_id
    where ts.is_active = true
    order by ts.stop_sequence asc
  `, routeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]RouteSegmentPriceStop, 0)
	for rows.Next() {
		var item RouteSegmentPriceStop
		if err := rows.Scan(&item.StopID, &item.DisplayName, &item.StopOrder); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *Repository) getRouteStopOrderMapTx(ctx context.Context, tx pgx.Tx, routeID string) (map[string]int, error) {
	rows, err := tx.Query(ctx, `
    with latest_trip as (
      select trip_id
      from trips
      where route_id = $1
      order by trip_date desc, created_at desc
      limit 1
    )
    select ts.stop_id, ts.stop_sequence
    from latest_trip lt
    join trip_stops ts on ts.trip_id = lt.trip_id
    where ts.is_active = true
  `, routeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]int)
	for rows.Next() {
		var stopID string
		var stopOrder int
		if err := rows.Scan(&stopID, &stopOrder); err != nil {
			return nil, err
		}
		result[stopID] = stopOrder
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *Repository) refreshAvailableSegmentsForRouteTx(ctx context.Context, tx pgx.Tx, routeID string) error {
	_, err := tx.Exec(ctx, `select refresh_available_segments_for_route($1)`, routeID)
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "42883" {
		return nil
	}
	return err
}

func normalizeStopInput(raw string) (displayName string, stopName string, state string) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", "", ""
	}

	if strings.Contains(trimmed, "/") {
		parts := strings.Split(trimmed, "/")
		namePart := strings.TrimSpace(strings.Join(parts[:len(parts)-1], "/"))
		statePart := strings.TrimSpace(parts[len(parts)-1])
		if namePart != "" && statePart != "" {
			stopName = namePart
			state = strings.ToUpper(statePart)
			displayName = stopName + "/" + state
			return displayName, stopName, state
		}
	}

	stopName = trimmed
	state = ""
	displayName = stopName
	return displayName, stopName, state
}

func sanitizeStopToken(raw string) string {
	upper := strings.ToUpper(strings.TrimSpace(raw))
	var builder strings.Builder
	lastUnderscore := false
	for _, ch := range upper {
		isLetter := ch >= 'A' && ch <= 'Z'
		isDigit := ch >= '0' && ch <= '9'
		if isLetter || isDigit {
			builder.WriteRune(ch)
			lastUnderscore = false
			continue
		}
		if !lastUnderscore {
			builder.WriteRune('_')
			lastUnderscore = true
		}
	}
	return strings.Trim(builder.String(), "_")
}

func valueOrDefaultInt(value *int, fallback int) int {
	if value == nil {
		return fallback
	}
	return *value
}

func nullableFloat64(value *float64) interface{} {
	if value == nil {
		return nil
	}
	return *value
}

func (r *Repository) allocateRouteIDTx(ctx context.Context, tx pgx.Tx, originCity string, destinationCity string) (string, error) {
	originCode := deriveRouteCode(originCity)
	destinationCode := deriveRouteCode(destinationCity)
	if originCode == "" || destinationCode == "" {
		return "", errors.New("could not derive route id from origin/destination")
	}

	base := originCode + "_" + destinationCode
	for attempt := 0; attempt < 100; attempt++ {
		candidate := base
		if attempt > 0 {
			candidate = fmt.Sprintf("%s_%d", base, attempt+1)
		}
		var exists bool
		if err := tx.QueryRow(ctx, `select exists(select 1 from routes where route_id = $1)`, candidate).Scan(&exists); err != nil {
			return "", err
		}
		if !exists {
			return candidate, nil
		}
	}
	return "", errors.New("could not allocate route_id")
}

func (r *Repository) allocateTripIDTx(ctx context.Context, tx pgx.Tx, routeID string) (string, error) {
	base := strings.ToLower(strings.ReplaceAll(routeID, "_", "-")) + "-template"
	for attempt := 0; attempt < 100; attempt++ {
		candidate := base
		if attempt > 0 {
			candidate = fmt.Sprintf("%s-%d", base, attempt+1)
		}
		var exists bool
		if err := tx.QueryRow(ctx, `select exists(select 1 from trips where trip_id = $1)`, candidate).Scan(&exists); err != nil {
			return "", err
		}
		if !exists {
			return candidate, nil
		}
	}
	return "", errors.New("could not allocate trip_id")
}

func deriveRouteCode(city string) string {
	trimmed := strings.TrimSpace(city)
	if trimmed == "" {
		return ""
	}
	if strings.Contains(trimmed, "/") {
		parts := strings.Split(trimmed, "/")
		last := strings.TrimSpace(parts[len(parts)-1])
		if last != "" {
			token := sanitizeStopToken(last)
			if len(token) > 4 {
				return token[:4]
			}
			return token
		}
	}
	token := sanitizeStopToken(trimmed)
	if len(token) > 4 {
		return token[:4]
	}
	return token
}
