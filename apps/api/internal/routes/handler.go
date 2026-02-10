package routes

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	httpx "schumacher-tur/api/internal/shared/http"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/routes", func(r chi.Router) {
		r.Get("/", h.list)
		r.Post("/", h.create)
		r.Patch("/{routeId}", h.update)
		r.Get("/{routeId}", h.get)
		r.Get("/{routeId}/stops", h.listStops)
		r.Post("/{routeId}/stops", h.createStop)
		r.Patch("/{routeId}/stops/{stopId}", h.updateStop)
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
		httpx.WriteError(w, http.StatusInternalServerError, "ROUTES_LIST_ERROR", "could not list routes", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusOK, items)
}

func parseListFilter(r *http.Request) (ListFilter, error) {
	filter := ListFilter{}
	q := r.URL.Query()
	filter.Search = strings.TrimSpace(q.Get("search"))
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

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.ParseUUIDParam(r, "routeId")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid route id", nil)
		return
	}
	item, err := h.svc.Get(r.Context(), id)
	if err != nil {
		if IsNotFound(err) {
			httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "route not found", nil)
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "ROUTE_GET_ERROR", "could not get route", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var input CreateRouteInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_BODY", "invalid json", err.Error())
		return
	}
	if input.Name == "" || input.OriginCity == "" || input.DestinationCity == "" {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "name, origin_city and destination_city are required", nil)
		return
	}

	item, err := h.svc.Create(r.Context(), input)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "ROUTE_CREATE_ERROR", "could not create route", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, item)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.ParseUUIDParam(r, "routeId")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid route id", nil)
		return
	}

	var input UpdateRouteInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_BODY", "invalid json", err.Error())
		return
	}

	item, err := h.svc.Update(r.Context(), id, input)
	if err != nil {
		if IsNotFound(err) {
			httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "route not found", nil)
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "ROUTE_UPDATE_ERROR", "could not update route", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) listStops(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.ParseUUIDParam(r, "routeId")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid route id", nil)
		return
	}
	items, err := h.svc.ListStops(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "ROUTE_STOPS_LIST_ERROR", "could not list route stops", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusOK, items)
}

func (h *Handler) createStop(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.ParseUUIDParam(r, "routeId")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid route id", nil)
		return
	}
	var input CreateRouteStopInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_BODY", "invalid json", err.Error())
		return
	}
	if input.City == "" || input.StopOrder <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "city and stop_order are required", nil)
		return
	}

	item, err := h.svc.CreateStop(r.Context(), id, input)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "ROUTE_STOP_CREATE_ERROR", "could not create route stop", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, item)
}

func (h *Handler) updateStop(w http.ResponseWriter, r *http.Request) {
	routeID, err := httpx.ParseUUIDParam(r, "routeId")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid route id", nil)
		return
	}
	stopID, err := httpx.ParseUUIDParam(r, "stopId")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid stop id", nil)
		return
	}

	var input UpdateRouteStopInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_BODY", "invalid json", err.Error())
		return
	}

	item, err := h.svc.UpdateStop(r.Context(), routeID, stopID, input)
	if err != nil {
		if IsNotFound(err) {
			httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "route stop not found", nil)
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "ROUTE_STOP_UPDATE_ERROR", "could not update route stop", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}
