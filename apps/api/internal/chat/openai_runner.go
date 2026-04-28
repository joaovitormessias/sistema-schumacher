package chat

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"schumacher-tur/api/internal/shared/config"
	"strings"
	"time"
)

var (
	ErrOpenAIRunnerNotConfigured = errors.New("openai runner not configured")
	ErrOpenAIRunFailed           = errors.New("openai agent run failed")
	ErrOpenAIEmptyReply          = errors.New("openai agent returned empty reply")
)

const maxInlineImageBytes = 8 * 1024 * 1024

type OpenAIRunner struct {
	baseURL     string
	apiKey      string
	model       string
	visionModel string
	client      *http.Client
}

func compactOpenAIErrorBody(body []byte) string {
	/* Sanitize the message error that comes from OpenAI */

	text := strings.TrimSpace(string(body))
	if text == "" {
		return ""
	}

	parsed := map[string]interface{}{}
	if err := json.Unmarshal(body, &parsed); err == nil {
		if errMap, ok := parsed["error"].(map[string]interface{}); ok {
			message := strings.TrimSpace(asString(errMap["message"]))
			errType := strings.TrimSpace(asString(errMap["type"]))
			code := strings.TrimSpace(asString(errMap["code"]))

			parts := []string{}
			if errType != "" {
				parts = append(parts, "type="+errType)
			}
			if code != "" {
				parts = append(parts, "code="+code)
			}
			if message != "" {
				parts = append(parts, "message="+message)
			}
			if len(parts) > 0 {
				return strings.Join(parts, " ")
			}

		}
	}

	if len(text) > 1500 {
		return text[:1500]
	}
	return text
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
		"input":        buildOpenAIInputContent(ctx, r.client, input),
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
		errorBody := compactOpenAIErrorBody(responseBody)
		if errorBody != "" {
			return RunAgentResult{}, fmt.Errorf(
				"%w: status %d %s",
				ErrOpenAIRunFailed,
				resp.StatusCode,
				errorBody,
			)
		}
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

func buildOpenAIInputContent(ctx context.Context, client *http.Client, input RunAgentInput) interface{} {
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
		imageURL := resolveOpenAIImageURL(ctx, client, item)
		if imageURL == "" {
			continue
		}
		content = append(content, map[string]interface{}{
			"type":      "input_image",
			"image_url": imageURL,
			"detail":    "high",
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

func resolveOpenAIImageURL(ctx context.Context, client *http.Client, item AgentMediaInput) string {
	url := strings.TrimSpace(item.URL)
	if url == "" {
		return ""
	}
	if strings.HasPrefix(strings.ToLower(url), "data:") {
		return url
	}
	if client == nil || !isHTTPURL(url) {
		return url
	}
	if dataURL, err := fetchImageAsDataURL(ctx, client, url, item.MimeType); err == nil {
		return dataURL
	}
	return url
}

func fetchImageAsDataURL(ctx context.Context, client *http.Client, url string, mimeType string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxInlineImageBytes+1))
	if err != nil {
		return "", err
	}
	if len(body) == 0 {
		return "", errors.New("empty image body")
	}
	if len(body) > maxInlineImageBytes {
		return "", errors.New("image too large to inline")
	}

	resolvedMime := strings.TrimSpace(mimeType)
	if resolvedMime == "" {
		resolvedMime = strings.TrimSpace(strings.Split(resp.Header.Get("Content-Type"), ";")[0])
	}
	if resolvedMime == "" {
		resolvedMime = http.DetectContentType(body)
	}

	return "data:" + resolvedMime + ";base64," + base64.StdEncoding.EncodeToString(body), nil
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
