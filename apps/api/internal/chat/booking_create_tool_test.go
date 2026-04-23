package chat

import (
	"context"
	"errors"
	"testing"

	"schumacher-tur/api/internal/bookings"
)

type fakeBookingCreateService struct {
	result bookings.BookingDetails
	err    error
}

func (f *fakeBookingCreateService) Create(_ context.Context, _ bookings.CreateBookingInput) (bookings.BookingDetails, error) {
	if f.err != nil {
		return bookings.BookingDetails{}, f.err
	}
	return f.result, nil
}

func TestBookingCreateToolMapsValidationErrorsToOperationalResult(t *testing.T) {
	tool := NewBookingCreateTool(&fakeBookingCreateService{err: bookings.ErrPassengerDocumentType})

	result, err := tool.Create(context.Background(), BookingCreateInput{
		TripID:       "trip-1",
		BoardStopID:  "board-1",
		AlightStopID: "alight-1",
		Passengers: []BookingCreatePassengerInput{
			{Name: "Joao", Document: "123", DocumentType: "INVALID"},
		},
	})
	if err != nil {
		t.Fatalf("expected no fatal error, got %v", err)
	}
	if result.Mode != "manual_review_required_validation_error" {
		t.Fatalf("expected validation mode, got %s", result.Mode)
	}
	if len(result.Errors) != 1 {
		t.Fatalf("expected one validation error, got %+v", result.Errors)
	}
}

func TestBookingCreateToolReturnsUnexpectedErrors(t *testing.T) {
	tool := NewBookingCreateTool(&fakeBookingCreateService{err: errors.New("db down")})

	_, err := tool.Create(context.Background(), BookingCreateInput{
		TripID:       "trip-1",
		BoardStopID:  "board-1",
		AlightStopID: "alight-1",
		Passengers: []BookingCreatePassengerInput{
			{Name: "Joao", Document: "06645648105", DocumentType: "CPF"},
		},
	})
	if err == nil {
		t.Fatalf("expected fatal error for unexpected failure")
	}
}
