package chat

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"schumacher-tur/api/internal/shared/config"
)

func TestIngestMessageCreatesSessionAndMessage(t *testing.T) {
	store := newFakeStore()
	handler := NewHandler(NewService(store, config.Config{ChatDebounceWindowMS: 1500}))

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	body := bytes.NewBufferString(`{
		"channel":"whatsapp",
		"contact_key":"5511999999999",
		"customer_phone":"5511999999999",
		"customer_name":"Cliente Teste",
		"metadata":{"source":"shadow"},
		"message":{
			"direction":"inbound",
			"provider_message_id":"msg-1",
			"idempotency_key":"idem-1",
			"body":"oi",
			"payload":{"raw":"ok"}
		}
	}`)

	req := httptest.NewRequest(http.MethodPost, "/chat/messages/ingest", body)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}

	var out IngestMessageResult
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if out.Idempotent {
		t.Fatalf("expected new message, got idempotent response")
	}
	if out.Session.ContactKey != "5511999999999" {
		t.Fatalf("expected contact key to be persisted")
	}
	if out.Message.ProviderMessageID != "msg-1" {
		t.Fatalf("expected provider message id to be persisted")
	}
	buffer, ok := out.Session.Metadata["buffer"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected buffer metadata to be present")
	}
	if got := asString(buffer["status"]); got != bufferStatusPending {
		t.Fatalf("expected buffer status %s, got %s", bufferStatusPending, got)
	}
	if got := asInt(buffer["message_count"]); got != 1 {
		t.Fatalf("expected one buffered message, got %d", got)
	}
	if out.Message.ProcessingStatus != "BUFFERED_PENDING" {
		t.Fatalf("expected buffered processing status, got %s", out.Message.ProcessingStatus)
	}
	if len(store.sessions) != 1 || len(store.messages) != 1 {
		t.Fatalf("expected one session and one message, got %d sessions and %d messages", len(store.sessions), len(store.messages))
	}
}

func TestIngestMessageReturnsExistingOnIdempotency(t *testing.T) {
	store := newFakeStore()
	handler := NewHandler(NewService(store, config.Config{ChatDebounceWindowMS: 1500}))

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	firstReq := httptest.NewRequest(http.MethodPost, "/chat/messages/ingest", bytes.NewBufferString(`{
		"contact_key":"5511999999999",
		"message":{"direction":"INBOUND","provider_message_id":"msg-dup","idempotency_key":"idem-dup","body":"oi"}
	}`))
	firstRec := httptest.NewRecorder()
	r.ServeHTTP(firstRec, firstReq)
	if firstRec.Code != http.StatusCreated {
		t.Fatalf("expected first call to create message, got %d", firstRec.Code)
	}

	secondReq := httptest.NewRequest(http.MethodPost, "/chat/messages/ingest", bytes.NewBufferString(`{
		"contact_key":"5511999999999",
		"message":{"direction":"INBOUND","provider_message_id":"msg-dup","idempotency_key":"idem-dup","body":"oi de novo"}
	}`))
	secondRec := httptest.NewRecorder()
	r.ServeHTTP(secondRec, secondReq)

	if secondRec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, secondRec.Code)
	}

	var out IngestMessageResult
	if err := json.Unmarshal(secondRec.Body.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if !out.Idempotent {
		t.Fatalf("expected idempotent response")
	}
	if len(store.messages) != 1 {
		t.Fatalf("expected only one persisted message, got %d", len(store.messages))
	}
}

func TestListAndGetSessionAndMessages(t *testing.T) {
	store := newFakeStore()
	session, message := store.seedSessionWithMessage("5511888888888", "ola")
	handler := NewHandler(NewService(store, config.Config{ChatDebounceWindowMS: 1500}))

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	listReq := httptest.NewRequest(http.MethodGet, "/chat/sessions?channel=whatsapp&limit=10", nil)
	listRec := httptest.NewRecorder()
	r.ServeHTTP(listRec, listReq)

	if listRec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, listRec.Code)
	}

	var sessions []Session
	if err := json.Unmarshal(listRec.Body.Bytes(), &sessions); err != nil {
		t.Fatalf("unmarshal sessions: %v", err)
	}
	if len(sessions) != 1 || sessions[0].ID != session.ID {
		t.Fatalf("expected listed session %s", session.ID)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/chat/sessions/"+session.ID, nil)
	getRec := httptest.NewRecorder()
	r.ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, getRec.Code)
	}

	msgReq := httptest.NewRequest(http.MethodGet, "/chat/sessions/"+session.ID+"/messages?limit=5", nil)
	msgRec := httptest.NewRecorder()
	r.ServeHTTP(msgRec, msgReq)
	if msgRec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, msgRec.Code)
	}

	var messages []Message
	if err := json.Unmarshal(msgRec.Body.Bytes(), &messages); err != nil {
		t.Fatalf("unmarshal messages: %v", err)
	}
	if len(messages) != 1 || messages[0].ID != message.ID {
		t.Fatalf("expected listed message %s", message.ID)
	}
}

func TestGetCurrentDraftReturnsNotFoundWhenSessionHasNoDraft(t *testing.T) {
	store := newFakeStore()
	session, _ := store.seedSessionWithMessage("5511999999999", "oi")
	handler := NewHandler(NewService(store, config.Config{ChatDebounceWindowMS: 1500}))

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/chat/sessions/"+session.ID+"/draft", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestGetCurrentDraftReturnsObservabilityForGeneratedDraft(t *testing.T) {
	store := newFakeStore()
	runner := &fakeAgentRunner{
		enabled: true,
		result: RunAgentResult{
			ReplyText:          "Encontrei uma opcao para esse trecho.",
			Model:              "gpt-test",
			ProviderResponseID: "resp-draft-view-1",
		},
	}
	searcher := &fakeAvailabilitySearcher{
		enabled: true,
		result: AvailabilitySearchResult{
			Results: []AvailabilitySearchItem{
				{
					SegmentID:              "seg-1",
					TripID:                 "trip-1",
					RouteID:                "route-1",
					BoardStopID:            "board-1",
					AlightStopID:           "alight-1",
					OriginStopID:           "stop-origin-1",
					DestinationStopID:      "stop-destination-1",
					OriginDisplayName:      "Videira/SC",
					DestinationDisplayName: "Sao Luis/MA",
					OriginDepartTime:       "18:30",
					TripDate:               "2026-05-10",
					SeatsAvailable:         12,
					Price:                  250,
					Currency:               "BRL",
					Status:                 "ACTIVE",
					TripStatus:             "SCHEDULED",
				},
			},
		},
	}
	svc := NewService(store, config.Config{ChatDebounceWindowMS: 1500}, runner, searcher)
	ingested, err := svc.Ingest(context.Background(), IngestMessageInput{
		ContactKey: "5511999999999",
		Message: IngestMessagePayload{
			Direction:         "INBOUND",
			ProviderMessageID: "msg-draft-view-1",
			IdempotencyKey:    "idem-draft-view-1",
			Body:              "qual o valor de Videira/SC para Sao Luis/MA em 10/05?",
		},
	})
	if err != nil {
		t.Fatalf("ingest message: %v", err)
	}
	reprocessed, err := svc.Reprocess(context.Background(), ReprocessInput{SessionID: ingested.Session.ID})
	if err != nil {
		t.Fatalf("reprocess message: %v", err)
	}
	if reprocessed.Draft == nil {
		t.Fatalf("expected draft to be generated")
	}

	handler := NewHandler(svc)
	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/chat/sessions/"+ingested.Session.ID+"/draft", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var out CurrentDraftResult
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal current draft: %v", err)
	}
	if out.Draft.ID != reprocessed.Draft.ID {
		t.Fatalf("expected draft id %s, got %s", reprocessed.Draft.ID, out.Draft.ID)
	}
	if out.DraftStatus != messageStatusAutomationDraft {
		t.Fatalf("expected draft status %s, got %s", messageStatusAutomationDraft, out.DraftStatus)
	}
	if out.AgentStatus != agentStatusDraftGenerated {
		t.Fatalf("expected agent status %s, got %s", agentStatusDraftGenerated, out.AgentStatus)
	}
	if out.Model != "gpt-test" {
		t.Fatalf("expected model gpt-test, got %s", out.Model)
	}
	if out.ProviderResponseID != "resp-draft-view-1" {
		t.Fatalf("expected provider response id resp-draft-view-1, got %s", out.ProviderResponseID)
	}
	if len(out.CurrentTurnMessageIDs) != 1 || out.CurrentTurnMessageIDs[0] == "" {
		t.Fatalf("expected current_turn_message_ids, got %+v", out.CurrentTurnMessageIDs)
	}
	if len(out.ToolNames) != 1 || out.ToolNames[0] != toolNameAvailabilitySearch {
		t.Fatalf("expected tool name %s, got %+v", toolNameAvailabilitySearch, out.ToolNames)
	}
	if out.ToolCallCount != 1 {
		t.Fatalf("expected tool call count 1, got %d", out.ToolCallCount)
	}
	if out.GeneratedAt == nil {
		t.Fatalf("expected generated_at to be present")
	}
	if out.LinkedReply != nil {
		t.Fatalf("expected no linked reply for generated draft")
	}
}

func TestGetCurrentDraftReturnsLinkedReplyForReviewedDraft(t *testing.T) {
	store := newFakeStore()
	ownerUserID := uuid.NewString()
	runner := &fakeAgentRunner{
		enabled: true,
		result: RunAgentResult{
			ReplyText:          "Posso seguir com sua reserva.",
			Model:              "gpt-test",
			ProviderResponseID: "resp-draft-view-2",
		},
	}
	svc := NewService(store, config.Config{ChatDebounceWindowMS: 1500}, runner)
	ingested, err := svc.Ingest(context.Background(), IngestMessageInput{
		ContactKey: "5511888888888",
		Message: IngestMessagePayload{
			Direction:         "INBOUND",
			ProviderMessageID: "msg-draft-view-2",
			IdempotencyKey:    "idem-draft-view-2",
			Body:              "quero continuar minha reserva",
		},
	})
	if err != nil {
		t.Fatalf("ingest message: %v", err)
	}
	reprocessed, err := svc.Reprocess(context.Background(), ReprocessInput{SessionID: ingested.Session.ID})
	if err != nil {
		t.Fatalf("reprocess message: %v", err)
	}
	if reprocessed.Draft == nil {
		t.Fatalf("expected draft to be generated")
	}
	if _, err := svc.RequestHandoff(context.Background(), RequestHandoffInput{
		SessionID:      ingested.Session.ID,
		RequestedBy:    "dashboard",
		AssignedUserID: ownerUserID,
	}); err != nil {
		t.Fatalf("request handoff: %v", err)
	}
	replied, err := svc.Reply(context.Background(), ReplyInput{
		SessionID:      ingested.Session.ID,
		OwnerUserID:    ownerUserID,
		DraftMessageID: reprocessed.Draft.ID,
		IdempotencyKey: "idem-reviewed-draft-1",
	})
	if err != nil {
		t.Fatalf("reply message: %v", err)
	}

	handler := NewHandler(svc)
	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/chat/sessions/"+ingested.Session.ID+"/draft", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var out CurrentDraftResult
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal current draft: %v", err)
	}
	if out.Draft.ID != reprocessed.Draft.ID {
		t.Fatalf("expected reviewed draft id %s, got %s", reprocessed.Draft.ID, out.Draft.ID)
	}
	if out.DraftStatus != messageStatusAutomationReviewed {
		t.Fatalf("expected draft status %s, got %s", messageStatusAutomationReviewed, out.DraftStatus)
	}
	if out.AgentStatus != agentStatusDraftReviewed {
		t.Fatalf("expected agent status %s, got %s", agentStatusDraftReviewed, out.AgentStatus)
	}
	if out.ReviewMode != "CONTROLLED" {
		t.Fatalf("expected review mode CONTROLLED, got %s", out.ReviewMode)
	}
	if out.ReviewAction != "APPROVED_AS_IS" {
		t.Fatalf("expected review action APPROVED_AS_IS, got %s", out.ReviewAction)
	}
	if out.ReviewedByUserID != ownerUserID {
		t.Fatalf("expected reviewed_by_user_id %s, got %s", ownerUserID, out.ReviewedByUserID)
	}
	if out.ReviewedAt == nil {
		t.Fatalf("expected reviewed_at to be present")
	}
	if out.LinkedReply == nil {
		t.Fatalf("expected linked reply to be present")
	}
	if out.LinkedReply.ID != replied.Message.ID {
		t.Fatalf("expected linked reply id %s, got %s", replied.Message.ID, out.LinkedReply.ID)
	}
}

func TestListSessionsIncludesDraftReviewSummary(t *testing.T) {
	store := newFakeStore()
	runner := &fakeAgentRunner{
		enabled: true,
		result: RunAgentResult{
			ReplyText:          "Encontrei uma opcao para esse trecho.",
			Model:              "gpt-test",
			ProviderResponseID: "resp-session-summary-1",
		},
	}
	searcher := &fakeAvailabilitySearcher{
		enabled: true,
		result: AvailabilitySearchResult{
			Results: []AvailabilitySearchItem{
				{
					SegmentID:              "seg-1",
					TripID:                 "trip-1",
					RouteID:                "route-1",
					BoardStopID:            "board-1",
					AlightStopID:           "alight-1",
					OriginStopID:           "stop-origin-1",
					DestinationStopID:      "stop-destination-1",
					OriginDisplayName:      "Videira/SC",
					DestinationDisplayName: "Sao Luis/MA",
					OriginDepartTime:       "18:30",
					TripDate:               "2026-05-10",
					SeatsAvailable:         12,
					Price:                  250,
					Currency:               "BRL",
					Status:                 "ACTIVE",
					TripStatus:             "SCHEDULED",
				},
			},
		},
	}
	svc := NewService(store, config.Config{ChatDebounceWindowMS: 1500}, runner, searcher)
	ingested, err := svc.Ingest(context.Background(), IngestMessageInput{
		ContactKey: "5511777777777",
		Message: IngestMessagePayload{
			Direction:         "INBOUND",
			ProviderMessageID: "msg-session-summary-1",
			IdempotencyKey:    "idem-session-summary-1",
			Body:              "qual o valor de Videira/SC para Sao Luis/MA em 10/05?",
		},
	})
	if err != nil {
		t.Fatalf("ingest message: %v", err)
	}
	if _, err := svc.Reprocess(context.Background(), ReprocessInput{SessionID: ingested.Session.ID}); err != nil {
		t.Fatalf("reprocess message: %v", err)
	}

	handler := NewHandler(svc)
	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/chat/sessions?channel=whatsapp&limit=10", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var sessions []Session
	if err := json.Unmarshal(rec.Body.Bytes(), &sessions); err != nil {
		t.Fatalf("unmarshal sessions: %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("expected one session, got %d", len(sessions))
	}
	if sessions[0].AgentStatus != agentStatusDraftGenerated {
		t.Fatalf("expected agent status %s, got %s", agentStatusDraftGenerated, sessions[0].AgentStatus)
	}
	if !sessions[0].HasAutomationDraft {
		t.Fatalf("expected has_automation_draft to be true")
	}
	if sessions[0].DraftReviewStatus != "PENDING_REVIEW" {
		t.Fatalf("expected draft review status PENDING_REVIEW, got %s", sessions[0].DraftReviewStatus)
	}
	if sessions[0].DraftGeneratedAt == nil {
		t.Fatalf("expected draft_generated_at to be present")
	}
	if sessions[0].DraftReviewSLASeconds != 15*60 {
		t.Fatalf("expected draft review SLA 900 seconds, got %d", sessions[0].DraftReviewSLASeconds)
	}
	if sessions[0].DraftPendingAgeBucket != "FRESH" {
		t.Fatalf("expected draft pending age bucket FRESH, got %s", sessions[0].DraftPendingAgeBucket)
	}
	if sessions[0].DraftReviewPriority != "LOW" {
		t.Fatalf("expected draft review priority LOW, got %s", sessions[0].DraftReviewPriority)
	}
	if sessions[0].DraftReviewAlertActive {
		t.Fatalf("expected draft review alert inactive for fresh draft")
	}
	if sessions[0].DraftReviewOverdue {
		t.Fatalf("expected draft review overdue false")
	}
	if len(sessions[0].DraftToolNames) != 1 || sessions[0].DraftToolNames[0] != toolNameAvailabilitySearch {
		t.Fatalf("expected draft tool names to include %s, got %+v", toolNameAvailabilitySearch, sessions[0].DraftToolNames)
	}
	if sessions[0].DraftToolCallCount != 1 {
		t.Fatalf("expected draft tool call count 1, got %d", sessions[0].DraftToolCallCount)
	}
	if sessions[0].DraftModel != "gpt-test" {
		t.Fatalf("expected draft model gpt-test, got %s", sessions[0].DraftModel)
	}
	if sessions[0].DraftProviderResponseID != "resp-session-summary-1" {
		t.Fatalf("expected provider response id resp-session-summary-1, got %s", sessions[0].DraftProviderResponseID)
	}
}

func TestGetSessionIncludesReviewedDraftSummary(t *testing.T) {
	store := newFakeStore()
	ownerUserID := uuid.NewString()
	runner := &fakeAgentRunner{
		enabled: true,
		result: RunAgentResult{
			ReplyText:          "Posso seguir com sua reserva.",
			Model:              "gpt-test",
			ProviderResponseID: "resp-session-summary-2",
		},
	}
	svc := NewService(store, config.Config{ChatDebounceWindowMS: 1500}, runner)
	ingested, err := svc.Ingest(context.Background(), IngestMessageInput{
		ContactKey: "5511666666666",
		Message: IngestMessagePayload{
			Direction:         "INBOUND",
			ProviderMessageID: "msg-session-summary-2",
			IdempotencyKey:    "idem-session-summary-2",
			Body:              "quero continuar minha reserva",
		},
	})
	if err != nil {
		t.Fatalf("ingest message: %v", err)
	}
	reprocessed, err := svc.Reprocess(context.Background(), ReprocessInput{SessionID: ingested.Session.ID})
	if err != nil {
		t.Fatalf("reprocess message: %v", err)
	}
	if _, err := svc.RequestHandoff(context.Background(), RequestHandoffInput{
		SessionID:      ingested.Session.ID,
		RequestedBy:    "dashboard",
		AssignedUserID: ownerUserID,
	}); err != nil {
		t.Fatalf("request handoff: %v", err)
	}
	if _, err := svc.Reply(context.Background(), ReplyInput{
		SessionID:      ingested.Session.ID,
		OwnerUserID:    ownerUserID,
		DraftMessageID: reprocessed.Draft.ID,
		IdempotencyKey: "idem-session-reviewed-1",
	}); err != nil {
		t.Fatalf("reply message: %v", err)
	}

	handler := NewHandler(svc)
	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/chat/sessions/"+ingested.Session.ID, nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var out Session
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal session: %v", err)
	}
	if out.AgentStatus != agentStatusDraftReviewed {
		t.Fatalf("expected agent status %s, got %s", agentStatusDraftReviewed, out.AgentStatus)
	}
	if !out.HasAutomationDraft {
		t.Fatalf("expected has_automation_draft to be true")
	}
	if out.DraftReviewStatus != "REVIEWED" {
		t.Fatalf("expected draft review status REVIEWED, got %s", out.DraftReviewStatus)
	}
	if out.DraftReviewedAt == nil {
		t.Fatalf("expected draft_reviewed_at to be present")
	}
	if out.DraftReviewedByUserID != ownerUserID {
		t.Fatalf("expected draft_reviewed_by_user_id %s, got %s", ownerUserID, out.DraftReviewedByUserID)
	}
	if out.DraftReviewAction != "APPROVED_AS_IS" {
		t.Fatalf("expected draft review action APPROVED_AS_IS, got %s", out.DraftReviewAction)
	}
}

func TestListSessionsFiltersByDraftReviewStatus(t *testing.T) {
	store := newFakeStore()
	ownerUserID := uuid.NewString()
	runner := &fakeAgentRunner{
		enabled: true,
		result: RunAgentResult{
			ReplyText:          "Posso seguir com sua reserva.",
			Model:              "gpt-test",
			ProviderResponseID: "resp-filter-review",
		},
	}
	svc := NewService(store, config.Config{ChatDebounceWindowMS: 1500}, runner)

	pending, err := svc.Ingest(context.Background(), IngestMessageInput{
		ContactKey: "5511555555555",
		Message: IngestMessagePayload{
			Direction:         "INBOUND",
			ProviderMessageID: "msg-filter-review-1",
			IdempotencyKey:    "idem-filter-review-1",
			Body:              "quero continuar minha reserva",
		},
	})
	if err != nil {
		t.Fatalf("ingest pending session: %v", err)
	}
	if _, err := svc.Reprocess(context.Background(), ReprocessInput{SessionID: pending.Session.ID}); err != nil {
		t.Fatalf("reprocess pending session: %v", err)
	}

	reviewed, err := svc.Ingest(context.Background(), IngestMessageInput{
		ContactKey: "5511444444444",
		Message: IngestMessagePayload{
			Direction:         "INBOUND",
			ProviderMessageID: "msg-filter-review-2",
			IdempotencyKey:    "idem-filter-review-2",
			Body:              "quero continuar minha reserva tambem",
		},
	})
	if err != nil {
		t.Fatalf("ingest reviewed session: %v", err)
	}
	reprocessed, err := svc.Reprocess(context.Background(), ReprocessInput{SessionID: reviewed.Session.ID})
	if err != nil {
		t.Fatalf("reprocess reviewed session: %v", err)
	}
	if _, err := svc.RequestHandoff(context.Background(), RequestHandoffInput{
		SessionID:      reviewed.Session.ID,
		RequestedBy:    "dashboard",
		AssignedUserID: ownerUserID,
	}); err != nil {
		t.Fatalf("request handoff: %v", err)
	}
	if _, err := svc.Reply(context.Background(), ReplyInput{
		SessionID:      reviewed.Session.ID,
		OwnerUserID:    ownerUserID,
		DraftMessageID: reprocessed.Draft.ID,
		IdempotencyKey: "idem-filter-review-3",
	}); err != nil {
		t.Fatalf("reply reviewed session: %v", err)
	}

	handler := NewHandler(svc)
	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/chat/sessions?draft_review_status=pending_review&limit=10", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var sessions []Session
	if err := json.Unmarshal(rec.Body.Bytes(), &sessions); err != nil {
		t.Fatalf("unmarshal sessions: %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("expected one pending review session, got %d", len(sessions))
	}
	if sessions[0].ID != pending.Session.ID {
		t.Fatalf("expected pending session %s, got %s", pending.Session.ID, sessions[0].ID)
	}
	if sessions[0].DraftReviewStatus != "PENDING_REVIEW" {
		t.Fatalf("expected draft review status PENDING_REVIEW, got %s", sessions[0].DraftReviewStatus)
	}
}

func TestListSessionsOrdersByReviewPriority(t *testing.T) {
	store := newFakeStore()
	now := time.Now().UTC()
	overdue, _ := store.UpsertSession(context.Background(), UpsertSessionInput{
		Channel:       "WHATSAPP",
		ContactKey:    "5511333333333",
		CustomerPhone: "5511333333333",
		LastMessageAt: timePointer(now.Add(-20 * time.Minute)),
		Metadata: map[string]interface{}{
			"agent": map[string]interface{}{
				"status":             "DRAFT_GENERATED",
				"draft_generated_at": now.Add(-20 * time.Minute).Format(time.RFC3339Nano),
			},
		},
	})
	dueSoon, _ := store.UpsertSession(context.Background(), UpsertSessionInput{
		Channel:       "WHATSAPP",
		ContactKey:    "5511222222222",
		CustomerPhone: "5511222222222",
		LastMessageAt: timePointer(now.Add(-10 * time.Minute)),
		Metadata: map[string]interface{}{
			"agent": map[string]interface{}{
				"status":             "DRAFT_GENERATED",
				"draft_generated_at": now.Add(-10 * time.Minute).Format(time.RFC3339Nano),
			},
		},
	})
	fresh, _ := store.UpsertSession(context.Background(), UpsertSessionInput{
		Channel:       "WHATSAPP",
		ContactKey:    "5511111111111",
		CustomerPhone: "5511111111111",
		LastMessageAt: timePointer(now.Add(-2 * time.Minute)),
		Metadata: map[string]interface{}{
			"agent": map[string]interface{}{
				"status":             "DRAFT_GENERATED",
				"draft_generated_at": now.Add(-2 * time.Minute).Format(time.RFC3339Nano),
			},
		},
	})
	reviewed, _ := store.UpsertSession(context.Background(), UpsertSessionInput{
		Channel:       "WHATSAPP",
		ContactKey:    "5511444444444",
		CustomerPhone: "5511444444444",
		LastMessageAt: timePointer(now.Add(-1 * time.Minute)),
		Metadata: map[string]interface{}{
			"agent": map[string]interface{}{
				"status": "DRAFT_REVIEWED",
			},
		},
	})
	normal, _ := store.UpsertSession(context.Background(), UpsertSessionInput{
		Channel:       "WHATSAPP",
		ContactKey:    "5511555555555",
		CustomerPhone: "5511555555555",
		LastMessageAt: timePointer(now),
	})

	handler := NewHandler(NewService(store, config.Config{ChatDebounceWindowMS: 1500}))
	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/chat/sessions?order_by=review_priority&limit=10", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var sessions []Session
	if err := json.Unmarshal(rec.Body.Bytes(), &sessions); err != nil {
		t.Fatalf("unmarshal sessions: %v", err)
	}
	if len(sessions) != 5 {
		t.Fatalf("expected five sessions, got %d", len(sessions))
	}
	if sessions[0].ID != overdue.ID {
		t.Fatalf("expected overdue draft session first, got %s", sessions[0].ID)
	}
	if sessions[1].ID != dueSoon.ID {
		t.Fatalf("expected due soon draft session second, got %s", sessions[1].ID)
	}
	if sessions[2].ID != fresh.ID {
		t.Fatalf("expected fresh draft session third, got %s", sessions[2].ID)
	}
	if sessions[3].ID != reviewed.ID {
		t.Fatalf("expected reviewed draft session fourth, got %s", sessions[3].ID)
	}
	if sessions[4].ID != normal.ID {
		t.Fatalf("expected normal session fifth, got %s", sessions[4].ID)
	}
}

func TestGetSessionsSummaryReturnsReviewCounters(t *testing.T) {
	store := newFakeStore()
	ownerUserID := uuid.NewString()
	now := time.Now().UTC()

	pending, _ := store.UpsertSession(context.Background(), UpsertSessionInput{
		Channel:       "WHATSAPP",
		ContactKey:    "5511000000001",
		CustomerPhone: "5511000000001",
		LastMessageAt: timePointer(now.Add(-3 * time.Minute)),
		Metadata: map[string]interface{}{
			"agent": map[string]interface{}{
				"status": "DRAFT_GENERATED",
			},
		},
	})
	reviewed, _ := store.UpsertSession(context.Background(), UpsertSessionInput{
		Channel:       "WHATSAPP",
		ContactKey:    "5511000000002",
		CustomerPhone: "5511000000002",
		LastMessageAt: timePointer(now.Add(-2 * time.Minute)),
		Metadata: map[string]interface{}{
			"agent": map[string]interface{}{
				"status":              "DRAFT_REVIEWED",
				"reviewed_by_user_id": ownerUserID,
			},
		},
	})
	item := store.sessions[reviewed.ID]
	item.HandoffStatus = "HUMAN"
	item.CurrentOwnerUserID = ownerUserID
	store.sessions[reviewed.ID] = item

	_, _ = store.UpsertSession(context.Background(), UpsertSessionInput{
		Channel:       "WHATSAPP",
		ContactKey:    "5511000000003",
		CustomerPhone: "5511000000003",
		LastMessageAt: timePointer(now.Add(-1 * time.Minute)),
	})

	handler := NewHandler(NewService(store, config.Config{ChatDebounceWindowMS: 1500}))
	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/chat/sessions/summary?channel=whatsapp&draft_review_status=pending_review", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var out SessionsSummary
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal summary: %v", err)
	}
	if out.TotalCount != 3 {
		t.Fatalf("expected total_count 3, got %d", out.TotalCount)
	}
	if out.ReviewSLASeconds != 15*60 {
		t.Fatalf("expected review_sla_seconds 900, got %d", out.ReviewSLASeconds)
	}
	if out.PendingReviewCount != 1 {
		t.Fatalf("expected pending_review_count 1, got %d", out.PendingReviewCount)
	}
	if out.ReviewedCount != 1 {
		t.Fatalf("expected reviewed_count 1, got %d", out.ReviewedCount)
	}
	if out.NoDraftCount != 1 {
		t.Fatalf("expected no_draft_count 1, got %d", out.NoDraftCount)
	}
	if out.HumanOwnedCount != 1 {
		t.Fatalf("expected human_owned_count 1, got %d", out.HumanOwnedCount)
	}
	if out.BotOwnedCount != 2 {
		t.Fatalf("expected bot_owned_count 2, got %d", out.BotOwnedCount)
	}
	if out.DueSoonReviewCount != 0 {
		t.Fatalf("expected due_soon_review_count 0, got %d", out.DueSoonReviewCount)
	}
	if out.OverdueReviewCount != 0 {
		t.Fatalf("expected overdue_review_count 0, got %d", out.OverdueReviewCount)
	}
	if out.HighPriorityReviewCount != 0 {
		t.Fatalf("expected high_priority_review_count 0, got %d", out.HighPriorityReviewCount)
	}
	if out.MediumPriorityReviewCount != 0 {
		t.Fatalf("expected medium_priority_review_count 0, got %d", out.MediumPriorityReviewCount)
	}
	if out.LowPriorityReviewCount != 1 {
		t.Fatalf("expected low_priority_review_count 1, got %d", out.LowPriorityReviewCount)
	}
	if out.HasReviewAlert {
		t.Fatalf("expected has_review_alert false, got true")
	}
	if out.OldestPendingAgeSeconds < 0 {
		t.Fatalf("expected oldest_pending_age_seconds >= 0, got %d", out.OldestPendingAgeSeconds)
	}
	_ = pending
}

func TestGetSessionsSummaryReturnsWarningAlertForDueSoonQueue(t *testing.T) {
	store := newFakeStore()
	now := time.Now().UTC()

	_, _ = store.UpsertSession(context.Background(), UpsertSessionInput{
		Channel:       "WHATSAPP",
		ContactKey:    "5511000000009",
		CustomerPhone: "5511000000009",
		LastMessageAt: timePointer(now.Add(-10 * time.Minute)),
		Metadata: map[string]interface{}{
			"agent": map[string]interface{}{
				"status":             "DRAFT_GENERATED",
				"draft_generated_at": now.Add(-10 * time.Minute).Format(time.RFC3339Nano),
			},
		},
	})

	handler := NewHandler(NewService(store, config.Config{ChatDebounceWindowMS: 1500, ChatReviewSLAMinutes: 15}))
	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/chat/sessions/summary?channel=whatsapp", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var out SessionsSummary
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal summary: %v", err)
	}
	if !out.HasReviewAlert {
		t.Fatalf("expected has_review_alert true")
	}
	if out.ReviewAlertLevel != "WARNING" {
		t.Fatalf("expected review_alert_level WARNING, got %s", out.ReviewAlertLevel)
	}
	if out.ReviewAlertCode != "REVIEW_QUEUE_DUE_SOON" {
		t.Fatalf("expected review_alert_code REVIEW_QUEUE_DUE_SOON, got %s", out.ReviewAlertCode)
	}
	if out.ReviewAlertSessionCount != 1 {
		t.Fatalf("expected review_alert_session_count 1, got %d", out.ReviewAlertSessionCount)
	}
}

func TestListSessionsAndSummaryExposeReviewAging(t *testing.T) {
	store := newFakeStore()
	now := time.Now().UTC()
	_, _ = store.UpsertSession(context.Background(), UpsertSessionInput{
		Channel:       "WHATSAPP",
		ContactKey:    "5511000000010",
		CustomerPhone: "5511000000010",
		LastMessageAt: timePointer(now.Add(-20 * time.Minute)),
		Metadata: map[string]interface{}{
			"agent": map[string]interface{}{
				"status":             "DRAFT_GENERATED",
				"draft_generated_at": now.Add(-20 * time.Minute).Format(time.RFC3339Nano),
			},
		},
	})
	_, _ = store.UpsertSession(context.Background(), UpsertSessionInput{
		Channel:       "WHATSAPP",
		ContactKey:    "5511000000011",
		CustomerPhone: "5511000000011",
		LastMessageAt: timePointer(now.Add(-10 * time.Minute)),
		Metadata: map[string]interface{}{
			"agent": map[string]interface{}{
				"status":             "DRAFT_GENERATED",
				"draft_generated_at": now.Add(-10 * time.Minute).Format(time.RFC3339Nano),
			},
		},
	})
	_, _ = store.UpsertSession(context.Background(), UpsertSessionInput{
		Channel:       "WHATSAPP",
		ContactKey:    "5511000000012",
		CustomerPhone: "5511000000012",
		LastMessageAt: timePointer(now.Add(-2 * time.Minute)),
		Metadata: map[string]interface{}{
			"agent": map[string]interface{}{
				"status":             "DRAFT_GENERATED",
				"draft_generated_at": now.Add(-2 * time.Minute).Format(time.RFC3339Nano),
			},
		},
	})

	svc := NewService(store, config.Config{ChatDebounceWindowMS: 1500, ChatReviewSLAMinutes: 15})
	handler := NewHandler(svc)
	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	listReq := httptest.NewRequest(http.MethodGet, "/chat/sessions?order_by=review_priority&limit=10", nil)
	listRec := httptest.NewRecorder()
	r.ServeHTTP(listRec, listReq)

	if listRec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, listRec.Code)
	}

	var sessions []Session
	if err := json.Unmarshal(listRec.Body.Bytes(), &sessions); err != nil {
		t.Fatalf("unmarshal sessions: %v", err)
	}
	if len(sessions) != 3 {
		t.Fatalf("expected three sessions, got %d", len(sessions))
	}
	if sessions[0].ContactKey != "5511000000010" {
		t.Fatalf("expected overdue session first, got %s", sessions[0].ContactKey)
	}
	if sessions[1].ContactKey != "5511000000011" {
		t.Fatalf("expected due soon session second, got %s", sessions[1].ContactKey)
	}
	if sessions[2].ContactKey != "5511000000012" {
		t.Fatalf("expected fresh session third, got %s", sessions[2].ContactKey)
	}
	byContactKey := map[string]Session{}
	for _, item := range sessions {
		byContactKey[item.ContactKey] = item
	}
	if byContactKey["5511000000010"].DraftPendingAgeBucket != "OVERDUE" {
		t.Fatalf("expected overdue bucket for 5511000000010, got %s", byContactKey["5511000000010"].DraftPendingAgeBucket)
	}
	if !byContactKey["5511000000010"].DraftReviewOverdue {
		t.Fatalf("expected overdue session to be flagged")
	}
	if byContactKey["5511000000010"].DraftReviewPriority != "HIGH" {
		t.Fatalf("expected HIGH priority for 5511000000010, got %s", byContactKey["5511000000010"].DraftReviewPriority)
	}
	if !byContactKey["5511000000010"].DraftReviewAlertActive {
		t.Fatalf("expected active alert for 5511000000010")
	}
	if byContactKey["5511000000010"].DraftReviewAlertLevel != "CRITICAL" {
		t.Fatalf("expected CRITICAL alert for 5511000000010, got %s", byContactKey["5511000000010"].DraftReviewAlertLevel)
	}
	if byContactKey["5511000000011"].DraftPendingAgeBucket != "DUE_SOON" {
		t.Fatalf("expected due soon bucket for 5511000000011, got %s", byContactKey["5511000000011"].DraftPendingAgeBucket)
	}
	if byContactKey["5511000000011"].DraftReviewOverdue {
		t.Fatalf("expected due soon session not to be overdue")
	}
	if byContactKey["5511000000011"].DraftReviewPriority != "MEDIUM" {
		t.Fatalf("expected MEDIUM priority for 5511000000011, got %s", byContactKey["5511000000011"].DraftReviewPriority)
	}
	if !byContactKey["5511000000011"].DraftReviewAlertActive {
		t.Fatalf("expected active alert for 5511000000011")
	}
	if byContactKey["5511000000011"].DraftReviewAlertLevel != "WARNING" {
		t.Fatalf("expected WARNING alert for 5511000000011, got %s", byContactKey["5511000000011"].DraftReviewAlertLevel)
	}
	if byContactKey["5511000000012"].DraftPendingAgeBucket != "FRESH" {
		t.Fatalf("expected fresh bucket for 5511000000012, got %s", byContactKey["5511000000012"].DraftPendingAgeBucket)
	}
	if byContactKey["5511000000012"].DraftReviewPriority != "LOW" {
		t.Fatalf("expected LOW priority for 5511000000012, got %s", byContactKey["5511000000012"].DraftReviewPriority)
	}
	if byContactKey["5511000000012"].DraftReviewAlertActive {
		t.Fatalf("expected inactive alert for 5511000000012")
	}

	summaryReq := httptest.NewRequest(http.MethodGet, "/chat/sessions/summary?channel=whatsapp", nil)
	summaryRec := httptest.NewRecorder()
	r.ServeHTTP(summaryRec, summaryReq)

	if summaryRec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, summaryRec.Code)
	}

	var summary SessionsSummary
	if err := json.Unmarshal(summaryRec.Body.Bytes(), &summary); err != nil {
		t.Fatalf("unmarshal summary: %v", err)
	}
	if summary.PendingReviewCount != 3 {
		t.Fatalf("expected pending_review_count 3, got %d", summary.PendingReviewCount)
	}
	if summary.DueSoonReviewCount != 1 {
		t.Fatalf("expected due_soon_review_count 1, got %d", summary.DueSoonReviewCount)
	}
	if summary.OverdueReviewCount != 1 {
		t.Fatalf("expected overdue_review_count 1, got %d", summary.OverdueReviewCount)
	}
	if summary.HighPriorityReviewCount != 1 {
		t.Fatalf("expected high_priority_review_count 1, got %d", summary.HighPriorityReviewCount)
	}
	if summary.MediumPriorityReviewCount != 1 {
		t.Fatalf("expected medium_priority_review_count 1, got %d", summary.MediumPriorityReviewCount)
	}
	if summary.LowPriorityReviewCount != 1 {
		t.Fatalf("expected low_priority_review_count 1, got %d", summary.LowPriorityReviewCount)
	}
	if !summary.HasReviewAlert {
		t.Fatalf("expected has_review_alert true")
	}
	if summary.ReviewAlertLevel != "CRITICAL" {
		t.Fatalf("expected review_alert_level CRITICAL, got %s", summary.ReviewAlertLevel)
	}
	if summary.ReviewAlertCode != "REVIEW_QUEUE_OVERDUE" {
		t.Fatalf("expected review_alert_code REVIEW_QUEUE_OVERDUE, got %s", summary.ReviewAlertCode)
	}
	if summary.ReviewAlertSessionCount != 1 {
		t.Fatalf("expected review_alert_session_count 1, got %d", summary.ReviewAlertSessionCount)
	}
	if summary.OldestPendingAgeSeconds < 20*60-5 {
		t.Fatalf("expected oldest_pending_age_seconds around 1200, got %d", summary.OldestPendingAgeSeconds)
	}
}

func TestReplyCreatesOutboundRecordForHumanOwner(t *testing.T) {
	store := newFakeStore()
	session, _ := store.seedSessionWithMessage("5511888888888", "ola")
	ownerID := uuid.NewString()
	item := store.sessions[session.ID]
	item.HandoffStatus = "HUMAN"
	item.CurrentOwnerUserID = ownerID
	store.sessions[session.ID] = item

	handler := NewHandler(NewService(store, config.Config{ChatDebounceWindowMS: 1500}))
	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodPost, "/chat/sessions/"+session.ID+"/reply", bytes.NewBufferString(`{
		"owner_user_id":"`+ownerID+`",
		"body":"posso te ajudar com mais algo?",
		"sender_name":"Operador",
		"idempotency_key":"reply-1",
		"metadata":{"source":"dashboard"}
	}`))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}

	var out ReplyResult
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if out.Idempotent {
		t.Fatalf("expected new reply, got idempotent response")
	}
	if out.Session.HandoffStatus != "HUMAN" {
		t.Fatalf("expected session to remain HUMAN, got %s", out.Session.HandoffStatus)
	}
	if out.Message.Direction != "OUTBOUND" {
		t.Fatalf("expected outbound message, got %s", out.Message.Direction)
	}
	if out.Message.ProcessingStatus != "MANUAL_PENDING" {
		t.Fatalf("expected MANUAL_PENDING processing status, got %s", out.Message.ProcessingStatus)
	}
	if out.Outbound.Status != "MANUAL_PENDING" {
		t.Fatalf("expected MANUAL_PENDING outbound status, got %s", out.Outbound.Status)
	}
	if out.Outbound.Recipient != session.ContactKey {
		t.Fatalf("expected outbound recipient %s, got %s", session.ContactKey, out.Outbound.Recipient)
	}
	buffer, ok := out.Session.Metadata["buffer"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected buffer metadata to be present")
	}
	if got := asString(buffer["status"]); got != bufferStatusIdle {
		t.Fatalf("expected buffer status %s, got %s", bufferStatusIdle, got)
	}
	if len(store.outbounds) != 1 {
		t.Fatalf("expected one outbound record, got %d", len(store.outbounds))
	}
}

func TestReplyDeliversImmediatelyWhenSenderIsEnabled(t *testing.T) {
	store := newFakeStore()
	session, _ := store.seedSessionWithMessage("5511888888888@s.whatsapp.net", "ola")
	ownerID := uuid.NewString()
	item := store.sessions[session.ID]
	item.HandoffStatus = "HUMAN"
	item.CurrentOwnerUserID = ownerID
	store.sessions[session.ID] = item

	sender := &fakeReplySender{
		enabled: true,
		result: SendReplyResult{
			ProviderMessageID: "MSG-SENT-1",
			ProviderStatus:    "SENT",
			Payload:           map[string]interface{}{"provider": "EVOLUTION"},
			SentAt:            time.Now().UTC(),
		},
	}
	handler := NewHandler(NewService(store, config.Config{ChatDebounceWindowMS: 1500}, sender))
	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodPost, "/chat/sessions/"+session.ID+"/reply", bytes.NewBufferString(`{
		"owner_user_id":"`+ownerID+`",
		"body":"posso te ajudar com mais algo?",
		"sender_name":"Operador",
		"idempotency_key":"reply-send-1"
	}`))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}

	var out ReplyResult
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if out.Outbound.Status != "SENT" {
		t.Fatalf("expected outbound status SENT, got %s", out.Outbound.Status)
	}
	if out.Outbound.ProviderMessageID != "MSG-SENT-1" {
		t.Fatalf("expected provider message id MSG-SENT-1, got %s", out.Outbound.ProviderMessageID)
	}
	if out.Message.ProcessingStatus != "SENT" {
		t.Fatalf("expected processing status SENT, got %s", out.Message.ProcessingStatus)
	}
	if sender.calls != 1 {
		t.Fatalf("expected one sender call, got %d", sender.calls)
	}
}

func TestReplyRejectsWhenHumanOwnerMismatch(t *testing.T) {
	store := newFakeStore()
	session, _ := store.seedSessionWithMessage("5511888888888", "ola")
	item := store.sessions[session.ID]
	item.HandoffStatus = "HUMAN"
	item.CurrentOwnerUserID = uuid.NewString()
	store.sessions[session.ID] = item

	handler := NewHandler(NewService(store, config.Config{ChatDebounceWindowMS: 1500}))
	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodPost, "/chat/sessions/"+session.ID+"/reply", bytes.NewBufferString(`{
		"owner_user_id":"`+uuid.NewString()+`",
		"body":"resposta"
	}`))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d", http.StatusConflict, rec.Code)
	}
}

func TestReplyRejectsWhenSessionHasNoActiveHumanOwner(t *testing.T) {
	store := newFakeStore()
	session, _ := store.seedSessionWithMessage("5511888888888", "ola")

	handler := NewHandler(NewService(store, config.Config{ChatDebounceWindowMS: 1500}))
	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodPost, "/chat/sessions/"+session.ID+"/reply", bytes.NewBufferString(`{
		"owner_user_id":"`+uuid.NewString()+`",
		"body":"resposta"
	}`))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d", http.StatusConflict, rec.Code)
	}
}

func TestReplyReturnsExistingOnIdempotency(t *testing.T) {
	store := newFakeStore()
	session, _ := store.seedSessionWithMessage("5511888888888", "ola")
	ownerID := uuid.NewString()
	item := store.sessions[session.ID]
	item.HandoffStatus = "HUMAN"
	item.CurrentOwnerUserID = ownerID
	store.sessions[session.ID] = item

	handler := NewHandler(NewService(store, config.Config{ChatDebounceWindowMS: 1500}))
	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	body := `{
		"owner_user_id":"` + ownerID + `",
		"body":"resposta assistida",
		"idempotency_key":"reply-idem-1"
	}`

	firstReq := httptest.NewRequest(http.MethodPost, "/chat/sessions/"+session.ID+"/reply", bytes.NewBufferString(body))
	firstRec := httptest.NewRecorder()
	r.ServeHTTP(firstRec, firstReq)
	if firstRec.Code != http.StatusCreated {
		t.Fatalf("expected first status %d, got %d", http.StatusCreated, firstRec.Code)
	}

	secondReq := httptest.NewRequest(http.MethodPost, "/chat/sessions/"+session.ID+"/reply", bytes.NewBufferString(body))
	secondRec := httptest.NewRecorder()
	r.ServeHTTP(secondRec, secondReq)
	if secondRec.Code != http.StatusOK {
		t.Fatalf("expected second status %d, got %d", http.StatusOK, secondRec.Code)
	}

	var out ReplyResult
	if err := json.Unmarshal(secondRec.Body.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if !out.Idempotent {
		t.Fatalf("expected idempotent response")
	}
	if len(store.outbounds) != 1 {
		t.Fatalf("expected one outbound record, got %d", len(store.outbounds))
	}
}

func TestReplyRetriesDeliveryOnSameIdempotencyKeyAfterFailure(t *testing.T) {
	store := newFakeStore()
	session, _ := store.seedSessionWithMessage("5511888888888@s.whatsapp.net", "ola")
	ownerID := uuid.NewString()
	item := store.sessions[session.ID]
	item.HandoffStatus = "HUMAN"
	item.CurrentOwnerUserID = ownerID
	store.sessions[session.ID] = item

	sender := &fakeReplySender{
		enabled: true,
		errs: []error{
			errors.New("gateway timeout"),
			nil,
		},
		results: []SendReplyResult{
			{},
			{
				ProviderMessageID: "MSG-RETRY-1",
				ProviderStatus:    "SENT",
				Payload:           map[string]interface{}{"provider": "EVOLUTION"},
				SentAt:            time.Now().UTC(),
			},
		},
	}
	handler := NewHandler(NewService(store, config.Config{ChatDebounceWindowMS: 1500}, sender))
	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	body := `{
		"owner_user_id":"` + ownerID + `",
		"body":"resposta assistida",
		"idempotency_key":"reply-send-retry-1"
	}`

	firstReq := httptest.NewRequest(http.MethodPost, "/chat/sessions/"+session.ID+"/reply", bytes.NewBufferString(body))
	firstRec := httptest.NewRecorder()
	r.ServeHTTP(firstRec, firstReq)
	if firstRec.Code != http.StatusBadGateway {
		t.Fatalf("expected first status %d, got %d", http.StatusBadGateway, firstRec.Code)
	}

	secondReq := httptest.NewRequest(http.MethodPost, "/chat/sessions/"+session.ID+"/reply", bytes.NewBufferString(body))
	secondRec := httptest.NewRecorder()
	r.ServeHTTP(secondRec, secondReq)
	if secondRec.Code != http.StatusOK {
		t.Fatalf("expected second status %d, got %d", http.StatusOK, secondRec.Code)
	}

	var out ReplyResult
	if err := json.Unmarshal(secondRec.Body.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if !out.Idempotent {
		t.Fatalf("expected idempotent response on retry")
	}
	if out.Outbound.Status != "SENT" {
		t.Fatalf("expected outbound status SENT after retry, got %s", out.Outbound.Status)
	}
	if out.Outbound.ProviderMessageID != "MSG-RETRY-1" {
		t.Fatalf("expected provider message id MSG-RETRY-1, got %s", out.Outbound.ProviderMessageID)
	}
	if sender.calls != 2 {
		t.Fatalf("expected two sender calls, got %d", sender.calls)
	}
}

func TestReplyApprovesAutomationDraftWhenDraftMessageIDIsProvided(t *testing.T) {
	store := newFakeStore()
	svc := NewService(store, config.Config{ChatDebounceWindowMS: 1500}, &fakeAgentRunner{
		enabled: true,
		result: RunAgentResult{
			ReplyText:          "Temos vaga para esse trecho.",
			Model:              "gpt-test",
			ProviderResponseID: "resp-review-1",
		},
	})

	ingested, err := svc.Ingest(context.Background(), IngestMessageInput{
		ContactKey: "5511888888888",
		Message: IngestMessagePayload{
			Direction:         "INBOUND",
			ProviderMessageID: "msg-review-1",
			IdempotencyKey:    "idem-review-1",
			Body:              "quero uma passagem",
		},
	})
	if err != nil {
		t.Fatalf("ingest message: %v", err)
	}

	reprocessed, err := svc.Reprocess(context.Background(), ReprocessInput{SessionID: ingested.Session.ID})
	if err != nil {
		t.Fatalf("reprocess message: %v", err)
	}
	if reprocessed.Draft == nil {
		t.Fatalf("expected draft to be created")
	}

	ownerID := uuid.NewString()
	session := store.sessions[ingested.Session.ID]
	session.HandoffStatus = "HUMAN"
	session.CurrentOwnerUserID = ownerID
	store.sessions[session.ID] = session

	handler := NewHandler(svc)
	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodPost, "/chat/sessions/"+session.ID+"/reply", bytes.NewBufferString(`{
		"owner_user_id":"`+ownerID+`",
		"draft_message_id":"`+reprocessed.Draft.ID+`",
		"idempotency_key":"reply-review-1"
	}`))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}

	var out ReplyResult
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if out.Draft == nil {
		t.Fatalf("expected reviewed draft in response")
	}
	if out.Draft.ProcessingStatus != messageStatusAutomationReviewed {
		t.Fatalf("expected reviewed draft status %s, got %s", messageStatusAutomationReviewed, out.Draft.ProcessingStatus)
	}
	if out.Message.Body != reprocessed.Draft.Body {
		t.Fatalf("expected reply body to reuse draft body")
	}
	if got := asString(out.Message.Payload["draft_message_id"]); got != reprocessed.Draft.ID {
		t.Fatalf("expected draft_message_id %s, got %s", reprocessed.Draft.ID, got)
	}
	agent, ok := out.Session.Metadata["agent"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected agent metadata")
	}
	if got := asString(agent["status"]); got != agentStatusDraftReviewed {
		t.Fatalf("expected agent status %s, got %s", agentStatusDraftReviewed, got)
	}
}

func TestReplyCanEditAutomationDraftBeforeSending(t *testing.T) {
	store := newFakeStore()
	svc := NewService(store, config.Config{ChatDebounceWindowMS: 1500}, &fakeAgentRunner{
		enabled: true,
		result: RunAgentResult{
			ReplyText:          "Temos vaga para esse trecho.",
			Model:              "gpt-test",
			ProviderResponseID: "resp-review-2",
		},
	})

	ingested, err := svc.Ingest(context.Background(), IngestMessageInput{
		ContactKey: "5511888888888",
		Message: IngestMessagePayload{
			Direction:         "INBOUND",
			ProviderMessageID: "msg-review-2",
			IdempotencyKey:    "idem-review-2",
			Body:              "quero uma passagem",
		},
	})
	if err != nil {
		t.Fatalf("ingest message: %v", err)
	}

	reprocessed, err := svc.Reprocess(context.Background(), ReprocessInput{SessionID: ingested.Session.ID})
	if err != nil {
		t.Fatalf("reprocess message: %v", err)
	}
	if reprocessed.Draft == nil {
		t.Fatalf("expected draft to be created")
	}

	ownerID := uuid.NewString()
	session := store.sessions[ingested.Session.ID]
	session.HandoffStatus = "HUMAN"
	session.CurrentOwnerUserID = ownerID
	store.sessions[session.ID] = session

	handler := NewHandler(svc)
	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodPost, "/chat/sessions/"+session.ID+"/reply", bytes.NewBufferString(`{
		"owner_user_id":"`+ownerID+`",
		"draft_message_id":"`+reprocessed.Draft.ID+`",
		"body":"Temos vaga. Quer que eu te passe as datas disponiveis?",
		"idempotency_key":"reply-review-2"
	}`))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}

	var out ReplyResult
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if out.Message.Body != "Temos vaga. Quer que eu te passe as datas disponiveis?" {
		t.Fatalf("expected edited reply body, got %q", out.Message.Body)
	}
	if out.Draft == nil {
		t.Fatalf("expected reviewed draft")
	}
	if got := asString(out.Draft.NormalizedPayload["review_action"]); got != "EDITED" {
		t.Fatalf("expected review_action EDITED, got %s", got)
	}
}

func TestReplyRejectsDraftReviewWhenDraftIsNotActiveAutomationDraft(t *testing.T) {
	store := newFakeStore()
	svc := NewService(store, config.Config{ChatDebounceWindowMS: 1500}, &fakeAgentRunner{
		enabled: true,
		result: RunAgentResult{
			ReplyText:          "Temos vaga para esse trecho.",
			Model:              "gpt-test",
			ProviderResponseID: "resp-review-3",
		},
	})

	ingested, err := svc.Ingest(context.Background(), IngestMessageInput{
		ContactKey: "5511888888888",
		Message: IngestMessagePayload{
			Direction:         "INBOUND",
			ProviderMessageID: "msg-review-3",
			IdempotencyKey:    "idem-review-3",
			Body:              "quero uma passagem",
		},
	})
	if err != nil {
		t.Fatalf("ingest message: %v", err)
	}

	reprocessed, err := svc.Reprocess(context.Background(), ReprocessInput{SessionID: ingested.Session.ID})
	if err != nil {
		t.Fatalf("reprocess message: %v", err)
	}
	if reprocessed.Draft == nil {
		t.Fatalf("expected draft to be created")
	}

	ownerID := uuid.NewString()
	session := store.sessions[ingested.Session.ID]
	session.HandoffStatus = "HUMAN"
	session.CurrentOwnerUserID = ownerID
	store.sessions[session.ID] = session

	handler := NewHandler(svc)
	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	firstReq := httptest.NewRequest(http.MethodPost, "/chat/sessions/"+session.ID+"/reply", bytes.NewBufferString(`{
		"owner_user_id":"`+ownerID+`",
		"draft_message_id":"`+reprocessed.Draft.ID+`",
		"idempotency_key":"reply-review-3a"
	}`))
	firstRec := httptest.NewRecorder()
	r.ServeHTTP(firstRec, firstReq)
	if firstRec.Code != http.StatusCreated {
		t.Fatalf("expected first status %d, got %d", http.StatusCreated, firstRec.Code)
	}

	secondReq := httptest.NewRequest(http.MethodPost, "/chat/sessions/"+session.ID+"/reply", bytes.NewBufferString(`{
		"owner_user_id":"`+ownerID+`",
		"draft_message_id":"`+reprocessed.Draft.ID+`",
		"idempotency_key":"reply-review-3b"
	}`))
	secondRec := httptest.NewRecorder()
	r.ServeHTTP(secondRec, secondReq)

	if secondRec.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d", http.StatusConflict, secondRec.Code)
	}
}

func TestReplyRejectsInvalidDraftMessageID(t *testing.T) {
	store := newFakeStore()
	session, _ := store.seedSessionWithMessage("5511888888888", "ola")
	ownerID := uuid.NewString()
	item := store.sessions[session.ID]
	item.HandoffStatus = "HUMAN"
	item.CurrentOwnerUserID = ownerID
	store.sessions[session.ID] = item

	handler := NewHandler(NewService(store, config.Config{ChatDebounceWindowMS: 1500}))
	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodPost, "/chat/sessions/"+session.ID+"/reply", bytes.NewBufferString(`{
		"owner_user_id":"`+ownerID+`",
		"draft_message_id":"not-a-uuid"
	}`))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestIngestMessageWithHumanOwnerBlocksAgentProcessing(t *testing.T) {
	store := newFakeStore()
	session, _ := store.seedSessionWithMessage("5511888888888", "ola")
	ownerID := uuid.NewString()
	item := store.sessions[session.ID]
	item.HandoffStatus = "HUMAN"
	item.CurrentOwnerUserID = ownerID
	store.sessions[session.ID] = item

	handler := NewHandler(NewService(store, config.Config{ChatDebounceWindowMS: 1500}))
	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodPost, "/chat/messages/ingest", bytes.NewBufferString(`{
		"contact_key":"5511888888888",
		"message":{"direction":"INBOUND","provider_message_id":"msg-human-1","idempotency_key":"idem-human-1","body":"preciso falar com atendente","processing_status":"RECEIVED"}
	}`))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}

	var out IngestMessageResult
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if out.Message.ProcessingStatus != "HUMAN_OWNED_PENDING" {
		t.Fatalf("expected HUMAN_OWNED_PENDING, got %s", out.Message.ProcessingStatus)
	}
	if got := asString(out.Message.NormalizedPayload["agent_block_reason"]); got != agentBlockReasonHumanOwnerActive {
		t.Fatalf("expected agent block reason %s, got %s", agentBlockReasonHumanOwnerActive, got)
	}
	if got := asString(out.Message.NormalizedPayload["current_owner_user_id"]); got != ownerID {
		t.Fatalf("expected current owner %s, got %s", ownerID, got)
	}
	buffer, ok := out.Session.Metadata["buffer"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected buffer metadata to be present")
	}
	if got := asString(buffer["status"]); got != bufferStatusIdle {
		t.Fatalf("expected blocked buffer to remain %s, got %s", bufferStatusIdle, got)
	}
	if got := asString(buffer["agent_block_reason"]); got != agentBlockReasonHumanOwnerActive {
		t.Fatalf("expected buffer block reason %s, got %s", agentBlockReasonHumanOwnerActive, got)
	}
}

func TestPresenceSignalIsSkippedWhenHumanOwnerIsActive(t *testing.T) {
	store := newFakeStore()
	svc := NewService(store, config.Config{ChatDebounceWindowMS: 1500})

	session, _ := store.seedSessionWithMessage("5511888888888", "ola")
	ownerID := uuid.NewString()
	item := store.sessions[session.ID]
	item.HandoffStatus = "HUMAN"
	item.CurrentOwnerUserID = ownerID
	item.Metadata["buffer"] = map[string]interface{}{
		"status":        bufferStatusPending,
		"pending_until": time.Now().UTC().Add(1 * time.Second).Format(time.RFC3339Nano),
	}
	store.sessions[session.ID] = item

	result, err := svc.ApplyPresenceSignal(context.Background(), ApplyPresenceSignalInput{
		ContactKey:     "5511888888888",
		PresenceStatus: "typing",
	})
	if err != nil {
		t.Fatalf("apply presence: %v", err)
	}
	if result.Status != "skipped" {
		t.Fatalf("expected skipped status, got %s", result.Status)
	}
	if result.Reason != agentBlockReasonHumanOwnerActive {
		t.Fatalf("expected reason %s, got %s", agentBlockReasonHumanOwnerActive, result.Reason)
	}
}

func TestRequestHandoffCreatesRecordAndUpdatesSession(t *testing.T) {
	store := newFakeStore()
	session, _ := store.seedSessionWithMessage("5511888888888", "ola")
	handler := NewHandler(NewService(store, config.Config{ChatDebounceWindowMS: 1500}))

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodPost, "/chat/sessions/"+session.ID+"/handoff", bytes.NewBufferString(`{
		"requested_by":"operator",
		"reason":"cliente pediu atendimento humano",
		"metadata":{"source":"dashboard"}
	}`))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}

	var out RequestHandoffResult
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if out.Session.HandoffStatus != "HUMAN_REQUESTED" {
		t.Fatalf("expected session handoff status HUMAN_REQUESTED, got %s", out.Session.HandoffStatus)
	}
	if out.Handoff.Status != "REQUESTED" {
		t.Fatalf("expected handoff status REQUESTED, got %s", out.Handoff.Status)
	}
	if out.Handoff.RequestedBy != "OPERATOR" {
		t.Fatalf("expected requested_by to be normalized, got %s", out.Handoff.RequestedBy)
	}
	if len(store.handoffs) != 1 {
		t.Fatalf("expected one persisted handoff, got %d", len(store.handoffs))
	}
	if out.Session.CurrentOwnerUserID != "" {
		t.Fatalf("expected session without owner assignment, got %s", out.Session.CurrentOwnerUserID)
	}
}

func TestRequestHandoffAssignsHumanOwnerWhenProvided(t *testing.T) {
	store := newFakeStore()
	session, _ := store.seedSessionWithMessage("5511888888888", "ola")
	handler := NewHandler(NewService(store, config.Config{ChatDebounceWindowMS: 1500}))

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	ownerID := uuid.NewString()
	req := httptest.NewRequest(http.MethodPost, "/chat/sessions/"+session.ID+"/handoff", bytes.NewBufferString(`{
		"requested_by":"operator",
		"assigned_user_id":"`+ownerID+`",
		"reason":"cliente pediu atendimento humano"
	}`))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}

	var out RequestHandoffResult
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if out.Session.HandoffStatus != "HUMAN" {
		t.Fatalf("expected session handoff status HUMAN, got %s", out.Session.HandoffStatus)
	}
	if out.Session.CurrentOwnerUserID != ownerID {
		t.Fatalf("expected current owner %s, got %s", ownerID, out.Session.CurrentOwnerUserID)
	}
	if out.Handoff.AssignedUserID != ownerID {
		t.Fatalf("expected handoff assigned user %s, got %s", ownerID, out.Handoff.AssignedUserID)
	}
}

func TestRequestHandoffRejectsInvalidAssignedUserID(t *testing.T) {
	store := newFakeStore()
	session, _ := store.seedSessionWithMessage("5511888888888", "ola")
	handler := NewHandler(NewService(store, config.Config{ChatDebounceWindowMS: 1500}))

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodPost, "/chat/sessions/"+session.ID+"/handoff", bytes.NewBufferString(`{
		"assigned_user_id":"not-a-uuid"
	}`))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestRequestHandoffRejectsWhenAlreadyActive(t *testing.T) {
	store := newFakeStore()
	session, _ := store.seedSessionWithMessage("5511888888888", "ola")
	item := store.sessions[session.ID]
	item.HandoffStatus = "HUMAN_REQUESTED"
	store.sessions[session.ID] = item
	handler := NewHandler(NewService(store, config.Config{ChatDebounceWindowMS: 1500}))

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodPost, "/chat/sessions/"+session.ID+"/handoff", bytes.NewBufferString(`{}`))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d", http.StatusConflict, rec.Code)
	}
}

func TestResumeSessionResolvesActiveHandoff(t *testing.T) {
	store := newFakeStore()
	session, _ := store.seedSessionWithMessage("5511888888888", "ola")
	ownerID := uuid.NewString()
	_, err := store.RequestHandoff(context.Background(), RequestHandoffInput{
		SessionID:      session.ID,
		RequestedBy:    "OPERATOR",
		Reason:         "cliente pediu atendimento humano",
		AssignedUserID: ownerID,
		Metadata:       map[string]interface{}{"source": "dashboard"},
	})
	if err != nil {
		t.Fatalf("seed handoff: %v", err)
	}

	handler := NewHandler(NewService(store, config.Config{ChatDebounceWindowMS: 1500}))
	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodPost, "/chat/sessions/"+session.ID+"/resume", bytes.NewBufferString(`{
		"resumed_by":"operator",
		"reason":"atendimento finalizado",
		"metadata":{"source":"dashboard"}
	}`))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var out ResumeSessionResult
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if out.Session.HandoffStatus != "BOT" {
		t.Fatalf("expected session handoff status BOT, got %s", out.Session.HandoffStatus)
	}
	if out.Session.CurrentOwnerUserID != "" {
		t.Fatalf("expected current owner to be cleared, got %s", out.Session.CurrentOwnerUserID)
	}
	if out.Handoff.Status != "RESOLVED" {
		t.Fatalf("expected handoff status RESOLVED, got %s", out.Handoff.Status)
	}
	if out.Handoff.AssignedUserID != ownerID {
		t.Fatalf("expected assigned user %s, got %s", ownerID, out.Handoff.AssignedUserID)
	}
	if out.Handoff.ResolvedAt == nil {
		t.Fatalf("expected resolved_at to be set")
	}
	if got := asString(out.Handoff.Metadata["resumed_by"]); got != "OPERATOR" {
		t.Fatalf("expected resumed_by metadata OPERATOR, got %s", got)
	}
}

func TestResumeSessionRejectsWhenNoActiveHandoff(t *testing.T) {
	store := newFakeStore()
	session, _ := store.seedSessionWithMessage("5511888888888", "ola")
	handler := NewHandler(NewService(store, config.Config{ChatDebounceWindowMS: 1500}))

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodPost, "/chat/sessions/"+session.ID+"/resume", bytes.NewBufferString(`{}`))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d", http.StatusConflict, rec.Code)
	}
}

func TestIngestMessageAggregatesWithinDebounceWindow(t *testing.T) {
	store := newFakeStore()
	handler := NewHandler(NewService(store, config.Config{ChatDebounceWindowMS: 2000}))

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	firstReq := httptest.NewRequest(http.MethodPost, "/chat/messages/ingest", bytes.NewBufferString(`{
		"contact_key":"5511999999999",
		"message":{"direction":"INBOUND","provider_message_id":"msg-buffer-1","idempotency_key":"idem-buffer-1","body":"primeira"}
	}`))
	firstRec := httptest.NewRecorder()
	r.ServeHTTP(firstRec, firstReq)
	if firstRec.Code != http.StatusCreated {
		t.Fatalf("expected first call to create message, got %d", firstRec.Code)
	}

	time.Sleep(10 * time.Millisecond)

	secondReq := httptest.NewRequest(http.MethodPost, "/chat/messages/ingest", bytes.NewBufferString(`{
		"contact_key":"5511999999999",
		"message":{"direction":"INBOUND","provider_message_id":"msg-buffer-2","idempotency_key":"idem-buffer-2","body":"segunda"}
	}`))
	secondRec := httptest.NewRecorder()
	r.ServeHTTP(secondRec, secondReq)
	if secondRec.Code != http.StatusCreated {
		t.Fatalf("expected second call to create message, got %d", secondRec.Code)
	}

	var out IngestMessageResult
	if err := json.Unmarshal(secondRec.Body.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	buffer, ok := out.Session.Metadata["buffer"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected buffer metadata to be present")
	}
	if got := asInt(buffer["message_count"]); got != 2 {
		t.Fatalf("expected two buffered messages, got %d", got)
	}
	if asString(buffer["last_message_body"]) != "segunda" {
		t.Fatalf("expected latest buffered body to be tracked")
	}
}

func TestOutboundMessageClearsPendingBuffer(t *testing.T) {
	store := newFakeStore()
	handler := NewHandler(NewService(store, config.Config{ChatDebounceWindowMS: 1500}))

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	inboundReq := httptest.NewRequest(http.MethodPost, "/chat/messages/ingest", bytes.NewBufferString(`{
		"contact_key":"5511999999999",
		"message":{"direction":"INBOUND","provider_message_id":"msg-inbound","idempotency_key":"idem-inbound","body":"oi"}
	}`))
	inboundRec := httptest.NewRecorder()
	r.ServeHTTP(inboundRec, inboundReq)
	if inboundRec.Code != http.StatusCreated {
		t.Fatalf("expected inbound call to create message, got %d", inboundRec.Code)
	}

	outboundReq := httptest.NewRequest(http.MethodPost, "/chat/messages/ingest", bytes.NewBufferString(`{
		"contact_key":"5511999999999",
		"message":{"direction":"OUTBOUND","provider_message_id":"msg-outbound","idempotency_key":"idem-outbound","body":"resposta"}
	}`))
	outboundRec := httptest.NewRecorder()
	r.ServeHTTP(outboundRec, outboundReq)
	if outboundRec.Code != http.StatusCreated {
		t.Fatalf("expected outbound call to create message, got %d", outboundRec.Code)
	}

	var out IngestMessageResult
	if err := json.Unmarshal(outboundRec.Body.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	buffer, ok := out.Session.Metadata["buffer"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected buffer metadata to be present")
	}
	if got := asString(buffer["status"]); got != bufferStatusIdle {
		t.Fatalf("expected buffer status %s, got %s", bufferStatusIdle, got)
	}
	if got := asInt(buffer["message_count"]); got != 0 {
		t.Fatalf("expected cleared buffer count, got %d", got)
	}
}

func TestPresenceTypingExtendsPendingBuffer(t *testing.T) {
	store := newFakeStore()
	svc := NewService(store, config.Config{ChatDebounceWindowMS: 2000})

	ingested, err := svc.Ingest(context.Background(), IngestMessageInput{
		ContactKey: "5511999999999",
		Message: IngestMessagePayload{
			Direction:         "INBOUND",
			ProviderMessageID: "msg-presence-1",
			IdempotencyKey:    "idem-presence-1",
			Body:              "oi",
		},
	})
	if err != nil {
		t.Fatalf("ingest message: %v", err)
	}

	before := parseBufferTime(ingested.Session.Metadata["buffer"].(map[string]interface{})["pending_until"])
	if before == nil {
		t.Fatalf("expected initial pending_until")
	}

	observedAt := before.Add(-500 * time.Millisecond)
	result, err := svc.ApplyPresenceSignal(context.Background(), ApplyPresenceSignalInput{
		ContactKey:     "5511999999999",
		PresenceStatus: "typing",
		ObservedAt:     &observedAt,
	})
	if err != nil {
		t.Fatalf("apply presence: %v", err)
	}

	after := parseBufferTime(result.Session.Metadata["buffer"].(map[string]interface{})["pending_until"])
	if after == nil || !after.After(*before) {
		t.Fatalf("expected pending_until to be extended")
	}
	if result.Status != "accepted" {
		t.Fatalf("expected accepted status, got %s", result.Status)
	}
}

func TestPresencePausedShortensPendingBuffer(t *testing.T) {
	store := newFakeStore()
	svc := NewService(store, config.Config{ChatDebounceWindowMS: 2000})

	ingested, err := svc.Ingest(context.Background(), IngestMessageInput{
		ContactKey: "5511999999999",
		Message: IngestMessagePayload{
			Direction:         "INBOUND",
			ProviderMessageID: "msg-presence-2",
			IdempotencyKey:    "idem-presence-2",
			Body:              "oi",
		},
	})
	if err != nil {
		t.Fatalf("ingest message: %v", err)
	}

	before := parseBufferTime(ingested.Session.Metadata["buffer"].(map[string]interface{})["pending_until"])
	if before == nil {
		t.Fatalf("expected initial pending_until")
	}

	observedAt := ingested.Message.ReceivedAt.Add(500 * time.Millisecond)
	result, err := svc.ApplyPresenceSignal(context.Background(), ApplyPresenceSignalInput{
		ContactKey:     "5511999999999",
		PresenceStatus: "paused",
		ObservedAt:     &observedAt,
	})
	if err != nil {
		t.Fatalf("apply presence: %v", err)
	}

	after := parseBufferTime(result.Session.Metadata["buffer"].(map[string]interface{})["pending_until"])
	if after == nil || !after.Before(*before) {
		t.Fatalf("expected pending_until to be shortened")
	}
}

func TestPresenceWithoutPendingBufferIsSkipped(t *testing.T) {
	store := newFakeStore()
	svc := NewService(store, config.Config{ChatDebounceWindowMS: 1500})

	_, err := svc.Ingest(context.Background(), IngestMessageInput{
		ContactKey: "5511999999999",
		Message: IngestMessagePayload{
			Direction:         "OUTBOUND",
			ProviderMessageID: "msg-presence-3",
			IdempotencyKey:    "idem-presence-3",
			Body:              "resposta",
		},
	})
	if err != nil {
		t.Fatalf("ingest outbound message: %v", err)
	}

	result, err := svc.ApplyPresenceSignal(context.Background(), ApplyPresenceSignalInput{
		ContactKey:     "5511999999999",
		PresenceStatus: "typing",
	})
	if err != nil {
		t.Fatalf("apply presence: %v", err)
	}
	if result.Status != "skipped" {
		t.Fatalf("expected skipped status, got %s", result.Status)
	}
	if result.Reason != "no_pending_buffer" {
		t.Fatalf("expected no_pending_buffer, got %s", result.Reason)
	}
}

func TestReprocessBuildsMemoryAndMarksMessagesAutomationPending(t *testing.T) {
	store := newFakeStore()
	svc := NewService(store, config.Config{ChatDebounceWindowMS: 1500})

	first, err := svc.Ingest(context.Background(), IngestMessageInput{
		ContactKey: "5511999999999",
		Message: IngestMessagePayload{
			Direction:         "INBOUND",
			ProviderMessageID: "msg-reprocess-1",
			IdempotencyKey:    "idem-reprocess-1",
			Body:              "quero saber o valor para Santa Catarina",
		},
	})
	if err != nil {
		t.Fatalf("ingest first message: %v", err)
	}
	_, err = svc.Ingest(context.Background(), IngestMessageInput{
		ContactKey: "5511999999999",
		Message: IngestMessagePayload{
			Direction:         "INBOUND",
			ProviderMessageID: "msg-reprocess-2",
			IdempotencyKey:    "idem-reprocess-2",
			Body:              "saindo de Sao Luis",
		},
	})
	if err != nil {
		t.Fatalf("ingest second message: %v", err)
	}

	handler := NewHandler(svc)
	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodPost, "/chat/sessions/"+first.Session.ID+"/reprocess", bytes.NewBufferString(`{
		"trigger":"dashboard",
		"metadata":{"source":"ops"}
	}`))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var out ReprocessResult
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if out.Status != "accepted" {
		t.Fatalf("expected accepted status, got %s", out.Status)
	}
	if len(out.Messages) != 2 {
		t.Fatalf("expected two reprocessed messages, got %d", len(out.Messages))
	}
	for _, item := range out.Messages {
		if item.ProcessingStatus != messageStatusAutomationPending {
			t.Fatalf("expected processing status %s, got %s", messageStatusAutomationPending, item.ProcessingStatus)
		}
	}
	currentTurnBody := asString(out.Memory["current_turn_body"])
	if currentTurnBody == "" || !strings.Contains(currentTurnBody, "Santa Catarina") || !strings.Contains(currentTurnBody, "Sao Luis") {
		t.Fatalf("expected current_turn_body to include buffered messages, got %q", currentTurnBody)
	}
	buffer, ok := out.Session.Metadata["buffer"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected buffer metadata to be present")
	}
	if got := asString(buffer["status"]); got != bufferStatusIdle {
		t.Fatalf("expected buffer status %s, got %s", bufferStatusIdle, got)
	}
	agent, ok := out.Session.Metadata["agent"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected agent metadata to be present")
	}
	if got := asString(agent["status"]); got != agentStatusReadyForAutomation {
		t.Fatalf("expected agent status %s, got %s", agentStatusReadyForAutomation, got)
	}
	if got := asString(agent["trigger"]); got != "DASHBOARD" {
		t.Fatalf("expected agent trigger DASHBOARD, got %s", got)
	}
}

func TestReprocessRejectsWhenHumanOwnerIsActive(t *testing.T) {
	store := newFakeStore()
	svc := NewService(store, config.Config{ChatDebounceWindowMS: 1500})

	ingested, err := svc.Ingest(context.Background(), IngestMessageInput{
		ContactKey: "5511999999999",
		Message: IngestMessagePayload{
			Direction:         "INBOUND",
			ProviderMessageID: "msg-reprocess-human-1",
			IdempotencyKey:    "idem-reprocess-human-1",
			Body:              "oi",
		},
	})
	if err != nil {
		t.Fatalf("ingest message: %v", err)
	}

	item := store.sessions[ingested.Session.ID]
	item.HandoffStatus = "HUMAN"
	item.CurrentOwnerUserID = uuid.NewString()
	store.sessions[item.ID] = item

	handler := NewHandler(svc)
	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodPost, "/chat/sessions/"+ingested.Session.ID+"/reprocess", bytes.NewBufferString(`{}`))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d", http.StatusConflict, rec.Code)
	}
}

func TestReprocessGeneratesAgentDraftWhenRunnerIsEnabled(t *testing.T) {
	store := newFakeStore()
	svc := NewService(store, config.Config{ChatDebounceWindowMS: 1500}, &fakeAgentRunner{
		enabled: true,
		result: RunAgentResult{
			ReplyText:          "Temos saidas para Santa Catarina. Qual cidade de destino voce quer consultar?",
			Model:              "gpt-test",
			ProviderResponseID: "resp_123",
			RequestPayload:     map[string]interface{}{"model": "gpt-test"},
			ResponsePayload:    map[string]interface{}{"id": "resp_123"},
		},
	})

	ingested, err := svc.Ingest(context.Background(), IngestMessageInput{
		ContactKey: "5511999999999",
		Message: IngestMessagePayload{
			Direction:         "INBOUND",
			ProviderMessageID: "msg-agent-1",
			IdempotencyKey:    "idem-agent-1",
			Body:              "quero ir para Santa Catarina",
		},
	})
	if err != nil {
		t.Fatalf("ingest message: %v", err)
	}

	handler := NewHandler(svc)
	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodPost, "/chat/sessions/"+ingested.Session.ID+"/reprocess", bytes.NewBufferString(`{"trigger":"dashboard"}`))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var out ReprocessResult
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if out.Draft == nil {
		t.Fatalf("expected draft to be present")
	}
	if out.Draft.ProcessingStatus != messageStatusAutomationDraft {
		t.Fatalf("expected draft status %s, got %s", messageStatusAutomationDraft, out.Draft.ProcessingStatus)
	}
	agent, ok := out.Session.Metadata["agent"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected agent metadata")
	}
	if got := asString(agent["status"]); got != agentStatusDraftGenerated {
		t.Fatalf("expected agent status %s, got %s", agentStatusDraftGenerated, got)
	}
}

func TestReprocessUsesAvailabilityToolWhenTurnHasStructuredRoute(t *testing.T) {
	store := newFakeStore()
	runner := &fakeAgentRunner{
		enabled: true,
		result: RunAgentResult{
			ReplyText:          "Encontrei uma opcao para esse trecho.",
			Model:              "gpt-test",
			ProviderResponseID: "resp_tool_1",
		},
	}
	searcher := &fakeAvailabilitySearcher{
		enabled: true,
		result: AvailabilitySearchResult{
			Results: []AvailabilitySearchItem{
				{
					SegmentID:              "seg-1",
					TripID:                 "trip-1",
					RouteID:                "route-1",
					OriginDisplayName:      "Videira/SC",
					DestinationDisplayName: "Sao Luis/MA",
					OriginDepartTime:       "18:30",
					TripDate:               "2026-05-10",
					SeatsAvailable:         12,
					Price:                  250,
					Currency:               "BRL",
					Status:                 "ACTIVE",
					TripStatus:             "SCHEDULED",
					PackageName:            "Pacote p/ Maranhao",
				},
			},
		},
	}
	svc := NewService(store, config.Config{ChatDebounceWindowMS: 1500}, runner, searcher)

	ingested, err := svc.Ingest(context.Background(), IngestMessageInput{
		ContactKey: "5511999999999",
		Message: IngestMessagePayload{
			Direction:         "INBOUND",
			ProviderMessageID: "msg-tool-1",
			IdempotencyKey:    "idem-tool-1",
			Body:              "quais horarios e o valor de Videira/SC para Sao Luis/MA em 10/05 para 2 pessoas?",
		},
	})
	if err != nil {
		t.Fatalf("ingest message: %v", err)
	}

	handler := NewHandler(svc)
	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodPost, "/chat/sessions/"+ingested.Session.ID+"/reprocess", bytes.NewBufferString(`{}`))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var out ReprocessResult
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if len(out.ToolCalls) != 1 {
		t.Fatalf("expected one tool call, got %d", len(out.ToolCalls))
	}
	if out.ToolCalls[0].ToolName != toolNameAvailabilitySearch {
		t.Fatalf("expected tool name %s, got %s", toolNameAvailabilitySearch, out.ToolCalls[0].ToolName)
	}
	if searcher.calls != 1 {
		t.Fatalf("expected one availability search, got %d", searcher.calls)
	}
	if runner.calls != 1 {
		t.Fatalf("expected one runner call, got %d", runner.calls)
	}
	if searcher.lastInput.Origin != "Videira/SC" || searcher.lastInput.Destination != "Sao Luis/MA" {
		t.Fatalf("unexpected normalized route: %+v", searcher.lastInput)
	}
	if searcher.lastInput.Qty != 2 {
		t.Fatalf("expected quantity 2, got %d", searcher.lastInput.Qty)
	}
	if searcher.lastInput.TripDate == nil || searcher.lastInput.TripDate.UTC().Format("2006-01-02") != "2026-05-10" {
		t.Fatalf("expected inferred date 2026-05-10, got %+v", searcher.lastInput.TripDate)
	}
	if !strings.Contains(runner.lastInput.UserPrompt, "RESULTADO DE FERRAMENTA") {
		t.Fatalf("expected prompt to include tool result section")
	}
	if !strings.Contains(runner.lastInput.UserPrompt, "Videira/SC") || !strings.Contains(runner.lastInput.UserPrompt, "Sao Luis/MA") {
		t.Fatalf("expected prompt to include normalized route")
	}
}

func TestReprocessSkipsAvailabilityToolWhenTurnIsBroad(t *testing.T) {
	store := newFakeStore()
	runner := &fakeAgentRunner{
		enabled: true,
		result: RunAgentResult{
			ReplyText:          "Qual cidade voce quer consultar?",
			Model:              "gpt-test",
			ProviderResponseID: "resp_tool_2",
		},
	}
	searcher := &fakeAvailabilitySearcher{enabled: true}
	svc := NewService(store, config.Config{ChatDebounceWindowMS: 1500}, runner, searcher)

	ingested, err := svc.Ingest(context.Background(), IngestMessageInput{
		ContactKey: "5511999999999",
		Message: IngestMessagePayload{
			Direction:         "INBOUND",
			ProviderMessageID: "msg-tool-2",
			IdempotencyKey:    "idem-tool-2",
			Body:              "quero ir para Santa Catarina",
		},
	})
	if err != nil {
		t.Fatalf("ingest message: %v", err)
	}

	handler := NewHandler(svc)
	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodPost, "/chat/sessions/"+ingested.Session.ID+"/reprocess", bytes.NewBufferString(`{}`))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var out ReprocessResult
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if len(out.ToolCalls) != 0 {
		t.Fatalf("expected no tool calls, got %d", len(out.ToolCalls))
	}
	if searcher.calls != 0 {
		t.Fatalf("expected no availability search, got %d", searcher.calls)
	}
}

func TestReprocessReturnsBadGatewayWhenAvailabilityToolFails(t *testing.T) {
	store := newFakeStore()
	runner := &fakeAgentRunner{
		enabled: true,
		result: RunAgentResult{
			ReplyText:          "draft",
			Model:              "gpt-test",
			ProviderResponseID: "resp_tool_3",
		},
	}
	searcher := &fakeAvailabilitySearcher{
		enabled: true,
		err:     errors.New("timeout"),
	}
	svc := NewService(store, config.Config{ChatDebounceWindowMS: 1500}, runner, searcher)

	ingested, err := svc.Ingest(context.Background(), IngestMessageInput{
		ContactKey: "5511999999999",
		Message: IngestMessagePayload{
			Direction:         "INBOUND",
			ProviderMessageID: "msg-tool-3",
			IdempotencyKey:    "idem-tool-3",
			Body:              "preciso de vagas de Videira/SC para Sao Luis/MA",
		},
	})
	if err != nil {
		t.Fatalf("ingest message: %v", err)
	}

	handler := NewHandler(svc)
	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodPost, "/chat/sessions/"+ingested.Session.ID+"/reprocess", bytes.NewBufferString(`{}`))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("expected status %d, got %d", http.StatusBadGateway, rec.Code)
	}
	if runner.calls != 0 {
		t.Fatalf("expected no runner call after tool failure, got %d", runner.calls)
	}
	if len(store.toolCalls) != 1 {
		t.Fatalf("expected one stored tool call, got %d", len(store.toolCalls))
	}
}

func TestReprocessUsesPricingQuoteToolAfterAvailabilityForPriceIntent(t *testing.T) {
	store := newFakeStore()
	runner := &fakeAgentRunner{
		enabled: true,
		result: RunAgentResult{
			ReplyText:          "Encontrei os valores confirmados para esse trecho.",
			Model:              "gpt-test",
			ProviderResponseID: "resp_pricing_1",
		},
	}
	availabilitySearcher := &fakeAvailabilitySearcher{
		enabled: true,
		result: AvailabilitySearchResult{
			Results: []AvailabilitySearchItem{
				{
					SegmentID:              "seg-1",
					TripID:                 "trip-1",
					RouteID:                "route-1",
					BoardStopID:            "board-1",
					AlightStopID:           "alight-1",
					OriginStopID:           "stop-origin-1",
					DestinationStopID:      "stop-destination-1",
					OriginDisplayName:      "Videira/SC",
					DestinationDisplayName: "Sao Luis/MA",
					OriginDepartTime:       "18:30",
					TripDate:               "2026-05-10",
					SeatsAvailable:         12,
					Price:                  250,
					Currency:               "BRL",
					Status:                 "ACTIVE",
					TripStatus:             "SCHEDULED",
					PackageName:            "Pacote p/ Maranhao",
				},
			},
		},
	}
	pricingSearcher := &fakePricingQuoteSearcher{
		enabled: true,
		result: PricingQuoteResult{
			Filter: PricingQuoteInput{FareMode: "AUTO"},
			Results: []PricingQuoteItem{
				{
					TripID:                 "trip-1",
					RouteID:                "route-1",
					BoardStopID:            "board-1",
					AlightStopID:           "alight-1",
					OriginStopID:           "stop-origin-1",
					DestinationStopID:      "stop-destination-1",
					OriginDisplayName:      "Videira/SC",
					DestinationDisplayName: "Sao Luis/MA",
					OriginDepartTime:       "18:30",
					TripDate:               "2026-05-10",
					BaseAmount:             250,
					CalcAmount:             250,
					FinalAmount:            250,
					Currency:               "BRL",
					FareMode:               "AUTO",
				},
			},
		},
	}
	svc := NewService(store, config.Config{ChatDebounceWindowMS: 1500}, runner, availabilitySearcher, pricingSearcher)

	ingested, err := svc.Ingest(context.Background(), IngestMessageInput{
		ContactKey: "5511999999999",
		Message: IngestMessagePayload{
			Direction:         "INBOUND",
			ProviderMessageID: "msg-pricing-tool-1",
			IdempotencyKey:    "idem-pricing-tool-1",
			Body:              "qual o valor de Videira/SC para Sao Luis/MA em 10/05 para 2 pessoas?",
		},
	})
	if err != nil {
		t.Fatalf("ingest message: %v", err)
	}

	handler := NewHandler(svc)
	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodPost, "/chat/sessions/"+ingested.Session.ID+"/reprocess", bytes.NewBufferString(`{}`))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var out ReprocessResult
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if len(out.ToolCalls) != 2 {
		t.Fatalf("expected two tool calls, got %d", len(out.ToolCalls))
	}
	if out.ToolCalls[0].ToolName != toolNameAvailabilitySearch {
		t.Fatalf("expected first tool name %s, got %s", toolNameAvailabilitySearch, out.ToolCalls[0].ToolName)
	}
	if out.ToolCalls[1].ToolName != toolNamePricingQuote {
		t.Fatalf("expected second tool name %s, got %s", toolNamePricingQuote, out.ToolCalls[1].ToolName)
	}
	if availabilitySearcher.calls != 1 {
		t.Fatalf("expected one availability search, got %d", availabilitySearcher.calls)
	}
	if pricingSearcher.calls != 1 {
		t.Fatalf("expected one pricing quote search, got %d", pricingSearcher.calls)
	}
	if len(pricingSearcher.lastInput.Candidates) != 1 {
		t.Fatalf("expected one pricing candidate, got %d", len(pricingSearcher.lastInput.Candidates))
	}
	if pricingSearcher.lastInput.Candidates[0].TripID != "trip-1" || pricingSearcher.lastInput.Candidates[0].BoardStopID != "board-1" {
		t.Fatalf("unexpected pricing candidate: %+v", pricingSearcher.lastInput.Candidates[0])
	}
	if !strings.Contains(runner.lastInput.UserPrompt, toolNamePricingQuote) {
		t.Fatalf("expected prompt to include pricing quote section")
	}
	if !strings.Contains(runner.lastInput.UserPrompt, "final R$ 250.00 BRL") {
		t.Fatalf("expected prompt to include quoted final amount")
	}
}

func TestReprocessReturnsBadGatewayWhenPricingQuoteToolFails(t *testing.T) {
	store := newFakeStore()
	runner := &fakeAgentRunner{
		enabled: true,
		result: RunAgentResult{
			ReplyText:          "draft",
			Model:              "gpt-test",
			ProviderResponseID: "resp_pricing_2",
		},
	}
	availabilitySearcher := &fakeAvailabilitySearcher{
		enabled: true,
		result: AvailabilitySearchResult{
			Results: []AvailabilitySearchItem{
				{
					SegmentID:              "seg-1",
					TripID:                 "trip-1",
					RouteID:                "route-1",
					BoardStopID:            "board-1",
					AlightStopID:           "alight-1",
					OriginStopID:           "stop-origin-1",
					DestinationStopID:      "stop-destination-1",
					OriginDisplayName:      "Videira/SC",
					DestinationDisplayName: "Sao Luis/MA",
					OriginDepartTime:       "18:30",
					TripDate:               "2026-05-10",
					Price:                  250,
					Currency:               "BRL",
					Status:                 "ACTIVE",
					TripStatus:             "SCHEDULED",
				},
			},
		},
	}
	pricingSearcher := &fakePricingQuoteSearcher{
		enabled: true,
		err:     errors.New("pricing timeout"),
	}
	svc := NewService(store, config.Config{ChatDebounceWindowMS: 1500}, runner, availabilitySearcher, pricingSearcher)

	ingested, err := svc.Ingest(context.Background(), IngestMessageInput{
		ContactKey: "5511999999999",
		Message: IngestMessagePayload{
			Direction:         "INBOUND",
			ProviderMessageID: "msg-pricing-tool-2",
			IdempotencyKey:    "idem-pricing-tool-2",
			Body:              "qual o preco de Videira/SC para Sao Luis/MA?",
		},
	})
	if err != nil {
		t.Fatalf("ingest message: %v", err)
	}

	handler := NewHandler(svc)
	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodPost, "/chat/sessions/"+ingested.Session.ID+"/reprocess", bytes.NewBufferString(`{}`))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("expected status %d, got %d", http.StatusBadGateway, rec.Code)
	}
	if runner.calls != 0 {
		t.Fatalf("expected no runner call after pricing tool failure, got %d", runner.calls)
	}
	if len(store.toolCalls) != 2 {
		t.Fatalf("expected two stored tool calls, got %d", len(store.toolCalls))
	}
	if store.toolCalls[store.toolCallOrder[1]].ToolName != toolNamePricingQuote {
		t.Fatalf("expected second stored tool call to be %s", toolNamePricingQuote)
	}
}

func TestReprocessUsesBookingLookupToolWhenTurnHasReservationCode(t *testing.T) {
	store := newFakeStore()
	runner := &fakeAgentRunner{
		enabled: true,
		result: RunAgentResult{
			ReplyText:          "Encontrei a sua reserva e ela segue pendente.",
			Model:              "gpt-test",
			ProviderResponseID: "resp_booking_1",
		},
	}
	searcher := &fakeBookingLookupSearcher{
		enabled: true,
		result: BookingLookupResult{
			Results: []BookingLookupItem{
				{
					ID:              "BK-ABC123456",
					TripID:          "trip-1",
					Status:          "PENDING",
					ReservationCode: "ABC12345",
					TotalAmount:     950,
					DepositAmount:   300,
					RemainderAmount: 650,
					PassengerName:   "Maria",
					PassengerPhone:  "48999999999",
					SeatNumber:      12,
					CreatedAt:       time.Now().UTC(),
				},
			},
		},
	}
	svc := NewService(store, config.Config{ChatDebounceWindowMS: 1500}, runner, searcher)

	ingested, err := svc.Ingest(context.Background(), IngestMessageInput{
		ContactKey: "5511999999999",
		Message: IngestMessagePayload{
			Direction:         "INBOUND",
			ProviderMessageID: "msg-booking-tool-1",
			IdempotencyKey:    "idem-booking-tool-1",
			Body:              "qual o status da minha reserva ABC12345?",
		},
	})
	if err != nil {
		t.Fatalf("ingest message: %v", err)
	}

	handler := NewHandler(svc)
	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodPost, "/chat/sessions/"+ingested.Session.ID+"/reprocess", bytes.NewBufferString(`{}`))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var out ReprocessResult
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if len(out.ToolCalls) != 1 {
		t.Fatalf("expected one tool call, got %d", len(out.ToolCalls))
	}
	if out.ToolCalls[0].ToolName != toolNameBookingLookup {
		t.Fatalf("expected tool name %s, got %s", toolNameBookingLookup, out.ToolCalls[0].ToolName)
	}
	if searcher.calls != 1 {
		t.Fatalf("expected one booking lookup, got %d", searcher.calls)
	}
	if searcher.lastInput.ReservationCode != "ABC12345" {
		t.Fatalf("expected reservation_code ABC12345, got %s", searcher.lastInput.ReservationCode)
	}
	if !strings.Contains(runner.lastInput.UserPrompt, "codigo ABC12345") && !strings.Contains(runner.lastInput.UserPrompt, "ABC12345") {
		t.Fatalf("expected prompt to include reservation code")
	}
	if !strings.Contains(runner.lastInput.UserPrompt, "PENDING") {
		t.Fatalf("expected prompt to include booking status")
	}
}

func TestReprocessReturnsBadGatewayWhenBookingLookupToolFails(t *testing.T) {
	store := newFakeStore()
	runner := &fakeAgentRunner{
		enabled: true,
		result: RunAgentResult{
			ReplyText:          "draft",
			Model:              "gpt-test",
			ProviderResponseID: "resp_booking_2",
		},
	}
	searcher := &fakeBookingLookupSearcher{
		enabled: true,
		err:     errors.New("db timeout"),
	}
	svc := NewService(store, config.Config{ChatDebounceWindowMS: 1500}, runner, searcher)

	ingested, err := svc.Ingest(context.Background(), IngestMessageInput{
		ContactKey: "5511999999999",
		Message: IngestMessagePayload{
			Direction:         "INBOUND",
			ProviderMessageID: "msg-booking-tool-2",
			IdempotencyKey:    "idem-booking-tool-2",
			Body:              "me diz o status da reserva ABC12345",
		},
	})
	if err != nil {
		t.Fatalf("ingest message: %v", err)
	}

	handler := NewHandler(svc)
	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodPost, "/chat/sessions/"+ingested.Session.ID+"/reprocess", bytes.NewBufferString(`{}`))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("expected status %d, got %d", http.StatusBadGateway, rec.Code)
	}
	if runner.calls != 0 {
		t.Fatalf("expected no runner call after booking tool failure, got %d", runner.calls)
	}
	if len(store.toolCalls) != 1 {
		t.Fatalf("expected one stored tool call, got %d", len(store.toolCalls))
	}
}

func TestReprocessUsesPaymentStatusToolWhenTurnAsksAboutPix(t *testing.T) {
	store := newFakeStore()
	runner := &fakeAgentRunner{
		enabled: true,
		result: RunAgentResult{
			ReplyText:          "Vi aqui que o PIX ainda esta pendente.",
			Model:              "gpt-test",
			ProviderResponseID: "resp_payment_1",
		},
	}
	bookingSearcher := &fakeBookingLookupSearcher{
		enabled: true,
		result: BookingLookupResult{
			Results: []BookingLookupItem{
				{
					ID:              "BK-ABC123456",
					TripID:          "trip-1",
					Status:          "PENDING",
					ReservationCode: "ABC12345",
					TotalAmount:     950,
					DepositAmount:   300,
					RemainderAmount: 650,
					PassengerName:   "Maria",
					PassengerPhone:  "48999999999",
					SeatNumber:      12,
					CreatedAt:       time.Now().UTC(),
				},
			},
		},
	}
	paymentSearcher := &fakePaymentStatusSearcher{
		enabled: true,
		result: PaymentStatusResult{
			Results: []PaymentStatusItem{
				{
					ID:          "pay-1",
					BookingID:   "BK-ABC123456",
					Amount:      300,
					Method:      "PIX",
					Status:      "PENDING",
					Provider:    "PAGARME",
					ProviderRef: "charge-1",
					CreatedAt:   time.Now().UTC(),
				},
			},
		},
	}
	svc := NewService(store, config.Config{ChatDebounceWindowMS: 1500}, runner, bookingSearcher, paymentSearcher)

	ingested, err := svc.Ingest(context.Background(), IngestMessageInput{
		ContactKey: "5511999999999",
		Message: IngestMessagePayload{
			Direction:         "INBOUND",
			ProviderMessageID: "msg-payment-tool-1",
			IdempotencyKey:    "idem-payment-tool-1",
			Body:              "o pix da reserva ABC12345 ja foi pago?",
		},
	})
	if err != nil {
		t.Fatalf("ingest message: %v", err)
	}

	handler := NewHandler(svc)
	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodPost, "/chat/sessions/"+ingested.Session.ID+"/reprocess", bytes.NewBufferString(`{}`))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var out ReprocessResult
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if len(out.ToolCalls) != 2 {
		t.Fatalf("expected two tool calls, got %d", len(out.ToolCalls))
	}
	if out.ToolCalls[0].ToolName != toolNameBookingLookup {
		t.Fatalf("expected first tool name %s, got %s", toolNameBookingLookup, out.ToolCalls[0].ToolName)
	}
	if out.ToolCalls[1].ToolName != toolNamePaymentStatus {
		t.Fatalf("expected second tool name %s, got %s", toolNamePaymentStatus, out.ToolCalls[1].ToolName)
	}
	if bookingSearcher.calls != 1 {
		t.Fatalf("expected one booking lookup, got %d", bookingSearcher.calls)
	}
	if paymentSearcher.calls != 1 {
		t.Fatalf("expected one payment lookup, got %d", paymentSearcher.calls)
	}
	if paymentSearcher.lastInput.BookingID != "BK-ABC123456" {
		t.Fatalf("expected booking_id BK-ABC123456, got %s", paymentSearcher.lastInput.BookingID)
	}
	if !strings.Contains(runner.lastInput.UserPrompt, toolNamePaymentStatus) {
		t.Fatalf("expected prompt to include payment tool section")
	}
	if !strings.Contains(runner.lastInput.UserPrompt, "Pagamento 1:") {
		t.Fatalf("expected prompt to include payment item")
	}
	if !strings.Contains(runner.lastInput.UserPrompt, "PENDING") {
		t.Fatalf("expected prompt to include payment status")
	}
}

func TestReprocessReturnsBadGatewayWhenPaymentStatusToolFails(t *testing.T) {
	store := newFakeStore()
	runner := &fakeAgentRunner{
		enabled: true,
		result: RunAgentResult{
			ReplyText:          "draft",
			Model:              "gpt-test",
			ProviderResponseID: "resp_payment_2",
		},
	}
	bookingSearcher := &fakeBookingLookupSearcher{
		enabled: true,
		result: BookingLookupResult{
			Results: []BookingLookupItem{
				{
					ID:              "BK-ABC123456",
					Status:          "PENDING",
					ReservationCode: "ABC12345",
					CreatedAt:       time.Now().UTC(),
				},
			},
		},
	}
	paymentSearcher := &fakePaymentStatusSearcher{
		enabled: true,
		err:     errors.New("payments timeout"),
	}
	svc := NewService(store, config.Config{ChatDebounceWindowMS: 1500}, runner, bookingSearcher, paymentSearcher)

	ingested, err := svc.Ingest(context.Background(), IngestMessageInput{
		ContactKey: "5511999999999",
		Message: IngestMessagePayload{
			Direction:         "INBOUND",
			ProviderMessageID: "msg-payment-tool-2",
			IdempotencyKey:    "idem-payment-tool-2",
			Body:              "confere o pagamento da reserva ABC12345",
		},
	})
	if err != nil {
		t.Fatalf("ingest message: %v", err)
	}

	handler := NewHandler(svc)
	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodPost, "/chat/sessions/"+ingested.Session.ID+"/reprocess", bytes.NewBufferString(`{}`))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("expected status %d, got %d", http.StatusBadGateway, rec.Code)
	}
	if runner.calls != 0 {
		t.Fatalf("expected no runner call after payment tool failure, got %d", runner.calls)
	}
	if len(store.toolCalls) != 2 {
		t.Fatalf("expected two stored tool calls, got %d", len(store.toolCalls))
	}
	if store.toolCalls[store.toolCallOrder[1]].ToolName != toolNamePaymentStatus {
		t.Fatalf("expected second stored tool call to be %s", toolNamePaymentStatus)
	}
}

func TestReprocessDraftIsIdempotentPerTurn(t *testing.T) {
	store := newFakeStore()
	runner := &fakeAgentRunner{
		enabled: true,
		result: RunAgentResult{
			ReplyText:          "Qual cidade de destino voce quer consultar?",
			Model:              "gpt-test",
			ProviderResponseID: "resp_456",
			RequestPayload:     map[string]interface{}{"model": "gpt-test"},
			ResponsePayload:    map[string]interface{}{"id": "resp_456"},
		},
	}
	svc := NewService(store, config.Config{ChatDebounceWindowMS: 1500}, runner)

	ingested, err := svc.Ingest(context.Background(), IngestMessageInput{
		ContactKey: "5511999999999",
		Message: IngestMessagePayload{
			Direction:         "INBOUND",
			ProviderMessageID: "msg-agent-2",
			IdempotencyKey:    "idem-agent-2",
			Body:              "quais datas para Santa Catarina?",
		},
	})
	if err != nil {
		t.Fatalf("ingest message: %v", err)
	}

	handler := NewHandler(svc)
	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	path := "/chat/sessions/" + ingested.Session.ID + "/reprocess"
	firstReq := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(`{}`))
	firstRec := httptest.NewRecorder()
	r.ServeHTTP(firstRec, firstReq)
	if firstRec.Code != http.StatusOK {
		t.Fatalf("expected first status %d, got %d", http.StatusOK, firstRec.Code)
	}

	secondReq := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(`{}`))
	secondRec := httptest.NewRecorder()
	r.ServeHTTP(secondRec, secondReq)
	if secondRec.Code != http.StatusOK {
		t.Fatalf("expected second status %d, got %d", http.StatusOK, secondRec.Code)
	}

	var out ReprocessResult
	if err := json.Unmarshal(secondRec.Body.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if !out.Idempotent {
		t.Fatalf("expected idempotent reprocess result")
	}
	if runner.calls != 1 {
		t.Fatalf("expected one runner call, got %d", runner.calls)
	}
}

func TestReprocessReturnsBadGatewayWhenAgentRunnerFails(t *testing.T) {
	store := newFakeStore()
	svc := NewService(store, config.Config{ChatDebounceWindowMS: 1500}, &fakeAgentRunner{
		enabled: true,
		err:     errors.New("upstream timeout"),
	})

	ingested, err := svc.Ingest(context.Background(), IngestMessageInput{
		ContactKey: "5511999999999",
		Message: IngestMessagePayload{
			Direction:         "INBOUND",
			ProviderMessageID: "msg-agent-3",
			IdempotencyKey:    "idem-agent-3",
			Body:              "preciso saber os horarios",
		},
	})
	if err != nil {
		t.Fatalf("ingest message: %v", err)
	}

	handler := NewHandler(svc)
	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodPost, "/chat/sessions/"+ingested.Session.ID+"/reprocess", bytes.NewBufferString(`{}`))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("expected status %d, got %d", http.StatusBadGateway, rec.Code)
	}
}

func TestReprocessRejectsWhenNoPendingMessagesExist(t *testing.T) {
	store := newFakeStore()
	session, _ := store.seedSessionWithMessage("5511999999999", "oi")
	_, err := store.CreateMessage(context.Background(), CreateMessageInput{
		SessionID:        session.ID,
		Direction:        "OUTBOUND",
		Kind:             "TEXT",
		ProcessingStatus: "SENT",
		ReceivedAt:       time.Now().UTC(),
		Body:             "resposta",
	})
	if err != nil {
		t.Fatalf("seed outbound message: %v", err)
	}

	handler := NewHandler(NewService(store, config.Config{ChatDebounceWindowMS: 1500}))
	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodPost, "/chat/sessions/"+session.ID+"/reprocess", bytes.NewBufferString(`{}`))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d", http.StatusConflict, rec.Code)
	}
}

type fakeStore struct {
	sessions         map[string]Session
	sessionsByKey    map[string]string
	messages         map[string]Message
	messageOrder     []string
	byProviderID     map[string]string
	byIdempotencyKey map[string]string
	handoffs         map[string]Handoff
	outbounds        map[string]ReplyOutbound
	toolCalls        map[string]ToolCall
	toolCallOrder    []string
}

type fakeReplySender struct {
	enabled bool
	calls   int
	result  SendReplyResult
	results []SendReplyResult
	err     error
	errs    []error
}

type fakeAgentRunner struct {
	enabled   bool
	calls     int
	result    RunAgentResult
	results   []RunAgentResult
	err       error
	errs      []error
	lastInput RunAgentInput
}

type fakeAvailabilitySearcher struct {
	enabled   bool
	calls     int
	result    AvailabilitySearchResult
	results   []AvailabilitySearchResult
	err       error
	errs      []error
	lastInput AvailabilitySearchInput
}

type fakePricingQuoteSearcher struct {
	enabled   bool
	calls     int
	result    PricingQuoteResult
	results   []PricingQuoteResult
	err       error
	errs      []error
	lastInput PricingQuoteInput
}

type fakeBookingLookupSearcher struct {
	enabled   bool
	calls     int
	result    BookingLookupResult
	results   []BookingLookupResult
	err       error
	errs      []error
	lastInput BookingLookupInput
}

type fakePaymentStatusSearcher struct {
	enabled   bool
	calls     int
	result    PaymentStatusResult
	results   []PaymentStatusResult
	err       error
	errs      []error
	lastInput PaymentStatusInput
}

func newFakeStore() *fakeStore {
	return &fakeStore{
		sessions:         map[string]Session{},
		sessionsByKey:    map[string]string{},
		messages:         map[string]Message{},
		messageOrder:     []string{},
		byProviderID:     map[string]string{},
		byIdempotencyKey: map[string]string{},
		handoffs:         map[string]Handoff{},
		outbounds:        map[string]ReplyOutbound{},
		toolCalls:        map[string]ToolCall{},
		toolCallOrder:    []string{},
	}
}

func (f *fakeReplySender) Enabled() bool {
	return f != nil && f.enabled
}

func (f *fakeReplySender) SendReply(_ context.Context, _ SendReplyInput) (SendReplyResult, error) {
	f.calls++
	if len(f.errs) > 0 {
		err := f.errs[0]
		f.errs = f.errs[1:]
		var result SendReplyResult
		if len(f.results) > 0 {
			result = f.results[0]
			f.results = f.results[1:]
		}
		if err != nil {
			return result, err
		}
		return result, nil
	}
	if f.err != nil {
		return SendReplyResult{}, f.err
	}
	return f.result, nil
}

func (f *fakeAgentRunner) Enabled() bool {
	return f != nil && f.enabled
}

func (f *fakeAgentRunner) Run(_ context.Context, input RunAgentInput) (RunAgentResult, error) {
	f.calls++
	f.lastInput = input
	if len(f.errs) > 0 {
		err := f.errs[0]
		f.errs = f.errs[1:]
		var result RunAgentResult
		if len(f.results) > 0 {
			result = f.results[0]
			f.results = f.results[1:]
		}
		if err != nil {
			return result, err
		}
		return result, nil
	}
	if f.err != nil {
		return RunAgentResult{}, f.err
	}
	return f.result, nil
}

func (f *fakeAvailabilitySearcher) Enabled() bool {
	return f != nil && f.enabled
}

func (f *fakeAvailabilitySearcher) Search(_ context.Context, input AvailabilitySearchInput) (AvailabilitySearchResult, error) {
	f.calls++
	f.lastInput = input
	if len(f.errs) > 0 {
		err := f.errs[0]
		f.errs = f.errs[1:]
		var result AvailabilitySearchResult
		if len(f.results) > 0 {
			result = f.results[0]
			f.results = f.results[1:]
		}
		if err != nil {
			return result, err
		}
		return result, nil
	}
	if f.err != nil {
		return AvailabilitySearchResult{}, f.err
	}
	return f.result, nil
}

func (f *fakePricingQuoteSearcher) Enabled() bool {
	return f != nil && f.enabled
}

func (f *fakePricingQuoteSearcher) Search(_ context.Context, input PricingQuoteInput) (PricingQuoteResult, error) {
	f.calls++
	f.lastInput = input
	if len(f.errs) > 0 {
		err := f.errs[0]
		f.errs = f.errs[1:]
		var result PricingQuoteResult
		if len(f.results) > 0 {
			result = f.results[0]
			f.results = f.results[1:]
		}
		if err != nil {
			return result, err
		}
		return result, nil
	}
	if f.err != nil {
		return PricingQuoteResult{}, f.err
	}
	return f.result, nil
}

func (f *fakeBookingLookupSearcher) Enabled() bool {
	return f != nil && f.enabled
}

func (f *fakeBookingLookupSearcher) Search(_ context.Context, input BookingLookupInput) (BookingLookupResult, error) {
	f.calls++
	f.lastInput = input
	if len(f.errs) > 0 {
		err := f.errs[0]
		f.errs = f.errs[1:]
		var result BookingLookupResult
		if len(f.results) > 0 {
			result = f.results[0]
			f.results = f.results[1:]
		}
		if err != nil {
			return result, err
		}
		return result, nil
	}
	if f.err != nil {
		return BookingLookupResult{}, f.err
	}
	return f.result, nil
}

func (f *fakePaymentStatusSearcher) Enabled() bool {
	return f != nil && f.enabled
}

func (f *fakePaymentStatusSearcher) Search(_ context.Context, input PaymentStatusInput) (PaymentStatusResult, error) {
	f.calls++
	f.lastInput = input
	if len(f.errs) > 0 {
		err := f.errs[0]
		f.errs = f.errs[1:]
		var result PaymentStatusResult
		if len(f.results) > 0 {
			result = f.results[0]
			f.results = f.results[1:]
		}
		if err != nil {
			return result, err
		}
		return result, nil
	}
	if f.err != nil {
		return PaymentStatusResult{}, f.err
	}
	return f.result, nil
}

func (s *fakeStore) FindMessageByKeys(_ context.Context, providerMessageID string, idempotencyKey string) (*Message, error) {
	if providerMessageID != "" {
		if id, ok := s.byProviderID[providerMessageID]; ok {
			item := s.messages[id]
			return &item, nil
		}
	}
	if idempotencyKey != "" {
		if id, ok := s.byIdempotencyKey[idempotencyKey]; ok {
			item := s.messages[id]
			return &item, nil
		}
	}
	return nil, nil
}

func (s *fakeStore) UpsertSession(_ context.Context, input UpsertSessionInput) (Session, error) {
	key := input.Channel + "::" + input.ContactKey
	if existingID, ok := s.sessionsByKey[key]; ok {
		item := s.sessions[existingID]
		if input.CustomerPhone != "" {
			item.CustomerPhone = input.CustomerPhone
		}
		if input.CustomerName != "" {
			item.CustomerName = input.CustomerName
		}
		if input.LastMessageAt != nil {
			item.LastMessageAt = input.LastMessageAt
		}
		if input.LastInboundAt != nil {
			item.LastInboundAt = input.LastInboundAt
		}
		if input.LastOutboundAt != nil {
			item.LastOutboundAt = input.LastOutboundAt
		}
		if len(input.Metadata) > 0 {
			if item.Metadata == nil {
				item.Metadata = map[string]interface{}{}
			}
			for k, v := range input.Metadata {
				item.Metadata[k] = v
			}
		}
		item.UpdatedAt = time.Now().UTC()
		s.sessions[item.ID] = item
		return item, nil
	}

	now := time.Now().UTC()
	item := Session{
		ID:             uuid.NewString(),
		Channel:        input.Channel,
		ContactKey:     input.ContactKey,
		CustomerPhone:  input.CustomerPhone,
		CustomerName:   input.CustomerName,
		Status:         "ACTIVE",
		HandoffStatus:  "BOT",
		LastMessageAt:  input.LastMessageAt,
		LastInboundAt:  input.LastInboundAt,
		LastOutboundAt: input.LastOutboundAt,
		Metadata:       map[string]interface{}{},
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	for k, v := range input.Metadata {
		item.Metadata[k] = v
	}

	s.sessions[item.ID] = item
	s.sessionsByKey[key] = item.ID
	return item, nil
}

func (s *fakeStore) CreateMessage(_ context.Context, input CreateMessageInput) (Message, error) {
	now := time.Now().UTC()
	item := Message{
		ID:                uuid.NewString(),
		SessionID:         input.SessionID,
		Direction:         input.Direction,
		Kind:              input.Kind,
		ProviderMessageID: input.ProviderMessageID,
		IdempotencyKey:    input.IdempotencyKey,
		SenderName:        input.SenderName,
		SenderPhone:       input.SenderPhone,
		Body:              input.Body,
		Payload:           input.Payload,
		NormalizedPayload: input.NormalizedPayload,
		ProcessingStatus:  input.ProcessingStatus,
		ReceivedAt:        input.ReceivedAt,
		SentAt:            input.SentAt,
		CreatedAt:         now,
	}

	s.messages[item.ID] = item
	s.messageOrder = append(s.messageOrder, item.ID)
	if item.ProviderMessageID != "" {
		s.byProviderID[item.ProviderMessageID] = item.ID
	}
	if item.IdempotencyKey != "" {
		s.byIdempotencyKey[item.IdempotencyKey] = item.ID
	}

	return item, nil
}

func (s *fakeStore) CreateToolCall(_ context.Context, input CreateToolCallInput) (ToolCall, error) {
	now := time.Now().UTC()
	item := ToolCall{
		ID:              uuid.NewString(),
		SessionID:       input.SessionID,
		MessageID:       input.MessageID,
		ToolName:        input.ToolName,
		RequestPayload:  input.RequestPayload,
		ResponsePayload: input.ResponsePayload,
		Status:          input.Status,
		ErrorCode:       input.ErrorCode,
		ErrorMessage:    input.ErrorMessage,
		StartedAt:       input.StartedAt,
		FinishedAt:      input.FinishedAt,
		CreatedAt:       now,
	}
	s.toolCalls[item.ID] = item
	s.toolCallOrder = append(s.toolCallOrder, item.ID)
	return item, nil
}

func (s *fakeStore) UpdateSessionBufferState(_ context.Context, input UpdateSessionBufferStateInput) (Session, error) {
	item := s.sessions[input.SessionID]
	if item.Metadata == nil {
		item.Metadata = map[string]interface{}{}
	}
	item.Metadata["buffer"] = input.Buffer
	item.UpdatedAt = time.Now().UTC()
	s.sessions[item.ID] = item
	return item, nil
}

func (s *fakeStore) RequestHandoff(_ context.Context, input RequestHandoffInput) (RequestHandoffResult, error) {
	item, ok := s.sessions[input.SessionID]
	if !ok {
		return RequestHandoffResult{}, ErrSessionNotFound
	}

	now := time.Now().UTC()
	handoff := Handoff{
		ID:             uuid.NewString(),
		SessionID:      input.SessionID,
		RequestedBy:    input.RequestedBy,
		Reason:         input.Reason,
		Status:         "REQUESTED",
		AssignedUserID: input.AssignedUserID,
		RequestedAt:    now,
		Metadata:       map[string]interface{}{},
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	for key, value := range input.Metadata {
		handoff.Metadata[key] = value
	}
	s.handoffs[handoff.ID] = handoff

	item.HandoffStatus = "HUMAN_REQUESTED"
	item.CurrentOwnerUserID = ""
	if input.AssignedUserID != "" {
		item.HandoffStatus = "HUMAN"
		item.CurrentOwnerUserID = input.AssignedUserID
	}
	item.UpdatedAt = now
	s.sessions[item.ID] = item

	return RequestHandoffResult{
		Session: item,
		Handoff: handoff,
	}, nil
}

func (s *fakeStore) ResumeSession(_ context.Context, input ResumeSessionInput) (ResumeSessionResult, error) {
	item, ok := s.sessions[input.SessionID]
	if !ok {
		return ResumeSessionResult{}, ErrSessionNotFound
	}

	var selected Handoff
	found := false
	for _, handoff := range s.handoffs {
		if handoff.SessionID == input.SessionID && handoff.Status == "REQUESTED" && handoff.ResolvedAt == nil {
			if !found || handoff.RequestedAt.After(selected.RequestedAt) {
				selected = handoff
				found = true
			}
		}
	}
	if !found {
		return ResumeSessionResult{}, ErrNoActiveHandoff
	}

	now := time.Now().UTC()
	if selected.Metadata == nil {
		selected.Metadata = map[string]interface{}{}
	}
	for key, value := range input.Metadata {
		selected.Metadata[key] = value
	}
	if input.ResumedBy != "" {
		selected.Metadata["resumed_by"] = input.ResumedBy
	}
	if input.Reason != "" {
		selected.Metadata["resume_reason"] = input.Reason
	}
	selected.Status = "RESOLVED"
	selected.ResolvedAt = &now
	selected.UpdatedAt = now
	s.handoffs[selected.ID] = selected

	item.HandoffStatus = "BOT"
	item.CurrentOwnerUserID = ""
	item.UpdatedAt = now
	s.sessions[item.ID] = item

	return ResumeSessionResult{
		Session: item,
		Handoff: selected,
	}, nil
}

func (s *fakeStore) FindReplyByIdempotency(_ context.Context, sessionID string, idempotencyKey string) (*ReplyResult, error) {
	messageID, ok := s.byIdempotencyKey[idempotencyKey]
	if !ok {
		return nil, nil
	}
	message := s.messages[messageID]
	if message.SessionID != sessionID || message.Direction != "OUTBOUND" {
		return nil, nil
	}
	session := s.sessions[sessionID]
	for _, outbound := range s.outbounds {
		if outbound.SessionID == sessionID && outbound.IdempotencyKey == idempotencyKey {
			result := ReplyResult{
				Session:  session,
				Message:  message,
				Outbound: outbound,
			}
			return &result, nil
		}
	}
	return nil, nil
}

func (s *fakeStore) CreateReply(_ context.Context, input ReplyInput, debounceWindow time.Duration) (ReplyResult, error) {
	session, ok := s.sessions[input.SessionID]
	if !ok {
		return ReplyResult{}, ErrSessionNotFound
	}

	now := time.Now().UTC()
	replyBody := strings.TrimSpace(input.Body)
	replyMode := "ASSISTED_REPLY"
	reviewAction := ""
	var reviewedDraft *Message
	if input.DraftMessageID != "" {
		draft, ok := s.messages[input.DraftMessageID]
		if !ok || draft.SessionID != input.SessionID || draft.Direction != "OUTBOUND" || !strings.EqualFold(draft.ProcessingStatus, messageStatusAutomationDraft) {
			return ReplyResult{}, ErrReplyDraftNotAllowed
		}
		if replyBody == "" {
			replyBody = strings.TrimSpace(draft.Body)
		}
		if replyBody == "" {
			return ReplyResult{}, ErrReplyBodyRequired
		}
		reviewAction = "APPROVED_AS_IS"
		if strings.TrimSpace(replyBody) != strings.TrimSpace(draft.Body) {
			reviewAction = "EDITED"
		}
		if draft.Payload == nil {
			draft.Payload = map[string]interface{}{}
		}
		if draft.NormalizedPayload == nil {
			draft.NormalizedPayload = map[string]interface{}{}
		}
		draft.Payload["review_mode"] = "CONTROLLED"
		draft.Payload["review_status"] = "APPROVED"
		draft.Payload["review_action"] = reviewAction
		draft.Payload["reviewed_at"] = now.Format(time.RFC3339Nano)
		draft.Payload["reviewed_by_user_id"] = input.OwnerUserID
		draft.NormalizedPayload["review_mode"] = "CONTROLLED"
		draft.NormalizedPayload["review_status"] = "APPROVED"
		draft.NormalizedPayload["review_action"] = reviewAction
		draft.NormalizedPayload["reviewed_at"] = now.Format(time.RFC3339Nano)
		draft.NormalizedPayload["reviewed_by_user_id"] = input.OwnerUserID
		draft.ProcessingStatus = messageStatusAutomationReviewed
		s.messages[draft.ID] = draft
		reviewedDraft = &draft
		replyMode = "DRAFT_REVIEW"
	}

	message := Message{
		ID:             uuid.NewString(),
		SessionID:      input.SessionID,
		Direction:      "OUTBOUND",
		Kind:           "TEXT",
		IdempotencyKey: input.IdempotencyKey,
		SenderName:     input.SenderName,
		Body:           replyBody,
		Payload: map[string]interface{}{
			"mode":          replyMode,
			"shadow_mode":   true,
			"owner_user_id": input.OwnerUserID,
			"sender_name":   input.SenderName,
		},
		NormalizedPayload: map[string]interface{}{
			"mode":          replyMode,
			"shadow_mode":   true,
			"owner_user_id": input.OwnerUserID,
			"sender_name":   input.SenderName,
		},
		ProcessingStatus: "MANUAL_PENDING",
		ReceivedAt:       now,
		CreatedAt:        now,
	}
	if reviewedDraft != nil {
		message.Payload["draft_message_id"] = reviewedDraft.ID
		message.Payload["review_mode"] = "CONTROLLED"
		message.Payload["review_action"] = reviewAction
		message.Payload["draft_reviewed"] = true
		message.NormalizedPayload["draft_message_id"] = reviewedDraft.ID
		message.NormalizedPayload["review_mode"] = "CONTROLLED"
		message.NormalizedPayload["review_action"] = reviewAction
		message.NormalizedPayload["draft_reviewed"] = true
	}
	for key, value := range input.Metadata {
		message.Payload[key] = value
		message.NormalizedPayload[key] = value
	}
	s.messages[message.ID] = message
	s.messageOrder = append(s.messageOrder, message.ID)
	s.byIdempotencyKey[message.IdempotencyKey] = message.ID

	if session.Metadata == nil {
		session.Metadata = map[string]interface{}{}
	}
	session.LastMessageAt = &now
	session.LastOutboundAt = &now
	session.Metadata["buffer"] = buildBufferState(session.Metadata, message, debounceWindow)
	if reviewedDraft != nil {
		session.Metadata["agent"] = buildDraftReviewedAgentState(session.Metadata, *reviewedDraft, input.OwnerUserID, reviewAction, now)
	}
	session.UpdatedAt = now
	s.sessions[session.ID] = session

	outbound := ReplyOutbound{
		ID:             uuid.NewString(),
		SessionID:      session.ID,
		Channel:        session.Channel,
		Recipient:      session.ContactKey,
		Payload:        map[string]interface{}{"body": replyBody, "owner_user_id": input.OwnerUserID, "mode": replyMode, "shadow_mode": true, "sender_name": input.SenderName, "message_id": message.ID},
		Provider:       "EVOLUTION",
		IdempotencyKey: input.IdempotencyKey,
		Status:         "MANUAL_PENDING",
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if reviewedDraft != nil {
		outbound.Payload["draft_message_id"] = reviewedDraft.ID
		outbound.Payload["review_mode"] = "CONTROLLED"
		outbound.Payload["review_action"] = reviewAction
		outbound.Payload["draft_reviewed"] = true
	}
	for key, value := range input.Metadata {
		outbound.Payload[key] = value
	}
	s.outbounds[outbound.ID] = outbound

	return ReplyResult{
		Session:  session,
		Message:  message,
		Outbound: outbound,
		Draft:    reviewedDraft,
	}, nil
}

func (s *fakeStore) MarkReplyDeliverySent(_ context.Context, input MarkReplyDeliverySentInput) (ReplyResult, error) {
	session := s.sessions[input.SessionID]
	message := s.messages[input.MessageID]
	outbound := s.outbounds[input.OutboundID]

	if message.NormalizedPayload == nil {
		message.NormalizedPayload = map[string]interface{}{}
	}
	for key, value := range input.Payload {
		message.NormalizedPayload[key] = value
	}
	message.ProviderMessageID = input.ProviderMessageID
	message.ProcessingStatus = input.ProviderStatus
	message.SentAt = timePointer(input.SentAt)
	s.messages[message.ID] = message
	if input.ProviderMessageID != "" {
		s.byProviderID[input.ProviderMessageID] = message.ID
	}

	if outbound.Payload == nil {
		outbound.Payload = map[string]interface{}{}
	}
	for key, value := range input.Payload {
		outbound.Payload[key] = value
	}
	outbound.ProviderMessageID = input.ProviderMessageID
	outbound.Status = input.ProviderStatus
	outbound.SentAt = timePointer(input.SentAt)
	outbound.UpdatedAt = time.Now().UTC()
	s.outbounds[outbound.ID] = outbound

	return ReplyResult{
		Session:  session,
		Message:  message,
		Outbound: outbound,
	}, nil
}

func (s *fakeStore) MarkReplyDeliveryFailure(_ context.Context, input MarkReplyDeliveryFailureInput) (ReplyResult, error) {
	session := s.sessions[input.SessionID]
	message := s.messages[input.MessageID]
	outbound := s.outbounds[input.OutboundID]

	if message.NormalizedPayload == nil {
		message.NormalizedPayload = map[string]interface{}{}
	}
	message.NormalizedPayload["delivery_error_text"] = input.ErrorText
	message.ProcessingStatus = "SEND_FAILED"
	s.messages[message.ID] = message

	if outbound.Payload == nil {
		outbound.Payload = map[string]interface{}{}
	}
	outbound.Payload["delivery_error_text"] = input.ErrorText
	outbound.Status = "SEND_FAILED"
	outbound.UpdatedAt = time.Now().UTC()
	s.outbounds[outbound.ID] = outbound

	return ReplyResult{
		Session:  session,
		Message:  message,
		Outbound: outbound,
	}, nil
}

func (s *fakeStore) SaveReprocessSnapshot(_ context.Context, input SaveReprocessSnapshotInput) (SaveReprocessSnapshotResult, error) {
	session := s.sessions[input.SessionID]
	if session.Metadata == nil {
		session.Metadata = map[string]interface{}{}
	}
	session.Metadata["memory"] = input.Memory
	session.Metadata["agent"] = input.Agent
	session.Metadata["buffer"] = input.Buffer
	session.UpdatedAt = time.Now().UTC()
	s.sessions[session.ID] = session

	updated := make([]Message, 0, len(input.MessageIDs))
	for _, messageID := range input.MessageIDs {
		message := s.messages[messageID]
		if message.NormalizedPayload == nil {
			message.NormalizedPayload = map[string]interface{}{}
		}
		for key, value := range input.MessageMetadata {
			message.NormalizedPayload[key] = value
		}
		message.ProcessingStatus = input.MessageStatus
		s.messages[message.ID] = message
		updated = append(updated, message)
	}

	return SaveReprocessSnapshotResult{
		Session:  session,
		Messages: updated,
	}, nil
}

func (s *fakeStore) SaveAgentDraft(_ context.Context, input SaveAgentDraftInput) (SaveAgentDraftResult, error) {
	session := s.sessions[input.SessionID]
	if session.Metadata == nil {
		session.Metadata = map[string]interface{}{}
	}
	session.Metadata["agent"] = input.Agent
	session.Metadata["buffer"] = input.Buffer
	session.UpdatedAt = time.Now().UTC()
	s.sessions[session.ID] = session

	message := Message{
		ID:                uuid.NewString(),
		SessionID:         input.SessionID,
		Direction:         "OUTBOUND",
		Kind:              "TEXT",
		IdempotencyKey:    input.IdempotencyKey,
		SenderName:        input.SenderName,
		Body:              input.Body,
		Payload:           input.Payload,
		NormalizedPayload: input.NormalizedPayload,
		ProcessingStatus:  input.ProcessingStatus,
		ReceivedAt:        input.RecordedAt,
		CreatedAt:         input.RecordedAt,
	}
	s.messages[message.ID] = message
	s.messageOrder = append(s.messageOrder, message.ID)
	s.byIdempotencyKey[message.IdempotencyKey] = message.ID

	return SaveAgentDraftResult{
		Session: session,
		Message: message,
	}, nil
}

func (s *fakeStore) ListSessions(_ context.Context, filter ListSessionsFilter) ([]Session, error) {
	items := []Session{}
	reviewSLASeconds := filter.ReviewSLASeconds
	if reviewSLASeconds <= 0 {
		reviewSLASeconds = 15 * 60
	}
	for _, item := range s.sessions {
		if filter.Channel != "" && item.Channel != filter.Channel {
			continue
		}
		if filter.Status != "" && item.Status != filter.Status {
			continue
		}
		if filter.HandoffStatus != "" && item.HandoffStatus != filter.HandoffStatus {
			continue
		}
		if filter.ContactKey != "" && item.ContactKey != filter.ContactKey {
			continue
		}
		agent := asMap(item.Metadata["agent"])
		agentStatus := strings.ToUpper(strings.TrimSpace(asString(agent["status"])))
		if filter.AgentStatus != "" && agentStatus != filter.AgentStatus {
			continue
		}
		switch filter.DraftReviewStatus {
		case "PENDING_REVIEW":
			if agentStatus != agentStatusDraftGenerated {
				continue
			}
		case "REVIEWED":
			if agentStatus != agentStatusDraftReviewed {
				continue
			}
		}
		items = append(items, item)
	}
	observedAt := time.Now().UTC()
	sort.Slice(items, func(i, j int) bool {
		if filter.OrderBy == "REVIEW_PRIORITY" {
			priority := func(session Session) int {
				decorated := decorateSessionDraftSummary(session, reviewSLASeconds, observedAt)
				switch decorated.DraftReviewPriority {
				case "HIGH":
					return 0
				case "MEDIUM":
					return 1
				case "LOW":
					return 2
				case "REVIEWED":
					return 3
				default:
					if decorated.DraftReviewStatus == "REVIEWED" {
						return 3
					}
					return 4
				}
			}
			left := priority(items[i])
			right := priority(items[j])
			if left != right {
				return left < right
			}
			leftDecorated := decorateSessionDraftSummary(items[i], reviewSLASeconds, observedAt)
			rightDecorated := decorateSessionDraftSummary(items[j], reviewSLASeconds, observedAt)
			if leftDecorated.DraftPendingAgeSeconds != rightDecorated.DraftPendingAgeSeconds {
				return leftDecorated.DraftPendingAgeSeconds > rightDecorated.DraftPendingAgeSeconds
			}
		}
		left := items[i].LastMessageAt
		right := items[j].LastMessageAt
		if left == nil {
			return false
		}
		if right == nil {
			return true
		}
		return left.After(*right)
	})
	return items, nil
}

func (s *fakeStore) CountSessionsSummary(_ context.Context, filter ListSessionsFilter, reviewSLASeconds int) (SessionsSummary, error) {
	var summary SessionsSummary
	for _, item := range s.sessions {
		if filter.Channel != "" && item.Channel != filter.Channel {
			continue
		}
		if filter.Status != "" && item.Status != filter.Status {
			continue
		}
		if filter.HandoffStatus != "" && item.HandoffStatus != filter.HandoffStatus {
			continue
		}
		if filter.ContactKey != "" && item.ContactKey != filter.ContactKey {
			continue
		}

		decorated := decorateSessionDraftSummary(item, reviewSLASeconds, time.Now().UTC())
		summary.TotalCount++
		switch decorated.DraftReviewStatus {
		case "PENDING_REVIEW":
			summary.PendingReviewCount++
		case "REVIEWED":
			summary.ReviewedCount++
		default:
			summary.NoDraftCount++
		}
		if decorated.HandoffStatus == "HUMAN" && strings.TrimSpace(decorated.CurrentOwnerUserID) != "" {
			summary.HumanOwnedCount++
		}
		if decorated.HandoffStatus == "BOT" && strings.TrimSpace(decorated.CurrentOwnerUserID) == "" {
			summary.BotOwnedCount++
		}
		switch decorated.DraftPendingAgeBucket {
		case "DUE_SOON":
			summary.DueSoonReviewCount++
		case "OVERDUE":
			summary.OverdueReviewCount++
		}
		switch decorated.DraftReviewPriority {
		case "HIGH":
			summary.HighPriorityReviewCount++
		case "MEDIUM":
			summary.MediumPriorityReviewCount++
		case "LOW":
			summary.LowPriorityReviewCount++
		}
		if decorated.DraftPendingAgeSeconds > summary.OldestPendingAgeSeconds {
			summary.OldestPendingAgeSeconds = decorated.DraftPendingAgeSeconds
		}
	}
	summary.ReviewSLASeconds = reviewSLASeconds
	return summary, nil
}

func (s *fakeStore) GetSession(_ context.Context, id string) (Session, error) {
	item, ok := s.sessions[id]
	if !ok {
		return Session{}, ErrSessionNotFound
	}
	return item, nil
}

func (s *fakeStore) ListMessages(_ context.Context, sessionID string, _ ListMessagesFilter) ([]Message, error) {
	items := []Message{}
	for _, id := range s.messageOrder {
		item := s.messages[id]
		if item.SessionID == sessionID {
			items = append(items, item)
		}
	}
	return items, nil
}

func (s *fakeStore) seedSessionWithMessage(contactKey string, body string) (Session, Message) {
	now := time.Now().UTC()
	session, _ := s.UpsertSession(context.Background(), UpsertSessionInput{
		Channel:       "WHATSAPP",
		ContactKey:    contactKey,
		CustomerPhone: contactKey,
		LastMessageAt: &now,
		LastInboundAt: &now,
	})
	message, _ := s.CreateMessage(context.Background(), CreateMessageInput{
		SessionID:        session.ID,
		Direction:        "INBOUND",
		Kind:             "TEXT",
		ProcessingStatus: "RECEIVED",
		ReceivedAt:       now,
		Body:             body,
	})
	return session, message
}

func timePointer(value time.Time) *time.Time {
	utc := value.UTC()
	return &utc
}
