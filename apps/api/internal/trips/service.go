package trips

import (
  "context"

  "github.com/google/uuid"
)

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
  return s.repo.Create(ctx, input)
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, input UpdateTripInput) (Trip, error) {
  return s.repo.Update(ctx, id, input)
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
