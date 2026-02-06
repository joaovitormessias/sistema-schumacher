package advance_returns

import (
  "context"

  "github.com/google/uuid"
)

// Service handles advance returns
//
type Service struct {
  repo *Repository
}

func NewService(repo *Repository) *Service {
  return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, filter ListFilter) ([]AdvanceReturn, error) {
  return s.repo.List(ctx, filter)
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (AdvanceReturn, error) {
  return s.repo.Get(ctx, id)
}

func (s *Service) Create(ctx context.Context, input CreateAdvanceReturnInput, receivedBy *string) (AdvanceReturn, error) {
  return s.repo.Create(ctx, input, receivedBy)
}
