package payments

import (
  "context"
  "encoding/json"
  "errors"
  "math"

  "github.com/google/uuid"

  "schumacher-tur/api/internal/shared/config"
)

type Service struct {
  repo   *Repository
  client *Client
}

func NewService(repo *Repository, cfg config.Config) *Service {
  client := NewClient(cfg.AbacatePayBaseURL, cfg.AbacatePayAPIKey)
  return &Service{repo: repo, client: client}
}

func (s *Service) Create(ctx context.Context, input CreatePaymentInput) (Payment, json.RawMessage, error) {
  if input.Method != "PIX" && input.Method != "CARD" {
    return Payment{}, nil, errors.New("invalid method")
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
    ReturnURL:     "",
    CompletionURL: "",
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

  meta, _ := json.Marshal(map[string]interface{}{
    "billing": billing,
  })

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
  if event.Event != "billing.paid" {
    return nil
  }
  if event.Data == nil || event.Data.Billing == nil {
    return nil
  }

  billingID := event.Data.Billing.ID
  _, err := s.repo.MarkPaidAndConfirmBooking(ctx, billingID, event.Raw)
  if err != nil {
    if IsNotFound(err) {
      _ = s.repo.AddEvent(ctx, nil, event.Event, event.Raw)
      return nil
    }
    return err
  }
  return nil
}
