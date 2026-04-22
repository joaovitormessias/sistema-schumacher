package chat

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
