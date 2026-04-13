package reports

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	httpx "schumacher-tur/api/internal/shared/http"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/reports", func(r chi.Router) {
		r.Get("/passengers", h.passengers)
	})
}

func (h *Handler) passengers(w http.ResponseWriter, r *http.Request) {
	filter := PassengerReportFilter{
		TripID:          strings.TrimSpace(r.URL.Query().Get("trip_id")),
		TripDate:        strings.TrimSpace(r.URL.Query().Get("trip_date")),
		BookingID:       strings.TrimSpace(r.URL.Query().Get("booking_id")),
		ReservationCode: strings.TrimSpace(r.URL.Query().Get("reservation_code")),
		IncludeCanceled: isTrueLike(r.URL.Query().Get("include_canceled")),
	}
	if filter.TripID == "" && filter.TripDate == "" && filter.BookingID == "" && filter.ReservationCode == "" {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "trip_id, trip_date, booking_id or reservation_code is required", nil)
		return
	}
	format := strings.TrimSpace(r.URL.Query().Get("format"))
	if format == "" {
		format = "json"
	}

	rows, err := h.svc.ListPassengers(r.Context(), filter)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "REPORT_ERROR", "could not generate report", err.Error())
		return
	}

	if format == "csv" {
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", "attachment; filename=manifesto.csv")

		writer := csv.NewWriter(w)
		_ = writer.Write([]string{
			"trip_date",
			"trip_id",
			"booking_id",
			"reservation_code",
			"seat_number",
			"name",
			"document",
			"phone",
			"email",
			"origin",
			"destination",
			"booking_status",
			"passenger_status",
			"total_amount",
			"deposit_amount",
			"remainder_amount",
			"amount_paid",
			"payment_stage",
		})

		for _, row := range rows {
			_ = writer.Write([]string{
				row.TripDate,
				row.TripID,
				row.BookingID,
				row.ReservationCode,
				row.SeatNumber,
				row.Name,
				row.Document,
				row.Phone,
				row.Email,
				row.Origin,
				row.Destination,
				row.BookingStatus,
				row.PassengerStatus,
				floatToString(row.TotalAmount),
				floatToString(row.DepositAmount),
				floatToString(row.RemainderAmount),
				floatToString(row.AmountPaid),
				row.PaymentStage,
			})
		}

		writer.Flush()
		if err := writer.Error(); err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, "CSV_ERROR", "could not write csv", err.Error())
			return
		}
		return
	}

	if format != "json" {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_FORMAT", "format must be json or csv", nil)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, rows)
}

func floatToString(v float64) string {
	return fmt.Sprintf("%.2f", v)
}

func isTrueLike(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "sim", "yes":
		return true
	default:
		return false
	}
}
