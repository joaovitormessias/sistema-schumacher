package routes

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrRouteHasLinkedTrips = errors.New("route has linked non-cancelled trips")

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
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

const routeSelectQuery = `select
    r.id::text,
    r.name,
    r.origin_city,
    r.destination_city,
    r.is_active,
    r.duplicated_from_route_id::text,
    r.created_at,
    stats.stop_count,
    stats.min_stop_order,
    stats.max_stop_order,
    stats.first_city,
    stats.last_city,
    stats.has_negative_eta,
    stats.has_eta_backtrack,
    links.has_linked_trips
  from routes r
  left join lateral (
    select
      count(*)::int as stop_count,
      coalesce(min(rs.stop_order), 0)::int as min_stop_order,
      coalesce(max(rs.stop_order), 0)::int as max_stop_order,
      coalesce((select rs_first.city from route_stops rs_first where rs_first.route_id = r.id order by rs_first.stop_order asc limit 1), '') as first_city,
      coalesce((select rs_last.city from route_stops rs_last where rs_last.route_id = r.id order by rs_last.stop_order desc limit 1), '') as last_city,
      exists(
        select 1 from route_stops rs_eta_negative
        where rs_eta_negative.route_id = r.id
          and rs_eta_negative.eta_offset_minutes is not null
          and rs_eta_negative.eta_offset_minutes < 0
      ) as has_negative_eta,
      exists(
        select 1
        from (
          select
            rs_eta.eta_offset_minutes,
            lag(rs_eta.eta_offset_minutes) over (order by rs_eta.stop_order asc, rs_eta.created_at asc) as prev_eta
          from route_stops rs_eta
          where rs_eta.route_id = r.id
            and rs_eta.eta_offset_minutes is not null
        ) eta
        where eta.prev_eta is not null
          and eta.eta_offset_minutes < eta.prev_eta
      ) as has_eta_backtrack
    from route_stops rs
    where rs.route_id = r.id
  ) stats on true
  left join lateral (
    select exists(
      select 1 from trips t
      where t.route_id = r.id
        and t.status <> 'CANCELLED'
    ) as has_linked_trips
  ) links on true`

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
      r.name ilike $%d
      or r.origin_city ilike $%d
      or r.destination_city ilike $%d
      or r.id::text ilike $%d
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

func (r *Repository) Get(ctx context.Context, id uuid.UUID) (Route, error) {
	query := routeSelectQuery + " where r.id=$1"
	row := r.pool.QueryRow(ctx, query, id)
	item, snapshot, err := scanRouteWithSnapshot(row)
	if err != nil {
		return Route{}, err
	}
	applyDerivedRouteState(&item, snapshot)
	return item, nil
}

func (r *Repository) Create(ctx context.Context, input CreateRouteInput) (Route, error) {
	isActive := false
	if input.IsActive != nil {
		isActive = *input.IsActive
	}

	var id uuid.UUID
	row := r.pool.QueryRow(ctx,
		`insert into routes (name, origin_city, destination_city, is_active)
     values ($1,$2,$3,$4)
     returning id`,
		input.Name, input.OriginCity, input.DestinationCity, isActive,
	)
	if err := row.Scan(&id); err != nil {
		return Route{}, err
	}
	return r.Get(ctx, id)
}

func (r *Repository) Update(ctx context.Context, id uuid.UUID, input UpdateRouteInput) (Route, error) {
	sets := []string{}
	args := []interface{}{}
	idx := 1

	if input.Name != nil {
		sets = append(sets, fmt.Sprintf("name=$%d", idx))
		args = append(args, *input.Name)
		idx++
	}
	if input.OriginCity != nil {
		sets = append(sets, fmt.Sprintf("origin_city=$%d", idx))
		args = append(args, *input.OriginCity)
		idx++
	}
	if input.DestinationCity != nil {
		sets = append(sets, fmt.Sprintf("destination_city=$%d", idx))
		args = append(args, *input.DestinationCity)
		idx++
	}
	if input.IsActive != nil {
		sets = append(sets, fmt.Sprintf("is_active=$%d", idx))
		args = append(args, *input.IsActive)
		idx++
	}

	if len(sets) == 0 {
		return r.Get(ctx, id)
	}

	args = append(args, id)
	query := fmt.Sprintf("update routes set %s where id=$%d", strings.Join(sets, ", "), idx)
	tag, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return Route{}, err
	}
	if tag.RowsAffected() == 0 {
		return Route{}, pgx.ErrNoRows
	}

	return r.Get(ctx, id)
}

func (r *Repository) Publish(ctx context.Context, id uuid.UUID) (Route, error) {
	tag, err := r.pool.Exec(ctx, `update routes set is_active=true where id=$1`, id)
	if err != nil {
		return Route{}, err
	}
	if tag.RowsAffected() == 0 {
		return Route{}, pgx.ErrNoRows
	}
	return r.Get(ctx, id)
}

func (r *Repository) Duplicate(ctx context.Context, id uuid.UUID) (Route, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Route{}, err
	}
	defer tx.Rollback(ctx)

	var name string
	var originCity string
	var destinationCity string
	row := tx.QueryRow(ctx, `select name, origin_city, destination_city from routes where id=$1`, id)
	if err := row.Scan(&name, &originCity, &destinationCity); err != nil {
		return Route{}, err
	}

	var duplicatedID uuid.UUID
	row = tx.QueryRow(ctx, `
    insert into routes (name, origin_city, destination_city, is_active, duplicated_from_route_id)
    values ($1, $2, $3, false, $4)
    returning id
  `, name+" (Copia)", originCity, destinationCity, id)
	if err := row.Scan(&duplicatedID); err != nil {
		return Route{}, err
	}

	if _, err := tx.Exec(ctx, `
    insert into route_stops (route_id, city, stop_order, eta_offset_minutes, notes)
    select $2, city, stop_order, eta_offset_minutes, notes
    from route_stops
    where route_id = $1
    order by stop_order asc
  `, id, duplicatedID); err != nil {
		return Route{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return Route{}, err
	}

	return r.Get(ctx, duplicatedID)
}

func (r *Repository) HasLinkedTrips(ctx context.Context, routeID uuid.UUID) (bool, error) {
	var hasLinkedTrips bool
	if err := r.pool.QueryRow(ctx, `
    select exists(
      select 1 from trips t
      where t.route_id = $1
        and t.status <> 'CANCELLED'
    )
  `, routeID).Scan(&hasLinkedTrips); err != nil {
		return false, err
	}
	return hasLinkedTrips, nil
}

func (r *Repository) ListStops(ctx context.Context, routeID uuid.UUID) ([]RouteStop, error) {
	rows, err := r.pool.Query(ctx,
		`select id, route_id, city, stop_order, eta_offset_minutes, notes, created_at
     from route_stops
     where route_id = $1
     order by stop_order asc`, routeID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []RouteStop{}
	for rows.Next() {
		var item RouteStop
		if err := rows.Scan(&item.ID, &item.RouteID, &item.City, &item.StopOrder, &item.ETAOffsetMinutes, &item.Notes, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) CreateStop(ctx context.Context, routeID uuid.UUID, input CreateRouteStopInput) (RouteStop, error) {
	var item RouteStop
	row := r.pool.QueryRow(ctx,
		`insert into route_stops (route_id, city, stop_order, eta_offset_minutes, notes)
     values ($1,$2,$3,$4,$5)
     returning id, route_id, city, stop_order, eta_offset_minutes, notes, created_at`,
		routeID, input.City, input.StopOrder, input.ETAOffsetMinutes, input.Notes,
	)
	if err := row.Scan(&item.ID, &item.RouteID, &item.City, &item.StopOrder, &item.ETAOffsetMinutes, &item.Notes, &item.CreatedAt); err != nil {
		return item, err
	}
	return item, nil
}

func (r *Repository) UpdateStop(ctx context.Context, routeID uuid.UUID, stopID uuid.UUID, input UpdateRouteStopInput) (RouteStop, error) {
	sets := []string{}
	args := []interface{}{}
	idx := 1

	if input.City != nil {
		sets = append(sets, fmt.Sprintf("city=$%d", idx))
		args = append(args, *input.City)
		idx++
	}
	if input.StopOrder != nil {
		sets = append(sets, fmt.Sprintf("stop_order=$%d", idx))
		args = append(args, *input.StopOrder)
		idx++
	}
	if input.ETAOffsetMinutes != nil {
		sets = append(sets, fmt.Sprintf("eta_offset_minutes=$%d", idx))
		args = append(args, *input.ETAOffsetMinutes)
		idx++
	}
	if input.Notes != nil {
		sets = append(sets, fmt.Sprintf("notes=$%d", idx))
		args = append(args, *input.Notes)
		idx++
	}

	if len(sets) == 0 {
		return r.GetStop(ctx, routeID, stopID)
	}

	args = append(args, routeID, stopID)
	query := fmt.Sprintf(`update route_stops set %s where route_id=$%d and id=$%d returning id, route_id, city, stop_order, eta_offset_minutes, notes, created_at`, strings.Join(sets, ", "), idx, idx+1)

	var item RouteStop
	row := r.pool.QueryRow(ctx, query, args...)
	if err := row.Scan(&item.ID, &item.RouteID, &item.City, &item.StopOrder, &item.ETAOffsetMinutes, &item.Notes, &item.CreatedAt); err != nil {
		return item, err
	}
	return item, nil
}

func (r *Repository) DeleteStop(ctx context.Context, routeID uuid.UUID, stopID uuid.UUID) error {
	hasLinkedTrips, err := r.HasLinkedTrips(ctx, routeID)
	if err != nil {
		return err
	}
	if hasLinkedTrips {
		return ErrRouteHasLinkedTrips
	}

	tag, err := r.pool.Exec(ctx, `delete from route_stops where route_id=$1 and id=$2`, routeID, stopID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *Repository) GetStop(ctx context.Context, routeID uuid.UUID, stopID uuid.UUID) (RouteStop, error) {
	var item RouteStop
	row := r.pool.QueryRow(ctx,
		`select id, route_id, city, stop_order, eta_offset_minutes, notes, created_at
     from route_stops
     where route_id=$1 and id=$2`, routeID, stopID,
	)
	if err := row.Scan(&item.ID, &item.RouteID, &item.City, &item.StopOrder, &item.ETAOffsetMinutes, &item.Notes, &item.CreatedAt); err != nil {
		return item, err
	}
	return item, nil
}

func IsNotFound(err error) bool {
	return err == pgx.ErrNoRows
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
	missing := []string{}

	if snapshot.StopCount < 2 {
		missing = append(missing, "at least two stops are required")
	}

	if snapshot.StopCount > 0 {
		if snapshot.MinStopOrder != 1 {
			missing = append(missing, "stop_order must start at 1")
		}
		if (snapshot.MaxStopOrder - snapshot.MinStopOrder + 1) != snapshot.StopCount {
			missing = append(missing, "stop_order must be sequential without gaps")
		}
		if !sameCity(item.OriginCity, snapshot.FirstCity) {
			missing = append(missing, "first stop city must match origin_city")
		}
		if !sameCity(item.DestinationCity, snapshot.LastCity) {
			missing = append(missing, "last stop city must match destination_city")
		}
	}

	if snapshot.HasNegativeETA {
		missing = append(missing, "eta_offset_minutes must be >= 0")
	}
	if snapshot.HasETABacktrack {
		missing = append(missing, "eta_offset_minutes must be non-decreasing")
	}

	item.StopCount = snapshot.StopCount
	item.MissingRequirements = missing
	item.HasLinkedTrips = snapshot.HasLinkedTrips
	item.ConfigurationStatus = resolveConfigurationStatus(item.IsActive, snapshot.HasLinkedTrips, len(missing) == 0)
}

func resolveConfigurationStatus(isActive bool, hasLinkedTrips bool, isReady bool) string {
	if isActive {
		if isReady {
			return "ACTIVE"
		}
		return "INCOMPLETE"
	}
	if hasLinkedTrips {
		return "SUSPENDED"
	}
	if isReady {
		return "READY"
	}
	return "INCOMPLETE"
}

func sameCity(a string, b string) bool {
	return strings.EqualFold(strings.TrimSpace(a), strings.TrimSpace(b))
}
