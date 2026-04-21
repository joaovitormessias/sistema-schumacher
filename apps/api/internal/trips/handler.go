package trips

import (
	"errors"
	"log"
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
	r.Get("/trips", h.list)
	r.Post("/trips", h.create)
	r.Patch("/trips/{tripId}", h.update)
	r.Get("/trips/{tripId}", h.get)
	r.Get("/trips/{tripId}/details", h.getDetails)
	r.Get("/trips/{tripId}/seats", h.listSeats)
	r.Get("/trips/{tripId}/stops", h.listStops)
	r.Get("/trips/{tripId}/segment-prices", h.listSegmentPrices)
	r.Put("/trips/{tripId}/segment-prices", h.upsertSegmentPrices)
	r.Post("/trips/{tripId}/stops", h.createStop)
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
	id := strings.TrimSpace(chi.URLParam(r, "tripId"))
	if id == "" {
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

func (h *Handler) getDetails(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(chi.URLParam(r, "tripId"))
	if id == "" {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid trip id", nil)
		return
	}
	item, err := h.svc.GetDetails(r.Context(), id)
	if err != nil {
		if IsNotFound(err) {
			httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "trip not found", nil)
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "TRIP_DETAILS_ERROR", "could not get trip details", err.Error())
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
	if input.RouteID == "" || input.BusID == "" {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "route_id and bus_id are required", nil)
		return
	}
	if input.DepartureAt.IsZero() && len(input.Stops) == 0 {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "departure_at is required when stops are not provided", nil)
		return
	}
	if input.EstimatedKM != nil && *input.EstimatedKM < 0 {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "estimated_km must be >= 0", nil)
		return
	}
	for index, stop := range input.Stops {
		if strings.TrimSpace(stop.RouteStopID) == "" {
			httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "each stop requires route_stop_id", map[string]any{
				"index": index,
			})
			return
		}
	}

	item, err := h.svc.Create(r.Context(), input)
	if err != nil {
		if errors.Is(err, ErrUnsupported) {
			httpx.WriteError(w, http.StatusNotImplemented, "NOT_SUPPORTED", "operation not supported for production ticketing schema", nil)
			return
		}
		var routeErr RouteNotReadyError
		if errors.As(err, &routeErr) {
			log.Printf(
				"event=trip_create_blocked_route_not_ready route_id=%s missing_count=%d missing_rules=%s",
				input.RouteID,
				len(routeErr.MissingRequirements),
				strings.Join(routeErr.MissingRequirements, ","),
			)
			log.Printf("metric=trip_create_itinerary_validation_total value=1 outcome=blocked route_id=%s", input.RouteID)
			httpx.WriteJSON(w, http.StatusUnprocessableEntity, map[string]interface{}{
				"code":                 "ROUTE_NOT_READY",
				"message":              "route is not ready for trip creation",
				"requirements_missing": routeErr.MissingRequirements,
			})
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "TRIP_CREATE_ERROR", "could not create trip", err.Error())
		return
	}
	log.Printf("event=trip_created trip_id=%s route_id=%s status=%s", item.ID, item.RouteID, item.Status)
	log.Printf("metric=trip_create_itinerary_validation_total value=1 outcome=ready route_id=%s", item.RouteID)
	httpx.WriteJSON(w, http.StatusCreated, item)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(chi.URLParam(r, "tripId"))
	if id == "" {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid trip id", nil)
		return
	}

	var input UpdateTripInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_BODY", "invalid json", err.Error())
		return
	}
	if input.EstimatedKM != nil && *input.EstimatedKM < 0 {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "estimated_km must be >= 0", nil)
		return
	}
	if input.OperationalStatus != nil {
		httpx.WriteError(w, http.StatusBadRequest, "WORKFLOW_LOCKED", "operational_status is managed by workflow", nil)
		return
	}

	item, err := h.svc.Update(r.Context(), id, input)
	if err != nil {
		if errors.Is(err, ErrUnsupported) {
			httpx.WriteError(w, http.StatusNotImplemented, "NOT_SUPPORTED", "operation not supported for production ticketing schema", nil)
			return
		}
		var routeErr RouteNotReadyError
		if errors.As(err, &routeErr) {
			httpx.WriteJSON(w, http.StatusUnprocessableEntity, map[string]interface{}{
				"code":                 "ROUTE_NOT_READY",
				"message":              "route is not ready for trip update",
				"requirements_missing": routeErr.MissingRequirements,
			})
			return
		}
		if errors.Is(err, ErrOperationalStatusManagedByWorkflow) {
			httpx.WriteError(w, http.StatusBadRequest, "WORKFLOW_LOCKED", "operational_status is managed by workflow", nil)
			return
		}
		if errors.Is(err, ErrTripStatusManagedByWorkflow) {
			httpx.WriteError(w, http.StatusBadRequest, "WORKFLOW_LOCKED", "trip status progression is managed by workflow", nil)
			return
		}
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
	id := strings.TrimSpace(chi.URLParam(r, "tripId"))
	if id == "" {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid trip id", nil)
		return
	}

	var boardStopID *string
	var alightStopID *string
	if v := r.URL.Query().Get("board_stop_id"); v != "" {
		value := strings.TrimSpace(v)
		boardStopID = &value
	}
	if v := r.URL.Query().Get("alight_stop_id"); v != "" {
		value := strings.TrimSpace(v)
		alightStopID = &value
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
	id := strings.TrimSpace(chi.URLParam(r, "tripId"))
	if id == "" {
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
	id := strings.TrimSpace(chi.URLParam(r, "tripId"))
	if id == "" {
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
	if input.LegDistanceKM != nil && *input.LegDistanceKM < 0 {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "leg_distance_km must be >= 0", nil)
		return
	}
	if input.CumulativeDistanceKM != nil && *input.CumulativeDistanceKM < 0 {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "cumulative_distance_km must be >= 0", nil)
		return
	}

	item, err := h.svc.CreateStop(r.Context(), id, input)
	if err != nil {
		if errors.Is(err, ErrUnsupported) {
			httpx.WriteError(w, http.StatusNotImplemented, "NOT_SUPPORTED", "operation not supported for production ticketing schema", nil)
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "TRIP_STOP_CREATE_ERROR", "could not create trip stop", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, item)
}

func (h *Handler) listSegmentPrices(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(chi.URLParam(r, "tripId"))
	if id == "" {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid trip id", nil)
		return
	}

	matrix, err := h.svc.ListSegmentPrices(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrTripNotFound) {
			httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "trip not found", nil)
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "TRIP_SEGMENT_PRICES_LIST_ERROR", "could not list trip segment prices", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusOK, matrix)
}

func (h *Handler) upsertSegmentPrices(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(chi.URLParam(r, "tripId"))
	if id == "" {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid trip id", nil)
		return
	}

	var input UpsertTripSegmentPricesInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_BODY", "invalid json", err.Error())
		return
	}

	matrix, err := h.svc.UpsertSegmentPrices(r.Context(), id, input)
	if err != nil {
		switch {
		case errors.Is(err, ErrTripNotFound):
			httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "trip not found", nil)
		case errors.Is(err, ErrInvalidSegmentPair):
			httpx.WriteError(w, http.StatusBadRequest, "INVALID_SEGMENT_PAIR", "origin_stop and destination_stop must belong to trip and keep route order", nil)
		case errors.Is(err, ErrInvalidSegmentPrice):
			httpx.WriteError(w, http.StatusBadRequest, "INVALID_SEGMENT_PRICE", "price must be >= 0", nil)
		case errors.Is(err, ErrInvalidSegmentStatus):
			httpx.WriteError(w, http.StatusBadRequest, "INVALID_SEGMENT_STATUS", "status must be ACTIVE or INACTIVE", nil)
		default:
			httpx.WriteError(w, http.StatusInternalServerError, "TRIP_SEGMENT_PRICES_UPDATE_ERROR", "could not update trip segment prices", err.Error())
		}
		return
	}

	httpx.WriteJSON(w, http.StatusOK, matrix)
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
