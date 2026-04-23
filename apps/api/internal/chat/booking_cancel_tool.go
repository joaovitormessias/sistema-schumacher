package chat

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"schumacher-tur/api/internal/bookings"
)

type BookingCancelTool struct {
	svc interface {
		Get(ctx context.Context, id string) (bookings.BookingDetails, error)
		UpdateStatus(ctx context.Context, id string, status string) (bookings.BookingDetails, error)
	}
}

func NewBookingCancelTool(svc interface {
	Get(ctx context.Context, id string) (bookings.BookingDetails, error)
	UpdateStatus(ctx context.Context, id string, status string) (bookings.BookingDetails, error)
}) *BookingCancelTool {
	return &BookingCancelTool{svc: svc}
}

func (t *BookingCancelTool) Enabled() bool {
	return t != nil && t.svc != nil
}

func (t *BookingCancelTool) Cancel(ctx context.Context, input BookingCancelInput) (BookingCancelResult, error) {
	result := BookingCancelResult{
		Filter: input,
		Reason: normalizeBookingCancelReason(input.Reason),
		Actor:  normalizeBookingCancelActor(input.Actor),
		Mode:   "manual_review_required_input_error",
	}
	if !t.Enabled() {
		return result, ErrAgentToolNotConfigured
	}

	bookingID := strings.TrimSpace(input.BookingID)
	if bookingID == "" {
		result.Errors = []string{"Nao foi possivel identificar a reserva para cancelar."}
		result.MessageForAgent = "Nao confirme cancelamento sem booking_id confirmado. Peca o codigo da reserva ou encaminhe para revisao humana."
		return result, nil
	}

	details, err := t.svc.Get(ctx, bookingID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			result.Mode = "manual_review_required_booking_not_found"
			result.Errors = []string{"Reserva nao encontrada."}
			result.MessageForAgent = "Nao foi possivel localizar a reserva. Peca confirmacao do codigo antes de tentar cancelar."
			return result, nil
		}
		return BookingCancelResult{}, err
	}

	populateBookingCancelResult(&result, details)
	result.PreviousStatus = result.BookingStatus

	switch result.BookingStatus {
	case "CANCELLED", "EXPIRED":
		result.Mode = "already_closed"
		result.Idempotent = true
		result.MessageForAgent = "A reserva ja estava encerrada. Responda de forma idempotente sem prometer nova alteracao."
		return result, nil
	}

	updated, err := t.svc.UpdateStatus(ctx, bookingID, "CANCELLED")
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			result.Mode = "manual_review_required_booking_not_found"
			result.Errors = []string{"Reserva nao encontrada no momento do cancelamento."}
			result.MessageForAgent = "Nao foi possivel concluir o cancelamento porque a reserva nao foi localizada novamente."
			return result, nil
		}
		return BookingCancelResult{}, err
	}

	populateBookingCancelResult(&result, updated)
	result.Mode = "cancel"
	result.MessageForAgent = "Cancelamento aplicado com sucesso. Confirme ao cliente que a reserva foi cancelada."
	return result, nil
}

func populateBookingCancelResult(result *BookingCancelResult, details bookings.BookingDetails) {
	result.BookingID = strings.TrimSpace(details.Booking.ID)
	result.ReservationCode = strings.TrimSpace(details.Booking.ReservationCode)
	result.TripID = strings.TrimSpace(details.Booking.TripID)
	result.BookingStatus = strings.ToUpper(strings.TrimSpace(details.Booking.Status))
	passengerCount := len(details.Passengers)
	if passengerCount == 0 && strings.TrimSpace(details.Passenger.ID) != "" {
		passengerCount = 1
	}
	result.PassengerCount = passengerCount
}

func normalizeBookingCancelReason(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "customer_requested", "customer_request":
		return "customer_requested"
	case "timeout", "timeout_50m", "payment_timeout":
		return "timeout_50m"
	default:
		return strings.ToLower(strings.TrimSpace(value))
	}
}

func normalizeBookingCancelActor(value string) string {
	if strings.TrimSpace(value) == "" {
		return "CUSTOMER"
	}
	return strings.ToUpper(strings.TrimSpace(value))
}
