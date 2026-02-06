package pricing

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

func (r *Repository) List(ctx context.Context, filter ListFilter) ([]PricingRule, error) {
  query := `select id, name, scope, scope_id, rule_type, priority, is_active, params, created_at, updated_at from pricing_rules`
  args := []interface{}{}
  clauses := []string{}

  if filter.Scope != nil {
    args = append(args, *filter.Scope)
    clauses = append(clauses, fmt.Sprintf("scope=$%d", len(args)))
  }
  if filter.ScopeID != nil {
    args = append(args, *filter.ScopeID)
    clauses = append(clauses, fmt.Sprintf("scope_id=$%d", len(args)))
  }
  if filter.RuleType != nil {
    args = append(args, *filter.RuleType)
    clauses = append(clauses, fmt.Sprintf("rule_type=$%d", len(args)))
  }
  if filter.IsActive != nil {
    args = append(args, *filter.IsActive)
    clauses = append(clauses, fmt.Sprintf("is_active=$%d", len(args)))
  }

  if len(clauses) > 0 {
    query += " where " + strings.Join(clauses, " and ")
  }

  query += " order by priority asc, created_at desc"

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

  items := []PricingRule{}
  for rows.Next() {
    var item PricingRule
    if err := rows.Scan(&item.ID, &item.Name, &item.Scope, &item.ScopeID, &item.RuleType, &item.Priority, &item.IsActive, &item.Params, &item.CreatedAt, &item.UpdatedAt); err != nil {
      return nil, err
    }
    items = append(items, item)
  }
  if err := rows.Err(); err != nil {
    return nil, err
  }
  return items, nil
}

func (r *Repository) Get(ctx context.Context, id uuid.UUID) (PricingRule, error) {
  var item PricingRule
  row := r.pool.QueryRow(ctx, `select id, name, scope, scope_id, rule_type, priority, is_active, params, created_at, updated_at from pricing_rules where id=$1`, id)
  if err := row.Scan(&item.ID, &item.Name, &item.Scope, &item.ScopeID, &item.RuleType, &item.Priority, &item.IsActive, &item.Params, &item.CreatedAt, &item.UpdatedAt); err != nil {
    return item, err
  }
  return item, nil
}

func (r *Repository) Create(ctx context.Context, input CreatePricingRuleInput) (PricingRule, error) {
  priority := 100
  if input.Priority != nil {
    priority = *input.Priority
  }
  isActive := true
  if input.IsActive != nil {
    isActive = *input.IsActive
  }

  var item PricingRule
  row := r.pool.QueryRow(ctx,
    `insert into pricing_rules (name, scope, scope_id, rule_type, priority, is_active, params)
     values ($1,$2,$3,$4,$5,$6,$7)
     returning id, name, scope, scope_id, rule_type, priority, is_active, params, created_at, updated_at`,
    input.Name, input.Scope, input.ScopeID, input.RuleType, priority, isActive, input.Params,
  )
  if err := row.Scan(&item.ID, &item.Name, &item.Scope, &item.ScopeID, &item.RuleType, &item.Priority, &item.IsActive, &item.Params, &item.CreatedAt, &item.UpdatedAt); err != nil {
    return item, err
  }
  return item, nil
}

func (r *Repository) Update(ctx context.Context, id uuid.UUID, input UpdatePricingRuleInput) (PricingRule, error) {
  sets := []string{}
  args := []interface{}{}
  idx := 1

  if input.Name != nil {
    sets = append(sets, fmt.Sprintf("name=$%d", idx))
    args = append(args, *input.Name)
    idx++
  }
  if input.Scope != nil {
    sets = append(sets, fmt.Sprintf("scope=$%d", idx))
    args = append(args, *input.Scope)
    idx++
  }
  if input.Scope != nil && *input.Scope == "GLOBAL" {
    sets = append(sets, "scope_id=NULL")
  }
  if input.ScopeID != nil {
    sets = append(sets, fmt.Sprintf("scope_id=$%d", idx))
    args = append(args, *input.ScopeID)
    idx++
  }
  if input.RuleType != nil {
    sets = append(sets, fmt.Sprintf("rule_type=$%d", idx))
    args = append(args, *input.RuleType)
    idx++
  }
  if input.Priority != nil {
    sets = append(sets, fmt.Sprintf("priority=$%d", idx))
    args = append(args, *input.Priority)
    idx++
  }
  if input.IsActive != nil {
    sets = append(sets, fmt.Sprintf("is_active=$%d", idx))
    args = append(args, *input.IsActive)
    idx++
  }
  if input.Params != nil {
    sets = append(sets, fmt.Sprintf("params=$%d", idx))
    args = append(args, input.Params)
    idx++
  }

  if len(sets) == 0 {
    return r.Get(ctx, id)
  }

  sets = append(sets, "updated_at=now()")
  args = append(args, id)
  query := fmt.Sprintf(`update pricing_rules set %s where id=$%d returning id, name, scope, scope_id, rule_type, priority, is_active, params, created_at, updated_at`, strings.Join(sets, ", "), idx)

  var item PricingRule
  row := r.pool.QueryRow(ctx, query, args...)
  if err := row.Scan(&item.ID, &item.Name, &item.Scope, &item.ScopeID, &item.RuleType, &item.Priority, &item.IsActive, &item.Params, &item.CreatedAt, &item.UpdatedAt); err != nil {
    return item, err
  }
  return item, nil
}

func IsNotFound(err error) bool {
  return err == pgx.ErrNoRows
}
