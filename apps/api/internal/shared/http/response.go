package httpx

import (
  "encoding/json"
  "net/http"
)

type APIError struct {
  Code    string      `json:"code"`
  Message string      `json:"message"`
  Details interface{} `json:"details,omitempty"`
}

func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(status)
  _ = json.NewEncoder(w).Encode(data)
}

func WriteError(w http.ResponseWriter, status int, code, message string, details interface{}) {
  WriteJSON(w, status, APIError{Code: code, Message: message, Details: details})
}
