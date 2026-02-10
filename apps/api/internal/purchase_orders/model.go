package purchase_orders

import "time"

type PurchaseOrderStatus string

const (
	StatusDraft     PurchaseOrderStatus = "DRAFT"
	StatusSent      PurchaseOrderStatus = "SENT"
	StatusPartial   PurchaseOrderStatus = "PARTIAL"
	StatusReceived  PurchaseOrderStatus = "RECEIVED"
	StatusCancelled PurchaseOrderStatus = "CANCELLED"
)

type PurchaseOrder struct {
	ID               string              `json:"id"`
	OrderNumber      int                 `json:"order_number"`
	ServiceOrderID   *string             `json:"service_order_id"`
	SupplierID       string              `json:"supplier_id"`
	SupplierName     *string             `json:"supplier_name,omitempty"`
	Status           PurchaseOrderStatus `json:"status"`
	OrderDate        time.Time           `json:"order_date"`
	ExpectedDelivery *time.Time          `json:"expected_delivery"`
	OwnDelivery      bool                `json:"own_delivery"`
	Subtotal         float64             `json:"subtotal"`
	Discount         float64             `json:"discount"`
	Freight          float64             `json:"freight"`
	Total            float64             `json:"total"`
	Notes            *string             `json:"notes"`
	CreatedBy        *string             `json:"created_by"`
	CreatedAt        time.Time           `json:"created_at"`
	UpdatedAt        time.Time           `json:"updated_at"`
	Items            []PurchaseOrderItem `json:"items,omitempty"`
}

type PurchaseOrderItem struct {
	ID               string    `json:"id"`
	PurchaseOrderID  string    `json:"purchase_order_id"`
	ProductID        string    `json:"product_id"`
	ProductName      *string   `json:"product_name,omitempty"`
	ProductCode      *string   `json:"product_code,omitempty"`
	Quantity         float64   `json:"quantity"`
	UnitPrice        float64   `json:"unit_price"`
	Discount         float64   `json:"discount"`
	Total            float64   `json:"total"`
	ReceivedQuantity float64   `json:"received_quantity"`
	CreatedAt        time.Time `json:"created_at"`
}

type CreatePurchaseOrderInput struct {
	ServiceOrderID   *string                `json:"service_order_id"`
	SupplierID       string                 `json:"supplier_id"`
	ExpectedDelivery *time.Time             `json:"expected_delivery"`
	OwnDelivery      *bool                  `json:"own_delivery"`
	Discount         *float64               `json:"discount"`
	Freight          *float64               `json:"freight"`
	Notes            *string                `json:"notes"`
	Items            []CreateOrderItemInput `json:"items"`
}

type CreateOrderItemInput struct {
	ProductID string   `json:"product_id"`
	Quantity  float64  `json:"quantity"`
	UnitPrice float64  `json:"unit_price"`
	Discount  *float64 `json:"discount"`
}

type UpdatePurchaseOrderInput struct {
	ExpectedDelivery *time.Time `json:"expected_delivery"`
	OwnDelivery      *bool      `json:"own_delivery"`
	Discount         *float64   `json:"discount"`
	Freight          *float64   `json:"freight"`
	Notes            *string    `json:"notes"`
}

type AddItemInput struct {
	ProductID string   `json:"product_id"`
	Quantity  float64  `json:"quantity"`
	UnitPrice float64  `json:"unit_price"`
	Discount  *float64 `json:"discount"`
}

type ListFilter struct {
	Limit          int
	Offset         int
	Status         *PurchaseOrderStatus
	SupplierID     *string
	ServiceOrderID *string
}
