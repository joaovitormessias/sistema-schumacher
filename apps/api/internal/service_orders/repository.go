package service_orders

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

const baseSelectQuery = `SELECT 
	so.id, so.order_number, so.bus_id, b.license_plate as bus_plate, so.driver_id, d.name as driver_name,
	so.order_type, so.status, so.description, so.odometer_km, so.scheduled_date, so.location,
	so.opened_at, so.closed_at, so.closed_odometer_km, so.next_preventive_km, so.notes,
	so.created_by, so.created_at, so.updated_at
FROM service_orders so
LEFT JOIN buses b ON so.bus_id = b.id
LEFT JOIN drivers d ON so.driver_id = d.id`

func (r *Repository) List(ctx context.Context, filter ListFilter) ([]ServiceOrder, error) {
	limit := filter.Limit
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	query := baseSelectQuery
	args := []interface{}{}
	conditions := []string{}
	idx := 1

	if filter.Status != nil {
		conditions = append(conditions, fmt.Sprintf("so.status = $%d", idx))
		args = append(args, *filter.Status)
		idx++
	}
	if filter.OrderType != nil {
		conditions = append(conditions, fmt.Sprintf("so.order_type = $%d", idx))
		args = append(args, *filter.OrderType)
		idx++
	}
	if filter.BusID != nil {
		conditions = append(conditions, fmt.Sprintf("so.bus_id = $%d", idx))
		args = append(args, *filter.BusID)
		idx++
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += fmt.Sprintf(" ORDER BY so.opened_at DESC LIMIT $%d OFFSET $%d", idx, idx+1)
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []ServiceOrder{}
	for rows.Next() {
		item, err := scanServiceOrder(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) Get(ctx context.Context, id uuid.UUID) (ServiceOrder, error) {
	query := baseSelectQuery + " WHERE so.id = $1"
	row := r.pool.QueryRow(ctx, query, id)
	return scanServiceOrderRow(row)
}

func (r *Repository) Create(ctx context.Context, input CreateServiceOrderInput) (ServiceOrder, error) {
	location := "SCHUMACHER"
	if input.Location != nil {
		location = *input.Location
	}

	var id uuid.UUID
	err := r.pool.QueryRow(ctx,
		`INSERT INTO service_orders (bus_id, driver_id, order_type, description, odometer_km, scheduled_date, location, notes)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		 RETURNING id`,
		input.BusID, input.DriverID, input.OrderType, input.Description, input.OdometerKm, input.ScheduledDate, location, input.Notes,
	).Scan(&id)
	if err != nil {
		return ServiceOrder{}, err
	}
	return r.Get(ctx, id)
}

func (r *Repository) Update(ctx context.Context, id uuid.UUID, input UpdateServiceOrderInput) (ServiceOrder, error) {
	sets := []string{"updated_at = now()"}
	args := []interface{}{}
	idx := 1

	if input.DriverID != nil {
		sets = append(sets, fmt.Sprintf("driver_id=$%d", idx))
		args = append(args, *input.DriverID)
		idx++
	}
	if input.Description != nil {
		sets = append(sets, fmt.Sprintf("description=$%d", idx))
		args = append(args, *input.Description)
		idx++
	}
	if input.OdometerKm != nil {
		sets = append(sets, fmt.Sprintf("odometer_km=$%d", idx))
		args = append(args, *input.OdometerKm)
		idx++
	}
	if input.Location != nil {
		sets = append(sets, fmt.Sprintf("location=$%d", idx))
		args = append(args, *input.Location)
		idx++
	}
	if input.Notes != nil {
		sets = append(sets, fmt.Sprintf("notes=$%d", idx))
		args = append(args, *input.Notes)
		idx++
	}
	if input.NextPreventiveKm != nil {
		sets = append(sets, fmt.Sprintf("next_preventive_km=$%d", idx))
		args = append(args, *input.NextPreventiveKm)
		idx++
	}

	args = append(args, id)
	query := fmt.Sprintf(`UPDATE service_orders SET %s WHERE id=$%d`, strings.Join(sets, ", "), idx)

	_, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return ServiceOrder{}, err
	}
	return r.Get(ctx, id)
}

func (r *Repository) SetStatus(ctx context.Context, id uuid.UUID, status ServiceOrderStatus) error {
	_, err := r.pool.Exec(ctx, `UPDATE service_orders SET status = $1, updated_at = now() WHERE id = $2`, status, id)
	return err
}

func (r *Repository) Close(ctx context.Context, id uuid.UUID, input CloseServiceOrderInput) (ServiceOrder, error) {
	now := time.Now()
	_, err := r.pool.Exec(ctx,
		`UPDATE service_orders SET status = 'CLOSED', closed_at = $1, closed_odometer_km = $2, next_preventive_km = $3, updated_at = now() WHERE id = $4`,
		now, input.ClosedOdometerKm, input.NextPreventiveKm, id,
	)
	if err != nil {
		return ServiceOrder{}, err
	}
	return r.Get(ctx, id)
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM service_orders WHERE id=$1`, id)
	return err
}

func scanServiceOrder(rows pgx.Rows) (ServiceOrder, error) {
	var item ServiceOrder
	err := rows.Scan(
		&item.ID, &item.OrderNumber, &item.BusID, &item.BusPlate, &item.DriverID, &item.DriverName,
		&item.OrderType, &item.Status, &item.Description, &item.OdometerKm, &item.ScheduledDate, &item.Location,
		&item.OpenedAt, &item.ClosedAt, &item.ClosedOdometerKm, &item.NextPreventiveKm, &item.Notes,
		&item.CreatedBy, &item.CreatedAt, &item.UpdatedAt,
	)
	return item, err
}

func scanServiceOrderRow(row pgx.Row) (ServiceOrder, error) {
	var item ServiceOrder
	err := row.Scan(
		&item.ID, &item.OrderNumber, &item.BusID, &item.BusPlate, &item.DriverID, &item.DriverName,
		&item.OrderType, &item.Status, &item.Description, &item.OdometerKm, &item.ScheduledDate, &item.Location,
		&item.OpenedAt, &item.ClosedAt, &item.ClosedOdometerKm, &item.NextPreventiveKm, &item.Notes,
		&item.CreatedBy, &item.CreatedAt, &item.UpdatedAt,
	)
	return item, err
}

func IsNotFound(err error) bool {
	return err == pgx.ErrNoRows
}
