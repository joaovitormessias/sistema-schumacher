package bookings

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"schumacher-tur/api/internal/pricing"
	httpx "schumacher-tur/api/internal/shared/http"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/bookings", func(r chi.Router) {
		r.Get("/", h.list)
		r.Post("/", h.create)
		r.Post("/checkout", h.checkout)
		r.Get("/{bookingId}", h.get)
		r.Patch("/{bookingId}", h.update)
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
		httpx.WriteError(w, http.StatusInternalServerError, "BOOKINGS_LIST_ERROR", "could not list bookings", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusOK, items)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var input CreateBookingInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_BODY", "invalid json", err.Error())
		return
	}
	if input.TripID == "" || input.SeatID == "" || input.Passenger.Name == "" {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "trip_id, seat_id and passenger.name are required", nil)
		return
	}
	if input.BoardStopID == "" || input.AlightStopID == "" {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "board_stop_id and alight_stop_id are required", nil)
		return
	}
	if input.TotalAmount < 0 || input.DepositAmount < 0 || input.RemainderAmount < 0 {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "amounts cannot be negative", nil)
		return
	}

	item, err := h.svc.Create(r.Context(), input)
	if err != nil {
		if errors.Is(err, ErrMissingFields) || errors.Is(err, ErrMissingStops) || errors.Is(err, ErrNegativeAmounts) || errors.Is(err, ErrInvalidAmounts) {
			httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
			return
		}
		if errors.Is(err, ErrSeatNotInTrip) {
			httpx.WriteError(w, http.StatusBadRequest, "SEAT_INVALID", "seat does not belong to trip bus", nil)
			return
		}
		if errors.Is(err, pricing.ErrTripNotFound) || errors.Is(err, pricing.ErrStopNotFound) {
			httpx.WriteError(w, http.StatusNotFound, "STOP_NOT_FOUND", "trip or stop not found", nil)
			return
		}
		if errors.Is(err, pricing.ErrInvalidStopOrder) {
			httpx.WriteError(w, http.StatusBadRequest, "INVALID_STOPS", "board_stop must be before alight_stop", nil)
			return
		}
		if errors.Is(err, pricing.ErrSegmentFareNotFound) {
			httpx.WriteError(w, http.StatusNotFound, "FARE_NOT_FOUND", "segment fare not found", nil)
			return
		}
		if errors.Is(err, pricing.ErrInvalidFareMode) {
			httpx.WriteError(w, http.StatusBadRequest, "INVALID_FARE_MODE", "invalid fare_mode", nil)
			return
		}
		if errors.Is(err, pricing.ErrFareAmountRequired) {
			httpx.WriteError(w, http.StatusBadRequest, "FARE_AMOUNT_REQUIRED", "fare_amount_final is required for MANUAL", nil)
			return
		}
		if IsUniqueViolation(err) {
			httpx.WriteError(w, http.StatusConflict, "SEAT_TAKEN", "seat already reserved", nil)
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "BOOKING_CREATE_ERROR", "could not create booking", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, item)
}

func (h *Handler) checkout(w http.ResponseWriter, r *http.Request) {
	var input CheckoutBookingInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_BODY", "invalid json", err.Error())
		return
	}
	if input.TripID == "" || input.SeatID == "" || input.Passenger.Name == "" {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "trip_id, seat_id and passenger.name are required", nil)
		return
	}
	if input.BoardStopID == "" || input.AlightStopID == "" {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "board_stop_id and alight_stop_id are required", nil)
		return
	}

	result, err := h.svc.Checkout(r.Context(), input)
	if err != nil {
		if errors.Is(err, ErrMissingFields) || errors.Is(err, ErrMissingStops) || errors.Is(err, ErrNegativeAmounts) || errors.Is(err, ErrInvalidAmounts) || errors.Is(err, ErrInvalidInitialPayment) || errors.Is(err, ErrInitialPaymentBelowMinimum) {
			httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
			return
		}
		if errors.Is(err, ErrSeatNotInTrip) {
			httpx.WriteError(w, http.StatusBadRequest, "SEAT_INVALID", "seat does not belong to trip bus", nil)
			return
		}
		if errors.Is(err, pricing.ErrTripNotFound) || errors.Is(err, pricing.ErrStopNotFound) {
			httpx.WriteError(w, http.StatusNotFound, "STOP_NOT_FOUND", "trip or stop not found", nil)
			return
		}
		if errors.Is(err, pricing.ErrInvalidStopOrder) {
			httpx.WriteError(w, http.StatusBadRequest, "INVALID_STOPS", "board_stop must be before alight_stop", nil)
			return
		}
		if errors.Is(err, pricing.ErrSegmentFareNotFound) {
			httpx.WriteError(w, http.StatusNotFound, "FARE_NOT_FOUND", "segment fare not found", nil)
			return
		}
		if errors.Is(err, pricing.ErrInvalidFareMode) {
			httpx.WriteError(w, http.StatusBadRequest, "INVALID_FARE_MODE", "invalid fare_mode", nil)
			return
		}
		if errors.Is(err, pricing.ErrFareAmountRequired) {
			httpx.WriteError(w, http.StatusBadRequest, "FARE_AMOUNT_REQUIRED", "fare_amount_final is required for MANUAL", nil)
			return
		}
		if IsUniqueViolation(err) {
			httpx.WriteError(w, http.StatusConflict, "SEAT_TAKEN", "seat already reserved", nil)
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "BOOKING_CHECKOUT_ERROR", "could not complete checkout", err.Error())
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, result)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "bookingId")
	if id == "" {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid booking id", nil)
		return
	}
	item, err := h.svc.Get(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "booking not found", nil)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "bookingId")
	if id == "" {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid booking id", nil)
		return
	}

	var input UpdateBookingInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_BODY", "invalid json", err.Error())
		return
	}

	if input.Status == nil || *input.Status == "" {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "status is required", nil)
		return
	}

	if !isValidBookingStatus(*input.Status) {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid status", nil)
		return
	}

	item, err := h.svc.UpdateStatus(r.Context(), id, *input.Status)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "BOOKING_UPDATE_ERROR", "could not update booking", err.Error())
		return
	}

	httpx.WriteJSON(w, http.StatusOK, item)
}

func isValidBookingStatus(status string) bool {
	switch status {
	case "PENDING", "CONFIRMED", "CANCELLED", "EXPIRED":
		return true
	default:
		return false
	}
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
	return filter, nil
}
