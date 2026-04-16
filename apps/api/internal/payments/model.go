package payments

import (
	"encoding/json"
	"time"
)

type Payment struct {
	ID          string          `json:"id"`
	BookingID   string          `json:"booking_id"`
	Amount      float64         `json:"amount"`
	Method      string          `json:"method"`
	Status      string          `json:"status"`
	Provider    *string         `json:"provider"`
	ProviderRef *string         `json:"provider_ref"`
	PaidAt      *time.Time      `json:"paid_at"`
	Metadata    json.RawMessage `json:"metadata"`
	CreatedAt   time.Time       `json:"created_at"`
}

type CreatePaymentInput struct {
	BookingID   string         `json:"booking_id"`
	Amount      float64        `json:"amount"`
	Method      string         `json:"method"`
	Description string         `json:"description"`
	Customer    *CustomerInput `json:"customer"`
}

type CustomerInput struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Document string `json:"document"`
}

type ManualPaymentInput struct {
	BookingID string  `json:"booking_id"`
	Amount    float64 `json:"amount"`
	Method    string  `json:"method"`
	Notes     string  `json:"notes"`
}

type PaymentStatusResponse struct {
	ID          string          `json:"id"`
	Status      string          `json:"status"`
	Amount      float64         `json:"amount"`
	Provider    *string         `json:"provider"`
	ProviderRef *string         `json:"provider_ref"`
	Metadata    json.RawMessage `json:"metadata"`
}

type CreatePaymentResponse struct {
	Payment     Payment     `json:"payment"`
	ProviderRaw interface{} `json:"provider_raw,omitempty"`
	CheckoutURL *string     `json:"checkout_url"`
	PixCode     *string     `json:"pix_code"`
}

type PaymentSyncResponse struct {
	Payment       Payment `json:"payment"`
	BookingStatus string  `json:"booking_status"`
	Synced        bool    `json:"synced"`
}

type PaymentListFilter struct {
	BookingID string
	Status    string
	Since     *time.Time
	Until     *time.Time
	PaidSince *time.Time
	PaidUntil *time.Time
	Limit     int
	Offset    int
}

type PaymentNotificationContext struct {
	BookingID       string  `json:"booking_id"`
	ReservationCode string  `json:"reservation_code"`
	CustomerName    string  `json:"customer_name"`
	CustomerPhone   string  `json:"customer_phone"`
	AmountTotal     float64 `json:"amount_total"`
	AmountPaid      float64 `json:"amount_paid"`
	AmountDue       float64 `json:"amount_due"`
	PaymentStatus   string  `json:"payment_status"`
}

type PaymentNotificationPayload struct {
	Event           string    `json:"event"`
	SentAt          time.Time `json:"sent_at"`
	PaymentID       string    `json:"payment_id"`
	PaymentAmount   float64   `json:"payment_amount"`
	PaymentMethod   string    `json:"payment_method"`
	BookingID       string    `json:"booking_id"`
	ReservationCode string    `json:"reservation_code"`
	CustomerName    string    `json:"customer_name"`
	CustomerPhone   string    `json:"customer_phone"`
	AmountTotal     float64   `json:"amount_total"`
	AmountPaid      float64   `json:"amount_paid"`
	AmountDue       float64   `json:"amount_due"`
	PaymentStatus   string    `json:"payment_status"`
}
