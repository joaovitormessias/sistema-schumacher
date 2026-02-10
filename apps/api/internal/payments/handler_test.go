package payments

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/http/httptest"
	"testing"
)

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

func TestParseWebhookExtractBillingID(t *testing.T) {
	tests := []struct {
		name    string
		payload string
		want    string
	}{
		{
			name:    "nested billing id",
			payload: `{"event":"billing.paid","data":{"billing":{"id":"bill-123"}}}`,
			want:    "bill-123",
		},
		{
			name:    "flat billing id",
			payload: `{"event":"billing.paid","data":{"billingId":"bill-456"}}`,
			want:    "bill-456",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			evt, err := ParseWebhook([]byte(tc.payload))
			if err != nil {
				t.Fatalf("unexpected parse error: %v", err)
			}
			if evt.BillingID != tc.want {
				t.Fatalf("expected %q, got %q", tc.want, evt.BillingID)
			}
		})
	}
}

func TestWebhookSecretAndSignatureHelpers(t *testing.T) {
	req := httptest.NewRequest("POST", "/webhooks/abacatepay?webhookSecret=secret123", nil)
	req.Header.Set("X-Webhook-Secret", "secret123")
	req.Header.Set("X-AbacatePay-Signature", "sig-abc")

	if got := WebhookSecretFromQuery(req.URL.RawQuery); got != "secret123" {
		t.Fatalf("expected webhookSecret from query, got %q", got)
	}
	if !secretMatches(req, "secret123") {
		t.Fatalf("expected header secret to match")
	}
	if got := webhookSignatureFromHeaders(req); got != "sig-abc" {
		t.Fatalf("expected signature header, got %q", got)
	}
}

func TestVerifyWebhookSignature(t *testing.T) {
	secret := "super-secret"
	body := []byte(`{"event":"billing.paid"}`)
	signature := signPayload(secret, body)

	if !VerifyWebhookSignature(secret, body, signature) {
		t.Fatalf("expected valid signature")
	}
	if VerifyWebhookSignature(secret, body, "wrong-signature") {
		t.Fatalf("expected invalid signature to fail")
	}
}

func TestVerifyWebhookSignatureWithFallback(t *testing.T) {
	body := []byte(`{"event":"billing.paid","data":{"id":"bill-1"}}`)
	t.Run("accept signature with public key when both keys configured", func(t *testing.T) {
		publicKey := "public-key-123"
		secret := "secret-key-123"
		signature := signPayload(publicKey, body)

		if !verifyWebhookSignatureWithFallback(signature, body, publicKey, secret) {
			t.Fatalf("expected signature validation to pass with public key")
		}
	})

	t.Run("fallback to secret when public key is invalid", func(t *testing.T) {
		publicKey := "invalid-public-key"
		secret := "secret-key-123"
		signature := signPayload(secret, body)

		if !verifyWebhookSignatureWithFallback(signature, body, publicKey, secret) {
			t.Fatalf("expected signature validation to pass with secret fallback")
		}
	})

	t.Run("reject invalid signatures for both keys", func(t *testing.T) {
		publicKey := "public-key-123"
		secret := "secret-key-123"
		signature := signPayload("different-key", body)

		if verifyWebhookSignatureWithFallback(signature, body, publicKey, secret) {
			t.Fatalf("expected signature validation to fail")
		}
	})

	t.Run("reject missing signature when signature validation is active", func(t *testing.T) {
		publicKey := "public-key-123"
		secret := "secret-key-123"

		if verifyWebhookSignatureWithFallback("", body, publicKey, secret) {
			t.Fatalf("expected empty signature to fail")
		}
	})
}

func signPayload(key string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write(body)
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}
