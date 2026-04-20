package availability

import (
	"net/http/httptest"
	"testing"
)

func TestParseSearchFilterNormalizesLocationAndDefaults(t *testing.T) {
	req := httptest.NewRequest("GET", "/availability/search?origin=Videira%20SC&destination=Sao%20Luis/MA", nil)

	filter, err := parseSearchFilter(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if filter.Origin != "Videira/SC" {
		t.Fatalf("unexpected origin: %q", filter.Origin)
	}
	if filter.Destination != "Sao Luis/MA" {
		t.Fatalf("unexpected destination: %q", filter.Destination)
	}
	if filter.Qty != 1 || filter.Limit != 10 || !filter.OnlyActive || filter.IncludePast {
		t.Fatalf("unexpected defaults: %+v", filter)
	}
}

func TestParseSearchFilterRejectsMissingSearchTerms(t *testing.T) {
	req := httptest.NewRequest("GET", "/availability/search", nil)

	if _, err := parseSearchFilter(req); err == nil {
		t.Fatalf("expected validation error")
	}
}

func TestParseSearchFilterRejectsInvalidDate(t *testing.T) {
	req := httptest.NewRequest("GET", "/availability/search?origin=Videira/SC&trip_date=08-04-2026", nil)

	if _, err := parseSearchFilter(req); err == nil {
		t.Fatalf("expected invalid date error")
	}
}

func TestParseSearchFilterRejectsInvalidBool(t *testing.T) {
	req := httptest.NewRequest("GET", "/availability/search?origin=Videira/SC&only_active=yes", nil)

	if _, err := parseSearchFilter(req); err == nil {
		t.Fatalf("expected invalid bool error")
	}
}
