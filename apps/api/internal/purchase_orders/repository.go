package purchase_orders

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

const baseSelectQuery = `SELECT 
	po.id, po.order_number, po.service_order_id, po.supplier_id, s.name as supplier_name,
	po.status, po.order_date, po.expected_delivery, po.own_delivery,
	po.subtotal, po.discount, po.freight, po.total, po.notes,
	po.created_by, po.created_at, po.updated_at
FROM purchase_orders po
LEFT JOIN suppliers s ON po.supplier_id = s.id`

func (r *Repository) List(ctx context.Context, filter ListFilter) ([]PurchaseOrder, error) {
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
		conditions = append(conditions, fmt.Sprintf("po.status = $%d", idx))
		args = append(args, *filter.Status)
		idx++
	}
	if filter.SupplierID != nil {
		conditions = append(conditions, fmt.Sprintf("po.supplier_id = $%d", idx))
		args = append(args, *filter.SupplierID)
		idx++
	}
	if filter.ServiceOrderID != nil {
		conditions = append(conditions, fmt.Sprintf("po.service_order_id = $%d", idx))
		args = append(args, *filter.ServiceOrderID)
		idx++
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += fmt.Sprintf(" ORDER BY po.order_date DESC LIMIT $%d OFFSET $%d", idx, idx+1)
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []PurchaseOrder{}
	for rows.Next() {
		item, err := scanPurchaseOrder(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) Get(ctx context.Context, id uuid.UUID) (PurchaseOrder, error) {
	query := baseSelectQuery + " WHERE po.id = $1"
	row := r.pool.QueryRow(ctx, query, id)
	order, err := scanPurchaseOrderRow(row)
	if err != nil {
		return order, err
	}

	// Load items
	itemsRows, err := r.pool.Query(ctx,
		`SELECT poi.id, poi.purchase_order_id, poi.product_id, p.name, p.code, poi.quantity, poi.unit_price, poi.discount, poi.total, poi.received_quantity, poi.created_at
		 FROM purchase_order_items poi
		 LEFT JOIN products p ON poi.product_id = p.id
		 WHERE poi.purchase_order_id = $1`, id)
	if err != nil {
		return order, err
	}
	defer itemsRows.Close()

	order.Items = []PurchaseOrderItem{}
	for itemsRows.Next() {
		var item PurchaseOrderItem
		if err := itemsRows.Scan(&item.ID, &item.PurchaseOrderID, &item.ProductID, &item.ProductName, &item.ProductCode, &item.Quantity, &item.UnitPrice, &item.Discount, &item.Total, &item.ReceivedQuantity, &item.CreatedAt); err != nil {
			return order, err
		}
		order.Items = append(order.Items, item)
	}

	return order, nil
}

func (r *Repository) Create(ctx context.Context, input CreatePurchaseOrderInput) (PurchaseOrder, error) {
	ownDelivery := true
	if input.OwnDelivery != nil {
		ownDelivery = *input.OwnDelivery
	}
	discount := 0.0
	if input.Discount != nil {
		discount = *input.Discount
	}
	freight := 0.0
	if input.Freight != nil {
		freight = *input.Freight
	}

	// Calculate totals
	subtotal := 0.0
	for _, item := range input.Items {
		itemDiscount := 0.0
		if item.Discount != nil {
			itemDiscount = *item.Discount
		}
		subtotal += (item.Quantity * item.UnitPrice) - itemDiscount
	}
	total := subtotal - discount + freight

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return PurchaseOrder{}, err
	}
	defer tx.Rollback(ctx)

	var id uuid.UUID
	err = tx.QueryRow(ctx,
		`INSERT INTO purchase_orders (service_order_id, supplier_id, expected_delivery, own_delivery, subtotal, discount, freight, total, notes)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		 RETURNING id`,
		input.ServiceOrderID, input.SupplierID, input.ExpectedDelivery, ownDelivery, subtotal, discount, freight, total, input.Notes,
	).Scan(&id)
	if err != nil {
		return PurchaseOrder{}, err
	}

	// Insert items
	for _, item := range input.Items {
		itemDiscount := 0.0
		if item.Discount != nil {
			itemDiscount = *item.Discount
		}
		itemTotal := (item.Quantity * item.UnitPrice) - itemDiscount
		_, err = tx.Exec(ctx,
			`INSERT INTO purchase_order_items (purchase_order_id, product_id, quantity, unit_price, discount, total)
			 VALUES ($1,$2,$3,$4,$5,$6)`,
			id, item.ProductID, item.Quantity, item.UnitPrice, itemDiscount, itemTotal,
		)
		if err != nil {
			return PurchaseOrder{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return PurchaseOrder{}, err
	}

	return r.Get(ctx, id)
}

func (r *Repository) Update(ctx context.Context, id uuid.UUID, input UpdatePurchaseOrderInput) (PurchaseOrder, error) {
	sets := []string{"updated_at = now()"}
	args := []interface{}{}
	idx := 1

	if input.ExpectedDelivery != nil {
		sets = append(sets, fmt.Sprintf("expected_delivery=$%d", idx))
		args = append(args, *input.ExpectedDelivery)
		idx++
	}
	if input.OwnDelivery != nil {
		sets = append(sets, fmt.Sprintf("own_delivery=$%d", idx))
		args = append(args, *input.OwnDelivery)
		idx++
	}
	if input.Discount != nil {
		sets = append(sets, fmt.Sprintf("discount=$%d", idx))
		args = append(args, *input.Discount)
		idx++
	}
	if input.Freight != nil {
		sets = append(sets, fmt.Sprintf("freight=$%d", idx))
		args = append(args, *input.Freight)
		idx++
	}
	if input.Notes != nil {
		sets = append(sets, fmt.Sprintf("notes=$%d", idx))
		args = append(args, *input.Notes)
		idx++
	}

	args = append(args, id)
	query := fmt.Sprintf(`UPDATE purchase_orders SET %s WHERE id=$%d`, strings.Join(sets, ", "), idx)
	_, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return PurchaseOrder{}, err
	}

	// Recalculate total
	r.recalculateTotal(ctx, id)

	return r.Get(ctx, id)
}

func (r *Repository) AddItem(ctx context.Context, orderID uuid.UUID, input AddItemInput) (PurchaseOrder, error) {
	discount := 0.0
	if input.Discount != nil {
		discount = *input.Discount
	}
	total := (input.Quantity * input.UnitPrice) - discount

	_, err := r.pool.Exec(ctx,
		`INSERT INTO purchase_order_items (purchase_order_id, product_id, quantity, unit_price, discount, total)
		 VALUES ($1,$2,$3,$4,$5,$6)`,
		orderID, input.ProductID, input.Quantity, input.UnitPrice, discount, total,
	)
	if err != nil {
		return PurchaseOrder{}, err
	}

	r.recalculateTotal(ctx, orderID)
	return r.Get(ctx, orderID)
}

func (r *Repository) RemoveItem(ctx context.Context, orderID, itemID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM purchase_order_items WHERE id = $1 AND purchase_order_id = $2`, itemID, orderID)
	if err != nil {
		return err
	}
	return r.recalculateTotal(ctx, orderID)
}

func (r *Repository) SetStatus(ctx context.Context, id uuid.UUID, status PurchaseOrderStatus) error {
	_, err := r.pool.Exec(ctx, `UPDATE purchase_orders SET status = $1, updated_at = now() WHERE id = $2`, status, id)
	return err
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM purchase_orders WHERE id=$1`, id)
	return err
}

func (r *Repository) recalculateTotal(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE purchase_orders 
		 SET subtotal = COALESCE((SELECT SUM(total) FROM purchase_order_items WHERE purchase_order_id = $1), 0),
		     total = COALESCE((SELECT SUM(total) FROM purchase_order_items WHERE purchase_order_id = $1), 0) - discount + freight,
		     updated_at = now()
		 WHERE id = $1`, id)
	return err
}

func scanPurchaseOrder(rows pgx.Rows) (PurchaseOrder, error) {
	var item PurchaseOrder
	err := rows.Scan(
		&item.ID, &item.OrderNumber, &item.ServiceOrderID, &item.SupplierID, &item.SupplierName,
		&item.Status, &item.OrderDate, &item.ExpectedDelivery, &item.OwnDelivery,
		&item.Subtotal, &item.Discount, &item.Freight, &item.Total, &item.Notes,
		&item.CreatedBy, &item.CreatedAt, &item.UpdatedAt,
	)
	return item, err
}

func scanPurchaseOrderRow(row pgx.Row) (PurchaseOrder, error) {
	var item PurchaseOrder
	err := row.Scan(
		&item.ID, &item.OrderNumber, &item.ServiceOrderID, &item.SupplierID, &item.SupplierName,
		&item.Status, &item.OrderDate, &item.ExpectedDelivery, &item.OwnDelivery,
		&item.Subtotal, &item.Discount, &item.Freight, &item.Total, &item.Notes,
		&item.CreatedBy, &item.CreatedAt, &item.UpdatedAt,
	)
	return item, err
}

func IsNotFound(err error) bool {
	return err == pgx.ErrNoRows
}
