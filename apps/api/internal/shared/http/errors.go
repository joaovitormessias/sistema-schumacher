package httpx

import (
  "net/http"
)

func NotImplemented(w http.ResponseWriter, r *http.Request) {
  WriteError(w, http.StatusNotImplemented, "NOT_IMPLEMENTED", "Endpoint not implemented", nil)
}
