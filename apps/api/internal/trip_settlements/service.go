package trip_settlements

import (
  "context"
  "errors"
  "math"

  "github.com/google/uuid"

  "schumacher-tur/api/internal/trip_advances"
  "schumacher-tur/api/internal/trip_expenses"
)

var ErrSettlementExists = errors.New("settlement already exists")
var ErrInvalidStatus = errors.New("invalid status transition")

// Service handles trip settlements
//
type Service struct {
  repo     *Repository
  advances *trip_advances.Repository
  expenses *trip_expenses.Repository
}

func NewService(repo *Repository, advances *trip_advances.Repository, expenses *trip_expenses.Repository) *Service {
  return &Service{repo: repo, advances: advances, expenses: expenses}
}

func (s *Service) List(ctx context.Context, filter ListFilter) ([]TripSettlement, error) {
  return s.repo.List(ctx, filter)
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (TripSettlement, error) {
  return s.repo.Get(ctx, id)
}

func (s *Service) Create(ctx context.Context, input CreateTripSettlementInput, createdBy *string) (TripSettlement, error) {
  tripID, err := uuid.Parse(input.TripID)
  if err != nil {
    return TripSettlement{}, err
  }

  driverID, err := s.repo.GetTripDriverID(ctx, tripID)
  if err != nil {
    return TripSettlement{}, err
  }

  advanceTotal, err := s.advances.SumDeliveredByTrip(ctx, tripID)
  if err != nil {
    return TripSettlement{}, err
  }
  expensesTotal, err := s.expenses.SumApprovedByTrip(ctx, tripID)
  if err != nil {
    return TripSettlement{}, err
  }

  balance := advanceTotal - expensesTotal
  amountToReturn := math.Max(balance, 0)
  amountToReimburse := math.Max(-balance, 0)

  item, err := s.repo.Create(ctx, input.TripID, driverID, advanceTotal, expensesTotal, balance, amountToReturn, amountToReimburse, input.Notes, createdBy)
  if err != nil {
    if IsUniqueViolation(err) {
      return TripSettlement{}, ErrSettlementExists
    }
    return TripSettlement{}, err
  }
  return item, nil
}

func (s *Service) Review(ctx context.Context, id uuid.UUID, reviewedBy *string) (TripSettlement, error) {
  current, err := s.repo.Get(ctx, id)
  if err != nil {
    return TripSettlement{}, err
  }
  if current.Status != "DRAFT" {
    return TripSettlement{}, ErrInvalidStatus
  }
  return s.repo.SetUnderReview(ctx, id, reviewedBy)
}

func (s *Service) Approve(ctx context.Context, id uuid.UUID, approvedBy *string) (TripSettlement, error) {
  current, err := s.repo.Get(ctx, id)
  if err != nil {
    return TripSettlement{}, err
  }
  if current.Status != "UNDER_REVIEW" {
    return TripSettlement{}, ErrInvalidStatus
  }
  return s.repo.SetApproved(ctx, id, approvedBy)
}

func (s *Service) Reject(ctx context.Context, id uuid.UUID, reviewedBy *string) (TripSettlement, error) {
  current, err := s.repo.Get(ctx, id)
  if err != nil {
    return TripSettlement{}, err
  }
  if current.Status != "UNDER_REVIEW" {
    return TripSettlement{}, ErrInvalidStatus
  }
  return s.repo.SetRejected(ctx, id, reviewedBy)
}

func (s *Service) Complete(ctx context.Context, id uuid.UUID) (TripSettlement, error) {
  current, err := s.repo.Get(ctx, id)
  if err != nil {
    return TripSettlement{}, err
  }
  if current.Status != "APPROVED" {
    return TripSettlement{}, ErrInvalidStatus
  }
  return s.repo.Complete(ctx, id)
}
