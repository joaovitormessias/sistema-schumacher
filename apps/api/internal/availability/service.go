package availability

import "context"

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Search(ctx context.Context, filter SearchFilter) ([]SearchResult, error) {
	return s.repo.Search(ctx, filter)
}
