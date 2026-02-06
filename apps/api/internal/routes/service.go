package routes

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

func (s *Service) List(ctx context.Context, filter ListFilter) ([]Route, error) {
  return s.repo.List(ctx, filter)
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (Route, error) {
  return s.repo.Get(ctx, id)
}

func (s *Service) Create(ctx context.Context, input CreateRouteInput) (Route, error) {
  return s.repo.Create(ctx, input)
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, input UpdateRouteInput) (Route, error) {
  return s.repo.Update(ctx, id, input)
}

func (s *Service) ListStops(ctx context.Context, routeID uuid.UUID) ([]RouteStop, error) {
  return s.repo.ListStops(ctx, routeID)
}

func (s *Service) CreateStop(ctx context.Context, routeID uuid.UUID, input CreateRouteStopInput) (RouteStop, error) {
  return s.repo.CreateStop(ctx, routeID, input)
}

func (s *Service) UpdateStop(ctx context.Context, routeID uuid.UUID, stopID uuid.UUID, input UpdateRouteStopInput) (RouteStop, error) {
  return s.repo.UpdateStop(ctx, routeID, stopID, input)
}
