package chat

import (
	"context"

	"schumacher-tur/api/internal/bookings"
)

type BookingLookupTool struct {
	svc interface {
		List(ctx context.Context, filter bookings.ListFilter) ([]bookings.BookingListItem, error)
	}
}

func NewBookingLookupTool(svc interface {
	List(ctx context.Context, filter bookings.ListFilter) ([]bookings.BookingListItem, error)
}) *BookingLookupTool {
	return &BookingLookupTool{svc: svc}
}

func (t *BookingLookupTool) Enabled() bool {
	return t != nil && t.svc != nil
}

func (t *BookingLookupTool) Search(ctx context.Context, input BookingLookupInput) (BookingLookupResult, error) {
	if !t.Enabled() {
		return BookingLookupResult{}, ErrAgentToolNotConfigured
	}

	filter := bookings.ListFilter{
		BookingID:       input.BookingID,
		ReservationCode: input.ReservationCode,
		Limit:           input.Limit,
	}
	items, err := t.svc.List(ctx, filter)
	if err != nil {
		return BookingLookupResult{}, err
	}

	result := BookingLookupResult{
		Filter:  input,
		Results: make([]BookingLookupItem, 0, len(items)),
	}
	for _, item := range items {
		result.Results = append(result.Results, BookingLookupItem{
			ID:              item.ID,
			TripID:          item.TripID,
			Status:          item.Status,
			ReservationCode: item.ReservationCode,
			TotalAmount:     item.TotalAmount,
			DepositAmount:   item.DepositAmount,
			RemainderAmount: item.RemainderAmount,
			PassengerName:   item.PassengerName,
			PassengerPhone:  item.PassengerPhone,
			SeatNumber:      item.SeatNumber,
			ExpiresAt:       item.ExpiresAt,
			CreatedAt:       item.CreatedAt,
		})
	}

	return result, nil
}
