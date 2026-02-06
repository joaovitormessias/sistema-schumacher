package pricing

import (
  "encoding/json"
  "errors"
  "net/http"
  "strconv"
  "strings"

  "github.com/go-chi/chi/v5"
  "github.com/google/uuid"
  httpx "schumacher-tur/api/internal/shared/http"
)

var allowedScopes = map[string]bool{
  "GLOBAL": true,
  "ROUTE":  true,
  "TRIP":   true,
}

var allowedRuleTypes = map[string]bool{
  "OCCUPANCY": true,
  "LEAD_TIME": true,
  "DOW":       true,
  "SEASON":    true,
}

type Handler struct {
  svc *Service
}

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) RegisterRoutes(r chi.Router) {
  r.Post("/pricing/quote", h.quote)
  r.Route("/pricing/rules", func(r chi.Router) {
    r.Get("/", h.list)
    r.Post("/", h.create)
    r.Patch("/{ruleId}", h.update)
    r.Get("/{ruleId}", h.get)
  })
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
  filter, err := parseListFilter(r)
  if err != nil {
    httpx.WriteError(w, http.StatusBadRequest, "INVALID_PAGINATION", "invalid query parameters", err.Error())
    return
  }
  items, err := h.svc.List(r.Context(), filter)
  if err != nil {
    httpx.WriteError(w, http.StatusInternalServerError, "PRICING_RULES_LIST_ERROR", "could not list pricing rules", err.Error())
    return
  }
  httpx.WriteJSON(w, http.StatusOK, items)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
  id, err := httpx.ParseUUIDParam(r, "ruleId")
  if err != nil {
    httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid rule id", nil)
    return
  }
  item, err := h.svc.Get(r.Context(), id)
  if err != nil {
    if IsNotFound(err) {
      httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "pricing rule not found", nil)
      return
    }
    httpx.WriteError(w, http.StatusInternalServerError, "PRICING_RULE_GET_ERROR", "could not get pricing rule", err.Error())
    return
  }
  httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
  var input CreatePricingRuleInput
  if err := httpx.DecodeJSON(r, &input); err != nil {
    httpx.WriteError(w, http.StatusBadRequest, "INVALID_BODY", "invalid json", err.Error())
    return
  }

  normalized, err := normalizeCreateInput(&input)
  if err != nil {
    httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
    return
  }
  input = *normalized

  item, err := h.svc.Create(r.Context(), input)
  if err != nil {
    httpx.WriteError(w, http.StatusInternalServerError, "PRICING_RULE_CREATE_ERROR", "could not create pricing rule", err.Error())
    return
  }
  httpx.WriteJSON(w, http.StatusCreated, item)
}

func (h *Handler) quote(w http.ResponseWriter, r *http.Request) {
  var input QuoteInput
  if err := httpx.DecodeJSON(r, &input); err != nil {
    httpx.WriteError(w, http.StatusBadRequest, "INVALID_BODY", "invalid json", err.Error())
    return
  }
  if input.TripID == "" || input.BoardStopID == "" || input.AlightStopID == "" {
    httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "trip_id, board_stop_id and alight_stop_id are required", nil)
    return
  }

  result, err := h.svc.Quote(r.Context(), input)
  if err != nil {
    switch err {
    case ErrTripNotFound:
      httpx.WriteError(w, http.StatusNotFound, "TRIP_NOT_FOUND", "trip not found", nil)
    case ErrStopNotFound:
      httpx.WriteError(w, http.StatusNotFound, "STOP_NOT_FOUND", "stop not found", nil)
    case ErrInvalidStopOrder:
      httpx.WriteError(w, http.StatusBadRequest, "INVALID_STOPS", "board_stop must be before alight_stop", nil)
    case ErrSegmentFareNotFound:
      httpx.WriteError(w, http.StatusNotFound, "FARE_NOT_FOUND", "segment fare not found", nil)
    case ErrInvalidFareMode:
      httpx.WriteError(w, http.StatusBadRequest, "INVALID_FARE_MODE", "invalid fare_mode", nil)
    case ErrFareAmountRequired:
      httpx.WriteError(w, http.StatusBadRequest, "FARE_AMOUNT_REQUIRED", "fare_amount_final is required for MANUAL", nil)
    default:
      httpx.WriteError(w, http.StatusInternalServerError, "PRICING_QUOTE_ERROR", "could not calculate quote", err.Error())
    }
    return
  }

  httpx.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
  id, err := httpx.ParseUUIDParam(r, "ruleId")
  if err != nil {
    httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid rule id", nil)
    return
  }

  var input UpdatePricingRuleInput
  if err := httpx.DecodeJSON(r, &input); err != nil {
    httpx.WriteError(w, http.StatusBadRequest, "INVALID_BODY", "invalid json", err.Error())
    return
  }

  normalized, err := normalizeUpdateInput(&input)
  if err != nil {
    httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
    return
  }
  input = *normalized

  item, err := h.svc.Update(r.Context(), id, input)
  if err != nil {
    if IsNotFound(err) {
      httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "pricing rule not found", nil)
      return
    }
    httpx.WriteError(w, http.StatusInternalServerError, "PRICING_RULE_UPDATE_ERROR", "could not update pricing rule", err.Error())
    return
  }
  httpx.WriteJSON(w, http.StatusOK, item)
}

func parseListFilter(r *http.Request) (ListFilter, error) {
  filter := ListFilter{}
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
  if v := strings.TrimSpace(q.Get("scope")); v != "" {
    scope := strings.ToUpper(v)
    if !allowedScopes[scope] {
      return filter, errors.New("invalid scope")
    }
    filter.Scope = &scope
  }
  if v := strings.TrimSpace(q.Get("scope_id")); v != "" {
    if _, err := uuid.Parse(v); err != nil {
      return filter, errors.New("invalid scope_id")
    }
    filter.ScopeID = &v
  }
  if v := strings.TrimSpace(q.Get("rule_type")); v != "" {
    rt := strings.ToUpper(v)
    if !allowedRuleTypes[rt] {
      return filter, errors.New("invalid rule_type")
    }
    filter.RuleType = &rt
  }
  if v := strings.TrimSpace(q.Get("is_active")); v != "" {
    normalized := strings.ToLower(v)
    if normalized == "true" || normalized == "1" {
      val := true
      filter.IsActive = &val
    } else if normalized == "false" || normalized == "0" {
      val := false
      filter.IsActive = &val
    } else {
      return filter, errors.New("invalid is_active")
    }
  }
  return filter, nil
}

func normalizeCreateInput(input *CreatePricingRuleInput) (*CreatePricingRuleInput, error) {
  if strings.TrimSpace(input.Name) == "" {
    return nil, errors.New("name is required")
  }
  input.Name = strings.TrimSpace(input.Name)

  scope := strings.ToUpper(strings.TrimSpace(input.Scope))
  if scope == "" {
    scope = "GLOBAL"
  }
  if !allowedScopes[scope] {
    return nil, errors.New("invalid scope")
  }
  input.Scope = scope

  if scope == "GLOBAL" {
    if input.ScopeID != nil && strings.TrimSpace(*input.ScopeID) != "" {
      return nil, errors.New("scope_id must be empty for GLOBAL scope")
    }
    input.ScopeID = nil
  } else {
    if input.ScopeID == nil || strings.TrimSpace(*input.ScopeID) == "" {
      return nil, errors.New("scope_id is required for ROUTE or TRIP scope")
    }
    if _, err := uuid.Parse(strings.TrimSpace(*input.ScopeID)); err != nil {
      return nil, errors.New("invalid scope_id")
    }
    normalized := strings.TrimSpace(*input.ScopeID)
    input.ScopeID = &normalized
  }

  ruleType := strings.ToUpper(strings.TrimSpace(input.RuleType))
  if ruleType == "" {
    return nil, errors.New("rule_type is required")
  }
  if !allowedRuleTypes[ruleType] {
    return nil, errors.New("invalid rule_type")
  }
  input.RuleType = ruleType

  params, err := normalizeParams(input.Params)
  if err != nil {
    return nil, err
  }
  input.Params = params

  return input, nil
}

func normalizeUpdateInput(input *UpdatePricingRuleInput) (*UpdatePricingRuleInput, error) {
  if input.Name != nil {
    trimmed := strings.TrimSpace(*input.Name)
    if trimmed == "" {
      return nil, errors.New("name cannot be empty")
    }
    input.Name = &trimmed
  }

  if input.Scope != nil {
    scope := strings.ToUpper(strings.TrimSpace(*input.Scope))
    if scope == "" {
      return nil, errors.New("scope cannot be empty")
    }
    if !allowedScopes[scope] {
      return nil, errors.New("invalid scope")
    }
    input.Scope = &scope
    if scope == "GLOBAL" {
      if input.ScopeID != nil && strings.TrimSpace(*input.ScopeID) != "" {
        return nil, errors.New("scope_id must be empty for GLOBAL scope")
      }
      input.ScopeID = nil
    } else {
      if input.ScopeID == nil || strings.TrimSpace(*input.ScopeID) == "" {
        return nil, errors.New("scope_id is required when changing to ROUTE or TRIP scope")
      }
      if _, err := uuid.Parse(strings.TrimSpace(*input.ScopeID)); err != nil {
        return nil, errors.New("invalid scope_id")
      }
      normalized := strings.TrimSpace(*input.ScopeID)
      input.ScopeID = &normalized
    }
  } else if input.ScopeID != nil {
    return nil, errors.New("scope is required when updating scope_id")
  }

  if input.RuleType != nil {
    rt := strings.ToUpper(strings.TrimSpace(*input.RuleType))
    if rt == "" {
      return nil, errors.New("rule_type cannot be empty")
    }
    if !allowedRuleTypes[rt] {
      return nil, errors.New("invalid rule_type")
    }
    input.RuleType = &rt
  }

  if input.Params != nil {
    params, err := normalizeParams(input.Params)
    if err != nil {
      return nil, err
    }
    input.Params = params
  }

  return input, nil
}

func normalizeParams(raw json.RawMessage) (json.RawMessage, error) {
  if len(raw) == 0 {
    return json.RawMessage(`{}`), nil
  }
  var decoded interface{}
  if err := json.Unmarshal(raw, &decoded); err != nil {
    return nil, errors.New("params must be valid json")
  }
  if _, ok := decoded.(map[string]interface{}); !ok {
    return nil, errors.New("params must be a json object")
  }
  return raw, nil
}
