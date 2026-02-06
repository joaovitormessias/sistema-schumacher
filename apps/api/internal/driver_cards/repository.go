package driver_cards

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
var ErrInvalidTransactionType = errors.New("invalid transaction_type")
var ErrInsufficientBalance = errors.New("insufficient balance")

func NewRepository(pool *pgxpool.Pool) *Repository {
  return &Repository{pool: pool}
}

func (r *Repository) List(ctx context.Context, filter ListFilter) ([]DriverCard, error) {
  query := `select id, driver_id, card_number, card_type, current_balance, is_active, is_blocked, issued_at, blocked_at, blocked_by, block_reason, notes, created_at, updated_at from driver_cards`
  args := []interface{}{}
  clauses := []string{}

  if filter.DriverID != "" {
    args = append(args, filter.DriverID)
    clauses = append(clauses, fmt.Sprintf("driver_id=$%d", len(args)))
  }
  if filter.IsActive != nil {
    args = append(args, *filter.IsActive)
    clauses = append(clauses, fmt.Sprintf("is_active=$%d", len(args)))
  }
  if filter.IsBlocked != nil {
    args = append(args, *filter.IsBlocked)
    clauses = append(clauses, fmt.Sprintf("is_blocked=$%d", len(args)))
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

  items := []DriverCard{}
  for rows.Next() {
    var item DriverCard
    if err := rows.Scan(&item.ID, &item.DriverID, &item.CardNumber, &item.CardType, &item.CurrentBalance, &item.IsActive, &item.IsBlocked, &item.IssuedAt, &item.BlockedAt, &item.BlockedBy, &item.BlockReason, &item.Notes, &item.CreatedAt, &item.UpdatedAt); err != nil {
      return nil, err
    }
    items = append(items, item)
  }
  return items, rows.Err()
}

func (r *Repository) Get(ctx context.Context, id uuid.UUID) (DriverCard, error) {
  var item DriverCard
  row := r.pool.QueryRow(ctx, `select id, driver_id, card_number, card_type, current_balance, is_active, is_blocked, issued_at, blocked_at, blocked_by, block_reason, notes, created_at, updated_at from driver_cards where id=$1`, id)
  if err := row.Scan(&item.ID, &item.DriverID, &item.CardNumber, &item.CardType, &item.CurrentBalance, &item.IsActive, &item.IsBlocked, &item.IssuedAt, &item.BlockedAt, &item.BlockedBy, &item.BlockReason, &item.Notes, &item.CreatedAt, &item.UpdatedAt); err != nil {
    return item, err
  }
  return item, nil
}

func (r *Repository) Create(ctx context.Context, input CreateDriverCardInput) (DriverCard, error) {
  balance := 0.0
  if input.CurrentBalance != nil {
    balance = *input.CurrentBalance
  }

  var item DriverCard
  row := r.pool.QueryRow(ctx,
    `insert into driver_cards (driver_id, card_number, card_type, current_balance, notes)
     values ($1,$2,$3,$4,$5)
     returning id, driver_id, card_number, card_type, current_balance, is_active, is_blocked, issued_at, blocked_at, blocked_by, block_reason, notes, created_at, updated_at`,
    input.DriverID, input.CardNumber, input.CardType, balance, input.Notes,
  )
  if err := row.Scan(&item.ID, &item.DriverID, &item.CardNumber, &item.CardType, &item.CurrentBalance, &item.IsActive, &item.IsBlocked, &item.IssuedAt, &item.BlockedAt, &item.BlockedBy, &item.BlockReason, &item.Notes, &item.CreatedAt, &item.UpdatedAt); err != nil {
    return item, err
  }
  return item, nil
}

func (r *Repository) Update(ctx context.Context, id uuid.UUID, input UpdateDriverCardInput) (DriverCard, error) {
  sets := []string{}
  args := []interface{}{}
  idx := 1

  if input.CardNumber != nil {
    sets = append(sets, fmt.Sprintf("card_number=$%d", idx))
    args = append(args, *input.CardNumber)
    idx++
  }
  if input.CardType != nil {
    sets = append(sets, fmt.Sprintf("card_type=$%d", idx))
    args = append(args, *input.CardType)
    idx++
  }
  if input.IsActive != nil {
    sets = append(sets, fmt.Sprintf("is_active=$%d", idx))
    args = append(args, *input.IsActive)
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
  query := fmt.Sprintf(`update driver_cards set %s where id=$%d returning id, driver_id, card_number, card_type, current_balance, is_active, is_blocked, issued_at, blocked_at, blocked_by, block_reason, notes, created_at, updated_at`, strings.Join(sets, ", "), idx)

  var item DriverCard
  row := r.pool.QueryRow(ctx, query, args...)
  if err := row.Scan(&item.ID, &item.DriverID, &item.CardNumber, &item.CardType, &item.CurrentBalance, &item.IsActive, &item.IsBlocked, &item.IssuedAt, &item.BlockedAt, &item.BlockedBy, &item.BlockReason, &item.Notes, &item.CreatedAt, &item.UpdatedAt); err != nil {
    return item, err
  }
  return item, nil
}

func (r *Repository) Block(ctx context.Context, id uuid.UUID, reason *string, blockedBy *string) (DriverCard, error) {
  var item DriverCard
  row := r.pool.QueryRow(ctx,
    `update driver_cards
     set is_blocked=true, blocked_at=now(), blocked_by=$2, block_reason=$3, updated_at=now()
     where id=$1
     returning id, driver_id, card_number, card_type, current_balance, is_active, is_blocked, issued_at, blocked_at, blocked_by, block_reason, notes, created_at, updated_at`,
    id, blockedBy, reason,
  )
  if err := row.Scan(&item.ID, &item.DriverID, &item.CardNumber, &item.CardType, &item.CurrentBalance, &item.IsActive, &item.IsBlocked, &item.IssuedAt, &item.BlockedAt, &item.BlockedBy, &item.BlockReason, &item.Notes, &item.CreatedAt, &item.UpdatedAt); err != nil {
    return item, err
  }
  return item, nil
}

func (r *Repository) Unblock(ctx context.Context, id uuid.UUID) (DriverCard, error) {
  var item DriverCard
  row := r.pool.QueryRow(ctx,
    `update driver_cards
     set is_blocked=false, blocked_at=null, blocked_by=null, block_reason=null, updated_at=now()
     where id=$1
     returning id, driver_id, card_number, card_type, current_balance, is_active, is_blocked, issued_at, blocked_at, blocked_by, block_reason, notes, created_at, updated_at`,
    id,
  )
  if err := row.Scan(&item.ID, &item.DriverID, &item.CardNumber, &item.CardType, &item.CurrentBalance, &item.IsActive, &item.IsBlocked, &item.IssuedAt, &item.BlockedAt, &item.BlockedBy, &item.BlockReason, &item.Notes, &item.CreatedAt, &item.UpdatedAt); err != nil {
    return item, err
  }
  return item, nil
}

func (r *Repository) CreateTransaction(ctx context.Context, cardID uuid.UUID, input CreateCardTransactionInput, tripExpenseID *uuid.UUID, performedBy *string) (DriverCardTransaction, error) {
  if input.Amount < 0 {
    return DriverCardTransaction{}, ErrInvalidAmount
  }

  tx, err := r.pool.Begin(ctx)
  if err != nil {
    return DriverCardTransaction{}, err
  }
  defer tx.Rollback(ctx)

  var balance float64
  if err := tx.QueryRow(ctx, `select current_balance from driver_cards where id=$1 for update`, cardID).Scan(&balance); err != nil {
    return DriverCardTransaction{}, err
  }

  delta := input.Amount
  switch input.TransactionType {
  case "DEBIT":
    delta = -input.Amount
  case "CREDIT", "REFUND", "ADJUSTMENT":
    delta = input.Amount
  default:
    return DriverCardTransaction{}, ErrInvalidTransactionType
  }

  newBalance := balance + delta
  if newBalance < 0 {
    return DriverCardTransaction{}, ErrInsufficientBalance
  }

  if _, err := tx.Exec(ctx, `update driver_cards set current_balance=$1, updated_at=now() where id=$2`, newBalance, cardID); err != nil {
    return DriverCardTransaction{}, err
  }

  var item DriverCardTransaction
  row := tx.QueryRow(ctx,
    `insert into driver_card_transactions (card_id, transaction_type, amount, balance_before, balance_after, description, trip_expense_id, performed_by)
     values ($1,$2,$3,$4,$5,$6,$7,$8)
     returning id, card_id, transaction_type, amount, balance_before, balance_after, description, trip_expense_id, performed_by, created_at`,
    cardID, input.TransactionType, input.Amount, balance, newBalance, input.Description, tripExpenseID, performedBy,
  )
  if err := row.Scan(&item.ID, &item.CardID, &item.TransactionType, &item.Amount, &item.BalanceBefore, &item.BalanceAfter, &item.Description, &item.TripExpenseID, &item.PerformedBy, &item.CreatedAt); err != nil {
    return DriverCardTransaction{}, err
  }

  if err := tx.Commit(ctx); err != nil {
    return DriverCardTransaction{}, err
  }
  return item, nil
}

func (r *Repository) ListTransactions(ctx context.Context, filter TransactionListFilter) ([]DriverCardTransaction, error) {
  query := `select id, card_id, transaction_type, amount, balance_before, balance_after, description, trip_expense_id, performed_by, created_at from driver_card_transactions`
  args := []interface{}{}
  clauses := []string{}

  if filter.CardID != "" {
    args = append(args, filter.CardID)
    clauses = append(clauses, fmt.Sprintf("card_id=$%d", len(args)))
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

  items := []DriverCardTransaction{}
  for rows.Next() {
    var item DriverCardTransaction
    if err := rows.Scan(&item.ID, &item.CardID, &item.TransactionType, &item.Amount, &item.BalanceBefore, &item.BalanceAfter, &item.Description, &item.TripExpenseID, &item.PerformedBy, &item.CreatedAt); err != nil {
      return nil, err
    }
    items = append(items, item)
  }
  return items, rows.Err()
}

func IsNotFound(err error) bool {
  return err == pgx.ErrNoRows
}
