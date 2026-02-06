package payments

import "testing"

func TestParseTimeParam(t *testing.T) {
  if _, err := parseTimeParam("2026-02-03T10:00:00Z"); err != nil {
    t.Fatalf("expected valid RFC3339, got %v", err)
  }
  if _, err := parseTimeParam("2026-02-03"); err != nil {
    t.Fatalf("expected valid date, got %v", err)
  }
  if _, err := parseTimeParam("invalid"); err == nil {
    t.Fatalf("expected error for invalid date")
  }
}

func TestIsValidPaymentStatus(t *testing.T) {
  if !isValidPaymentStatus("PAID") {
    t.Fatalf("expected PAID to be valid")
  }
  if isValidPaymentStatus("UNKNOWN") {
    t.Fatalf("expected UNKNOWN to be invalid")
  }
}
