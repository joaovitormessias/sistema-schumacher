package reports

import (
  "encoding/csv"
  "fmt"
  "net/http"

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
  tripID := r.URL.Query().Get("trip_id")
  if tripID == "" {
    httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "trip_id is required", nil)
    return
  }
  format := r.URL.Query().Get("format")
  if format == "" {
    format = "json"
  }

  rows, err := h.svc.ListPassengers(r.Context(), tripID)
  if err != nil {
    httpx.WriteError(w, http.StatusInternalServerError, "REPORT_ERROR", "could not generate report", err.Error())
    return
  }

  if format == "csv" {
    w.Header().Set("Content-Type", "text/csv")
    w.Header().Set("Content-Disposition", "attachment; filename=manifesto.csv")

    writer := csv.NewWriter(w)
    _ = writer.Write([]string{
      "seat_number",
      "name",
      "document",
      "phone",
      "email",
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
        intToString(row.SeatNumber),
        row.Name,
        row.Document,
        row.Phone,
        row.Email,
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

func intToString(v int) string {
  return fmt.Sprintf("%d", v)
}

func floatToString(v float64) string {
  return fmt.Sprintf("%.2f", v)
}
