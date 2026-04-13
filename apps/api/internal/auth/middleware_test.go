package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAudMatchesString(t *testing.T) {
	if !audMatches("authenticated", "authenticated") {
		t.Fatalf("expected audience match")
	}
	if audMatches("authenticated", "other") {
		t.Fatalf("expected audience mismatch")
	}
}

func TestAudMatchesSlice(t *testing.T) {
	aud := []interface{}{"other", "authenticated"}
	if !audMatches("authenticated", aud) {
		t.Fatalf("expected audience match in slice")
	}
}

func TestMiddlewareSkip(t *testing.T) {
	auth, err := NewAuthenticator("", "", "", nil, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	handler := auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected %d, got %d", http.StatusNoContent, rr.Code)
	}
}

func TestMiddlewareServiceToken(t *testing.T) {
	subject := serviceSubject("service-token-123")
	auth := &Authenticator{
		services: map[string]string{
			"service-token-123": subject,
		},
	}

	var gotUserID string
	handler := auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := UserIDFromContext(r.Context())
		if !ok {
			t.Fatalf("expected user id in context")
		}
		gotUserID = userID
		w.WriteHeader(http.StatusNoContent)
	}))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer service-token-123")
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected %d, got %d", http.StatusNoContent, rr.Code)
	}
	if gotUserID != subject {
		t.Fatalf("expected subject %q, got %q", subject, gotUserID)
	}
}
