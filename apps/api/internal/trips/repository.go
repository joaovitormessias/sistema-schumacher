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

	query := `select id, route_id, bus_id, driver_id, fare_id, departure_at, arrival_at, status, pair_trip_id, notes, created_at
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
    )`, len(args), len(args), len(args), len(args), len(args)))
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
		if err := rows.Scan(&item.ID, &item.RouteID, &item.BusID, &item.DriverID, &item.FareID, &item.DepartureAt, &item.ArrivalAt, &item.Status, &item.PairTripID, &item.Notes, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) Get(ctx context.Context, id uuid.UUID) (Trip, error) {
	var item Trip
	row := r.pool.QueryRow(ctx, `select id, route_id, bus_id, driver_id, fare_id, departure_at, arrival_at, status, pair_trip_id, notes, created_at from trips where id=$1`, id)
	if err := row.Scan(&item.ID, &item.RouteID, &item.BusID, &item.DriverID, &item.FareID, &item.DepartureAt, &item.ArrivalAt, &item.Status, &item.PairTripID, &item.Notes, &item.CreatedAt); err != nil {
		return item, err
	}
	return item, nil
}

func (r *Repository) Create(ctx context.Context, input CreateTripInput) (Trip, error) {
	status := "SCHEDULED"
	if input.Status != nil {
		status = *input.Status
	}

	var item Trip
	row := r.pool.QueryRow(ctx,
		`insert into trips (route_id, bus_id, driver_id, fare_id, departure_at, arrival_at, status, pair_trip_id, notes)
     values ($1,$2,$3,$4,$5,$6,$7,$8,$9)
     returning id, route_id, bus_id, driver_id, fare_id, departure_at, arrival_at, status, pair_trip_id, notes, created_at`,
		input.RouteID, input.BusID, input.DriverID, input.FareID, input.DepartureAt, input.ArrivalAt, status, input.PairTripID, input.Notes,
	)
	if err := row.Scan(&item.ID, &item.RouteID, &item.BusID, &item.DriverID, &item.FareID, &item.DepartureAt, &item.ArrivalAt, &item.Status, &item.PairTripID, &item.Notes, &item.CreatedAt); err != nil {
		return item, err
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
		sets = append(sets, fmt.Sprintf("driver_id=$%d", idx))
		args = append(args, *input.DriverID)
		idx++
	}
	if input.FareID != nil {
		sets = append(sets, fmt.Sprintf("fare_id=$%d", idx))
		args = append(args, *input.FareID)
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
	if input.PairTripID != nil {
		sets = append(sets, fmt.Sprintf("pair_trip_id=$%d", idx))
		args = append(args, *input.PairTripID)
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

	args = append(args, id)
	query := fmt.Sprintf(`update trips set %s where id=$%d returning id, route_id, bus_id, driver_id, fare_id, departure_at, arrival_at, status, pair_trip_id, notes, created_at`, strings.Join(sets, ", "), idx)

	var item Trip
	row := r.pool.QueryRow(ctx, query, args...)
	if err := row.Scan(&item.ID, &item.RouteID, &item.BusID, &item.DriverID, &item.FareID, &item.DepartureAt, &item.ArrivalAt, &item.Status, &item.PairTripID, &item.Notes, &item.CreatedAt); err != nil {
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
    select ts.id, ts.trip_id, ts.route_stop_id, rs.city, ts.stop_order, ts.arrive_at, ts.depart_at, ts.created_at
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
		if err := rows.Scan(&item.ID, &item.TripID, &item.RouteStopID, &item.City, &item.StopOrder, &item.ArriveAt, &item.DepartAt, &item.CreatedAt); err != nil {
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

	var item TripStop
	row = tx.QueryRow(ctx, `
    with inserted as (
      insert into trip_stops (trip_id, route_stop_id, stop_order, arrive_at, depart_at)
      values ($1,$2,$3,$4,$5)
      returning id, trip_id, route_stop_id, stop_order, arrive_at, depart_at, created_at
    )
    select inserted.id, inserted.trip_id, inserted.route_stop_id, rs.city, inserted.stop_order, inserted.arrive_at, inserted.depart_at, inserted.created_at
    from inserted
    join route_stops rs on rs.id = inserted.route_stop_id
  `, tripID, input.RouteStopID, stopOrder, input.ArriveAt, input.DepartAt)
	if err := row.Scan(&item.ID, &item.TripID, &item.RouteStopID, &item.City, &item.StopOrder, &item.ArriveAt, &item.DepartAt, &item.CreatedAt); err != nil {
		return TripStop{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return TripStop{}, err
	}

	return item, nil
}

func IsNotFound(err error) bool {
	return err == pgx.ErrNoRows
}
