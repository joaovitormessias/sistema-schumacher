package buses

import (
  "net/http/httptest"
  "testing"
)

func TestParseListFilterBusesValid(t *testing.T) {
  req := httptest.NewRequest("GET", "/buses?limit=15&offset=5", nil)
  filter, err := parseListFilter(req)
  if err != nil {
    t.Fatalf("unexpected error: %v", err)
  }
  if filter.Limit != 15 || filter.Offset != 5 {
    t.Fatalf("unexpected filter: %+v", filter)
  }
}

func TestParseListFilterBusesInvalidLimit(t *testing.T) {
  req := httptest.NewRequest("GET", "/buses?limit=0", nil)
  if _, err := parseListFilter(req); err == nil {
    t.Fatalf("expected error for invalid limit")
  }
}

func TestParseListFilterBusesInvalidOffset(t *testing.T) {
  req := httptest.NewRequest("GET", "/buses?offset=-2", nil)
  if _, err := parseListFilter(req); err == nil {
    t.Fatalf("expected error for invalid offset")
  }
}
