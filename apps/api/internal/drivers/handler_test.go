package drivers

import (
  "net/http/httptest"
  "testing"
)

func TestParseListFilterDriversValid(t *testing.T) {
  req := httptest.NewRequest("GET", "/drivers?limit=25&offset=0", nil)
  filter, err := parseListFilter(req)
  if err != nil {
    t.Fatalf("unexpected error: %v", err)
  }
  if filter.Limit != 25 || filter.Offset != 0 {
    t.Fatalf("unexpected filter: %+v", filter)
  }
}

func TestParseListFilterDriversInvalidLimit(t *testing.T) {
  req := httptest.NewRequest("GET", "/drivers?limit=-1", nil)
  if _, err := parseListFilter(req); err == nil {
    t.Fatalf("expected error for invalid limit")
  }
}

func TestParseListFilterDriversInvalidOffset(t *testing.T) {
  req := httptest.NewRequest("GET", "/drivers?offset=-1", nil)
  if _, err := parseListFilter(req); err == nil {
    t.Fatalf("expected error for invalid offset")
  }
}
