package chat

import "strings"

func parseBookingCancelInput(history []Message, text string, currentBooking *BookingLookupResult, currentBookingCreate *BookingCreateResult) (BookingCancelInput, bool) {
	body := strings.Join(strings.Fields(strings.TrimSpace(text)), " ")
	if body == "" || !looksLikeBookingCancelIntent(body) {
		return BookingCancelInput{}, false
	}

	bookingID, reservationCode := resolveBookingCancelTarget(body, history, currentBooking, currentBookingCreate)
	if bookingID == "" {
		return BookingCancelInput{}, false
	}

	return BookingCancelInput{
		BookingID:       bookingID,
		ReservationCode: reservationCode,
		Reason:          extractBookingCancelReason(body),
		Actor:           "CUSTOMER",
	}, true
}

func looksLikeBookingCancelIntent(text string) bool {
	lower := strings.ToLower(strings.TrimSpace(text))
	if lower == "" {
		return false
	}

	keywords := []string{
		"cancelar",
		"cancela",
		"cancelamento",
		"desist",
		"nao vou mais",
		"não vou mais",
	}
	for _, keyword := range keywords {
		if strings.Contains(lower, keyword) {
			return true
		}
	}
	return false
}

func resolveBookingCancelTarget(text string, history []Message, currentBooking *BookingLookupResult, currentBookingCreate *BookingCreateResult) (string, string) {
	bookingID, reservationCode := extractBookingIdentifiers(text)
	if bookingID != "" {
		return bookingID, reservationCode
	}
	if currentBooking != nil && len(currentBooking.Results) > 0 {
		item := currentBooking.Results[0]
		return strings.TrimSpace(item.ID), firstNonEmpty(reservationCode, strings.TrimSpace(item.ReservationCode))
	}
	if currentBookingCreate != nil && strings.TrimSpace(currentBookingCreate.BookingID) != "" {
		return strings.TrimSpace(currentBookingCreate.BookingID), firstNonEmpty(reservationCode, strings.TrimSpace(currentBookingCreate.ReservationCode))
	}
	if previous := findLatestBookingCreateContext(history); previous != nil && strings.TrimSpace(previous.BookingID) != "" {
		return strings.TrimSpace(previous.BookingID), firstNonEmpty(reservationCode, strings.TrimSpace(previous.ReservationCode))
	}
	if previous := findLatestBookingLookupContext(history); previous != nil && len(previous.Results) > 0 {
		item := previous.Results[0]
		return strings.TrimSpace(item.ID), firstNonEmpty(reservationCode, strings.TrimSpace(item.ReservationCode))
	}
	return "", reservationCode
}

func extractBookingCancelReason(text string) string {
	lower := strings.ToLower(strings.TrimSpace(text))
	switch {
	case strings.Contains(lower, "timeout"), strings.Contains(lower, "expir"), strings.Contains(lower, "venceu"), strings.Contains(lower, "tempo do pagamento"):
		return "timeout_50m"
	default:
		return "customer_requested"
	}
}

func buildBookingCancelRequestPayload(input BookingCancelInput) map[string]interface{} {
	payload := map[string]interface{}{
		"booking_id":       strings.TrimSpace(input.BookingID),
		"reservation_code": strings.TrimSpace(input.ReservationCode),
		"reason":           normalizeBookingCancelReason(input.Reason),
		"actor":            normalizeBookingCancelActor(input.Actor),
	}
	return payload
}

func buildBookingCancelResponsePayload(result BookingCancelResult) map[string]interface{} {
	return map[string]interface{}{
		"mode":              result.Mode,
		"booking_id":        result.BookingID,
		"reservation_code":  result.ReservationCode,
		"trip_id":           result.TripID,
		"booking_status":    result.BookingStatus,
		"previous_status":   result.PreviousStatus,
		"reason":            result.Reason,
		"actor":             result.Actor,
		"passenger_count":   result.PassengerCount,
		"idempotent":        result.Idempotent,
		"errors":            result.Errors,
		"message_for_agent": result.MessageForAgent,
	}
}
