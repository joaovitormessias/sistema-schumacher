package trip_expenses

import "time"

type TripExpense struct {
	ID                       string     `json:"id"`
	TripID                   string     `json:"trip_id"`
	DriverID                 string     `json:"driver_id"`
	ExpenseType              string     `json:"expense_type"`
	Amount                   float64    `json:"amount"`
	Description              string     `json:"description"`
	ExpenseDate              time.Time  `json:"expense_date"`
	PaymentMethod            string     `json:"payment_method"`
	DriverCardID             *string    `json:"driver_card_id"`
	ReceiptNumber            *string    `json:"receipt_number"`
	ReceiptVerified          bool       `json:"receipt_verified"`
	ReceiptVerifiedBy        *string    `json:"receipt_verified_by"`
	ReceiptVerifiedAt        *time.Time `json:"receipt_verified_at"`
	ReceiptVerificationNotes *string    `json:"receipt_verification_notes"`
	IsApproved               bool       `json:"is_approved"`
	ApprovedBy               *string    `json:"approved_by"`
	ApprovedAt               *time.Time `json:"approved_at"`
	Notes                    *string    `json:"notes"`
	CreatedBy                *string    `json:"created_by"`
	CreatedAt                time.Time  `json:"created_at"`
	UpdatedAt                time.Time  `json:"updated_at"`
}

type CreateTripExpenseInput struct {
	TripID        string    `json:"trip_id"`
	DriverID      string    `json:"driver_id"`
	ExpenseType   string    `json:"expense_type"`
	Amount        float64   `json:"amount"`
	Description   string    `json:"description"`
	ExpenseDate   time.Time `json:"expense_date"`
	PaymentMethod string    `json:"payment_method"`
	DriverCardID  *string   `json:"driver_card_id"`
	ReceiptNumber *string   `json:"receipt_number"`
	Notes         *string   `json:"notes"`
}

type UpdateTripExpenseInput struct {
	Amount        *float64   `json:"amount"`
	Description   *string    `json:"description"`
	ExpenseDate   *time.Time `json:"expense_date"`
	ReceiptNumber *string    `json:"receipt_number"`
	Notes         *string    `json:"notes"`
}

type ListFilter struct {
	TripID        string
	DriverID      string
	ExpenseType   string
	PaymentMethod string
	Approved      *bool
	Limit         int
	Offset        int
}
