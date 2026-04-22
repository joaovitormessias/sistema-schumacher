package chat

import (
	"fmt"
	"strings"
	"time"
)

const (
	bufferStatusPending = "PENDING"
	bufferStatusIdle    = "IDLE"
	presenceTyping      = "TYPING"
	presenceRecording   = "RECORDING"
	presencePaused      = "PAUSED"
	presenceAvailable   = "AVAILABLE"
	presenceUnavailable = "UNAVAILABLE"
)

const (
	agentBlockReasonHumanOwnerActive = "human_owner_active"
)

type bufferState struct {
	Status           string
	WindowStartedAt  *time.Time
	PendingUntil     *time.Time
	MessageCount     int
	LastMessageID    string
	LastMessageAt    *time.Time
	LastMessageBody  string
	LastDirection    string
	DebounceWindowMS int
}

func buildBufferState(metadata map[string]interface{}, message Message, debounceWindow time.Duration) map[string]interface{} {
	observedAt := message.ReceivedAt.UTC()
	windowMS := int(debounceWindow / time.Millisecond)
	if message.Direction != "INBOUND" || debounceWindow <= 0 {
		return map[string]interface{}{
			"status":             bufferStatusIdle,
			"cleared_at":         observedAt.Format(time.RFC3339Nano),
			"last_message_id":    message.ID,
			"last_message_at":    observedAt.Format(time.RFC3339Nano),
			"last_direction":     message.Direction,
			"debounce_window_ms": windowMS,
		}
	}

	current := parseBufferState(metadata)
	startedAt := observedAt
	messageCount := 1
	if current.Status == bufferStatusPending && current.PendingUntil != nil && !observedAt.After(*current.PendingUntil) {
		if current.WindowStartedAt != nil {
			startedAt = current.WindowStartedAt.UTC()
		}
		if current.MessageCount > 0 {
			messageCount = current.MessageCount + 1
		}
	}

	pendingUntil := observedAt.Add(debounceWindow)
	return map[string]interface{}{
		"status":              bufferStatusPending,
		"window_started_at":   startedAt.Format(time.RFC3339Nano),
		"pending_until":       pendingUntil.Format(time.RFC3339Nano),
		"message_count":       messageCount,
		"last_message_id":     message.ID,
		"last_message_at":     observedAt.Format(time.RFC3339Nano),
		"last_message_body":   truncateBufferBody(message.Body),
		"last_direction":      message.Direction,
		"debounce_window_ms":  windowMS,
		"processing_status":   message.ProcessingStatus,
		"provider_message_id": message.ProviderMessageID,
	}
}

func buildHumanBlockedBufferState(metadata map[string]interface{}, message Message, ownerUserID string) map[string]interface{} {
	observedAt := message.ReceivedAt.UTC()
	buffer := cloneBuffer(metadata)

	buffer["status"] = bufferStatusIdle
	buffer["agent_blocked"] = true
	buffer["agent_block_reason"] = agentBlockReasonHumanOwnerActive
	buffer["current_owner_user_id"] = ownerUserID
	buffer["blocked_at"] = observedAt.Format(time.RFC3339Nano)
	buffer["last_message_id"] = message.ID
	buffer["last_message_at"] = observedAt.Format(time.RFC3339Nano)
	buffer["last_message_body"] = truncateBufferBody(message.Body)
	buffer["last_direction"] = message.Direction
	buffer["message_count"] = 0
	buffer["window_started_at"] = nil
	buffer["pending_until"] = nil
	buffer["processing_status"] = message.ProcessingStatus
	buffer["provider_message_id"] = message.ProviderMessageID

	return buffer
}

func applyPresenceSignalToBuffer(metadata map[string]interface{}, presenceStatus string, observedAt time.Time, debounceWindow time.Duration) (map[string]interface{}, string, bool) {
	current := parseBufferState(metadata)
	if current.Status != bufferStatusPending || current.PendingUntil == nil {
		return nil, "no_pending_buffer", false
	}

	buffer := cloneBuffer(metadata)
	buffer["presence_status"] = presenceStatus
	buffer["presence_observed_at"] = observedAt.UTC().Format(time.RFC3339Nano)
	buffer["presence_debounce_window_ms"] = int(debounceWindow / time.Millisecond)

	switch presenceStatus {
	case presenceTyping, presenceRecording:
		nextPendingUntil := observedAt.Add(debounceWindow)
		if current.PendingUntil == nil || nextPendingUntil.After(*current.PendingUntil) {
			buffer["pending_until"] = nextPendingUntil.UTC().Format(time.RFC3339Nano)
		}
		buffer["presence_action"] = "extend"
		return buffer, "presence_extended", true
	case presencePaused, presenceAvailable:
		flushWindow := resolvePresenceFlushWindow(debounceWindow)
		nextPendingUntil := observedAt.Add(flushWindow)
		if current.PendingUntil == nil || current.PendingUntil.After(nextPendingUntil) {
			buffer["pending_until"] = nextPendingUntil.UTC().Format(time.RFC3339Nano)
		}
		buffer["presence_action"] = "flush_window"
		return buffer, "presence_flush_window", true
	case presenceUnavailable:
		buffer["presence_action"] = "observe_only"
		return buffer, "presence_observed", true
	default:
		return nil, "unsupported_presence", false
	}
}

func parseBufferState(metadata map[string]interface{}) bufferState {
	raw, ok := metadata["buffer"].(map[string]interface{})
	if !ok || len(raw) == 0 {
		return bufferState{}
	}

	return bufferState{
		Status:           asString(raw["status"]),
		WindowStartedAt:  parseBufferTime(raw["window_started_at"]),
		PendingUntil:     parseBufferTime(raw["pending_until"]),
		MessageCount:     asInt(raw["message_count"]),
		LastMessageID:    asString(raw["last_message_id"]),
		LastMessageAt:    parseBufferTime(raw["last_message_at"]),
		LastMessageBody:  asString(raw["last_message_body"]),
		LastDirection:    asString(raw["last_direction"]),
		DebounceWindowMS: asInt(raw["debounce_window_ms"]),
	}
}

func cloneBuffer(metadata map[string]interface{}) map[string]interface{} {
	raw, ok := metadata["buffer"].(map[string]interface{})
	if !ok || len(raw) == 0 {
		return map[string]interface{}{}
	}

	cloned := make(map[string]interface{}, len(raw))
	for key, value := range raw {
		cloned[key] = value
	}
	return cloned
}

func truncateBufferBody(body string) string {
	body = strings.TrimSpace(body)
	if len(body) <= 160 {
		return body
	}
	return body[:157] + "..."
}

func resolvePresenceFlushWindow(debounceWindow time.Duration) time.Duration {
	if debounceWindow <= 0 {
		return 400 * time.Millisecond
	}
	flushWindow := debounceWindow / 4
	if flushWindow < 400*time.Millisecond {
		flushWindow = 400 * time.Millisecond
	}
	if flushWindow > debounceWindow {
		return debounceWindow
	}
	return flushWindow
}

func parseBufferTime(value interface{}) *time.Time {
	raw := strings.TrimSpace(asString(value))
	if raw == "" {
		return nil
	}
	parsed, err := time.Parse(time.RFC3339Nano, raw)
	if err != nil {
		return nil
	}
	utc := parsed.UTC()
	return &utc
}

func asString(value interface{}) string {
	switch typed := value.(type) {
	case string:
		return typed
	case fmt.Stringer:
		return typed.String()
	default:
		return ""
	}
}

func asInt(value interface{}) int {
	switch typed := value.(type) {
	case int:
		return typed
	case int32:
		return int(typed)
	case int64:
		return int(typed)
	case float32:
		return int(typed)
	case float64:
		return int(typed)
	default:
		return 0
	}
}
