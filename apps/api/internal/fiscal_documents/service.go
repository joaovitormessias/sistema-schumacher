package fiscal_documents

import (
  "context"

  "github.com/google/uuid"
)

// Service handles fiscal documents
//
type Service struct {
  repo *Repository
}

func NewService(repo *Repository) *Service {
  return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, filter ListFilter) ([]FiscalDocument, error) {
  return s.repo.List(ctx, filter)
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (FiscalDocument, error) {
  return s.repo.Get(ctx, id)
}

func (s *Service) Create(ctx context.Context, input CreateFiscalDocumentInput, createdBy *string) (FiscalDocument, error) {
  return s.repo.Create(ctx, input, createdBy)
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, input UpdateFiscalDocumentInput) (FiscalDocument, error) {
  return s.repo.Update(ctx, id, input)
}
