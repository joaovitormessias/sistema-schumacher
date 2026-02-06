package advance_returns

import (
  "context"
  "fmt"
  "strings"
  "time"

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

func (r *Repository) List(ctx context.Context, filter ListFilter) ([]AdvanceReturn, error) {
  query := `select id, trip_advance_id, trip_settlement_id, amount, return_date, payment_method, received_by, notes, created_at from advance_returns`
  args := []interface{}{}
  clauses := []string{}

  if filter.TripAdvanceID != "" {
    args = append(args, filter.TripAdvanceID)
    clauses = append(clauses, fmt.Sprintf("trip_advance_id=$%d", len(args)))
  }
  if filter.TripSettlementID != "" {
    args = append(args, filter.TripSettlementID)
    clauses = append(clauses, fmt.Sprintf("trip_settlement_id=$%d", len(args)))
  }

  if len(clauses) > 0 {
    query += " where " + strings.Join(clauses, " and ")
  }

  query += " order by return_date desc"

  limit := filter.Limit
  if limit <= 0 || limit > 500 {
    limit = 200
  }
  args = append(args, limit)
  query += fmt.Sprintf(" limit $%d", len(args))

  if filter.Offset > 0 {
    args = append(args, filter.Offset)
    query += fmt.Sprintf(" offset $%d", len(args))
  }

  rows, err := r.pool.Query(ctx, query, args...)
  if err != nil {
    return nil, err
  }
  defer rows.Close()

  items := []AdvanceReturn{}
  for rows.Next() {
    var item AdvanceReturn
    if err := rows.Scan(&item.ID, &item.TripAdvanceID, &item.TripSettlementID, &item.Amount, &item.ReturnDate, &item.PaymentMethod, &item.ReceivedBy, &item.Notes, &item.CreatedAt); err != nil {
      return nil, err
    }
    items = append(items, item)
  }
  return items, rows.Err()
}

func (r *Repository) Get(ctx context.Context, id uuid.UUID) (AdvanceReturn, error) {
  var item AdvanceReturn
  row := r.pool.QueryRow(ctx, `select id, trip_advance_id, trip_settlement_id, amount, return_date, payment_method, received_by, notes, created_at from advance_returns where id=$1`, id)
  if err := row.Scan(&item.ID, &item.TripAdvanceID, &item.TripSettlementID, &item.Amount, &item.ReturnDate, &item.PaymentMethod, &item.ReceivedBy, &item.Notes, &item.CreatedAt); err != nil {
    return item, err
  }
  return item, nil
}

func (r *Repository) Create(ctx context.Context, input CreateAdvanceReturnInput, receivedBy *string) (AdvanceReturn, error) {
  returnDate := time.Now()
  if input.ReturnDate != nil {
    returnDate = *input.ReturnDate
  }

  var item AdvanceReturn
  row := r.pool.QueryRow(ctx,
    `insert into advance_returns (trip_advance_id, trip_settlement_id, amount, return_date, payment_method, received_by, notes)
     values ($1,$2,$3,$4,$5,$6,$7)
     returning id, trip_advance_id, trip_settlement_id, amount, return_date, payment_method, received_by, notes, created_at`,
    input.TripAdvanceID, input.TripSettlementID, input.Amount, returnDate, input.PaymentMethod, receivedBy, input.Notes,
  )
  if err := row.Scan(&item.ID, &item.TripAdvanceID, &item.TripSettlementID, &item.Amount, &item.ReturnDate, &item.PaymentMethod, &item.ReceivedBy, &item.Notes, &item.CreatedAt); err != nil {
    return item, err
  }
  return item, nil
}

func IsNotFound(err error) bool {
  return err == pgx.ErrNoRows
}
