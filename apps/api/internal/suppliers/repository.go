package suppliers

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

func (r *Repository) List(ctx context.Context, filter ListFilter) ([]Supplier, error) {
	limit := filter.Limit
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	query := `SELECT id, name, document, phone, email, payment_terms, billing_day, is_active, notes, created_at, updated_at
		FROM suppliers`
	args := []interface{}{}
	conditions := []string{}
	idx := 1

	if filter.Active != nil {
		conditions = append(conditions, fmt.Sprintf("is_active = $%d", idx))
		args = append(args, *filter.Active)
		idx++
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += fmt.Sprintf(" ORDER BY name ASC LIMIT $%d OFFSET $%d", idx, idx+1)
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []Supplier{}
	for rows.Next() {
		var item Supplier
		if err := rows.Scan(&item.ID, &item.Name, &item.Document, &item.Phone, &item.Email, &item.PaymentTerms, &item.BillingDay, &item.IsActive, &item.Notes, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) Get(ctx context.Context, id uuid.UUID) (Supplier, error) {
	var item Supplier
	row := r.pool.QueryRow(ctx, `SELECT id, name, document, phone, email, payment_terms, billing_day, is_active, notes, created_at, updated_at FROM suppliers WHERE id=$1`, id)
	if err := row.Scan(&item.ID, &item.Name, &item.Document, &item.Phone, &item.Email, &item.PaymentTerms, &item.BillingDay, &item.IsActive, &item.Notes, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return item, err
	}
	return item, nil
}

func (r *Repository) Create(ctx context.Context, input CreateSupplierInput) (Supplier, error) {
	isActive := true
	if input.IsActive != nil {
		isActive = *input.IsActive
	}

	var item Supplier
	row := r.pool.QueryRow(ctx,
		`INSERT INTO suppliers (name, document, phone, email, payment_terms, billing_day, is_active, notes)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		 RETURNING id, name, document, phone, email, payment_terms, billing_day, is_active, notes, created_at, updated_at`,
		input.Name, input.Document, input.Phone, input.Email, input.PaymentTerms, input.BillingDay, isActive, input.Notes,
	)
	if err := row.Scan(&item.ID, &item.Name, &item.Document, &item.Phone, &item.Email, &item.PaymentTerms, &item.BillingDay, &item.IsActive, &item.Notes, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return item, err
	}
	return item, nil
}

func (r *Repository) Update(ctx context.Context, id uuid.UUID, input UpdateSupplierInput) (Supplier, error) {
	sets := []string{"updated_at = now()"}
	args := []interface{}{}
	idx := 1

	if input.Name != nil {
		sets = append(sets, fmt.Sprintf("name=$%d", idx))
		args = append(args, *input.Name)
		idx++
	}
	if input.Document != nil {
		sets = append(sets, fmt.Sprintf("document=$%d", idx))
		args = append(args, *input.Document)
		idx++
	}
	if input.Phone != nil {
		sets = append(sets, fmt.Sprintf("phone=$%d", idx))
		args = append(args, *input.Phone)
		idx++
	}
	if input.Email != nil {
		sets = append(sets, fmt.Sprintf("email=$%d", idx))
		args = append(args, *input.Email)
		idx++
	}
	if input.PaymentTerms != nil {
		sets = append(sets, fmt.Sprintf("payment_terms=$%d", idx))
		args = append(args, *input.PaymentTerms)
		idx++
	}
	if input.BillingDay != nil {
		sets = append(sets, fmt.Sprintf("billing_day=$%d", idx))
		args = append(args, *input.BillingDay)
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

	args = append(args, id)
	query := fmt.Sprintf(`UPDATE suppliers SET %s WHERE id=$%d RETURNING id, name, document, phone, email, payment_terms, billing_day, is_active, notes, created_at, updated_at`, strings.Join(sets, ", "), idx)

	var item Supplier
	row := r.pool.QueryRow(ctx, query, args...)
	if err := row.Scan(&item.ID, &item.Name, &item.Document, &item.Phone, &item.Email, &item.PaymentTerms, &item.BillingDay, &item.IsActive, &item.Notes, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return item, err
	}
	return item, nil
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM suppliers WHERE id=$1`, id)
	return err
}

func IsNotFound(err error) bool {
	return err == pgx.ErrNoRows
}
