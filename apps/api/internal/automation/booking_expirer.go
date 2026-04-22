package automation

import (
	"time"

	"schumacher-tur/api/internal/bookings"
)

const (
	defaultBookingsExpireLimit       = 500
	defaultBookingsExpireHoldMinutes = 50
)

type bookingExpirationCandidate struct {
	Booking            bookings.BookingListItem
	EffectiveExpiresAt time.Time
	ExpirationSource   string
}

func normalizeRunBookingsExpireInput(input RunBookingsExpireInput) (RunBookingsExpireInput, error) {
	switch {
	case input.Limit < 0:
		return RunBookingsExpireInput{}, ErrInvalidBookingsExpireLimit
	case input.Limit == 0:
		input.Limit = defaultBookingsExpireLimit
	case input.Limit > defaultBookingsExpireLimit:
		input.Limit = defaultBookingsExpireLimit
	}

	switch {
	case input.HoldMinutes < 0:
		return RunBookingsExpireInput{}, ErrInvalidBookingsExpireHold
	case input.HoldMinutes == 0:
		input.HoldMinutes = defaultBookingsExpireHoldMinutes
	}

	return input, nil
}

func collectExpiredBookings(items []bookings.BookingListItem, observedAt time.Time, holdMinutes int) []bookingExpirationCandidate {
	expired := make([]bookingExpirationCandidate, 0, len(items))
	for _, item := range items {
		expiresAt, source, ok := resolveBookingExpiration(item, holdMinutes)
		if !ok {
			continue
		}
		if !expiresAt.Before(observedAt) {
			continue
		}

		expired = append(expired, bookingExpirationCandidate{
			Booking:            item,
			EffectiveExpiresAt: expiresAt,
			ExpirationSource:   source,
		})
	}
	return expired
}

func resolveBookingExpiration(item bookings.BookingListItem, holdMinutes int) (time.Time, string, bool) {
	if item.ExpiresAt != nil {
		return item.ExpiresAt.UTC(), "RESERVED_UNTIL", true
	}

	if item.CreatedAt.IsZero() {
		return time.Time{}, "", false
	}

	return item.CreatedAt.UTC().Add(time.Duration(holdMinutes) * time.Minute), "CREATED_AT_FALLBACK", true
}

func serializeExpiredBookings(items []ExpiredBookingResult) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		out = append(out, map[string]interface{}{
			"booking_id":           item.BookingID,
			"reservation_code":     item.ReservationCode,
			"trip_id":              item.TripID,
			"previous_status":      item.PreviousStatus,
			"updated_status":       item.UpdatedStatus,
			"effective_expires_at": item.EffectiveExpiresAt.Format(time.RFC3339Nano),
			"expiration_source":    item.ExpirationSource,
		})
	}
	return out
}
