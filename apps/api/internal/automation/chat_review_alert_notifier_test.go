package automation

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"schumacher-tur/api/internal/shared/config"
)

func TestWebhookChatReviewAlertNotifierPostsPayload(t *testing.T) {
	var received ChatReviewAlertNotificationPayload
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	notifier := newChatReviewAlertNotifier(config.Config{
		ChatReviewAlertWebhookURL: server.URL,
	})

	payload := ChatReviewAlertNotificationPayload{
		Source:            "chat_review_queue",
		ObservedAt:        time.Date(2026, 4, 22, 15, 0, 0, 0, time.UTC),
		AlertLevel:        "CRITICAL",
		AlertCode:         "REVIEW_QUEUE_OVERDUE",
		AlertMessage:      "Fila com drafts acima do SLA de revisao.",
		AlertSessionCount: 2,
		Summary: ChatReviewAlertSummarySnapshot{
			OverdueReviewCount:      2,
			OldestPendingAgeSeconds: 1800,
			ReviewSLASeconds:        900,
			HasReviewAlert:          true,
		},
	}

	if err := notifier.NotifyReviewAlert(context.Background(), payload); err != nil {
		t.Fatalf("notify review alert: %v", err)
	}
	if received.AlertCode != payload.AlertCode {
		t.Fatalf("expected alert code %s, got %s", payload.AlertCode, received.AlertCode)
	}
	if received.AlertSessionCount != 2 {
		t.Fatalf("expected alert session count 2, got %d", received.AlertSessionCount)
	}
}
