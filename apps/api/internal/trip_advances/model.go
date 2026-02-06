package trip_advances

import "time"

type TripAdvance struct {
  ID          string     `json:"id"`
  TripID      string     `json:"trip_id"`
  DriverID    string     `json:"driver_id"`
  Amount      float64    `json:"amount"`
  Status      string     `json:"status"`
  Purpose     *string    `json:"purpose"`
  DeliveredAt *time.Time `json:"delivered_at"`
  DeliveredBy *string    `json:"delivered_by"`
  SettledAt   *time.Time `json:"settled_at"`
  Notes       *string    `json:"notes"`
  CreatedBy   *string    `json:"created_by"`
  CreatedAt   time.Time  `json:"created_at"`
  UpdatedAt   time.Time  `json:"updated_at"`
}

type CreateTripAdvanceInput struct {
  TripID   string  `json:"trip_id"`
  DriverID string  `json:"driver_id"`
  Amount   float64 `json:"amount"`
  Purpose  *string `json:"purpose"`
  Notes    *string `json:"notes"`
}

type UpdateTripAdvanceInput struct {
  Amount  *float64 `json:"amount"`
  Purpose *string  `json:"purpose"`
  Notes   *string  `json:"notes"`
}

type DeliverAdvanceInput struct {
  DeliveredBy *string `json:"delivered_by"`
}

type ListFilter struct {
  TripID   string
  DriverID string
  Status   string
  Limit    int
  Offset   int
}
