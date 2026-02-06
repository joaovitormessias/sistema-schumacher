package trip_expenses

import (
  "context"
  "errors"
  "fmt"
  "strings"

  "github.com/google/uuid"
  "github.com/jackc/pgx/v5"
  "github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
  pool *pgxpool.Pool
}

var ErrInvalidAmount = errors.New("amount must be positive")
var ErrInsufficientBalance = errors.New("insufficient balance")
var ErrCardInactive = errors.New("card inactive")
var ErrCardBlocked = errors.New("card blocked")

func NewRepository(pool *pgxpool.Pool) *Repository {
  return &Repository{pool: pool}
}

func (r *Repository) List(ctx context.Context, filter ListFilter) ([]TripExpense, error) {
  query := `select id, trip_id, driver_id, expense_type, amount, description, expense_date, payment_method, driver_card_id, receipt_number, is_approved, approved_by, approved_at, notes, created_by, created_at, updated_at from trip_expenses`
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
  if filter.ExpenseType != "" {
    args = append(args, filter.ExpenseType)
    clauses = append(clauses, fmt.Sprintf("expense_type=$%d", len(args)))
  }
  if filter.PaymentMethod != "" {
    args = append(args, filter.PaymentMethod)
    clauses = append(clauses, fmt.Sprintf("payment_method=$%d", len(args)))
  }
  if filter.Approved != nil {
    args = append(args, *filter.Approved)
    clauses = append(clauses, fmt.Sprintf("is_approved=$%d", len(args)))
  }

  if len(clauses) > 0 {
    query += " where " + strings.Join(clauses, " and ")
  }

  query += " order by expense_date desc"

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

  items := []TripExpense{}
  for rows.Next() {
    var item TripExpense
    if err := rows.Scan(&item.ID, &item.TripID, &item.DriverID, &item.ExpenseType, &item.Amount, &item.Description, &item.ExpenseDate, &item.PaymentMethod, &item.DriverCardID, &item.ReceiptNumber, &item.IsApproved, &item.ApprovedBy, &item.ApprovedAt, &item.Notes, &item.CreatedBy, &item.CreatedAt, &item.UpdatedAt); err != nil {
      return nil, err
    }
    items = append(items, item)
  }
  return items, rows.Err()
}

func (r *Repository) Get(ctx context.Context, id uuid.UUID) (TripExpense, error) {
  var item TripExpense
  row := r.pool.QueryRow(ctx, `select id, trip_id, driver_id, expense_type, amount, description, expense_date, payment_method, driver_card_id, receipt_number, is_approved, approved_by, approved_at, notes, created_by, created_at, updated_at from trip_expenses where id=$1`, id)
  if err := row.Scan(&item.ID, &item.TripID, &item.DriverID, &item.ExpenseType, &item.Amount, &item.Description, &item.ExpenseDate, &item.PaymentMethod, &item.DriverCardID, &item.ReceiptNumber, &item.IsApproved, &item.ApprovedBy, &item.ApprovedAt, &item.Notes, &item.CreatedBy, &item.CreatedAt, &item.UpdatedAt); err != nil {
    return item, err
  }
  return item, nil
}

func (r *Repository) Create(ctx context.Context, input CreateTripExpenseInput, createdBy *string, performedBy *string) (TripExpense, error) {
  if input.Amount < 0 {
    return TripExpense{}, ErrInvalidAmount
  }

  tx, err := r.pool.Begin(ctx)
  if err != nil {
    return TripExpense{}, err
  }
  defer tx.Rollback(ctx)

  var item TripExpense
  row := tx.QueryRow(ctx,
    `insert into trip_expenses (trip_id, driver_id, expense_type, amount, description, expense_date, payment_method, driver_card_id, receipt_number, notes, created_by)
     values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
     returning id, trip_id, driver_id, expense_type, amount, description, expense_date, payment_method, driver_card_id, receipt_number, is_approved, approved_by, approved_at, notes, created_by, created_at, updated_at`,
    input.TripID, input.DriverID, input.ExpenseType, input.Amount, input.Description, input.ExpenseDate, input.PaymentMethod, input.DriverCardID, input.ReceiptNumber, input.Notes, createdBy,
  )
  if err := row.Scan(&item.ID, &item.TripID, &item.DriverID, &item.ExpenseType, &item.Amount, &item.Description, &item.ExpenseDate, &item.PaymentMethod, &item.DriverCardID, &item.ReceiptNumber, &item.IsApproved, &item.ApprovedBy, &item.ApprovedAt, &item.Notes, &item.CreatedBy, &item.CreatedAt, &item.UpdatedAt); err != nil {
    return TripExpense{}, err
  }

  if input.PaymentMethod == "CARD" && input.DriverCardID != nil {
    cardID, err := uuid.Parse(*input.DriverCardID)
    if err != nil {
      return TripExpense{}, err
    }

    var balance float64
    var isActive bool
    var isBlocked bool
    if err := tx.QueryRow(ctx, `select current_balance, is_active, is_blocked from driver_cards where id=$1 for update`, cardID).Scan(&balance, &isActive, &isBlocked); err != nil {
      return TripExpense{}, err
    }
    if !isActive {
      return TripExpense{}, ErrCardInactive
    }
    if isBlocked {
      return TripExpense{}, ErrCardBlocked
    }

    newBalance := balance - input.Amount
    if newBalance < 0 {
      return TripExpense{}, ErrInsufficientBalance
    }

    if _, err := tx.Exec(ctx, `update driver_cards set current_balance=$1, updated_at=now() where id=$2`, newBalance, cardID); err != nil {
      return TripExpense{}, err
    }

    _, err = tx.Exec(ctx,
      `insert into driver_card_transactions (card_id, transaction_type, amount, balance_before, balance_after, description, trip_expense_id, performed_by)
       values ($1,'DEBIT',$2,$3,$4,$5,$6,$7)`,
      cardID, input.Amount, balance, newBalance, input.Description, item.ID, performedBy,
    )
    if err != nil {
      return TripExpense{}, err
    }
  }

  if err := tx.Commit(ctx); err != nil {
    return TripExpense{}, err
  }
  return item, nil
}

func (r *Repository) Update(ctx context.Context, id uuid.UUID, input UpdateTripExpenseInput) (TripExpense, error) {
  sets := []string{}
  args := []interface{}{}
  idx := 1

  if input.Amount != nil {
    sets = append(sets, fmt.Sprintf("amount=$%d", idx))
    args = append(args, *input.Amount)
    idx++
  }
  if input.Description != nil {
    sets = append(sets, fmt.Sprintf("description=$%d", idx))
    args = append(args, *input.Description)
    idx++
  }
  if input.ExpenseDate != nil {
    sets = append(sets, fmt.Sprintf("expense_date=$%d", idx))
    args = append(args, *input.ExpenseDate)
    idx++
  }
  if input.ReceiptNumber != nil {
    sets = append(sets, fmt.Sprintf("receipt_number=$%d", idx))
    args = append(args, *input.ReceiptNumber)
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
  query := fmt.Sprintf(`update trip_expenses set %s where id=$%d returning id, trip_id, driver_id, expense_type, amount, description, expense_date, payment_method, driver_card_id, receipt_number, is_approved, approved_by, approved_at, notes, created_by, created_at, updated_at`, strings.Join(sets, ", "), idx)

  var item TripExpense
  row := r.pool.QueryRow(ctx, query, args...)
  if err := row.Scan(&item.ID, &item.TripID, &item.DriverID, &item.ExpenseType, &item.Amount, &item.Description, &item.ExpenseDate, &item.PaymentMethod, &item.DriverCardID, &item.ReceiptNumber, &item.IsApproved, &item.ApprovedBy, &item.ApprovedAt, &item.Notes, &item.CreatedBy, &item.CreatedAt, &item.UpdatedAt); err != nil {
    return item, err
  }
  return item, nil
}

func (r *Repository) Approve(ctx context.Context, id uuid.UUID, approvedBy *string) (TripExpense, error) {
  var item TripExpense
  row := r.pool.QueryRow(ctx,
    `update trip_expenses
     set is_approved=true, approved_by=$2, approved_at=now(), updated_at=now()
     where id=$1
     returning id, trip_id, driver_id, expense_type, amount, description, expense_date, payment_method, driver_card_id, receipt_number, is_approved, approved_by, approved_at, notes, created_by, created_at, updated_at`,
    id, approvedBy,
  )
  if err := row.Scan(&item.ID, &item.TripID, &item.DriverID, &item.ExpenseType, &item.Amount, &item.Description, &item.ExpenseDate, &item.PaymentMethod, &item.DriverCardID, &item.ReceiptNumber, &item.IsApproved, &item.ApprovedBy, &item.ApprovedAt, &item.Notes, &item.CreatedBy, &item.CreatedAt, &item.UpdatedAt); err != nil {
    return item, err
  }
  return item, nil
}

func (r *Repository) SumApprovedByTrip(ctx context.Context, tripID uuid.UUID) (float64, error) {
  var total float64
  if err := r.pool.QueryRow(ctx, `select coalesce(sum(amount), 0) from trip_expenses where trip_id=$1 and is_approved=true`, tripID).Scan(&total); err != nil {
    return 0, err
  }
  return total, nil
}

func IsNotFound(err error) bool {
  return err == pgx.ErrNoRows
}
