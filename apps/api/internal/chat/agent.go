package chat

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	messageStatusAutomationPending  = "AUTOMATION_PENDING"
	agentStatusReadyForAutomation   = "READY_FOR_AUTOMATION"
	messageStatusAutomationDraft    = "AUTOMATION_DRAFT"
	messageStatusAutomationReviewed = "AUTOMATION_REVIEWED"
	agentStatusDraftGenerated       = "DRAFT_GENERATED"
	agentStatusDraftReviewed        = "DRAFT_REVIEWED"
)

func selectReprocessCandidateMessages(history []Message) []Message {
	candidates := make([]Message, 0, 4)
	for i := len(history) - 1; i >= 0; i-- {
		message := history[i]
		if message.Direction == "OUTBOUND" {
			if strings.EqualFold(strings.TrimSpace(message.ProcessingStatus), messageStatusAutomationDraft) {
				continue
			}
			if len(candidates) > 0 {
				break
			}
			return nil
		}
		if message.Direction != "INBOUND" {
			continue
		}
		if !isReprocessableInboundStatus(message.ProcessingStatus) {
			if len(candidates) > 0 {
				break
			}
			continue
		}
		candidates = append([]Message{message}, candidates...)
	}
	return candidates
}

func isReprocessableInboundStatus(status string) bool {
	switch strings.ToUpper(strings.TrimSpace(status)) {
	case "", "RECEIVED", "BUFFERED_PENDING", messageStatusAutomationPending:
		return true
	default:
		return false
	}
}

func isAutomationDraftStatus(status string) bool {
	switch strings.ToUpper(strings.TrimSpace(status)) {
	case messageStatusAutomationDraft, messageStatusAutomationReviewed:
		return true
	default:
		return false
	}
}

func buildReprocessMemory(session Session, history []Message, candidates []Message, observedAt time.Time) map[string]interface{} {
	recent := history
	if len(recent) > 8 {
		recent = recent[len(recent)-8:]
	}

	items := make([]map[string]interface{}, 0, len(recent))
	lastCustomerMessage := ""
	lastAssistantMessage := ""
	for _, message := range recent {
		body := strings.TrimSpace(message.Body)
		items = append(items, map[string]interface{}{
			"id":                message.ID,
			"direction":         message.Direction,
			"body":              truncateBufferBody(body),
			"processing_status": message.ProcessingStatus,
			"created_at":        message.CreatedAt.UTC().Format(time.RFC3339Nano),
		})
		if message.Direction == "INBOUND" && body != "" {
			lastCustomerMessage = body
		}
		if message.Direction == "OUTBOUND" && body != "" {
			lastAssistantMessage = body
		}
	}

	return map[string]interface{}{
		"generated_at":             observedAt.UTC().Format(time.RFC3339Nano),
		"customer_phone":           session.CustomerPhone,
		"customer_name":            session.CustomerName,
		"current_turn_body":        joinCandidateBodies(candidates),
		"current_turn_message_ids": candidateMessageIDs(candidates),
		"last_customer_message":    lastCustomerMessage,
		"last_assistant_message":   lastAssistantMessage,
		"recent_messages":          items,
	}
}

func buildReprocessAgentState(session Session, candidates []Message, trigger string, observedAt time.Time, metadata map[string]interface{}) map[string]interface{} {
	state := map[string]interface{}{
		"status":                   agentStatusReadyForAutomation,
		"trigger":                  trigger,
		"requested_at":             observedAt.UTC().Format(time.RFC3339Nano),
		"handoff_status":           session.HandoffStatus,
		"current_turn_body":        joinCandidateBodies(candidates),
		"current_turn_message_ids": candidateMessageIDs(candidates),
	}
	for key, value := range metadata {
		state[key] = value
	}
	return state
}

func buildReprocessBufferState(metadata map[string]interface{}, candidates []Message, trigger string, observedAt time.Time) map[string]interface{} {
	buffer := cloneBuffer(metadata)
	buffer["status"] = bufferStatusIdle
	buffer["message_count"] = 0
	buffer["window_started_at"] = nil
	buffer["pending_until"] = nil
	buffer["agent_status"] = agentStatusReadyForAutomation
	buffer["reprocess_trigger"] = trigger
	buffer["reprocess_requested_at"] = observedAt.UTC().Format(time.RFC3339Nano)
	buffer["current_turn_message_ids"] = candidateMessageIDs(candidates)
	buffer["current_turn_body"] = truncateBufferBody(joinCandidateBodies(candidates))
	if len(candidates) == 0 {
		return buffer
	}

	last := candidates[len(candidates)-1]
	buffer["last_message_id"] = last.ID
	buffer["last_message_at"] = last.ReceivedAt.UTC().Format(time.RFC3339Nano)
	buffer["last_message_body"] = truncateBufferBody(last.Body)
	buffer["last_direction"] = last.Direction
	buffer["processing_status"] = messageStatusAutomationPending
	buffer["provider_message_id"] = last.ProviderMessageID

	return buffer
}

func candidateMessageIDs(messages []Message) []string {
	ids := make([]string, 0, len(messages))
	for _, message := range messages {
		ids = append(ids, message.ID)
	}
	return ids
}

func joinCandidateBodies(messages []Message) string {
	parts := make([]string, 0, len(messages))
	for _, message := range messages {
		body := strings.TrimSpace(message.Body)
		if body == "" {
			continue
		}
		parts = append(parts, body)
	}
	return strings.Join(parts, "\n")
}

func buildAgentDraftIdempotencyKey(sessionID string, messageIDs []string) string {
	return "chat-agent-draft-" + uuid.NewSHA1(uuid.Nil, []byte(sessionID+"|"+strings.Join(messageIDs, ","))).String()
}

func buildDraftGeneratedAgentState(metadata map[string]interface{}, candidates []Message, draftID string, run RunAgentResult, toolCalls []ToolCall, observedAt time.Time) map[string]interface{} {
	state := cloneNestedMetadataMap(metadata, "agent")
	state["status"] = agentStatusDraftGenerated
	state["draft_idempotency_key"] = draftID
	state["draft_generated_at"] = observedAt.UTC().Format(time.RFC3339Nano)
	state["draft_message_ids"] = candidateMessageIDs(candidates)
	state["draft_model"] = strings.TrimSpace(run.Model)
	state["provider_response_id"] = strings.TrimSpace(run.ProviderResponseID)
	state["tool_calls_count"] = len(toolCalls)
	if len(toolCalls) > 0 {
		state["tool_names"] = toolCallNames(toolCalls)
		state["tool_call_ids"] = toolCallIDs(toolCalls)
	}
	return state
}

func buildDraftGeneratedBufferState(metadata map[string]interface{}, candidates []Message, draftID string, observedAt time.Time) map[string]interface{} {
	buffer := cloneBuffer(metadata)
	buffer["status"] = bufferStatusIdle
	buffer["agent_status"] = agentStatusDraftGenerated
	buffer["draft_idempotency_key"] = draftID
	buffer["draft_generated_at"] = observedAt.UTC().Format(time.RFC3339Nano)
	buffer["current_turn_message_ids"] = candidateMessageIDs(candidates)
	return buffer
}

func buildDraftReviewedAgentState(metadata map[string]interface{}, draft Message, ownerUserID string, action string, observedAt time.Time) map[string]interface{} {
	state := cloneNestedMetadataMap(metadata, "agent")
	state["status"] = agentStatusDraftReviewed
	state["reviewed_draft_message_id"] = draft.ID
	state["reviewed_at"] = observedAt.UTC().Format(time.RFC3339Nano)
	state["reviewed_by_user_id"] = strings.TrimSpace(ownerUserID)
	state["review_action"] = strings.TrimSpace(action)
	return state
}

func buildAgentDraftPayload(session Session, candidates []Message, draftID string, systemPrompt string, userPrompt string, run RunAgentResult, tools agentToolContext, observedAt time.Time) (map[string]interface{}, map[string]interface{}) {
	payload := map[string]interface{}{
		"mode":                     "AUTOMATION_DRAFT",
		"draft_idempotency_key":    draftID,
		"generated_at":             observedAt.UTC().Format(time.RFC3339Nano),
		"current_turn_message_ids": candidateMessageIDs(candidates),
		"model":                    strings.TrimSpace(run.Model),
		"provider_response_id":     strings.TrimSpace(run.ProviderResponseID),
		"request_payload":          run.RequestPayload,
		"response_payload":         run.ResponsePayload,
		"system_prompt":            systemPrompt,
		"user_prompt":              userPrompt,
		"session_channel":          session.Channel,
		"contact_key":              session.ContactKey,
	}
	if len(tools.Calls) > 0 {
		payload["tool_calls"] = serializeToolCalls(tools.Calls)
	}
	if tools.Availability != nil || tools.Pricing != nil || tools.Booking != nil || tools.Payments != nil {
		toolContext := map[string]interface{}{}
		if tools.Availability != nil {
			toolContext[toolNameAvailabilitySearch] = buildAvailabilityToolResponsePayload(*tools.Availability)
		}
		if tools.Pricing != nil {
			toolContext[toolNamePricingQuote] = buildPricingQuoteResponsePayload(*tools.Pricing)
		}
		if tools.Booking != nil {
			toolContext[toolNameBookingLookup] = buildBookingLookupResponsePayload(*tools.Booking)
		}
		if tools.Payments != nil {
			toolContext[toolNamePaymentStatus] = buildPaymentStatusResponsePayload(*tools.Payments)
		}
		payload["tool_context"] = toolContext
	}
	normalized := map[string]interface{}{
		"mode":                     "AUTOMATION_DRAFT",
		"draft_idempotency_key":    draftID,
		"generated_at":             observedAt.UTC().Format(time.RFC3339Nano),
		"agent_status":             agentStatusDraftGenerated,
		"current_turn_message_ids": candidateMessageIDs(candidates),
		"model":                    strings.TrimSpace(run.Model),
		"provider_response_id":     strings.TrimSpace(run.ProviderResponseID),
		"tool_call_count":          len(tools.Calls),
	}
	if len(tools.Calls) > 0 {
		normalized["tool_names"] = toolCallNames(tools.Calls)
	}
	return payload, normalized
}

func cloneNestedMetadataMap(metadata map[string]interface{}, key string) map[string]interface{} {
	raw, ok := metadata[key].(map[string]interface{})
	if !ok || len(raw) == 0 {
		return map[string]interface{}{}
	}
	cloned := make(map[string]interface{}, len(raw))
	for nestedKey, value := range raw {
		cloned[nestedKey] = value
	}
	return cloned
}

func toolCallNames(toolCalls []ToolCall) []string {
	names := make([]string, 0, len(toolCalls))
	for _, item := range toolCalls {
		if name := strings.TrimSpace(item.ToolName); name != "" {
			names = append(names, name)
		}
	}
	return names
}

func toolCallIDs(toolCalls []ToolCall) []string {
	ids := make([]string, 0, len(toolCalls))
	for _, item := range toolCalls {
		if id := strings.TrimSpace(item.ID); id != "" {
			ids = append(ids, id)
		}
	}
	return ids
}

func serializeToolCalls(toolCalls []ToolCall) []map[string]interface{} {
	items := make([]map[string]interface{}, 0, len(toolCalls))
	for _, item := range toolCalls {
		payload := map[string]interface{}{
			"id":               item.ID,
			"tool_name":        item.ToolName,
			"status":           item.Status,
			"request_payload":  item.RequestPayload,
			"response_payload": item.ResponsePayload,
			"started_at":       item.StartedAt.UTC().Format(time.RFC3339Nano),
		}
		if item.ErrorCode != "" {
			payload["error_code"] = item.ErrorCode
		}
		if item.ErrorMessage != "" {
			payload["error_message"] = item.ErrorMessage
		}
		if item.FinishedAt != nil {
			payload["finished_at"] = item.FinishedAt.UTC().Format(time.RFC3339Nano)
		}
		items = append(items, payload)
	}
	return items
}
