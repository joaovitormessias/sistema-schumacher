package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"github.com/MicahParks/keyfunc"
	"github.com/golang-jwt/jwt/v4"
)

type contextKey string

const userIDKey contextKey = "user_id"

// Authenticator validates Supabase JWTs using JWKS.
type Authenticator struct {
	jwks     *keyfunc.JWKS
	issuer   string
	audience string
	services map[string]string
	skip     bool
}

func NewAuthenticator(jwksURL, issuer, audience string, serviceTokens []string, skip bool) (*Authenticator, error) {
	if skip {
		return &Authenticator{skip: true}, nil
	}

	services := make(map[string]string, len(serviceTokens))
	for _, token := range serviceTokens {
		trimmed := strings.TrimSpace(token)
		if trimmed == "" {
			continue
		}
		services[trimmed] = serviceSubject(trimmed)
	}

	options := keyfunc.Options{
		RefreshInterval:     time.Hour,
		RefreshErrorHandler: func(err error) {},
		RefreshTimeout:      10 * time.Second,
	}
	jwks, err := keyfunc.Get(jwksURL, options)
	if err != nil {
		return nil, err
	}
	return &Authenticator{jwks: jwks, issuer: issuer, audience: audience, services: services}, nil
}

func (a *Authenticator) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if a.skip {
			next.ServeHTTP(w, r)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "missing authorization", http.StatusUnauthorized)
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			http.Error(w, "invalid authorization", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]
		if subject, ok := a.services[tokenString]; ok {
			ctx := context.WithValue(r.Context(), userIDKey, subject)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		token, err := jwt.Parse(tokenString, a.jwks.Keyfunc)
		if err != nil || !token.Valid {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "invalid claims", http.StatusUnauthorized)
			return
		}

		if iss, _ := claims["iss"].(string); a.issuer != "" && iss != a.issuer {
			http.Error(w, "invalid issuer", http.StatusUnauthorized)
			return
		}

		if a.audience != "" {
			aud := claims["aud"]
			if !audMatches(a.audience, aud) {
				http.Error(w, "invalid audience", http.StatusUnauthorized)
				return
			}
		}

		sub, _ := claims["sub"].(string)
		if sub == "" {
			http.Error(w, "invalid subject", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), userIDKey, sub)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func UserIDFromContext(ctx context.Context) (string, bool) {
	v := ctx.Value(userIDKey)
	if v == nil {
		return "", false
	}
	id, ok := v.(string)
	return id, ok
}

func audMatches(expected string, aud interface{}) bool {
	switch v := aud.(type) {
	case string:
		return v == expected
	case []interface{}:
		for _, item := range v {
			if s, ok := item.(string); ok && s == expected {
				return true
			}
		}
	}
	return false
}

func serviceSubject(token string) string {
	sum := sha256.Sum256([]byte(token))
	return "service:" + hex.EncodeToString(sum[:6])
}
