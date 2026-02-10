package payments

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	baseURL string
	apiKey  string
	http    *http.Client
}

func NewClient(baseURL, apiKey string) *Client {
	if baseURL == "" {
		baseURL = "https://api.abacatepay.com/v1"
	}
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  apiKey,
		http:    &http.Client{Timeout: 15 * time.Second},
	}
}

type BillingProduct struct {
	ExternalID string `json:"externalId"`
	Name       string `json:"name"`
	Quantity   int64  `json:"quantity"`
	Price      int64  `json:"price"`
}

type BillingCustomer struct {
	Name      string `json:"name"`
	Email     string `json:"email"`
	Cellphone string `json:"cellphone"`
	TaxID     string `json:"taxId"`
}

type BillingRequest struct {
	Frequency     string           `json:"frequency"`
	Methods       []string         `json:"methods"`
	Products      []BillingProduct `json:"products"`
	ReturnURL     string           `json:"returnUrl,omitempty"`
	CompletionURL string           `json:"completionUrl,omitempty"`
	Customer      *BillingCustomer `json:"customer,omitempty"`
}

type BillingResponse struct {
	ID       string           `json:"id"`
	URL      string           `json:"url"`
	Status   string           `json:"status"`
	Methods  []string         `json:"methods"`
	Amount   int64            `json:"amount"`
	DevMode  bool             `json:"devMode"`
	Products []BillingProduct `json:"products"`
}

type abacateResponse[T any] struct {
	Data  *T          `json:"data"`
	Error interface{} `json:"error"`
}

func (c *Client) CreateBilling(ctx context.Context, req BillingRequest) (BillingResponse, json.RawMessage, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return BillingResponse{}, nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/billing/create", bytes.NewReader(body))
	if err != nil {
		return BillingResponse{}, nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return BillingResponse{}, nil, err
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return BillingResponse{}, raw, fmt.Errorf("abacatepay error: %s", string(raw))
	}

	var envelope abacateResponse[BillingResponse]
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return BillingResponse{}, raw, err
	}
	if envelope.Data == nil {
		return BillingResponse{}, raw, errors.New("empty response")
	}
	return *envelope.Data, raw, nil
}

func (c *Client) ListBillings(ctx context.Context) ([]BillingResponse, json.RawMessage, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/billing/list", nil)
	if err != nil {
		return nil, nil, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return nil, raw, fmt.Errorf("abacatepay error: %s", string(raw))
	}

	var envelope abacateResponse[[]BillingResponse]
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return nil, raw, err
	}
	if envelope.Data == nil {
		return nil, raw, errors.New("empty response")
	}
	return *envelope.Data, raw, nil
}

func (c *Client) GetBillingByID(ctx context.Context, billingID string) (BillingResponse, json.RawMessage, error) {
	billings, raw, err := c.ListBillings(ctx)
	if err != nil {
		return BillingResponse{}, raw, err
	}
	for _, billing := range billings {
		if billing.ID == billingID {
			return billing, raw, nil
		}
	}
	return BillingResponse{}, raw, fmt.Errorf("billing not found")
}

// Webhook handling

type AbacateWebhookEvent struct {
	Event     string          `json:"event"`
	Data      json.RawMessage `json:"data"`
	Raw       json.RawMessage `json:"-"`
	BillingID string          `json:"-"`
}

func ParseWebhook(body []byte) (AbacateWebhookEvent, error) {
	var evt AbacateWebhookEvent
	if err := json.Unmarshal(body, &evt); err != nil {
		return evt, err
	}
	evt.BillingID = extractBillingID(evt.Data)
	evt.Raw = body
	return evt, nil
}

func VerifyWebhookSignature(secret string, body []byte, signature string) bool {
	if secret == "" || signature == "" {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expected := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}

func extractBillingID(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}

	type webhookData struct {
		Billing *struct {
			ID string `json:"id"`
		} `json:"billing"`
		BillingID string `json:"billingId"`
		ID        string `json:"id"`
	}

	var payload webhookData
	if err := json.Unmarshal(raw, &payload); err != nil {
		return ""
	}
	if payload.Billing != nil && payload.Billing.ID != "" {
		return payload.Billing.ID
	}
	if payload.BillingID != "" {
		return payload.BillingID
	}
	return payload.ID
}

func WebhookSecretFromQuery(rawQuery string) string {
	values, err := url.ParseQuery(rawQuery)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(values.Get("webhookSecret"))
}
