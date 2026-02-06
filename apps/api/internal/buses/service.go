package buses

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

func (s *Service) List(ctx context.Context, filter ListFilter) ([]Bus, error) {
  return s.repo.List(ctx, filter)
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (Bus, error) {
  return s.repo.Get(ctx, id)
}

func (s *Service) Create(ctx context.Context, input CreateBusInput) (Bus, error) {
  return s.repo.Create(ctx, input)
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, input UpdateBusInput) (Bus, error) {
  return s.repo.Update(ctx, id, input)
}
