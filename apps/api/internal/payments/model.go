package payments

import (
  "encoding/json"
  "time"
)

type Payment struct {
  ID         string          `json:"id"`
  BookingID  string          `json:"booking_id"`
  Amount     float64         `json:"amount"`
  Method     string          `json:"method"`
  Status     string          `json:"status"`
  Provider   *string         `json:"provider"`
  ProviderRef *string        `json:"provider_ref"`
  PaidAt     *time.Time      `json:"paid_at"`
  Metadata   json.RawMessage `json:"metadata"`
  CreatedAt  time.Time       `json:"created_at"`
}

type CreatePaymentInput struct {
  BookingID   string  `json:"booking_id"`
  Amount      float64 `json:"amount"`
  Method      string  `json:"method"` // PIX or CARD
  Description string  `json:"description"`
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
  Method    string  `json:"method"` // CASH, TRANSFER, OTHER
  Notes     string  `json:"notes"`
}

type PaymentStatusResponse struct {
  ID        string  `json:"id"`
  Status    string  `json:"status"`
  Amount    float64 `json:"amount"`
  Provider  *string `json:"provider"`
  ProviderRef *string `json:"provider_ref"`
  Metadata  json.RawMessage `json:"metadata"`
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
