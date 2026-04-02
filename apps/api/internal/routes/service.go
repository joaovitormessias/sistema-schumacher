package routes

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

type RoutePublishBlockedError struct {
	MissingRequirements []string `json:"requirements_missing"`
}

func (e RoutePublishBlockedError) Error() string {
	return "route does not meet publish requirements"
}

var ErrRouteStopLocked = errors.New("route stop update is locked for routes with linked trips")
var ErrInvalidSegmentPair = errors.New("invalid segment pair")
var ErrInvalidSegmentPrice = errors.New("invalid segment price")
var ErrInvalidSegmentStatus = errors.New("invalid segment status")

type Service struct {
	repo     *Repository
	geocoder Geocoder
}

func NewService(repo *Repository) *Service {
	return &Service{
		repo:     repo,
		geocoder: NewNominatimGeocoderFromEnv(),
	}
}

func (s *Service) List(ctx context.Context, filter ListFilter) ([]Route, error) {
	return s.repo.List(ctx, filter)
}

func (s *Service) SearchCityCandidates(ctx context.Context, query string, limit int) ([]CityCandidate, error) {
	trimmedQuery := strings.TrimSpace(query)
	if trimmedQuery == "" {
		return []CityCandidate{}, nil
	}
	if s.geocoder == nil {
		return nil, errors.New("automatic geocoding is not configured")
	}
	return s.geocoder.SearchCityCandidates(ctx, trimmedQuery, limit)
}

func (s *Service) Get(ctx context.Context, id string) (Route, error) {
	return s.repo.Get(ctx, id)
}

func (s *Service) Create(ctx context.Context, input CreateRouteInput) (Route, error) {
	var err error
	input.OriginLatitude, input.OriginLongitude, err = s.resolveCoordinatesIfMissing(ctx, input.OriginCity, input.OriginLatitude, input.OriginLongitude)
	if err != nil {
		return Route{}, err
	}
	input.DestinationLatitude, input.DestinationLongitude, err = s.resolveCoordinatesIfMissing(ctx, input.DestinationCity, input.DestinationLatitude, input.DestinationLongitude)
	if err != nil {
		return Route{}, err
	}
	if err := validateRouteEndpointCoordinates("origin", input.OriginLatitude, input.OriginLongitude); err != nil {
		return Route{}, err
	}
	if err := validateRouteEndpointCoordinates("destination", input.DestinationLatitude, input.DestinationLongitude); err != nil {
		return Route{}, err
	}
	if input.IsActive != nil && *input.IsActive {
		// New routes should start as draft/inactive and be published later.
		forceInactive := false
		input.IsActive = &forceInactive
	}
	return s.repo.Create(ctx, input)
}

func (s *Service) Update(ctx context.Context, id string, input UpdateRouteInput) (Route, error) {
	if input.IsActive != nil && *input.IsActive {
		if err := s.ensureRouteCanBePublished(ctx, id); err != nil {
			return Route{}, err
		}
	}
	return s.repo.Update(ctx, id, input)
}

func (s *Service) Publish(ctx context.Context, id string) (Route, error) {
	if err := s.ensureRouteCanBePublished(ctx, id); err != nil {
		return Route{}, err
	}
	return s.repo.Publish(ctx, id)
}

func (s *Service) Duplicate(ctx context.Context, id string) (Route, error) {
	return s.repo.Duplicate(ctx, id)
}

func (s *Service) ListStops(ctx context.Context, routeID string) ([]RouteStop, error) {
	return s.repo.ListStops(ctx, routeID)
}

func (s *Service) CreateStop(ctx context.Context, routeID string, input CreateRouteStopInput) (RouteStop, error) {
	var err error
	input.Latitude, input.Longitude, err = s.resolveCoordinatesIfMissing(ctx, input.City, input.Latitude, input.Longitude)
	if err != nil {
		return RouteStop{}, err
	}
	if input.ETAOffsetMinutes != nil && *input.ETAOffsetMinutes < 0 {
		return RouteStop{}, errors.New("eta_offset_minutes must be >= 0")
	}
	if err := validateStopCoordinates(input.Latitude, input.Longitude); err != nil {
		return RouteStop{}, err
	}
	return s.repo.CreateStop(ctx, routeID, input)
}

func (s *Service) UpdateStop(ctx context.Context, routeID string, stopID string, input UpdateRouteStopInput) (RouteStop, error) {
	if input.City != nil && strings.TrimSpace(*input.City) != "" {
		resolvedLatitude, resolvedLongitude, err := s.resolveCoordinatesIfMissing(ctx, *input.City, input.Latitude, input.Longitude)
		if err != nil {
			return RouteStop{}, err
		}
		input.Latitude = resolvedLatitude
		input.Longitude = resolvedLongitude
	}
	if input.ETAOffsetMinutes != nil && *input.ETAOffsetMinutes < 0 {
		return RouteStop{}, errors.New("eta_offset_minutes must be >= 0")
	}
	if err := validateStopCoordinates(input.Latitude, input.Longitude); err != nil {
		return RouteStop{}, err
	}
	return s.repo.UpdateStop(ctx, routeID, stopID, input)
}

func (s *Service) DeleteStop(ctx context.Context, routeID string, stopID string) error {
	return s.repo.DeleteStop(ctx, routeID, stopID)
}

func (s *Service) ListSegmentPrices(ctx context.Context, routeID string) (RouteSegmentPriceMatrix, error) {
	return s.repo.ListSegmentPrices(ctx, routeID)
}

func (s *Service) UpsertSegmentPrices(ctx context.Context, routeID string, input UpsertRouteSegmentPricesInput) (RouteSegmentPriceMatrix, error) {
	for _, item := range input.Items {
		if item.OriginStopID == "" || item.DestinationStopID == "" || item.OriginStopID == item.DestinationStopID {
			return RouteSegmentPriceMatrix{}, ErrInvalidSegmentPair
		}
		if item.Price != nil && *item.Price < 0 {
			return RouteSegmentPriceMatrix{}, ErrInvalidSegmentPrice
		}
		if item.Status != nil {
			status := normalizeSegmentStatus(*item.Status)
			if status != "ACTIVE" && status != "INACTIVE" {
				return RouteSegmentPriceMatrix{}, ErrInvalidSegmentStatus
			}
		}
	}
	return s.repo.UpsertSegmentPrices(ctx, routeID, input)
}

func (s *Service) ensureRouteCanBePublished(ctx context.Context, routeID string) error {
	route, err := s.repo.Get(ctx, routeID)
	if err != nil {
		return err
	}
	if len(route.MissingRequirements) > 0 {
		return RoutePublishBlockedError{MissingRequirements: route.MissingRequirements}
	}
	return nil
}

func normalizeSegmentStatus(status string) string {
	return strings.ToUpper(strings.TrimSpace(status))
}

func validateStopCoordinates(latitude *float64, longitude *float64) error {
	if (latitude == nil) != (longitude == nil) {
		return errors.New("latitude and longitude must be informed together")
	}
	if latitude == nil || longitude == nil {
		return nil
	}
	if *latitude < -90 || *latitude > 90 {
		return errors.New("latitude must be between -90 and 90")
	}
	if *longitude < -180 || *longitude > 180 {
		return errors.New("longitude must be between -180 and 180")
	}
	return nil
}

func validateRouteEndpointCoordinates(label string, latitude *float64, longitude *float64) error {
	if (latitude == nil) != (longitude == nil) {
		return errors.New(label + "_latitude and " + label + "_longitude must be informed together")
	}
	if latitude == nil || longitude == nil {
		return nil
	}
	if *latitude < -90 || *latitude > 90 {
		return errors.New(label + "_latitude must be between -90 and 90")
	}
	if *longitude < -180 || *longitude > 180 {
		return errors.New(label + "_longitude must be between -180 and 180")
	}
	return nil
}

func (s *Service) resolveCoordinatesIfMissing(
	ctx context.Context,
	city string,
	latitude *float64,
	longitude *float64,
) (*float64, *float64, error) {
	if latitude != nil || longitude != nil {
		return latitude, longitude, nil
	}
	trimmedCity := strings.TrimSpace(city)
	if trimmedCity == "" {
		return latitude, longitude, nil
	}
	if s.geocoder == nil {
		return nil, nil, errors.New("automatic geocoding is not configured")
	}
	resolvedLatitude, resolvedLongitude, err := s.geocoder.GeocodeCity(ctx, trimmedCity)
	if err != nil {
		return nil, nil, fmt.Errorf("could not resolve coordinates for city %q: %w", trimmedCity, err)
	}
	return &resolvedLatitude, &resolvedLongitude, nil
}
