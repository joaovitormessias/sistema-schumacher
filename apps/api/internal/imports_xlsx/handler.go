package imports_xlsx

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	httpx "schumacher-tur/api/internal/shared/http"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/imports/xlsx", func(r chi.Router) {
		r.Post("/upload", h.upload)
		r.Post("/{batchId}/validate", h.validate)
		r.Post("/{batchId}/promote", h.promote)
		r.Get("/{batchId}/report", h.report)
	})
}

func (h *Handler) upload(w http.ResponseWriter, r *http.Request) {
	var input UploadInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_BODY", "invalid json", err.Error())
		return
	}
	if len(input.Sheets) == 0 {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "sheets is required", nil)
		return
	}

	result, err := h.svc.Upload(r.Context(), input)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "XLSX_UPLOAD_ERROR", "could not upload staging rows", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, result)
}

func (h *Handler) validate(w http.ResponseWriter, r *http.Request) {
	batchID := chi.URLParam(r, "batchId")
	if batchID == "" {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid batch id", nil)
		return
	}

	result, err := h.svc.Validate(r.Context(), batchID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "XLSX_VALIDATE_ERROR", "could not validate batch", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) promote(w http.ResponseWriter, r *http.Request) {
	batchID := chi.URLParam(r, "batchId")
	if batchID == "" {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid batch id", nil)
		return
	}

	result, err := h.svc.Promote(r.Context(), batchID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "XLSX_PROMOTE_ERROR", "could not promote batch", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) report(w http.ResponseWriter, r *http.Request) {
	batchID := chi.URLParam(r, "batchId")
	if batchID == "" {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid batch id", nil)
		return
	}

	result, err := h.svc.Report(r.Context(), batchID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "XLSX_REPORT_ERROR", "could not get report", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result)
}
