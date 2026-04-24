package chat

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strings"
	"time"
)

var (
	optionIndexPattern          = regexp.MustCompile(`(?i)\bop[cç][aã]o\s*([1-5])\b`)
	passengerNamePattern        = regexp.MustCompile(`(?i)\bnome(?:\s+completo)?\s*(?:é|e|:|-)?\s*([A-ZÀ-ÿ][A-Za-zÀ-ÿ' ]{3,100}?)(?:\s+(?:cpf|rg|cnh|certid[aã]o|matr[ií]cula)\b|$)`)
	passengerAltNamePattern     = regexp.MustCompile(`(?i)\b(?:meu nome|sou)\s*(?:é|e)?\s*([A-ZÀ-ÿ][A-Za-zÀ-ÿ' ]{3,100}?)(?:\s+(?:cpf|rg|cnh|certid[aã]o|matr[ií]cula)\b|$)`)
	passengerCPFPattern         = regexp.MustCompile(`(?i)\bcpf\b[^0-9]*([0-9.\-]{11,14})`)
	passengerRGPattern          = regexp.MustCompile(`(?i)\brg\b[^A-Z0-9]*([A-Z0-9.\-]{4,20})`)
	passengerCNHPattern         = regexp.MustCompile(`(?i)\bcnh\b[^A-Z0-9]*([A-Z0-9.\-]{4,20})`)
	passengerBirthRecordPattern = regexp.MustCompile(`(?i)\b(?:certid[aã]o(?: de nascimento)?|matr[ií]cula)\b[^A-Z0-9]*([A-Z0-9.\-]{8,40})`)
	passengerLooseCPFLinePattern = regexp.MustCompile(`(?i)^\s*([A-Za-zÀ-ÿ][A-Za-zÀ-ÿ' ]{3,100}?)\s+([0-9.\-]{11,14})\s*$`)
)

func parseBookingCreateInput(session Session, history []Message, text string, currentAvailability *AvailabilitySearchResult) (BookingCreateInput, bool) {
	body := strings.TrimSpace(text)
	if body == "" {
		return BookingCreateInput{}, false
	}
	confirmationOnly := looksLikeBookingCreateConfirmation(body)
	if !looksLikeCreateBookingIntent(body) && !confirmationOnly {
		return BookingCreateInput{}, false
	}

	selectedOptionIndex, selected, ok := resolveBookingCreateSelection(body, history, currentAvailability)
	if !ok {
		return BookingCreateInput{}, false
	}

	passengerSource := body
	passengers := extractBookingCreatePassengers(passengerSource, session)
	if len(passengers) == 0 && confirmationOnly {
		passengerSource = findLatestPassengerDetailsText(history, session)
		passengers = extractBookingCreatePassengers(passengerSource, session)
	}
	if len(passengers) == 0 {
		return BookingCreateInput{}, false
	}

	qty := extractPassengerQuantity(body)
	if qty <= 0 {
		qty = extractPassengerQuantity(passengerSource)
	}
	if qty <= 0 {
		qty = len(passengers)
	}
	if qty != len(passengers) {
		return BookingCreateInput{}, false
	}
	applyLapChildFlags(passengers, inferLapChildCount(history))

	input := BookingCreateInput{
		SelectedOptionIndex:    selectedOptionIndex,
		TripID:                 strings.TrimSpace(selected.TripID),
		BoardStopID:            strings.TrimSpace(selected.BoardStopID),
		AlightStopID:           strings.TrimSpace(selected.AlightStopID),
		OriginDisplayName:      strings.TrimSpace(selected.OriginDisplayName),
		DestinationDisplayName: strings.TrimSpace(selected.DestinationDisplayName),
		TripDate:               strings.TrimSpace(selected.TripDate),
		DepartureTime:          strings.TrimSpace(selected.OriginDepartTime),
		Qty:                    qty,
		CustomerName:           firstNonEmpty(strings.TrimSpace(session.CustomerName), strings.TrimSpace(passengers[0].Name)),
		CustomerPhone:          strings.TrimSpace(session.CustomerPhone),
		Passengers:             passengers,
	}
	input.IdempotencyKey = buildBookingCreateIdempotencyKey(session, input)
	return input, true
}

func looksLikeBookingCreateConfirmation(text string) bool {
	folded := strings.TrimSpace(foldChatText(text))
	switch folded {
	case "sim", "isso", "isso mesmo", "pode seguir", "pode reservar", "confirmo", "confirmado", "ok", "certo":
		return true
	default:
		return false
	}
}

func looksLikeCreateBookingIntent(text string) bool {
	lower := strings.ToLower(strings.TrimSpace(text))
	if lower == "" {
		return false
	}

	keywords := []string{
		"quero reservar",
		"quero fazer a reserva",
		"pode reservar",
		"faz a reserva",
		"fazer a reserva",
		"seguir com a reserva",
		"vou querer",
		"quero a opcao",
		"quero a opção",
		"fechar essa",
		"fechar a passagem",
	}
	for _, keyword := range keywords {
		if strings.Contains(lower, keyword) {
			return true
		}
	}
	if extractSelectedOptionIndex(text) > 0 {
		actionWords := []string{"quero", "vou", "pode", "faz", "fechar", "reserv"}
		for _, keyword := range actionWords {
			if strings.Contains(lower, keyword) {
				return true
			}
		}
	}
	return false
}

func resolveBookingCreateSelection(text string, history []Message, currentAvailability *AvailabilitySearchResult) (int, AvailabilitySearchItem, bool) {
	options := []AvailabilitySearchItem{}
	if currentAvailability != nil && len(currentAvailability.Results) > 0 {
		options = append(options, currentAvailability.Results...)
	} else if previous := findLatestAvailabilityContext(history); previous != nil {
		options = append(options, previous.Results...)
	}
	if len(options) == 0 {
		return 0, AvailabilitySearchItem{}, false
	}

	index := extractSelectedOptionIndex(text)
	if index <= 0 {
		index = findLatestSelectedOptionIndex(history)
	}
	if index > 0 {
		if index > len(options) {
			return 0, AvailabilitySearchItem{}, false
		}
		return index, options[index-1], true
	}
	if len(options) == 1 {
		return 1, options[0], true
	}
	return 0, AvailabilitySearchItem{}, false
}

func extractSelectedOptionIndex(text string) int {
	if match := optionIndexPattern.FindStringSubmatch(text); len(match) == 2 {
		return asInt(float64(match[1][0] - '0'))
	}
	lower := strings.ToLower(strings.TrimSpace(text))
	switch {
	case strings.Contains(lower, "primeira opção"), strings.Contains(lower, "primeira opcao"):
		return 1
	case strings.Contains(lower, "segunda opção"), strings.Contains(lower, "segunda opcao"):
		return 2
	case strings.Contains(lower, "terceira opção"), strings.Contains(lower, "terceira opcao"):
		return 3
	}
	return 0
}

func extractBookingCreatePassengers(text string, session Session) []BookingCreatePassengerInput {
	if passengers := extractBookingCreatePassengersByLines(text, session); len(passengers) > 0 {
		return passengers
	}
	document, documentType := extractBookingPassengerDocument(text)
	if document == "" || documentType == "" {
		return nil
	}
	name := extractBookingPassengerName(text, session.CustomerName)
	if name == "" {
		return nil
	}
	return []BookingCreatePassengerInput{
		{
			Name:         name,
			Document:     document,
			DocumentType: documentType,
			Phone:        strings.TrimSpace(session.CustomerPhone),
		},
	}
}

func extractBookingCreatePassengersByLines(text string, session Session) []BookingCreatePassengerInput {
	segments := splitPassengerSegments(text)
	if len(segments) == 0 {
		return nil
	}

	passengers := make([]BookingCreatePassengerInput, 0, len(segments))
	for _, segment := range segments {
		match := passengerLooseCPFLinePattern.FindStringSubmatch(segment)
		if len(match) != 3 {
			return nil
		}
		name := normalizePassengerName(match[1])
		document := normalizeDigits(match[2])
		if name == "" || len(document) != 11 {
			return nil
		}
		passengers = append(passengers, BookingCreatePassengerInput{
			Name:         name,
			Document:     document,
			DocumentType: "CPF",
			Phone:        strings.TrimSpace(session.CustomerPhone),
		})
	}
	if len(passengers) == 0 {
		return nil
	}
	return passengers
}

func splitPassengerSegments(text string) []string {
	candidates := strings.FieldsFunc(text, func(r rune) bool {
		return r == '\n' || r == ';'
	})
	segments := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		candidate = strings.TrimSpace(candidate)
		if candidate == "" {
			continue
		}
		segments = append(segments, candidate)
	}
	if len(segments) > 1 {
		return segments
	}
	return nil
}

func extractBookingPassengerName(text string, fallback string) string {
	if match := passengerNamePattern.FindStringSubmatch(text); len(match) == 2 {
		return normalizePassengerName(match[1])
	}
	if match := passengerAltNamePattern.FindStringSubmatch(text); len(match) == 2 {
		return normalizePassengerName(match[1])
	}
	return normalizePassengerName(fallback)
}

func normalizePassengerName(value string) string {
	value = strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
	value = strings.Trim(value, " ,.;:-")
	if len(strings.Fields(value)) < 2 {
		return ""
	}
	return value
}

func extractBookingPassengerDocument(text string) (string, string) {
	if match := passengerCPFPattern.FindStringSubmatch(strings.ToUpper(text)); len(match) == 2 {
		document := normalizeDigits(match[1])
		if len(document) == 11 {
			return document, "CPF"
		}
	}
	if match := passengerRGPattern.FindStringSubmatch(strings.ToUpper(text)); len(match) == 2 {
		document := normalizeAlphaNumeric(match[1])
		if document != "" {
			return document, "RG"
		}
	}
	if match := passengerBirthRecordPattern.FindStringSubmatch(strings.ToUpper(text)); len(match) == 2 {
		document := normalizeAlphaNumeric(match[1])
		if document != "" {
			return document, "CERTIDAO_NASCIMENTO"
		}
	}
	if match := passengerCNHPattern.FindStringSubmatch(strings.ToUpper(text)); len(match) == 2 {
		document := normalizeAlphaNumeric(match[1])
		if document != "" {
			return document, "CNH"
		}
	}
	return "", ""
}

func normalizeDigits(value string) string {
	var builder strings.Builder
	for _, char := range value {
		if char >= '0' && char <= '9' {
			builder.WriteRune(char)
		}
	}
	return builder.String()
}

func normalizeAlphaNumeric(value string) string {
	var builder strings.Builder
	for _, char := range strings.ToUpper(strings.TrimSpace(value)) {
		if (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') {
			builder.WriteRune(char)
		}
	}
	return builder.String()
}

func findLatestSelectedOptionIndex(history []Message) int {
	for i := len(history) - 1; i >= 0; i-- {
		body := strings.TrimSpace(history[i].Body)
		if body == "" {
			continue
		}
		if index := extractSelectedOptionIndex(body); index > 0 {
			return index
		}
	}
	return 0
}

func findLatestPassengerDetailsText(history []Message, session Session) string {
	for i := len(history) - 1; i >= 0; i-- {
		message := history[i]
		if message.Direction != "INBOUND" {
			continue
		}
		body := strings.TrimSpace(message.Body)
		if body == "" || looksLikeBookingCreateConfirmation(body) {
			continue
		}
		if len(extractBookingCreatePassengers(body, session)) > 0 {
			return body
		}
	}
	return ""
}

func inferLapChildCount(history []Message) int {
	for i := len(history) - 1; i >= 0; i-- {
		message := history[i]
		if message.Direction != "INBOUND" || !looksLikeBookingCreateConfirmation(message.Body) {
			continue
		}
		for j := i - 1; j >= 0; j-- {
			previous := history[j]
			if previous.Direction != "OUTBOUND" {
				continue
			}
			folded := foldChatText(previous.Body)
			if strings.Contains(folded, "ate 5 anos") || strings.Contains(folded, "crianca de 5 anos ou menos") {
				return 1
			}
			break
		}
	}
	return 0
}

func applyLapChildFlags(passengers []BookingCreatePassengerInput, lapChildCount int) {
	if lapChildCount <= 0 || len(passengers) == 0 {
		return
	}
	if lapChildCount > len(passengers)-1 {
		lapChildCount = len(passengers) - 1
	}
	for i := len(passengers) - lapChildCount; i < len(passengers); i++ {
		if i >= 0 && i < len(passengers) {
			passengers[i].IsLapChild = true
		}
	}
}

func buildBookingCreateIdempotencyKey(session Session, input BookingCreateInput) string {
	parts := []string{
		strings.TrimSpace(session.ContactKey),
		strings.TrimSpace(input.TripID),
		strings.TrimSpace(input.BoardStopID),
		strings.TrimSpace(input.AlightStopID),
	}
	for _, item := range input.Passengers {
		parts = append(parts, strings.TrimSpace(item.DocumentType)+"="+strings.TrimSpace(item.Document))
	}
	hash := sha256.Sum256([]byte(strings.Join(parts, "|")))
	return "chat-booking-create-" + hex.EncodeToString(hash[:16])
}

func findLatestAvailabilityContext(history []Message) *AvailabilitySearchResult {
	for i := len(history) - 1; i >= 0; i-- {
		message := history[i]
		if message.Direction != "OUTBOUND" || !isAutomationDraftStatus(message.ProcessingStatus) {
			continue
		}
		toolContext := asMap(message.Payload["tool_context"])
		if len(toolContext) == 0 {
			continue
		}
		payload := asMap(toolContext[toolNameAvailabilitySearch])
		if len(payload) == 0 {
			continue
		}
		result := parseAvailabilityContextPayload(payload)
		if len(result.Results) > 0 {
			return &result
		}
	}
	return nil
}

func parseAvailabilityContextPayload(payload map[string]interface{}) AvailabilitySearchResult {
	result := AvailabilitySearchResult{
		Filter: AvailabilitySearchInput{
			Origin:      strings.TrimSpace(asString(payload["origin"])),
			Destination: strings.TrimSpace(asString(payload["destination"])),
			Qty:         asInt(payload["qtd"]),
			Limit:       len(asInterfaceSliceMaps(payload["results"])),
		},
		Results: []AvailabilitySearchItem{},
	}
	if rawDate := strings.TrimSpace(asString(payload["trip_date"])); rawDate != "" {
		if parsed, err := time.Parse("2006-01-02", rawDate); err == nil {
			parsed = parsed.UTC()
			result.Filter.TripDate = &parsed
		}
	}
	for _, item := range asInterfaceSliceMaps(payload["results"]) {
		result.Results = append(result.Results, AvailabilitySearchItem{
			SegmentID:              strings.TrimSpace(asString(item["segment_id"])),
			TripID:                 strings.TrimSpace(asString(item["trip_id"])),
			RouteID:                strings.TrimSpace(asString(item["route_id"])),
			BoardStopID:            strings.TrimSpace(asString(item["board_stop_id"])),
			AlightStopID:           strings.TrimSpace(asString(item["alight_stop_id"])),
			OriginStopID:           strings.TrimSpace(asString(item["origin_stop_id"])),
			DestinationStopID:      strings.TrimSpace(asString(item["destination_stop_id"])),
			OriginDisplayName:      strings.TrimSpace(asString(item["origin_display_name"])),
			DestinationDisplayName: strings.TrimSpace(asString(item["destination_display_name"])),
			OriginDepartTime:       strings.TrimSpace(asString(item["origin_depart_time"])),
			TripDate:               strings.TrimSpace(asString(item["trip_date"])),
			SeatsAvailable:         asInt(item["seats_available"]),
			Price:                  asFloat64(item["price"]),
			Currency:               strings.TrimSpace(asString(item["currency"])),
			Status:                 strings.TrimSpace(asString(item["status"])),
			TripStatus:             strings.TrimSpace(asString(item["trip_status"])),
			PackageName:            strings.TrimSpace(asString(item["package_name"])),
		})
	}
	return result
}

func buildBookingCreateRequestPayload(input BookingCreateInput) map[string]interface{} {
	passengers := make([]map[string]interface{}, 0, len(input.Passengers))
	for _, item := range input.Passengers {
		passengers = append(passengers, map[string]interface{}{
			"name":          item.Name,
			"document":      item.Document,
			"document_type": item.DocumentType,
			"phone":         item.Phone,
			"is_lap_child":  item.IsLapChild,
		})
	}
	return map[string]interface{}{
		"selected_option_index":    input.SelectedOptionIndex,
		"trip_id":                  input.TripID,
		"board_stop_id":            input.BoardStopID,
		"alight_stop_id":           input.AlightStopID,
		"origin_display_name":      input.OriginDisplayName,
		"destination_display_name": input.DestinationDisplayName,
		"trip_date":                input.TripDate,
		"departure_time":           input.DepartureTime,
		"qtd":                      input.Qty,
		"customer_name":            input.CustomerName,
		"customer_phone":           input.CustomerPhone,
		"idempotency_key":          input.IdempotencyKey,
		"passengers":               passengers,
	}
}

func buildBookingCreateResponsePayload(result BookingCreateResult) map[string]interface{} {
	passengers := make([]map[string]interface{}, 0, len(result.Passengers))
	for _, item := range result.Passengers {
		passengers = append(passengers, map[string]interface{}{
			"name":          item.Name,
			"document":      item.Document,
			"document_type": item.DocumentType,
			"phone":         item.Phone,
			"seat_id":       item.SeatID,
			"is_lap_child":  item.IsLapChild,
		})
	}
	payload := map[string]interface{}{
		"mode":              result.Mode,
		"booking_id":        result.BookingID,
		"reservation_code":  result.ReservationCode,
		"status":            result.Status,
		"total_amount":      result.TotalAmount,
		"deposit_amount":    result.DepositAmount,
		"remainder_amount":  result.RemainderAmount,
		"passenger_count":   len(result.Passengers),
		"passengers":        passengers,
		"errors":            result.Errors,
		"message_for_agent": result.MessageForAgent,
	}
	if result.ReservedUntil != nil {
		payload["reserved_until"] = result.ReservedUntil.UTC().Format(time.RFC3339Nano)
	}
	return payload
}

func asFloat64(value interface{}) float64 {
	switch typed := value.(type) {
	case float64:
		return typed
	case float32:
		return float64(typed)
	case int:
		return float64(typed)
	case int32:
		return float64(typed)
	case int64:
		return float64(typed)
	default:
		return 0
	}
}
