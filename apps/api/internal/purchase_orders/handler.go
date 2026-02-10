package purchase_orders

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
	r.Route("/purchase-orders", func(r chi.Router) {
		r.Get("/", h.list)
		r.Post("/", h.create)
		r.Get("/{orderId}", h.get)
		r.Patch("/{orderId}", h.update)
		r.Post("/{orderId}/items", h.addItem)
		r.Delete("/{orderId}/items/{itemId}", h.removeItem)
		r.Post("/{orderId}/send", h.send)
		r.Post("/{orderId}/receive", h.receive)
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
		httpx.WriteError(w, http.StatusInternalServerError, "PURCHASE_ORDERS_LIST_ERROR", "could not list purchase orders", err.Error())
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
			httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "purchase order not found", nil)
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "PURCHASE_ORDER_GET_ERROR", "could not get purchase order", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var input CreatePurchaseOrderInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_BODY", "invalid json", err.Error())
		return
	}
	if input.SupplierID == "" {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "supplier_id is required", nil)
		return
	}

	item, err := h.svc.Create(r.Context(), input)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "PURCHASE_ORDER_CREATE_ERROR", "could not create purchase order", err.Error())
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

	var input UpdatePurchaseOrderInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_BODY", "invalid json", err.Error())
		return
	}

	item, err := h.svc.Update(r.Context(), id, input)
	if err != nil {
		if IsNotFound(err) {
			httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "purchase order not found", nil)
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "PURCHASE_ORDER_UPDATE_ERROR", "could not update purchase order", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) addItem(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.ParseUUIDParam(r, "orderId")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid order id", nil)
		return
	}

	var input AddItemInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_BODY", "invalid json", err.Error())
		return
	}
	if input.ProductID == "" {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "product_id is required", nil)
		return
	}

	item, err := h.svc.AddItem(r.Context(), id, input)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "PURCHASE_ORDER_ADD_ITEM_ERROR", "could not add item", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) removeItem(w http.ResponseWriter, r *http.Request) {
	orderID, err := httpx.ParseUUIDParam(r, "orderId")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid order id", nil)
		return
	}
	itemID, err := httpx.ParseUUIDParam(r, "itemId")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid item id", nil)
		return
	}

	if err := h.svc.RemoveItem(r.Context(), orderID, itemID); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "PURCHASE_ORDER_REMOVE_ITEM_ERROR", "could not remove item", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusNoContent, nil)
}

func (h *Handler) send(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.ParseUUIDParam(r, "orderId")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid order id", nil)
		return
	}

	if err := h.svc.Send(r.Context(), id); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "PURCHASE_ORDER_SEND_ERROR", "could not send purchase order", err.Error())
		return
	}

	item, _ := h.svc.Get(r.Context(), id)
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) receive(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.ParseUUIDParam(r, "orderId")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid order id", nil)
		return
	}

	if err := h.svc.MarkReceived(r.Context(), id); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "PURCHASE_ORDER_RECEIVE_ERROR", "could not mark purchase order as received", err.Error())
		return
	}

	item, _ := h.svc.Get(r.Context(), id)
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) cancel(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.ParseUUIDParam(r, "orderId")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid order id", nil)
		return
	}

	if err := h.svc.Cancel(r.Context(), id); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "PURCHASE_ORDER_CANCEL_ERROR", "could not cancel purchase order", err.Error())
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
		httpx.WriteError(w, http.StatusInternalServerError, "PURCHASE_ORDER_DELETE_ERROR", "could not delete purchase order", err.Error())
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
		status := PurchaseOrderStatus(v)
		filter.Status = &status
	}
	if v := q.Get("supplier_id"); v != "" {
		filter.SupplierID = &v
	}
	if v := q.Get("service_order_id"); v != "" {
		filter.ServiceOrderID = &v
	}
	return filter, nil
}
