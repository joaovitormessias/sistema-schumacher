package chat

import (
	"context"
	"testing"

	"schumacher-tur/api/internal/bookings"
)

type fakeBookingCancelService struct {
	getResult    bookings.BookingDetails
	getErr       error
	updateResult bookings.BookingDetails
	updateErr    error
	updateCalls  int
}

func (f *fakeBookingCancelService) Get(_ context.Context, _ string) (bookings.BookingDetails, error) {
	if f.getErr != nil {
		return bookings.BookingDetails{}, f.getErr
	}
	return f.getResult, nil
}

func (f *fakeBookingCancelService) UpdateStatus(_ context.Context, _ string, _ string) (bookings.BookingDetails, error) {
	f.updateCalls++
	if f.updateErr != nil {
		return bookings.BookingDetails{}, f.updateErr
	}
	return f.updateResult, nil
}

func TestBookingCancelToolCancelsActiveBooking(t *testing.T) {
	service := &fakeBookingCancelService{
		getResult: bookings.BookingDetails{
			Booking: bookings.Booking{
				ID:              "BK-ABC123456",
				ReservationCode: "ABC12345",
				TripID:          "trip-1",
				Status:          "PENDING",
			},
			Passengers: []bookings.BookingPassenger{{ID: "PS-1"}, {ID: "PS-2"}},
		},
		updateResult: bookings.BookingDetails{
			Booking: bookings.Booking{
				ID:              "BK-ABC123456",
				ReservationCode: "ABC12345",
				TripID:          "trip-1",
				Status:          "CANCELLED",
			},
			Passengers: []bookings.BookingPassenger{{ID: "PS-1"}, {ID: "PS-2"}},
		},
	}
	tool := NewBookingCancelTool(service)

	result, err := tool.Cancel(context.Background(), BookingCancelInput{
		BookingID:       "BK-ABC123456",
		ReservationCode: "ABC12345",
		Reason:          "customer_requested",
	})
	if err != nil {
		t.Fatalf("cancel booking: %v", err)
	}
	if result.Mode != "cancel" {
		t.Fatalf("expected cancel mode, got %s", result.Mode)
	}
	if result.BookingStatus != "CANCELLED" {
		t.Fatalf("expected cancelled status, got %s", result.BookingStatus)
	}
	if result.PreviousStatus != "PENDING" {
		t.Fatalf("expected previous status PENDING, got %s", result.PreviousStatus)
	}
	if service.updateCalls != 1 {
		t.Fatalf("expected one update call, got %d", service.updateCalls)
	}
}

func TestBookingCancelToolReturnsIdempotentResultForClosedBooking(t *testing.T) {
	service := &fakeBookingCancelService{
		getResult: bookings.BookingDetails{
			Booking: bookings.Booking{
				ID:              "BK-ABC123456",
				ReservationCode: "ABC12345",
				TripID:          "trip-1",
				Status:          "CANCELLED",
			},
			Passengers: []bookings.BookingPassenger{{ID: "PS-1"}},
		},
	}
	tool := NewBookingCancelTool(service)

	result, err := tool.Cancel(context.Background(), BookingCancelInput{
		BookingID:       "BK-ABC123456",
		ReservationCode: "ABC12345",
		Reason:          "customer_requested",
	})
	if err != nil {
		t.Fatalf("cancel booking: %v", err)
	}
	if result.Mode != "already_closed" {
		t.Fatalf("expected already_closed mode, got %s", result.Mode)
	}
	if !result.Idempotent {
		t.Fatalf("expected idempotent result")
	}
	if service.updateCalls != 0 {
		t.Fatalf("expected no update call, got %d", service.updateCalls)
	}
}
