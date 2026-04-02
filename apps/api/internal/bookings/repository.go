package bookings

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrSeatNotInTrip = errors.New("seat does not belong to trip bus")

// Repository handles booking persistence.
type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) List(ctx context.Context, filter ListFilter) ([]BookingListItem, error) {
	limit := filter.Limit
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	rows, err := r.pool.Query(ctx, `
    select
      b.booking_id as id,
      b.trip_id,
      b.status,
      coalesce(pd.amount_total, 0)::numeric as total_amount,
      coalesce(pd.amount_paid, 0)::numeric as deposit_amount,
      coalesce(pd.amount_due, greatest(coalesce(pd.amount_total, 0) - coalesce(pd.amount_paid, 0), 0))::numeric as remainder_amount,
      coalesce(p.full_name, b.customer_name) as passenger_name,
      coalesce(p.phone, b.customer_phone) as passenger_phone,
      ''::text as passenger_email,
      case when coalesce(p.seat_number, '') ~ '^[0-9]+$' then p.seat_number::int else 0 end as seat_number,
      b.created_at
    from bookings b
    left join booking_payment_details pd on pd.booking_id = b.booking_id
    left join lateral (
      select full_name, phone, seat_number
      from passengers p
      where p.booking_id = b.booking_id
      order by p.created_at asc
      limit 1
    ) p on true
    order by b.created_at desc
    limit $1 offset $2
  `, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []BookingListItem{}
	for rows.Next() {
		var item BookingListItem
		if err := rows.Scan(
			&item.ID,
			&item.TripID,
			&item.Status,
			&item.TotalAmount,
			&item.DepositAmount,
			&item.RemainderAmount,
			&item.PassengerName,
			&item.PassengerPhone,
			&item.PassengerEmail,
			&item.SeatNumber,
			&item.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) Create(ctx context.Context, input CreateBookingData) (BookingDetails, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return BookingDetails{}, err
	}
	defer tx.Rollback(ctx)

	var tripExists bool
	var seatsTotal int
	if err := tx.QueryRow(ctx, `select true, greatest(coalesce(seats_total, 0), 0) from trips where trip_id = $1`, input.TripID).Scan(&tripExists, &seatsTotal); err != nil {
		return BookingDetails{}, err
	}
	if !tripExists {
		return BookingDetails{}, fmt.Errorf("trip not found")
	}

	seatID := strings.TrimSpace(input.SeatID)
	if seatID == "" {
		return BookingDetails{}, ErrSeatNotInTrip
	}
	if seatsTotal > 0 {
		var seatNum int
		if _, err := fmt.Sscanf(seatID, "%d", &seatNum); err != nil || seatNum <= 0 || seatNum > seatsTotal {
			return BookingDetails{}, ErrSeatNotInTrip
		}
	}

	var booking Booking
	row := tx.QueryRow(ctx, `
    with generated as (
      select 'BK-' || upper(replace(gen_random_uuid()::text, '-', '')) as booking_id
    )
    insert into bookings (
      booking_id, trip_id, origin_stop_id, destination_stop_id,
      passenger_qty, customer_name, customer_phone, status, reservation_code, row_type, idempotency_key,
      created_at, updated_at
    )
    select
      g.booking_id,
      $1, $2, $3,
      1, $4, $5, 'PENDING',
      right(g.booking_id, 8), 'booking', nullif($6, ''),
      now(), now()
    from generated g
    returning booking_id, trip_id, status, created_at
  `, input.TripID, input.BoardStopID, input.AlightStopID, input.Passenger.Name, input.Passenger.Phone, input.SeatID)
	if err := row.Scan(&booking.ID, &booking.TripID, &booking.Status, &booking.CreatedAt); err != nil {
		return BookingDetails{}, err
	}
	booking.Source = "WHATSAPP"
	booking.TotalAmount = input.TotalAmount
	booking.DepositAmount = input.DepositAmount
	booking.RemainderAmount = input.RemainderAmount

	var passenger BookingPassenger
	row = tx.QueryRow(ctx, `
    with generated as (
      select 'PS-' || upper(replace(gen_random_uuid()::text, '-', '')) as passenger_id
    )
    insert into passengers (
      passenger_id, booking_id, trip_id, full_name, document, seat_number,
      origin_stop_id, destination_stop_id, phone, notes, status, row_type, created_at, updated_at
    )
    select
      g.passenger_id,
      $1, $2, $3, nullif($4, ''), $5,
      $6, $7, nullif($8, ''), null, 'RESERVED', 'passenger', now(), now()
    from generated g
    returning passenger_id, booking_id, trip_id, full_name, coalesce(document, ''), coalesce(phone, ''), coalesce(seat_number, ''), status, created_at
  `, booking.ID, input.TripID, input.Passenger.Name, input.Passenger.Document, seatID, input.BoardStopID, input.AlightStopID, input.Passenger.Phone)
	var seatNumber string
	if err := row.Scan(&passenger.ID, &passenger.BookingID, &passenger.TripID, &passenger.Name, &passenger.Document, &passenger.Phone, &seatNumber, &passenger.Status, &passenger.CreatedAt); err != nil {
		return BookingDetails{}, err
	}
	passenger.SeatID = seatNumber
	passenger.BoardStopID = input.BoardStopID
	passenger.AlightStopID = input.AlightStopID
	passenger.FareMode = input.FareMode
	passenger.FareAmountCalc = input.FareAmountCalc
	passenger.FareAmountFinal = input.FareAmountFinal

	if _, err := tx.Exec(ctx, `
    insert into booking_payment_details (
      booking_id, payment_type, payment_status, payment_method,
      amount_total, amount_paid
    ) values ($1, 'PARTIAL', 'PENDING', 'PIX', $2, $3)
    on conflict (booking_id) do update
      set amount_total = excluded.amount_total,
          amount_paid = excluded.amount_paid
  `, booking.ID, input.TotalAmount, input.DepositAmount); err != nil {
		return BookingDetails{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return BookingDetails{}, err
	}

	return BookingDetails{Booking: booking, Passenger: passenger}, nil
}

func (r *Repository) Get(ctx context.Context, id string) (BookingDetails, error) {
	var booking Booking
	row := r.pool.QueryRow(ctx, `
    select
      b.booking_id, b.trip_id, b.status,
      coalesce(pd.amount_total, 0)::numeric as total_amount,
      coalesce(pd.amount_paid, 0)::numeric as deposit_amount,
      coalesce(pd.amount_due, greatest(coalesce(pd.amount_total, 0) - coalesce(pd.amount_paid, 0), 0))::numeric as remainder_amount,
      b.reserved_until,
      b.created_at
    from bookings b
    left join booking_payment_details pd on pd.booking_id = b.booking_id
    where b.booking_id=$1
  `, id)
	if err := row.Scan(&booking.ID, &booking.TripID, &booking.Status, &booking.TotalAmount, &booking.DepositAmount, &booking.RemainderAmount, &booking.ExpiresAt, &booking.CreatedAt); err != nil {
		return BookingDetails{}, err
	}
	booking.Source = "WHATSAPP"

	var passenger BookingPassenger
	row = r.pool.QueryRow(ctx, `
    select
      p.passenger_id, p.booking_id, p.trip_id, p.full_name, coalesce(p.document, ''), coalesce(p.phone, ''),
      ''::text as email,
      coalesce(p.seat_number, '') as seat_id,
      coalesce(p.origin_stop_id, ''),
      coalesce(p.destination_stop_id, ''),
      0::int as board_stop_order,
      0::int as alight_stop_order,
      'AUTO'::text as fare_mode,
      coalesce(pd.amount_total, 0)::numeric as fare_amount_calc,
      coalesce(pd.amount_total, 0)::numeric as fare_amount_final,
      p.status,
      p.created_at
    from passengers p
    left join booking_payment_details pd on pd.booking_id = p.booking_id
    where p.booking_id=$1
    order by p.created_at asc
    limit 1`, id)
	if err := row.Scan(
		&passenger.ID, &passenger.BookingID, &passenger.TripID, &passenger.Name, &passenger.Document, &passenger.Phone, &passenger.Email, &passenger.SeatID,
		&passenger.BoardStopID, &passenger.AlightStopID, &passenger.BoardStopOrder, &passenger.AlightStopOrder,
		&passenger.FareMode, &passenger.FareAmountCalc, &passenger.FareAmountFinal, &passenger.Status, &passenger.CreatedAt,
	); err != nil {
		return BookingDetails{}, err
	}

	return BookingDetails{Booking: booking, Passenger: passenger}, nil
}

func (r *Repository) UpdateStatus(ctx context.Context, id string, status string) (BookingDetails, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return BookingDetails{}, err
	}
	defer tx.Rollback(ctx)

	var booking Booking
	row := tx.QueryRow(ctx, `
    update bookings
    set status=$1, updated_at=now()
    where booking_id=$2
    returning booking_id, trip_id, status, created_at
  `, status, id)
	if err := row.Scan(&booking.ID, &booking.TripID, &booking.Status, &booking.CreatedAt); err != nil {
		return BookingDetails{}, err
	}
	booking.Source = "WHATSAPP"

	if status == "CANCELLED" || status == "EXPIRED" {
		if _, err := tx.Exec(ctx, `update passengers set status='CANCELLED', updated_at=now() where booking_id=$1`, id); err != nil {
			return BookingDetails{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return BookingDetails{}, err
	}

	return r.Get(ctx, id)
}

func IsUniqueViolation(err error) bool {
	pgErr, ok := err.(*pgconn.PgError)
	if !ok {
		return false
	}
	return pgErr.Code == "23505" || pgErr.Code == "23P01"
}

