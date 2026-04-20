package config

import (
	"errors"
	"os"
	"strings"
)

type Config struct {
	AppEnv                  string
	Port                    string
	CORSOrigins             string
	DatabaseURL             string
	SupabaseURL             string
	SupabaseAnonKey         string
	SupabaseServiceRoleKey  string
	SupabaseJWKSURL         string
	SupabaseIssuer          string
	SupabaseAudience        string
	AuthDisabled            bool
	APIServiceTokens        []string
	PagarmeSecretKey        string
	PagarmeBaseURL          string
	PagarmeAPIBaseURL       string
	PagarmeDebugRecipientID string
	PagarmeWebhookBasicUser string
	PagarmeWebhookBasicPass string
	AbacatePayAPIKey        string
	AbacatePayWebhookSecret string
	AbacatePayBaseURL       string
	AbacatePayPublicKey     string
	// TODO(abacatepay-domain): Use hosted frontend URLs (not localhost) in production.
	AbacatePayReturnURL string
	// TODO(abacatepay-domain): Use hosted frontend URLs (not localhost) in production.
	AbacatePayCompletionURL       string
	PaymentNotificationWebhookURL string
}

func Load() (Config, error) {
	cfg := Config{
		AppEnv:                        getEnv("APP_ENV", "production"),
		Port:                          getEnv("PORT", "8080"),
		CORSOrigins:                   os.Getenv("CORS_ORIGINS"),
		DatabaseURL:                   os.Getenv("DATABASE_URL"),
		SupabaseURL:                   os.Getenv("SUPABASE_URL"),
		SupabaseAnonKey:               os.Getenv("SUPABASE_ANON_KEY"),
		SupabaseServiceRoleKey:        os.Getenv("SUPABASE_SERVICE_ROLE_KEY"),
		SupabaseJWKSURL:               os.Getenv("SUPABASE_JWKS_URL"),
		SupabaseIssuer:                os.Getenv("SUPABASE_ISSUER"),
		SupabaseAudience:              getEnv("SUPABASE_AUDIENCE", "authenticated"),
		AuthDisabled:                  parseBool(os.Getenv("AUTH_DISABLED")),
		APIServiceTokens:              splitCSV(os.Getenv("API_SERVICE_TOKENS")),
		PagarmeSecretKey:              os.Getenv("PAGARME_SECRET_KEY"),
		PagarmeBaseURL:                os.Getenv("PAGARME_BASE_URL"),
		PagarmeAPIBaseURL:             getEnv("PAGARME_API_BASE_URL", getEnv("PAGARME_BASE_URL", "https://api.pagar.me/core/v5")),
		PagarmeDebugRecipientID:       strings.TrimSpace(os.Getenv("PAGARME_DEBUG_RECIPIENT_ID")),
		PagarmeWebhookBasicUser:       os.Getenv("PAGARME_WEBHOOK_BASIC_USER"),
		PagarmeWebhookBasicPass:       os.Getenv("PAGARME_WEBHOOK_BASIC_PASS"),
		AbacatePayAPIKey:              os.Getenv("ABACATEPAY_API_KEY"),
		AbacatePayWebhookSecret:       os.Getenv("ABACATEPAY_WEBHOOK_SECRET"),
		AbacatePayBaseURL:             os.Getenv("ABACATEPAY_BASE_URL"),
		AbacatePayPublicKey:           os.Getenv("ABACATEPAY_PUBLIC_KEY"),
		AbacatePayReturnURL:           os.Getenv("ABACATEPAY_RETURN_URL"),
		AbacatePayCompletionURL:       os.Getenv("ABACATEPAY_COMPLETION_URL"),
		PaymentNotificationWebhookURL: strings.TrimSpace(os.Getenv("PAYMENT_NOTIFICATION_WEBHOOK_URL")),
	}

	if cfg.DatabaseURL == "" {
		return cfg, errors.New("DATABASE_URL is required")
	}
	if cfg.SupabaseJWKSURL == "" || cfg.SupabaseIssuer == "" {
		return cfg, errors.New("SUPABASE_JWKS_URL and SUPABASE_ISSUER are required")
	}
	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func parseBool(val string) bool {
	switch strings.ToLower(strings.TrimSpace(val)) {
	case "1", "true", "yes", "y", "on":
		return true
	default:
		return false
	}
}

func splitCSV(val string) []string {
	if strings.TrimSpace(val) == "" {
		return nil
	}

	parts := strings.Split(val, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		out = append(out, trimmed)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
