package routes

import (
  "net/http/httptest"
  "testing"
)

func TestParseListFilterRoutesValid(t *testing.T) {
  req := httptest.NewRequest("GET", "/routes?limit=10&offset=20", nil)
  filter, err := parseListFilter(req)
  if err != nil {
    t.Fatalf("unexpected error: %v", err)
  }
  if filter.Limit != 10 || filter.Offset != 20 {
    t.Fatalf("unexpected filter: %+v", filter)
  }
}

func TestParseListFilterRoutesInvalidLimit(t *testing.T) {
  req := httptest.NewRequest("GET", "/routes?limit=0", nil)
  if _, err := parseListFilter(req); err == nil {
    t.Fatalf("expected error for invalid limit")
  }
}

func TestParseListFilterRoutesInvalidOffset(t *testing.T) {
  req := httptest.NewRequest("GET", "/routes?offset=-1", nil)
  if _, err := parseListFilter(req); err == nil {
    t.Fatalf("expected error for invalid offset")
  }
}
