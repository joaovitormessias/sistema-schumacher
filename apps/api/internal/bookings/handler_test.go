package bookings

import (
	"net/http/httptest"
	"testing"
)

func TestParseListFilterBookingsValid(t *testing.T) {
	req := httptest.NewRequest("GET", "/bookings?limit=50&offset=25&booking_id=BK-1&reservation_code=RSV1&trip_id=TRIP-1&status=confirmed", nil)
	filter, err := parseListFilter(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filter.Limit != 50 || filter.Offset != 25 {
		t.Fatalf("unexpected pagination: %+v", filter)
	}
	if filter.BookingID != "BK-1" || filter.ReservationCode != "RSV1" || filter.TripID != "TRIP-1" || filter.Status != "CONFIRMED" {
		t.Fatalf("unexpected filter fields: %+v", filter)
	}
}

func TestParseListFilterBookingsInvalidLimit(t *testing.T) {
	req := httptest.NewRequest("GET", "/bookings?limit=0", nil)
	if _, err := parseListFilter(req); err == nil {
		t.Fatalf("expected error for invalid limit")
	}
}

func TestParseListFilterBookingsInvalidOffset(t *testing.T) {
	req := httptest.NewRequest("GET", "/bookings?offset=-1", nil)
	if _, err := parseListFilter(req); err == nil {
		t.Fatalf("expected error for invalid offset")
	}
}

func TestParseListFilterBookingsInvalidStatus(t *testing.T) {
	req := httptest.NewRequest("GET", "/bookings?status=broken", nil)
	if _, err := parseListFilter(req); err == nil {
		t.Fatalf("expected error for invalid status")
	}
}

func TestIsValidBookingStatus(t *testing.T) {
	if !isValidBookingStatus("PENDING") {
		t.Fatalf("expected PENDING to be valid")
	}
	if isValidBookingStatus("INVALID") {
		t.Fatalf("expected INVALID to be invalid")
	}
}
