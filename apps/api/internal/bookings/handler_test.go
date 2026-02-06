package bookings

import (
  "net/http/httptest"
  "testing"
)

func TestParseListFilterBookingsValid(t *testing.T) {
  req := httptest.NewRequest("GET", "/bookings?limit=50&offset=25", nil)
  filter, err := parseListFilter(req)
  if err != nil {
    t.Fatalf("unexpected error: %v", err)
  }
  if filter.Limit != 50 || filter.Offset != 25 {
    t.Fatalf("unexpected filter: %+v", filter)
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

func TestIsValidBookingStatus(t *testing.T) {
  if !isValidBookingStatus("PENDING") {
    t.Fatalf("expected PENDING to be valid")
  }
  if isValidBookingStatus("INVALID") {
    t.Fatalf("expected INVALID to be invalid")
  }
}
