package trip_advances

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

func (r *Repository) List(ctx context.Context, filter ListFilter) ([]TripAdvance, error) {
  query := `select id, trip_id, driver_id, amount, status, purpose, delivered_at, delivered_by, settled_at, notes, created_by, created_at, updated_at from trip_advances`
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

  items := []TripAdvance{}
  for rows.Next() {
    var item TripAdvance
    if err := rows.Scan(&item.ID, &item.TripID, &item.DriverID, &item.Amount, &item.Status, &item.Purpose, &item.DeliveredAt, &item.DeliveredBy, &item.SettledAt, &item.Notes, &item.CreatedBy, &item.CreatedAt, &item.UpdatedAt); err != nil {
      return nil, err
    }
    items = append(items, item)
  }
  return items, rows.Err()
}

func (r *Repository) Get(ctx context.Context, id uuid.UUID) (TripAdvance, error) {
  var item TripAdvance
  row := r.pool.QueryRow(ctx, `select id, trip_id, driver_id, amount, status, purpose, delivered_at, delivered_by, settled_at, notes, created_by, created_at, updated_at from trip_advances where id=$1`, id)
  if err := row.Scan(&item.ID, &item.TripID, &item.DriverID, &item.Amount, &item.Status, &item.Purpose, &item.DeliveredAt, &item.DeliveredBy, &item.SettledAt, &item.Notes, &item.CreatedBy, &item.CreatedAt, &item.UpdatedAt); err != nil {
    return item, err
  }
  return item, nil
}

func (r *Repository) Create(ctx context.Context, input CreateTripAdvanceInput, createdBy *string) (TripAdvance, error) {
  var item TripAdvance
  row := r.pool.QueryRow(ctx,
    `insert into trip_advances (trip_id, driver_id, amount, purpose, notes, created_by)
     values ($1,$2,$3,$4,$5,$6)
     returning id, trip_id, driver_id, amount, status, purpose, delivered_at, delivered_by, settled_at, notes, created_by, created_at, updated_at`,
    input.TripID, input.DriverID, input.Amount, input.Purpose, input.Notes, createdBy,
  )
  if err := row.Scan(&item.ID, &item.TripID, &item.DriverID, &item.Amount, &item.Status, &item.Purpose, &item.DeliveredAt, &item.DeliveredBy, &item.SettledAt, &item.Notes, &item.CreatedBy, &item.CreatedAt, &item.UpdatedAt); err != nil {
    return item, err
  }
  return item, nil
}

func (r *Repository) Update(ctx context.Context, id uuid.UUID, input UpdateTripAdvanceInput) (TripAdvance, error) {
  sets := []string{}
  args := []interface{}{}
  idx := 1

  if input.Amount != nil {
    sets = append(sets, fmt.Sprintf("amount=$%d", idx))
    args = append(args, *input.Amount)
    idx++
  }
  if input.Purpose != nil {
    sets = append(sets, fmt.Sprintf("purpose=$%d", idx))
    args = append(args, *input.Purpose)
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
  query := fmt.Sprintf(`update trip_advances set %s where id=$%d returning id, trip_id, driver_id, amount, status, purpose, delivered_at, delivered_by, settled_at, notes, created_by, created_at, updated_at`, strings.Join(sets, ", "), idx)

  var item TripAdvance
  row := r.pool.QueryRow(ctx, query, args...)
  if err := row.Scan(&item.ID, &item.TripID, &item.DriverID, &item.Amount, &item.Status, &item.Purpose, &item.DeliveredAt, &item.DeliveredBy, &item.SettledAt, &item.Notes, &item.CreatedBy, &item.CreatedAt, &item.UpdatedAt); err != nil {
    return item, err
  }
  return item, nil
}

func (r *Repository) MarkDelivered(ctx context.Context, id uuid.UUID, deliveredBy *string) (TripAdvance, error) {
  var item TripAdvance
  row := r.pool.QueryRow(ctx,
    `update trip_advances
     set status='DELIVERED', delivered_at=now(), delivered_by=$2, updated_at=now()
     where id=$1
     returning id, trip_id, driver_id, amount, status, purpose, delivered_at, delivered_by, settled_at, notes, created_by, created_at, updated_at`,
    id, deliveredBy,
  )
  if err := row.Scan(&item.ID, &item.TripID, &item.DriverID, &item.Amount, &item.Status, &item.Purpose, &item.DeliveredAt, &item.DeliveredBy, &item.SettledAt, &item.Notes, &item.CreatedBy, &item.CreatedAt, &item.UpdatedAt); err != nil {
    return item, err
  }
  return item, nil
}

func (r *Repository) MarkSettledByTrip(ctx context.Context, tripID uuid.UUID) error {
  _, err := r.pool.Exec(ctx, `update trip_advances set status='SETTLED', settled_at=now(), updated_at=now() where trip_id=$1 and status='DELIVERED'`, tripID)
  return err
}

func (r *Repository) SumDeliveredByTrip(ctx context.Context, tripID uuid.UUID) (float64, error) {
  var total float64
  if err := r.pool.QueryRow(ctx, `select coalesce(sum(amount), 0) from trip_advances where trip_id=$1 and status in ('DELIVERED','SETTLED')`, tripID).Scan(&total); err != nil {
    return 0, err
  }
  return total, nil
}

func IsNotFound(err error) bool {
  return err == pgx.ErrNoRows
}
