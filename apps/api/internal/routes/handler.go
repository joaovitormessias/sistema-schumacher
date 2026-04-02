package routes

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	httpx "schumacher-tur/api/internal/shared/http"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/routes", func(r chi.Router) {
		r.Get("/", h.list)
		r.Post("/", h.create)
		r.Get("/cities/candidates", h.searchCityCandidates)
		r.Patch("/{routeId}", h.update)
		r.Get("/{routeId}", h.get)
		r.Post("/{routeId}/publish", h.publish)
		r.Post("/{routeId}/duplicate", h.duplicate)
		r.Get("/{routeId}/stops", h.listStops)
		r.Post("/{routeId}/stops", h.createStop)
		r.Patch("/{routeId}/stops/{stopId}", h.updateStop)
		r.Delete("/{routeId}/stops/{stopId}", h.deleteStop)
		r.Get("/{routeId}/segment-prices", h.listSegmentPrices)
		r.Put("/{routeId}/segment-prices", h.upsertSegmentPrices)
	})
}

func (h *Handler) searchCityCandidates(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("query"))
	if query == "" {
		httpx.WriteJSON(w, http.StatusOK, []CityCandidate{})
		return
	}

	limit := 8
	if rawLimit := strings.TrimSpace(r.URL.Query().Get("limit")); rawLimit != "" {
		parsed, err := strconv.Atoi(rawLimit)
		if err != nil || parsed <= 0 {
			httpx.WriteError(w, http.StatusBadRequest, "INVALID_LIMIT", "invalid limit", nil)
			return
		}
		if parsed > 10 {
			parsed = 10
		}
		limit = parsed
	}

	items, err := h.svc.SearchCityCandidates(r.Context(), query, limit)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "ROUTE_CITY_CANDIDATES_ERROR", "could not search city candidates", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusOK, items)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	filter, err := parseListFilter(r)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_PAGINATION", "invalid pagination parameters", nil)
		return
	}
	items, err := h.svc.List(r.Context(), filter)
	if err != nil {
		log.Printf("Routes list error: %v", err)
		httpx.WriteError(w, http.StatusInternalServerError, "ROUTES_LIST_ERROR", "could not list routes", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusOK, items)
}

func parseListFilter(r *http.Request) (ListFilter, error) {
	filter := ListFilter{}
	q := r.URL.Query()
	filter.Search = strings.TrimSpace(q.Get("search"))
	status := strings.ToLower(strings.TrimSpace(q.Get("status")))
	if status != "" && status != "active" && status != "inactive" && status != "all" {
		return filter, errors.New("invalid status")
	}
	filter.Status = status
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
	id := strings.TrimSpace(chi.URLParam(r, "routeId"))
	if id == "" {
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
		if errors.Is(err, ErrUnsupported) {
			httpx.WriteError(w, http.StatusNotImplemented, "NOT_SUPPORTED", "operation not supported for production ticketing schema", nil)
			return
		}
		if strings.Contains(err.Error(), "origin_latitude") ||
			strings.Contains(err.Error(), "origin_longitude") ||
			strings.Contains(err.Error(), "destination_latitude") ||
			strings.Contains(err.Error(), "destination_longitude") ||
			strings.Contains(err.Error(), "latitude and longitude") ||
			strings.Contains(err.Error(), "could not resolve coordinates") {
			httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "ROUTE_CREATE_ERROR", "could not create route", err.Error())
		return
	}
	logRouteDraftCreated(item, "create")
	httpx.WriteJSON(w, http.StatusCreated, item)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(chi.URLParam(r, "routeId"))
	if id == "" {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid route id", nil)
		return
	}

	var input UpdateRouteInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_BODY", "invalid json", err.Error())
		return
	}

	isPublishAttempt := input.IsActive != nil && *input.IsActive
	if isPublishAttempt {
		logRoutePublishAttempt(id, "update")
	}

	item, err := h.svc.Update(r.Context(), id, input)
	if err != nil {
		if errors.Is(err, ErrUnsupported) {
			httpx.WriteError(w, http.StatusNotImplemented, "NOT_SUPPORTED", "operation not supported for production ticketing schema", nil)
			return
		}
		var publishErr RoutePublishBlockedError
		if errors.As(err, &publishErr) {
			if isPublishAttempt {
				logRoutePublishBlocked(id, "update", publishErr.MissingRequirements)
			}
			httpx.WriteJSON(w, http.StatusUnprocessableEntity, map[string]interface{}{
				"code":                 "ROUTE_PUBLISH_BLOCKED",
				"message":              "route does not meet publish requirements",
				"requirements_missing": publishErr.MissingRequirements,
			})
			return
		}
		if IsNotFound(err) {
			httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "route not found", nil)
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "ROUTE_UPDATE_ERROR", "could not update route", err.Error())
		return
	}
	if isPublishAttempt && item.IsActive {
		logRoutePublished(item, "update")
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) publish(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(chi.URLParam(r, "routeId"))
	if id == "" {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid route id", nil)
		return
	}

	logRoutePublishAttempt(id, "publish")

	item, err := h.svc.Publish(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrUnsupported) {
			httpx.WriteError(w, http.StatusNotImplemented, "NOT_SUPPORTED", "operation not supported for production ticketing schema", nil)
			return
		}
		var publishErr RoutePublishBlockedError
		if errors.As(err, &publishErr) {
			logRoutePublishBlocked(id, "publish", publishErr.MissingRequirements)
			httpx.WriteJSON(w, http.StatusUnprocessableEntity, map[string]interface{}{
				"code":                 "ROUTE_PUBLISH_BLOCKED",
				"message":              "route does not meet publish requirements",
				"requirements_missing": publishErr.MissingRequirements,
			})
			return
		}
		if IsNotFound(err) {
			httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "route not found", nil)
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "ROUTE_PUBLISH_ERROR", "could not publish route", err.Error())
		return
	}
	logRoutePublished(item, "publish")
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) duplicate(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(chi.URLParam(r, "routeId"))
	if id == "" {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid route id", nil)
		return
	}

	item, err := h.svc.Duplicate(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrUnsupported) {
			httpx.WriteError(w, http.StatusNotImplemented, "NOT_SUPPORTED", "operation not supported for production ticketing schema", nil)
			return
		}
		if IsNotFound(err) {
			httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "route not found", nil)
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "ROUTE_DUPLICATE_ERROR", "could not duplicate route", err.Error())
		return
	}
	logRouteDraftCreated(item, "duplicate")
	httpx.WriteJSON(w, http.StatusCreated, item)
}

func (h *Handler) listStops(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(chi.URLParam(r, "routeId"))
	if id == "" {
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
	id := strings.TrimSpace(chi.URLParam(r, "routeId"))
	if id == "" {
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
		if errors.Is(err, ErrUnsupported) {
			httpx.WriteError(w, http.StatusNotImplemented, "NOT_SUPPORTED", "operation not supported for production ticketing schema", nil)
			return
		}
		if IsNotFound(err) {
			httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "route not found", nil)
			return
		}
		if strings.Contains(err.Error(), "eta_offset_minutes") ||
			strings.Contains(err.Error(), "latitude") ||
			strings.Contains(err.Error(), "longitude") ||
			strings.Contains(err.Error(), "could not resolve coordinates") {
			httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "ROUTE_STOP_CREATE_ERROR", "could not create route stop", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, item)
}

func (h *Handler) updateStop(w http.ResponseWriter, r *http.Request) {
	routeID := strings.TrimSpace(chi.URLParam(r, "routeId"))
	if routeID == "" {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid route id", nil)
		return
	}
	stopID := strings.TrimSpace(chi.URLParam(r, "stopId"))
	if stopID == "" {
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
		if errors.Is(err, ErrUnsupported) {
			httpx.WriteError(w, http.StatusNotImplemented, "NOT_SUPPORTED", "operation not supported for production ticketing schema", nil)
			return
		}
		if errors.Is(err, ErrRouteStopLocked) {
			httpx.WriteError(w, http.StatusConflict, "ROUTE_STOP_LOCKED", "route stop reorder is locked for routes with linked trips", nil)
			return
		}
		if strings.Contains(err.Error(), "eta_offset_minutes") ||
			strings.Contains(err.Error(), "latitude") ||
			strings.Contains(err.Error(), "longitude") ||
			strings.Contains(err.Error(), "could not resolve coordinates") {
			httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
			return
		}
		if IsNotFound(err) {
			httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "route stop not found", nil)
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "ROUTE_STOP_UPDATE_ERROR", "could not update route stop", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) deleteStop(w http.ResponseWriter, r *http.Request) {
	routeID := strings.TrimSpace(chi.URLParam(r, "routeId"))
	if routeID == "" {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid route id", nil)
		return
	}
	stopID := strings.TrimSpace(chi.URLParam(r, "stopId"))
	if stopID == "" {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid stop id", nil)
		return
	}

	if err := h.svc.DeleteStop(r.Context(), routeID, stopID); err != nil {
		if errors.Is(err, ErrUnsupported) {
			httpx.WriteError(w, http.StatusNotImplemented, "NOT_SUPPORTED", "operation not supported for production ticketing schema", nil)
			return
		}
		if errors.Is(err, ErrRouteHasLinkedTrips) {
			httpx.WriteError(w, http.StatusConflict, "ROUTE_STOP_LOCKED", "route stop delete is locked for routes with linked trips", nil)
			return
		}
		if IsNotFound(err) {
			httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "route stop not found", nil)
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "ROUTE_STOP_DELETE_ERROR", "could not delete route stop", err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) listSegmentPrices(w http.ResponseWriter, r *http.Request) {
	routeID := strings.TrimSpace(chi.URLParam(r, "routeId"))
	if routeID == "" {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid route id", nil)
		return
	}

	matrix, err := h.svc.ListSegmentPrices(r.Context(), routeID)
	if err != nil {
		if IsNotFound(err) {
			httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "route not found", nil)
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "ROUTE_SEGMENT_PRICES_LIST_ERROR", "could not list route segment prices", err.Error())
		return
	}

	httpx.WriteJSON(w, http.StatusOK, matrix)
}

func (h *Handler) upsertSegmentPrices(w http.ResponseWriter, r *http.Request) {
	routeID := strings.TrimSpace(chi.URLParam(r, "routeId"))
	if routeID == "" {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid route id", nil)
		return
	}

	var input UpsertRouteSegmentPricesInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_BODY", "invalid json", err.Error())
		return
	}

	matrix, err := h.svc.UpsertSegmentPrices(r.Context(), routeID, input)
	if err != nil {
		switch {
		case IsNotFound(err):
			httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "route not found", nil)
		case errors.Is(err, ErrInvalidSegmentPair):
			httpx.WriteError(w, http.StatusBadRequest, "INVALID_SEGMENT_PAIR", "origin_stop and destination_stop must belong to route and keep route order", nil)
		case errors.Is(err, ErrInvalidSegmentPrice):
			httpx.WriteError(w, http.StatusBadRequest, "INVALID_SEGMENT_PRICE", "price must be >= 0", nil)
		case errors.Is(err, ErrInvalidSegmentStatus):
			httpx.WriteError(w, http.StatusBadRequest, "INVALID_SEGMENT_STATUS", "status must be ACTIVE or INACTIVE", nil)
		default:
			httpx.WriteError(w, http.StatusInternalServerError, "ROUTE_SEGMENT_PRICES_UPDATE_ERROR", "could not update route segment prices", err.Error())
		}
		return
	}

	httpx.WriteJSON(w, http.StatusOK, matrix)
}

func logRouteDraftCreated(item Route, source string) {
	duplicatedFrom := ""
	if item.DuplicatedFromRouteID != nil {
		duplicatedFrom = *item.DuplicatedFromRouteID
	}
	log.Printf(
		"event=route_draft_created route_id=%s source=%s stop_count=%d configuration_status=%s duplicated_from_route_id=%s",
		item.ID,
		source,
		item.StopCount,
		item.ConfigurationStatus,
		duplicatedFrom,
	)
}

func logRoutePublishAttempt(routeID string, source string) {
	log.Printf("event=route_publish_attempt route_id=%s source=%s", routeID, source)
	log.Printf("metric=route_publish_attempt_total value=1 source=%s", source)
}

func logRoutePublishBlocked(routeID string, source string, missingRequirements []string) {
	log.Printf("metric=route_publish_blocked_total value=1 rule=any source=%s", source)
	normalized := make([]string, 0, len(missingRequirements))
	for _, requirement := range missingRequirements {
		rule := normalizeRequirementRule(requirement)
		normalized = append(normalized, rule)
		log.Printf("metric=route_publish_blocked_total value=1 rule=%s source=%s", rule, source)
	}
	log.Printf(
		"event=route_publish_blocked route_id=%s source=%s missing_count=%d missing_rules=%s",
		routeID,
		source,
		len(normalized),
		strings.Join(normalized, ","),
	)
}

func logRoutePublished(item Route, source string) {
	draftToPublishMinutes := time.Since(item.CreatedAt).Minutes()
	if draftToPublishMinutes < 0 {
		draftToPublishMinutes = 0
	}
	log.Printf(
		"event=route_published route_id=%s source=%s stop_count=%d configuration_status=%s",
		item.ID,
		source,
		item.StopCount,
		item.ConfigurationStatus,
	)
	log.Printf(
		"metric=route_draft_to_publish_minutes value=%.2f route_id=%s source=%s",
		draftToPublishMinutes,
		item.ID,
		source,
	)
	log.Printf("metric=route_publish_success_total value=1 source=%s", source)
}

func normalizeRequirementRule(requirement string) string {
	key := strings.TrimSpace(strings.ToLower(requirement))
	switch key {
	case "route must be active":
		return "route_must_be_active"
	case "at least two stops are required":
		return "at_least_two_stops_required"
	case "stop_order must start at 1":
		return "stop_order_starts_at_1"
	case "stop_order must be sequential without gaps":
		return "stop_order_sequential_without_gaps"
	case "first stop city must match origin_city":
		return "first_stop_matches_origin_city"
	case "last stop city must match destination_city":
		return "last_stop_matches_destination_city"
	case "eta_offset_minutes must be >= 0":
		return "eta_offset_minutes_non_negative"
	case "eta_offset_minutes must be non-decreasing":
		return "eta_offset_minutes_non_decreasing"
	case "route not found":
		return "route_not_found"
	default:
		key = strings.ReplaceAll(key, " ", "_")
		key = strings.ReplaceAll(key, "-", "_")
		key = strings.ReplaceAll(key, ">", "gt")
		key = strings.ReplaceAll(key, "<", "lt")
		key = strings.ReplaceAll(key, "=", "")
		key = strings.ReplaceAll(key, ".", "")
		for strings.Contains(key, "__") {
			key = strings.ReplaceAll(key, "__", "_")
		}
		return strings.Trim(key, "_")
	}
}
