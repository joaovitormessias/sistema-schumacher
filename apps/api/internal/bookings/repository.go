package bookings

import (
  "context"
  "errors"

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
    select b.id, b.trip_id, b.status, b.total_amount, b.deposit_amount, b.remainder_amount,
           bp.name, bp.phone, bp.email, bs.seat_number, b.created_at
    from bookings b
    join booking_passengers bp on bp.booking_id = b.id
    join bus_seats bs on bs.id = bp.seat_id
    where bp.is_active = true
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
    if err := rows.Scan(&item.ID, &item.TripID, &item.Status, &item.TotalAmount, &item.DepositAmount, &item.RemainderAmount,
      &item.PassengerName, &item.PassengerPhone, &item.PassengerEmail, &item.SeatNumber, &item.CreatedAt); err != nil {
      return nil, err
    }
    items = append(items, item)
  }
  return items, rows.Err()
}

func (r *Repository) Create(ctx context.Context, input CreateBookingData) (BookingDetails, error) {
  source := "WHATSAPP"
  if input.Source != nil {
    source = *input.Source
  }

  tx, err := r.pool.Begin(ctx)
  if err != nil {
    return BookingDetails{}, err
  }
  defer tx.Rollback(ctx)

  var seatValid bool
   if err := tx.QueryRow(ctx,
     `select exists(
        select 1
        from trips t
        join bus_seats s on s.bus_id = t.bus_id
        where t.id = $1 and s.id = $2 and s.is_active = true
      )`, input.TripID, input.SeatID,
   ).Scan(&seatValid); err != nil {
    return BookingDetails{}, err
  }
  if !seatValid {
    return BookingDetails{}, ErrSeatNotInTrip
  }

  var booking Booking
  row := tx.QueryRow(ctx,
    `insert into bookings (trip_id, status, source, total_amount, deposit_amount, remainder_amount)
     values ($1, 'PENDING', $2, $3, $4, $5)
     returning id, trip_id, status, source, total_amount, deposit_amount, remainder_amount, expires_at, created_at`,
    input.TripID, source, input.TotalAmount, input.DepositAmount, input.RemainderAmount,
  )
  if err := row.Scan(&booking.ID, &booking.TripID, &booking.Status, &booking.Source, &booking.TotalAmount, &booking.DepositAmount, &booking.RemainderAmount, &booking.ExpiresAt, &booking.CreatedAt); err != nil {
    return BookingDetails{}, err
  }

  var passenger BookingPassenger
  row = tx.QueryRow(ctx,
    `insert into booking_passengers (
        booking_id, trip_id, name, document, phone, email, seat_id,
        board_stop_id, alight_stop_id, board_stop_order, alight_stop_order,
        fare_mode, fare_amount_calc, fare_amount_final, fare_snapshot
     )
     values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
     returning id, booking_id, trip_id, name, document, phone, email, seat_id,
       board_stop_id, alight_stop_id, board_stop_order, alight_stop_order,
       fare_mode, fare_amount_calc, fare_amount_final, status, created_at`,
    booking.ID, input.TripID, input.Passenger.Name, input.Passenger.Document, input.Passenger.Phone, input.Passenger.Email, input.SeatID,
    input.BoardStopID, input.AlightStopID, input.BoardStopOrder, input.AlightStopOrder,
    input.FareMode, input.FareAmountCalc, input.FareAmountFinal, input.FareSnapshot,
  )
  if err := row.Scan(&passenger.ID, &passenger.BookingID, &passenger.TripID, &passenger.Name, &passenger.Document, &passenger.Phone, &passenger.Email, &passenger.SeatID,
    &passenger.BoardStopID, &passenger.AlightStopID, &passenger.BoardStopOrder, &passenger.AlightStopOrder,
    &passenger.FareMode, &passenger.FareAmountCalc, &passenger.FareAmountFinal, &passenger.Status, &passenger.CreatedAt); err != nil {
    return BookingDetails{}, err
  }

  if err := tx.Commit(ctx); err != nil {
    return BookingDetails{}, err
  }

  return BookingDetails{Booking: booking, Passenger: passenger}, nil
}

func (r *Repository) Get(ctx context.Context, id string) (BookingDetails, error) {
  var booking Booking
  row := r.pool.QueryRow(ctx,
    `select id, trip_id, status, source, total_amount, deposit_amount, remainder_amount, expires_at, created_at
     from bookings where id=$1`, id)
  if err := row.Scan(&booking.ID, &booking.TripID, &booking.Status, &booking.Source, &booking.TotalAmount, &booking.DepositAmount, &booking.RemainderAmount, &booking.ExpiresAt, &booking.CreatedAt); err != nil {
    return BookingDetails{}, err
  }

  var passenger BookingPassenger
  row = r.pool.QueryRow(ctx,
    `select id, booking_id, trip_id, name, document, phone, email, seat_id,
            board_stop_id, alight_stop_id, board_stop_order, alight_stop_order,
            fare_mode, fare_amount_calc, fare_amount_final, status, created_at
     from booking_passengers where booking_id=$1 limit 1`, id)
  if err := row.Scan(&passenger.ID, &passenger.BookingID, &passenger.TripID, &passenger.Name, &passenger.Document, &passenger.Phone, &passenger.Email, &passenger.SeatID,
    &passenger.BoardStopID, &passenger.AlightStopID, &passenger.BoardStopOrder, &passenger.AlightStopOrder,
    &passenger.FareMode, &passenger.FareAmountCalc, &passenger.FareAmountFinal, &passenger.Status, &passenger.CreatedAt); err != nil {
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
  row := tx.QueryRow(ctx,
    `update bookings set status=$1, updated_at=now() where id=$2
     returning id, trip_id, status, source, total_amount, deposit_amount, remainder_amount, expires_at, created_at`,
    status, id,
  )
  if err := row.Scan(&booking.ID, &booking.TripID, &booking.Status, &booking.Source, &booking.TotalAmount, &booking.DepositAmount, &booking.RemainderAmount, &booking.ExpiresAt, &booking.CreatedAt); err != nil {
    return BookingDetails{}, err
  }

  if status == "CANCELLED" || status == "EXPIRED" {
    if _, err := tx.Exec(ctx, `update booking_passengers set is_active=false, status='CANCELLED' where booking_id=$1`, id); err != nil {
      return BookingDetails{}, err
    }
  }

  var passenger BookingPassenger
  row = tx.QueryRow(ctx,
    `select id, booking_id, trip_id, name, document, phone, email, seat_id,
            board_stop_id, alight_stop_id, board_stop_order, alight_stop_order,
            fare_mode, fare_amount_calc, fare_amount_final, status, created_at
     from booking_passengers where booking_id=$1 limit 1`, id)
  if err := row.Scan(&passenger.ID, &passenger.BookingID, &passenger.TripID, &passenger.Name, &passenger.Document, &passenger.Phone, &passenger.Email, &passenger.SeatID,
    &passenger.BoardStopID, &passenger.AlightStopID, &passenger.BoardStopOrder, &passenger.AlightStopOrder,
    &passenger.FareMode, &passenger.FareAmountCalc, &passenger.FareAmountFinal, &passenger.Status, &passenger.CreatedAt); err != nil {
    return BookingDetails{}, err
  }

  if err := tx.Commit(ctx); err != nil {
    return BookingDetails{}, err
  }

  return BookingDetails{Booking: booking, Passenger: passenger}, nil
}

func IsUniqueViolation(err error) bool {
  pgErr, ok := err.(*pgconn.PgError)
  if !ok {
    return false
  }
  return pgErr.Code == "23505" || pgErr.Code == "23P01"
}
