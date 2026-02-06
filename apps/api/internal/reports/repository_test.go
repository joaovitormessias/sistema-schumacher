package reports

import "testing"

func TestPaymentStage(t *testing.T) {
  if stage := paymentStage(100, 100, 20); stage != "PAID" {
    t.Fatalf("expected PAID, got %s", stage)
  }
  if stage := paymentStage(20, 100, 20); stage != "DEPOSIT" {
    t.Fatalf("expected DEPOSIT, got %s", stage)
  }
  if stage := paymentStage(10, 100, 20); stage != "PENDING" {
    t.Fatalf("expected PENDING, got %s", stage)
  }
}
