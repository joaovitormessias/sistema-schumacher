package chat

import (
	"strings"
)

func parsePaymentCreateInput(session Session, history []Message, text string, currentBooking *BookingLookupResult, currentBookingCreate *BookingCreateResult) (PaymentCreateInput, bool) {
	body := strings.Join(strings.Fields(strings.TrimSpace(text)), " ")
	if body == "" || !looksLikePaymentCreateIntent(body) {
		return PaymentCreateInput{}, false
	}

	bookingID, reservationCode := resolvePaymentCreateTarget(body, history, currentBooking, currentBookingCreate)
	if bookingID == "" {
		return PaymentCreateInput{}, false
	}

	input := PaymentCreateInput{
		BookingID:        bookingID,
		ReservationCode:  reservationCode,
		PaymentType:      extractRequestedPaymentType(body),
		DepositPerPerson: 250,
		CustomerName:     strings.TrimSpace(session.CustomerName),
		CustomerPhone:    strings.TrimSpace(session.CustomerPhone),
		CustomerDocument: extractExplicitPaymentDocument(body),
		Note:             buildPaymentCreateNote(bookingID, reservationCode, body),
	}
	return input, true
}

func looksLikePaymentCreateIntent(text string) bool {
	lower := strings.ToLower(strings.TrimSpace(text))
	if lower == "" {
		return false
	}

	keywords := []string{
		"gerar pix",
		"gerar o pix",
		"gera o pix",
		"gera pix",
		"criar pix",
		"manda o pix",
		"manda pix",
		"envia o pix",
		"enviar pix",
		"quero pagar",
		"pagar via pix",
		"fazer o pagamento",
		"fazer pagamento",
		"gerar pagamento",
		"criar cobranca",
		"criar cobrança",
		"copia e cola do pix",
		"copia e cola pix",
	}
	for _, keyword := range keywords {
		if strings.Contains(lower, keyword) {
			return true
		}
	}
	return false
}

func extractRequestedPaymentType(text string) string {
	lower := strings.ToLower(strings.TrimSpace(text))
	integralKeywords := []string{
		"integral",
		"valor total",
		"pagar tudo",
		"quitar",
		"quitacao",
		"quitação",
		"restante",
		"saldo",
	}
	for _, keyword := range integralKeywords {
		if strings.Contains(lower, keyword) {
			return "integral"
		}
	}

	signalKeywords := []string{
		"sinal",
		"entrada",
		"deposito",
		"depósito",
	}
	for _, keyword := range signalKeywords {
		if strings.Contains(lower, keyword) {
			return "sinal"
		}
	}
	return "sinal"
}

func extractExplicitPaymentDocument(text string) string {
	if match := passengerCPFPattern.FindStringSubmatch(strings.ToUpper(text)); len(match) == 2 {
		return normalizeDigits(match[1])
	}
	return ""
}

func resolvePaymentCreateTarget(text string, history []Message, currentBooking *BookingLookupResult, currentBookingCreate *BookingCreateResult) (string, string) {
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

func buildPaymentCreateNote(bookingID string, reservationCode string, text string) string {
	target := firstNonEmpty(strings.TrimSpace(reservationCode), strings.TrimSpace(bookingID))
	return "Pagamento " + extractRequestedPaymentType(text) + " reserva " + target
}

func findLatestBookingCreateContext(history []Message) *BookingCreateResult {
	for i := len(history) - 1; i >= 0; i-- {
		message := history[i]
		if message.Direction != "OUTBOUND" || !isAutomationDraftStatus(message.ProcessingStatus) {
			continue
		}
		toolContext := asMap(message.Payload["tool_context"])
		payload := asMap(toolContext[toolNameBookingCreate])
		if len(payload) == 0 {
			continue
		}
		result := parseBookingCreateContextPayload(payload)
		if strings.TrimSpace(result.BookingID) != "" {
			return &result
		}
	}
	return nil
}

func findLatestBookingLookupContext(history []Message) *BookingLookupResult {
	for i := len(history) - 1; i >= 0; i-- {
		message := history[i]
		if message.Direction != "OUTBOUND" || !isAutomationDraftStatus(message.ProcessingStatus) {
			continue
		}
		toolContext := asMap(message.Payload["tool_context"])
		payload := asMap(toolContext[toolNameBookingLookup])
		if len(payload) == 0 {
			continue
		}
		result := parseBookingLookupContextPayload(payload)
		if len(result.Results) > 0 {
			return &result
		}
	}
	return nil
}

func parseBookingCreateContextPayload(payload map[string]interface{}) BookingCreateResult {
	result := BookingCreateResult{
		Mode:            strings.TrimSpace(asString(payload["mode"])),
		BookingID:       strings.TrimSpace(asString(payload["booking_id"])),
		ReservationCode: strings.TrimSpace(asString(payload["reservation_code"])),
		Status:          strings.TrimSpace(asString(payload["status"])),
		TotalAmount:     asFloat64(payload["total_amount"]),
		DepositAmount:   asFloat64(payload["deposit_amount"]),
		RemainderAmount: asFloat64(payload["remainder_amount"]),
		Errors:          asStringSlice(payload["errors"]),
		MessageForAgent: strings.TrimSpace(asString(payload["message_for_agent"])),
	}
	for _, item := range asInterfaceSliceMaps(payload["passengers"]) {
		result.Passengers = append(result.Passengers, BookingCreatePassengerResult{
			Name:         strings.TrimSpace(asString(item["name"])),
			Document:     strings.TrimSpace(asString(item["document"])),
			DocumentType: strings.TrimSpace(asString(item["document_type"])),
			Phone:        strings.TrimSpace(asString(item["phone"])),
			SeatID:       strings.TrimSpace(asString(item["seat_id"])),
			IsLapChild:   strings.EqualFold(strings.TrimSpace(asString(item["is_lap_child"])), "true"),
		})
	}
	return result
}

func parseBookingLookupContextPayload(payload map[string]interface{}) BookingLookupResult {
	result := BookingLookupResult{
		Filter: BookingLookupInput{
			BookingID:       strings.TrimSpace(asString(payload["booking_id"])),
			ReservationCode: strings.TrimSpace(asString(payload["reservation_code"])),
			Limit:           asInt(payload["result_count"]),
		},
	}
	for _, item := range asInterfaceSliceMaps(payload["results"]) {
		result.Results = append(result.Results, BookingLookupItem{
			ID:              strings.TrimSpace(asString(item["booking_id"])),
			ReservationCode: strings.TrimSpace(asString(item["reservation_code"])),
			Status:          strings.TrimSpace(asString(item["status"])),
			TotalAmount:     asFloat64(item["total_amount"]),
			DepositAmount:   asFloat64(item["deposit_amount"]),
			RemainderAmount: asFloat64(item["remainder_amount"]),
			PassengerName:   strings.TrimSpace(asString(item["passenger_name"])),
			PassengerPhone:  strings.TrimSpace(asString(item["passenger_phone"])),
			SeatNumber:      asInt(item["seat_number"]),
		})
	}
	return result
}

func buildPaymentCreateRequestPayload(input PaymentCreateInput) map[string]interface{} {
	payload := map[string]interface{}{
		"booking_id":         strings.TrimSpace(input.BookingID),
		"reservation_code":   strings.TrimSpace(input.ReservationCode),
		"payment_type":       normalizePaymentType(input.PaymentType),
		"confirm_paid":       input.ConfirmPaid,
		"paid_amount":        input.PaidAmount,
		"deposit_per_person": input.DepositPerPerson,
		"customer_name":      strings.TrimSpace(input.CustomerName),
		"customer_phone":     strings.TrimSpace(input.CustomerPhone),
		"customer_document":  strings.TrimSpace(input.CustomerDocument),
		"customer_email":     strings.TrimSpace(input.CustomerEmail),
		"note":               strings.TrimSpace(input.Note),
	}
	return payload
}

func buildPaymentCreateResponsePayload(result PaymentCreateResult) map[string]interface{} {
	payload := map[string]interface{}{
		"mode":              result.Mode,
		"booking_id":        result.BookingID,
		"reservation_code":  result.ReservationCode,
		"booking_status":    result.BookingStatus,
		"payment_type":      result.PaymentType,
		"stage":             result.Stage,
		"amount_total":      result.AmountTotal,
		"amount_paid":       result.AmountPaid,
		"amount_due":        result.AmountDue,
		"payment_id":        result.PaymentID,
		"payment_status":    result.PaymentStatus,
		"provider":          result.Provider,
		"provider_ref":      result.ProviderRef,
		"pix_code":          result.PixCode,
		"checkout_url":      result.CheckoutURL,
		"errors":            result.Errors,
		"message_for_agent": result.MessageForAgent,
	}
	return payload
}
