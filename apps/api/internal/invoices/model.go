package invoices

import "time"

type InvoiceStatus string

const (
	StatusPending   InvoiceStatus = "PENDING"
	StatusProcessed InvoiceStatus = "PROCESSED"
	StatusCancelled InvoiceStatus = "CANCELLED"
)

type Invoice struct {
	ID              string        `json:"id"`
	InvoiceNumber   string        `json:"invoice_number"`
	Barcode         *string       `json:"barcode"`
	SupplierID      string        `json:"supplier_id"`
	SupplierName    *string       `json:"supplier_name,omitempty"`
	PurchaseOrderID *string       `json:"purchase_order_id"`
	ServiceOrderID  *string       `json:"service_order_id"`
	BusID           *string       `json:"bus_id"`
	BusPlate        *string       `json:"bus_plate,omitempty"`
	IssueDate       time.Time     `json:"issue_date"`
	IssueTime       *string       `json:"issue_time"`
	EntryDate       time.Time     `json:"entry_date"`
	EntryTime       *string       `json:"entry_time"`
	CFOP            string        `json:"cfop"`
	PaymentType     *string       `json:"payment_type"`
	DueDate         *time.Time    `json:"due_date"`
	Subtotal        float64       `json:"subtotal"`
	Discount        float64       `json:"discount"`
	Freight         float64       `json:"freight"`
	Total           float64       `json:"total"`
	Status          InvoiceStatus `json:"status"`
	Notes           *string       `json:"notes"`
	DriverID        *string       `json:"driver_id"`
	DriverName      *string       `json:"driver_name,omitempty"`
	OdometerKm      *int          `json:"odometer_km"`
	CreatedBy       *string       `json:"created_by"`
	CreatedAt       time.Time     `json:"created_at"`
	Items           []InvoiceItem `json:"items,omitempty"`
}

type InvoiceItem struct {
	ID          string    `json:"id"`
	InvoiceID   string    `json:"invoice_id"`
	ProductID   string    `json:"product_id"`
	ProductName *string   `json:"product_name,omitempty"`
	ProductCode *string   `json:"product_code,omitempty"`
	Quantity    float64   `json:"quantity"`
	UnitPrice   float64   `json:"unit_price"`
	Discount    float64   `json:"discount"`
	Total       float64   `json:"total"`
	CreatedAt   time.Time `json:"created_at"`
}

type CreateInvoiceInput struct {
	InvoiceNumber   string            `json:"invoice_number"`
	Barcode         *string           `json:"barcode"`
	SupplierID      string            `json:"supplier_id"`
	PurchaseOrderID *string           `json:"purchase_order_id"`
	ServiceOrderID  *string           `json:"service_order_id"`
	BusID           *string           `json:"bus_id"`
	IssueDate       time.Time         `json:"issue_date"`
	IssueTime       *string           `json:"issue_time"`
	CFOP            *string           `json:"cfop"`
	PaymentType     *string           `json:"payment_type"`
	DueDate         *time.Time        `json:"due_date"`
	Discount        *float64          `json:"discount"`
	Freight         *float64          `json:"freight"`
	Notes           *string           `json:"notes"`
	DriverID        *string           `json:"driver_id"`
	OdometerKm      *int              `json:"odometer_km"`
	Items           []CreateItemInput `json:"items"`
}

type CreateItemInput struct {
	ProductID string   `json:"product_id"`
	Quantity  float64  `json:"quantity"`
	UnitPrice float64  `json:"unit_price"`
	Discount  *float64 `json:"discount"`
}

type UpdateInvoiceInput struct {
	Barcode     *string    `json:"barcode"`
	PaymentType *string    `json:"payment_type"`
	DueDate     *time.Time `json:"due_date"`
	Notes       *string    `json:"notes"`
	DriverID    *string    `json:"driver_id"`
	OdometerKm  *int       `json:"odometer_km"`
}

type AddItemInput struct {
	ProductID string   `json:"product_id"`
	Quantity  float64  `json:"quantity"`
	UnitPrice float64  `json:"unit_price"`
	Discount  *float64 `json:"discount"`
}

type ListFilter struct {
	Limit           int
	Offset          int
	Status          *InvoiceStatus
	SupplierID      *string
	ServiceOrderID  *string
	PurchaseOrderID *string
	BusID           *string
}
