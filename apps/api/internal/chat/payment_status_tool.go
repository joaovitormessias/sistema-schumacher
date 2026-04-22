package chat

import (
	"context"

	"schumacher-tur/api/internal/payments"
)

type PaymentStatusTool struct {
	svc interface {
		List(ctx context.Context, filter payments.PaymentListFilter) ([]payments.Payment, error)
	}
}

func NewPaymentStatusTool(svc interface {
	List(ctx context.Context, filter payments.PaymentListFilter) ([]payments.Payment, error)
}) *PaymentStatusTool {
	return &PaymentStatusTool{svc: svc}
}

func (t *PaymentStatusTool) Enabled() bool {
	return t != nil && t.svc != nil
}

func (t *PaymentStatusTool) Search(ctx context.Context, input PaymentStatusInput) (PaymentStatusResult, error) {
	if !t.Enabled() {
		return PaymentStatusResult{}, ErrAgentToolNotConfigured
	}

	filter := payments.PaymentListFilter{
		BookingID: input.BookingID,
		Limit:     input.Limit,
	}
	items, err := t.svc.List(ctx, filter)
	if err != nil {
		return PaymentStatusResult{}, err
	}

	result := PaymentStatusResult{
		Filter:  input,
		Results: make([]PaymentStatusItem, 0, len(items)),
	}
	for _, item := range items {
		provider := ""
		if item.Provider != nil {
			provider = *item.Provider
		}
		providerRef := ""
		if item.ProviderRef != nil {
			providerRef = *item.ProviderRef
		}
		result.Results = append(result.Results, PaymentStatusItem{
			ID:          item.ID,
			BookingID:   item.BookingID,
			Amount:      item.Amount,
			Method:      item.Method,
			Status:      item.Status,
			Provider:    provider,
			ProviderRef: providerRef,
			PaidAt:      item.PaidAt,
			CreatedAt:   item.CreatedAt,
		})
	}

	return result, nil
}
