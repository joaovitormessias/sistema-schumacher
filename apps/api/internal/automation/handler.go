package automation

import (
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"schumacher-tur/api/internal/payments"
	httpx "schumacher-tur/api/internal/shared/http"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterWebhooks(r chi.Router) {
	r.Route("/webhooks/evolution", func(r chi.Router) {
		r.Post("/messages", h.handleEvolutionMessages)
		r.Post("/status", h.handleEvolutionStatus)
		r.Post("/presence", h.handleEvolutionPresence)
	})
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/automation/jobs", func(r chi.Router) {
		r.Post("/chat-review-alerts/run", h.runChatReviewAlerts)
		r.Post("/bookings-expire/run", h.runBookingsExpire)
		r.Post("/payment-notifications/run", h.runPaymentNotifications)
		r.Post("/sheet-sync/run", h.runSheetSync)
		r.Get("/{jobName}/runs", h.listJobRuns)
	})
}

func (h *Handler) handleEvolutionMessages(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_BODY", "could not read request body", err.Error())
		return
	}

	result, err := h.svc.HandleEvolutionMessages(r.Context(), body)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidEvolutionPayload), errors.Is(err, ErrMissingEvolutionChatKey):
			httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
		default:
			httpx.WriteError(w, http.StatusInternalServerError, "EVOLUTION_WEBHOOK_ERROR", "could not process evolution webhook", err.Error())
		}
		return
	}

	httpx.WriteJSON(w, http.StatusAccepted, result)
}

func (h *Handler) handleEvolutionStatus(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_BODY", "could not read request body", err.Error())
		return
	}

	result, err := h.svc.HandleEvolutionStatus(r.Context(), body)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidEvolutionPayload), errors.Is(err, ErrMissingEvolutionMessageID), errors.Is(err, ErrMissingEvolutionStatus):
			httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
		default:
			httpx.WriteError(w, http.StatusInternalServerError, "EVOLUTION_STATUS_WEBHOOK_ERROR", "could not process evolution status webhook", err.Error())
		}
		return
	}

	httpx.WriteJSON(w, http.StatusAccepted, result)
}

func (h *Handler) handleEvolutionPresence(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_BODY", "could not read request body", err.Error())
		return
	}

	result, err := h.svc.HandleEvolutionPresence(r.Context(), body)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidEvolutionPayload), errors.Is(err, ErrMissingEvolutionChatKey), errors.Is(err, ErrMissingEvolutionPresence):
			httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
		default:
			httpx.WriteError(w, http.StatusInternalServerError, "EVOLUTION_PRESENCE_WEBHOOK_ERROR", "could not process evolution presence webhook", err.Error())
		}
		return
	}

	httpx.WriteJSON(w, http.StatusAccepted, result)
}

func (h *Handler) runBookingsExpire(w http.ResponseWriter, r *http.Request) {
	var input RunBookingsExpireInput
	if err := httpx.DecodeJSON(r, &input); err != nil && !errors.Is(err, httpx.ErrEmptyBody) {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_BODY", "invalid json", err.Error())
		return
	}

	result, err := h.svc.RunBookingsExpire(r.Context(), input)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidBookingsExpireLimit), errors.Is(err, ErrInvalidBookingsExpireHold):
			httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
		default:
			httpx.WriteError(w, http.StatusInternalServerError, "BOOKINGS_EXPIRE_ERROR", "could not expire pending bookings", err.Error())
		}
		return
	}

	httpx.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) runChatReviewAlerts(w http.ResponseWriter, r *http.Request) {
	var input RunChatReviewAlertsInput
	if err := httpx.DecodeJSON(r, &input); err != nil && !errors.Is(err, httpx.ErrEmptyBody) {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_BODY", "invalid json", err.Error())
		return
	}

	result, err := h.svc.RunChatReviewAlerts(r.Context(), input)
	if err != nil {
		httpx.WriteError(w, http.StatusBadGateway, "CHAT_REVIEW_ALERT_ERROR", "could not send chat review alert", err.Error())
		return
	}

	httpx.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) runPaymentNotifications(w http.ResponseWriter, r *http.Request) {
	var input RunPaymentNotificationsInput
	if err := httpx.DecodeJSON(r, &input); err != nil && !errors.Is(err, httpx.ErrEmptyBody) {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_BODY", "invalid json", err.Error())
		return
	}

	result, err := h.svc.RunPaymentNotifications(r.Context(), input)
	if err != nil {
		switch {
		case errors.Is(err, ErrPaymentNotificationTarget):
			httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
		case payments.IsNotFound(err):
			httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "payment not found", nil)
		default:
			httpx.WriteError(w, http.StatusInternalServerError, "PAYMENT_NOTIFICATION_ERROR", "could not queue payment notification draft", err.Error())
		}
		return
	}

	httpx.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) runSheetSync(w http.ResponseWriter, r *http.Request) {
	httpx.NotImplemented(w, r)
}

func (h *Handler) listJobRuns(w http.ResponseWriter, r *http.Request) {
	limit := 0
	if raw := r.URL.Query().Get("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "INVALID_QUERY", "invalid limit", err.Error())
			return
		}
		limit = parsed
	}

	result, err := h.svc.ListJobRuns(r.Context(), ListJobRunsInput{
		JobName:       chi.URLParam(r, "jobName"),
		Status:        r.URL.Query().Get("status"),
		TriggerSource: r.URL.Query().Get("trigger_source"),
		Limit:         limit,
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrMissingJobName), errors.Is(err, ErrInvalidJobRunLimit):
			httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
		default:
			httpx.WriteError(w, http.StatusInternalServerError, "AUTOMATION_JOB_RUNS_ERROR", "could not list automation job runs", err.Error())
		}
		return
	}

	httpx.WriteJSON(w, http.StatusOK, result)
}
