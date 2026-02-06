package trip_validations

import (
  "context"
  "errors"

  "github.com/google/uuid"
)

var ErrValidationExists = errors.New("validation already exists")

// Service handles trip validations
//
type Service struct {
  repo *Repository
}

func NewService(repo *Repository) *Service {
  return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, filter ListFilter) ([]TripValidation, error) {
  return s.repo.List(ctx, filter)
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (TripValidation, error) {
  return s.repo.Get(ctx, id)
}

func (s *Service) Create(ctx context.Context, input CreateTripValidationInput, validatedBy *string) (TripValidation, error) {
  item, err := s.repo.Create(ctx, input, validatedBy)
  if err != nil {
    if IsUniqueViolation(err) {
      return TripValidation{}, ErrValidationExists
    }
    return TripValidation{}, err
  }
  return item, nil
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, input UpdateTripValidationInput, validatedBy *string) (TripValidation, error) {
  return s.repo.Update(ctx, id, input, validatedBy)
}
