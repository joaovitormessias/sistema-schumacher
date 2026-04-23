package chat

import (
	"context"
	"sort"
	"strings"

	"schumacher-tur/api/internal/reports"
)

type RescheduleAssistTool struct {
	bookings interface {
		Enabled() bool
		Search(ctx context.Context, input BookingLookupInput) (BookingLookupResult, error)
	}
	reports interface {
		ListPassengers(ctx context.Context, filter reports.PassengerReportFilter) ([]reports.PassengerReportRow, error)
	}
	availability interface {
		Enabled() bool
		Search(ctx context.Context, input AvailabilitySearchInput) (AvailabilitySearchResult, error)
	}
}

func NewRescheduleAssistTool(
	bookings interface {
		Enabled() bool
		Search(ctx context.Context, input BookingLookupInput) (BookingLookupResult, error)
	},
	reportsSvc interface {
		ListPassengers(ctx context.Context, filter reports.PassengerReportFilter) ([]reports.PassengerReportRow, error)
	},
	availability interface {
		Enabled() bool
		Search(ctx context.Context, input AvailabilitySearchInput) (AvailabilitySearchResult, error)
	},
) *RescheduleAssistTool {
	return &RescheduleAssistTool{
		bookings:     bookings,
		reports:      reportsSvc,
		availability: availability,
	}
}

func (t *RescheduleAssistTool) Enabled() bool {
	return t != nil &&
		t.bookings != nil &&
		t.bookings.Enabled() &&
		t.reports != nil &&
		t.availability != nil &&
		t.availability.Enabled()
}

func (t *RescheduleAssistTool) Search(ctx context.Context, input RescheduleAssistInput) (RescheduleAssistResult, error) {
	result := RescheduleAssistResult{
		Filter:              input,
		NextStep:            "revisao_humana",
		HumanReviewRequired: true,
		CanAutoReschedule:   false,
		Options:             []RescheduleAssistOption{},
	}
	if !t.Enabled() {
		return result, ErrAgentToolNotConfigured
	}

	if strings.TrimSpace(input.BookingID) == "" && strings.TrimSpace(input.ReservationCode) == "" {
		result.Mode = "manual_review_required_input_error"
		result.Errors = []string{"Informe booking_id ou reservation_code."}
		result.MessageForAgent = "Nao confirme reagendamento automaticamente. Peca a identificacao correta da reserva ou encaminhe para revisao humana."
		return result, nil
	}
	if input.RequestedTripDate == nil {
		result.Mode = "manual_review_required_input_error"
		result.Errors = []string{"Informe a nova data do reagendamento."}
		result.MessageForAgent = "Nao confirme reagendamento automaticamente. Peca a nova data ao cliente ou encaminhe para revisao humana."
		return result, nil
	}

	bookingResult, err := t.bookings.Search(ctx, BookingLookupInput{
		BookingID:       strings.TrimSpace(input.BookingID),
		ReservationCode: strings.ToUpper(strings.TrimSpace(input.ReservationCode)),
		Limit:           1,
	})
	if err != nil {
		return RescheduleAssistResult{}, err
	}
	if len(bookingResult.Results) == 0 {
		result.Mode = "manual_review_required_booking_not_found"
		result.Errors = []string{"Reserva nao encontrada."}
		result.MessageForAgent = "Nao foi possivel localizar a reserva. Nao confirme reagendamento; peca confirmacao do codigo ou encaminhe para revisao humana."
		return result, nil
	}

	booking := bookingResult.Results[0]
	result.Booking = &RescheduleAssistBooking{
		ID:              strings.TrimSpace(booking.ID),
		TripID:          strings.TrimSpace(booking.TripID),
		Status:          strings.ToUpper(strings.TrimSpace(booking.Status)),
		ReservationCode: strings.TrimSpace(booking.ReservationCode),
	}

	rows, err := t.reports.ListPassengers(ctx, reports.PassengerReportFilter{
		BookingID:       strings.TrimSpace(booking.ID),
		ReservationCode: strings.TrimSpace(booking.ReservationCode),
	})
	if err != nil {
		return RescheduleAssistResult{}, err
	}

	currentOrigin, currentDestination, currentTripDate := deriveCurrentRouteFromPassengerReport(rows)
	passengerCount := len(rows)
	requestedOrigin := firstNonEmpty(strings.TrimSpace(input.RequestedOrigin), currentOrigin)
	requestedDestination := firstNonEmpty(strings.TrimSpace(input.RequestedDestination), currentDestination)
	requestedQty := input.RequestedQty
	if requestedQty <= 0 {
		requestedQty = passengerCount
	}
	if requestedQty <= 0 {
		requestedQty = 1
	}

	result.Current = RescheduleAssistRoute{
		Origin:         currentOrigin,
		Destination:    currentDestination,
		TripDate:       currentTripDate,
		PassengerCount: passengerCount,
	}
	result.Requested = RescheduleAssistRequest{
		Origin:      requestedOrigin,
		Destination: requestedDestination,
		TripDate:    input.RequestedTripDate.UTC().Format("2006-01-02"),
		Qty:         requestedQty,
	}

	if requestedOrigin == "" || requestedDestination == "" {
		result.Mode = "manual_review_required_route_error"
		result.Errors = []string{"Nao foi possivel identificar a rota para a nova busca."}
		result.MessageForAgent = "Nao foi possivel montar a rota do reagendamento. Nao confirme a troca; peca confirmacao da origem e do destino ou encaminhe para revisao humana."
		return result, nil
	}

	if isRescheduleBlockedStatus(result.Booking.Status) {
		result.Mode = "manual_review_required_booking_ineligible"
		result.MessageForAgent = "A reserva localizada esta em status impeditivo. Nao confirme reagendamento; informe que o caso precisa de analise humana."
		return result, nil
	}

	availabilityResult, err := t.availability.Search(ctx, AvailabilitySearchInput{
		Origin:      requestedOrigin,
		Destination: requestedDestination,
		TripDate:    input.RequestedTripDate,
		Qty:         requestedQty,
		Limit:       5,
	})
	if err != nil {
		return RescheduleAssistResult{}, err
	}

	result.Options = mapRescheduleAvailabilityOptions(availabilityResult.Results)
	if len(result.Options) == 0 {
		result.Mode = "manual_review_required_no_availability"
		result.MessageForAgent = "Nao foi encontrada disponibilidade para a nova data pesquisada. Nao confirme reagendamento; ofereca apoio humano."
		return result, nil
	}

	result.Mode = "manual_review_required_with_options"
	result.FieldsRequiredForManualCompletion = []string{"trip_id", "board_stop_id", "alight_stop_id"}
	result.MessageForAgent = "Ha opcoes de viagem para a nova data. Apresente as alternativas, mas nao confirme o reagendamento como concluido; encaminhe para revisao humana."
	return result, nil
}

func deriveCurrentRouteFromPassengerReport(rows []reports.PassengerReportRow) (string, string, string) {
	origins := make([]string, 0, len(rows))
	destinations := make([]string, 0, len(rows))
	seenOrigins := map[string]struct{}{}
	seenDestinations := map[string]struct{}{}
	tripDate := ""
	for _, row := range rows {
		origin := strings.TrimSpace(row.Origin)
		if origin != "" {
			key := strings.ToUpper(origin)
			if _, ok := seenOrigins[key]; !ok {
				seenOrigins[key] = struct{}{}
				origins = append(origins, origin)
			}
		}
		destination := strings.TrimSpace(row.Destination)
		if destination != "" {
			key := strings.ToUpper(destination)
			if _, ok := seenDestinations[key]; !ok {
				seenDestinations[key] = struct{}{}
				destinations = append(destinations, destination)
			}
		}
		if tripDate == "" {
			tripDate = strings.TrimSpace(row.TripDate)
		}
	}

	var origin string
	if len(origins) > 0 {
		origin = origins[0]
	}
	var destination string
	if len(destinations) > 0 {
		destination = destinations[0]
	}
	return origin, destination, tripDate
}

func mapRescheduleAvailabilityOptions(items []AvailabilitySearchItem) []RescheduleAssistOption {
	sorted := append([]AvailabilitySearchItem(nil), items...)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].TripDate != sorted[j].TripDate {
			return sorted[i].TripDate < sorted[j].TripDate
		}
		return sorted[i].OriginDepartTime < sorted[j].OriginDepartTime
	})
	if len(sorted) > 5 {
		sorted = sorted[:5]
	}

	options := make([]RescheduleAssistOption, 0, len(sorted))
	for _, item := range sorted {
		options = append(options, RescheduleAssistOption{
			TripID:         strings.TrimSpace(item.TripID),
			TripDate:       strings.TrimSpace(item.TripDate),
			DepartureTime:  strings.TrimSpace(item.OriginDepartTime),
			Origin:         strings.TrimSpace(item.OriginDisplayName),
			Destination:    strings.TrimSpace(item.DestinationDisplayName),
			BoardStopID:    strings.TrimSpace(item.BoardStopID),
			AlightStopID:   strings.TrimSpace(item.AlightStopID),
			SeatsAvailable: item.SeatsAvailable,
			Price:          item.Price,
			Currency:       strings.TrimSpace(item.Currency),
			PackageName:    strings.TrimSpace(item.PackageName),
		})
	}
	return options
}

func isRescheduleBlockedStatus(status string) bool {
	switch strings.ToUpper(strings.TrimSpace(status)) {
	case "CANCELLED", "EXPIRED":
		return true
	default:
		return false
	}
}
