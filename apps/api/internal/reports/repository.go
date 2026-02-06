package reports

import (
  "context"

  "github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
  pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
  return &Repository{pool: pool}
}

func (r *Repository) ListPassengers(ctx context.Context, tripID string) ([]PassengerReportRow, error) {
  rows, err := r.pool.Query(ctx, `
    select
      bp.id,
      bp.name,
      bp.document,
      bp.phone,
      bp.email,
      bs.seat_number,
      b.status,
      bp.status,
      b.total_amount,
      b.deposit_amount,
      b.remainder_amount,
      coalesce(sum(p.amount) filter (where p.status = 'PAID'), 0) as amount_paid,
      bp.created_at
    from booking_passengers bp
    join bookings b on b.id = bp.booking_id
    join bus_seats bs on bs.id = bp.seat_id
    left join payments p on p.booking_id = b.id
    where bp.trip_id = $1
      and bp.is_active = true
      and b.status not in ('CANCELLED','EXPIRED')
    group by bp.id, bp.name, bp.document, bp.phone, bp.email, bs.seat_number, b.status, bp.status, b.total_amount, b.deposit_amount, b.remainder_amount, bp.created_at
    order by bs.seat_number asc
  `, tripID)
  if err != nil {
    return nil, err
  }
  defer rows.Close()

  items := []PassengerReportRow{}
  for rows.Next() {
    var row PassengerReportRow
    if err := rows.Scan(
      &row.PassengerID,
      &row.Name,
      &row.Document,
      &row.Phone,
      &row.Email,
      &row.SeatNumber,
      &row.BookingStatus,
      &row.PassengerStatus,
      &row.TotalAmount,
      &row.DepositAmount,
      &row.RemainderAmount,
      &row.AmountPaid,
      &row.CreatedAt,
    ); err != nil {
      return nil, err
    }
    row.PaymentStage = paymentStage(row.AmountPaid, row.TotalAmount, row.DepositAmount)
    items = append(items, row)
  }
  return items, rows.Err()
}

func paymentStage(paid, total, deposit float64) string {
  if total > 0 && paid >= total {
    return "PAID"
  }
  if deposit > 0 && paid >= deposit {
    return "DEPOSIT"
  }
  return "PENDING"
}
