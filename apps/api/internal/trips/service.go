package trips

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

var ErrOperationalStatusManagedByWorkflow = errors.New("operational status is managed by workflow")
var ErrTripStatusManagedByWorkflow = errors.New("trip status progression is managed by workflow")

type RouteNotReadyError struct {
	MissingRequirements []string `json:"requirements_missing"`
}

func (e RouteNotReadyError) Error() string {
	return "route is not ready for trip creation"
}

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, filter ListFilter) ([]Trip, error) {
	return s.repo.List(ctx, filter)
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (Trip, error) {
	return s.repo.Get(ctx, id)
}

func (s *Service) Create(ctx context.Context, input CreateTripInput) (Trip, error) {
	missingRequirements, err := s.repo.ValidateRouteReadiness(ctx, input.RouteID)
	if err != nil {
		return Trip{}, err
	}
	if len(missingRequirements) > 0 {
		return Trip{}, RouteNotReadyError{MissingRequirements: missingRequirements}
	}
	return s.repo.Create(ctx, input)
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, input UpdateTripInput) (Trip, error) {
	if input.OperationalStatus != nil {
		return Trip{}, ErrOperationalStatusManagedByWorkflow
	}
	if input.Status != nil && *input.Status != "SCHEDULED" && *input.Status != "CANCELLED" {
		return Trip{}, ErrTripStatusManagedByWorkflow
	}
	if input.RouteID != nil {
		missingRequirements, err := s.repo.ValidateRouteReadiness(ctx, *input.RouteID)
		if err != nil {
			return Trip{}, err
		}
		if len(missingRequirements) > 0 {
			return Trip{}, RouteNotReadyError{MissingRequirements: missingRequirements}
		}
	}
	item, err := s.repo.Update(ctx, id, input)
	if err != nil {
		return Trip{}, err
	}
	return item, nil
}

func (s *Service) ListSeats(ctx context.Context, tripID uuid.UUID, boardStopID *uuid.UUID, alightStopID *uuid.UUID) ([]TripSeat, error) {
	return s.repo.ListSeats(ctx, tripID, boardStopID, alightStopID)
}

func (s *Service) ListStops(ctx context.Context, tripID uuid.UUID) ([]TripStop, error) {
	return s.repo.ListStops(ctx, tripID)
}

func (s *Service) CreateStop(ctx context.Context, tripID uuid.UUID, input CreateTripStopInput) (TripStop, error) {
	return s.repo.CreateStop(ctx, tripID, input)
}
