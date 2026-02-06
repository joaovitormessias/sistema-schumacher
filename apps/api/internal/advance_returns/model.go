package advance_returns

import "time"

type AdvanceReturn struct {
  ID              string    `json:"id"`
  TripAdvanceID   string    `json:"trip_advance_id"`
  TripSettlementID *string  `json:"trip_settlement_id"`
  Amount          float64   `json:"amount"`
  ReturnDate      time.Time `json:"return_date"`
  PaymentMethod   string    `json:"payment_method"`
  ReceivedBy      *string   `json:"received_by"`
  Notes           *string   `json:"notes"`
  CreatedAt       time.Time `json:"created_at"`
}

type CreateAdvanceReturnInput struct {
  TripAdvanceID    string     `json:"trip_advance_id"`
  TripSettlementID *string    `json:"trip_settlement_id"`
  Amount           float64    `json:"amount"`
  ReturnDate       *time.Time `json:"return_date"`
  PaymentMethod    string     `json:"payment_method"`
  Notes            *string    `json:"notes"`
}

type ListFilter struct {
  TripAdvanceID string
  TripSettlementID string
  Limit         int
  Offset        int
}
