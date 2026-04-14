package reports

import "time"

type PassengerReportFilter struct {
	TripID          string
	TripDate        string
	BookingID       string
	ReservationCode string
	IncludeCanceled bool
}

type PassengerReportRow struct {
	PassengerID     string    `json:"passenger_id"`
	BookingID       string    `json:"booking_id"`
	ReservationCode string    `json:"reservation_code"`
	TripID          string    `json:"trip_id"`
	TripDate        string    `json:"trip_date"`
	Name            string    `json:"name"`
	Document        string    `json:"document"`
	DocumentType    string    `json:"document_type"`
	Phone           string    `json:"phone"`
	Email           string    `json:"email"`
	Origin          string    `json:"origin"`
	Destination     string    `json:"destination"`
	SeatNumber      string    `json:"seat_number"`
	BookingStatus   string    `json:"booking_status"`
	PassengerStatus string    `json:"passenger_status"`
	TotalAmount     float64   `json:"total_amount"`
	DepositAmount   float64   `json:"deposit_amount"`
	RemainderAmount float64   `json:"remainder_amount"`
	AmountPaid      float64   `json:"amount_paid"`
	PaymentStage    string    `json:"payment_stage"`
	CreatedAt       time.Time `json:"created_at"`
}
