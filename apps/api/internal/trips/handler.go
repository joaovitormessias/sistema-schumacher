package trips

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	httpx "schumacher-tur/api/internal/shared/http"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/trips", func(r chi.Router) {
		r.Get("/", h.list)
		r.Post("/", h.create)
		r.Patch("/{tripId}", h.update)
		r.Get("/{tripId}", h.get)
		r.Get("/{tripId}/seats", h.listSeats)
		r.Get("/{tripId}/stops", h.listStops)
		r.Post("/{tripId}/stops", h.createStop)
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
		httpx.WriteError(w, http.StatusInternalServerError, "TRIPS_LIST_ERROR", "could not list trips", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusOK, items)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.ParseUUIDParam(r, "tripId")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid trip id", nil)
		return
	}
	item, err := h.svc.Get(r.Context(), id)
	if err != nil {
		if IsNotFound(err) {
			httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "trip not found", nil)
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "TRIP_GET_ERROR", "could not get trip", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var input CreateTripInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_BODY", "invalid json", err.Error())
		return
	}
	if input.RouteID == "" || input.BusID == "" || input.DepartureAt.IsZero() {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "route_id, bus_id and departure_at are required", nil)
		return
	}

	item, err := h.svc.Create(r.Context(), input)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "TRIP_CREATE_ERROR", "could not create trip", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, item)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.ParseUUIDParam(r, "tripId")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid trip id", nil)
		return
	}

	var input UpdateTripInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_BODY", "invalid json", err.Error())
		return
	}

	item, err := h.svc.Update(r.Context(), id, input)
	if err != nil {
		if IsNotFound(err) {
			httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "trip not found", nil)
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "TRIP_UPDATE_ERROR", "could not update trip", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) listSeats(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.ParseUUIDParam(r, "tripId")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid trip id", nil)
		return
	}

	var boardStopID *uuid.UUID
	var alightStopID *uuid.UUID
	if v := r.URL.Query().Get("board_stop_id"); v != "" {
		parsed, err := uuid.Parse(v)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid board_stop_id", nil)
			return
		}
		boardStopID = &parsed
	}
	if v := r.URL.Query().Get("alight_stop_id"); v != "" {
		parsed, err := uuid.Parse(v)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid alight_stop_id", nil)
			return
		}
		alightStopID = &parsed
	}

	if (boardStopID == nil) != (alightStopID == nil) {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_PARAMS", "board_stop_id and alight_stop_id must be provided together", nil)
		return
	}

	seats, err := h.svc.ListSeats(r.Context(), id, boardStopID, alightStopID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "SEATS_LIST_ERROR", "could not list seats", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusOK, seats)
}

func (h *Handler) listStops(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.ParseUUIDParam(r, "tripId")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid trip id", nil)
		return
	}
	items, err := h.svc.ListStops(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "TRIP_STOPS_LIST_ERROR", "could not list trip stops", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusOK, items)
}

func (h *Handler) createStop(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.ParseUUIDParam(r, "tripId")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid trip id", nil)
		return
	}
	var input CreateTripStopInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_BODY", "invalid json", err.Error())
		return
	}
	if input.RouteStopID == "" {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "route_stop_id is required", nil)
		return
	}

	item, err := h.svc.CreateStop(r.Context(), id, input)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "TRIP_STOP_CREATE_ERROR", "could not create trip stop", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, item)
}

func parseListFilter(r *http.Request) (ListFilter, error) {
	filter := ListFilter{}
	q := r.URL.Query()
	filter.Search = strings.TrimSpace(q.Get("search"))
	filter.Status = strings.TrimSpace(q.Get("status"))
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
