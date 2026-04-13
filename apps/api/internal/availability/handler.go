package availability

import (
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	httpx "schumacher-tur/api/internal/shared/http"
)

var cityStatePattern = regexp.MustCompile(`^(.+?)[/\s,-]+([A-Za-z]{2})$`)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/availability", func(r chi.Router) {
		r.Get("/search", h.search)
	})
}

func (h *Handler) search(w http.ResponseWriter, r *http.Request) {
	filter, err := parseSearchFilter(r)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
		return
	}

	items, err := h.svc.Search(r.Context(), filter)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "AVAILABILITY_SEARCH_ERROR", "could not search availability", err.Error())
		return
	}

	httpx.WriteJSON(w, http.StatusOK, items)
}

func parseSearchFilter(r *http.Request) (SearchFilter, error) {
	q := r.URL.Query()
	filter := SearchFilter{
		Origin:      normalizeDisplayName(q.Get("origin")),
		Destination: normalizeDisplayName(q.Get("destination")),
		PackageName: strings.TrimSpace(q.Get("package_name")),
		Qty:         1,
		Limit:       10,
		OnlyActive:  true,
		IncludePast: false,
	}

	if filter.Origin == "" && filter.Destination == "" && filter.PackageName == "" {
		return filter, errors.New("origin, destination or package_name is required")
	}
	if filter.Origin != "" && filter.Destination != "" && strings.EqualFold(filter.Origin, filter.Destination) {
		return filter, errors.New("origin and destination must be different")
	}

	if rawDate := strings.TrimSpace(q.Get("trip_date")); rawDate != "" {
		parsed, err := time.Parse("2006-01-02", rawDate)
		if err != nil {
			return filter, errors.New("trip_date must be in YYYY-MM-DD format")
		}
		filter.TripDate = &parsed
	}

	if rawQty := strings.TrimSpace(q.Get("qtd")); rawQty != "" {
		qty, err := strconv.Atoi(rawQty)
		if err != nil || qty <= 0 {
			return filter, errors.New("qtd must be a positive integer")
		}
		filter.Qty = qty
	}

	if rawLimit := strings.TrimSpace(q.Get("limit")); rawLimit != "" {
		limit, err := strconv.Atoi(rawLimit)
		if err != nil || limit <= 0 {
			return filter, errors.New("limit must be a positive integer")
		}
		if limit > 50 {
			limit = 50
		}
		filter.Limit = limit
	}

	if q.Has("only_active") {
		val, err := parseBoolQuery(q.Get("only_active"))
		if err != nil {
			return filter, errors.New("only_active must be true, false, 1 or 0")
		}
		filter.OnlyActive = val
	}

	if q.Has("include_past") {
		val, err := parseBoolQuery(q.Get("include_past"))
		if err != nil {
			return filter, errors.New("include_past must be true, false, 1 or 0")
		}
		filter.IncludePast = val
	}

	return filter, nil
}

func parseBoolQuery(value string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true":
		return true, nil
	case "0", "false":
		return false, nil
	default:
		return false, errors.New("invalid boolean")
	}
}

func normalizeDisplayName(value string) string {
	raw := strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
	if raw == "" {
		return ""
	}

	if match := cityStatePattern.FindStringSubmatch(raw); len(match) == 3 {
		city := strings.TrimSpace(strings.TrimRight(match[1], "-, "))
		uf := strings.ToUpper(strings.TrimSpace(match[2]))
		if city != "" && uf != "" {
			return city + "/" + uf
		}
	}

	return raw
}
