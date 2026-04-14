package pricing

import (
  "encoding/json"
  "time"
)

type PricingRule struct {
  ID        string          `json:"id"`
  Name      string          `json:"name"`
  Scope     string          `json:"scope"`
  ScopeID   *string         `json:"scope_id,omitempty"`
  RuleType  string          `json:"rule_type"`
  Priority  int             `json:"priority"`
  IsActive  bool            `json:"is_active"`
  Params    json.RawMessage `json:"params"`
  CreatedAt time.Time       `json:"created_at"`
  UpdatedAt time.Time       `json:"updated_at"`
}

type CreatePricingRuleInput struct {
  Name     string          `json:"name"`
  Scope    string          `json:"scope"`
  ScopeID  *string         `json:"scope_id"`
  RuleType string          `json:"rule_type"`
  Priority *int            `json:"priority"`
  IsActive *bool           `json:"is_active"`
  Params   json.RawMessage `json:"params"`
}

type UpdatePricingRuleInput struct {
  Name     *string         `json:"name"`
  Scope    *string         `json:"scope"`
  ScopeID  *string         `json:"scope_id"`
  RuleType *string         `json:"rule_type"`
  Priority *int            `json:"priority"`
  IsActive *bool           `json:"is_active"`
  Params   json.RawMessage `json:"params"`
}

type ListFilter struct {
  Limit    int
  Offset   int
  Scope    *string
  ScopeID  *string
  RuleType *string
  IsActive *bool
}

type QuoteInput struct {
  TripID         string   `json:"trip_id"`
  BoardStopID    string   `json:"board_stop_id"`
  AlightStopID   string   `json:"alight_stop_id"`
  FareMode       string   `json:"fare_mode"`
  FareAmountFinal *float64 `json:"fare_amount_final"`
}

type AppliedRule struct {
  ID         string  `json:"id"`
  Name       string  `json:"name"`
  RuleType   string  `json:"rule_type"`
  Multiplier float64 `json:"multiplier"`
  Priority   int     `json:"priority"`
}

type QuoteResult struct {
  TripID          string        `json:"trip_id"`
  RouteID         string        `json:"route_id"`
  BoardStopID     string        `json:"board_stop_id"`
  AlightStopID    string        `json:"alight_stop_id"`
  OriginStopID    string        `json:"origin_stop_id"`
  DestinationStopID string      `json:"destination_stop_id"`
  BoardStopOrder  int           `json:"board_stop_order"`
  AlightStopOrder int           `json:"alight_stop_order"`
  BaseAmount      float64       `json:"base_amount"`
  CalcAmount      float64       `json:"calc_amount"`
  FinalAmount     float64       `json:"final_amount"`
  Currency        string        `json:"currency"`
  FareMode        string        `json:"fare_mode"`
  OccupancyRatio  float64       `json:"occupancy_ratio"`
  AppliedRules    []AppliedRule `json:"applied_rules"`
  Snapshot        json.RawMessage `json:"snapshot"`
}
