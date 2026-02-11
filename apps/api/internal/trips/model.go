package trips

import "time"

type Trip struct {
	ID                  string     `json:"id"`
	RouteID             string     `json:"route_id"`
	BusID               string     `json:"bus_id"`
	DriverID            *string    `json:"driver_id"`
	FareID              *string    `json:"fare_id"`
	RequestID           *string    `json:"request_id"`
	DepartureAt         time.Time  `json:"departure_at"`
	ArrivalAt           *time.Time `json:"arrival_at"`
	Status              string     `json:"status"`
	OperationalStatus   string     `json:"operational_status"`
	EstimatedKM         float64    `json:"estimated_km"`
	DispatchValidatedAt *time.Time `json:"dispatch_validated_at"`
	DispatchValidatedBy *string    `json:"dispatch_validated_by"`
	PairTripID          *string    `json:"pair_trip_id"`
	Notes               *string    `json:"notes"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

type TripSeat struct {
	ID         string `json:"id"`
	SeatNumber int    `json:"seat_number"`
	IsActive   bool   `json:"is_active"`
	IsTaken    bool   `json:"is_taken"`
}

type TripStop struct {
	ID                   string     `json:"id"`
	TripID               string     `json:"trip_id"`
	RouteStopID          string     `json:"route_stop_id"`
	City                 string     `json:"city"`
	StopOrder            int        `json:"stop_order"`
	LegDistanceKM        float64    `json:"leg_distance_km"`
	CumulativeDistanceKM float64    `json:"cumulative_distance_km"`
	ArriveAt             *time.Time `json:"arrive_at"`
	DepartAt             *time.Time `json:"depart_at"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

type CreateTripStopInput struct {
	RouteStopID          string     `json:"route_stop_id"`
	LegDistanceKM        *float64   `json:"leg_distance_km"`
	CumulativeDistanceKM *float64   `json:"cumulative_distance_km"`
	ArriveAt             *time.Time `json:"arrive_at"`
	DepartAt             *time.Time `json:"depart_at"`
}

type CreateTripInput struct {
	RouteID     string     `json:"route_id"`
	BusID       string     `json:"bus_id"`
	DriverID    *string    `json:"driver_id"`
	FareID      *string    `json:"fare_id"`
	RequestID   *string    `json:"request_id"`
	DepartureAt time.Time  `json:"departure_at"`
	ArrivalAt   *time.Time `json:"arrival_at"`
	Status      *string    `json:"status"`
	EstimatedKM *float64   `json:"estimated_km"`
	PairTripID  *string    `json:"pair_trip_id"`
	Notes       *string    `json:"notes"`
}

type UpdateTripInput struct {
	RouteID           *string    `json:"route_id"`
	BusID             *string    `json:"bus_id"`
	DriverID          *string    `json:"driver_id"`
	FareID            *string    `json:"fare_id"`
	RequestID         *string    `json:"request_id"`
	DepartureAt       *time.Time `json:"departure_at"`
	ArrivalAt         *time.Time `json:"arrival_at"`
	Status            *string    `json:"status"`
	OperationalStatus *string    `json:"operational_status"`
	EstimatedKM       *float64   `json:"estimated_km"`
	PairTripID        *string    `json:"pair_trip_id"`
	Notes             *string    `json:"notes"`
}

type ListFilter struct {
	Limit  int
	Offset int
	Search string
	Status string
}
