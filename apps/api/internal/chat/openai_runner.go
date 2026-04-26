package chat

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"schumacher-tur/api/internal/shared/config"
)

var (
	ErrOpenAIRunnerNotConfigured = errors.New("openai runner not configured")
	ErrOpenAIRunFailed           = errors.New("openai agent run failed")
	ErrOpenAIEmptyReply          = errors.New("openai agent returned empty reply")
)

type OpenAIRunner struct {
	baseURL     string
	apiKey      string
	model       string
	visionModel string
	client      *http.Client
}

func NewOpenAIRunner(cfg config.Config) *OpenAIRunner {
	return &OpenAIRunner{
		baseURL:     "https://api.openai.com/v1",
		apiKey:      strings.TrimSpace(cfg.OpenAIAPIKey),
		model:       strings.TrimSpace(cfg.OpenAIModel),
		visionModel: strings.TrimSpace(cfg.OpenAIVisionModel),
		client: &http.Client{
			Timeout: 45 * time.Second,
		},
	}
}

func (r *OpenAIRunner) Enabled() bool {
	return r != nil && r.apiKey != "" && r.model != ""
}

func (r *OpenAIRunner) Run(ctx context.Context, input RunAgentInput) (RunAgentResult, error) {
	if !r.Enabled() {
		return RunAgentResult{}, ErrOpenAIRunnerNotConfigured
	}

	requestPayload := map[string]interface{}{
		"model":        r.requestModel(input),
		"instructions": input.SystemPrompt,
		"input":        buildOpenAIInputContent(input),
	}
	result, err := r.runRequest(ctx, requestPayload, input.IdempotencyKey)
	if err == nil {
		return result, nil
	}
	if len(input.CurrentTurnMedia) == 0 {
		return RunAgentResult{}, err
	}

	fallbackPayload := map[string]interface{}{
		"model":        r.model,
		"instructions": input.SystemPrompt,
		"input":        input.UserPrompt,
	}
	return r.runRequest(ctx, fallbackPayload, input.IdempotencyKey)
}

func (r *OpenAIRunner) requestModel(input RunAgentInput) string {
	if len(input.CurrentTurnMedia) > 0 {
		if model := strings.TrimSpace(r.visionModel); model != "" {
			return model
		}
	}
	return r.model
}

func (r *OpenAIRunner) runRequest(ctx context.Context, requestPayload map[string]interface{}, idempotencyKey string) (RunAgentResult, error) {
	body, err := json.Marshal(requestPayload)
	if err != nil {
		return RunAgentResult{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(r.baseURL, "/")+"/responses", bytes.NewReader(body))
	if err != nil {
		return RunAgentResult{}, err
	}
	req.Header.Set("Authorization", "Bearer "+r.apiKey)
	req.Header.Set("Content-Type", "application/json")
	if trimmed := strings.TrimSpace(idempotencyKey); trimmed != "" {
		req.Header.Set("X-Client-Request-Id", trimmed)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return RunAgentResult{}, fmt.Errorf("%w: %v", ErrOpenAIRunFailed, err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return RunAgentResult{}, fmt.Errorf("%w: %v", ErrOpenAIRunFailed, err)
	}

	responsePayload := map[string]interface{}{}
	if len(responseBody) > 0 {
		_ = json.Unmarshal(responseBody, &responsePayload)
	}
	if resp.StatusCode >= 300 {
		return RunAgentResult{}, fmt.Errorf("%w: status %d", ErrOpenAIRunFailed, resp.StatusCode)
	}

	replyText := strings.TrimSpace(extractOpenAIResponseText(responsePayload))
	if replyText == "" {
		return RunAgentResult{}, ErrOpenAIEmptyReply
	}

	return RunAgentResult{
		ReplyText:          replyText,
		Model:              strings.TrimSpace(asString(requestPayload["model"])),
		ProviderResponseID: strings.TrimSpace(asString(responsePayload["id"])),
		RequestPayload:     requestPayload,
		ResponsePayload:    responsePayload,
	}, nil
}

func buildOpenAIInputContent(input RunAgentInput) interface{} {
	content := []map[string]interface{}{
		{
			"type": "input_text",
			"text": input.UserPrompt,
		},
	}
	for _, item := range input.CurrentTurnMedia {
		if !strings.EqualFold(strings.TrimSpace(item.Kind), "IMAGE") {
			continue
		}
		url := strings.TrimSpace(item.URL)
		if url == "" {
			continue
		}
		content = append(content, map[string]interface{}{
			"type":      "input_image",
			"image_url": url,
		})
	}
	if len(content) == 1 {
		return input.UserPrompt
	}
	return []map[string]interface{}{
		{
			"role":    "user",
			"content": content,
		},
	}
}

func extractOpenAIResponseText(payload map[string]interface{}) string {
	if text := strings.TrimSpace(asString(payload["output_text"])); text != "" {
		return text
	}

	rawOutput, ok := payload["output"].([]interface{})
	if !ok {
		return ""
	}
	parts := make([]string, 0, 2)
	for _, entry := range rawOutput {
		message, ok := entry.(map[string]interface{})
		if !ok {
			continue
		}
		contentItems, ok := message["content"].([]interface{})
		if !ok {
			continue
		}
		for _, rawContent := range contentItems {
			content, ok := rawContent.(map[string]interface{})
			if !ok {
				continue
			}
			if text := strings.TrimSpace(asString(content["text"])); text != "" {
				parts = append(parts, text)
				continue
			}
			if text := strings.TrimSpace(asString(content["output_text"])); text != "" {
				parts = append(parts, text)
			}
		}
	}
	return strings.Join(parts, "\n")
}
