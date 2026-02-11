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
	Name            string `json:"name"`
	OriginCity      string `json:"origin_city"`
	DestinationCity string `json:"destination_city"`
	IsActive        *bool  `json:"is_active"`
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
	StopOrder        int       `json:"stop_order"`
	ETAOffsetMinutes *int      `json:"eta_offset_minutes"`
	Notes            *string   `json:"notes"`
	CreatedAt        time.Time `json:"created_at"`
}

type CreateRouteStopInput struct {
	City             string  `json:"city"`
	StopOrder        int     `json:"stop_order"`
	ETAOffsetMinutes *int    `json:"eta_offset_minutes"`
	Notes            *string `json:"notes"`
}

type UpdateRouteStopInput struct {
	City             *string `json:"city"`
	StopOrder        *int    `json:"stop_order"`
	ETAOffsetMinutes *int    `json:"eta_offset_minutes"`
	Notes            *string `json:"notes"`
}
