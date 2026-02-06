package trip_settlements

import (
  "context"
  "errors"
  "fmt"
  "strings"

  "github.com/google/uuid"
  "github.com/jackc/pgx/v5"
  "github.com/jackc/pgx/v5/pgconn"
  "github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
  pool *pgxpool.Pool
}

var ErrTripDriverMissing = errors.New("trip driver not set")

func NewRepository(pool *pgxpool.Pool) *Repository {
  return &Repository{pool: pool}
}

func (r *Repository) List(ctx context.Context, filter ListFilter) ([]TripSettlement, error) {
  query := `select id, trip_id, driver_id, status, advance_amount, expenses_total, balance, amount_to_return, amount_to_reimburse, reviewed_by, reviewed_at, approved_by, approved_at, completed_at, notes, created_by, created_at, updated_at from trip_settlements`
  args := []interface{}{}
  clauses := []string{}

  if filter.TripID != "" {
    args = append(args, filter.TripID)
    clauses = append(clauses, fmt.Sprintf("trip_id=$%d", len(args)))
  }
  if filter.DriverID != "" {
    args = append(args, filter.DriverID)
    clauses = append(clauses, fmt.Sprintf("driver_id=$%d", len(args)))
  }
  if filter.Status != "" {
    args = append(args, filter.Status)
    clauses = append(clauses, fmt.Sprintf("status=$%d", len(args)))
  }

  if len(clauses) > 0 {
    query += " where " + strings.Join(clauses, " and ")
  }

  query += " order by created_at desc"

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

  items := []TripSettlement{}
  for rows.Next() {
    var item TripSettlement
    if err := rows.Scan(&item.ID, &item.TripID, &item.DriverID, &item.Status, &item.AdvanceAmount, &item.ExpensesTotal, &item.Balance, &item.AmountToReturn, &item.AmountToReimburse, &item.ReviewedBy, &item.ReviewedAt, &item.ApprovedBy, &item.ApprovedAt, &item.CompletedAt, &item.Notes, &item.CreatedBy, &item.CreatedAt, &item.UpdatedAt); err != nil {
      return nil, err
    }
    items = append(items, item)
  }
  return items, rows.Err()
}

func (r *Repository) Get(ctx context.Context, id uuid.UUID) (TripSettlement, error) {
  var item TripSettlement
  row := r.pool.QueryRow(ctx, `select id, trip_id, driver_id, status, advance_amount, expenses_total, balance, amount_to_return, amount_to_reimburse, reviewed_by, reviewed_at, approved_by, approved_at, completed_at, notes, created_by, created_at, updated_at from trip_settlements where id=$1`, id)
  if err := row.Scan(&item.ID, &item.TripID, &item.DriverID, &item.Status, &item.AdvanceAmount, &item.ExpensesTotal, &item.Balance, &item.AmountToReturn, &item.AmountToReimburse, &item.ReviewedBy, &item.ReviewedAt, &item.ApprovedBy, &item.ApprovedAt, &item.CompletedAt, &item.Notes, &item.CreatedBy, &item.CreatedAt, &item.UpdatedAt); err != nil {
    return item, err
  }
  return item, nil
}

func (r *Repository) GetTripDriverID(ctx context.Context, tripID uuid.UUID) (string, error) {
  var driverID *string
  if err := r.pool.QueryRow(ctx, `select driver_id from trips where id=$1`, tripID).Scan(&driverID); err != nil {
    return "", err
  }
  if driverID == nil || *driverID == "" {
    return "", ErrTripDriverMissing
  }
  return *driverID, nil
}

func (r *Repository) Create(ctx context.Context, tripID string, driverID string, advanceAmount float64, expensesTotal float64, balance float64, amountToReturn float64, amountToReimburse float64, notes *string, createdBy *string) (TripSettlement, error) {
  var item TripSettlement
  row := r.pool.QueryRow(ctx,
    `insert into trip_settlements (trip_id, driver_id, advance_amount, expenses_total, balance, amount_to_return, amount_to_reimburse, notes, created_by)
     values ($1,$2,$3,$4,$5,$6,$7,$8,$9)
     returning id, trip_id, driver_id, status, advance_amount, expenses_total, balance, amount_to_return, amount_to_reimburse, reviewed_by, reviewed_at, approved_by, approved_at, completed_at, notes, created_by, created_at, updated_at`,
    tripID, driverID, advanceAmount, expensesTotal, balance, amountToReturn, amountToReimburse, notes, createdBy,
  )
  if err := row.Scan(&item.ID, &item.TripID, &item.DriverID, &item.Status, &item.AdvanceAmount, &item.ExpensesTotal, &item.Balance, &item.AmountToReturn, &item.AmountToReimburse, &item.ReviewedBy, &item.ReviewedAt, &item.ApprovedBy, &item.ApprovedAt, &item.CompletedAt, &item.Notes, &item.CreatedBy, &item.CreatedAt, &item.UpdatedAt); err != nil {
    return item, err
  }
  return item, nil
}

func (r *Repository) SetUnderReview(ctx context.Context, id uuid.UUID, reviewedBy *string) (TripSettlement, error) {
  var item TripSettlement
  row := r.pool.QueryRow(ctx,
    `update trip_settlements set status='UNDER_REVIEW', reviewed_by=$2, reviewed_at=now(), updated_at=now()
     where id=$1
     returning id, trip_id, driver_id, status, advance_amount, expenses_total, balance, amount_to_return, amount_to_reimburse, reviewed_by, reviewed_at, approved_by, approved_at, completed_at, notes, created_by, created_at, updated_at`,
    id, reviewedBy,
  )
  if err := row.Scan(&item.ID, &item.TripID, &item.DriverID, &item.Status, &item.AdvanceAmount, &item.ExpensesTotal, &item.Balance, &item.AmountToReturn, &item.AmountToReimburse, &item.ReviewedBy, &item.ReviewedAt, &item.ApprovedBy, &item.ApprovedAt, &item.CompletedAt, &item.Notes, &item.CreatedBy, &item.CreatedAt, &item.UpdatedAt); err != nil {
    return item, err
  }
  return item, nil
}

func (r *Repository) SetApproved(ctx context.Context, id uuid.UUID, approvedBy *string) (TripSettlement, error) {
  var item TripSettlement
  row := r.pool.QueryRow(ctx,
    `update trip_settlements set status='APPROVED', approved_by=$2, approved_at=now(), updated_at=now()
     where id=$1
     returning id, trip_id, driver_id, status, advance_amount, expenses_total, balance, amount_to_return, amount_to_reimburse, reviewed_by, reviewed_at, approved_by, approved_at, completed_at, notes, created_by, created_at, updated_at`,
    id, approvedBy,
  )
  if err := row.Scan(&item.ID, &item.TripID, &item.DriverID, &item.Status, &item.AdvanceAmount, &item.ExpensesTotal, &item.Balance, &item.AmountToReturn, &item.AmountToReimburse, &item.ReviewedBy, &item.ReviewedAt, &item.ApprovedBy, &item.ApprovedAt, &item.CompletedAt, &item.Notes, &item.CreatedBy, &item.CreatedAt, &item.UpdatedAt); err != nil {
    return item, err
  }
  return item, nil
}

func (r *Repository) SetRejected(ctx context.Context, id uuid.UUID, reviewedBy *string) (TripSettlement, error) {
  var item TripSettlement
  row := r.pool.QueryRow(ctx,
    `update trip_settlements set status='REJECTED', reviewed_by=$2, reviewed_at=now(), updated_at=now()
     where id=$1
     returning id, trip_id, driver_id, status, advance_amount, expenses_total, balance, amount_to_return, amount_to_reimburse, reviewed_by, reviewed_at, approved_by, approved_at, completed_at, notes, created_by, created_at, updated_at`,
    id, reviewedBy,
  )
  if err := row.Scan(&item.ID, &item.TripID, &item.DriverID, &item.Status, &item.AdvanceAmount, &item.ExpensesTotal, &item.Balance, &item.AmountToReturn, &item.AmountToReimburse, &item.ReviewedBy, &item.ReviewedAt, &item.ApprovedBy, &item.ApprovedAt, &item.CompletedAt, &item.Notes, &item.CreatedBy, &item.CreatedAt, &item.UpdatedAt); err != nil {
    return item, err
  }
  return item, nil
}

func (r *Repository) Complete(ctx context.Context, id uuid.UUID) (TripSettlement, error) {
  tx, err := r.pool.Begin(ctx)
  if err != nil {
    return TripSettlement{}, err
  }
  defer tx.Rollback(ctx)

  var item TripSettlement
  row := tx.QueryRow(ctx,
    `update trip_settlements set status='COMPLETED', completed_at=now(), updated_at=now()
     where id=$1
     returning id, trip_id, driver_id, status, advance_amount, expenses_total, balance, amount_to_return, amount_to_reimburse, reviewed_by, reviewed_at, approved_by, approved_at, completed_at, notes, created_by, created_at, updated_at`,
    id,
  )
  if err := row.Scan(&item.ID, &item.TripID, &item.DriverID, &item.Status, &item.AdvanceAmount, &item.ExpensesTotal, &item.Balance, &item.AmountToReturn, &item.AmountToReimburse, &item.ReviewedBy, &item.ReviewedAt, &item.ApprovedBy, &item.ApprovedAt, &item.CompletedAt, &item.Notes, &item.CreatedBy, &item.CreatedAt, &item.UpdatedAt); err != nil {
    return TripSettlement{}, err
  }

  if _, err := tx.Exec(ctx, `update trip_advances set status='SETTLED', settled_at=now(), updated_at=now() where trip_id=$1 and status='DELIVERED'`, item.TripID); err != nil {
    return TripSettlement{}, err
  }

  if err := tx.Commit(ctx); err != nil {
    return TripSettlement{}, err
  }
  return item, nil
}

func IsNotFound(err error) bool {
  return err == pgx.ErrNoRows
}

func IsUniqueViolation(err error) bool {
  pgErr, ok := err.(*pgconn.PgError)
  if !ok {
    return false
  }
  return pgErr.Code == "23505"
}
