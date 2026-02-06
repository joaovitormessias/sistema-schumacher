package trips

import "time"

type Trip struct {
  ID          string     `json:"id"`
  RouteID     string     `json:"route_id"`
  BusID       string     `json:"bus_id"`
  DriverID    *string    `json:"driver_id"`
  FareID      *string    `json:"fare_id"`
  DepartureAt time.Time  `json:"departure_at"`
  ArrivalAt   *time.Time `json:"arrival_at"`
  Status      string     `json:"status"`
  PairTripID  *string    `json:"pair_trip_id"`
  Notes       *string    `json:"notes"`
  CreatedAt   time.Time  `json:"created_at"`
}

type TripSeat struct {
  ID         string `json:"id"`
  SeatNumber int    `json:"seat_number"`
  IsActive   bool   `json:"is_active"`
  IsTaken    bool   `json:"is_taken"`
}

type TripStop struct {
  ID          string     `json:"id"`
  TripID      string     `json:"trip_id"`
  RouteStopID string     `json:"route_stop_id"`
  City        string     `json:"city"`
  StopOrder   int        `json:"stop_order"`
  ArriveAt    *time.Time `json:"arrive_at"`
  DepartAt    *time.Time `json:"depart_at"`
  CreatedAt   time.Time  `json:"created_at"`
}

type CreateTripStopInput struct {
  RouteStopID string     `json:"route_stop_id"`
  ArriveAt    *time.Time `json:"arrive_at"`
  DepartAt    *time.Time `json:"depart_at"`
}

type CreateTripInput struct {
  RouteID     string     `json:"route_id"`
  BusID       string     `json:"bus_id"`
  DriverID    *string    `json:"driver_id"`
  FareID      *string    `json:"fare_id"`
  DepartureAt time.Time  `json:"departure_at"`
  ArrivalAt   *time.Time `json:"arrival_at"`
  Status      *string    `json:"status"`
  PairTripID  *string    `json:"pair_trip_id"`
  Notes       *string    `json:"notes"`
}

type UpdateTripInput struct {
  RouteID     *string    `json:"route_id"`
  BusID       *string    `json:"bus_id"`
  DriverID    *string    `json:"driver_id"`
  FareID      *string    `json:"fare_id"`
  DepartureAt *time.Time `json:"departure_at"`
  ArrivalAt   *time.Time `json:"arrival_at"`
  Status      *string    `json:"status"`
  PairTripID  *string    `json:"pair_trip_id"`
  Notes       *string    `json:"notes"`
}

type ListFilter struct {
  Limit  int
  Offset int
}
