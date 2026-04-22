package automation

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

type noopChatReviewAlertNotifier struct{}

func (noopChatReviewAlertNotifier) Enabled() bool {
	return false
}

func (noopChatReviewAlertNotifier) NotifyReviewAlert(context.Context, ChatReviewAlertNotificationPayload) error {
	return nil
}

type webhookChatReviewAlertNotifier struct {
	url  string
	http *http.Client
}

func newChatReviewAlertNotifier(cfg config.Config) chatReviewAlertNotifier {
	if strings.TrimSpace(cfg.ChatReviewAlertWebhookURL) == "" {
		return noopChatReviewAlertNotifier{}
	}
	return &webhookChatReviewAlertNotifier{
		url:  strings.TrimSpace(cfg.ChatReviewAlertWebhookURL),
		http: &http.Client{Timeout: 15 * time.Second},
	}
}

func (n *webhookChatReviewAlertNotifier) Enabled() bool {
	return n != nil && strings.TrimSpace(n.url) != ""
}

func (n *webhookChatReviewAlertNotifier) NotifyReviewAlert(ctx context.Context, payload ChatReviewAlertNotificationPayload) error {
	if !n.Enabled() {
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
	return fmt.Errorf("chat review alert webhook returned %d: %s", resp.StatusCode, strings.TrimSpace(string(snippet)))
}
