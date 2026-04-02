package trips

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

var ErrUnsupported = errors.New("operation not supported in production ticketing schema")
var ErrTripNotFound = errors.New("trip not found")

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) List(ctx context.Context, filter ListFilter) ([]Trip, error) {
	limit := filter.Limit
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	query := `select
    t.trip_id as id,
    t.route_id,
    coalesce(t.bus_id, '') as bus_id,
    null::text as driver_id,
    null::text as fare_id,
    null::text as request_id,
    (t.trip_date::timestamp) as departure_at,
    null::timestamp as arrival_at,
    t.status,
    t.status as operational_status,
    greatest(coalesce(t.seats_total, 0), 0) as seats_total,
    greatest(coalesce(t.seats_available, 0), 0) as seats_available,
    0::numeric as estimated_km,
    null::timestamp as dispatch_validated_at,
    null::text as dispatch_validated_by,
    null::text as pair_trip_id,
    null::text as notes,
    t.created_at,
    t.created_at as updated_at
    from trips t`
	args := []interface{}{}
	clauses := []string{}

	if filter.Status != "" && filter.Status != "ALL" {
		args = append(args, filter.Status)
		clauses = append(clauses, fmt.Sprintf("t.status = $%d", len(args)))
	}

	if filter.Search != "" {
		args = append(args, "%"+filter.Search+"%")
		clauses = append(clauses, fmt.Sprintf(`(
      t.trip_id ilike $%d
      or t.route_id ilike $%d
      or coalesce(t.bus_id, '') ilike $%d
      or t.status ilike $%d
    )`, len(args), len(args), len(args), len(args)))
	}

	if len(clauses) > 0 {
		query += " where " + strings.Join(clauses, " and ")
	}

	query += " order by t.trip_date desc, t.created_at desc"
	args = append(args, limit)
	query += fmt.Sprintf(" limit $%d", len(args))
	args = append(args, offset)
	query += fmt.Sprintf(" offset $%d", len(args))

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []Trip{}
	for rows.Next() {
		var item Trip
		if err := scanTrip(rows, &item); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) Get(ctx context.Context, id string) (Trip, error) {
	var item Trip
	row := r.pool.QueryRow(ctx, `select
    t.trip_id as id,
    t.route_id,
    coalesce(t.bus_id, '') as bus_id,
    null::text as driver_id,
    null::text as fare_id,
    null::text as request_id,
    (t.trip_date::timestamp) as departure_at,
    null::timestamp as arrival_at,
    t.status,
    t.status as operational_status,
    greatest(coalesce(t.seats_total, 0), 0) as seats_total,
    greatest(coalesce(t.seats_available, 0), 0) as seats_available,
    0::numeric as estimated_km,
    null::timestamp as dispatch_validated_at,
    null::text as dispatch_validated_by,
    null::text as pair_trip_id,
    null::text as notes,
    t.created_at,
    t.created_at as updated_at
    from trips t where t.trip_id=$1`, id)
	if err := scanTrip(row, &item); err != nil {
		return item, err
	}
	return item, nil
}

func (r *Repository) Create(ctx context.Context, input CreateTripInput) (Trip, error) {
	return Trip{}, ErrUnsupported
}

func (r *Repository) Update(ctx context.Context, id string, input UpdateTripInput) (Trip, error) {
	return Trip{}, ErrUnsupported
}

func (r *Repository) ListSeats(ctx context.Context, tripID string, boardStopID *string, alightStopID *string) ([]TripSeat, error) {
	rows, err := r.pool.Query(ctx, `
    with trip_capacity as (
      select greatest(coalesce(seats_total, 0), 0) as seats_total
      from trips
      where trip_id = $1
      limit 1
    )
    select
      gs::text as id,
      gs as seat_number,
      true as is_active,
      exists (
        select 1
        from passengers p
        join bookings b on b.booking_id = p.booking_id
        where p.trip_id = $1
          and p.seat_number = gs::text
          and upper(coalesce(b.status, '')) not in ('CANCELLED', 'EXPIRED')
      ) as is_taken
    from trip_capacity tc
    cross join generate_series(1, coalesce(nullif(tc.seats_total, 0), 50)) as gs
    order by gs asc
  `, tripID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	seats := []TripSeat{}
	for rows.Next() {
		var seat TripSeat
		if err := rows.Scan(&seat.ID, &seat.SeatNumber, &seat.IsActive, &seat.IsTaken); err != nil {
			return nil, err
		}
		seats = append(seats, seat)
	}
	return seats, rows.Err()
}

func (r *Repository) ListStops(ctx context.Context, tripID string) ([]TripStop, error) {
	rows, err := r.pool.Query(ctx, `
    select
      ts.trip_stop_id as id,
      ts.trip_id,
      ts.stop_id as route_stop_id,
      s.display_name as city,
      ts.stop_sequence as stop_order,
      0::numeric as leg_distance_km,
      0::numeric as cumulative_distance_km,
      null::timestamp as arrive_at,
      (ts.depart_time::timestamp) as depart_at,
      ts.created_at,
      ts.created_at as updated_at
    from trip_stops ts
    join stops s on s.stop_id = ts.stop_id
    where ts.trip_id = $1
      and ts.is_active = true
    order by ts.stop_sequence asc
  `, tripID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []TripStop{}
	for rows.Next() {
		var item TripStop
		if err := rowscanTripStop(rows, &item); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) CreateStop(ctx context.Context, tripID string, input CreateTripStopInput) (TripStop, error) {
	return TripStop{}, ErrUnsupported
}

func (r *Repository) ListSegmentPrices(ctx context.Context, tripID string) (TripSegmentPriceMatrix, error) {
	routeID, err := r.getTripRouteID(ctx, tripID)
	if err != nil {
		return TripSegmentPriceMatrix{}, err
	}

	stops, err := r.listTripOrderedStops(ctx, tripID)
	if err != nil {
		return TripSegmentPriceMatrix{}, err
	}

	rows, err := r.pool.Query(ctx, `
    with ordered_stops as (
      select
        ts.stop_id,
        s.display_name,
        ts.stop_sequence
      from trip_stops ts
      join stops s on s.stop_id = ts.stop_id
      where ts.trip_id = $1
        and ts.is_active = true
      order by ts.stop_sequence asc
    )
    select
      o.stop_id as origin_stop_id,
      o.display_name as origin_display_name,
      o.stop_sequence as origin_stop_order,
      d.stop_id as destination_stop_id,
      d.display_name as destination_display_name,
      d.stop_sequence as destination_stop_order,
      rsp.price,
      rsp.status,
      rsp.created_at,
      rsp.updated_at,
      (rsp.route_id is not null) as configured
    from ordered_stops o
    join ordered_stops d on d.stop_sequence > o.stop_sequence
    left join route_segment_prices rsp
      on rsp.route_id = $2
     and rsp.origin_stop_id = o.stop_id
     and rsp.destination_stop_id = d.stop_id
    order by o.stop_sequence asc, d.stop_sequence asc
  `, tripID, routeID)
	if err != nil {
		return TripSegmentPriceMatrix{}, err
	}
	defer rows.Close()

	items := make([]TripSegmentPriceItem, 0)
	for rows.Next() {
		var item TripSegmentPriceItem
		var status *string
		var createdAt *time.Time
		var updatedAt *time.Time
		if err := rows.Scan(
			&item.OriginStopID,
			&item.OriginDisplayName,
			&item.OriginStopOrder,
			&item.DestinationStopID,
			&item.DestinationDisplayName,
			&item.DestinationStopOrder,
			&item.Price,
			&status,
			&createdAt,
			&updatedAt,
			&item.Configured,
		); err != nil {
			return TripSegmentPriceMatrix{}, err
		}

		if status == nil || !item.Configured {
			item.Status = "MISSING"
		} else {
			item.Status = strings.ToUpper(strings.TrimSpace(*status))
		}
		item.CreatedAt = createdAt
		item.UpdatedAt = updatedAt
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return TripSegmentPriceMatrix{}, err
	}

	return TripSegmentPriceMatrix{
		TripID:  tripID,
		RouteID: routeID,
		Stops:   stops,
		Items:   items,
	}, nil
}

func (r *Repository) UpsertSegmentPrices(ctx context.Context, tripID string, input UpsertTripSegmentPricesInput) (TripSegmentPriceMatrix, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return TripSegmentPriceMatrix{}, err
	}
	defer tx.Rollback(ctx)

	routeID, err := r.getTripRouteIDTx(ctx, tx, tripID)
	if err != nil {
		return TripSegmentPriceMatrix{}, err
	}

	stopOrderByStopID, err := r.getTripStopOrderMapTx(ctx, tx, tripID)
	if err != nil {
		return TripSegmentPriceMatrix{}, err
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
			return TripSegmentPriceMatrix{}, ErrInvalidSegmentPair
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
				return TripSegmentPriceMatrix{}, err
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
			return TripSegmentPriceMatrix{}, err
		}
	}

	if err := r.refreshAvailableSegmentsForRouteTx(ctx, tx, routeID); err != nil {
		return TripSegmentPriceMatrix{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return TripSegmentPriceMatrix{}, err
	}

	return r.ListSegmentPrices(ctx, tripID)
}

func (r *Repository) EnsureTripStopsFromRoute(ctx context.Context, tripID string) error {
	return nil
}

func (r *Repository) ValidateRouteReadiness(ctx context.Context, routeID string) ([]string, error) {
	return []string{}, nil
}

func (r *Repository) getTripRouteID(ctx context.Context, tripID string) (string, error) {
	var routeID string
	if err := r.pool.QueryRow(ctx, `select route_id from trips where trip_id = $1`, tripID).Scan(&routeID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrTripNotFound
		}
		return "", err
	}
	return routeID, nil
}

func (r *Repository) getTripRouteIDTx(ctx context.Context, tx pgx.Tx, tripID string) (string, error) {
	var routeID string
	if err := tx.QueryRow(ctx, `select route_id from trips where trip_id = $1`, tripID).Scan(&routeID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrTripNotFound
		}
		return "", err
	}
	return routeID, nil
}

func (r *Repository) listTripOrderedStops(ctx context.Context, tripID string) ([]TripSegmentPriceStop, error) {
	rows, err := r.pool.Query(ctx, `
    select
      ts.stop_id,
      s.display_name,
      ts.stop_sequence
    from trip_stops ts
    join stops s on s.stop_id = ts.stop_id
    where ts.trip_id = $1
      and ts.is_active = true
    order by ts.stop_sequence asc
  `, tripID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]TripSegmentPriceStop, 0)
	for rows.Next() {
		var item TripSegmentPriceStop
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

func (r *Repository) getTripStopOrderMapTx(ctx context.Context, tx pgx.Tx, tripID string) (map[string]int, error) {
	rows, err := tx.Query(ctx, `
    select stop_id, stop_sequence
    from trip_stops
    where trip_id = $1
      and is_active = true
  `, tripID)
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
		// Function might not exist before migration rollout.
		return nil
	}
	return err
}

func scanTrip(scanner interface {
	Scan(dest ...interface{}) error
}, item *Trip) error {
	return scanner.Scan(
		&item.ID,
		&item.RouteID,
		&item.BusID,
		&item.DriverID,
		&item.FareID,
		&item.RequestID,
		&item.DepartureAt,
		&item.ArrivalAt,
		&item.Status,
		&item.OperationalStatus,
		&item.SeatsTotal,
		&item.SeatsAvailable,
		&item.EstimatedKM,
		&item.DispatchValidatedAt,
		&item.DispatchValidatedBy,
		&item.PairTripID,
		&item.Notes,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
}

func rowscanTripStop(scanner interface {
	Scan(dest ...interface{}) error
}, item *TripStop) error {
	return scanner.Scan(
		&item.ID,
		&item.TripID,
		&item.RouteStopID,
		&item.City,
		&item.StopOrder,
		&item.LegDistanceKM,
		&item.CumulativeDistanceKM,
		&item.ArriveAt,
		&item.DepartAt,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
}

func IsNotFound(err error) bool {
	return err == pgx.ErrNoRows
}
