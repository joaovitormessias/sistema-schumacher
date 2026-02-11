package routes

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

type RoutePublishBlockedError struct {
	MissingRequirements []string `json:"requirements_missing"`
}

func (e RoutePublishBlockedError) Error() string {
	return "route does not meet publish requirements"
}

var ErrRouteStopLocked = errors.New("route stop update is locked for routes with linked trips")

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, filter ListFilter) ([]Route, error) {
	return s.repo.List(ctx, filter)
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (Route, error) {
	return s.repo.Get(ctx, id)
}

func (s *Service) Create(ctx context.Context, input CreateRouteInput) (Route, error) {
	if input.IsActive != nil && *input.IsActive {
		// New routes should start as draft/inactive and be published later.
		forceInactive := false
		input.IsActive = &forceInactive
	}
	return s.repo.Create(ctx, input)
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, input UpdateRouteInput) (Route, error) {
	if input.IsActive != nil && *input.IsActive {
		if err := s.ensureRouteCanBePublished(ctx, id); err != nil {
			return Route{}, err
		}
	}
	return s.repo.Update(ctx, id, input)
}

func (s *Service) Publish(ctx context.Context, id uuid.UUID) (Route, error) {
	if err := s.ensureRouteCanBePublished(ctx, id); err != nil {
		return Route{}, err
	}
	return s.repo.Publish(ctx, id)
}

func (s *Service) Duplicate(ctx context.Context, id uuid.UUID) (Route, error) {
	return s.repo.Duplicate(ctx, id)
}

func (s *Service) ListStops(ctx context.Context, routeID uuid.UUID) ([]RouteStop, error) {
	return s.repo.ListStops(ctx, routeID)
}

func (s *Service) CreateStop(ctx context.Context, routeID uuid.UUID, input CreateRouteStopInput) (RouteStop, error) {
	if input.ETAOffsetMinutes != nil && *input.ETAOffsetMinutes < 0 {
		return RouteStop{}, errors.New("eta_offset_minutes must be >= 0")
	}
	return s.repo.CreateStop(ctx, routeID, input)
}

func (s *Service) UpdateStop(ctx context.Context, routeID uuid.UUID, stopID uuid.UUID, input UpdateRouteStopInput) (RouteStop, error) {
	if input.ETAOffsetMinutes != nil && *input.ETAOffsetMinutes < 0 {
		return RouteStop{}, errors.New("eta_offset_minutes must be >= 0")
	}
	if input.StopOrder != nil {
		hasLinkedTrips, err := s.repo.HasLinkedTrips(ctx, routeID)
		if err != nil {
			return RouteStop{}, err
		}
		if hasLinkedTrips {
			return RouteStop{}, ErrRouteStopLocked
		}
	}
	return s.repo.UpdateStop(ctx, routeID, stopID, input)
}

func (s *Service) DeleteStop(ctx context.Context, routeID uuid.UUID, stopID uuid.UUID) error {
	return s.repo.DeleteStop(ctx, routeID, stopID)
}

func (s *Service) ensureRouteCanBePublished(ctx context.Context, routeID uuid.UUID) error {
	route, err := s.repo.Get(ctx, routeID)
	if err != nil {
		return err
	}
	if len(route.MissingRequirements) > 0 {
		return RoutePublishBlockedError{MissingRequirements: route.MissingRequirements}
	}
	return nil
}
