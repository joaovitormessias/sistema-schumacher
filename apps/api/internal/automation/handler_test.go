package automation

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"schumacher-tur/api/internal/bookings"
	"schumacher-tur/api/internal/chat"
	"schumacher-tur/api/internal/payments"
	"schumacher-tur/api/internal/shared/config"
)

func TestHandleEvolutionMessagesAcceptsDirectPayload(t *testing.T) {
	chatSvc := &fakeChatIngestor{}
	handler := NewHandler(NewService(&fakeAutomationStore{}, chatSvc, config.Config{}))

	r := chi.NewRouter()
	handler.RegisterWebhooks(r)

	req := httptest.NewRequest(http.MethodPost, "/webhooks/evolution/messages", bytes.NewBufferString(`{
		"event":"messages.upsert",
		"instance":"belle",
		"data":{
			"key":{"remoteJid":"554988709047@s.whatsapp.net","fromMe":false,"id":"MSG-1"},
			"pushName":"messias",
			"message":{"conversation":"06645648109"},
			"messageType":"conversation",
			"messageTimestamp":1772544357,
			"source":"ios"
		},
		"server_url":"https://evo.bellasoftware.com.br"
	}`))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d", http.StatusAccepted, rec.Code)
	}
	if chatSvc.lastInput.ContactKey != "554988709047@s.whatsapp.net" {
		t.Fatalf("unexpected contact key: %s", chatSvc.lastInput.ContactKey)
	}
	if chatSvc.lastInput.Message.Body != "06645648109" {
		t.Fatalf("unexpected message body: %s", chatSvc.lastInput.Message.Body)
	}
	if chatSvc.lastInput.Message.Kind != "TEXT" {
		t.Fatalf("unexpected kind: %s", chatSvc.lastInput.Message.Kind)
	}
	if chatSvc.lastInput.Message.ProviderMessageID != "MSG-1" {
		t.Fatalf("unexpected provider message id: %s", chatSvc.lastInput.Message.ProviderMessageID)
	}
}

func TestHandleEvolutionMessagesAcceptsN8NEnvelope(t *testing.T) {
	chatSvc := &fakeChatIngestor{}
	handler := NewHandler(NewService(&fakeAutomationStore{}, chatSvc, config.Config{}))

	r := chi.NewRouter()
	handler.RegisterWebhooks(r)

	req := httptest.NewRequest(http.MethodPost, "/webhooks/evolution/messages", bytes.NewBufferString(`{
		"body":{
			"event":"messages.upsert",
			"instance":"belle",
			"data":{
				"key":{"remoteJid":"554998208115@s.whatsapp.net","fromMe":false,"id":"MSG-2"},
				"pushName":"Cliente",
				"message":{"extendedTextMessage":{"text":"quero reservar"}},
				"messageType":"extendedTextMessage",
				"messageTimestamp":1772544357
			}
		}
	}`))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d", http.StatusAccepted, rec.Code)
	}
	if chatSvc.lastInput.Message.Body != "quero reservar" {
		t.Fatalf("unexpected extracted body: %s", chatSvc.lastInput.Message.Body)
	}
}

func TestHandleEvolutionMessagesSkipsFromMe(t *testing.T) {
	chatSvc := &fakeChatIngestor{}
	handler := NewHandler(NewService(&fakeAutomationStore{}, chatSvc, config.Config{}))

	r := chi.NewRouter()
	handler.RegisterWebhooks(r)

	req := httptest.NewRequest(http.MethodPost, "/webhooks/evolution/messages", bytes.NewBufferString(`{
		"event":"messages.upsert",
		"data":{
			"key":{"remoteJid":"554998208115@s.whatsapp.net","fromMe":true,"id":"MSG-3"},
			"messageType":"conversation"
		}
	}`))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d", http.StatusAccepted, rec.Code)
	}
	if chatSvc.calls != 0 {
		t.Fatalf("expected chat ingest not to be called, got %d", chatSvc.calls)
	}
}

func TestUnimplementedAutomationRoutesRemainUnimplemented(t *testing.T) {
	handler := NewHandler(NewService(&fakeAutomationStore{}, &fakeChatIngestor{}, config.Config{}))

	r := chi.NewRouter()
	handler.RegisterWebhooks(r)
	handler.RegisterRoutes(r)

	tests := []struct {
		method string
		path   string
	}{
		{method: http.MethodPost, path: "/automation/jobs/sheet-sync/run"},
	}

	for _, tc := range tests {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotImplemented {
			t.Fatalf("expected status %d, got %d for %s", http.StatusNotImplemented, rec.Code, tc.path)
		}
	}
}

func TestRunBookingsExpireRejectsInvalidInput(t *testing.T) {
	handler := NewHandler(NewService(&fakeAutomationStore{}, &fakeChatIngestor{}, config.Config{}, &fakeBookingExpirationSource{}))

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodPost, "/automation/jobs/bookings-expire/run", bytes.NewBufferString(`{"limit":-1}`))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestRunBookingsExpireProcessesExpiredBookings(t *testing.T) {
	store := &fakeAutomationStore{}
	bookingSvc := &fakeBookingExpirationSource{
		items: []bookings.BookingListItem{
			{
				ID:              "booking-1",
				TripID:          "trip-1",
				Status:          "PENDING",
				ReservationCode: "ABCD1234",
				CreatedAt:       time.Date(2026, 4, 22, 10, 0, 0, 0, time.UTC),
				ExpiresAt:       timePointer(time.Date(2026, 4, 22, 10, 45, 0, 0, time.UTC)),
			},
			{
				ID:              "booking-2",
				TripID:          "trip-2",
				Status:          "PENDING",
				ReservationCode: "EFGH5678",
				CreatedAt:       time.Date(2026, 4, 22, 11, 30, 0, 0, time.UTC),
				ExpiresAt:       timePointer(time.Date(2099, 4, 22, 12, 0, 0, 0, time.UTC)),
			},
			{
				ID:              "booking-3",
				TripID:          "trip-3",
				Status:          "PENDING",
				ReservationCode: "IJKL9012",
				CreatedAt:       time.Date(2026, 4, 22, 9, 0, 0, 0, time.UTC),
			},
		},
	}
	handler := NewHandler(NewService(store, &fakeChatIngestor{}, config.Config{}, bookingSvc))

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodPost, "/automation/jobs/bookings-expire/run", bytes.NewBufferString(`{"limit":500,"hold_minutes":50}`))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if bookingSvc.listCalls != 1 {
		t.Fatalf("expected list to be called once, got %d", bookingSvc.listCalls)
	}
	if bookingSvc.lastFilter.Status != "PENDING" {
		t.Fatalf("expected PENDING filter, got %s", bookingSvc.lastFilter.Status)
	}
	if len(bookingSvc.updateCalls) != 2 {
		t.Fatalf("expected two expired bookings, got %d", len(bookingSvc.updateCalls))
	}
	if bookingSvc.updateCalls[0].Status != "EXPIRED" || bookingSvc.updateCalls[1].Status != "EXPIRED" {
		t.Fatalf("expected EXPIRED updates, got %+v", bookingSvc.updateCalls)
	}
	if len(store.jobRuns) != 1 {
		t.Fatalf("expected one job run, got %d", len(store.jobRuns))
	}
	if store.jobRuns[0].Status != "COMPLETED" {
		t.Fatalf("expected COMPLETED job run, got %s", store.jobRuns[0].Status)
	}

	var out RunBookingsExpireResult
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	if out.Status != "processed" {
		t.Fatalf("expected processed status, got %s", out.Status)
	}
	if out.ExpiredCount != 2 {
		t.Fatalf("expected expired_count 2, got %d", out.ExpiredCount)
	}
	if out.CheckedCount != 3 {
		t.Fatalf("expected checked_count 3, got %d", out.CheckedCount)
	}
}

func TestRunBookingsExpireSkipsWhenNothingExpired(t *testing.T) {
	store := &fakeAutomationStore{}
	bookingSvc := &fakeBookingExpirationSource{
		items: []bookings.BookingListItem{
			{
				ID:              "booking-1",
				TripID:          "trip-1",
				Status:          "PENDING",
				ReservationCode: "ABCD1234",
				CreatedAt:       time.Now().UTC(),
				ExpiresAt:       timePointer(time.Now().UTC().Add(30 * time.Minute)),
			},
		},
	}
	handler := NewHandler(NewService(store, &fakeChatIngestor{}, config.Config{}, bookingSvc))

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodPost, "/automation/jobs/bookings-expire/run", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if len(bookingSvc.updateCalls) != 0 {
		t.Fatalf("expected no updates, got %d", len(bookingSvc.updateCalls))
	}
	if len(store.jobRuns) != 1 {
		t.Fatalf("expected one job run, got %d", len(store.jobRuns))
	}
	if store.jobRuns[0].Status != "SKIPPED" {
		t.Fatalf("expected SKIPPED job run, got %s", store.jobRuns[0].Status)
	}

	var out RunBookingsExpireResult
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	if out.Status != "skipped" {
		t.Fatalf("expected skipped status, got %s", out.Status)
	}
	if out.Reason != "no_expired_bookings" {
		t.Fatalf("expected no_expired_bookings, got %s", out.Reason)
	}
}

func TestRunPaymentNotificationsRejectsMissingTarget(t *testing.T) {
	handler := NewHandler(NewService(&fakeAutomationStore{}, &fakeChatIngestor{}, config.Config{}, &fakePaymentNotificationSource{}))

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodPost, "/automation/jobs/payment-notifications/run", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestRunPaymentNotificationsQueuesReviewDraft(t *testing.T) {
	store := &fakeAutomationStore{}
	chatSvc := &fakeChatIngestor{
		draftResult: chat.QueueAutomationDraftResult{
			Session: chat.Session{
				ID:            "session-1",
				Channel:       "WHATSAPP",
				ContactKey:    "554998887766@s.whatsapp.net",
				CustomerPhone: "554998887766",
				CustomerName:  "Joao",
				Status:        "ACTIVE",
				HandoffStatus: "BOT",
				CreatedAt:     time.Now().UTC(),
				UpdatedAt:     time.Now().UTC(),
			},
			Message: chat.Message{
				ID:               "draft-1",
				SessionID:        "session-1",
				Direction:        "OUTBOUND",
				Kind:             "TEXT",
				ProcessingStatus: "AUTOMATION_DRAFT",
				ReceivedAt:       time.Now().UTC(),
				CreatedAt:        time.Now().UTC(),
			},
		},
	}
	paymentSource := &fakePaymentNotificationSource{
		payment: payments.Payment{
			ID:        "pay-1",
			BookingID: "booking-1",
			Amount:    250,
			Method:    "PIX",
			Status:    "PAID",
		},
		notification: payments.PaymentNotificationContext{
			BookingID:       "booking-1",
			ReservationCode: "ABCD1234",
			CustomerName:    "Joao",
			CustomerPhone:   "54998887766",
			AmountTotal:     1100,
			AmountPaid:      250,
			AmountDue:       850,
			PaymentStatus:   "PARTIAL",
		},
	}
	handler := NewHandler(NewService(store, chatSvc, config.Config{}, paymentSource))

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodPost, "/automation/jobs/payment-notifications/run", bytes.NewBufferString(`{"payment_id":"pay-1"}`))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if chatSvc.draftCalls != 1 {
		t.Fatalf("expected draft to be queued once, got %d", chatSvc.draftCalls)
	}
	if chatSvc.lastDraftInput.ContactKey != "5554998887766@s.whatsapp.net" {
		t.Fatalf("unexpected contact key: %s", chatSvc.lastDraftInput.ContactKey)
	}
	if chatSvc.lastDraftInput.IdempotencyKey != "payment-notification-pay-1" {
		t.Fatalf("unexpected draft idempotency key: %s", chatSvc.lastDraftInput.IdempotencyKey)
	}
	if !strings.Contains(chatSvc.lastDraftInput.Body, "ABCD1234") {
		t.Fatalf("expected reservation code in draft body, got %q", chatSvc.lastDraftInput.Body)
	}
	if len(store.jobRuns) != 1 {
		t.Fatalf("expected one job run, got %d", len(store.jobRuns))
	}
	if store.jobRuns[0].Status != "QUEUED" {
		t.Fatalf("expected QUEUED job run, got %s", store.jobRuns[0].Status)
	}

	var out RunPaymentNotificationsResult
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	if out.Status != "queued" {
		t.Fatalf("expected status queued, got %s", out.Status)
	}
	if !out.DraftQueued {
		t.Fatalf("expected draft_queued true")
	}
	if out.DraftMessageID != "draft-1" {
		t.Fatalf("expected draft_message_id draft-1, got %s", out.DraftMessageID)
	}
	if out.JobRunID == "" {
		t.Fatalf("expected job_run_id to be present")
	}
}

func TestRunPaymentNotificationsSkipsWhenPhoneMissing(t *testing.T) {
	store := &fakeAutomationStore{}
	chatSvc := &fakeChatIngestor{}
	paymentSource := &fakePaymentNotificationSource{
		payment: payments.Payment{
			ID:        "pay-1",
			BookingID: "booking-1",
			Amount:    250,
			Method:    "PIX",
			Status:    "PAID",
		},
		notification: payments.PaymentNotificationContext{
			BookingID:       "booking-1",
			ReservationCode: "ABCD1234",
			CustomerName:    "Joao",
			AmountTotal:     1100,
			AmountPaid:      250,
			AmountDue:       850,
			PaymentStatus:   "PARTIAL",
		},
	}
	handler := NewHandler(NewService(store, chatSvc, config.Config{}, paymentSource))

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodPost, "/automation/jobs/payment-notifications/run", bytes.NewBufferString(`{"payment_id":"pay-1"}`))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if chatSvc.draftCalls != 0 {
		t.Fatalf("expected no draft queue calls, got %d", chatSvc.draftCalls)
	}
	if len(store.jobRuns) != 1 {
		t.Fatalf("expected one job run, got %d", len(store.jobRuns))
	}
	if store.jobRuns[0].Status != "SKIPPED" {
		t.Fatalf("expected SKIPPED job run, got %s", store.jobRuns[0].Status)
	}

	var out RunPaymentNotificationsResult
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	if out.Reason != "missing_customer_phone" {
		t.Fatalf("expected missing_customer_phone, got %s", out.Reason)
	}
}

func TestRunPaymentNotificationsDeduplicatesDraftByPayment(t *testing.T) {
	store := &fakeAutomationStore{}
	chatSvc := &fakeChatIngestor{
		draftResult: chat.QueueAutomationDraftResult{
			Session: chat.Session{
				ID:            "session-1",
				Channel:       "WHATSAPP",
				ContactKey:    "554998887766@s.whatsapp.net",
				CustomerPhone: "554998887766",
				Status:        "ACTIVE",
				HandoffStatus: "BOT",
				CreatedAt:     time.Now().UTC(),
				UpdatedAt:     time.Now().UTC(),
			},
			Message: chat.Message{
				ID:               "draft-1",
				SessionID:        "session-1",
				Direction:        "OUTBOUND",
				Kind:             "TEXT",
				ProcessingStatus: "AUTOMATION_DRAFT",
				ReceivedAt:       time.Now().UTC(),
				CreatedAt:        time.Now().UTC(),
			},
			Idempotent: true,
		},
	}
	paymentSource := &fakePaymentNotificationSource{
		payment: payments.Payment{
			ID:        "pay-1",
			BookingID: "booking-1",
			Amount:    250,
			Method:    "PIX",
			Status:    "PAID",
		},
		notification: payments.PaymentNotificationContext{
			BookingID:       "booking-1",
			ReservationCode: "ABCD1234",
			CustomerName:    "Joao",
			CustomerPhone:   "54998887766",
			AmountTotal:     1100,
			AmountPaid:      250,
			AmountDue:       850,
			PaymentStatus:   "PARTIAL",
		},
	}
	handler := NewHandler(NewService(store, chatSvc, config.Config{}, paymentSource))

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodPost, "/automation/jobs/payment-notifications/run", bytes.NewBufferString(`{"payment_id":"pay-1"}`))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if len(store.jobRuns) != 1 {
		t.Fatalf("expected one job run, got %d", len(store.jobRuns))
	}
	if store.jobRuns[0].Status != "SKIPPED" {
		t.Fatalf("expected SKIPPED job run, got %s", store.jobRuns[0].Status)
	}

	var out RunPaymentNotificationsResult
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	if out.Status != "skipped" {
		t.Fatalf("expected status skipped, got %s", out.Status)
	}
	if out.Reason != "draft_already_exists" {
		t.Fatalf("expected reason draft_already_exists, got %s", out.Reason)
	}
	if !out.Idempotent {
		t.Fatalf("expected idempotent true")
	}
}

func TestListJobRunsReturnsNormalizedJobRuns(t *testing.T) {
	store := &fakeAutomationStore{
		jobRuns: []JobRun{
			{
				ID:            uuid.NewString(),
				JobName:       "CHAT_REVIEW_ALERTS",
				TriggerSource: "MANUAL",
				Status:        "SENT",
				StartedAt:     time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC),
				CreatedAt:     time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC),
			},
			{
				ID:            uuid.NewString(),
				JobName:       "CHAT_REVIEW_ALERTS",
				TriggerSource: "SYSTEM",
				Status:        "FAILED",
				StartedAt:     time.Date(2026, 4, 22, 11, 0, 0, 0, time.UTC),
				CreatedAt:     time.Date(2026, 4, 22, 11, 0, 0, 0, time.UTC),
			},
			{
				ID:            uuid.NewString(),
				JobName:       "PAYMENT_NOTIFICATIONS",
				TriggerSource: "MANUAL",
				Status:        "SENT",
				StartedAt:     time.Date(2026, 4, 22, 10, 0, 0, 0, time.UTC),
				CreatedAt:     time.Date(2026, 4, 22, 10, 0, 0, 0, time.UTC),
			},
		},
	}
	handler := NewHandler(NewService(store, &fakeChatIngestor{}, config.Config{}))

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/automation/jobs/chat-review-alerts/runs?status=sent&limit=1", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var out ListJobRunsResult
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	if out.JobName != "CHAT_REVIEW_ALERTS" {
		t.Fatalf("expected job name CHAT_REVIEW_ALERTS, got %s", out.JobName)
	}
	if out.Filter.Status != "SENT" {
		t.Fatalf("expected filter status SENT, got %s", out.Filter.Status)
	}
	if out.Filter.Limit != 1 {
		t.Fatalf("expected filter limit 1, got %d", out.Filter.Limit)
	}
	if out.Count != 1 || len(out.Runs) != 1 {
		t.Fatalf("expected one run, got count=%d len=%d", out.Count, len(out.Runs))
	}
	if out.Runs[0].TriggerSource != "MANUAL" {
		t.Fatalf("expected MANUAL trigger source, got %s", out.Runs[0].TriggerSource)
	}
}

func TestListJobRunsFiltersByTriggerSource(t *testing.T) {
	store := &fakeAutomationStore{
		jobRuns: []JobRun{
			{
				ID:            uuid.NewString(),
				JobName:       "CHAT_REVIEW_ALERTS",
				TriggerSource: "MANUAL",
				Status:        "SENT",
				StartedAt:     time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC),
				CreatedAt:     time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC),
			},
			{
				ID:            uuid.NewString(),
				JobName:       "CHAT_REVIEW_ALERTS",
				TriggerSource: "SYSTEM",
				Status:        "SENT",
				StartedAt:     time.Date(2026, 4, 22, 11, 0, 0, 0, time.UTC),
				CreatedAt:     time.Date(2026, 4, 22, 11, 0, 0, 0, time.UTC),
			},
		},
	}
	handler := NewHandler(NewService(store, &fakeChatIngestor{}, config.Config{}))

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/automation/jobs/chat-review-alerts/runs?trigger_source=system", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var out ListJobRunsResult
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	if out.Count != 1 || len(out.Runs) != 1 {
		t.Fatalf("expected one filtered run, got count=%d len=%d", out.Count, len(out.Runs))
	}
	if out.Runs[0].TriggerSource != "SYSTEM" {
		t.Fatalf("expected SYSTEM trigger source, got %s", out.Runs[0].TriggerSource)
	}
}

func TestListJobRunsRejectsInvalidLimit(t *testing.T) {
	handler := NewHandler(NewService(&fakeAutomationStore{}, &fakeChatIngestor{}, config.Config{}))

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/automation/jobs/chat-review-alerts/runs?limit=nope", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestRunChatReviewAlertsSkipsWithoutActiveAlert(t *testing.T) {
	chatSvc := &fakeChatIngestor{
		summary: chat.SessionsSummary{
			TotalCount:         3,
			PendingReviewCount: 1,
			ReviewSLASeconds:   900,
		},
	}
	notifier := &fakeChatReviewAlertNotifier{enabled: true}
	handler := NewHandler(NewService(&fakeAutomationStore{}, chatSvc, config.Config{}, notifier))

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodPost, "/automation/jobs/chat-review-alerts/run", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if notifier.calls != 0 {
		t.Fatalf("expected notifier not to be called, got %d", notifier.calls)
	}

	var out RunChatReviewAlertsResult
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	if out.Status != "skipped" {
		t.Fatalf("expected status skipped, got %s", out.Status)
	}
	if out.Reason != "no_active_alert" {
		t.Fatalf("expected reason no_active_alert, got %s", out.Reason)
	}
}

func TestRunChatReviewAlertsSendsWebhookWhenAlertIsActive(t *testing.T) {
	store := &fakeAutomationStore{}
	chatSvc := &fakeChatIngestor{
		summary: chat.SessionsSummary{
			TotalCount:              5,
			PendingReviewCount:      2,
			OverdueReviewCount:      1,
			HighPriorityReviewCount: 1,
			OldestPendingAgeSeconds: 1200,
			ReviewSLASeconds:        900,
			HasReviewAlert:          true,
			ReviewAlertLevel:        "CRITICAL",
			ReviewAlertCode:         "REVIEW_QUEUE_OVERDUE",
			ReviewAlertMessage:      "Fila com drafts acima do SLA de revisao.",
			ReviewAlertSessionCount: 1,
		},
	}
	notifier := &fakeChatReviewAlertNotifier{enabled: true}
	handler := NewHandler(NewService(store, chatSvc, config.Config{}, notifier))

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodPost, "/automation/jobs/chat-review-alerts/run", bytes.NewBufferString(`{"channel":"whatsapp","handoff_status":"bot"}`))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if notifier.calls != 1 {
		t.Fatalf("expected notifier to be called once, got %d", notifier.calls)
	}
	if notifier.lastPayload.Filter.Channel != "WHATSAPP" {
		t.Fatalf("expected filter channel WHATSAPP, got %s", notifier.lastPayload.Filter.Channel)
	}
	if notifier.lastPayload.AlertCode != "REVIEW_QUEUE_OVERDUE" {
		t.Fatalf("expected alert code REVIEW_QUEUE_OVERDUE, got %s", notifier.lastPayload.AlertCode)
	}

	var out RunChatReviewAlertsResult
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	if out.Status != "sent" {
		t.Fatalf("expected status sent, got %s", out.Status)
	}
	if !out.NotificationTriggered {
		t.Fatalf("expected notification_triggered true")
	}
	if out.JobRunID == "" {
		t.Fatalf("expected job_run_id to be present")
	}
	if len(store.jobRuns) != 1 {
		t.Fatalf("expected one persisted job run, got %d", len(store.jobRuns))
	}
	if store.jobRuns[0].Status != "SENT" {
		t.Fatalf("expected persisted job run status SENT, got %s", store.jobRuns[0].Status)
	}
}

func TestRunChatReviewAlertsDeduplicatesSameAlertState(t *testing.T) {
	store := &fakeAutomationStore{}
	chatSvc := &fakeChatIngestor{
		summary: chat.SessionsSummary{
			TotalCount:              5,
			PendingReviewCount:      2,
			OverdueReviewCount:      1,
			HighPriorityReviewCount: 1,
			OldestPendingAgeSeconds: 1200,
			ReviewSLASeconds:        900,
			HasReviewAlert:          true,
			ReviewAlertLevel:        "CRITICAL",
			ReviewAlertCode:         "REVIEW_QUEUE_OVERDUE",
			ReviewAlertMessage:      "Fila com drafts acima do SLA de revisao.",
			ReviewAlertSessionCount: 1,
		},
	}
	notifier := &fakeChatReviewAlertNotifier{enabled: true}
	handler := NewHandler(NewService(store, chatSvc, config.Config{}, notifier))

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	for i := 0; i < 2; i++ {
		if i == 1 {
			chatSvc.summary.OldestPendingAgeSeconds = 1500
		}
		req := httptest.NewRequest(http.MethodPost, "/automation/jobs/chat-review-alerts/run", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d on run %d", http.StatusOK, rec.Code, i+1)
		}
		if i == 1 {
			var out RunChatReviewAlertsResult
			if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
				t.Fatalf("unmarshal result: %v", err)
			}
			if out.Reason != "duplicate_alert_state" {
				t.Fatalf("expected duplicate_alert_state, got %s", out.Reason)
			}
			if !out.Deduplicated {
				t.Fatalf("expected deduplicated true")
			}
		}
	}

	if notifier.calls != 1 {
		t.Fatalf("expected notifier to be called once, got %d", notifier.calls)
	}
	if len(store.jobRuns) != 1 {
		t.Fatalf("expected one persisted sent job run, got %d", len(store.jobRuns))
	}
}

func TestRunChatReviewAlertsSkipsWhenNotifierNotConfigured(t *testing.T) {
	chatSvc := &fakeChatIngestor{
		summary: chat.SessionsSummary{
			HasReviewAlert:          true,
			ReviewAlertLevel:        "WARNING",
			ReviewAlertCode:         "REVIEW_QUEUE_DUE_SOON",
			ReviewAlertSessionCount: 2,
		},
	}
	handler := NewHandler(NewService(&fakeAutomationStore{}, chatSvc, config.Config{}))

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodPost, "/automation/jobs/chat-review-alerts/run", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var out RunChatReviewAlertsResult
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	if out.Reason != "notifier_not_configured" {
		t.Fatalf("expected reason notifier_not_configured, got %s", out.Reason)
	}
	if out.WebhookConfigured {
		t.Fatalf("expected webhook_configured false")
	}
}

type fakeChatIngestor struct {
	calls             int
	lastInput         chat.IngestMessageInput
	presenceCalls     int
	lastPresenceInput chat.ApplyPresenceSignalInput
	summaryCalls      int
	lastSummaryFilter chat.ListSessionsFilter
	summary           chat.SessionsSummary
	draftCalls        int
	lastDraftInput    chat.QueueAutomationDraftInput
	draftResult       chat.QueueAutomationDraftResult
}

func (f *fakeChatIngestor) Ingest(_ context.Context, input chat.IngestMessageInput) (chat.IngestMessageResult, error) {
	f.calls++
	f.lastInput = input
	now := time.Now().UTC()
	return chat.IngestMessageResult{
		Session: chat.Session{
			ID:            uuid.NewString(),
			Channel:       input.Channel,
			ContactKey:    input.ContactKey,
			CustomerPhone: input.CustomerPhone,
			CustomerName:  input.CustomerName,
			Status:        "ACTIVE",
			HandoffStatus: "BOT",
			CreatedAt:     now,
			UpdatedAt:     now,
		},
		Message: chat.Message{
			ID:                uuid.NewString(),
			SessionID:         uuid.NewString(),
			Direction:         input.Message.Direction,
			Kind:              input.Message.Kind,
			ProviderMessageID: input.Message.ProviderMessageID,
			IdempotencyKey:    input.Message.IdempotencyKey,
			SenderName:        input.Message.SenderName,
			SenderPhone:       input.Message.SenderPhone,
			Body:              input.Message.Body,
			Payload:           input.Message.Payload,
			NormalizedPayload: input.Message.NormalizedPayload,
			ProcessingStatus:  input.Message.ProcessingStatus,
			ReceivedAt:        now,
			CreatedAt:         now,
		},
	}, nil
}

func (f *fakeChatIngestor) ApplyPresenceSignal(_ context.Context, input chat.ApplyPresenceSignalInput) (chat.ApplyPresenceSignalResult, error) {
	f.presenceCalls++
	f.lastPresenceInput = input
	now := time.Now().UTC()
	return chat.ApplyPresenceSignalResult{
		Status:        "accepted",
		PresenceState: input.PresenceStatus,
		Reason:        "presence_extended",
		Session: chat.Session{
			ID:            uuid.NewString(),
			Channel:       input.Channel,
			ContactKey:    input.ContactKey,
			CustomerPhone: normalizePhone(input.ContactKey),
			Status:        "ACTIVE",
			HandoffStatus: "BOT",
			CreatedAt:     now,
			UpdatedAt:     now,
		},
	}, nil
}

func (f *fakeChatIngestor) GetSessionsSummary(_ context.Context, filter chat.ListSessionsFilter) (chat.SessionsSummary, error) {
	f.summaryCalls++
	f.lastSummaryFilter = filter
	return f.summary, nil
}

func (f *fakeChatIngestor) QueueAutomationDraft(_ context.Context, input chat.QueueAutomationDraftInput) (chat.QueueAutomationDraftResult, error) {
	f.draftCalls++
	f.lastDraftInput = input
	if f.draftResult.Session.ID == "" {
		now := time.Now().UTC()
		f.draftResult = chat.QueueAutomationDraftResult{
			Session: chat.Session{
				ID:            uuid.NewString(),
				Channel:       input.Channel,
				ContactKey:    input.ContactKey,
				CustomerPhone: input.CustomerPhone,
				CustomerName:  input.CustomerName,
				Status:        "ACTIVE",
				HandoffStatus: "BOT",
				CreatedAt:     now,
				UpdatedAt:     now,
			},
			Message: chat.Message{
				ID:                uuid.NewString(),
				SessionID:         uuid.NewString(),
				Direction:         "OUTBOUND",
				Kind:              "TEXT",
				IdempotencyKey:    input.IdempotencyKey,
				SenderName:        input.SenderName,
				Body:              input.Body,
				Payload:           input.Metadata,
				NormalizedPayload: input.Metadata,
				ProcessingStatus:  "AUTOMATION_DRAFT",
				ReceivedAt:        now,
				CreatedAt:         now,
			},
		}
	}
	return f.draftResult, nil
}

type fakePaymentNotificationSource struct {
	payment      payments.Payment
	list         []payments.Payment
	notification payments.PaymentNotificationContext
	getErr       error
	listErr      error
	notifyCtxErr error
}

func (f *fakePaymentNotificationSource) Get(_ context.Context, _ string) (payments.Payment, error) {
	if f.getErr != nil {
		return payments.Payment{}, f.getErr
	}
	return f.payment, nil
}

func (f *fakePaymentNotificationSource) List(_ context.Context, _ payments.PaymentListFilter) ([]payments.Payment, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.list, nil
}

func (f *fakePaymentNotificationSource) GetBookingNotificationContext(_ context.Context, _ string) (payments.PaymentNotificationContext, error) {
	if f.notifyCtxErr != nil {
		return payments.PaymentNotificationContext{}, f.notifyCtxErr
	}
	return f.notification, nil
}

type fakeBookingExpirationSource struct {
	items       []bookings.BookingListItem
	listErr     error
	updateErr   error
	listCalls   int
	lastFilter  bookings.ListFilter
	updateCalls []fakeBookingUpdateCall
}

type fakeBookingUpdateCall struct {
	ID     string
	Status string
}

func (f *fakeBookingExpirationSource) List(_ context.Context, filter bookings.ListFilter) ([]bookings.BookingListItem, error) {
	f.listCalls++
	f.lastFilter = filter
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.items, nil
}

func (f *fakeBookingExpirationSource) UpdateStatus(_ context.Context, id string, status string) (bookings.BookingDetails, error) {
	f.updateCalls = append(f.updateCalls, fakeBookingUpdateCall{ID: id, Status: status})
	if f.updateErr != nil {
		return bookings.BookingDetails{}, f.updateErr
	}

	var matched bookings.BookingListItem
	for _, item := range f.items {
		if item.ID == id {
			matched = item
			break
		}
	}

	return bookings.BookingDetails{
		Booking: bookings.Booking{
			ID:              id,
			TripID:          matched.TripID,
			Status:          status,
			ReservationCode: matched.ReservationCode,
			ExpiresAt:       matched.ExpiresAt,
			CreatedAt:       matched.CreatedAt,
		},
	}, nil
}

func TestNormalizePhone(t *testing.T) {
	if got := normalizePhone("5549988709047@s.whatsapp.net"); got != "5549988709047" {
		t.Fatalf("unexpected normalized phone: %s", got)
	}
}

func TestHandleEvolutionMessagesResponseBody(t *testing.T) {
	chatSvc := &fakeChatIngestor{}
	handler := NewHandler(NewService(&fakeAutomationStore{}, chatSvc, config.Config{}))

	r := chi.NewRouter()
	handler.RegisterWebhooks(r)

	req := httptest.NewRequest(http.MethodPost, "/webhooks/evolution/messages", bytes.NewBufferString(`{
		"event":"messages.upsert",
		"data":{
			"key":{"remoteJid":"554998208115@s.whatsapp.net","fromMe":false,"id":"MSG-4"},
			"message":{"imageMessage":{"caption":"foto do documento"}},
			"messageType":"imageMessage"
		}
	}`))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	var out EvolutionWebhookResult
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if out.Status != "accepted" {
		t.Fatalf("expected accepted status, got %s", out.Status)
	}
	if out.MessageType != "imageMessage" {
		t.Fatalf("expected imageMessage type, got %s", out.MessageType)
	}
}

func TestHandleEvolutionStatusAcceptsDirectPayload(t *testing.T) {
	store := &fakeAutomationStore{
		result: RecordEvolutionStatusResult{MatchedChatMessages: 1, MatchedOutboundMessages: 2},
	}
	handler := NewHandler(NewService(store, &fakeChatIngestor{}, config.Config{}))

	r := chi.NewRouter()
	handler.RegisterWebhooks(r)

	req := httptest.NewRequest(http.MethodPost, "/webhooks/evolution/status", bytes.NewBufferString(`{
		"event":"messages.update",
		"data":{
			"key":{"id":"MSG-STATUS-1"},
			"status":"READ",
			"messageType":"conversation",
			"messageTimestamp":1772544357
		}
	}`))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d", http.StatusAccepted, rec.Code)
	}
	if store.lastStatus.ProviderMessageID != "MSG-STATUS-1" {
		t.Fatalf("unexpected provider message id: %s", store.lastStatus.ProviderMessageID)
	}
	if store.lastStatus.ProviderStatus != "READ" {
		t.Fatalf("unexpected provider status: %s", store.lastStatus.ProviderStatus)
	}
}

func TestHandleEvolutionStatusAcceptsN8NEnvelope(t *testing.T) {
	store := &fakeAutomationStore{
		result: RecordEvolutionStatusResult{MatchedChatMessages: 0, MatchedOutboundMessages: 1},
	}
	handler := NewHandler(NewService(store, &fakeChatIngestor{}, config.Config{}))

	r := chi.NewRouter()
	handler.RegisterWebhooks(r)

	req := httptest.NewRequest(http.MethodPost, "/webhooks/evolution/status", bytes.NewBufferString(`{
		"body":{
			"event":"messages.update",
			"data":{
				"key":{"id":"MSG-STATUS-2"},
				"status":"DELIVERY_ACK",
				"messageTimestamp":1772544357
			}
		}
	}`))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d", http.StatusAccepted, rec.Code)
	}
	if store.lastStatus.ProviderStatus != "DELIVERY_ACK" {
		t.Fatalf("unexpected provider status: %s", store.lastStatus.ProviderStatus)
	}
}

func TestHandleEvolutionStatusRejectsMissingStatus(t *testing.T) {
	handler := NewHandler(NewService(&fakeAutomationStore{}, &fakeChatIngestor{}, config.Config{}))

	r := chi.NewRouter()
	handler.RegisterWebhooks(r)

	req := httptest.NewRequest(http.MethodPost, "/webhooks/evolution/status", bytes.NewBufferString(`{
		"event":"messages.update",
		"data":{"key":{"id":"MSG-STATUS-3"}}
	}`))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestHandleEvolutionPresenceAcceptsDirectPayload(t *testing.T) {
	chatSvc := &fakeChatIngestor{}
	handler := NewHandler(NewService(&fakeAutomationStore{}, chatSvc, config.Config{}))

	r := chi.NewRouter()
	handler.RegisterWebhooks(r)

	req := httptest.NewRequest(http.MethodPost, "/webhooks/evolution/presence", bytes.NewBufferString(`{
		"event":"presence.update",
		"instance":"belle",
		"data":{
			"id":"554998208115@s.whatsapp.net",
			"presences":{
				"554998208115@s.whatsapp.net":{"lastKnownPresence":"typing"}
			}
		},
		"date_time":"2026-04-21T14:00:00Z"
	}`))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d", http.StatusAccepted, rec.Code)
	}
	if chatSvc.presenceCalls != 1 {
		t.Fatalf("expected one presence call, got %d", chatSvc.presenceCalls)
	}
	if chatSvc.lastPresenceInput.ContactKey != "554998208115@s.whatsapp.net" {
		t.Fatalf("unexpected contact key: %s", chatSvc.lastPresenceInput.ContactKey)
	}
	if chatSvc.lastPresenceInput.PresenceStatus != "TYPING" {
		t.Fatalf("unexpected presence status: %s", chatSvc.lastPresenceInput.PresenceStatus)
	}
}

func TestHandleEvolutionPresenceAcceptsN8NEnvelope(t *testing.T) {
	chatSvc := &fakeChatIngestor{}
	handler := NewHandler(NewService(&fakeAutomationStore{}, chatSvc, config.Config{}))

	r := chi.NewRouter()
	handler.RegisterWebhooks(r)

	req := httptest.NewRequest(http.MethodPost, "/webhooks/evolution/presence", bytes.NewBufferString(`{
		"body":{
			"event":"presence.update",
			"data":{
				"id":"554998208115@s.whatsapp.net",
				"presences":{
					"554998208115@s.whatsapp.net":{"lastKnownPresence":"paused"}
				}
			}
		}
	}`))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d", http.StatusAccepted, rec.Code)
	}
	if chatSvc.lastPresenceInput.PresenceStatus != "PAUSED" {
		t.Fatalf("unexpected presence status: %s", chatSvc.lastPresenceInput.PresenceStatus)
	}
}

func TestHandleEvolutionPresenceRejectsMissingPresence(t *testing.T) {
	handler := NewHandler(NewService(&fakeAutomationStore{}, &fakeChatIngestor{}, config.Config{}))

	r := chi.NewRouter()
	handler.RegisterWebhooks(r)

	req := httptest.NewRequest(http.MethodPost, "/webhooks/evolution/presence", bytes.NewBufferString(`{
		"event":"presence.update",
		"data":{"id":"554998208115@s.whatsapp.net","presences":{}}
	}`))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

type fakeAutomationStore struct {
	lastStatus RecordEvolutionStatusInput
	result     RecordEvolutionStatusResult
	jobRuns    []JobRun
}

func (f *fakeAutomationStore) RecordEvolutionStatus(_ context.Context, input RecordEvolutionStatusInput) (RecordEvolutionStatusResult, error) {
	f.lastStatus = input
	return f.result, nil
}

func (f *fakeAutomationStore) CreateJobRun(_ context.Context, input CreateJobRunInput) (JobRun, error) {
	finishedAt := input.FinishedAt
	item := JobRun{
		ID:            uuid.NewString(),
		JobName:       input.JobName,
		TriggerSource: input.TriggerSource,
		Status:        input.Status,
		InputPayload:  input.InputPayload,
		ResultPayload: input.ResultPayload,
		ErrorText:     input.ErrorText,
		StartedAt:     input.StartedAt,
		FinishedAt:    finishedAt,
		CreatedAt:     time.Now().UTC(),
	}
	f.jobRuns = append(f.jobRuns, item)
	return item, nil
}

func (f *fakeAutomationStore) UpdateJobRun(_ context.Context, input UpdateJobRunInput) (JobRun, error) {
	for i := range f.jobRuns {
		if f.jobRuns[i].ID != input.ID {
			continue
		}
		f.jobRuns[i].Status = input.Status
		f.jobRuns[i].ResultPayload = input.ResultPayload
		f.jobRuns[i].ErrorText = input.ErrorText
		finishedAt := input.FinishedAt
		f.jobRuns[i].FinishedAt = &finishedAt
		return f.jobRuns[i], nil
	}
	return JobRun{}, nil
}

func (f *fakeAutomationStore) FindLatestJobRunByKey(_ context.Context, jobName string, alertSignature string) (*JobRun, error) {
	for i := len(f.jobRuns) - 1; i >= 0; i-- {
		item := f.jobRuns[i]
		if item.JobName != jobName || item.Status != "SENT" {
			continue
		}
		if asString(item.ResultPayload["alert_signature"]) != alertSignature {
			continue
		}
		copy := item
		return &copy, nil
	}
	return nil, nil
}

func (f *fakeAutomationStore) ListJobRuns(_ context.Context, input ListJobRunsInput) ([]JobRun, error) {
	items := make([]JobRun, 0, len(f.jobRuns))
	for _, item := range f.jobRuns {
		if item.JobName != input.JobName {
			continue
		}
		if input.Status != "" && item.Status != input.Status {
			continue
		}
		if input.TriggerSource != "" && item.TriggerSource != input.TriggerSource {
			continue
		}
		items = append(items, item)
		if input.Limit > 0 && len(items) >= input.Limit {
			break
		}
	}
	return items, nil
}

type fakeChatReviewAlertNotifier struct {
	enabled     bool
	calls       int
	lastPayload ChatReviewAlertNotificationPayload
}

func (f *fakeChatReviewAlertNotifier) Enabled() bool {
	return f != nil && f.enabled
}

func (f *fakeChatReviewAlertNotifier) NotifyReviewAlert(_ context.Context, payload ChatReviewAlertNotificationPayload) error {
	f.calls++
	f.lastPayload = payload
	return nil
}
