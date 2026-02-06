package trip_validations

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
  r.Route("/trip-validations", func(r chi.Router) {
    r.Get("/", h.list)
    r.Post("/", h.create)
    r.Get("/{validationId}", h.get)
    r.Patch("/{validationId}", h.update)
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
    httpx.WriteError(w, http.StatusInternalServerError, "VALIDATIONS_LIST_ERROR", "could not list validations", err.Error())
    return
  }
  httpx.WriteJSON(w, http.StatusOK, items)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
  id, err := httpx.ParseUUIDParam(r, "validationId")
  if err != nil {
    httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid validation id", nil)
    return
  }
  item, err := h.svc.Get(r.Context(), id)
  if err != nil {
    if IsNotFound(err) {
      httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "validation not found", nil)
      return
    }
    httpx.WriteError(w, http.StatusInternalServerError, "VALIDATION_GET_ERROR", "could not get validation", err.Error())
    return
  }
  httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
  var input CreateTripValidationInput
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
  if input.OdometerInitial != nil && *input.OdometerInitial < 0 {
    httpx.WriteError(w, http.StatusBadRequest, "INVALID_ODOMETER", "odometer_initial must be positive", nil)
    return
  }
  if input.OdometerFinal != nil && *input.OdometerFinal < 0 {
    httpx.WriteError(w, http.StatusBadRequest, "INVALID_ODOMETER", "odometer_final must be positive", nil)
    return
  }
  if input.OdometerInitial != nil && input.OdometerFinal != nil && *input.OdometerFinal < *input.OdometerInitial {
    httpx.WriteError(w, http.StatusBadRequest, "INVALID_ODOMETER", "odometer_final must be >= odometer_initial", nil)
    return
  }
  if input.PassengersExpected < 0 || input.PassengersBoarded < 0 || input.PassengersNoShow < 0 {
    httpx.WriteError(w, http.StatusBadRequest, "INVALID_PASSENGERS", "passenger counts must be positive", nil)
    return
  }

  var validatedBy *string
  if userID, ok := auth.UserIDFromContext(r.Context()); ok {
    validatedBy = &userID
  }

  item, err := h.svc.Create(r.Context(), input, validatedBy)
  if err != nil {
    if errors.Is(err, ErrValidationExists) {
      httpx.WriteError(w, http.StatusConflict, "ALREADY_EXISTS", "validation already exists for trip", nil)
      return
    }
    httpx.WriteError(w, http.StatusInternalServerError, "VALIDATION_CREATE_ERROR", "could not create validation", err.Error())
    return
  }
  httpx.WriteJSON(w, http.StatusCreated, item)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
  id, err := httpx.ParseUUIDParam(r, "validationId")
  if err != nil {
    httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid validation id", nil)
    return
  }

  var input UpdateTripValidationInput
  if err := httpx.DecodeJSON(r, &input); err != nil {
    httpx.WriteError(w, http.StatusBadRequest, "INVALID_BODY", "invalid json", err.Error())
    return
  }
  if input.OdometerInitial != nil && *input.OdometerInitial < 0 {
    httpx.WriteError(w, http.StatusBadRequest, "INVALID_ODOMETER", "odometer_initial must be positive", nil)
    return
  }
  if input.OdometerFinal != nil && *input.OdometerFinal < 0 {
    httpx.WriteError(w, http.StatusBadRequest, "INVALID_ODOMETER", "odometer_final must be positive", nil)
    return
  }
  if input.OdometerInitial != nil && input.OdometerFinal != nil && *input.OdometerFinal < *input.OdometerInitial {
    httpx.WriteError(w, http.StatusBadRequest, "INVALID_ODOMETER", "odometer_final must be >= odometer_initial", nil)
    return
  }
  if input.PassengersExpected != nil && *input.PassengersExpected < 0 {
    httpx.WriteError(w, http.StatusBadRequest, "INVALID_PASSENGERS", "passengers_expected must be positive", nil)
    return
  }
  if input.PassengersBoarded != nil && *input.PassengersBoarded < 0 {
    httpx.WriteError(w, http.StatusBadRequest, "INVALID_PASSENGERS", "passengers_boarded must be positive", nil)
    return
  }
  if input.PassengersNoShow != nil && *input.PassengersNoShow < 0 {
    httpx.WriteError(w, http.StatusBadRequest, "INVALID_PASSENGERS", "passengers_no_show must be positive", nil)
    return
  }

  var validatedBy *string
  if userID, ok := auth.UserIDFromContext(r.Context()); ok {
    validatedBy = &userID
  }

  item, err := h.svc.Update(r.Context(), id, input, validatedBy)
  if err != nil {
    if IsNotFound(err) {
      httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "validation not found", nil)
      return
    }
    httpx.WriteError(w, http.StatusInternalServerError, "VALIDATION_UPDATE_ERROR", "could not update validation", err.Error())
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
