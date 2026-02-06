package pricing

import (
  "context"
  "encoding/json"
  "errors"
  "math"
  "strings"
  "time"
)

var (
  ErrTripNotFound        = errors.New("trip not found")
  ErrStopNotFound        = errors.New("stop not found")
  ErrInvalidStopOrder    = errors.New("invalid stop order")
  ErrSegmentFareNotFound = errors.New("segment fare not found")
  ErrInvalidFareMode     = errors.New("invalid fare mode")
  ErrFareAmountRequired  = errors.New("fare_amount_final is required for MANUAL mode")
)

type tripInfo struct {
  RouteID     string
  DepartureAt time.Time
  Capacity    int
}

type stopInfo struct {
  RouteStopID string
  StopOrder   int
}

type segmentFare struct {
  BaseAmount float64
  Currency   string
}

func (r *Repository) Quote(ctx context.Context, input QuoteInput) (QuoteResult, error) {
  result := QuoteResult{}

  mode := normalizeFareMode(input.FareMode)
  if mode == "" {
    return result, ErrInvalidFareMode
  }
  if mode == "MANUAL" && input.FareAmountFinal == nil {
    return result, ErrFareAmountRequired
  }

  trip, err := r.getTripInfo(ctx, input.TripID)
  if err != nil {
    return result, err
  }

  boardStop, err := r.getTripStopInfo(ctx, input.TripID, input.BoardStopID)
  if err != nil {
    return result, err
  }
  alightStop, err := r.getTripStopInfo(ctx, input.TripID, input.AlightStopID)
  if err != nil {
    return result, err
  }
  if boardStop.StopOrder >= alightStop.StopOrder {
    return result, ErrInvalidStopOrder
  }

  missingSegmentFare := false
  fare, err := r.getSegmentFare(ctx, input.TripID, trip.RouteID, boardStop.RouteStopID, alightStop.RouteStopID)
  if err != nil {
    if errors.Is(err, ErrSegmentFareNotFound) && mode == "MANUAL" && input.FareAmountFinal != nil {
      missingSegmentFare = true
      fare = segmentFare{
        BaseAmount: roundMoney(*input.FareAmountFinal),
        Currency:   "BRL",
      }
    } else {
      return result, err
    }
  }

  calcAmount := fare.BaseAmount
  appliedRules := []AppliedRule{}
  occupancyRatio := 0.0

  if mode != "FIXED" && !missingSegmentFare {
    rules, err := r.getApplicableRules(ctx, trip.RouteID, input.TripID)
    if err != nil {
      return result, err
    }

    seenTypes := map[string]bool{}
    for _, rule := range rules {
      if seenTypes[rule.RuleType] {
        continue
      }
      multiplier, ratio, ok := r.evaluateRule(ctx, rule, input.TripID, trip, boardStop, alightStop)
      if !ok {
        continue
      }
      calcAmount = roundMoney(calcAmount * multiplier)
      appliedRules = append(appliedRules, AppliedRule{
        ID:         rule.ID,
        Name:       rule.Name,
        RuleType:   rule.RuleType,
        Multiplier: multiplier,
        Priority:   rule.Priority,
      })
      if rule.RuleType == "OCCUPANCY" {
        occupancyRatio = ratio
      }
      seenTypes[rule.RuleType] = true
    }
  }

  finalAmount := calcAmount
  if mode == "MANUAL" {
    finalAmount = roundMoney(*input.FareAmountFinal)
  }

  snapshot, _ := json.Marshal(map[string]interface{}{
    "base_amount":     fare.BaseAmount,
    "currency":        fare.Currency,
    "calc_amount":     calcAmount,
    "final_amount":    finalAmount,
    "fare_mode":       mode,
    "occupancy_ratio": occupancyRatio,
    "applied_rules":   appliedRules,
  })

  result = QuoteResult{
    TripID:          input.TripID,
    RouteID:         trip.RouteID,
    BoardStopID:     input.BoardStopID,
    AlightStopID:    input.AlightStopID,
    BoardStopOrder:  boardStop.StopOrder,
    AlightStopOrder: alightStop.StopOrder,
    BaseAmount:      fare.BaseAmount,
    CalcAmount:      calcAmount,
    FinalAmount:     finalAmount,
    Currency:        fare.Currency,
    FareMode:        mode,
    OccupancyRatio:  occupancyRatio,
    AppliedRules:    appliedRules,
    Snapshot:        snapshot,
  }

  return result, nil
}

func (r *Repository) getTripInfo(ctx context.Context, tripID string) (tripInfo, error) {
  var info tripInfo
  row := r.pool.QueryRow(ctx, `
    select t.route_id, t.departure_at, b.capacity
    from trips t
    join buses b on b.id = t.bus_id
    where t.id = $1
  `, tripID)
  if err := row.Scan(&info.RouteID, &info.DepartureAt, &info.Capacity); err != nil {
    return info, ErrTripNotFound
  }
  return info, nil
}

func (r *Repository) getTripStopInfo(ctx context.Context, tripID string, stopID string) (stopInfo, error) {
  var info stopInfo
  row := r.pool.QueryRow(ctx, `
    select route_stop_id, stop_order
    from trip_stops
    where trip_id = $1 and id = $2
  `, tripID, stopID)
  if err := row.Scan(&info.RouteStopID, &info.StopOrder); err != nil {
    return info, ErrStopNotFound
  }
  return info, nil
}

func (r *Repository) getSegmentFare(ctx context.Context, tripID string, routeID string, fromRouteStopID string, toRouteStopID string) (segmentFare, error) {
  var fare segmentFare
  row := r.pool.QueryRow(ctx, `
    select base_amount, currency
    from segment_fares
    where is_active = true
      and from_route_stop_id = $1
      and to_route_stop_id = $2
      and (
        (scope = 'TRIP' and scope_id = $3)
        or (scope = 'ROUTE' and scope_id = $4)
      )
      and (valid_from is null or valid_from <= now())
      and (valid_to is null or valid_to >= now())
    order by (scope = 'TRIP') desc, priority asc, created_at desc
    limit 1
  `, fromRouteStopID, toRouteStopID, tripID, routeID)
  if err := row.Scan(&fare.BaseAmount, &fare.Currency); err != nil {
    return fare, ErrSegmentFareNotFound
  }
  return fare, nil
}

type ruleRow struct {
  ID       string
  Name     string
  RuleType string
  Priority int
  Params   json.RawMessage
}

func (r *Repository) getApplicableRules(ctx context.Context, routeID string, tripID string) ([]ruleRow, error) {
  rows, err := r.pool.Query(ctx, `
    select id, name, rule_type, priority, params
    from pricing_rules
    where is_active = true
      and (
        scope = 'GLOBAL'
        or (scope = 'ROUTE' and scope_id = $1)
        or (scope = 'TRIP' and scope_id = $2)
      )
    order by priority asc, created_at desc
  `, routeID, tripID)
  if err != nil {
    return nil, err
  }
  defer rows.Close()

  items := []ruleRow{}
  for rows.Next() {
    var item ruleRow
    if err := rows.Scan(&item.ID, &item.Name, &item.RuleType, &item.Priority, &item.Params); err != nil {
      return nil, err
    }
    items = append(items, item)
  }
  return items, rows.Err()
}

func (r *Repository) evaluateRule(ctx context.Context, rule ruleRow, tripID string, trip tripInfo, board stopInfo, alight stopInfo) (float64, float64, bool) {
  switch rule.RuleType {
  case "OCCUPANCY":
    ratio, ok := r.getOccupancyRatio(ctx, tripID, trip, board, alight)
    if !ok {
      return 1, 0, false
    }
    mult := occupancyMultiplier(rule.Params, ratio)
    return mult, ratio, true
  case "LEAD_TIME":
    hours := time.Until(trip.DepartureAt).Hours()
    mult := leadTimeMultiplier(rule.Params, hours)
    return mult, 0, true
  case "DOW":
    dow := int(trip.DepartureAt.Weekday())
    mult := dowMultiplier(rule.Params, dow)
    return mult, 0, true
  case "SEASON":
    mult := seasonMultiplier(rule.Params, trip.DepartureAt)
    return mult, 0, true
  default:
    return 1, 0, false
  }
}

func (r *Repository) getOccupancyRatio(ctx context.Context, tripID string, trip tripInfo, board stopInfo, alight stopInfo) (float64, bool) {
  if trip.Capacity <= 0 {
    return 0, false
  }

  var maxTaken int
  row := r.pool.QueryRow(ctx, `
    with segments as (
      select generate_series($1, $2 - 1) as seg
    ),
    occupied as (
      select s.seg, count(distinct bp.seat_id) as taken
      from segments s
      join booking_passengers bp on bp.trip_id = $3
      join bookings b on b.id = bp.booking_id
      where bp.is_active = true
        and b.status not in ('CANCELLED', 'EXPIRED')
        and bp.board_stop_order < s.seg + 1
        and bp.alight_stop_order > s.seg
      group by s.seg
    )
    select coalesce(max(taken), 0) from occupied
  `, board.StopOrder, alight.StopOrder, tripID)
  if err := row.Scan(&maxTaken); err != nil {
    return 0, false
  }

  ratio := float64(maxTaken) / float64(trip.Capacity)
  return ratio, true
}

func occupancyMultiplier(params json.RawMessage, ratio float64) float64 {
  if len(params) == 0 {
    return 1
  }
  type band struct {
    Min        float64 `json:"min"`
    Max        float64 `json:"max"`
    Multiplier float64 `json:"multiplier"`
  }
  type cfg struct {
    Bands []band `json:"bands"`
  }

  var data cfg
  if err := json.Unmarshal(params, &data); err != nil {
    return 1
  }
  for _, b := range data.Bands {
    if ratio >= b.Min && ratio <= b.Max {
      if b.Multiplier > 0 {
        return b.Multiplier
      }
      return 1
    }
  }
  return 1
}

func leadTimeMultiplier(params json.RawMessage, hours float64) float64 {
  if len(params) == 0 {
    return 1
  }
  type band struct {
    MinHours   float64 `json:"min_hours"`
    MaxHours   float64 `json:"max_hours"`
    Multiplier float64 `json:"multiplier"`
  }
  type cfg struct {
    Bands []band `json:"bands"`
  }

  var data cfg
  if err := json.Unmarshal(params, &data); err != nil {
    return 1
  }
  for _, b := range data.Bands {
    if hours >= b.MinHours && hours <= b.MaxHours {
      if b.Multiplier > 0 {
        return b.Multiplier
      }
      return 1
    }
  }
  return 1
}

func dowMultiplier(params json.RawMessage, dow int) float64 {
  if len(params) == 0 {
    return 1
  }
  type band struct {
    Days       []int   `json:"days"`
    Multiplier float64 `json:"multiplier"`
  }
  type cfg struct {
    Days []band `json:"days"`
  }

  var data cfg
  if err := json.Unmarshal(params, &data); err != nil {
    return 1
  }
  for _, b := range data.Days {
    for _, d := range b.Days {
      if d == dow {
        if b.Multiplier > 0 {
          return b.Multiplier
        }
        return 1
      }
    }
  }
  return 1
}

func seasonMultiplier(params json.RawMessage, departure time.Time) float64 {
  if len(params) == 0 {
    return 1
  }
  type window struct {
    From       string  `json:"from"`
    To         string  `json:"to"`
    Multiplier float64 `json:"multiplier"`
  }
  type cfg struct {
    Windows []window `json:"windows"`
  }

  var data cfg
  if err := json.Unmarshal(params, &data); err != nil {
    return 1
  }
  for _, w := range data.Windows {
    from, err1 := time.Parse("2006-01-02", w.From)
    to, err2 := time.Parse("2006-01-02", w.To)
    if err1 != nil || err2 != nil {
      continue
    }
    if !departure.Before(from) && !departure.After(to.Add(24*time.Hour-time.Nanosecond)) {
      if w.Multiplier > 0 {
        return w.Multiplier
      }
      return 1
    }
  }
  return 1
}

func normalizeFareMode(mode string) string {
  switch strings.ToUpper(strings.TrimSpace(mode)) {
  case "", "AUTO":
    return "AUTO"
  case "FIXED":
    return "FIXED"
  case "MANUAL":
    return "MANUAL"
  default:
    return ""
  }
}

func roundMoney(val float64) float64 {
  return math.Round(val*100) / 100
}
