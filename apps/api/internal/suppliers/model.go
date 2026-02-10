package suppliers

import "time"

type Supplier struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Document     *string   `json:"document"`
	Phone        *string   `json:"phone"`
	Email        *string   `json:"email"`
	PaymentTerms *string   `json:"payment_terms"`
	BillingDay   *int      `json:"billing_day"`
	IsActive     bool      `json:"is_active"`
	Notes        *string   `json:"notes"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type CreateSupplierInput struct {
	Name         string  `json:"name"`
	Document     *string `json:"document"`
	Phone        *string `json:"phone"`
	Email        *string `json:"email"`
	PaymentTerms *string `json:"payment_terms"`
	BillingDay   *int    `json:"billing_day"`
	IsActive     *bool   `json:"is_active"`
	Notes        *string `json:"notes"`
}

type UpdateSupplierInput struct {
	Name         *string `json:"name"`
	Document     *string `json:"document"`
	Phone        *string `json:"phone"`
	Email        *string `json:"email"`
	PaymentTerms *string `json:"payment_terms"`
	BillingDay   *int    `json:"billing_day"`
	IsActive     *bool   `json:"is_active"`
	Notes        *string `json:"notes"`
}

type ListFilter struct {
	Limit  int
	Offset int
	Active *bool
}
