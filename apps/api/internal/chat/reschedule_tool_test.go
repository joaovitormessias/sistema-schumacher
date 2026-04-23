package chat

import (
	"context"
	"errors"
	"testing"
	"time"

	"schumacher-tur/api/internal/bookings"
	"schumacher-tur/api/internal/reports"
)

type fakeBookingLookupBackend struct {
	items []bookings.BookingListItem
	err   error
}

func (f *fakeBookingLookupBackend) Enabled() bool {
	return f != nil
}

func (f *fakeBookingLookupBackend) Search(_ context.Context, input BookingLookupInput) (BookingLookupResult, error) {
	if f.err != nil {
		return BookingLookupResult{}, f.err
	}
	result := BookingLookupResult{
		Filter:  input,
		Results: make([]BookingLookupItem, 0, len(f.items)),
	}
	for _, item := range f.items {
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

type fakePassengerReportsBackend struct {
	rows []reports.PassengerReportRow
	err  error
}

func (f *fakePassengerReportsBackend) ListPassengers(_ context.Context, _ reports.PassengerReportFilter) ([]reports.PassengerReportRow, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.rows, nil
}

type fakeAvailabilityBackend struct {
	result AvailabilitySearchResult
	err    error
}

func (f *fakeAvailabilityBackend) Enabled() bool {
	return f != nil
}

func (f *fakeAvailabilityBackend) Search(_ context.Context, input AvailabilitySearchInput) (AvailabilitySearchResult, error) {
	if f.err != nil {
		return AvailabilitySearchResult{}, f.err
	}
	result := f.result
	result.Filter = input
	return result, nil
}

func TestRescheduleAssistToolSearchReturnsOptions(t *testing.T) {
	tool := NewRescheduleAssistTool(
		&fakeBookingLookupBackend{
			items: []bookings.BookingListItem{
				{
					ID:              "BK-ABC123456",
					TripID:          "trip-1",
					Status:          "PENDING",
					ReservationCode: "ABC12345",
					CreatedAt:       time.Now().UTC(),
				},
			},
		},
		&fakePassengerReportsBackend{
			rows: []reports.PassengerReportRow{
				{BookingID: "BK-ABC123456", TripDate: "2026-06-10", Origin: "Videira/SC", Destination: "Sao Luis/MA"},
				{BookingID: "BK-ABC123456", TripDate: "2026-06-10", Origin: "Videira/SC", Destination: "Sao Luis/MA"},
			},
		},
		&fakeAvailabilityBackend{
			result: AvailabilitySearchResult{
				Results: []AvailabilitySearchItem{
					{
						TripID:                 "trip-2",
						TripDate:               "2026-06-12",
						OriginDepartTime:       "18:30",
						OriginDisplayName:      "Videira/SC",
						DestinationDisplayName: "Sao Luis/MA",
						BoardStopID:            "stop-1",
						AlightStopID:           "stop-2",
						SeatsAvailable:         6,
						Price:                  950,
						Currency:               "BRL",
						PackageName:            "Convencional",
					},
				},
			},
		},
	)

	requestedDate := time.Date(2026, time.June, 12, 0, 0, 0, 0, time.UTC)
	result, err := tool.Search(context.Background(), RescheduleAssistInput{
		ReservationCode:   "ABC12345",
		RequestedTripDate: &requestedDate,
	})
	if err != nil {
		t.Fatalf("search reschedule assist: %v", err)
	}
	if result.Mode != "manual_review_required_with_options" {
		t.Fatalf("expected with options mode, got %s", result.Mode)
	}
	if len(result.Options) != 1 {
		t.Fatalf("expected one option, got %d", len(result.Options))
	}
	if result.Requested.Origin != "Videira/SC" || result.Requested.Destination != "Sao Luis/MA" {
		t.Fatalf("expected requested route derived from booking report, got %+v", result.Requested)
	}
	if result.Requested.Qty != 2 {
		t.Fatalf("expected requested qty 2, got %d", result.Requested.Qty)
	}
	if !result.HumanReviewRequired || result.CanAutoReschedule {
		t.Fatalf("expected manual review result")
	}
}

func TestRescheduleAssistToolSearchBlocksIneligibleBooking(t *testing.T) {
	tool := NewRescheduleAssistTool(
		&fakeBookingLookupBackend{
			items: []bookings.BookingListItem{
				{
					ID:              "BK-ABC123456",
					TripID:          "trip-1",
					Status:          "CANCELLED",
					ReservationCode: "ABC12345",
					CreatedAt:       time.Now().UTC(),
				},
			},
		},
		&fakePassengerReportsBackend{
			rows: []reports.PassengerReportRow{
				{BookingID: "BK-ABC123456", TripDate: "2026-06-10", Origin: "Videira/SC", Destination: "Sao Luis/MA"},
			},
		},
		&fakeAvailabilityBackend{},
	)

	requestedDate := time.Date(2026, time.June, 12, 0, 0, 0, 0, time.UTC)
	result, err := tool.Search(context.Background(), RescheduleAssistInput{
		ReservationCode:   "ABC12345",
		RequestedTripDate: &requestedDate,
	})
	if err != nil {
		t.Fatalf("search reschedule assist: %v", err)
	}
	if result.Mode != "manual_review_required_booking_ineligible" {
		t.Fatalf("expected ineligible mode, got %s", result.Mode)
	}
	if len(result.Options) != 0 {
		t.Fatalf("expected no options for blocked booking, got %d", len(result.Options))
	}
}

func TestRescheduleAssistToolSearchReturnsNoAvailabilityMode(t *testing.T) {
	tool := NewRescheduleAssistTool(
		&fakeBookingLookupBackend{
			items: []bookings.BookingListItem{
				{
					ID:              "BK-ABC123456",
					TripID:          "trip-1",
					Status:          "PENDING",
					ReservationCode: "ABC12345",
					CreatedAt:       time.Now().UTC(),
				},
			},
		},
		&fakePassengerReportsBackend{
			rows: []reports.PassengerReportRow{
				{BookingID: "BK-ABC123456", TripDate: "2026-06-10", Origin: "Videira/SC", Destination: "Sao Luis/MA"},
			},
		},
		&fakeAvailabilityBackend{
			result: AvailabilitySearchResult{Results: []AvailabilitySearchItem{}},
		},
	)

	requestedDate := time.Date(2026, time.June, 15, 0, 0, 0, 0, time.UTC)
	result, err := tool.Search(context.Background(), RescheduleAssistInput{
		ReservationCode:   "ABC12345",
		RequestedTripDate: &requestedDate,
	})
	if err != nil {
		t.Fatalf("search reschedule assist: %v", err)
	}
	if result.Mode != "manual_review_required_no_availability" {
		t.Fatalf("expected no availability mode, got %s", result.Mode)
	}
}

func TestRescheduleAssistToolSearchReturnsErrorWhenReportsFail(t *testing.T) {
	tool := NewRescheduleAssistTool(
		&fakeBookingLookupBackend{
			items: []bookings.BookingListItem{
				{
					ID:              "BK-ABC123456",
					TripID:          "trip-1",
					Status:          "PENDING",
					ReservationCode: "ABC12345",
					CreatedAt:       time.Now().UTC(),
				},
			},
		},
		&fakePassengerReportsBackend{err: errors.New("reports timeout")},
		&fakeAvailabilityBackend{},
	)

	requestedDate := time.Date(2026, time.June, 15, 0, 0, 0, 0, time.UTC)
	if _, err := tool.Search(context.Background(), RescheduleAssistInput{
		ReservationCode:   "ABC12345",
		RequestedTripDate: &requestedDate,
	}); err == nil {
		t.Fatalf("expected reports error")
	}
}

var (
	_ BookingLookupSearcher    = (*fakeBookingLookupBackend)(nil)
	_ RescheduleAssistSearcher = (*RescheduleAssistTool)(nil)
	_ AvailabilitySearcher     = (*fakeAvailabilityBackend)(nil)
)
