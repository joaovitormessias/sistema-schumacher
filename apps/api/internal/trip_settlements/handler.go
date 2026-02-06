package trip_settlements

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
  r.Route("/trip-settlements", func(r chi.Router) {
    r.Get("/", h.list)
    r.Post("/", h.create)
    r.Get("/{settlementId}", h.get)
    r.Post("/{settlementId}/review", h.review)
    r.Post("/{settlementId}/approve", h.approve)
    r.Post("/{settlementId}/reject", h.reject)
    r.Post("/{settlementId}/complete", h.complete)
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
    httpx.WriteError(w, http.StatusInternalServerError, "SETTLEMENTS_LIST_ERROR", "could not list settlements", err.Error())
    return
  }
  httpx.WriteJSON(w, http.StatusOK, items)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
  id, err := httpx.ParseUUIDParam(r, "settlementId")
  if err != nil {
    httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid settlement id", nil)
    return
  }
  item, err := h.svc.Get(r.Context(), id)
  if err != nil {
    if IsNotFound(err) {
      httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "settlement not found", nil)
      return
    }
    httpx.WriteError(w, http.StatusInternalServerError, "SETTLEMENT_GET_ERROR", "could not get settlement", err.Error())
    return
  }
  httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
  var input CreateTripSettlementInput
  if err := httpx.DecodeJSON(r, &input); err != nil {
    httpx.WriteError(w, http.StatusBadRequest, "INVALID_BODY", "invalid json", err.Error())
    return
  }
  if input.TripID == "" {
    httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "trip_id is required", nil)
    return
  }
  if _, err := uuid.Parse(input.TripID); err != nil {
    httpx.WriteError(w, http.StatusBadRequest, "INVALID_TRIP", "invalid trip_id", nil)
    return
  }

  var createdBy *string
  if userID, ok := auth.UserIDFromContext(r.Context()); ok {
    createdBy = &userID
  }

  item, err := h.svc.Create(r.Context(), input, createdBy)
  if err != nil {
    if errors.Is(err, ErrSettlementExists) {
      httpx.WriteError(w, http.StatusConflict, "ALREADY_EXISTS", "settlement already exists for trip", nil)
      return
    }
    if errors.Is(err, ErrTripDriverMissing) {
      httpx.WriteError(w, http.StatusConflict, "MISSING_DRIVER", "trip has no driver assigned", nil)
      return
    }
    httpx.WriteError(w, http.StatusInternalServerError, "SETTLEMENT_CREATE_ERROR", "could not create settlement", err.Error())
    return
  }
  httpx.WriteJSON(w, http.StatusCreated, item)
}

func (h *Handler) review(w http.ResponseWriter, r *http.Request) {
  id, err := httpx.ParseUUIDParam(r, "settlementId")
  if err != nil {
    httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid settlement id", nil)
    return
  }

  var reviewedBy *string
  if userID, ok := auth.UserIDFromContext(r.Context()); ok {
    reviewedBy = &userID
  }

  item, err := h.svc.Review(r.Context(), id, reviewedBy)
  if err != nil {
    if IsNotFound(err) {
      httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "settlement not found", nil)
      return
    }
    if errors.Is(err, ErrInvalidStatus) {
      httpx.WriteError(w, http.StatusBadRequest, "INVALID_STATUS", "settlement is not in draft", nil)
      return
    }
    httpx.WriteError(w, http.StatusInternalServerError, "SETTLEMENT_REVIEW_ERROR", "could not move settlement to review", err.Error())
    return
  }
  httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) approve(w http.ResponseWriter, r *http.Request) {
  id, err := httpx.ParseUUIDParam(r, "settlementId")
  if err != nil {
    httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid settlement id", nil)
    return
  }

  var approvedBy *string
  if userID, ok := auth.UserIDFromContext(r.Context()); ok {
    approvedBy = &userID
  }

  item, err := h.svc.Approve(r.Context(), id, approvedBy)
  if err != nil {
    if IsNotFound(err) {
      httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "settlement not found", nil)
      return
    }
    if errors.Is(err, ErrInvalidStatus) {
      httpx.WriteError(w, http.StatusBadRequest, "INVALID_STATUS", "settlement is not under review", nil)
      return
    }
    httpx.WriteError(w, http.StatusInternalServerError, "SETTLEMENT_APPROVE_ERROR", "could not approve settlement", err.Error())
    return
  }
  httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) reject(w http.ResponseWriter, r *http.Request) {
  id, err := httpx.ParseUUIDParam(r, "settlementId")
  if err != nil {
    httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid settlement id", nil)
    return
  }

  var reviewedBy *string
  if userID, ok := auth.UserIDFromContext(r.Context()); ok {
    reviewedBy = &userID
  }

  item, err := h.svc.Reject(r.Context(), id, reviewedBy)
  if err != nil {
    if IsNotFound(err) {
      httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "settlement not found", nil)
      return
    }
    if errors.Is(err, ErrInvalidStatus) {
      httpx.WriteError(w, http.StatusBadRequest, "INVALID_STATUS", "settlement is not under review", nil)
      return
    }
    httpx.WriteError(w, http.StatusInternalServerError, "SETTLEMENT_REJECT_ERROR", "could not reject settlement", err.Error())
    return
  }
  httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) complete(w http.ResponseWriter, r *http.Request) {
  id, err := httpx.ParseUUIDParam(r, "settlementId")
  if err != nil {
    httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid settlement id", nil)
    return
  }

  item, err := h.svc.Complete(r.Context(), id)
  if err != nil {
    if IsNotFound(err) {
      httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "settlement not found", nil)
      return
    }
    if errors.Is(err, ErrInvalidStatus) {
      httpx.WriteError(w, http.StatusBadRequest, "INVALID_STATUS", "settlement is not approved", nil)
      return
    }
    httpx.WriteError(w, http.StatusInternalServerError, "SETTLEMENT_COMPLETE_ERROR", "could not complete settlement", err.Error())
    return
  }
  httpx.WriteJSON(w, http.StatusOK, item)
}

func parseListFilter(r *http.Request) (ListFilter, error) {
  filter := ListFilter{}
  q := r.URL.Query()

  if v := q.Get("trip_id"); v != "" {
    if _, err := uuid.Parse(v); err != nil {
      return filter, err
    }
    filter.TripID = v
  }
  if v := q.Get("driver_id"); v != "" {
    if _, err := uuid.Parse(v); err != nil {
      return filter, err
    }
    filter.DriverID = v
  }
  if v := q.Get("status"); v != "" {
    filter.Status = v
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
