package chat

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"schumacher-tur/api/internal/shared/config"
)

func TestOpenAIRunnerRun(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/responses" {
			t.Fatalf("expected path /responses, got %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer sk-test" {
			t.Fatalf("expected bearer auth, got %s", got)
		}
		if got := r.Header.Get("X-Client-Request-Id"); got != "draft-1" {
			t.Fatalf("expected idempotency header draft-1, got %s", got)
		}

		var payload map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if payload["model"] != "gpt-test" {
			t.Fatalf("expected model gpt-test, got %v", payload["model"])
		}
		if payload["input"] != "user" {
			t.Fatalf("expected plain text input, got %#v", payload["input"])
		}

		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"id":          "resp_test_1",
			"output_text": "Resposta do Shabas",
		})
	}))
	defer server.Close()

	runner := NewOpenAIRunner(config.Config{
		OpenAIAPIKey: "sk-test",
		OpenAIModel:  "gpt-test",
	})
	runner.baseURL = server.URL

	result, err := runner.Run(context.Background(), RunAgentInput{
		SystemPrompt:   "system",
		UserPrompt:     "user",
		IdempotencyKey: "draft-1",
	})
	if err != nil {
		t.Fatalf("run agent: %v", err)
	}
	if result.ReplyText != "Resposta do Shabas" {
		t.Fatalf("expected reply text, got %s", result.ReplyText)
	}
	if result.ProviderResponseID != "resp_test_1" {
		t.Fatalf("expected provider response id resp_test_1, got %s", result.ProviderResponseID)
	}
}

func TestOpenAIRunnerRunIncludesImageInputWhenCurrentTurnHasMedia(t *testing.T) {
	imageServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/jpeg")
		_, _ = w.Write([]byte("fake-image-binary"))
	}))
	defer imageServer.Close()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if payload["model"] != "gpt-vision-test" {
			t.Fatalf("expected vision model gpt-vision-test, got %v", payload["model"])
		}

		items, ok := payload["input"].([]interface{})
		if !ok || len(items) != 1 {
			t.Fatalf("expected structured input with one message, got %#v", payload["input"])
		}
		message, ok := items[0].(map[string]interface{})
		if !ok {
			t.Fatalf("expected input message map, got %#v", items[0])
		}
		content, ok := message["content"].([]interface{})
		if !ok || len(content) != 2 {
			t.Fatalf("expected text + image content, got %#v", message["content"])
		}
		image, ok := content[1].(map[string]interface{})
		expectedDataURL := "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString([]byte("fake-image-binary"))
		if !ok || image["type"] != "input_image" || image["image_url"] != expectedDataURL {
			t.Fatalf("unexpected image payload: %#v", content[1])
		}

		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"id":          "resp_test_2",
			"output_text": "Dados extraidos",
		})
	}))
	defer server.Close()

	runner := NewOpenAIRunner(config.Config{
		OpenAIAPIKey:      "sk-test",
		OpenAIModel:       "gpt-test",
		OpenAIVisionModel: "gpt-vision-test",
	})
	runner.baseURL = server.URL

	result, err := runner.Run(context.Background(), RunAgentInput{
		SystemPrompt: "system",
		UserPrompt:   "user",
		CurrentTurnMedia: []AgentMediaInput{
			{Kind: "IMAGE", URL: imageServer.URL, MimeType: "image/jpeg"},
		},
	})
	if err != nil {
		t.Fatalf("run agent: %v", err)
	}
	if result.ReplyText != "Dados extraidos" {
		t.Fatalf("expected image reply text, got %s", result.ReplyText)
	}
	if result.Model != "gpt-vision-test" {
		t.Fatalf("expected result model gpt-vision-test, got %s", result.Model)
	}
}

func TestOpenAIRunnerRunFallsBackToTextWhenImageRequestFails(t *testing.T) {
	var calls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if atomic.AddInt32(&calls, 1) == 1 {
			if _, ok := payload["input"].([]interface{}); !ok {
				t.Fatalf("expected first request to be multimodal, got %#v", payload["input"])
			}
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": map[string]interface{}{"message": "unsupported image"}})
			return
		}
		if payload["input"] != "user" {
			t.Fatalf("expected fallback request to be text-only, got %#v", payload["input"])
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"id":          "resp_test_3",
			"output_text": "Recebi a foto do documento.",
		})
	}))
	defer server.Close()

	runner := NewOpenAIRunner(config.Config{
		OpenAIAPIKey: "sk-test",
		OpenAIModel:  "gpt-test",
	})
	runner.baseURL = server.URL

	result, err := runner.Run(context.Background(), RunAgentInput{
		SystemPrompt: "system",
		UserPrompt:   "user",
		CurrentTurnMedia: []AgentMediaInput{
			{Kind: "IMAGE", URL: "https://files.example.test/rg.jpg", MimeType: "image/jpeg"},
		},
	})
	if err != nil {
		t.Fatalf("run agent: %v", err)
	}
	if atomic.LoadInt32(&calls) != 2 {
		t.Fatalf("expected two requests after fallback, got %d", calls)
	}
	if result.ReplyText != "Recebi a foto do documento." {
		t.Fatalf("unexpected fallback reply text: %s", result.ReplyText)
	}
}
