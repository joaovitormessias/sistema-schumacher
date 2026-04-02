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
	StopID    string
	StopOrder int
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

	fare, err := r.getSegmentFare(ctx, trip.RouteID, boardStop.StopID, alightStop.StopID)
	if err != nil {
		if mode == "MANUAL" && input.FareAmountFinal != nil {
			fare = segmentFare{BaseAmount: roundMoney(*input.FareAmountFinal), Currency: "BRL"}
		} else {
			return result, err
		}
	}

	calcAmount := fare.BaseAmount
	finalAmount := calcAmount
	if mode == "MANUAL" {
		finalAmount = roundMoney(*input.FareAmountFinal)
	}

	snapshot, _ := json.Marshal(map[string]interface{}{
		"base_amount":    fare.BaseAmount,
		"currency":       fare.Currency,
		"calc_amount":    calcAmount,
		"final_amount":   finalAmount,
		"fare_mode":      mode,
		"applied_rules":  []AppliedRule{},
		"pricing_source": "route_segment_prices",
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
		OccupancyRatio:  0,
		AppliedRules:    []AppliedRule{},
		Snapshot:        snapshot,
	}

	return result, nil
}

func (r *Repository) getTripInfo(ctx context.Context, tripID string) (tripInfo, error) {
	var info tripInfo
	row := r.pool.QueryRow(ctx, `
    select route_id, (trip_date::timestamp), coalesce(seats_total, 0)
    from trips
    where trip_id = $1
  `, tripID)
	if err := row.Scan(&info.RouteID, &info.DepartureAt, &info.Capacity); err != nil {
		return info, ErrTripNotFound
	}
	return info, nil
}

func (r *Repository) getTripStopInfo(ctx context.Context, tripID string, stopID string) (stopInfo, error) {
	var info stopInfo
	row := r.pool.QueryRow(ctx, `
    select stop_id, stop_sequence
    from trip_stops
    where trip_id = $1 and trip_stop_id = $2
  `, tripID, stopID)
	if err := row.Scan(&info.StopID, &info.StopOrder); err != nil {
		return info, ErrStopNotFound
	}
	return info, nil
}

func (r *Repository) getSegmentFare(ctx context.Context, routeID string, fromStopID string, toStopID string) (segmentFare, error) {
	var fare segmentFare
	row := r.pool.QueryRow(ctx, `
    select price
    from route_segment_prices
    where route_id = $1
      and origin_stop_id = $2
      and destination_stop_id = $3
      and upper(coalesce(status, 'ACTIVE')) = 'ACTIVE'
    limit 1
  `, routeID, fromStopID, toStopID)
	if err := row.Scan(&fare.BaseAmount); err != nil {
		return fare, ErrSegmentFareNotFound
	}
	fare.Currency = "BRL"
	return fare, nil
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
