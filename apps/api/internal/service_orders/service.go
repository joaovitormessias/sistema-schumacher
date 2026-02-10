package service_orders

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

func (s *Service) List(ctx context.Context, filter ListFilter) ([]ServiceOrder, error) {
	return s.repo.List(ctx, filter)
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (ServiceOrder, error) {
	return s.repo.Get(ctx, id)
}

func (s *Service) Create(ctx context.Context, input CreateServiceOrderInput) (ServiceOrder, error) {
	return s.repo.Create(ctx, input)
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, input UpdateServiceOrderInput) (ServiceOrder, error) {
	return s.repo.Update(ctx, id, input)
}

func (s *Service) StartProgress(ctx context.Context, id uuid.UUID) error {
	return s.repo.SetStatus(ctx, id, StatusInProgress)
}

func (s *Service) Close(ctx context.Context, id uuid.UUID, input CloseServiceOrderInput) (ServiceOrder, error) {
	return s.repo.Close(ctx, id, input)
}

func (s *Service) Cancel(ctx context.Context, id uuid.UUID) error {
	return s.repo.SetStatus(ctx, id, StatusCancelled)
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
