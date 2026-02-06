package advance_returns

import (
  "errors"
  "net/http"
  "strconv"

  "github.com/go-chi/chi/v5"
  "github.com/google/uuid"

  "schumacher-tur/api/internal/auth"
  httpx "schumacher-tur/api/internal/shared/http"
)

type Handler struct {
  svc *Service
}

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) RegisterRoutes(r chi.Router) {
  r.Route("/advance-returns", func(r chi.Router) {
    r.Get("/", h.list)
    r.Post("/", h.create)
    r.Get("/{returnId}", h.get)
  })
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
  filter, err := parseListFilter(r)
  if err != nil {
    httpx.WriteError(w, http.StatusBadRequest, "INVALID_QUERY", "invalid query parameters", nil)
    return
  }
  items, err := h.svc.List(r.Context(), filter)
  if err != nil {
    httpx.WriteError(w, http.StatusInternalServerError, "RETURNS_LIST_ERROR", "could not list returns", err.Error())
    return
  }
  httpx.WriteJSON(w, http.StatusOK, items)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
  id, err := httpx.ParseUUIDParam(r, "returnId")
  if err != nil {
    httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid return id", nil)
    return
  }
  item, err := h.svc.Get(r.Context(), id)
  if err != nil {
    if IsNotFound(err) {
      httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "return not found", nil)
      return
    }
    httpx.WriteError(w, http.StatusInternalServerError, "RETURN_GET_ERROR", "could not get return", err.Error())
    return
  }
  httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
  var input CreateAdvanceReturnInput
  if err := httpx.DecodeJSON(r, &input); err != nil {
    httpx.WriteError(w, http.StatusBadRequest, "INVALID_BODY", "invalid json", err.Error())
    return
  }
  if input.TripAdvanceID == "" || input.PaymentMethod == "" {
    httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "trip_advance_id and payment_method are required", nil)
    return
  }
  if _, err := uuid.Parse(input.TripAdvanceID); err != nil {
    httpx.WriteError(w, http.StatusBadRequest, "INVALID_ADVANCE", "invalid trip_advance_id", nil)
    return
  }
  if input.TripSettlementID != nil {
    if _, err := uuid.Parse(*input.TripSettlementID); err != nil {
      httpx.WriteError(w, http.StatusBadRequest, "INVALID_SETTLEMENT", "invalid trip_settlement_id", nil)
      return
    }
  }
  if input.Amount < 0 {
    httpx.WriteError(w, http.StatusBadRequest, "INVALID_AMOUNT", "amount must be positive", nil)
    return
  }

  var receivedBy *string
  if userID, ok := auth.UserIDFromContext(r.Context()); ok {
    receivedBy = &userID
  }

  item, err := h.svc.Create(r.Context(), input, receivedBy)
  if err != nil {
    httpx.WriteError(w, http.StatusInternalServerError, "RETURN_CREATE_ERROR", "could not create return", err.Error())
    return
  }
  httpx.WriteJSON(w, http.StatusCreated, item)
}

func parseListFilter(r *http.Request) (ListFilter, error) {
  filter := ListFilter{}
  q := r.URL.Query()

  if v := q.Get("trip_advance_id"); v != "" {
    if _, err := uuid.Parse(v); err != nil {
      return filter, err
    }
    filter.TripAdvanceID = v
  }
  if v := q.Get("trip_settlement_id"); v != "" {
    if _, err := uuid.Parse(v); err != nil {
      return filter, err
    }
    filter.TripSettlementID = v
  }
  if v := q.Get("limit"); v != "" {
    n, err := strconv.Atoi(v)
    if err != nil || n <= 0 {
      return filter, errors.New("invalid limit")
    }
    filter.Limit = n
  }
  if v := q.Get("offset"); v != "" {
    n, err := strconv.Atoi(v)
    if err != nil || n < 0 {
      return filter, errors.New("invalid offset")
    }
    filter.Offset = n
  }

  return filter, nil
}
