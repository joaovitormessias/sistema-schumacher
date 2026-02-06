package trip_advances

import (
  "context"
  "errors"

  "github.com/google/uuid"
)

var ErrAdvanceLocked = errors.New("advance locked")
var ErrAdvanceStatus = errors.New("invalid advance status")

type Service struct {
  repo *Repository
}

func NewService(repo *Repository) *Service {
  return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, filter ListFilter) ([]TripAdvance, error) {
  return s.repo.List(ctx, filter)
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (TripAdvance, error) {
  return s.repo.Get(ctx, id)
}

func (s *Service) Create(ctx context.Context, input CreateTripAdvanceInput, createdBy *string) (TripAdvance, error) {
  return s.repo.Create(ctx, input, createdBy)
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, input UpdateTripAdvanceInput) (TripAdvance, error) {
  current, err := s.repo.Get(ctx, id)
  if err != nil {
    return TripAdvance{}, err
  }
  if current.Status != "PENDING" {
    return TripAdvance{}, ErrAdvanceLocked
  }
  return s.repo.Update(ctx, id, input)
}

func (s *Service) Deliver(ctx context.Context, id uuid.UUID, deliveredBy *string) (TripAdvance, error) {
  current, err := s.repo.Get(ctx, id)
  if err != nil {
    return TripAdvance{}, err
  }
  if current.Status != "PENDING" {
    return TripAdvance{}, ErrAdvanceStatus
  }
  return s.repo.MarkDelivered(ctx, id, deliveredBy)
}
