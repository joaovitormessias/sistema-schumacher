package trip_advances

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
  r.Route("/trip-advances", func(r chi.Router) {
    r.Get("/", h.list)
    r.Post("/", h.create)
    r.Get("/{advanceId}", h.get)
    r.Patch("/{advanceId}", h.update)
    r.Post("/{advanceId}/deliver", h.deliver)
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
    httpx.WriteError(w, http.StatusInternalServerError, "ADVANCES_LIST_ERROR", "could not list advances", err.Error())
    return
  }
  httpx.WriteJSON(w, http.StatusOK, items)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
  id, err := httpx.ParseUUIDParam(r, "advanceId")
  if err != nil {
    httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid advance id", nil)
    return
  }
  item, err := h.svc.Get(r.Context(), id)
  if err != nil {
    if IsNotFound(err) {
      httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "advance not found", nil)
      return
    }
    httpx.WriteError(w, http.StatusInternalServerError, "ADVANCE_GET_ERROR", "could not get advance", err.Error())
    return
  }
  httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
  var input CreateTripAdvanceInput
  if err := httpx.DecodeJSON(r, &input); err != nil {
    httpx.WriteError(w, http.StatusBadRequest, "INVALID_BODY", "invalid json", err.Error())
    return
  }
  if input.TripID == "" || input.DriverID == "" {
    httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "trip_id and driver_id are required", nil)
    return
  }
  if _, err := uuid.Parse(input.TripID); err != nil {
    httpx.WriteError(w, http.StatusBadRequest, "INVALID_TRIP", "invalid trip_id", nil)
    return
  }
  if _, err := uuid.Parse(input.DriverID); err != nil {
    httpx.WriteError(w, http.StatusBadRequest, "INVALID_DRIVER", "invalid driver_id", nil)
    return
  }
  if input.Amount < 0 {
    httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "amount must be positive", nil)
    return
  }

  var createdBy *string
  if userID, ok := auth.UserIDFromContext(r.Context()); ok {
    createdBy = &userID
  }

  item, err := h.svc.Create(r.Context(), input, createdBy)
  if err != nil {
    httpx.WriteError(w, http.StatusInternalServerError, "ADVANCE_CREATE_ERROR", "could not create advance", err.Error())
    return
  }
  httpx.WriteJSON(w, http.StatusCreated, item)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
  id, err := httpx.ParseUUIDParam(r, "advanceId")
  if err != nil {
    httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid advance id", nil)
    return
  }

  var input UpdateTripAdvanceInput
  if err := httpx.DecodeJSON(r, &input); err != nil {
    httpx.WriteError(w, http.StatusBadRequest, "INVALID_BODY", "invalid json", err.Error())
    return
  }

  item, err := h.svc.Update(r.Context(), id, input)
  if err != nil {
    if IsNotFound(err) {
      httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "advance not found", nil)
      return
    }
    if errors.Is(err, ErrAdvanceLocked) {
      httpx.WriteError(w, http.StatusConflict, "ADVANCE_LOCKED", "advance cannot be edited after delivery", nil)
      return
    }
    httpx.WriteError(w, http.StatusInternalServerError, "ADVANCE_UPDATE_ERROR", "could not update advance", err.Error())
    return
  }
  httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) deliver(w http.ResponseWriter, r *http.Request) {
  id, err := httpx.ParseUUIDParam(r, "advanceId")
  if err != nil {
    httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid advance id", nil)
    return
  }

  var deliveredBy *string
  if userID, ok := auth.UserIDFromContext(r.Context()); ok {
    deliveredBy = &userID
  }

  item, err := h.svc.Deliver(r.Context(), id, deliveredBy)
  if err != nil {
    if IsNotFound(err) {
      httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "advance not found", nil)
      return
    }
    if errors.Is(err, ErrAdvanceStatus) {
      httpx.WriteError(w, http.StatusBadRequest, "INVALID_STATUS", "advance is not pending", nil)
      return
    }
    httpx.WriteError(w, http.StatusInternalServerError, "ADVANCE_DELIVER_ERROR", "could not deliver advance", err.Error())
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
