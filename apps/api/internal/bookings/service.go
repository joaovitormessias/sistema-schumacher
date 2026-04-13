package bookings

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"strings"

	"schumacher-tur/api/internal/payments"
	"schumacher-tur/api/internal/pricing"
)

var (
	ErrInvalidAmounts              = errors.New("deposit + remainder must equal total_amount")
	ErrNegativeAmounts             = errors.New("amounts cannot be negative")
	ErrMissingStops                = errors.New("board_stop_id and alight_stop_id are required")
	ErrMissingFields               = errors.New("trip_id and at least one passenger.name are required")
	ErrPassengerNameRequired       = errors.New("every passenger must include name")
	ErrSeatRequiresSinglePassenger = errors.New("seat_id can be used only with a single passenger")
	ErrInitialPaymentRequired      = errors.New("initial_payment is required")
	ErrInvalidInitialPayment       = errors.New("initial payment method or amount is invalid")
	ErrInitialPaymentBelowMinimum  = errors.New("initial_payment.amount must be at least 30% of total_amount")
)

type Service struct {
	repo     *Repository
	pricing  *pricing.Service
	payments *payments.Service
}

func NewService(repo *Repository, pricingSvc *pricing.Service, paymentsSvc *payments.Service) *Service {
	return &Service{repo: repo, pricing: pricingSvc, payments: paymentsSvc}
}

func (s *Service) List(ctx context.Context, filter ListFilter) ([]BookingListItem, error) {
	return s.repo.List(ctx, filter)
}

func (s *Service) Create(ctx context.Context, input CreateBookingInput) (BookingDetails, error) {
	passengers := normalizePassengers(input.Passenger, input.Passengers)
	if input.TripID == "" || len(passengers) == 0 {
		return BookingDetails{}, ErrMissingFields
	}
	for _, passenger := range passengers {
		if strings.TrimSpace(passenger.Name) == "" {
			return BookingDetails{}, ErrPassengerNameRequired
		}
	}
	if strings.TrimSpace(input.SeatID) != "" && len(passengers) > 1 {
		return BookingDetails{}, ErrSeatRequiresSinglePassenger
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

	total := roundTo2(quote.FinalAmount * float64(len(passengers)))
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
		Passengers:      passengers,
		IdempotencyKey:  strings.TrimSpace(input.IdempotencyKey),
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

func (s *Service) Checkout(ctx context.Context, input CheckoutBookingInput) (CheckoutResponse, error) {
	if strings.TrimSpace(input.InitialPayment.Method) == "" || input.InitialPayment.Amount <= 0 {
		return CheckoutResponse{}, ErrInvalidInitialPayment
	}
	if !isCheckoutPaymentMethod(input.InitialPayment.Method) {
		return CheckoutResponse{}, ErrInvalidInitialPayment
	}

	minRequired := roundTo2(math.Max(input.TotalAmount*0.30, 0.01))
	if roundTo2(input.InitialPayment.Amount) < minRequired {
		return CheckoutResponse{}, ErrInitialPaymentBelowMinimum
	}

	deposit := roundTo2(input.InitialPayment.Amount)
	remainder := roundTo2(math.Max(input.TotalAmount-deposit, 0))
	booking, err := s.Create(ctx, CreateBookingInput{
		TripID:          input.TripID,
		SeatID:          input.SeatID,
		BoardStopID:     input.BoardStopID,
		AlightStopID:    input.AlightStopID,
		FareMode:        input.FareMode,
		FareAmountFinal: input.FareAmountFinal,
		Passenger:       input.Passenger,
		Passengers:      input.Passengers,
		IdempotencyKey:  input.IdempotencyKey,
		Source:          input.Source,
		TotalAmount:     input.TotalAmount,
		DepositAmount:   deposit,
		RemainderAmount: remainder,
	})
	if err != nil {
		return CheckoutResponse{}, err
	}
	method := strings.ToUpper(strings.TrimSpace(input.InitialPayment.Method))
	payment := CheckoutPayment{
		ID:        "",
		BookingID: booking.Booking.ID,
		Amount:    deposit,
		Method:    method,
		Status:    "PENDING",
		CreatedAt: booking.Booking.CreatedAt,
	}

	return CheckoutResponse{
		Booking:     booking,
		Payment:     payment,
		ProviderRaw: nil,
		CheckoutURL: nil,
		PixCode:     nil,
	}, nil
}

func mapCustomer(input *CheckoutCustomerInput) *payments.CustomerInput {
	if input == nil {
		return nil
	}
	return &payments.CustomerInput{
		Name:     input.Name,
		Email:    input.Email,
		Phone:    input.Phone,
		Document: input.Document,
	}
}

func mapPayment(payment payments.Payment) CheckoutPayment {
	return CheckoutPayment{
		ID:          payment.ID,
		BookingID:   payment.BookingID,
		Amount:      payment.Amount,
		Method:      payment.Method,
		Status:      payment.Status,
		Provider:    payment.Provider,
		ProviderRef: payment.ProviderRef,
		PaidAt:      payment.PaidAt,
		Metadata:    payment.Metadata,
		CreatedAt:   payment.CreatedAt,
	}
}

func parseProviderData(raw []byte) (interface{}, *string, *string) {
	if len(raw) == 0 {
		return nil, nil, nil
	}
	var parsed interface{}
	_ = json.Unmarshal(raw, &parsed)

	obj, ok := parsed.(map[string]interface{})
	if !ok {
		return parsed, nil, nil
	}

	checkout := readStringFromMap(obj, "url")
	if checkout == nil {
		if dataObj, ok := obj["data"].(map[string]interface{}); ok {
			checkout = readStringFromMap(dataObj, "url")
		}
	}

	var pixCode *string
	if dataObj, ok := obj["data"].(map[string]interface{}); ok {
		pixCode = readStringFromMap(dataObj, "pixQrCode")
	}
	if pixCode == nil {
		pixCode = readStringFromMap(obj, "pixQrCode")
	}

	return parsed, checkout, pixCode
}

func readStringFromMap(obj map[string]interface{}, key string) *string {
	raw, ok := obj[key]
	if !ok {
		return nil
	}
	value, ok := raw.(string)
	if !ok {
		return nil
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}

func isAutomaticCheckoutMethod(method string) bool {
	switch method {
	case "PIX", "CARD":
		return true
	default:
		return false
	}
}

func isCheckoutPaymentMethod(method string) bool {
	switch strings.ToUpper(strings.TrimSpace(method)) {
	case "PIX", "CARD", "CASH", "TRANSFER", "OTHER":
		return true
	default:
		return false
	}
}

func roundTo2(value float64) float64 {
	return math.Round(value*100) / 100
}

func normalizePassengers(primary PassengerInput, list []PassengerInput) []PassengerInput {
	if len(list) > 0 {
		passengers := make([]PassengerInput, 0, len(list))
		for _, passenger := range list {
			passengers = append(passengers, PassengerInput{
				Name:     strings.TrimSpace(passenger.Name),
				Document: strings.TrimSpace(passenger.Document),
				Phone:    strings.TrimSpace(passenger.Phone),
				Email:    strings.TrimSpace(passenger.Email),
			})
		}
		return passengers
	}
	if !hasPassengerPayload(primary) {
		return nil
	}
	return []PassengerInput{{
		Name:     strings.TrimSpace(primary.Name),
		Document: strings.TrimSpace(primary.Document),
		Phone:    strings.TrimSpace(primary.Phone),
		Email:    strings.TrimSpace(primary.Email),
	}}
}

func hasPassengerPayload(passenger PassengerInput) bool {
	return strings.TrimSpace(passenger.Name) != "" ||
		strings.TrimSpace(passenger.Document) != "" ||
		strings.TrimSpace(passenger.Phone) != "" ||
		strings.TrimSpace(passenger.Email) != ""
}
