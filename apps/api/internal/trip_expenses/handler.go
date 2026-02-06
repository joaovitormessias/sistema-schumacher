package trip_expenses

import (
  "errors"
  "net/http"
  "strconv"

  "github.com/go-chi/chi/v5"
  "github.com/google/uuid"

  "schumacher-tur/api/internal/auth"
  httpx "schumacher-tur/api/internal/shared/http"
)

type Handler struct {
  svc *Service
}

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) RegisterRoutes(r chi.Router) {
  r.Route("/trip-expenses", func(r chi.Router) {
    r.Get("/", h.list)
    r.Post("/", h.create)
    r.Get("/{expenseId}", h.get)
    r.Patch("/{expenseId}", h.update)
    r.Post("/{expenseId}/approve", h.approve)
  })
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
  filter, err := parseListFilter(r)
  if err != nil {
    httpx.WriteError(w, http.StatusBadRequest, "INVALID_QUERY", "invalid query parameters", nil)
    return
  }
  items, err := h.svc.List(r.Context(), filter)
  if err != nil {
    httpx.WriteError(w, http.StatusInternalServerError, "EXPENSES_LIST_ERROR", "could not list expenses", err.Error())
    return
  }
  httpx.WriteJSON(w, http.StatusOK, items)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
  id, err := httpx.ParseUUIDParam(r, "expenseId")
  if err != nil {
    httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid expense id", nil)
    return
  }
  item, err := h.svc.Get(r.Context(), id)
  if err != nil {
    if IsNotFound(err) {
      httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "expense not found", nil)
      return
    }
    httpx.WriteError(w, http.StatusInternalServerError, "EXPENSE_GET_ERROR", "could not get expense", err.Error())
    return
  }
  httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
  var input CreateTripExpenseInput
  if err := httpx.DecodeJSON(r, &input); err != nil {
    httpx.WriteError(w, http.StatusBadRequest, "INVALID_BODY", "invalid json", err.Error())
    return
  }
  if input.TripID == "" || input.DriverID == "" || input.ExpenseType == "" || input.Description == "" {
    httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "trip_id, driver_id, expense_type and description are required", nil)
    return
  }
  if _, err := uuid.Parse(input.TripID); err != nil {
    httpx.WriteError(w, http.StatusBadRequest, "INVALID_TRIP", "invalid trip_id", nil)
    return
  }
  if _, err := uuid.Parse(input.DriverID); err != nil {
    httpx.WriteError(w, http.StatusBadRequest, "INVALID_DRIVER", "invalid driver_id", nil)
    return
  }
  if !isValidExpenseType(input.ExpenseType) {
    httpx.WriteError(w, http.StatusBadRequest, "INVALID_TYPE", "invalid expense_type", nil)
    return
  }
  if input.Amount < 0 {
    httpx.WriteError(w, http.StatusBadRequest, "INVALID_AMOUNT", "amount must be positive", nil)
    return
  }
  if input.ExpenseDate.IsZero() {
    httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "expense_date is required", nil)
    return
  }
  if input.PaymentMethod != "" && !isValidPaymentMethod(input.PaymentMethod) {
    httpx.WriteError(w, http.StatusBadRequest, "INVALID_PAYMENT", "invalid payment_method", nil)
    return
  }
  if input.PaymentMethod == "CARD" {
    if input.DriverCardID == nil || *input.DriverCardID == "" {
      httpx.WriteError(w, http.StatusBadRequest, "CARD_REQUIRED", "driver_card_id is required for card expenses", nil)
      return
    }
    if _, err := uuid.Parse(*input.DriverCardID); err != nil {
      httpx.WriteError(w, http.StatusBadRequest, "INVALID_CARD", "invalid driver_card_id", nil)
      return
    }
  } else if input.DriverCardID != nil {
    httpx.WriteError(w, http.StatusBadRequest, "CARD_NOT_ALLOWED", "driver_card_id only allowed for card payments", nil)
    return
  }

  var createdBy *string
  if userID, ok := auth.UserIDFromContext(r.Context()); ok {
    createdBy = &userID
  }

  item, err := h.svc.Create(r.Context(), input, createdBy, createdBy)
  if err != nil {
    switch {
    case errors.Is(err, ErrInvalidPaymentMethod):
      httpx.WriteError(w, http.StatusBadRequest, "INVALID_PAYMENT", "invalid payment_method", nil)
      return
    case errors.Is(err, ErrCardRequired):
      httpx.WriteError(w, http.StatusBadRequest, "CARD_REQUIRED", "driver_card_id is required", nil)
      return
    case errors.Is(err, ErrInvalidAmount):
      httpx.WriteError(w, http.StatusBadRequest, "INVALID_AMOUNT", "amount must be positive", nil)
      return
    case errors.Is(err, ErrInsufficientBalance):
      httpx.WriteError(w, http.StatusConflict, "INSUFFICIENT_BALANCE", "insufficient card balance", nil)
      return
    case errors.Is(err, ErrCardInactive):
      httpx.WriteError(w, http.StatusConflict, "CARD_INACTIVE", "card is inactive", nil)
      return
    case errors.Is(err, ErrCardBlocked):
      httpx.WriteError(w, http.StatusConflict, "CARD_BLOCKED", "card is blocked", nil)
      return
    }
    httpx.WriteError(w, http.StatusInternalServerError, "EXPENSE_CREATE_ERROR", "could not create expense", err.Error())
    return
  }
  httpx.WriteJSON(w, http.StatusCreated, item)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
  id, err := httpx.ParseUUIDParam(r, "expenseId")
  if err != nil {
    httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid expense id", nil)
    return
  }

  var input UpdateTripExpenseInput
  if err := httpx.DecodeJSON(r, &input); err != nil {
    httpx.WriteError(w, http.StatusBadRequest, "INVALID_BODY", "invalid json", err.Error())
    return
  }
  if input.Amount != nil && *input.Amount < 0 {
    httpx.WriteError(w, http.StatusBadRequest, "INVALID_AMOUNT", "amount must be positive", nil)
    return
  }

  item, err := h.svc.Update(r.Context(), id, input)
  if err != nil {
    if IsNotFound(err) {
      httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "expense not found", nil)
      return
    }
    if errors.Is(err, ErrExpenseLocked) {
      httpx.WriteError(w, http.StatusConflict, "EXPENSE_LOCKED", "approved expenses cannot be edited", nil)
      return
    }
    httpx.WriteError(w, http.StatusInternalServerError, "EXPENSE_UPDATE_ERROR", "could not update expense", err.Error())
    return
  }
  httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) approve(w http.ResponseWriter, r *http.Request) {
  id, err := httpx.ParseUUIDParam(r, "expenseId")
  if err != nil {
    httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid expense id", nil)
    return
  }

  var approvedBy *string
  if userID, ok := auth.UserIDFromContext(r.Context()); ok {
    approvedBy = &userID
  }

  item, err := h.svc.Approve(r.Context(), id, approvedBy)
  if err != nil {
    if IsNotFound(err) {
      httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "expense not found", nil)
      return
    }
    if errors.Is(err, ErrExpenseApproved) {
      httpx.WriteError(w, http.StatusConflict, "ALREADY_APPROVED", "expense already approved", nil)
      return
    }
    httpx.WriteError(w, http.StatusInternalServerError, "EXPENSE_APPROVE_ERROR", "could not approve expense", err.Error())
    return
  }
  httpx.WriteJSON(w, http.StatusOK, item)
}

func parseListFilter(r *http.Request) (ListFilter, error) {
  filter := ListFilter{}
  q := r.URL.Query()

  if v := q.Get("trip_id"); v != "" {
    if _, err := uuid.Parse(v); err != nil {
      return filter, err
    }
    filter.TripID = v
  }
  if v := q.Get("driver_id"); v != "" {
    if _, err := uuid.Parse(v); err != nil {
      return filter, err
    }
    filter.DriverID = v
  }
  if v := q.Get("expense_type"); v != "" {
    filter.ExpenseType = v
  }
  if v := q.Get("payment_method"); v != "" {
    filter.PaymentMethod = v
  }
  if v := q.Get("approved"); v != "" {
    parsed, err := strconv.ParseBool(v)
    if err != nil {
      return filter, err
    }
    filter.Approved = &parsed
  }
  if v := q.Get("limit"); v != "" {
    n, err := strconv.Atoi(v)
    if err != nil || n <= 0 {
      return filter, errors.New("invalid limit")
    }
    filter.Limit = n
  }
  if v := q.Get("offset"); v != "" {
    n, err := strconv.Atoi(v)
    if err != nil || n < 0 {
      return filter, errors.New("invalid offset")
    }
    filter.Offset = n
  }

  return filter, nil
}

func isValidExpenseType(expenseType string) bool {
  switch expenseType {
  case "FUEL", "FOOD", "LODGING", "TOLL", "MAINTENANCE", "OTHER":
    return true
  default:
    return false
  }
}
