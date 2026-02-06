package bookings

import (
  "context"
  "errors"
  "math"

  "schumacher-tur/api/internal/pricing"
)

var (
  ErrInvalidAmounts  = errors.New("deposit + remainder must equal total_amount")
  ErrNegativeAmounts = errors.New("amounts cannot be negative")
  ErrMissingStops    = errors.New("board_stop_id and alight_stop_id are required")
  ErrMissingFields   = errors.New("trip_id, seat_id and passenger.name are required")
)

type Service struct {
  repo *Repository
  pricing *pricing.Service
}

func NewService(repo *Repository, pricingSvc *pricing.Service) *Service {
  return &Service{repo: repo, pricing: pricingSvc}
}

func (s *Service) List(ctx context.Context, filter ListFilter) ([]BookingListItem, error) {
  return s.repo.List(ctx, filter)
}

func (s *Service) Create(ctx context.Context, input CreateBookingInput) (BookingDetails, error) {
  if input.TripID == "" || input.SeatID == "" || input.Passenger.Name == "" {
    return BookingDetails{}, ErrMissingFields
  }
  if input.BoardStopID == "" || input.AlightStopID == "" {
    return BookingDetails{}, ErrMissingStops
  }
  if input.TotalAmount < 0 || input.DepositAmount < 0 || input.RemainderAmount < 0 {
    return BookingDetails{}, ErrNegativeAmounts
  }

  fareMode := "AUTO"
  if input.FareMode != nil && *input.FareMode != "" {
    fareMode = *input.FareMode
  }

  quote, err := s.pricing.Quote(ctx, pricing.QuoteInput{
    TripID:          input.TripID,
    BoardStopID:     input.BoardStopID,
    AlightStopID:    input.AlightStopID,
    FareMode:        fareMode,
    FareAmountFinal: input.FareAmountFinal,
  })
  if err != nil {
    return BookingDetails{}, err
  }

  total := quote.FinalAmount
  deposit := input.DepositAmount
  remainder := input.RemainderAmount
  if deposit == 0 && remainder == 0 {
    remainder = total
  }
  if total > 0 {
    sum := deposit + remainder
    if math.Abs(sum-total) > 0.01 {
      return BookingDetails{}, ErrInvalidAmounts
    }
  }

  data := CreateBookingData{
    TripID:          input.TripID,
    SeatID:          input.SeatID,
    BoardStopID:     input.BoardStopID,
    AlightStopID:    input.AlightStopID,
    BoardStopOrder:  quote.BoardStopOrder,
    AlightStopOrder: quote.AlightStopOrder,
    FareMode:        quote.FareMode,
    FareAmountCalc:  quote.CalcAmount,
    FareAmountFinal: quote.FinalAmount,
    FareSnapshot:    quote.Snapshot,
    Passenger:       input.Passenger,
    Source:          input.Source,
    TotalAmount:     total,
    DepositAmount:   deposit,
    RemainderAmount: remainder,
  }

  return s.repo.Create(ctx, data)
}

func (s *Service) Get(ctx context.Context, id string) (BookingDetails, error) {
  return s.repo.Get(ctx, id)
}

func (s *Service) UpdateStatus(ctx context.Context, id string, status string) (BookingDetails, error) {
  return s.repo.UpdateStatus(ctx, id, status)
}
