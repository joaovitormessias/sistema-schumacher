package products

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

func (r *Repository) List(ctx context.Context, filter ListFilter) ([]Product, error) {
	limit := filter.Limit
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	query := `SELECT id, code, name, category, unit, min_stock, current_stock, last_cost, is_active, created_at, updated_at
		FROM products`
	args := []interface{}{}
	conditions := []string{}
	idx := 1

	if filter.Active != nil {
		conditions = append(conditions, fmt.Sprintf("is_active = $%d", idx))
		args = append(args, *filter.Active)
		idx++
	}
	if filter.Category != nil {
		conditions = append(conditions, fmt.Sprintf("category = $%d", idx))
		args = append(args, *filter.Category)
		idx++
	}
	if filter.Search != nil && *filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(name ILIKE $%d OR code ILIKE $%d)", idx, idx))
		args = append(args, "%"+*filter.Search+"%")
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

	items := []Product{}
	for rows.Next() {
		var item Product
		if err := rows.Scan(&item.ID, &item.Code, &item.Name, &item.Category, &item.Unit, &item.MinStock, &item.CurrentStock, &item.LastCost, &item.IsActive, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) Get(ctx context.Context, id uuid.UUID) (Product, error) {
	var item Product
	row := r.pool.QueryRow(ctx, `SELECT id, code, name, category, unit, min_stock, current_stock, last_cost, is_active, created_at, updated_at FROM products WHERE id=$1`, id)
	if err := row.Scan(&item.ID, &item.Code, &item.Name, &item.Category, &item.Unit, &item.MinStock, &item.CurrentStock, &item.LastCost, &item.IsActive, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return item, err
	}
	return item, nil
}

func (r *Repository) Create(ctx context.Context, input CreateProductInput) (Product, error) {
	isActive := true
	if input.IsActive != nil {
		isActive = *input.IsActive
	}
	unit := "UN"
	if input.Unit != nil {
		unit = *input.Unit
	}
	minStock := 0.0
	if input.MinStock != nil {
		minStock = *input.MinStock
	}

	var item Product
	row := r.pool.QueryRow(ctx,
		`INSERT INTO products (code, name, category, unit, min_stock, is_active)
		 VALUES ($1,$2,$3,$4,$5,$6)
		 RETURNING id, code, name, category, unit, min_stock, current_stock, last_cost, is_active, created_at, updated_at`,
		input.Code, input.Name, input.Category, unit, minStock, isActive,
	)
	if err := row.Scan(&item.ID, &item.Code, &item.Name, &item.Category, &item.Unit, &item.MinStock, &item.CurrentStock, &item.LastCost, &item.IsActive, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return item, err
	}
	return item, nil
}

func (r *Repository) Update(ctx context.Context, id uuid.UUID, input UpdateProductInput) (Product, error) {
	sets := []string{"updated_at = now()"}
	args := []interface{}{}
	idx := 1

	if input.Code != nil {
		sets = append(sets, fmt.Sprintf("code=$%d", idx))
		args = append(args, *input.Code)
		idx++
	}
	if input.Name != nil {
		sets = append(sets, fmt.Sprintf("name=$%d", idx))
		args = append(args, *input.Name)
		idx++
	}
	if input.Category != nil {
		sets = append(sets, fmt.Sprintf("category=$%d", idx))
		args = append(args, *input.Category)
		idx++
	}
	if input.Unit != nil {
		sets = append(sets, fmt.Sprintf("unit=$%d", idx))
		args = append(args, *input.Unit)
		idx++
	}
	if input.MinStock != nil {
		sets = append(sets, fmt.Sprintf("min_stock=$%d", idx))
		args = append(args, *input.MinStock)
		idx++
	}
	if input.IsActive != nil {
		sets = append(sets, fmt.Sprintf("is_active=$%d", idx))
		args = append(args, *input.IsActive)
		idx++
	}

	args = append(args, id)
	query := fmt.Sprintf(`UPDATE products SET %s WHERE id=$%d RETURNING id, code, name, category, unit, min_stock, current_stock, last_cost, is_active, created_at, updated_at`, strings.Join(sets, ", "), idx)

	var item Product
	row := r.pool.QueryRow(ctx, query, args...)
	if err := row.Scan(&item.ID, &item.Code, &item.Name, &item.Category, &item.Unit, &item.MinStock, &item.CurrentStock, &item.LastCost, &item.IsActive, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return item, err
	}
	return item, nil
}

func (r *Repository) UpdateStock(ctx context.Context, id uuid.UUID, quantity float64, cost *float64) error {
	query := `UPDATE products SET current_stock = current_stock + $1, updated_at = now()`
	args := []interface{}{quantity}
	idx := 2
	if cost != nil {
		query += fmt.Sprintf(`, last_cost = $%d`, idx)
		args = append(args, *cost)
		idx++
	}
	query += fmt.Sprintf(` WHERE id = $%d`, idx)
	args = append(args, id)

	_, err := r.pool.Exec(ctx, query, args...)
	return err
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM products WHERE id=$1`, id)
	return err
}

func IsNotFound(err error) bool {
	return err == pgx.ErrNoRows
}
