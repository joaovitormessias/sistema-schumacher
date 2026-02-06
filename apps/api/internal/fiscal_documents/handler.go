package fiscal_documents

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
  r.Route("/fiscal-documents", func(r chi.Router) {
    r.Get("/", h.list)
    r.Post("/", h.create)
    r.Get("/{documentId}", h.get)
    r.Patch("/{documentId}", h.update)
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
    httpx.WriteError(w, http.StatusInternalServerError, "DOCS_LIST_ERROR", "could not list documents", err.Error())
    return
  }
  httpx.WriteJSON(w, http.StatusOK, items)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
  id, err := httpx.ParseUUIDParam(r, "documentId")
  if err != nil {
    httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid document id", nil)
    return
  }
  item, err := h.svc.Get(r.Context(), id)
  if err != nil {
    if IsNotFound(err) {
      httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "document not found", nil)
      return
    }
    httpx.WriteError(w, http.StatusInternalServerError, "DOC_GET_ERROR", "could not get document", err.Error())
    return
  }
  httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
  var input CreateFiscalDocumentInput
  if err := httpx.DecodeJSON(r, &input); err != nil {
    httpx.WriteError(w, http.StatusBadRequest, "INVALID_BODY", "invalid json", err.Error())
    return
  }
  if input.TripID == "" || input.DocumentType == "" {
    httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "trip_id and document_type are required", nil)
    return
  }
  if _, err := uuid.Parse(input.TripID); err != nil {
    httpx.WriteError(w, http.StatusBadRequest, "INVALID_TRIP", "invalid trip_id", nil)
    return
  }
  if input.Amount < 0 {
    httpx.WriteError(w, http.StatusBadRequest, "INVALID_AMOUNT", "amount must be positive", nil)
    return
  }

  var createdBy *string
  if userID, ok := auth.UserIDFromContext(r.Context()); ok {
    createdBy = &userID
  }

  item, err := h.svc.Create(r.Context(), input, createdBy)
  if err != nil {
    httpx.WriteError(w, http.StatusInternalServerError, "DOC_CREATE_ERROR", "could not create document", err.Error())
    return
  }
  httpx.WriteJSON(w, http.StatusCreated, item)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
  id, err := httpx.ParseUUIDParam(r, "documentId")
  if err != nil {
    httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid document id", nil)
    return
  }

  var input UpdateFiscalDocumentInput
  if err := httpx.DecodeJSON(r, &input); err != nil {
    httpx.WriteError(w, http.StatusBadRequest, "INVALID_BODY", "invalid json", err.Error())
    return
  }
  if input.Amount != nil && *input.Amount < 0 {
    httpx.WriteError(w, http.StatusBadRequest, "INVALID_AMOUNT", "amount must be positive", nil)
    return
  }

  item, err := h.svc.Update(r.Context(), id, input)
  if err != nil {
    if IsNotFound(err) {
      httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "document not found", nil)
      return
    }
    httpx.WriteError(w, http.StatusInternalServerError, "DOC_UPDATE_ERROR", "could not update document", err.Error())
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
  if v := q.Get("document_type"); v != "" {
    filter.DocumentType = v
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
