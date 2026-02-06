package reports

import "time"

type PassengerReportRow struct {
  PassengerID    string    `json:"passenger_id"`
  Name           string    `json:"name"`
  Document       string    `json:"document"`
  Phone          string    `json:"phone"`
  Email          string    `json:"email"`
  SeatNumber     int       `json:"seat_number"`
  BookingStatus  string    `json:"booking_status"`
  PassengerStatus string   `json:"passenger_status"`
  TotalAmount    float64   `json:"total_amount"`
  DepositAmount  float64   `json:"deposit_amount"`
  RemainderAmount float64  `json:"remainder_amount"`
  AmountPaid     float64   `json:"amount_paid"`
  PaymentStage   string    `json:"payment_stage"`
  CreatedAt      time.Time `json:"created_at"`
}
