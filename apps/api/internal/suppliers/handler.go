package suppliers

import (
	"errors"
	"net/http"
	"strconv"

	httpx "schumacher-tur/api/internal/shared/http"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/suppliers", func(r chi.Router) {
		r.Get("/", h.list)
		r.Post("/", h.create)
		r.Get("/{supplierId}", h.get)
		r.Patch("/{supplierId}", h.update)
		r.Delete("/{supplierId}", h.delete)
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
		httpx.WriteError(w, http.StatusInternalServerError, "SUPPLIERS_LIST_ERROR", "could not list suppliers", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusOK, items)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.ParseUUIDParam(r, "supplierId")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid supplier id", nil)
		return
	}
	item, err := h.svc.Get(r.Context(), id)
	if err != nil {
		if IsNotFound(err) {
			httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "supplier not found", nil)
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "SUPPLIER_GET_ERROR", "could not get supplier", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var input CreateSupplierInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_BODY", "invalid json", err.Error())
		return
	}
	if input.Name == "" {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "name is required", nil)
		return
	}

	item, err := h.svc.Create(r.Context(), input)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "SUPPLIER_CREATE_ERROR", "could not create supplier", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, item)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.ParseUUIDParam(r, "supplierId")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid supplier id", nil)
		return
	}

	var input UpdateSupplierInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_BODY", "invalid json", err.Error())
		return
	}

	item, err := h.svc.Update(r.Context(), id, input)
	if err != nil {
		if IsNotFound(err) {
			httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "supplier not found", nil)
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "SUPPLIER_UPDATE_ERROR", "could not update supplier", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.ParseUUIDParam(r, "supplierId")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid supplier id", nil)
		return
	}

	if err := h.svc.Delete(r.Context(), id); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "SUPPLIER_DELETE_ERROR", "could not delete supplier", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusNoContent, nil)
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
	if v := q.Get("active"); v != "" {
		active := v == "true"
		filter.Active = &active
	}
	return filter, nil
}
