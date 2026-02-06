package trip_settlements

import "time"

type TripSettlement struct {
  ID                string     `json:"id"`
  TripID            string     `json:"trip_id"`
  DriverID          string     `json:"driver_id"`
  Status            string     `json:"status"`
  AdvanceAmount     float64    `json:"advance_amount"`
  ExpensesTotal     float64    `json:"expenses_total"`
  Balance           float64    `json:"balance"`
  AmountToReturn    float64    `json:"amount_to_return"`
  AmountToReimburse float64    `json:"amount_to_reimburse"`
  ReviewedBy        *string    `json:"reviewed_by"`
  ReviewedAt        *time.Time `json:"reviewed_at"`
  ApprovedBy        *string    `json:"approved_by"`
  ApprovedAt        *time.Time `json:"approved_at"`
  CompletedAt       *time.Time `json:"completed_at"`
  Notes             *string    `json:"notes"`
  CreatedBy         *string    `json:"created_by"`
  CreatedAt         time.Time  `json:"created_at"`
  UpdatedAt         time.Time  `json:"updated_at"`
}

type CreateTripSettlementInput struct {
  TripID string  `json:"trip_id"`
  Notes  *string `json:"notes"`
}

type ListFilter struct {
  TripID   string
  DriverID string
  Status   string
  Limit    int
  Offset   int
}
