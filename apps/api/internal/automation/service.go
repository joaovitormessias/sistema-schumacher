package automation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"schumacher-tur/api/internal/bookings"
	"schumacher-tur/api/internal/chat"
	"schumacher-tur/api/internal/payments"
	"schumacher-tur/api/internal/shared/config"
)

var (
	ErrInvalidEvolutionPayload    = errors.New("invalid evolution payload")
	ErrMissingEvolutionChatKey    = errors.New("missing evolution remoteJid")
	ErrMissingEvolutionMessageID  = errors.New("missing evolution message id")
	ErrMissingEvolutionStatus     = errors.New("missing evolution status")
	ErrMissingEvolutionPresence   = errors.New("missing evolution presence status")
	ErrMissingJobName             = errors.New("missing automation job name")
	ErrInvalidJobRunLimit         = errors.New("invalid automation job run limit")
	ErrInvalidBookingsExpireLimit = errors.New("invalid bookings expire limit")
	ErrInvalidBookingsExpireHold  = errors.New("invalid bookings expire hold_minutes")
	ErrPaymentNotificationTarget  = errors.New("payment_id or booking_id is required")
)

type ChatIngestor interface {
	Ingest(ctx context.Context, input chat.IngestMessageInput) (chat.IngestMessageResult, error)
	ApplyPresenceSignal(ctx context.Context, input chat.ApplyPresenceSignalInput) (chat.ApplyPresenceSignalResult, error)
	GetSessionsSummary(ctx context.Context, filter chat.ListSessionsFilter) (chat.SessionsSummary, error)
	QueueAutomationDraft(ctx context.Context, input chat.QueueAutomationDraftInput) (chat.QueueAutomationDraftResult, error)
}

type chatReviewAlertNotifier interface {
	Enabled() bool
	NotifyReviewAlert(ctx context.Context, payload ChatReviewAlertNotificationPayload) error
}

type paymentNotificationSource interface {
	Get(ctx context.Context, id string) (payments.Payment, error)
	List(ctx context.Context, filter payments.PaymentListFilter) ([]payments.Payment, error)
	GetBookingNotificationContext(ctx context.Context, bookingID string) (payments.PaymentNotificationContext, error)
}

type bookingExpirationSource interface {
	List(ctx context.Context, filter bookings.ListFilter) ([]bookings.BookingListItem, error)
	UpdateStatus(ctx context.Context, id string, status string) (bookings.BookingDetails, error)
}

type Service struct {
	store               Store
	chat                ChatIngestor
	cfg                 config.Config
	reviewAlertNotifier chatReviewAlertNotifier
	paymentSource       paymentNotificationSource
	bookingSource       bookingExpirationSource
}

func NewService(store Store, chatSvc ChatIngestor, cfg config.Config, deps ...interface{}) *Service {
	notifier := newChatReviewAlertNotifier(cfg)
	var paymentSource paymentNotificationSource
	var bookingSource bookingExpirationSource
	for _, dep := range deps {
		switch typed := dep.(type) {
		case chatReviewAlertNotifier:
			if typed != nil {
				notifier = typed
			}
		case paymentNotificationSource:
			if typed != nil {
				paymentSource = typed
			}
		case bookingExpirationSource:
			if typed != nil {
				bookingSource = typed
			}
		}
	}
	return &Service{
		store:               store,
		chat:                chatSvc,
		cfg:                 cfg,
		reviewAlertNotifier: notifier,
		paymentSource:       paymentSource,
		bookingSource:       bookingSource,
	}
}

func (s *Service) HandleEvolutionMessages(ctx context.Context, body []byte) (EvolutionWebhookResult, error) {
	payload, payloadMap, err := decodeEvolutionPayload(body)
	if err != nil {
		return EvolutionWebhookResult{}, err
	}

	event := strings.TrimSpace(payload.Event)
	if event != "messages.upsert" {
		return EvolutionWebhookResult{
			Status: "skipped",
			Event:  event,
			Reason: "unsupported_event",
		}, nil
	}

	contactKey := strings.TrimSpace(payload.Data.Key.RemoteJID)
	if contactKey == "" {
		contactKey = strings.TrimSpace(payload.Data.Key.RemoteJIDAlt)
	}
	if contactKey == "" {
		return EvolutionWebhookResult{}, ErrMissingEvolutionChatKey
	}

	if payload.Data.Key.FromMe {
		return EvolutionWebhookResult{
			Status:            "skipped",
			Event:             event,
			Direction:         "OUTBOUND",
			ContactKey:        contactKey,
			ProviderMessageID: strings.TrimSpace(payload.Data.Key.ID),
			MessageType:       strings.TrimSpace(payload.Data.MessageType),
			Reason:            "from_me",
		}, nil
	}

	receivedAt := resolveEvolutionTimestamp(payload)
	textBody := extractEvolutionText(payload.Data)
	normalized := map[string]interface{}{
		"provider":                 "EVOLUTION",
		"event":                    event,
		"instance":                 strings.TrimSpace(payload.Instance),
		"message_type":             strings.TrimSpace(payload.Data.MessageType),
		"message_text":             textBody,
		"from_me":                  payload.Data.Key.FromMe,
		"remote_jid":               contactKey,
		"customer_phone":           normalizePhone(contactKey),
		"push_name":                strings.TrimSpace(payload.Data.PushName),
		"chatwoot_message_id":      payload.Data.ChatwootMessageID,
		"chatwoot_inbox_id":        payload.Data.ChatwootInboxID,
		"chatwoot_conversation_id": payload.Data.ChatwootConversationID,
	}

	result, err := s.chat.Ingest(ctx, chat.IngestMessageInput{
		Channel:       "WHATSAPP",
		ContactKey:    contactKey,
		CustomerPhone: normalizePhone(contactKey),
		CustomerName:  strings.TrimSpace(payload.Data.PushName),
		Metadata: map[string]interface{}{
			"provider":   "EVOLUTION",
			"instance":   strings.TrimSpace(payload.Instance),
			"server_url": strings.TrimSpace(payload.ServerURL),
			"source":     strings.TrimSpace(payload.Data.Source),
		},
		Message: chat.IngestMessagePayload{
			Direction:         "INBOUND",
			Kind:              mapEvolutionKind(payload.Data.MessageType),
			ProviderMessageID: strings.TrimSpace(payload.Data.Key.ID),
			IdempotencyKey:    buildEvolutionIdempotencyKey(contactKey, payload.Data.Key.ID),
			SenderName:        strings.TrimSpace(payload.Data.PushName),
			SenderPhone:       normalizePhone(contactKey),
			Body:              textBody,
			Payload:           payloadMap,
			NormalizedPayload: normalized,
			ProcessingStatus:  "RECEIVED",
			ReceivedAt:        &receivedAt,
		},
	})
	if err != nil {
		return EvolutionWebhookResult{}, err
	}

	return EvolutionWebhookResult{
		Status:            "accepted",
		Event:             event,
		Direction:         "INBOUND",
		ContactKey:        contactKey,
		CustomerPhone:     normalizePhone(contactKey),
		ProviderMessageID: strings.TrimSpace(payload.Data.Key.ID),
		MessageType:       strings.TrimSpace(payload.Data.MessageType),
		SessionID:         result.Session.ID,
		MessageID:         result.Message.ID,
		Idempotent:        result.Idempotent,
	}, nil
}

func (s *Service) HandleEvolutionStatus(ctx context.Context, body []byte) (EvolutionStatusWebhookResult, error) {
	payload, payloadMap, err := decodeEvolutionPayload(body)
	if err != nil {
		return EvolutionStatusWebhookResult{}, err
	}

	providerMessageID := strings.TrimSpace(payload.Data.Key.ID)
	if providerMessageID == "" {
		return EvolutionStatusWebhookResult{}, ErrMissingEvolutionMessageID
	}

	providerStatus := normalizeProviderStatus(payload.Data.Status)
	if providerStatus == "" {
		return EvolutionStatusWebhookResult{}, ErrMissingEvolutionStatus
	}

	record, err := s.store.RecordEvolutionStatus(ctx, RecordEvolutionStatusInput{
		ProviderMessageID: providerMessageID,
		ProviderStatus:    providerStatus,
		Event:             strings.TrimSpace(payload.Event),
		ObservedAt:        resolveEvolutionTimestamp(payload),
		Payload:           payloadMap,
	})
	if err != nil {
		return EvolutionStatusWebhookResult{}, err
	}

	return EvolutionStatusWebhookResult{
		Status:                  "accepted",
		Event:                   strings.TrimSpace(payload.Event),
		ProviderMessageID:       providerMessageID,
		ProviderStatus:          providerStatus,
		MessageType:             strings.TrimSpace(payload.Data.MessageType),
		MatchedChatMessages:     record.MatchedChatMessages,
		MatchedOutboundMessages: record.MatchedOutboundMessages,
	}, nil
}

func (s *Service) HandleEvolutionPresence(ctx context.Context, body []byte) (EvolutionPresenceWebhookResult, error) {
	payload, payloadMap, err := decodeEvolutionPayload(body)
	if err != nil {
		return EvolutionPresenceWebhookResult{}, err
	}

	event := strings.TrimSpace(payload.Event)
	if event != "presence.update" && event != "presence-update" {
		return EvolutionPresenceWebhookResult{
			Status: "skipped",
			Event:  event,
			Reason: "unsupported_event",
		}, nil
	}

	contactKey := resolveEvolutionPresenceContactKey(payload.Data)
	if contactKey == "" {
		return EvolutionPresenceWebhookResult{}, ErrMissingEvolutionChatKey
	}

	presenceStatus := resolveEvolutionPresenceStatus(payload.Data)
	if presenceStatus == "" {
		return EvolutionPresenceWebhookResult{}, ErrMissingEvolutionPresence
	}

	result, err := s.chat.ApplyPresenceSignal(ctx, chat.ApplyPresenceSignalInput{
		Channel:        "WHATSAPP",
		ContactKey:     contactKey,
		PresenceStatus: presenceStatus,
		ObservedAt:     timePointer(resolveEvolutionTimestamp(payload)),
		Metadata: map[string]interface{}{
			"provider":   "EVOLUTION",
			"instance":   strings.TrimSpace(payload.Instance),
			"event":      event,
			"source":     strings.TrimSpace(payload.Data.Source),
			"payload":    payloadMap,
			"server_url": strings.TrimSpace(payload.ServerURL),
		},
	})
	if err != nil {
		return EvolutionPresenceWebhookResult{}, err
	}

	return EvolutionPresenceWebhookResult{
		Status:        result.Status,
		Event:         event,
		ContactKey:    contactKey,
		CustomerPhone: normalizePhone(contactKey),
		PresenceState: presenceStatus,
		SessionID:     result.Session.ID,
		Reason:        result.Reason,
	}, nil
}

func (s *Service) RunChatReviewAlerts(ctx context.Context, input RunChatReviewAlertsInput) (RunChatReviewAlertsResult, error) {
	filter := chat.ListSessionsFilter{
		Channel:       strings.ToUpper(strings.TrimSpace(input.Channel)),
		Status:        strings.ToUpper(strings.TrimSpace(input.Status)),
		HandoffStatus: strings.ToUpper(strings.TrimSpace(input.HandoffStatus)),
		ContactKey:    strings.TrimSpace(input.ContactKey),
	}
	observedAt := time.Now().UTC()
	summary, err := s.chat.GetSessionsSummary(ctx, filter)
	if err != nil {
		return RunChatReviewAlertsResult{}, err
	}

	result := RunChatReviewAlertsResult{
		Status:            "skipped",
		Reason:            "no_active_alert",
		WebhookConfigured: s.reviewAlertNotifier != nil && s.reviewAlertNotifier.Enabled(),
		ObservedAt:        observedAt,
		Filter: RunChatReviewAlertsInput{
			Channel:       filter.Channel,
			Status:        filter.Status,
			HandoffStatus: filter.HandoffStatus,
			ContactKey:    filter.ContactKey,
		},
		Summary: buildChatReviewAlertSummarySnapshot(summary),
	}
	if !summary.HasReviewAlert {
		return result, nil
	}
	if s.reviewAlertNotifier == nil || !s.reviewAlertNotifier.Enabled() {
		result.Reason = "notifier_not_configured"
		return result, nil
	}

	alertSignature := buildChatReviewAlertSignature(result.Filter, summary)
	if existing, err := s.store.FindLatestJobRunByKey(ctx, "CHAT_REVIEW_ALERTS", alertSignature); err != nil {
		return RunChatReviewAlertsResult{}, err
	} else if existing != nil {
		result.Reason = "duplicate_alert_state"
		result.Deduplicated = true
		result.JobRunID = existing.ID
		return result, nil
	}

	jobRun, err := s.store.CreateJobRun(ctx, CreateJobRunInput{
		JobName:       "CHAT_REVIEW_ALERTS",
		TriggerSource: "MANUAL",
		Status:        "RUNNING",
		InputPayload: map[string]interface{}{
			"filter":          result.Filter,
			"summary":         result.Summary,
			"alert_signature": alertSignature,
		},
		StartedAt: observedAt,
	})
	if err != nil {
		return RunChatReviewAlertsResult{}, err
	}
	result.JobRunID = jobRun.ID

	payload := ChatReviewAlertNotificationPayload{
		Source:            "chat_review_queue",
		ObservedAt:        observedAt,
		AlertLevel:        summary.ReviewAlertLevel,
		AlertCode:         summary.ReviewAlertCode,
		AlertMessage:      summary.ReviewAlertMessage,
		AlertSessionCount: summary.ReviewAlertSessionCount,
		Filter:            result.Filter,
		Summary:           result.Summary,
	}
	if err := s.reviewAlertNotifier.NotifyReviewAlert(ctx, payload); err != nil {
		_, updateErr := s.store.UpdateJobRun(ctx, UpdateJobRunInput{
			ID:         jobRun.ID,
			Status:     "FAILED",
			ErrorText:  err.Error(),
			FinishedAt: time.Now().UTC(),
			ResultPayload: map[string]interface{}{
				"filter":          result.Filter,
				"summary":         result.Summary,
				"alert_signature": alertSignature,
				"delivery_status": "FAILED",
			},
		})
		if updateErr != nil {
			return RunChatReviewAlertsResult{}, fmt.Errorf("notify review alert: %w; update job run: %v", err, updateErr)
		}
		return RunChatReviewAlertsResult{}, err
	}

	if _, err := s.store.UpdateJobRun(ctx, UpdateJobRunInput{
		ID:         jobRun.ID,
		Status:     "SENT",
		FinishedAt: time.Now().UTC(),
		ResultPayload: map[string]interface{}{
			"filter":          result.Filter,
			"summary":         result.Summary,
			"alert_signature": alertSignature,
			"delivery_status": "SENT",
		},
	}); err != nil {
		return RunChatReviewAlertsResult{}, err
	}

	result.Status = "sent"
	result.Reason = "alert_sent"
	result.NotificationTriggered = true
	return result, nil
}

func (s *Service) RunBookingsExpire(ctx context.Context, input RunBookingsExpireInput) (RunBookingsExpireResult, error) {
	if s.bookingSource == nil {
		return RunBookingsExpireResult{}, errors.New("booking expiration source is not configured")
	}

	filter, err := normalizeRunBookingsExpireInput(input)
	if err != nil {
		return RunBookingsExpireResult{}, err
	}

	observedAt := time.Now().UTC()
	jobRun, err := s.store.CreateJobRun(ctx, CreateJobRunInput{
		JobName:       "BOOKINGS_EXPIRE",
		TriggerSource: "MANUAL",
		Status:        "RUNNING",
		InputPayload: map[string]interface{}{
			"status":       "PENDING",
			"limit":        filter.Limit,
			"hold_minutes": filter.HoldMinutes,
		},
		StartedAt: observedAt,
	})
	if err != nil {
		return RunBookingsExpireResult{}, err
	}

	result := RunBookingsExpireResult{
		Status:      "skipped",
		Reason:      "no_expired_bookings",
		ObservedAt:  observedAt,
		JobRunID:    jobRun.ID,
		Limit:       filter.Limit,
		HoldMinutes: filter.HoldMinutes,
	}

	items, err := s.bookingSource.List(ctx, bookings.ListFilter{
		Status: "PENDING",
		Limit:  filter.Limit,
	})
	if err != nil {
		if _, updateErr := s.store.UpdateJobRun(ctx, UpdateJobRunInput{
			ID:         jobRun.ID,
			Status:     "FAILED",
			ErrorText:  err.Error(),
			FinishedAt: time.Now().UTC(),
			ResultPayload: map[string]interface{}{
				"reason": "list_pending_bookings_failed",
			},
		}); updateErr != nil {
			return RunBookingsExpireResult{}, fmt.Errorf("list pending bookings: %w; update job run: %v", err, updateErr)
		}
		return RunBookingsExpireResult{}, err
	}

	result.CheckedCount = len(items)
	expired := collectExpiredBookings(items, observedAt, filter.HoldMinutes)
	if len(expired) == 0 {
		if _, err := s.store.UpdateJobRun(ctx, UpdateJobRunInput{
			ID:         jobRun.ID,
			Status:     "SKIPPED",
			FinishedAt: time.Now().UTC(),
			ResultPayload: map[string]interface{}{
				"reason":        result.Reason,
				"checked_count": result.CheckedCount,
				"expired_count": 0,
			},
		}); err != nil {
			return RunBookingsExpireResult{}, err
		}
		return result, nil
	}

	expiredResults := make([]ExpiredBookingResult, 0, len(expired))
	for _, candidate := range expired {
		details, err := s.bookingSource.UpdateStatus(ctx, candidate.Booking.ID, "EXPIRED")
		if err != nil {
			if _, updateErr := s.store.UpdateJobRun(ctx, UpdateJobRunInput{
				ID:         jobRun.ID,
				Status:     "FAILED",
				ErrorText:  err.Error(),
				FinishedAt: time.Now().UTC(),
				ResultPayload: map[string]interface{}{
					"reason":               "expire_booking_failed",
					"checked_count":        result.CheckedCount,
					"expired_count":        len(expiredResults),
					"failed_booking_id":    candidate.Booking.ID,
					"reservation_code":     candidate.Booking.ReservationCode,
					"effective_expires_at": candidate.EffectiveExpiresAt.Format(time.RFC3339Nano),
				},
			}); updateErr != nil {
				return RunBookingsExpireResult{}, fmt.Errorf("expire booking %s: %w; update job run: %v", candidate.Booking.ID, err, updateErr)
			}
			return RunBookingsExpireResult{}, err
		}

		expiredResults = append(expiredResults, ExpiredBookingResult{
			BookingID:          details.Booking.ID,
			ReservationCode:    details.Booking.ReservationCode,
			TripID:             details.Booking.TripID,
			PreviousStatus:     candidate.Booking.Status,
			UpdatedStatus:      details.Booking.Status,
			EffectiveExpiresAt: candidate.EffectiveExpiresAt,
			ExpirationSource:   candidate.ExpirationSource,
		})
	}

	result.Status = "processed"
	result.Reason = "expired_bookings_processed"
	result.ExpiredCount = len(expiredResults)
	result.ExpiredBookings = expiredResults

	if _, err := s.store.UpdateJobRun(ctx, UpdateJobRunInput{
		ID:         jobRun.ID,
		Status:     "COMPLETED",
		FinishedAt: time.Now().UTC(),
		ResultPayload: map[string]interface{}{
			"reason":           result.Reason,
			"checked_count":    result.CheckedCount,
			"expired_count":    result.ExpiredCount,
			"expired_bookings": serializeExpiredBookings(expiredResults),
		},
	}); err != nil {
		return RunBookingsExpireResult{}, err
	}

	return result, nil
}

func (s *Service) RunPaymentNotifications(ctx context.Context, input RunPaymentNotificationsInput) (RunPaymentNotificationsResult, error) {
	if s.paymentSource == nil {
		return RunPaymentNotificationsResult{}, errors.New("payment notification source is not configured")
	}

	payment, err := s.resolvePaymentNotificationPayment(ctx, input)
	if err != nil {
		return RunPaymentNotificationsResult{}, err
	}

	notification, err := s.paymentSource.GetBookingNotificationContext(ctx, payment.BookingID)
	if err != nil {
		return RunPaymentNotificationsResult{}, err
	}

	observedAt := time.Now().UTC()
	channel := strings.ToUpper(strings.TrimSpace(input.Channel))
	if channel == "" {
		channel = "WHATSAPP"
	}

	payload := buildPaymentNotificationPayload(payment, notification, observedAt)
	preview := buildPaymentNotificationPreview(payload)
	draftKey := buildPaymentNotificationDraftKey(payment.ID)

	result := RunPaymentNotificationsResult{
		Status:          "skipped",
		ObservedAt:      observedAt,
		PaymentID:       payment.ID,
		BookingID:       payment.BookingID,
		ReservationCode: notification.ReservationCode,
		CustomerName:    notification.CustomerName,
		CustomerPhone:   preview.CustomerPhone,
		ContactKey:      preview.ContactKey,
		PaymentStatus:   notification.PaymentStatus,
		AmountPaid:      notification.AmountPaid,
		AmountDue:       notification.AmountDue,
		Messages:        preview.Messages,
	}

	jobRun, err := s.store.CreateJobRun(ctx, CreateJobRunInput{
		JobName:       "PAYMENT_NOTIFICATIONS",
		TriggerSource: "MANUAL",
		Status:        "RUNNING",
		InputPayload: map[string]interface{}{
			"payment_id":            payment.ID,
			"booking_id":            payment.BookingID,
			"channel":               channel,
			"reservation_code":      notification.ReservationCode,
			"customer_phone":        preview.CustomerPhone,
			"contact_key":           preview.ContactKey,
			"payment_status":        notification.PaymentStatus,
			"amount_paid":           notification.AmountPaid,
			"amount_due":            notification.AmountDue,
			"draft_idempotency_key": draftKey,
		},
		StartedAt: observedAt,
	})
	if err != nil {
		return RunPaymentNotificationsResult{}, err
	}
	result.JobRunID = jobRun.ID

	if preview.ContactKey == "" {
		if _, err := s.store.UpdateJobRun(ctx, UpdateJobRunInput{
			ID:         jobRun.ID,
			Status:     "SKIPPED",
			FinishedAt: time.Now().UTC(),
			ResultPayload: map[string]interface{}{
				"reason":     "missing_customer_phone",
				"payment_id": payment.ID,
				"booking_id": payment.BookingID,
			},
		}); err != nil {
			return RunPaymentNotificationsResult{}, err
		}
		result.Reason = "missing_customer_phone"
		return result, nil
	}

	draft, err := s.chat.QueueAutomationDraft(ctx, chat.QueueAutomationDraftInput{
		Channel:        channel,
		ContactKey:     preview.ContactKey,
		CustomerPhone:  preview.CustomerPhone,
		CustomerName:   notification.CustomerName,
		Body:           preview.Body,
		SenderName:     "AUTOMATION",
		IdempotencyKey: draftKey,
		Metadata: map[string]interface{}{
			"source":                "PAYMENT_NOTIFICATION",
			"automation_job_name":   "PAYMENT_NOTIFICATIONS",
			"automation_job_run_id": jobRun.ID,
			"draft_model":           "PAYMENT_NOTIFICATION_TEMPLATE",
			"shadow_mode":           true,
			"requires_human_review": true,
			"payment_id":            payment.ID,
			"booking_id":            payment.BookingID,
			"reservation_code":      notification.ReservationCode,
			"payment_status":        notification.PaymentStatus,
			"amount_paid":           notification.AmountPaid,
			"amount_due":            notification.AmountDue,
			"event":                 payload.Event,
			"messages":              preview.Messages,
		},
	})
	if err != nil {
		if _, updateErr := s.store.UpdateJobRun(ctx, UpdateJobRunInput{
			ID:         jobRun.ID,
			Status:     "FAILED",
			ErrorText:  err.Error(),
			FinishedAt: time.Now().UTC(),
			ResultPayload: map[string]interface{}{
				"payment_id": payment.ID,
				"booking_id": payment.BookingID,
				"reason":     "draft_queue_failed",
			},
		}); updateErr != nil {
			return RunPaymentNotificationsResult{}, fmt.Errorf("queue payment notification draft: %w; update job run: %v", err, updateErr)
		}
		return RunPaymentNotificationsResult{}, err
	}

	jobStatus := "QUEUED"
	result.Status = "queued"
	result.Reason = "draft_created"
	result.DraftQueued = true
	result.SessionID = draft.Session.ID
	result.DraftMessageID = draft.Message.ID
	if draft.Idempotent {
		jobStatus = "SKIPPED"
		result.Status = "skipped"
		result.Reason = "draft_already_exists"
		result.DraftQueued = false
		result.Idempotent = true
	}

	if _, err := s.store.UpdateJobRun(ctx, UpdateJobRunInput{
		ID:         jobRun.ID,
		Status:     jobStatus,
		FinishedAt: time.Now().UTC(),
		ResultPayload: map[string]interface{}{
			"payment_id":            payment.ID,
			"booking_id":            payment.BookingID,
			"reservation_code":      notification.ReservationCode,
			"contact_key":           preview.ContactKey,
			"draft_idempotency_key": draftKey,
			"session_id":            draft.Session.ID,
			"draft_message_id":      draft.Message.ID,
			"idempotent":            draft.Idempotent,
			"messages_count":        len(preview.Messages),
			"reason":                result.Reason,
		},
	}); err != nil {
		return RunPaymentNotificationsResult{}, err
	}

	return result, nil
}

func (s *Service) ListJobRuns(ctx context.Context, input ListJobRunsInput) (ListJobRunsResult, error) {
	jobName := normalizeJobName(input.JobName)
	if jobName == "" {
		return ListJobRunsResult{}, ErrMissingJobName
	}

	limit := input.Limit
	switch {
	case limit == 0:
		limit = 20
	case limit < 0:
		return ListJobRunsResult{}, ErrInvalidJobRunLimit
	case limit > 100:
		limit = 100
	}

	filter := ListJobRunsInput{
		JobName:       jobName,
		Status:        strings.ToUpper(strings.TrimSpace(input.Status)),
		TriggerSource: strings.ToUpper(strings.TrimSpace(input.TriggerSource)),
		Limit:         limit,
	}
	runs, err := s.store.ListJobRuns(ctx, filter)
	if err != nil {
		return ListJobRunsResult{}, err
	}

	return ListJobRunsResult{
		JobName: filter.JobName,
		Filter:  filter,
		Count:   len(runs),
		Runs:    runs,
	}, nil
}

func buildChatReviewAlertSummarySnapshot(summary chat.SessionsSummary) ChatReviewAlertSummarySnapshot {
	return ChatReviewAlertSummarySnapshot{
		TotalCount:                summary.TotalCount,
		PendingReviewCount:        summary.PendingReviewCount,
		ReviewedCount:             summary.ReviewedCount,
		NoDraftCount:              summary.NoDraftCount,
		HumanOwnedCount:           summary.HumanOwnedCount,
		BotOwnedCount:             summary.BotOwnedCount,
		DueSoonReviewCount:        summary.DueSoonReviewCount,
		OverdueReviewCount:        summary.OverdueReviewCount,
		HighPriorityReviewCount:   summary.HighPriorityReviewCount,
		MediumPriorityReviewCount: summary.MediumPriorityReviewCount,
		LowPriorityReviewCount:    summary.LowPriorityReviewCount,
		OldestPendingAgeSeconds:   summary.OldestPendingAgeSeconds,
		ReviewSLASeconds:          summary.ReviewSLASeconds,
		HasReviewAlert:            summary.HasReviewAlert,
		ReviewAlertLevel:          summary.ReviewAlertLevel,
		ReviewAlertCode:           summary.ReviewAlertCode,
		ReviewAlertMessage:        summary.ReviewAlertMessage,
		ReviewAlertSessionCount:   summary.ReviewAlertSessionCount,
	}
}

func buildChatReviewAlertSignature(filter RunChatReviewAlertsInput, summary chat.SessionsSummary) string {
	return strings.Join([]string{
		strings.ToUpper(strings.TrimSpace(summary.ReviewAlertCode)),
		strings.ToUpper(strings.TrimSpace(summary.ReviewAlertLevel)),
		fmt.Sprintf("%d", summary.ReviewAlertSessionCount),
		fmt.Sprintf("%d", summary.OverdueReviewCount),
		fmt.Sprintf("%d", summary.DueSoonReviewCount),
		fmt.Sprintf("%d", summary.HighPriorityReviewCount),
		fmt.Sprintf("%d", summary.MediumPriorityReviewCount),
		fmt.Sprintf("%d", summary.LowPriorityReviewCount),
		fmt.Sprintf("%d", summary.ReviewSLASeconds),
		strings.ToUpper(strings.TrimSpace(filter.Channel)),
		strings.ToUpper(strings.TrimSpace(filter.Status)),
		strings.ToUpper(strings.TrimSpace(filter.HandoffStatus)),
		strings.TrimSpace(filter.ContactKey),
	}, "|")
}

func normalizeJobName(input string) string {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return ""
	}
	trimmed = strings.ReplaceAll(trimmed, "-", "_")
	trimmed = strings.ReplaceAll(trimmed, " ", "_")
	return strings.ToUpper(trimmed)
}

func (s *Service) resolvePaymentNotificationPayment(ctx context.Context, input RunPaymentNotificationsInput) (payments.Payment, error) {
	paymentID := strings.TrimSpace(input.PaymentID)
	if paymentID != "" {
		return s.paymentSource.Get(ctx, paymentID)
	}

	bookingID := strings.TrimSpace(input.BookingID)
	if bookingID == "" {
		return payments.Payment{}, ErrPaymentNotificationTarget
	}

	items, err := s.paymentSource.List(ctx, payments.PaymentListFilter{
		BookingID: bookingID,
		Status:    "PAID",
		Limit:     1,
	})
	if err != nil {
		return payments.Payment{}, err
	}
	if len(items) == 0 {
		return payments.Payment{}, pgx.ErrNoRows
	}
	return items[0], nil
}

func decodeEvolutionPayload(body []byte) (EvolutionWebhookPayload, map[string]interface{}, error) {
	var payload EvolutionWebhookPayload
	if err := json.Unmarshal(body, &payload); err == nil && payload.Event != "" {
		rawMap := map[string]interface{}{}
		if err := json.Unmarshal(body, &rawMap); err != nil {
			return EvolutionWebhookPayload{}, nil, ErrInvalidEvolutionPayload
		}
		return payload, rawMap, nil
	}

	var envelope EvolutionEnvelope
	if err := json.Unmarshal(body, &envelope); err != nil || len(envelope.Body) == 0 {
		return EvolutionWebhookPayload{}, nil, ErrInvalidEvolutionPayload
	}
	if err := json.Unmarshal(envelope.Body, &payload); err != nil || payload.Event == "" {
		return EvolutionWebhookPayload{}, nil, ErrInvalidEvolutionPayload
	}
	rawMap := map[string]interface{}{}
	if err := json.Unmarshal(envelope.Body, &rawMap); err != nil {
		return EvolutionWebhookPayload{}, nil, ErrInvalidEvolutionPayload
	}
	return payload, rawMap, nil
}

func resolveEvolutionTimestamp(payload EvolutionWebhookPayload) time.Time {
	if payload.Data.MessageTimestamp > 0 {
		return time.Unix(payload.Data.MessageTimestamp, 0).UTC()
	}
	if parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(payload.DateTime)); err == nil {
		return parsed.UTC()
	}
	return time.Now().UTC()
}

func extractEvolutionText(data EvolutionMessageData) string {
	switch strings.TrimSpace(data.MessageType) {
	case "conversation":
		return strings.TrimSpace(data.Message.Conversation)
	case "extendedTextMessage":
		if data.Message.ExtendedTextMessage != nil {
			return strings.TrimSpace(data.Message.ExtendedTextMessage.Text)
		}
	case "imageMessage":
		if data.Message.ImageMessage != nil {
			return strings.TrimSpace(data.Message.ImageMessage.Caption)
		}
	case "videoMessage":
		if data.Message.VideoMessage != nil {
			return strings.TrimSpace(data.Message.VideoMessage.Caption)
		}
	}

	if text := strings.TrimSpace(data.Message.Conversation); text != "" {
		return text
	}
	if data.Message.ExtendedTextMessage != nil {
		if text := strings.TrimSpace(data.Message.ExtendedTextMessage.Text); text != "" {
			return text
		}
	}
	if data.Message.ImageMessage != nil {
		if text := strings.TrimSpace(data.Message.ImageMessage.Caption); text != "" {
			return text
		}
	}
	if data.Message.VideoMessage != nil {
		return strings.TrimSpace(data.Message.VideoMessage.Caption)
	}
	return ""
}

func normalizePhone(contactKey string) string {
	trimmed := strings.TrimSpace(contactKey)
	if trimmed == "" {
		return ""
	}
	if idx := strings.Index(trimmed, "@"); idx >= 0 {
		trimmed = trimmed[:idx]
	}

	builder := strings.Builder{}
	for _, r := range trimmed {
		if r >= '0' && r <= '9' {
			builder.WriteRune(r)
		}
	}
	return builder.String()
}

func mapEvolutionKind(messageType string) string {
	switch strings.TrimSpace(messageType) {
	case "conversation", "extendedTextMessage":
		return "TEXT"
	case "imageMessage":
		return "IMAGE"
	case "videoMessage":
		return "VIDEO"
	case "audioMessage":
		return "AUDIO"
	case "documentMessage":
		return "DOCUMENT"
	default:
		return strings.ToUpper(strings.TrimSpace(messageType))
	}
}

func buildEvolutionIdempotencyKey(contactKey string, providerMessageID string) string {
	contactKey = strings.TrimSpace(contactKey)
	providerMessageID = strings.TrimSpace(providerMessageID)
	if contactKey == "" && providerMessageID == "" {
		return ""
	}
	return "evolution:" + contactKey + ":" + providerMessageID
}

func normalizeProviderStatus(status string) string {
	return strings.ToUpper(strings.TrimSpace(status))
}

func resolveEvolutionPresenceContactKey(data EvolutionMessageData) string {
	if value := strings.TrimSpace(data.ID); value != "" {
		return value
	}
	if value := strings.TrimSpace(data.Key.RemoteJID); value != "" {
		return value
	}
	for key := range data.Presences {
		if strings.TrimSpace(key) != "" {
			return key
		}
	}
	return ""
}

func resolveEvolutionPresenceStatus(data EvolutionMessageData) string {
	for _, raw := range data.Presences {
		payload, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		for _, key := range []string{"lastKnownPresence", "presence", "status"} {
			if value := normalizeProviderStatus(stringValue(payload[key])); value != "" {
				return value
			}
		}
	}
	return normalizeProviderStatus(data.Status)
}

func timePointer(value time.Time) *time.Time {
	utc := value.UTC()
	return &utc
}

func stringValue(value interface{}) string {
	typed, ok := value.(string)
	if !ok {
		return ""
	}
	return typed
}
