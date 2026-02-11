package trips

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

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

	query := `select id, route_id, bus_id, driver_id, fare_id, request_id, departure_at, arrival_at, status,
    operational_status, estimated_km, dispatch_validated_at, dispatch_validated_by, pair_trip_id, notes,
    created_at, updated_at
    from trips`
	args := []interface{}{}
	clauses := []string{}

	if filter.Status != "" && filter.Status != "ALL" {
		args = append(args, filter.Status)
		clauses = append(clauses, fmt.Sprintf("status = $%d", len(args)))
	}

	if filter.Search != "" {
		args = append(args, "%"+filter.Search+"%")
		clauses = append(clauses, fmt.Sprintf(`(
      id::text ilike $%d
      or route_id::text ilike $%d
      or bus_id::text ilike $%d
      or coalesce(driver_id::text, '') ilike $%d
      or status ilike $%d
      or operational_status ilike $%d
    )`, len(args), len(args), len(args), len(args), len(args), len(args)))
	}

	if len(clauses) > 0 {
		query += " where " + strings.Join(clauses, " and ")
	}

	query += " order by departure_at desc"
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

func (r *Repository) Get(ctx context.Context, id uuid.UUID) (Trip, error) {
	var item Trip
	row := r.pool.QueryRow(ctx, `select id, route_id, bus_id, driver_id, fare_id, request_id, departure_at, arrival_at, status,
    operational_status, estimated_km, dispatch_validated_at, dispatch_validated_by, pair_trip_id, notes,
    created_at, updated_at
    from trips where id=$1`, id)
	if err := scanTrip(row, &item); err != nil {
		return item, err
	}
	return item, nil
}

func (r *Repository) Create(ctx context.Context, input CreateTripInput) (Trip, error) {
	status := "SCHEDULED"
	if input.Status != nil {
		status = *input.Status
	}

	estimatedKM := 0.0
	if input.EstimatedKM != nil {
		estimatedKM = *input.EstimatedKM
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Trip{}, err
	}
	defer tx.Rollback(ctx)

	var item Trip
	row := tx.QueryRow(ctx,
		`insert into trips (route_id, bus_id, driver_id, fare_id, request_id, departure_at, arrival_at, status, operational_status, estimated_km, pair_trip_id, notes, updated_at)
     values ($1,$2,$3,$4,$5,$6,$7,$8,'REQUESTED',$9,$10,$11,now())
     returning id, route_id, bus_id, driver_id, fare_id, request_id, departure_at, arrival_at, status,
       operational_status, estimated_km, dispatch_validated_at, dispatch_validated_by, pair_trip_id, notes,
       created_at, updated_at`,
		input.RouteID, input.BusID, input.DriverID, input.FareID, nullableUUID(input.RequestID), input.DepartureAt, input.ArrivalAt, status, estimatedKM, nullableUUID(input.PairTripID), input.Notes,
	)
	if err := scanTrip(row, &item); err != nil {
		return item, err
	}

	if err := r.ensureTripStopsFromRouteTx(ctx, tx, uuid.MustParse(item.ID)); err != nil {
		return Trip{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return Trip{}, err
	}
	return item, nil
}

func (r *Repository) Update(ctx context.Context, id uuid.UUID, input UpdateTripInput) (Trip, error) {
	sets := []string{}
	args := []interface{}{}
	idx := 1

	if input.RouteID != nil {
		sets = append(sets, fmt.Sprintf("route_id=$%d", idx))
		args = append(args, *input.RouteID)
		idx++
	}
	if input.BusID != nil {
		sets = append(sets, fmt.Sprintf("bus_id=$%d", idx))
		args = append(args, *input.BusID)
		idx++
	}
	if input.DriverID != nil {
		if *input.DriverID == "" {
			sets = append(sets, fmt.Sprintf("driver_id=$%d", idx))
			args = append(args, nil)
		} else {
			sets = append(sets, fmt.Sprintf("driver_id=$%d", idx))
			args = append(args, *input.DriverID)
		}
		idx++
	}
	if input.FareID != nil {
		sets = append(sets, fmt.Sprintf("fare_id=$%d", idx))
		args = append(args, nullableUUID(input.FareID))
		idx++
	}
	if input.RequestID != nil {
		sets = append(sets, fmt.Sprintf("request_id=$%d", idx))
		args = append(args, nullableUUID(input.RequestID))
		idx++
	}
	if input.DepartureAt != nil {
		sets = append(sets, fmt.Sprintf("departure_at=$%d", idx))
		args = append(args, *input.DepartureAt)
		idx++
	}
	if input.ArrivalAt != nil {
		sets = append(sets, fmt.Sprintf("arrival_at=$%d", idx))
		args = append(args, *input.ArrivalAt)
		idx++
	}
	if input.Status != nil {
		sets = append(sets, fmt.Sprintf("status=$%d", idx))
		args = append(args, *input.Status)
		idx++
	}
	if input.OperationalStatus != nil {
		sets = append(sets, fmt.Sprintf("operational_status=$%d", idx))
		args = append(args, *input.OperationalStatus)
		idx++
	}
	if input.EstimatedKM != nil {
		sets = append(sets, fmt.Sprintf("estimated_km=$%d", idx))
		args = append(args, *input.EstimatedKM)
		idx++
	}
	if input.PairTripID != nil {
		sets = append(sets, fmt.Sprintf("pair_trip_id=$%d", idx))
		args = append(args, nullableUUID(input.PairTripID))
		idx++
	}
	if input.Notes != nil {
		sets = append(sets, fmt.Sprintf("notes=$%d", idx))
		args = append(args, *input.Notes)
		idx++
	}

	if len(sets) == 0 {
		return r.Get(ctx, id)
	}

	sets = append(sets, "updated_at=now()")

	args = append(args, id)
	query := fmt.Sprintf(`update trips set %s where id=$%d returning id, route_id, bus_id, driver_id, fare_id, request_id, departure_at, arrival_at, status,
    operational_status, estimated_km, dispatch_validated_at, dispatch_validated_by, pair_trip_id, notes,
    created_at, updated_at`, strings.Join(sets, ", "), idx)

	var item Trip
	row := r.pool.QueryRow(ctx, query, args...)
	if err := scanTrip(row, &item); err != nil {
		return item, err
	}
	return item, nil
}

func (r *Repository) ListSeats(ctx context.Context, tripID uuid.UUID, boardStopID *uuid.UUID, alightStopID *uuid.UUID) ([]TripSeat, error) {
	if boardStopID == nil || alightStopID == nil {
		rows, err := r.pool.Query(ctx, `
    select s.id, s.seat_number, s.is_active,
      exists(
        select 1
        from booking_passengers bp
        join bookings b on b.id = bp.booking_id
        where bp.trip_id=$1
          and bp.seat_id=s.id
          and bp.is_active = true
          and b.status not in ('CANCELLED','EXPIRED')
      ) as is_taken
    from trips t
    join bus_seats s on s.bus_id = t.bus_id
    where t.id = $1
    order by s.seat_number asc
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

	var boardOrder int
	if err := r.pool.QueryRow(ctx, `select stop_order from trip_stops where trip_id=$1 and id=$2`, tripID, *boardStopID).Scan(&boardOrder); err != nil {
		return nil, err
	}
	var alightOrder int
	if err := r.pool.QueryRow(ctx, `select stop_order from trip_stops where trip_id=$1 and id=$2`, tripID, *alightStopID).Scan(&alightOrder); err != nil {
		return nil, err
	}

	rows, err := r.pool.Query(ctx, `
    select s.id, s.seat_number, s.is_active,
      exists(
        select 1
        from booking_passengers bp
        join bookings b on b.id = bp.booking_id
        where bp.trip_id=$1
          and bp.seat_id=s.id
          and bp.is_active = true
          and b.status not in ('CANCELLED','EXPIRED')
          and bp.board_stop_order < $3
          and bp.alight_stop_order > $2
      ) as is_taken
    from trips t
    join bus_seats s on s.bus_id = t.bus_id
    where t.id = $1
    order by s.seat_number asc
  `, tripID, boardOrder, alightOrder)
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

func (r *Repository) ListStops(ctx context.Context, tripID uuid.UUID) ([]TripStop, error) {
	rows, err := r.pool.Query(ctx, `
    select ts.id, ts.trip_id, ts.route_stop_id, rs.city, ts.stop_order, ts.leg_distance_km, ts.cumulative_distance_km,
      ts.arrive_at, ts.depart_at, ts.created_at, ts.updated_at
    from trip_stops ts
    join route_stops rs on rs.id = ts.route_stop_id
    where ts.trip_id = $1
    order by ts.stop_order asc
  `, tripID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []TripStop{}
	for rows.Next() {
		var item TripStop
		if err := rows.Scan(&item.ID, &item.TripID, &item.RouteStopID, &item.City, &item.StopOrder, &item.LegDistanceKM, &item.CumulativeDistanceKM, &item.ArriveAt, &item.DepartAt, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) CreateStop(ctx context.Context, tripID uuid.UUID, input CreateTripStopInput) (TripStop, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return TripStop{}, err
	}
	defer tx.Rollback(ctx)

	var stopOrder int
	row := tx.QueryRow(ctx, `
    select rs.stop_order
    from trips t
    join route_stops rs on rs.route_id = t.route_id
    where t.id = $1 and rs.id = $2
  `, tripID, input.RouteStopID)
	if err := row.Scan(&stopOrder); err != nil {
		return TripStop{}, err
	}

	legDistanceKM := 0.0
	if input.LegDistanceKM != nil {
		legDistanceKM = *input.LegDistanceKM
	}
	cumulativeDistanceKM := 0.0
	if input.CumulativeDistanceKM != nil {
		cumulativeDistanceKM = *input.CumulativeDistanceKM
	}

	var item TripStop
	row = tx.QueryRow(ctx, `
    with inserted as (
      insert into trip_stops (trip_id, route_stop_id, stop_order, leg_distance_km, cumulative_distance_km, arrive_at, depart_at, updated_at)
      values ($1,$2,$3,$4,$5,$6,$7,now())
      returning id, trip_id, route_stop_id, stop_order, leg_distance_km, cumulative_distance_km, arrive_at, depart_at, created_at, updated_at
    )
    select inserted.id, inserted.trip_id, inserted.route_stop_id, rs.city, inserted.stop_order,
      inserted.leg_distance_km, inserted.cumulative_distance_km, inserted.arrive_at, inserted.depart_at,
      inserted.created_at, inserted.updated_at
    from inserted
    join route_stops rs on rs.id = inserted.route_stop_id
  `, tripID, input.RouteStopID, stopOrder, legDistanceKM, cumulativeDistanceKM, input.ArriveAt, input.DepartAt)
	if err := rowscanTripStop(row, &item); err != nil {
		return TripStop{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return TripStop{}, err
	}

	return item, nil
}

func (r *Repository) EnsureTripStopsFromRoute(ctx context.Context, tripID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
    with base as (
      select t.id as trip_id, rs.id as route_stop_id, rs.stop_order,
        case
          when rs.eta_offset_minutes is not null then t.departure_at + make_interval(mins => rs.eta_offset_minutes)
          else null
        end as arrive_at
      from trips t
      join route_stops rs on rs.route_id = t.route_id
      where t.id = $1
    )
    insert into trip_stops (trip_id, route_stop_id, stop_order, arrive_at, depart_at, leg_distance_km, cumulative_distance_km, updated_at)
    select base.trip_id, base.route_stop_id, base.stop_order, base.arrive_at, null, 0, 0, now()
    from base
    where not exists (
      select 1 from trip_stops ts where ts.trip_id = base.trip_id
    )
    on conflict (trip_id, route_stop_id) do nothing
  `, tripID)
	return err
}

func (r *Repository) ensureTripStopsFromRouteTx(ctx context.Context, tx pgx.Tx, tripID uuid.UUID) error {
	_, err := tx.Exec(ctx, `
    with base as (
      select t.id as trip_id, rs.id as route_stop_id, rs.stop_order,
        case
          when rs.eta_offset_minutes is not null then t.departure_at + make_interval(mins => rs.eta_offset_minutes)
          else null
        end as arrive_at
      from trips t
      join route_stops rs on rs.route_id = t.route_id
      where t.id = $1
    )
    insert into trip_stops (trip_id, route_stop_id, stop_order, arrive_at, depart_at, leg_distance_km, cumulative_distance_km, updated_at)
    select base.trip_id, base.route_stop_id, base.stop_order, base.arrive_at, null, 0, 0, now()
    from base
    on conflict (trip_id, route_stop_id) do nothing
	`, tripID)
	return err
}

func (r *Repository) ValidateRouteReadiness(ctx context.Context, routeID string) ([]string, error) {
	type routeValidation struct {
		IsActive        bool
		OriginCity      string
		DestinationCity string
		StopCount       int
		MinStopOrder    int
		MaxStopOrder    int
		FirstCity       string
		LastCity        string
		HasNegativeETA  bool
		HasETABacktrack bool
	}

	var validation routeValidation
	row := r.pool.QueryRow(ctx, `
    select
      r.is_active,
      r.origin_city,
      r.destination_city,
      count(rs.id)::int as stop_count,
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
    from routes r
    left join route_stops rs on rs.route_id = r.id
    where r.id = $1
    group by r.id, r.is_active, r.origin_city, r.destination_city
  `, routeID)

	if err := row.Scan(
		&validation.IsActive,
		&validation.OriginCity,
		&validation.DestinationCity,
		&validation.StopCount,
		&validation.MinStopOrder,
		&validation.MaxStopOrder,
		&validation.FirstCity,
		&validation.LastCity,
		&validation.HasNegativeETA,
		&validation.HasETABacktrack,
	); err != nil {
		if err == pgx.ErrNoRows {
			return []string{"route not found"}, nil
		}
		return nil, err
	}

	missing := []string{}
	if !validation.IsActive {
		missing = append(missing, "route must be active")
	}
	if validation.StopCount < 2 {
		missing = append(missing, "at least two stops are required")
	}
	if validation.StopCount > 0 {
		if validation.MinStopOrder != 1 {
			missing = append(missing, "stop_order must start at 1")
		}
		if (validation.MaxStopOrder - validation.MinStopOrder + 1) != validation.StopCount {
			missing = append(missing, "stop_order must be sequential without gaps")
		}
		if !strings.EqualFold(strings.TrimSpace(validation.OriginCity), strings.TrimSpace(validation.FirstCity)) {
			missing = append(missing, "first stop city must match origin_city")
		}
		if !strings.EqualFold(strings.TrimSpace(validation.DestinationCity), strings.TrimSpace(validation.LastCity)) {
			missing = append(missing, "last stop city must match destination_city")
		}
	}
	if validation.HasNegativeETA {
		missing = append(missing, "eta_offset_minutes must be >= 0")
	}
	if validation.HasETABacktrack {
		missing = append(missing, "eta_offset_minutes must be non-decreasing")
	}

	return missing, nil
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

func nullableUUID(v *string) interface{} {
	if v == nil || *v == "" {
		return nil
	}
	return *v
}

func IsNotFound(err error) bool {
	return err == pgx.ErrNoRows
}
