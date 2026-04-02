package routes

import "time"

// Route represents a travel route.
type Route struct {
	ID                    string    `json:"id"`
	Name                  string    `json:"name"`
	OriginCity            string    `json:"origin_city"`
	DestinationCity       string    `json:"destination_city"`
	IsActive              bool      `json:"is_active"`
	StopCount             int       `json:"stop_count"`
	ConfigurationStatus   string    `json:"configuration_status"`
	MissingRequirements   []string  `json:"missing_requirements"`
	HasLinkedTrips        bool      `json:"has_linked_trips"`
	DuplicatedFromRouteID *string   `json:"duplicated_from_route_id,omitempty"`
	CreatedAt             time.Time `json:"created_at"`
}

type CreateRouteInput struct {
	Name                 string   `json:"name"`
	OriginCity           string   `json:"origin_city"`
	OriginLatitude       *float64 `json:"origin_latitude"`
	OriginLongitude      *float64 `json:"origin_longitude"`
	DestinationCity      string   `json:"destination_city"`
	DestinationLatitude  *float64 `json:"destination_latitude"`
	DestinationLongitude *float64 `json:"destination_longitude"`
	IsActive             *bool    `json:"is_active"`
}

type UpdateRouteInput struct {
	Name            *string `json:"name"`
	OriginCity      *string `json:"origin_city"`
	DestinationCity *string `json:"destination_city"`
	IsActive        *bool   `json:"is_active"`
}

type ListFilter struct {
	Limit  int
	Offset int
	Search string
	Status string
}

type RouteStop struct {
	ID               string    `json:"id"`
	RouteID          string    `json:"route_id"`
	City             string    `json:"city"`
	Latitude         *float64  `json:"latitude,omitempty"`
	Longitude        *float64  `json:"longitude,omitempty"`
	StopOrder        int       `json:"stop_order"`
	ETAOffsetMinutes *int      `json:"eta_offset_minutes"`
	Notes            *string   `json:"notes"`
	CreatedAt        time.Time `json:"created_at"`
}

type CreateRouteStopInput struct {
	City             string   `json:"city"`
	Latitude         *float64 `json:"latitude"`
	Longitude        *float64 `json:"longitude"`
	StopOrder        int      `json:"stop_order"`
	ETAOffsetMinutes *int     `json:"eta_offset_minutes"`
	Notes            *string  `json:"notes"`
}

type UpdateRouteStopInput struct {
	City             *string  `json:"city"`
	Latitude         *float64 `json:"latitude"`
	Longitude        *float64 `json:"longitude"`
	StopOrder        *int     `json:"stop_order"`
	ETAOffsetMinutes *int     `json:"eta_offset_minutes"`
	Notes            *string  `json:"notes"`
}

type RouteSegmentPriceStop struct {
	StopID      string `json:"stop_id"`
	DisplayName string `json:"display_name"`
	StopOrder   int    `json:"stop_order"`
}

type RouteSegmentPriceItem struct {
	OriginStopID           string     `json:"origin_stop_id"`
	OriginDisplayName      string     `json:"origin_display_name"`
	OriginStopOrder        int        `json:"origin_stop_order"`
	DestinationStopID      string     `json:"destination_stop_id"`
	DestinationDisplayName string     `json:"destination_display_name"`
	DestinationStopOrder   int        `json:"destination_stop_order"`
	Price                  *float64   `json:"price"`
	Status                 string     `json:"status"`
	Configured             bool       `json:"configured"`
	CreatedAt              *time.Time `json:"created_at"`
	UpdatedAt              *time.Time `json:"updated_at"`
}

type RouteSegmentPriceMatrix struct {
	RouteID string                  `json:"route_id"`
	Stops   []RouteSegmentPriceStop `json:"stops"`
	Items   []RouteSegmentPriceItem `json:"items"`
}

type UpsertRouteSegmentPriceItem struct {
	OriginStopID      string   `json:"origin_stop_id"`
	DestinationStopID string   `json:"destination_stop_id"`
	Price             *float64 `json:"price"`
	Status            *string  `json:"status"`
}

type UpsertRouteSegmentPricesInput struct {
	Items []UpsertRouteSegmentPriceItem `json:"items"`
}

type CityCandidate struct {
	PlaceID     string  `json:"place_id"`
	City        string  `json:"city"`
	StateCode   string  `json:"state_code,omitempty"`
	DisplayName string  `json:"display_name"`
	Addresstype string  `json:"addresstype"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
}
