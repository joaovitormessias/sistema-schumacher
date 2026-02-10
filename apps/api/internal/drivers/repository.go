package drivers

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

func (r *Repository) List(ctx context.Context, filter ListFilter) ([]Driver, error) {
	limit := filter.Limit
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	query := `select id, name, document, phone, is_active, created_at
     from drivers`
	args := []interface{}{}
	clauses := []string{}

	if filter.Search != "" {
		args = append(args, "%"+filter.Search+"%")
		clauses = append(clauses, fmt.Sprintf(`(
      name ilike $%d
      or document ilike $%d
      or phone ilike $%d
      or id::text ilike $%d
    )`, len(args), len(args), len(args), len(args)))
	}

	if len(clauses) > 0 {
		query += " where " + strings.Join(clauses, " and ")
	}

	query += " order by created_at desc"
	args = append(args, limit)
	query += fmt.Sprintf(" limit $%d", len(args))
	args = append(args, offset)
	query += fmt.Sprintf(" offset $%d", len(args))

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []Driver{}
	for rows.Next() {
		var item Driver
		if err := rows.Scan(&item.ID, &item.Name, &item.Document, &item.Phone, &item.IsActive, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) Get(ctx context.Context, id uuid.UUID) (Driver, error) {
	var item Driver
	row := r.pool.QueryRow(ctx, `select id, name, document, phone, is_active, created_at from drivers where id=$1`, id)
	if err := row.Scan(&item.ID, &item.Name, &item.Document, &item.Phone, &item.IsActive, &item.CreatedAt); err != nil {
		return item, err
	}
	return item, nil
}

func (r *Repository) Create(ctx context.Context, input CreateDriverInput) (Driver, error) {
	isActive := true
	if input.IsActive != nil {
		isActive = *input.IsActive
	}

	var item Driver
	row := r.pool.QueryRow(ctx,
		`insert into drivers (name, document, phone, is_active)
     values ($1,$2,$3,$4)
     returning id, name, document, phone, is_active, created_at`,
		input.Name, input.Document, input.Phone, isActive,
	)
	if err := row.Scan(&item.ID, &item.Name, &item.Document, &item.Phone, &item.IsActive, &item.CreatedAt); err != nil {
		return item, err
	}
	return item, nil
}

func (r *Repository) Update(ctx context.Context, id uuid.UUID, input UpdateDriverInput) (Driver, error) {
	sets := []string{}
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
	if input.IsActive != nil {
		sets = append(sets, fmt.Sprintf("is_active=$%d", idx))
		args = append(args, *input.IsActive)
		idx++
	}

	if len(sets) == 0 {
		return r.Get(ctx, id)
	}

	args = append(args, id)
	query := fmt.Sprintf(`update drivers set %s where id=$%d returning id, name, document, phone, is_active, created_at`, strings.Join(sets, ", "), idx)

	var item Driver
	row := r.pool.QueryRow(ctx, query, args...)
	if err := row.Scan(&item.ID, &item.Name, &item.Document, &item.Phone, &item.IsActive, &item.CreatedAt); err != nil {
		return item, err
	}
	return item, nil
}

func IsNotFound(err error) bool {
	return err == pgx.ErrNoRows
}
