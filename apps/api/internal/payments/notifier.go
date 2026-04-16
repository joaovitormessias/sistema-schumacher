package payments

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"schumacher-tur/api/internal/shared/config"
)

type paymentConfirmationNotifier interface {
	NotifyPaymentConfirmed(context.Context, PaymentNotificationPayload) error
}

type noopPaymentConfirmationNotifier struct{}

func (noopPaymentConfirmationNotifier) NotifyPaymentConfirmed(context.Context, PaymentNotificationPayload) error {
	return nil
}

type webhookPaymentConfirmationNotifier struct {
	url  string
	http *http.Client
}

func newPaymentConfirmationNotifier(cfg config.Config) paymentConfirmationNotifier {
	if strings.TrimSpace(cfg.PaymentNotificationWebhookURL) == "" {
		return noopPaymentConfirmationNotifier{}
	}
	return &webhookPaymentConfirmationNotifier{
		url:  strings.TrimSpace(cfg.PaymentNotificationWebhookURL),
		http: &http.Client{Timeout: 15 * time.Second},
	}
}

func (n *webhookPaymentConfirmationNotifier) NotifyPaymentConfirmed(ctx context.Context, payload PaymentNotificationPayload) error {
	if n == nil || strings.TrimSpace(n.url) == "" {
		return nil
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, n.url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := n.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	snippet, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
	return fmt.Errorf("payment notification webhook returned %d: %s", resp.StatusCode, strings.TrimSpace(string(snippet)))
}
