package trip_operations

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"schumacher-tur/api/internal/auth"
	httpx "schumacher-tur/api/internal/shared/http"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/trip-requests", func(r chi.Router) {
		r.Get("/", h.listTripRequests)
		r.Post("/", h.createTripRequest)
	})

	r.Get("/trips/{tripId}/manifest", h.listManifest)
	r.Post("/trips/{tripId}/manifest", h.createManifest)
	r.Post("/trips/{tripId}/manifest/sync", h.syncManifest)
	r.Patch("/trips/{tripId}/manifest/{entryId}", h.updateManifest)

	r.Get("/trips/{tripId}/authorizations", h.listAuthorizations)
	r.Post("/trips/{tripId}/authorizations", h.createAuthorization)
	r.Patch("/trips/{tripId}/authorizations/{authorizationId}", h.updateAuthorization)

	r.Get("/trips/{tripId}/checklists/{stage}", h.getChecklist)
	r.Put("/trips/{tripId}/checklists/{stage}", h.upsertChecklist)

	r.Get("/trips/{tripId}/driver-report", h.getDriverReport)
	r.Put("/trips/{tripId}/driver-report", h.upsertDriverReport)

	r.Get("/trips/{tripId}/reconciliation", h.getReconciliation)
	r.Put("/trips/{tripId}/reconciliation", h.upsertReconciliation)

	r.Get("/trips/{tripId}/attachments", h.listAttachments)
	r.Post("/trips/{tripId}/attachments", h.createAttachment)

	r.Post("/trips/{tripId}/workflow/advance", h.advanceWorkflow)
}

func (h *Handler) listTripRequests(w http.ResponseWriter, r *http.Request) {
	limit, offset, err := parseLimitOffset(r)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_QUERY", "invalid query parameters", nil)
		return
	}
	items, err := h.svc.ListTripRequests(r.Context(), limit, offset)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "TRIP_REQUESTS_LIST_ERROR", "could not list trip requests", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusOK, items)
}

func (h *Handler) createTripRequest(w http.ResponseWriter, r *http.Request) {
	var input CreateTripRequestInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_BODY", "invalid json", err.Error())
		return
	}
	var createdBy *string
	if userID, ok := auth.UserIDFromContext(r.Context()); ok {
		createdBy = &userID
	}
	item, err := h.svc.CreateTripRequest(r.Context(), input, createdBy)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidSource), errors.Is(err, ErrInvalidStatus):
			httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
		default:
			httpx.WriteError(w, http.StatusInternalServerError, "TRIP_REQUEST_CREATE_ERROR", "could not create trip request", err.Error())
		}
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, item)
}

func (h *Handler) listManifest(w http.ResponseWriter, r *http.Request) {
	tripID, err := httpx.ParseUUIDParam(r, "tripId")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid trip id", nil)
		return
	}
	items, err := h.svc.ListManifestEntries(r.Context(), tripID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "MANIFEST_LIST_ERROR", "could not list trip manifest", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusOK, items)
}

func (h *Handler) createManifest(w http.ResponseWriter, r *http.Request) {
	tripID, err := httpx.ParseUUIDParam(r, "tripId")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid trip id", nil)
		return
	}
	var input CreateManifestEntryInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_BODY", "invalid json", err.Error())
		return
	}
	item, err := h.svc.CreateManifestEntry(r.Context(), tripID, input)
	if err != nil {
		if errors.Is(err, ErrInvalidStatus) {
			httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "MANIFEST_CREATE_ERROR", "could not create manifest entry", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, item)
}

func (h *Handler) syncManifest(w http.ResponseWriter, r *http.Request) {
	tripID, err := httpx.ParseUUIDParam(r, "tripId")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid trip id", nil)
		return
	}
	items, err := h.svc.SyncManifestFromBookings(r.Context(), tripID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "MANIFEST_SYNC_ERROR", "could not sync manifest", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusOK, items)
}

func (h *Handler) updateManifest(w http.ResponseWriter, r *http.Request) {
	tripID, err := httpx.ParseUUIDParam(r, "tripId")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid trip id", nil)
		return
	}
	entryID, err := httpx.ParseUUIDParam(r, "entryId")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid manifest entry id", nil)
		return
	}
	var input UpdateManifestEntryInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_BODY", "invalid json", err.Error())
		return
	}
	item, err := h.svc.UpdateManifestEntry(r.Context(), tripID, entryID, input)
	if err != nil {
		if errors.Is(err, ErrInvalidStatus) {
			httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
			return
		}
		if IsNotFound(err) {
			httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "manifest entry not found", nil)
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "MANIFEST_UPDATE_ERROR", "could not update manifest entry", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) listAuthorizations(w http.ResponseWriter, r *http.Request) {
	tripID, err := httpx.ParseUUIDParam(r, "tripId")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid trip id", nil)
		return
	}
	items, err := h.svc.ListTripAuthorizations(r.Context(), tripID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "AUTHORIZATION_LIST_ERROR", "could not list authorizations", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusOK, items)
}

func (h *Handler) createAuthorization(w http.ResponseWriter, r *http.Request) {
	tripID, err := httpx.ParseUUIDParam(r, "tripId")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid trip id", nil)
		return
	}
	var input CreateTripAuthorizationInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_BODY", "invalid json", err.Error())
		return
	}
	var createdBy *string
	if userID, ok := auth.UserIDFromContext(r.Context()); ok {
		createdBy = &userID
	}
	item, err := h.svc.CreateTripAuthorization(r.Context(), tripID, input, createdBy)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidAuthority), errors.Is(err, ErrInvalidAuthorization):
			httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
		default:
			httpx.WriteError(w, http.StatusInternalServerError, "AUTHORIZATION_CREATE_ERROR", "could not create authorization", err.Error())
		}
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, item)
}

func (h *Handler) updateAuthorization(w http.ResponseWriter, r *http.Request) {
	tripID, err := httpx.ParseUUIDParam(r, "tripId")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid trip id", nil)
		return
	}
	authorizationID, err := httpx.ParseUUIDParam(r, "authorizationId")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid authorization id", nil)
		return
	}
	var input UpdateTripAuthorizationInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_BODY", "invalid json", err.Error())
		return
	}
	item, err := h.svc.UpdateTripAuthorization(r.Context(), tripID, authorizationID, input)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidAuthority), errors.Is(err, ErrInvalidAuthorization):
			httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
		case IsNotFound(err):
			httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "authorization not found", nil)
		default:
			httpx.WriteError(w, http.StatusInternalServerError, "AUTHORIZATION_UPDATE_ERROR", "could not update authorization", err.Error())
		}
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) getChecklist(w http.ResponseWriter, r *http.Request) {
	tripID, stage, ok := parseTripAndStage(w, r)
	if !ok {
		return
	}
	item, err := h.svc.GetTripChecklist(r.Context(), tripID, stage)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidStage):
			httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
		case IsNotFound(err):
			httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "checklist not found", nil)
		default:
			httpx.WriteError(w, http.StatusInternalServerError, "CHECKLIST_GET_ERROR", "could not get checklist", err.Error())
		}
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) upsertChecklist(w http.ResponseWriter, r *http.Request) {
	tripID, stage, ok := parseTripAndStage(w, r)
	if !ok {
		return
	}
	var input UpsertChecklistInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_BODY", "invalid json", err.Error())
		return
	}
	var completedBy *string
	if userID, ok := auth.UserIDFromContext(r.Context()); ok {
		completedBy = &userID
	}
	item, err := h.svc.UpsertTripChecklist(r.Context(), tripID, stage, input, completedBy)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidStage):
			httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
		default:
			httpx.WriteError(w, http.StatusInternalServerError, "CHECKLIST_UPSERT_ERROR", "could not save checklist", err.Error())
		}
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) getDriverReport(w http.ResponseWriter, r *http.Request) {
	tripID, err := httpx.ParseUUIDParam(r, "tripId")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid trip id", nil)
		return
	}
	item, err := h.svc.GetTripDriverReport(r.Context(), tripID)
	if err != nil {
		if IsNotFound(err) {
			httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "driver report not found", nil)
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "DRIVER_REPORT_GET_ERROR", "could not get driver report", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) upsertDriverReport(w http.ResponseWriter, r *http.Request) {
	tripID, err := httpx.ParseUUIDParam(r, "tripId")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid trip id", nil)
		return
	}
	var input UpsertDriverReportInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_BODY", "invalid json", err.Error())
		return
	}
	var submittedBy *string
	if userID, ok := auth.UserIDFromContext(r.Context()); ok {
		submittedBy = &userID
	}
	item, err := h.svc.UpsertTripDriverReport(r.Context(), tripID, input, submittedBy)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "DRIVER_REPORT_UPSERT_ERROR", "could not save driver report", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) getReconciliation(w http.ResponseWriter, r *http.Request) {
	tripID, err := httpx.ParseUUIDParam(r, "tripId")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid trip id", nil)
		return
	}
	item, err := h.svc.GetTripReceiptReconciliation(r.Context(), tripID)
	if err != nil {
		if IsNotFound(err) {
			httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "reconciliation not found", nil)
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "RECONCILIATION_GET_ERROR", "could not get reconciliation", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) upsertReconciliation(w http.ResponseWriter, r *http.Request) {
	tripID, err := httpx.ParseUUIDParam(r, "tripId")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid trip id", nil)
		return
	}
	var input UpsertReceiptReconciliationInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_BODY", "invalid json", err.Error())
		return
	}
	var reconciledBy *string
	if userID, ok := auth.UserIDFromContext(r.Context()); ok {
		reconciledBy = &userID
	}
	item, err := h.svc.UpsertTripReceiptReconciliation(r.Context(), tripID, input, reconciledBy)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "RECONCILIATION_UPSERT_ERROR", err.Error(), nil)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) listAttachments(w http.ResponseWriter, r *http.Request) {
	tripID, err := httpx.ParseUUIDParam(r, "tripId")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid trip id", nil)
		return
	}
	items, err := h.svc.ListTripAttachments(r.Context(), tripID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "ATTACHMENTS_LIST_ERROR", "could not list attachments", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusOK, items)
}

func (h *Handler) createAttachment(w http.ResponseWriter, r *http.Request) {
	tripID, err := httpx.ParseUUIDParam(r, "tripId")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid trip id", nil)
		return
	}
	var input CreateAttachmentInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_BODY", "invalid json", err.Error())
		return
	}
	var uploadedBy *string
	if userID, ok := auth.UserIDFromContext(r.Context()); ok {
		uploadedBy = &userID
	}
	item, err := h.svc.CreateTripAttachment(r.Context(), tripID, input, uploadedBy)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "ATTACHMENT_CREATE_ERROR", err.Error(), nil)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, item)
}

func (h *Handler) advanceWorkflow(w http.ResponseWriter, r *http.Request) {
	tripID, err := httpx.ParseUUIDParam(r, "tripId")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid trip id", nil)
		return
	}
	var input WorkflowAdvanceInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_BODY", "invalid json", err.Error())
		return
	}
	var actorID *string
	if userID, ok := auth.UserIDFromContext(r.Context()); ok {
		actorID = &userID
	}
	item, err := h.svc.AdvanceWorkflow(r.Context(), tripID, input.ToStatus, actorID)
	if err != nil {
		switch typed := err.(type) {
		case WorkflowBlockedError:
			httpx.WriteJSON(w, http.StatusConflict, typed.Response)
		default:
			if errors.Is(err, ErrInvalidOperationalStatus) {
				httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
				return
			}
			if IsNotFound(err) {
				httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "trip not found", nil)
				return
			}
			httpx.WriteError(w, http.StatusInternalServerError, "WORKFLOW_ADVANCE_ERROR", "could not advance workflow", err.Error())
		}
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func parseTripAndStage(w http.ResponseWriter, r *http.Request) (uuid.UUID, string, bool) {
	tripID, err := httpx.ParseUUIDParam(r, "tripId")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid trip id", nil)
		return uuid.Nil, "", false
	}
	stage := strings.ToUpper(strings.TrimSpace(chi.URLParam(r, "stage")))
	if stage == "" {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "stage is required", nil)
		return uuid.Nil, "", false
	}
	return tripID, stage, true
}

func parseLimitOffset(r *http.Request) (int, int, error) {
	limit := 200
	offset := 0
	q := r.URL.Query()
	if v := q.Get("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n <= 0 {
			return 0, 0, errors.New("invalid limit")
		}
		limit = n
	}
	if v := q.Get("offset"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			return 0, 0, errors.New("invalid offset")
		}
		offset = n
	}
	return limit, offset, nil
}
