package chat

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	ErrAgentToolFailed        = errors.New("chat agent tool execution failed")
	ErrAgentToolNotConfigured = errors.New("chat agent tool not configured")

	locationMentionPattern = regexp.MustCompile(`(?i)\b([\p{L}][\p{L}' ]{1,60}?)[/\s,-]+([a-z]{2})\b`)
	routeFromToPattern     = regexp.MustCompile(`(?i)\bde\s+(.+?)\s+para\s+(.+?)(?:\s+(?:em|dia|para)\b|$)`)
	isoDatePattern         = regexp.MustCompile(`\b(\d{4})-(\d{2})-(\d{2})\b`)
	brDatePattern          = regexp.MustCompile(`\b(\d{2})/(\d{2})(?:/(\d{4}))?\b`)
	qtyPattern             = regexp.MustCompile(`\b(\d{1,2})\s*(?:pessoas?|passagens?|assentos?|lugares?)\b`)
	bookingIDPattern       = regexp.MustCompile(`(?i)\bBK-[A-Z0-9]+\b`)
	reservationCodePattern = regexp.MustCompile(`\b[A-Z0-9]{8}\b`)
)

const toolNameAvailabilitySearch = "availability_search"
const toolNamePricingQuote = "pricing_quote"
const toolNameBookingLookup = "booking_lookup"
const toolNamePaymentStatus = "payment_status"

type agentToolContext struct {
	Calls        []ToolCall
	Availability *AvailabilitySearchResult
	Pricing      *PricingQuoteResult
	Booking      *BookingLookupResult
	Payments     *PaymentStatusResult
}

func (s *Service) canSearchAvailability() bool {
	return s.availability != nil && s.availability.Enabled()
}

func (s *Service) canSearchPricing() bool {
	return s.pricing != nil && s.pricing.Enabled()
}

func (s *Service) canSearchBookings() bool {
	return s.bookings != nil && s.bookings.Enabled()
}

func (s *Service) canSearchPayments() bool {
	return s.payments != nil && s.payments.Enabled()
}

func (s *Service) resolveAgentToolContext(ctx context.Context, session Session, memory map[string]interface{}) (agentToolContext, error) {
	currentTurn := strings.TrimSpace(asString(memory["current_turn_body"]))
	context := agentToolContext{}
	if s.canSearchBookings() {
		searchInput, ok := parseBookingLookupInput(currentTurn)
		if ok {
			startedAt := time.Now().UTC()
			requestPayload := buildBookingLookupRequestPayload(searchInput)
			result, err := s.bookings.Search(ctx, searchInput)
			finishedAt := time.Now().UTC()
			finishedAtPtr := &finishedAt
			if err != nil {
				call, recordErr := s.store.CreateToolCall(ctx, CreateToolCallInput{
					SessionID:      session.ID,
					ToolName:       toolNameBookingLookup,
					RequestPayload: requestPayload,
					Status:         "FAILED",
					ErrorCode:      "BOOKING_LOOKUP_ERROR",
					ErrorMessage:   strings.TrimSpace(err.Error()),
					StartedAt:      startedAt,
					FinishedAt:     finishedAtPtr,
				})
				if recordErr != nil {
					return agentToolContext{}, recordErr
				}
				return agentToolContext{Calls: []ToolCall{call}}, fmt.Errorf("%w: %v", ErrAgentToolFailed, err)
			}

			call, err := s.store.CreateToolCall(ctx, CreateToolCallInput{
				SessionID:       session.ID,
				ToolName:        toolNameBookingLookup,
				RequestPayload:  requestPayload,
				ResponsePayload: buildBookingLookupResponsePayload(result),
				Status:          "COMPLETED",
				StartedAt:       startedAt,
				FinishedAt:      finishedAtPtr,
			})
			if err != nil {
				return agentToolContext{}, err
			}

			context.Calls = append(context.Calls, call)
			context.Booking = &result

			if s.canSearchPayments() && looksLikePaymentLookupIntent(currentTurn) && len(result.Results) > 0 {
				paymentInput := PaymentStatusInput{
					BookingID:       strings.TrimSpace(result.Results[0].ID),
					ReservationCode: strings.TrimSpace(result.Results[0].ReservationCode),
					Limit:           5,
				}
				paymentStartedAt := time.Now().UTC()
				paymentRequestPayload := buildPaymentStatusRequestPayload(paymentInput)
				paymentResult, paymentErr := s.payments.Search(ctx, paymentInput)
				paymentFinishedAt := time.Now().UTC()
				paymentFinishedAtPtr := &paymentFinishedAt
				if paymentErr != nil {
					call, recordErr := s.store.CreateToolCall(ctx, CreateToolCallInput{
						SessionID:      session.ID,
						ToolName:       toolNamePaymentStatus,
						RequestPayload: paymentRequestPayload,
						Status:         "FAILED",
						ErrorCode:      "PAYMENT_STATUS_ERROR",
						ErrorMessage:   strings.TrimSpace(paymentErr.Error()),
						StartedAt:      paymentStartedAt,
						FinishedAt:     paymentFinishedAtPtr,
					})
					if recordErr != nil {
						return agentToolContext{}, recordErr
					}
					context.Calls = append(context.Calls, call)
					return context, fmt.Errorf("%w: %v", ErrAgentToolFailed, paymentErr)
				}

				call, err := s.store.CreateToolCall(ctx, CreateToolCallInput{
					SessionID:       session.ID,
					ToolName:        toolNamePaymentStatus,
					RequestPayload:  paymentRequestPayload,
					ResponsePayload: buildPaymentStatusResponsePayload(paymentResult),
					Status:          "COMPLETED",
					StartedAt:       paymentStartedAt,
					FinishedAt:      paymentFinishedAtPtr,
				})
				if err != nil {
					return agentToolContext{}, err
				}
				context.Calls = append(context.Calls, call)
				context.Payments = &paymentResult
			}

			return context, nil
		}
	}

	if !s.canSearchAvailability() {
		return context, nil
	}

	searchInput, ok := parseAvailabilitySearchInput(currentTurn, time.Now().UTC())
	if !ok {
		return context, nil
	}

	startedAt := time.Now().UTC()
	requestPayload := buildAvailabilityToolRequestPayload(searchInput)
	result, err := s.availability.Search(ctx, searchInput)
	finishedAt := time.Now().UTC()
	finishedAtPtr := &finishedAt
	if err != nil {
		call, recordErr := s.store.CreateToolCall(ctx, CreateToolCallInput{
			SessionID:      session.ID,
			ToolName:       toolNameAvailabilitySearch,
			RequestPayload: requestPayload,
			Status:         "FAILED",
			ErrorCode:      "AVAILABILITY_SEARCH_ERROR",
			ErrorMessage:   strings.TrimSpace(err.Error()),
			StartedAt:      startedAt,
			FinishedAt:     finishedAtPtr,
		})
		if recordErr != nil {
			return agentToolContext{}, recordErr
		}
		return agentToolContext{Calls: []ToolCall{call}}, fmt.Errorf("%w: %v", ErrAgentToolFailed, err)
	}

	call, err := s.store.CreateToolCall(ctx, CreateToolCallInput{
		SessionID:       session.ID,
		ToolName:        toolNameAvailabilitySearch,
		RequestPayload:  requestPayload,
		ResponsePayload: buildAvailabilityToolResponsePayload(result),
		Status:          "COMPLETED",
		StartedAt:       startedAt,
		FinishedAt:      finishedAtPtr,
	})
	if err != nil {
		return agentToolContext{}, err
	}
	context.Calls = append(context.Calls, call)
	context.Availability = &result

	if s.canSearchPricing() && looksLikePricingQuoteIntent(currentTurn) && len(result.Results) > 0 {
		pricingInput := buildPricingQuoteInput(result)
		pricingStartedAt := time.Now().UTC()
		pricingRequestPayload := buildPricingQuoteRequestPayload(pricingInput)
		pricingResult, pricingErr := s.pricing.Search(ctx, pricingInput)
		pricingFinishedAt := time.Now().UTC()
		pricingFinishedAtPtr := &pricingFinishedAt
		if pricingErr != nil {
			call, recordErr := s.store.CreateToolCall(ctx, CreateToolCallInput{
				SessionID:      session.ID,
				ToolName:       toolNamePricingQuote,
				RequestPayload: pricingRequestPayload,
				Status:         "FAILED",
				ErrorCode:      "PRICING_QUOTE_ERROR",
				ErrorMessage:   strings.TrimSpace(pricingErr.Error()),
				StartedAt:      pricingStartedAt,
				FinishedAt:     pricingFinishedAtPtr,
			})
			if recordErr != nil {
				return agentToolContext{}, recordErr
			}
			context.Calls = append(context.Calls, call)
			return context, fmt.Errorf("%w: %v", ErrAgentToolFailed, pricingErr)
		}

		call, err := s.store.CreateToolCall(ctx, CreateToolCallInput{
			SessionID:       session.ID,
			ToolName:        toolNamePricingQuote,
			RequestPayload:  pricingRequestPayload,
			ResponsePayload: buildPricingQuoteResponsePayload(pricingResult),
			Status:          "COMPLETED",
			StartedAt:       pricingStartedAt,
			FinishedAt:      pricingFinishedAtPtr,
		})
		if err != nil {
			return agentToolContext{}, err
		}
		context.Calls = append(context.Calls, call)
		context.Pricing = &pricingResult
	}

	return context, nil
}

func parseAvailabilitySearchInput(text string, observedAt time.Time) (AvailabilitySearchInput, bool) {
	body := strings.Join(strings.Fields(strings.TrimSpace(text)), " ")
	if body == "" {
		return AvailabilitySearchInput{}, false
	}

	locations := extractCanonicalLocations(body)
	if !looksLikeAvailabilityIntent(body, len(locations)) {
		return AvailabilitySearchInput{}, false
	}

	origin, destination := "", ""
	if match := routeFromToPattern.FindStringSubmatch(body); len(match) == 3 {
		origin = normalizeLocationDisplayName(match[1])
		destination = normalizeLocationDisplayName(match[2])
	}
	if origin == "" || destination == "" {
		if len(locations) < 2 {
			return AvailabilitySearchInput{}, false
		}
		origin = locations[0]
		destination = locations[1]
	}
	if origin == "" || destination == "" || strings.EqualFold(origin, destination) {
		return AvailabilitySearchInput{}, false
	}

	input := AvailabilitySearchInput{
		Origin:      origin,
		Destination: destination,
		TripDate:    extractTripDate(body, observedAt),
		Qty:         extractPassengerQuantity(body),
		Limit:       5,
	}
	if input.Qty <= 0 {
		input.Qty = 1
	}
	return input, true
}

func looksLikeAvailabilityIntent(text string, locationCount int) bool {
	lower := strings.ToLower(strings.TrimSpace(text))
	if lower == "" {
		return false
	}

	keywords := []string{
		"dispon",
		"vaga",
		"assento",
		"lugar",
		"horario",
		"horario",
		"saida",
		"saída",
		"valor",
		"preco",
		"preço",
		"passagem",
	}
	for _, keyword := range keywords {
		if strings.Contains(lower, keyword) {
			return true
		}
	}

	return locationCount >= 2 && strings.Contains(lower, " para ")
}

func looksLikePricingQuoteIntent(text string) bool {
	lower := strings.ToLower(strings.TrimSpace(text))
	if lower == "" {
		return false
	}

	keywords := []string{
		"valor",
		"preco",
		"preço",
		"quanto",
		"custa",
		"cotacao",
		"cotação",
		"tarifa",
	}
	for _, keyword := range keywords {
		if strings.Contains(lower, keyword) {
			return true
		}
	}
	return false
}

func extractCanonicalLocations(text string) []string {
	matches := locationMentionPattern.FindAllStringSubmatch(text, -1)
	seen := make(map[string]struct{}, len(matches))
	items := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) != 3 {
			continue
		}
		location := normalizeLocationDisplayName(match[0])
		if location == "" {
			continue
		}
		if _, ok := seen[strings.ToUpper(location)]; ok {
			continue
		}
		seen[strings.ToUpper(location)] = struct{}{}
		items = append(items, location)
	}
	return items
}

func normalizeLocationDisplayName(value string) string {
	raw := strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
	if raw == "" {
		return ""
	}

	match := locationMentionPattern.FindStringSubmatch(raw)
	if len(match) != 3 {
		return ""
	}
	city := strings.TrimSpace(strings.TrimRight(match[1], "-, "))
	uf := strings.ToUpper(strings.TrimSpace(match[2]))
	if city == "" || uf == "" {
		return ""
	}

	parts := strings.Fields(strings.ToLower(city))
	for i, part := range parts {
		runes := []rune(part)
		if len(runes) == 0 {
			continue
		}
		parts[i] = strings.ToUpper(string(runes[0])) + string(runes[1:])
	}
	return strings.Join(parts, " ") + "/" + uf
}

func extractTripDate(text string, observedAt time.Time) *time.Time {
	if match := isoDatePattern.FindStringSubmatch(text); len(match) == 4 {
		if parsed, err := time.Parse("2006-01-02", match[0]); err == nil {
			parsed = parsed.UTC()
			return &parsed
		}
	}
	if match := brDatePattern.FindStringSubmatch(text); len(match) >= 3 {
		day, _ := strconv.Atoi(match[1])
		month, _ := strconv.Atoi(match[2])
		year := observedAt.UTC().Year()
		if len(match) >= 4 && strings.TrimSpace(match[3]) != "" {
			if parsedYear, err := strconv.Atoi(match[3]); err == nil {
				year = parsedYear
			}
		}
		parsed := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
		if parsed.Month() == time.Month(month) && parsed.Day() == day {
			if len(match) < 4 || strings.TrimSpace(match[3]) == "" {
				threshold := observedAt.UTC().Add(-24 * time.Hour)
				if parsed.Before(time.Date(threshold.Year(), threshold.Month(), threshold.Day(), 0, 0, 0, 0, time.UTC)) {
					nextYear := time.Date(year+1, time.Month(month), day, 0, 0, 0, 0, time.UTC)
					if nextYear.Month() == time.Month(month) && nextYear.Day() == day {
						parsed = nextYear
					}
				}
			}
			return &parsed
		}
	}
	return nil
}

func extractPassengerQuantity(text string) int {
	match := qtyPattern.FindStringSubmatch(text)
	if len(match) != 2 {
		return 0
	}
	value, err := strconv.Atoi(match[1])
	if err != nil || value <= 0 {
		return 0
	}
	return value
}

func buildAvailabilityToolRequestPayload(input AvailabilitySearchInput) map[string]interface{} {
	payload := map[string]interface{}{
		"origin":      input.Origin,
		"destination": input.Destination,
		"qtd":         input.Qty,
		"limit":       input.Limit,
		"only_active": true,
	}
	if input.TripDate != nil {
		payload["trip_date"] = input.TripDate.UTC().Format("2006-01-02")
	}
	return payload
}

func buildAvailabilityToolResponsePayload(result AvailabilitySearchResult) map[string]interface{} {
	items := make([]map[string]interface{}, 0, len(result.Results))
	for _, item := range result.Results {
		items = append(items, map[string]interface{}{
			"segment_id":               item.SegmentID,
			"trip_id":                  item.TripID,
			"route_id":                 item.RouteID,
			"board_stop_id":            item.BoardStopID,
			"alight_stop_id":           item.AlightStopID,
			"origin_stop_id":           item.OriginStopID,
			"destination_stop_id":      item.DestinationStopID,
			"origin_display_name":      item.OriginDisplayName,
			"destination_display_name": item.DestinationDisplayName,
			"origin_depart_time":       item.OriginDepartTime,
			"trip_date":                item.TripDate,
			"seats_available":          item.SeatsAvailable,
			"price":                    item.Price,
			"currency":                 item.Currency,
			"status":                   item.Status,
			"trip_status":              item.TripStatus,
			"package_name":             item.PackageName,
		})
	}

	payload := map[string]interface{}{
		"result_count": len(result.Results),
		"results":      items,
	}
	if result.Filter.Origin != "" {
		payload["origin"] = result.Filter.Origin
	}
	if result.Filter.Destination != "" {
		payload["destination"] = result.Filter.Destination
	}
	if result.Filter.TripDate != nil {
		payload["trip_date"] = result.Filter.TripDate.UTC().Format("2006-01-02")
	}
	if result.Filter.Qty > 0 {
		payload["qtd"] = result.Filter.Qty
	}
	return payload
}

func buildPricingQuoteInput(result AvailabilitySearchResult) PricingQuoteInput {
	candidates := make([]PricingQuoteCandidate, 0, len(result.Results))
	for index, item := range result.Results {
		if index >= 3 {
			break
		}
		candidates = append(candidates, PricingQuoteCandidate{
			TripID:                 item.TripID,
			RouteID:                item.RouteID,
			BoardStopID:            item.BoardStopID,
			AlightStopID:           item.AlightStopID,
			OriginStopID:           item.OriginStopID,
			DestinationStopID:      item.DestinationStopID,
			OriginDisplayName:      item.OriginDisplayName,
			DestinationDisplayName: item.DestinationDisplayName,
			OriginDepartTime:       item.OriginDepartTime,
			TripDate:               item.TripDate,
		})
	}
	return PricingQuoteInput{
		FareMode:   "AUTO",
		Candidates: candidates,
	}
}

func buildPricingQuoteRequestPayload(input PricingQuoteInput) map[string]interface{} {
	candidates := make([]map[string]interface{}, 0, len(input.Candidates))
	for _, candidate := range input.Candidates {
		candidates = append(candidates, map[string]interface{}{
			"trip_id":                  candidate.TripID,
			"route_id":                 candidate.RouteID,
			"board_stop_id":            candidate.BoardStopID,
			"alight_stop_id":           candidate.AlightStopID,
			"origin_stop_id":           candidate.OriginStopID,
			"destination_stop_id":      candidate.DestinationStopID,
			"origin_display_name":      candidate.OriginDisplayName,
			"destination_display_name": candidate.DestinationDisplayName,
			"origin_depart_time":       candidate.OriginDepartTime,
			"trip_date":                candidate.TripDate,
		})
	}
	return map[string]interface{}{
		"fare_mode":  input.FareMode,
		"candidates": candidates,
	}
}

func buildPricingQuoteResponsePayload(result PricingQuoteResult) map[string]interface{} {
	items := make([]map[string]interface{}, 0, len(result.Results))
	for _, item := range result.Results {
		items = append(items, map[string]interface{}{
			"trip_id":                  item.TripID,
			"route_id":                 item.RouteID,
			"board_stop_id":            item.BoardStopID,
			"alight_stop_id":           item.AlightStopID,
			"origin_stop_id":           item.OriginStopID,
			"destination_stop_id":      item.DestinationStopID,
			"origin_display_name":      item.OriginDisplayName,
			"destination_display_name": item.DestinationDisplayName,
			"origin_depart_time":       item.OriginDepartTime,
			"trip_date":                item.TripDate,
			"base_amount":              item.BaseAmount,
			"calc_amount":              item.CalcAmount,
			"final_amount":             item.FinalAmount,
			"currency":                 item.Currency,
			"fare_mode":                item.FareMode,
		})
	}
	return map[string]interface{}{
		"fare_mode":    result.Filter.FareMode,
		"result_count": len(result.Results),
		"results":      items,
	}
}

func parseBookingLookupInput(text string) (BookingLookupInput, bool) {
	body := strings.Join(strings.Fields(strings.TrimSpace(text)), " ")
	if body == "" || !looksLikeBookingLookupIntent(body) {
		return BookingLookupInput{}, false
	}

	bookingID := strings.ToUpper(strings.TrimSpace(bookingIDPattern.FindString(body)))
	reservationCode := ""
	for _, token := range reservationCodePattern.FindAllString(strings.ToUpper(body), -1) {
		if token == "" || token == bookingID || !containsLetter(token) {
			continue
		}
		reservationCode = token
		break
	}

	if bookingID == "" && reservationCode == "" {
		return BookingLookupInput{}, false
	}

	return BookingLookupInput{
		BookingID:       bookingID,
		ReservationCode: reservationCode,
		Limit:           3,
	}, true
}

func looksLikeBookingLookupIntent(text string) bool {
	lower := strings.ToLower(strings.TrimSpace(text))
	if lower == "" {
		return false
	}

	keywords := []string{
		"reserva",
		"booking",
		"codigo",
		"código",
		"status",
		"confirmad",
		"pagament",
		"expir",
		"pix",
	}
	for _, keyword := range keywords {
		if strings.Contains(lower, keyword) {
			return true
		}
	}
	return false
}

func looksLikePaymentLookupIntent(text string) bool {
	lower := strings.ToLower(strings.TrimSpace(text))
	if lower == "" {
		return false
	}

	keywords := []string{
		"pagament",
		"pix",
		"cobranc",
		"cobranç",
		"pago",
		"quitad",
		"checkout",
	}
	for _, keyword := range keywords {
		if strings.Contains(lower, keyword) {
			return true
		}
	}
	return false
}

func containsLetter(value string) bool {
	for _, char := range value {
		if (char >= 'A' && char <= 'Z') || (char >= 'a' && char <= 'z') {
			return true
		}
	}
	return false
}

func buildBookingLookupRequestPayload(input BookingLookupInput) map[string]interface{} {
	payload := map[string]interface{}{
		"limit": input.Limit,
	}
	if input.BookingID != "" {
		payload["booking_id"] = input.BookingID
	}
	if input.ReservationCode != "" {
		payload["reservation_code"] = input.ReservationCode
	}
	return payload
}

func buildBookingLookupResponsePayload(result BookingLookupResult) map[string]interface{} {
	items := make([]map[string]interface{}, 0, len(result.Results))
	for _, item := range result.Results {
		payload := map[string]interface{}{
			"booking_id":       item.ID,
			"trip_id":          item.TripID,
			"status":           item.Status,
			"reservation_code": item.ReservationCode,
			"total_amount":     item.TotalAmount,
			"deposit_amount":   item.DepositAmount,
			"remainder_amount": item.RemainderAmount,
			"passenger_name":   item.PassengerName,
			"passenger_phone":  item.PassengerPhone,
			"seat_number":      item.SeatNumber,
			"created_at":       item.CreatedAt.UTC().Format(time.RFC3339Nano),
		}
		if item.ExpiresAt != nil {
			payload["expires_at"] = item.ExpiresAt.UTC().Format(time.RFC3339Nano)
		}
		items = append(items, payload)
	}

	payload := map[string]interface{}{
		"result_count": len(result.Results),
		"results":      items,
	}
	if result.Filter.BookingID != "" {
		payload["booking_id"] = result.Filter.BookingID
	}
	if result.Filter.ReservationCode != "" {
		payload["reservation_code"] = result.Filter.ReservationCode
	}
	return payload
}

func buildPaymentStatusRequestPayload(input PaymentStatusInput) map[string]interface{} {
	payload := map[string]interface{}{
		"limit": input.Limit,
	}
	if input.BookingID != "" {
		payload["booking_id"] = input.BookingID
	}
	if input.ReservationCode != "" {
		payload["reservation_code"] = input.ReservationCode
	}
	return payload
}

func buildPaymentStatusResponsePayload(result PaymentStatusResult) map[string]interface{} {
	items := make([]map[string]interface{}, 0, len(result.Results))
	totalAmount := 0.0
	paidAmount := 0.0
	latestStatus := ""
	for index, item := range result.Results {
		payload := map[string]interface{}{
			"payment_id":   item.ID,
			"booking_id":   item.BookingID,
			"amount":       item.Amount,
			"method":       item.Method,
			"status":       item.Status,
			"provider":     item.Provider,
			"provider_ref": item.ProviderRef,
			"created_at":   item.CreatedAt.UTC().Format(time.RFC3339Nano),
		}
		if item.PaidAt != nil {
			payload["paid_at"] = item.PaidAt.UTC().Format(time.RFC3339Nano)
		}
		items = append(items, payload)
		totalAmount += item.Amount
		if strings.EqualFold(item.Status, "PAID") {
			paidAmount += item.Amount
		}
		if index == 0 {
			latestStatus = strings.TrimSpace(item.Status)
		}
	}

	payload := map[string]interface{}{
		"result_count":  len(result.Results),
		"results":       items,
		"total_amount":  totalAmount,
		"paid_amount":   paidAmount,
		"latest_status": latestStatus,
	}
	if result.Filter.BookingID != "" {
		payload["booking_id"] = result.Filter.BookingID
	}
	if result.Filter.ReservationCode != "" {
		payload["reservation_code"] = result.Filter.ReservationCode
	}
	return payload
}
