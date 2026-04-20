package payments

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestParseProviderData(t *testing.T) {
	raw := []byte(`{"data":{"url":"https://checkout.example/pix","pixQrCode":"000201abc"}}`)
	parsed, checkoutURL, pixCode := parseProviderData(raw)

	if parsed == nil {
		t.Fatalf("expected parsed provider payload")
	}
	if checkoutURL == nil || *checkoutURL != "https://checkout.example/pix" {
		t.Fatalf("unexpected checkout url: %#v", checkoutURL)
	}
	if pixCode == nil || *pixCode != "000201abc" {
		t.Fatalf("unexpected pix code: %#v", pixCode)
	}
}

func TestBuildCustomerSynthesizesEmailWhenMissing(t *testing.T) {
	customer := BuildCustomer(&CustomerInput{
		Name:     "Joao Vitor Messias",
		Email:    "",
		Phone:    "554988709047",
		Document: "06645648103",
	}, "BK-C9A55BC8C67846F9B09EF5EFFC576A50")

	if customer == nil {
		t.Fatalf("expected customer payload")
	}
	if customer.Email != "reserva.bkc9a55bc8c67846f9b09ef5effc576a50@schumachertur.com" {
		t.Fatalf("unexpected fallback email: %q", customer.Email)
	}
}

func TestWebhookPaymentConfirmationNotifierPostsPayload(t *testing.T) {
	var gotMethod string
	var gotContentType string
	var gotPayload PaymentNotificationPayload

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotContentType = r.Header.Get("Content-Type")
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&gotPayload); err != nil {
			t.Fatalf("could not decode payload: %v", err)
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	notifier := &webhookPaymentConfirmationNotifier{
		url:  server.URL,
		http: server.Client(),
	}

	payload := PaymentNotificationPayload{
		Event:           "payment.confirmed",
		SentAt:          time.Date(2026, 4, 15, 12, 0, 0, 0, time.UTC),
		PaymentID:       "pay-1",
		PaymentAmount:   250,
		PaymentMethod:   "PIX",
		BookingID:       "BK-1",
		ReservationCode: "ABCD1234",
		CustomerName:    "Joao",
		CustomerPhone:   "554998887766",
		AmountTotal:     1100,
		AmountPaid:      250,
		AmountDue:       850,
		PaymentStatus:   "PARTIAL",
	}

	if err := notifier.NotifyPaymentConfirmed(context.Background(), payload); err != nil {
		t.Fatalf("unexpected notify error: %v", err)
	}
	if gotMethod != http.MethodPost {
		t.Fatalf("expected POST, got %s", gotMethod)
	}
	if gotContentType != "application/json" {
		t.Fatalf("expected application/json, got %s", gotContentType)
	}
	if gotPayload.BookingID != payload.BookingID || gotPayload.AmountDue != payload.AmountDue || gotPayload.PaymentStatus != payload.PaymentStatus {
		t.Fatalf("unexpected payload: %+v", gotPayload)
	}
}
