package trip_expenses

import (
  "context"
  "errors"

  "github.com/google/uuid"
)

var ErrExpenseLocked = errors.New("expense locked")
var ErrExpenseApproved = errors.New("expense already approved")
var ErrCardRequired = errors.New("card required")
var ErrInvalidPaymentMethod = errors.New("invalid payment method")

// Service handles trip expenses operations
//
type Service struct {
  repo *Repository
}

func NewService(repo *Repository) *Service {
  return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, filter ListFilter) ([]TripExpense, error) {
  return s.repo.List(ctx, filter)
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (TripExpense, error) {
  return s.repo.Get(ctx, id)
}

func (s *Service) Create(ctx context.Context, input CreateTripExpenseInput, createdBy *string, performedBy *string) (TripExpense, error) {
  if input.PaymentMethod == "" {
    input.PaymentMethod = "ADVANCE"
  }
  if !isValidPaymentMethod(input.PaymentMethod) {
    return TripExpense{}, ErrInvalidPaymentMethod
  }
  if input.PaymentMethod == "CARD" && input.DriverCardID == nil {
    return TripExpense{}, ErrCardRequired
  }
  if input.PaymentMethod != "CARD" {
    input.DriverCardID = nil
  }
  return s.repo.Create(ctx, input, createdBy, performedBy)
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, input UpdateTripExpenseInput) (TripExpense, error) {
  current, err := s.repo.Get(ctx, id)
  if err != nil {
    return TripExpense{}, err
  }
  if current.IsApproved {
    return TripExpense{}, ErrExpenseLocked
  }
  return s.repo.Update(ctx, id, input)
}

func (s *Service) Approve(ctx context.Context, id uuid.UUID, approvedBy *string) (TripExpense, error) {
  current, err := s.repo.Get(ctx, id)
  if err != nil {
    return TripExpense{}, err
  }
  if current.IsApproved {
    return TripExpense{}, ErrExpenseApproved
  }
  return s.repo.Approve(ctx, id, approvedBy)
}

func isValidPaymentMethod(method string) bool {
  switch method {
  case "ADVANCE", "CARD", "PERSONAL", "COMPANY":
    return true
  default:
    return false
  }
}
