package payments

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"math"
	"strings"

	"github.com/google/uuid"
	"schumacher-tur/api/internal/shared/config"
)

type Service struct {
	repo          *Repository
	client        *Client
	returnURL     string
	completionURL string
}

var (
	errMissingCheckoutURLs = errors.New("ABACATEPAY_RETURN_URL and ABACATEPAY_COMPLETION_URL are required for automatic payments")
)

func NewService(repo *Repository, cfg config.Config) *Service {
	client := NewClient(cfg.AbacatePayBaseURL, cfg.AbacatePayAPIKey)
	return &Service{
		repo:          repo,
		client:        client,
		returnURL:     cfg.AbacatePayReturnURL,
		completionURL: cfg.AbacatePayCompletionURL,
	}
}

func (s *Service) Create(ctx context.Context, input CreatePaymentInput) (Payment, json.RawMessage, error) {
	if input.Method != "PIX" && input.Method != "CARD" {
		return Payment{}, nil, errors.New("invalid method")
	}
	if strings.TrimSpace(s.returnURL) == "" || strings.TrimSpace(s.completionURL) == "" {
		return Payment{}, nil, errMissingCheckoutURLs
	}

	desc := input.Description
	if desc == "" {
		desc = "Passagem"
	}

	paymentID := uuid.NewString()
	cents := int64(math.Round(input.Amount * 100))
	billingReq := BillingRequest{
		Frequency: "ONE_TIME",
		Methods:   []string{input.Method},
		Products: []BillingProduct{{
			ExternalID: paymentID,
			Name:       desc,
			Quantity:   1,
			Price:      cents,
		}},
		ReturnURL:     s.returnURL,
		CompletionURL: s.completionURL,
		Customer:      nil,
	}

	if input.Customer != nil {
		billingReq.Customer = &BillingCustomer{
			Name:      input.Customer.Name,
			Email:     input.Customer.Email,
			Cellphone: input.Customer.Phone,
			TaxID:     input.Customer.Document,
		}
	}

	billing, raw, err := s.client.CreateBilling(ctx, billingReq)
	if err != nil {
		return Payment{}, nil, err
	}

	meta, _ := json.Marshal(map[string]interface{}{"billing": billing})
	created, err := s.repo.CreateWithProvider(ctx, paymentID, input.BookingID, input.Amount, input.Method, "ABACATEPAY", billing.ID, meta)
	if err != nil {
		return Payment{}, raw, err
	}
	return created, raw, nil
}

func (s *Service) CreateManual(ctx context.Context, input ManualPaymentInput) (Payment, error) {
	return s.repo.CreateManual(ctx, input.BookingID, input.Amount, input.Method, input.Notes)
}

func (s *Service) GetStatus(ctx context.Context, id string) (PaymentStatusResponse, error) {
	payment, err := s.repo.Get(ctx, id)
	if err != nil {
		return PaymentStatusResponse{}, err
	}
	return PaymentStatusResponse{
		ID:          payment.ID,
		Status:      payment.Status,
		Amount:      payment.Amount,
		Provider:    payment.Provider,
		ProviderRef: payment.ProviderRef,
		Metadata:    payment.Metadata,
	}, nil
}

func (s *Service) List(ctx context.Context, filter PaymentListFilter) ([]Payment, error) {
	return s.repo.List(ctx, filter)
}

func (s *Service) HandleWebhook(ctx context.Context, event AbacateWebhookEvent) error {
	if event.Event != "billing.paid" || event.BillingID == "" {
		return nil
	}
	_, err := s.repo.MarkPaidAndConfirmBooking(ctx, event.BillingID, event.Raw)
	if err != nil {
		if IsNotFound(err) {
			_ = s.repo.AddEvent(ctx, nil, event.Event, event.Raw)
			return nil
		}
		return err
	}
	return nil
}

func (s *Service) Sync(ctx context.Context, paymentID string) (PaymentSyncResponse, error) {
	payment, err := s.repo.Get(ctx, paymentID)
	if err != nil {
		return PaymentSyncResponse{}, err
	}
	bookingStatus, err := s.repo.GetBookingStatus(ctx, payment.BookingID)
	if err != nil {
		return PaymentSyncResponse{}, err
	}
	if payment.Provider == nil || strings.TrimSpace(*payment.Provider) != "ABACATEPAY" {
		return PaymentSyncResponse{Payment: payment, BookingStatus: bookingStatus, Synced: false}, nil
	}
	if payment.ProviderRef == nil || strings.TrimSpace(*payment.ProviderRef) == "" {
		return PaymentSyncResponse{Payment: payment, BookingStatus: bookingStatus, Synced: false}, nil
	}

	billing, raw, err := s.client.GetBillingByID(ctx, *payment.ProviderRef)
	if err != nil {
		return PaymentSyncResponse{}, err
	}
	if isPaidBillingStatus(billing.Status) && payment.Status != "PAID" {
		updated, err := s.repo.MarkPaidAndConfirmBooking(ctx, *payment.ProviderRef, raw)
		if err != nil {
			return PaymentSyncResponse{}, err
		}
		payment = updated
	}

	bookingStatus, err = s.repo.GetBookingStatus(ctx, payment.BookingID)
	if err != nil {
		return PaymentSyncResponse{}, err
	}
	providerRef := ""
	if payment.ProviderRef != nil {
		providerRef = strings.TrimSpace(*payment.ProviderRef)
	}
	log.Printf("event=payment_synced payment_id=%s booking_id=%s payment_status=%s booking_status=%s provider_ref=%s", payment.ID, payment.BookingID, payment.Status, bookingStatus, providerRef)
	return PaymentSyncResponse{Payment: payment, BookingStatus: bookingStatus, Synced: true}, nil
}

func parseProviderData(raw []byte) (interface{}, *string, *string) {
	if len(raw) == 0 {
		return nil, nil, nil
	}
	var parsed interface{}
	_ = json.Unmarshal(raw, &parsed)

	obj, ok := parsed.(map[string]interface{})
	if !ok {
		return parsed, nil, nil
	}

	checkout := readStringFromMap(obj, "url")
	if checkout == nil {
		if dataObj, ok := obj["data"].(map[string]interface{}); ok {
			checkout = readStringFromMap(dataObj, "url")
		}
	}

	var pixCode *string
	if dataObj, ok := obj["data"].(map[string]interface{}); ok {
		pixCode = readStringFromMap(dataObj, "pixQrCode")
	}
	if pixCode == nil {
		pixCode = readStringFromMap(obj, "pixQrCode")
	}
	return parsed, checkout, pixCode
}

func readStringFromMap(obj map[string]interface{}, key string) *string {
	raw, ok := obj[key]
	if !ok {
		return nil
	}
	value, ok := raw.(string)
	if !ok {
		return nil
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}

func isPaidBillingStatus(status string) bool {
	switch strings.ToUpper(strings.TrimSpace(status)) {
	case "PAID", "COMPLETED", "RECEIVED":
		return true
	default:
		return false
	}
}
