package invoices

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
	r.Route("/invoices", func(r chi.Router) {
		r.Get("/", h.list)
		r.Post("/", h.create)
		r.Get("/{invoiceId}", h.get)
		r.Patch("/{invoiceId}", h.update)
		r.Post("/{invoiceId}/process", h.process)
		r.Post("/{invoiceId}/cancel", h.cancel)
		r.Delete("/{invoiceId}", h.delete)
	})
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	filter, err := parseListFilter(r)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_PARAMS", err.Error(), nil)
		return
	}
	items, err := h.svc.List(r.Context(), filter)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INVOICES_LIST_ERROR", "could not list invoices", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusOK, items)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.ParseUUIDParam(r, "invoiceId")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid invoice id", nil)
		return
	}
	item, err := h.svc.Get(r.Context(), id)
	if err != nil {
		if IsNotFound(err) {
			httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "invoice not found", nil)
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "INVOICE_GET_ERROR", "could not get invoice", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var input CreateInvoiceInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_BODY", "invalid json", err.Error())
		return
	}
	if input.InvoiceNumber == "" {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invoice_number is required", nil)
		return
	}
	if input.SupplierID == "" {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "supplier_id is required", nil)
		return
	}

	item, err := h.svc.Create(r.Context(), input)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INVOICE_CREATE_ERROR", "could not create invoice", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, item)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.ParseUUIDParam(r, "invoiceId")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid invoice id", nil)
		return
	}

	var input UpdateInvoiceInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_BODY", "invalid json", err.Error())
		return
	}

	item, err := h.svc.Update(r.Context(), id, input)
	if err != nil {
		if IsNotFound(err) {
			httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "invoice not found", nil)
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "INVOICE_UPDATE_ERROR", "could not update invoice", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) process(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.ParseUUIDParam(r, "invoiceId")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid invoice id", nil)
		return
	}

	if err := h.svc.Process(r.Context(), id); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INVOICE_PROCESS_ERROR", "could not process invoice", err.Error())
		return
	}

	item, _ := h.svc.Get(r.Context(), id)
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) cancel(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.ParseUUIDParam(r, "invoiceId")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid invoice id", nil)
		return
	}

	if err := h.svc.Cancel(r.Context(), id); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INVOICE_CANCEL_ERROR", "could not cancel invoice", err.Error())
		return
	}

	item, _ := h.svc.Get(r.Context(), id)
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.ParseUUIDParam(r, "invoiceId")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid invoice id", nil)
		return
	}

	if err := h.svc.Delete(r.Context(), id); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INVOICE_DELETE_ERROR", "could not delete invoice", err.Error())
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
	if v := q.Get("status"); v != "" {
		status := InvoiceStatus(v)
		filter.Status = &status
	}
	if v := q.Get("supplier_id"); v != "" {
		filter.SupplierID = &v
	}
	if v := q.Get("service_order_id"); v != "" {
		filter.ServiceOrderID = &v
	}
	if v := q.Get("purchase_order_id"); v != "" {
		filter.PurchaseOrderID = &v
	}
	if v := q.Get("bus_id"); v != "" {
		filter.BusID = &v
	}
	return filter, nil
}
