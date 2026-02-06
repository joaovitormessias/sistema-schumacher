package reports

import "context"

type Service struct {
  repo *Repository
}

func NewService(repo *Repository) *Service {
  return &Service{repo: repo}
}

func (s *Service) ListPassengers(ctx context.Context, tripID string) ([]PassengerReportRow, error) {
  return s.repo.ListPassengers(ctx, tripID)
}
