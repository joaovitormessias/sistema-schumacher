package chat

import (
	"context"
	"strings"

	"schumacher-tur/api/internal/pricing"
)

type PricingQuoteTool struct {
	svc interface {
		Quote(ctx context.Context, input pricing.QuoteInput) (pricing.QuoteResult, error)
	}
}

func NewPricingQuoteTool(svc interface {
	Quote(ctx context.Context, input pricing.QuoteInput) (pricing.QuoteResult, error)
}) *PricingQuoteTool {
	return &PricingQuoteTool{svc: svc}
}

func (t *PricingQuoteTool) Enabled() bool {
	return t != nil && t.svc != nil
}

func (t *PricingQuoteTool) Search(ctx context.Context, input PricingQuoteInput) (PricingQuoteResult, error) {
	if !t.Enabled() {
		return PricingQuoteResult{}, ErrAgentToolNotConfigured
	}

	result := PricingQuoteResult{
		Filter:  input,
		Results: make([]PricingQuoteItem, 0, len(input.Candidates)),
	}
	for _, candidate := range input.Candidates {
		quote, err := t.svc.Quote(ctx, pricing.QuoteInput{
			TripID:       candidate.TripID,
			BoardStopID:  candidate.BoardStopID,
			AlightStopID: candidate.AlightStopID,
			FareMode:     strings.TrimSpace(input.FareMode),
		})
		if err != nil {
			return PricingQuoteResult{}, err
		}
		result.Results = append(result.Results, PricingQuoteItem{
			TripID:                 quote.TripID,
			RouteID:                quote.RouteID,
			BoardStopID:            quote.BoardStopID,
			AlightStopID:           quote.AlightStopID,
			OriginStopID:           quote.OriginStopID,
			DestinationStopID:      quote.DestinationStopID,
			OriginDisplayName:      candidate.OriginDisplayName,
			DestinationDisplayName: candidate.DestinationDisplayName,
			OriginDepartTime:       candidate.OriginDepartTime,
			TripDate:               candidate.TripDate,
			BaseAmount:             quote.BaseAmount,
			CalcAmount:             quote.CalcAmount,
			FinalAmount:            quote.FinalAmount,
			Currency:               quote.Currency,
			FareMode:               quote.FareMode,
		})
	}

	return result, nil
}
