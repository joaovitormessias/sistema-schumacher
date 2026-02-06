package httpx

import (
  "bytes"
  "net/http/httptest"
  "testing"
)

type payload struct {
  Name string `json:"name"`
}

func TestDecodeJSONValid(t *testing.T) {
  req := httptest.NewRequest("POST", "/", bytes.NewBufferString(`{"name":"ok"}`))
  var out payload
  if err := DecodeJSON(req, &out); err != nil {
    t.Fatalf("unexpected error: %v", err)
  }
  if out.Name != "ok" {
    t.Fatalf("expected name to be ok, got %q", out.Name)
  }
}

func TestDecodeJSONEmpty(t *testing.T) {
  req := httptest.NewRequest("POST", "/", bytes.NewBufferString(""))
  var out payload
  if err := DecodeJSON(req, &out); err != ErrEmptyBody {
    t.Fatalf("expected ErrEmptyBody, got %v", err)
  }
}

func TestDecodeJSONUnknownField(t *testing.T) {
  req := httptest.NewRequest("POST", "/", bytes.NewBufferString(`{"name":"ok","extra":1}`))
  var out payload
  if err := DecodeJSON(req, &out); err == nil {
    t.Fatalf("expected error for unknown field")
  }
}

func TestDecodeJSONTrailingData(t *testing.T) {
  req := httptest.NewRequest("POST", "/", bytes.NewBufferString(`{"name":"ok"}{"name":"no"}`))
  var out payload
  if err := DecodeJSON(req, &out); err == nil {
    t.Fatalf("expected error for trailing data")
  }
}
