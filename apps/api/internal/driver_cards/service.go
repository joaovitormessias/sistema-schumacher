package driver_cards

import (
  "context"
  "errors"

  "github.com/google/uuid"
)

var ErrCardBlocked = errors.New("card blocked")
var ErrCardInactive = errors.New("card inactive")
var ErrCardStatus = errors.New("invalid card status")

// Service handles driver card operations
//
type Service struct {
  repo *Repository
}

func NewService(repo *Repository) *Service {
  return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, filter ListFilter) ([]DriverCard, error) {
  return s.repo.List(ctx, filter)
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (DriverCard, error) {
  return s.repo.Get(ctx, id)
}

func (s *Service) Create(ctx context.Context, input CreateDriverCardInput) (DriverCard, error) {
  return s.repo.Create(ctx, input)
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, input UpdateDriverCardInput) (DriverCard, error) {
  return s.repo.Update(ctx, id, input)
}

func (s *Service) Block(ctx context.Context, id uuid.UUID, reason *string, blockedBy *string) (DriverCard, error) {
  current, err := s.repo.Get(ctx, id)
  if err != nil {
    return DriverCard{}, err
  }
  if current.IsBlocked {
    return DriverCard{}, ErrCardStatus
  }
  return s.repo.Block(ctx, id, reason, blockedBy)
}

func (s *Service) Unblock(ctx context.Context, id uuid.UUID) (DriverCard, error) {
  current, err := s.repo.Get(ctx, id)
  if err != nil {
    return DriverCard{}, err
  }
  if !current.IsBlocked {
    return DriverCard{}, ErrCardStatus
  }
  return s.repo.Unblock(ctx, id)
}

func (s *Service) CreateTransaction(ctx context.Context, cardID uuid.UUID, input CreateCardTransactionInput, tripExpenseID *uuid.UUID, performedBy *string) (DriverCardTransaction, error) {
  card, err := s.repo.Get(ctx, cardID)
  if err != nil {
    return DriverCardTransaction{}, err
  }
  if !card.IsActive {
    return DriverCardTransaction{}, ErrCardInactive
  }
  if card.IsBlocked {
    return DriverCardTransaction{}, ErrCardBlocked
  }
  return s.repo.CreateTransaction(ctx, cardID, input, tripExpenseID, performedBy)
}

func (s *Service) ListTransactions(ctx context.Context, filter TransactionListFilter) ([]DriverCardTransaction, error) {
  return s.repo.ListTransactions(ctx, filter)
}
