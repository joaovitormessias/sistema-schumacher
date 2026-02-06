package bookings

import "time"

type Booking struct {
  ID              string     `json:"id"`
  TripID          string     `json:"trip_id"`
  Status          string     `json:"status"`
  Source          string     `json:"source"`
  TotalAmount     float64    `json:"total_amount"`
  DepositAmount   float64    `json:"deposit_amount"`
  RemainderAmount float64    `json:"remainder_amount"`
  ExpiresAt       *time.Time `json:"expires_at"`
  CreatedAt       time.Time  `json:"created_at"`
}

type BookingListItem struct {
  ID              string    `json:"id"`
  TripID          string    `json:"trip_id"`
  Status          string    `json:"status"`
  TotalAmount     float64   `json:"total_amount"`
  DepositAmount   float64   `json:"deposit_amount"`
  RemainderAmount float64   `json:"remainder_amount"`
  PassengerName   string    `json:"passenger_name"`
  PassengerPhone  string    `json:"passenger_phone"`
  PassengerEmail  string    `json:"passenger_email"`
  SeatNumber      int       `json:"seat_number"`
  CreatedAt       time.Time `json:"created_at"`
}

type PassengerInput struct {
  Name     string `json:"name"`
  Document string `json:"document"`
  Phone    string `json:"phone"`
  Email    string `json:"email"`
}

type CreateBookingInput struct {
  TripID          string         `json:"trip_id"`
  SeatID          string         `json:"seat_id"`
  BoardStopID     string         `json:"board_stop_id"`
  AlightStopID    string         `json:"alight_stop_id"`
  FareMode        *string        `json:"fare_mode"`
  FareAmountFinal *float64       `json:"fare_amount_final"`
  Passenger       PassengerInput `json:"passenger"`
  Source          *string        `json:"source"`
  TotalAmount     float64        `json:"total_amount"`
  DepositAmount   float64        `json:"deposit_amount"`
  RemainderAmount float64        `json:"remainder_amount"`
}

type UpdateBookingInput struct {
  Status *string `json:"status"`
}

type ListFilter struct {
  Limit  int
  Offset int
}

type BookingPassenger struct {
  ID              string    `json:"id"`
  BookingID       string    `json:"booking_id"`
  TripID          string    `json:"trip_id"`
  Name            string    `json:"name"`
  Document        string    `json:"document"`
  Phone           string    `json:"phone"`
  Email           string    `json:"email"`
  SeatID          string    `json:"seat_id"`
  BoardStopID     string    `json:"board_stop_id"`
  AlightStopID    string    `json:"alight_stop_id"`
  BoardStopOrder  int       `json:"board_stop_order"`
  AlightStopOrder int       `json:"alight_stop_order"`
  FareMode        string    `json:"fare_mode"`
  FareAmountCalc  float64   `json:"fare_amount_calc"`
  FareAmountFinal float64   `json:"fare_amount_final"`
  Status          string    `json:"status"`
  CreatedAt       time.Time `json:"created_at"`
}

type BookingDetails struct {
  Booking   Booking          `json:"booking"`
  Passenger BookingPassenger `json:"passenger"`
}

type CreateBookingData struct {
  TripID          string
  SeatID          string
  BoardStopID     string
  AlightStopID    string
  BoardStopOrder  int
  AlightStopOrder int
  FareMode        string
  FareAmountCalc  float64
  FareAmountFinal float64
  FareSnapshot    []byte
  Passenger       PassengerInput
  Source          *string
  TotalAmount     float64
  DepositAmount   float64
  RemainderAmount float64
}
