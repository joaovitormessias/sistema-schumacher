package service_orders

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
	r.Route("/service-orders", func(r chi.Router) {
		r.Get("/", h.list)
		r.Post("/", h.create)
		r.Get("/{orderId}", h.get)
		r.Patch("/{orderId}", h.update)
		r.Post("/{orderId}/start", h.start)
		r.Post("/{orderId}/close", h.close)
		r.Post("/{orderId}/cancel", h.cancel)
		r.Delete("/{orderId}", h.delete)
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
		httpx.WriteError(w, http.StatusInternalServerError, "SERVICE_ORDERS_LIST_ERROR", "could not list service orders", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusOK, items)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.ParseUUIDParam(r, "orderId")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid order id", nil)
		return
	}
	item, err := h.svc.Get(r.Context(), id)
	if err != nil {
		if IsNotFound(err) {
			httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "service order not found", nil)
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "SERVICE_ORDER_GET_ERROR", "could not get service order", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var input CreateServiceOrderInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_BODY", "invalid json", err.Error())
		return
	}
	if input.BusID == "" {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "bus_id is required", nil)
		return
	}
	if input.OrderType == "" {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "order_type is required", nil)
		return
	}
	if input.Description == "" {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "description is required", nil)
		return
	}

	item, err := h.svc.Create(r.Context(), input)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "SERVICE_ORDER_CREATE_ERROR", "could not create service order", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, item)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.ParseUUIDParam(r, "orderId")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid order id", nil)
		return
	}

	var input UpdateServiceOrderInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_BODY", "invalid json", err.Error())
		return
	}

	item, err := h.svc.Update(r.Context(), id, input)
	if err != nil {
		if IsNotFound(err) {
			httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "service order not found", nil)
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "SERVICE_ORDER_UPDATE_ERROR", "could not update service order", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) start(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.ParseUUIDParam(r, "orderId")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid order id", nil)
		return
	}

	if err := h.svc.StartProgress(r.Context(), id); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "SERVICE_ORDER_START_ERROR", "could not start service order", err.Error())
		return
	}

	item, _ := h.svc.Get(r.Context(), id)
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) close(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.ParseUUIDParam(r, "orderId")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid order id", nil)
		return
	}

	var input CloseServiceOrderInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_BODY", "invalid json", err.Error())
		return
	}

	item, err := h.svc.Close(r.Context(), id, input)
	if err != nil {
		if IsNotFound(err) {
			httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "service order not found", nil)
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "SERVICE_ORDER_CLOSE_ERROR", "could not close service order", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) cancel(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.ParseUUIDParam(r, "orderId")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid order id", nil)
		return
	}

	if err := h.svc.Cancel(r.Context(), id); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "SERVICE_ORDER_CANCEL_ERROR", "could not cancel service order", err.Error())
		return
	}

	item, _ := h.svc.Get(r.Context(), id)
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.ParseUUIDParam(r, "orderId")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid order id", nil)
		return
	}

	if err := h.svc.Delete(r.Context(), id); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "SERVICE_ORDER_DELETE_ERROR", "could not delete service order", err.Error())
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
		status := ServiceOrderStatus(v)
		filter.Status = &status
	}
	if v := q.Get("order_type"); v != "" {
		orderType := ServiceOrderType(v)
		filter.OrderType = &orderType
	}
	if v := q.Get("bus_id"); v != "" {
		filter.BusID = &v
	}
	return filter, nil
}
