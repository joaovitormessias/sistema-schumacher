package invoices

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
	inv.id, inv.invoice_number, inv.barcode, inv.supplier_id, s.name as supplier_name,
	inv.purchase_order_id, inv.service_order_id, inv.bus_id, b.license_plate as bus_plate,
	inv.issue_date, inv.issue_time::text, inv.entry_date, inv.entry_time::text, inv.cfop,
	inv.payment_type, inv.due_date, inv.subtotal, inv.discount, inv.freight, inv.total,
	inv.status, inv.notes, inv.driver_id, d.name as driver_name, inv.odometer_km,
	inv.created_by, inv.created_at
FROM invoices inv
LEFT JOIN suppliers s ON inv.supplier_id = s.id
LEFT JOIN buses b ON inv.bus_id = b.id
LEFT JOIN drivers d ON inv.driver_id = d.id`

func (r *Repository) List(ctx context.Context, filter ListFilter) ([]Invoice, error) {
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
		conditions = append(conditions, fmt.Sprintf("inv.status = $%d", idx))
		args = append(args, *filter.Status)
		idx++
	}
	if filter.SupplierID != nil {
		conditions = append(conditions, fmt.Sprintf("inv.supplier_id = $%d", idx))
		args = append(args, *filter.SupplierID)
		idx++
	}
	if filter.ServiceOrderID != nil {
		conditions = append(conditions, fmt.Sprintf("inv.service_order_id = $%d", idx))
		args = append(args, *filter.ServiceOrderID)
		idx++
	}
	if filter.PurchaseOrderID != nil {
		conditions = append(conditions, fmt.Sprintf("inv.purchase_order_id = $%d", idx))
		args = append(args, *filter.PurchaseOrderID)
		idx++
	}
	if filter.BusID != nil {
		conditions = append(conditions, fmt.Sprintf("inv.bus_id = $%d", idx))
		args = append(args, *filter.BusID)
		idx++
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += fmt.Sprintf(" ORDER BY inv.entry_date DESC LIMIT $%d OFFSET $%d", idx, idx+1)
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []Invoice{}
	for rows.Next() {
		item, err := scanInvoice(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) Get(ctx context.Context, id uuid.UUID) (Invoice, error) {
	query := baseSelectQuery + " WHERE inv.id = $1"
	row := r.pool.QueryRow(ctx, query, id)
	invoice, err := scanInvoiceRow(row)
	if err != nil {
		return invoice, err
	}

	// Load items
	itemsRows, err := r.pool.Query(ctx,
		`SELECT ii.id, ii.invoice_id, ii.product_id, p.name, p.code, ii.quantity, ii.unit_price, ii.discount, ii.total, ii.created_at
		 FROM invoice_items ii
		 LEFT JOIN products p ON ii.product_id = p.id
		 WHERE ii.invoice_id = $1`, id)
	if err != nil {
		return invoice, err
	}
	defer itemsRows.Close()

	invoice.Items = []InvoiceItem{}
	for itemsRows.Next() {
		var item InvoiceItem
		if err := itemsRows.Scan(&item.ID, &item.InvoiceID, &item.ProductID, &item.ProductName, &item.ProductCode, &item.Quantity, &item.UnitPrice, &item.Discount, &item.Total, &item.CreatedAt); err != nil {
			return invoice, err
		}
		invoice.Items = append(invoice.Items, item)
	}

	return invoice, nil
}

func (r *Repository) Create(ctx context.Context, input CreateInvoiceInput) (Invoice, error) {
	cfop := "1000"
	if input.CFOP != nil {
		cfop = *input.CFOP
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
		return Invoice{}, err
	}
	defer tx.Rollback(ctx)

	var id uuid.UUID
	err = tx.QueryRow(ctx,
		`INSERT INTO invoices (invoice_number, barcode, supplier_id, purchase_order_id, service_order_id, bus_id, issue_date, issue_time, cfop, payment_type, due_date, subtotal, discount, freight, total, notes, driver_id, odometer_km)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18)
		 RETURNING id`,
		input.InvoiceNumber, input.Barcode, input.SupplierID, input.PurchaseOrderID, input.ServiceOrderID, input.BusID, input.IssueDate, input.IssueTime, cfop, input.PaymentType, input.DueDate, subtotal, discount, freight, total, input.Notes, input.DriverID, input.OdometerKm,
	).Scan(&id)
	if err != nil {
		return Invoice{}, err
	}

	// Insert items and update stock
	for _, item := range input.Items {
		itemDiscount := 0.0
		if item.Discount != nil {
			itemDiscount = *item.Discount
		}
		itemTotal := (item.Quantity * item.UnitPrice) - itemDiscount

		_, err = tx.Exec(ctx,
			`INSERT INTO invoice_items (invoice_id, product_id, quantity, unit_price, discount, total)
			 VALUES ($1,$2,$3,$4,$5,$6)`,
			id, item.ProductID, item.Quantity, item.UnitPrice, itemDiscount, itemTotal,
		)
		if err != nil {
			return Invoice{}, err
		}

		// Update product stock and cost
		_, err = tx.Exec(ctx,
			`UPDATE products SET current_stock = current_stock + $1, last_cost = $2, updated_at = now() WHERE id = $3`,
			item.Quantity, item.UnitPrice, item.ProductID,
		)
		if err != nil {
			return Invoice{}, err
		}

		// Record stock movement
		_, err = tx.Exec(ctx,
			`INSERT INTO stock_movements (product_id, movement_type, quantity, unit_cost, reference_type, reference_id, notes)
			 VALUES ($1, 'IN', $2, $3, 'INVOICE', $4, $5)`,
			item.ProductID, item.Quantity, item.UnitPrice, id, "Entrada por NF",
		)
		if err != nil {
			return Invoice{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return Invoice{}, err
	}

	return r.Get(ctx, id)
}

func (r *Repository) Update(ctx context.Context, id uuid.UUID, input UpdateInvoiceInput) (Invoice, error) {
	sets := []string{}
	args := []interface{}{}
	idx := 1

	if input.Barcode != nil {
		sets = append(sets, fmt.Sprintf("barcode=$%d", idx))
		args = append(args, *input.Barcode)
		idx++
	}
	if input.PaymentType != nil {
		sets = append(sets, fmt.Sprintf("payment_type=$%d", idx))
		args = append(args, *input.PaymentType)
		idx++
	}
	if input.DueDate != nil {
		sets = append(sets, fmt.Sprintf("due_date=$%d", idx))
		args = append(args, *input.DueDate)
		idx++
	}
	if input.Notes != nil {
		sets = append(sets, fmt.Sprintf("notes=$%d", idx))
		args = append(args, *input.Notes)
		idx++
	}
	if input.DriverID != nil {
		sets = append(sets, fmt.Sprintf("driver_id=$%d", idx))
		args = append(args, *input.DriverID)
		idx++
	}
	if input.OdometerKm != nil {
		sets = append(sets, fmt.Sprintf("odometer_km=$%d", idx))
		args = append(args, *input.OdometerKm)
		idx++
	}

	if len(sets) == 0 {
		return r.Get(ctx, id)
	}

	args = append(args, id)
	query := fmt.Sprintf(`UPDATE invoices SET %s WHERE id=$%d`, strings.Join(sets, ", "), idx)
	_, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return Invoice{}, err
	}

	return r.Get(ctx, id)
}

func (r *Repository) SetStatus(ctx context.Context, id uuid.UUID, status InvoiceStatus) error {
	_, err := r.pool.Exec(ctx, `UPDATE invoices SET status = $1 WHERE id = $2`, status, id)
	return err
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM invoices WHERE id=$1`, id)
	return err
}

func scanInvoice(rows pgx.Rows) (Invoice, error) {
	var item Invoice
	err := rows.Scan(
		&item.ID, &item.InvoiceNumber, &item.Barcode, &item.SupplierID, &item.SupplierName,
		&item.PurchaseOrderID, &item.ServiceOrderID, &item.BusID, &item.BusPlate,
		&item.IssueDate, &item.IssueTime, &item.EntryDate, &item.EntryTime, &item.CFOP,
		&item.PaymentType, &item.DueDate, &item.Subtotal, &item.Discount, &item.Freight, &item.Total,
		&item.Status, &item.Notes, &item.DriverID, &item.DriverName, &item.OdometerKm,
		&item.CreatedBy, &item.CreatedAt,
	)
	return item, err
}

func scanInvoiceRow(row pgx.Row) (Invoice, error) {
	var item Invoice
	err := row.Scan(
		&item.ID, &item.InvoiceNumber, &item.Barcode, &item.SupplierID, &item.SupplierName,
		&item.PurchaseOrderID, &item.ServiceOrderID, &item.BusID, &item.BusPlate,
		&item.IssueDate, &item.IssueTime, &item.EntryDate, &item.EntryTime, &item.CFOP,
		&item.PaymentType, &item.DueDate, &item.Subtotal, &item.Discount, &item.Freight, &item.Total,
		&item.Status, &item.Notes, &item.DriverID, &item.DriverName, &item.OdometerKm,
		&item.CreatedBy, &item.CreatedAt,
	)
	return item, err
}

func IsNotFound(err error) bool {
	return err == pgx.ErrNoRows
}
