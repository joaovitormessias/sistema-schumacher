package pricing

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

func (s *Service) List(ctx context.Context, filter ListFilter) ([]PricingRule, error) {
  return s.repo.List(ctx, filter)
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (PricingRule, error) {
  return s.repo.Get(ctx, id)
}

func (s *Service) Create(ctx context.Context, input CreatePricingRuleInput) (PricingRule, error) {
  return s.repo.Create(ctx, input)
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, input UpdatePricingRuleInput) (PricingRule, error) {
  return s.repo.Update(ctx, id, input)
}

func (s *Service) Quote(ctx context.Context, input QuoteInput) (QuoteResult, error) {
  return s.repo.Quote(ctx, input)
}
