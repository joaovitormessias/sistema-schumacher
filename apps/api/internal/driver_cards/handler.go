package driver_cards

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
  r.Route("/driver-cards", func(r chi.Router) {
    r.Get("/", h.list)
    r.Post("/", h.create)
    r.Get("/{cardId}", h.get)
    r.Patch("/{cardId}", h.update)
    r.Post("/{cardId}/block", h.block)
    r.Post("/{cardId}/unblock", h.unblock)
    r.Get("/{cardId}/transactions", h.listTransactions)
    r.Post("/{cardId}/transactions", h.createTransaction)
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
    httpx.WriteError(w, http.StatusInternalServerError, "CARDS_LIST_ERROR", "could not list cards", err.Error())
    return
  }
  httpx.WriteJSON(w, http.StatusOK, items)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
  id, err := httpx.ParseUUIDParam(r, "cardId")
  if err != nil {
    httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid card id", nil)
    return
  }
  item, err := h.svc.Get(r.Context(), id)
  if err != nil {
    if IsNotFound(err) {
      httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "card not found", nil)
      return
    }
    httpx.WriteError(w, http.StatusInternalServerError, "CARD_GET_ERROR", "could not get card", err.Error())
    return
  }
  httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
  var input CreateDriverCardInput
  if err := httpx.DecodeJSON(r, &input); err != nil {
    httpx.WriteError(w, http.StatusBadRequest, "INVALID_BODY", "invalid json", err.Error())
    return
  }
  if input.DriverID == "" || input.CardNumber == "" || input.CardType == "" {
    httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "driver_id, card_number and card_type are required", nil)
    return
  }
  if _, err := uuid.Parse(input.DriverID); err != nil {
    httpx.WriteError(w, http.StatusBadRequest, "INVALID_DRIVER", "invalid driver_id", nil)
    return
  }
  if !isValidCardType(input.CardType) {
    httpx.WriteError(w, http.StatusBadRequest, "INVALID_CARD_TYPE", "invalid card_type", nil)
    return
  }
  if input.CurrentBalance != nil && *input.CurrentBalance < 0 {
    httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "current_balance must be positive", nil)
    return
  }

  item, err := h.svc.Create(r.Context(), input)
  if err != nil {
    httpx.WriteError(w, http.StatusInternalServerError, "CARD_CREATE_ERROR", "could not create card", err.Error())
    return
  }
  httpx.WriteJSON(w, http.StatusCreated, item)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
  id, err := httpx.ParseUUIDParam(r, "cardId")
  if err != nil {
    httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid card id", nil)
    return
  }

  var input UpdateDriverCardInput
  if err := httpx.DecodeJSON(r, &input); err != nil {
    httpx.WriteError(w, http.StatusBadRequest, "INVALID_BODY", "invalid json", err.Error())
    return
  }
  if input.CardType != nil && !isValidCardType(*input.CardType) {
    httpx.WriteError(w, http.StatusBadRequest, "INVALID_CARD_TYPE", "invalid card_type", nil)
    return
  }

  item, err := h.svc.Update(r.Context(), id, input)
  if err != nil {
    if IsNotFound(err) {
      httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "card not found", nil)
      return
    }
    httpx.WriteError(w, http.StatusInternalServerError, "CARD_UPDATE_ERROR", "could not update card", err.Error())
    return
  }
  httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) block(w http.ResponseWriter, r *http.Request) {
  id, err := httpx.ParseUUIDParam(r, "cardId")
  if err != nil {
    httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid card id", nil)
    return
  }

  var input BlockCardInput
  if err := httpx.DecodeJSON(r, &input); err != nil && !errors.Is(err, httpx.ErrEmptyBody) {
    httpx.WriteError(w, http.StatusBadRequest, "INVALID_BODY", "invalid json", err.Error())
    return
  }

  var blockedBy *string
  if userID, ok := auth.UserIDFromContext(r.Context()); ok {
    blockedBy = &userID
  }

  item, err := h.svc.Block(r.Context(), id, input.Reason, blockedBy)
  if err != nil {
    if IsNotFound(err) {
      httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "card not found", nil)
      return
    }
    if errors.Is(err, ErrCardStatus) {
      httpx.WriteError(w, http.StatusBadRequest, "INVALID_STATUS", "card already blocked", nil)
      return
    }
    httpx.WriteError(w, http.StatusInternalServerError, "CARD_BLOCK_ERROR", "could not block card", err.Error())
    return
  }
  httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) unblock(w http.ResponseWriter, r *http.Request) {
  id, err := httpx.ParseUUIDParam(r, "cardId")
  if err != nil {
    httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid card id", nil)
    return
  }

  item, err := h.svc.Unblock(r.Context(), id)
  if err != nil {
    if IsNotFound(err) {
      httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "card not found", nil)
      return
    }
    if errors.Is(err, ErrCardStatus) {
      httpx.WriteError(w, http.StatusBadRequest, "INVALID_STATUS", "card is not blocked", nil)
      return
    }
    httpx.WriteError(w, http.StatusInternalServerError, "CARD_UNBLOCK_ERROR", "could not unblock card", err.Error())
    return
  }
  httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) listTransactions(w http.ResponseWriter, r *http.Request) {
  cardID, err := httpx.ParseUUIDParam(r, "cardId")
  if err != nil {
    httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid card id", nil)
    return
  }

  filter, err := parseTransactionFilter(r)
  if err != nil {
    httpx.WriteError(w, http.StatusBadRequest, "INVALID_QUERY", "invalid query parameters", nil)
    return
  }
  filter.CardID = cardID.String()

  items, err := h.svc.ListTransactions(r.Context(), filter)
  if err != nil {
    httpx.WriteError(w, http.StatusInternalServerError, "TRANSACTIONS_LIST_ERROR", "could not list transactions", err.Error())
    return
  }
  httpx.WriteJSON(w, http.StatusOK, items)
}

func (h *Handler) createTransaction(w http.ResponseWriter, r *http.Request) {
  cardID, err := httpx.ParseUUIDParam(r, "cardId")
  if err != nil {
    httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid card id", nil)
    return
  }

  var input CreateCardTransactionInput
  if err := httpx.DecodeJSON(r, &input); err != nil {
    httpx.WriteError(w, http.StatusBadRequest, "INVALID_BODY", "invalid json", err.Error())
    return
  }
  if input.TransactionType == "" {
    httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "transaction_type is required", nil)
    return
  }

  var performedBy *string
  if userID, ok := auth.UserIDFromContext(r.Context()); ok {
    performedBy = &userID
  }

  item, err := h.svc.CreateTransaction(r.Context(), cardID, input, nil, performedBy)
  if err != nil {
    if IsNotFound(err) {
      httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "card not found", nil)
      return
    }
    if errors.Is(err, ErrCardBlocked) {
      httpx.WriteError(w, http.StatusConflict, "CARD_BLOCKED", "card is blocked", nil)
      return
    }
    if errors.Is(err, ErrCardInactive) {
      httpx.WriteError(w, http.StatusConflict, "CARD_INACTIVE", "card is inactive", nil)
      return
    }
    if errors.Is(err, ErrInvalidAmount) {
      httpx.WriteError(w, http.StatusBadRequest, "INVALID_AMOUNT", "amount must be positive", nil)
      return
    }
    if errors.Is(err, ErrInvalidTransactionType) {
      httpx.WriteError(w, http.StatusBadRequest, "INVALID_TYPE", "invalid transaction_type", nil)
      return
    }
    if errors.Is(err, ErrInsufficientBalance) {
      httpx.WriteError(w, http.StatusConflict, "INSUFFICIENT_BALANCE", "insufficient card balance", nil)
      return
    }
    httpx.WriteError(w, http.StatusInternalServerError, "TRANSACTION_CREATE_ERROR", "could not create transaction", err.Error())
    return
  }
  httpx.WriteJSON(w, http.StatusCreated, item)
}

func parseListFilter(r *http.Request) (ListFilter, error) {
  filter := ListFilter{}
  q := r.URL.Query()

  if v := q.Get("driver_id"); v != "" {
    if _, err := uuid.Parse(v); err != nil {
      return filter, err
    }
    filter.DriverID = v
  }
  if v := q.Get("is_active"); v != "" {
    parsed, err := strconv.ParseBool(v)
    if err != nil {
      return filter, err
    }
    filter.IsActive = &parsed
  }
  if v := q.Get("is_blocked"); v != "" {
    parsed, err := strconv.ParseBool(v)
    if err != nil {
      return filter, err
    }
    filter.IsBlocked = &parsed
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

func parseTransactionFilter(r *http.Request) (TransactionListFilter, error) {
  filter := TransactionListFilter{}
  q := r.URL.Query()

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

func isValidCardType(cardType string) bool {
  switch cardType {
  case "FUEL", "MULTIPURPOSE", "FOOD":
    return true
  default:
    return false
  }
}
