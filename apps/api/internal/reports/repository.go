package reports

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) ListPassengers(ctx context.Context, filter PassengerReportFilter) ([]PassengerReportRow, error) {
	query := `
		select
			p.passenger_id,
			p.booking_id,
			coalesce(b.reservation_code, '') as reservation_code,
			p.trip_id,
			to_char(t.trip_date, 'YYYY-MM-DD') as trip_date,
			p.full_name,
			coalesce(p.document, ''),
			coalesce(p.phone, ''),
			coalesce(p.email, ''),
			coalesce(origin.display_name, ''),
			coalesce(destination.display_name, ''),
			coalesce(p.seat_number, ''),
			b.status,
			p.status,
			coalesce(pd.amount_total, 0)::numeric as total_amount,
			coalesce(pd.amount_paid, 0)::numeric as deposit_amount,
			coalesce(pd.amount_due, greatest(coalesce(pd.amount_total, 0) - coalesce(pd.amount_paid, 0), 0))::numeric as remainder_amount,
			coalesce(sum(pay.amount) filter (where pay.status = 'PAID'), 0)::numeric as amount_paid,
			p.created_at
		from passengers p
		join bookings b on b.booking_id = p.booking_id
		join trips t on t.trip_id = p.trip_id
		left join booking_payment_details pd on pd.booking_id = b.booking_id
		left join payments pay on pay.booking_id = b.booking_id
		left join stops origin on origin.stop_id = p.origin_stop_id
		left join stops destination on destination.stop_id = p.destination_stop_id`

	args := []interface{}{}
	clauses := []string{}

	if filter.TripID != "" {
		args = append(args, filter.TripID)
		clauses = append(clauses, fmt.Sprintf("p.trip_id = $%d", len(args)))
	}
	if filter.TripDate != "" {
		args = append(args, filter.TripDate)
		clauses = append(clauses, fmt.Sprintf("t.trip_date = $%d::date", len(args)))
	}
	if filter.BookingID != "" {
		args = append(args, filter.BookingID)
		clauses = append(clauses, fmt.Sprintf("p.booking_id = $%d", len(args)))
	}
	if filter.ReservationCode != "" {
		args = append(args, filter.ReservationCode)
		clauses = append(clauses, fmt.Sprintf("b.reservation_code = $%d", len(args)))
	}
	if !filter.IncludeCanceled {
		clauses = append(clauses, "upper(coalesce(b.status, '')) not in ('CANCELLED', 'EXPIRED')")
		clauses = append(clauses, "upper(coalesce(p.status, '')) <> 'CANCELLED'")
	}

	if len(clauses) > 0 {
		query += "\nwhere " + strings.Join(clauses, " and ")
	}

	query += `
		group by
			p.passenger_id,
			p.booking_id,
			b.reservation_code,
			p.trip_id,
			t.trip_date,
			p.full_name,
			p.document,
			p.phone,
			p.email,
			origin.display_name,
			destination.display_name,
			p.seat_number,
			b.status,
			p.status,
			pd.amount_total,
			pd.amount_paid,
			pd.amount_due,
			p.created_at
		order by t.trip_date asc, p.trip_id asc, p.created_at asc`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []PassengerReportRow{}
	for rows.Next() {
		var row PassengerReportRow
		if err := rows.Scan(
			&row.PassengerID,
			&row.BookingID,
			&row.ReservationCode,
			&row.TripID,
			&row.TripDate,
			&row.Name,
			&row.Document,
			&row.Phone,
			&row.Email,
			&row.Origin,
			&row.Destination,
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
