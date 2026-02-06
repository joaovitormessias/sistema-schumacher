package buses

import (
  "errors"
  "net/http"
  "strconv"

  "github.com/go-chi/chi/v5"
  httpx "schumacher-tur/api/internal/shared/http"
)

type Handler struct {
  svc *Service
}

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) RegisterRoutes(r chi.Router) {
  r.Route("/buses", func(r chi.Router) {
    r.Get("/", h.list)
    r.Post("/", h.create)
    r.Patch("/{busId}", h.update)
    r.Get("/{busId}", h.get)
  })
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
  filter, err := parseListFilter(r)
  if err != nil {
    httpx.WriteError(w, http.StatusBadRequest, "INVALID_PAGINATION", "invalid pagination parameters", nil)
    return
  }
  items, err := h.svc.List(r.Context(), filter)
  if err != nil {
    httpx.WriteError(w, http.StatusInternalServerError, "BUSES_LIST_ERROR", "could not list buses", err.Error())
    return
  }
  httpx.WriteJSON(w, http.StatusOK, items)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
  id, err := httpx.ParseUUIDParam(r, "busId")
  if err != nil {
    httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid bus id", nil)
    return
  }
  item, err := h.svc.Get(r.Context(), id)
  if err != nil {
    if IsNotFound(err) {
      httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "bus not found", nil)
      return
    }
    httpx.WriteError(w, http.StatusInternalServerError, "BUS_GET_ERROR", "could not get bus", err.Error())
    return
  }
  httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
  var input CreateBusInput
  if err := httpx.DecodeJSON(r, &input); err != nil {
    httpx.WriteError(w, http.StatusBadRequest, "INVALID_BODY", "invalid json", err.Error())
    return
  }
  if input.Name == "" || input.Capacity <= 0 {
    httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "name and capacity are required", nil)
    return
  }

  item, err := h.svc.Create(r.Context(), input)
  if err != nil {
    httpx.WriteError(w, http.StatusInternalServerError, "BUS_CREATE_ERROR", "could not create bus", err.Error())
    return
  }
  httpx.WriteJSON(w, http.StatusCreated, item)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
  id, err := httpx.ParseUUIDParam(r, "busId")
  if err != nil {
    httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid bus id", nil)
    return
  }

  var input UpdateBusInput
  if err := httpx.DecodeJSON(r, &input); err != nil {
    httpx.WriteError(w, http.StatusBadRequest, "INVALID_BODY", "invalid json", err.Error())
    return
  }

  item, err := h.svc.Update(r.Context(), id, input)
  if err != nil {
    if errors.Is(err, ErrCapacityInUse) {
      httpx.WriteError(w, http.StatusBadRequest, "CAPACITY_IN_USE", "cannot reduce capacity; seats in use", nil)
      return
    }
    if IsNotFound(err) {
      httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "bus not found", nil)
      return
    }
    httpx.WriteError(w, http.StatusInternalServerError, "BUS_UPDATE_ERROR", "could not update bus", err.Error())
    return
  }
  httpx.WriteJSON(w, http.StatusOK, item)
}

func parseListFilter(r *http.Request) (ListFilter, error) {
  filter := ListFilter{}
  q := r.URL.Query()
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
