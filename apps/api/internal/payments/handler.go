package payments

import (
  "encoding/json"
  "fmt"
  "io"
  "net/http"
  "strconv"
  "strings"
  "time"

  httpx "schumacher-tur/api/internal/shared/http"

  "github.com/go-chi/chi/v5"
)

type Handler struct {
  svc       *Service
  publicKey string
  secret    string
}

func NewHandler(svc *Service, publicKey, secret string) *Handler {
  return &Handler{svc: svc, publicKey: publicKey, secret: secret}
}

// RegisterRoutes registers authenticated payment routes.
func (h *Handler) RegisterRoutes(r chi.Router) {
  r.Route("/payments", func(r chi.Router) {
    r.Get("/", h.list)
    r.Get("/{paymentId}/status", h.getStatus)
    r.Post("/", h.create)
    r.Post("/manual", h.createManual)
  })
}

// RegisterWebhooks registers public webhook routes.
func (h *Handler) RegisterWebhooks(r chi.Router) {
  r.Post("/webhooks/abacatepay", h.handleWebhook)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
  var input CreatePaymentInput
  if err := httpx.DecodeJSON(r, &input); err != nil {
    httpx.WriteError(w, http.StatusBadRequest, "INVALID_BODY", "invalid json", err.Error())
    return
  }
  if input.BookingID == "" || input.Amount <= 0 || input.Method == "" {
    httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "booking_id, amount and method are required", nil)
    return
  }
  if !isProviderMethod(input.Method) {
    httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid provider method", nil)
    return
  }

  payment, raw, err := h.svc.Create(r.Context(), input)
  if err != nil {
    httpx.WriteError(w, http.StatusInternalServerError, "PAYMENT_CREATE_ERROR", "could not create payment", err.Error())
    return
  }

  httpx.WriteJSON(w, http.StatusCreated, map[string]interface{}{
    "payment":      payment,
    "provider_raw": jsonRawToMap(raw),
  })
}

func (h *Handler) createManual(w http.ResponseWriter, r *http.Request) {
  var input ManualPaymentInput
  if err := httpx.DecodeJSON(r, &input); err != nil {
    httpx.WriteError(w, http.StatusBadRequest, "INVALID_BODY", "invalid json", err.Error())
    return
  }
  if input.BookingID == "" || input.Amount <= 0 || input.Method == "" {
    httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "booking_id, amount and method are required", nil)
    return
  }

  if !isManualMethod(input.Method) {
    httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid manual method", nil)
    return
  }

  payment, err := h.svc.CreateManual(r.Context(), input)
  if err != nil {
    httpx.WriteError(w, http.StatusInternalServerError, "PAYMENT_MANUAL_ERROR", "could not create manual payment", err.Error())
    return
  }
  httpx.WriteJSON(w, http.StatusCreated, payment)
}

func (h *Handler) getStatus(w http.ResponseWriter, r *http.Request) {
  id := chi.URLParam(r, "paymentId")
  if id == "" {
    httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid payment id", nil)
    return
  }
  status, err := h.svc.GetStatus(r.Context(), id)
  if err != nil {
    httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "payment not found", nil)
    return
  }
  httpx.WriteJSON(w, http.StatusOK, status)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
  filter := PaymentListFilter{}
  q := r.URL.Query()

  if v := q.Get("booking_id"); v != "" {
    filter.BookingID = v
  }
  if v := q.Get("status"); v != "" {
    if !isValidPaymentStatus(v) {
      httpx.WriteError(w, http.StatusBadRequest, "INVALID_STATUS", "invalid status parameter", nil)
      return
    }
    filter.Status = v
  }

  if v := q.Get("since"); v != "" {
    parsed, err := parseTimeParam(v)
    if err != nil {
      httpx.WriteError(w, http.StatusBadRequest, "INVALID_DATE", "invalid since parameter", err.Error())
      return
    }
    filter.Since = parsed
  }
  if v := q.Get("until"); v != "" {
    parsed, err := parseTimeParam(v)
    if err != nil {
      httpx.WriteError(w, http.StatusBadRequest, "INVALID_DATE", "invalid until parameter", err.Error())
      return
    }
    filter.Until = parsed
  }
  if v := q.Get("paid_since"); v != "" {
    parsed, err := parseTimeParam(v)
    if err != nil {
      httpx.WriteError(w, http.StatusBadRequest, "INVALID_DATE", "invalid paid_since parameter", err.Error())
      return
    }
    filter.PaidSince = parsed
  }
  if v := q.Get("paid_until"); v != "" {
    parsed, err := parseTimeParam(v)
    if err != nil {
      httpx.WriteError(w, http.StatusBadRequest, "INVALID_DATE", "invalid paid_until parameter", err.Error())
      return
    }
    filter.PaidUntil = parsed
  }

  if v := q.Get("limit"); v != "" {
    n, err := strconv.Atoi(v)
    if err != nil || n <= 0 {
      httpx.WriteError(w, http.StatusBadRequest, "INVALID_LIMIT", "invalid limit parameter", nil)
      return
    }
    filter.Limit = n
  }
  if v := q.Get("offset"); v != "" {
    n, err := strconv.Atoi(v)
    if err != nil || n < 0 {
      httpx.WriteError(w, http.StatusBadRequest, "INVALID_OFFSET", "invalid offset parameter", nil)
      return
    }
    filter.Offset = n
  }

  items, err := h.svc.List(r.Context(), filter)
  if err != nil {
    httpx.WriteError(w, http.StatusInternalServerError, "PAYMENT_LIST_ERROR", "could not list payments", err.Error())
    return
  }

  httpx.WriteJSON(w, http.StatusOK, items)
}

func (h *Handler) handleWebhook(w http.ResponseWriter, r *http.Request) {
  if h.secret == "" && h.publicKey == "" {
    httpx.WriteError(w, http.StatusServiceUnavailable, "WEBHOOK_NOT_CONFIGURED", "webhook security not configured", nil)
    return
  }

  body, _ := io.ReadAll(r.Body)

  if h.secret != "" {
    if !secretMatches(r, h.secret) {
      httpx.WriteError(w, http.StatusUnauthorized, "INVALID_SECRET", "invalid webhook secret", nil)
      return
    }
  }

  if h.publicKey != "" {
    sig := r.Header.Get("X-Webhook-Signature")
    if !VerifyWebhookSignature(h.publicKey, body, sig) {
      httpx.WriteError(w, http.StatusUnauthorized, "INVALID_SIGNATURE", "invalid webhook signature", nil)
      return
    }
  }

  evt, err := ParseWebhook(body)
  if err != nil {
    httpx.WriteError(w, http.StatusBadRequest, "INVALID_BODY", "invalid webhook body", err.Error())
    return
  }

  if err := h.svc.HandleWebhook(r.Context(), evt); err != nil {
    httpx.WriteError(w, http.StatusInternalServerError, "WEBHOOK_ERROR", "could not handle webhook", err.Error())
    return
  }

  httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func jsonRawToMap(raw []byte) interface{} {
  if len(raw) == 0 {
    return nil
  }
  var out interface{}
  _ = json.Unmarshal(raw, &out)
  return out
}

func secretMatches(r *http.Request, secret string) bool {
  if v := r.Header.Get("X-Webhook-Secret"); v != "" {
    return v == secret
  }
  auth := r.Header.Get("Authorization")
  if strings.HasPrefix(strings.ToLower(auth), "bearer ") {
    return strings.TrimSpace(auth[7:]) == secret
  }
  return false
}

func parseTimeParam(value string) (*time.Time, error) {
  if value == "" {
    return nil, nil
  }
  if t, err := time.Parse(time.RFC3339, value); err == nil {
    return &t, nil
  }
  if t, err := time.Parse("2006-01-02", value); err == nil {
    return &t, nil
  }
  return nil, fmt.Errorf("invalid time format")
}

func isProviderMethod(method string) bool {
  switch method {
  case "PIX", "CARD":
    return true
  default:
    return false
  }
}

func isManualMethod(method string) bool {
  switch method {
  case "CASH", "TRANSFER", "OTHER":
    return true
  default:
    return false
  }
}

func isValidPaymentStatus(status string) bool {
  switch status {
  case "PENDING", "PAID", "FAILED", "REFUNDED", "CANCELLED":
    return true
  default:
    return false
  }
}
