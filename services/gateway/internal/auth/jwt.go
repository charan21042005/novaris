// ==============================================================================
// NOVARIS: Distributed Payment Settlement & Adaptive Fraud Intelligence Platform
// MODULE 1: Cryptographic JWT Authentication & Security Middleware
// ==============================================================================
//
// Theoretical Context:
// The API Gateway sits on the public perimeter. While /healthz remains public,
// financial routes (/v1/payments) must verify caller identity before invoking
// any downstream processing.
//
// Key Security Protections in this package:
// 1. Strict HMAC-SHA256 enforcement: Defends against the classic "alg: none" exploit
//    and algorithm confusion attacks by strictly validating token.Method.
// 2. Strongly-typed Custom Claims: Binds merchant_id and role to standard RegisteredClaims.
// 3. Collision-free Context Injection: Injects validated claims into request context
//    using an unexported package type to prevent key collisions across the call stack.
// ==============================================================================

package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// contextKey is an unexported type for keys defined in this package.
// This prevents collisions with context keys defined in other packages.
type contextKey string

// ClaimsContextKey is the specific key used to store/retrieve Claims from context.Context.
const ClaimsContextKey contextKey = "novaris_jwt_claims"

// DefaultDevSecret is used for local development when JWT_SECRET is not configured.
// In production, this must be injected from Kubernetes Secrets or HashiCorp Vault.
const DefaultDevSecret = "novaris_dev_super_secret_signing_key_2026_do_not_use_in_prod"

// Claims defines the payload schema of a NOVARIS JWT.
// It bundles application-specific merchant identity with RFC 7519 standard claims.
type Claims struct {
	MerchantID string `json:"merchant_id"` // Unique merchant/store identifier
	Role       string `json:"role"`        // RBAC role: e.g. "merchant_admin", "analyst", "service"
	jwt.RegisteredClaims
}

// ErrorResponse defines the standard JSON error payload returned for 401 Unauthorized responses.
type ErrorResponse struct {
	Error     string    `json:"error"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// GetJWTSecret retrieves the HMAC secret key from the environment, falling back to development secret.
func GetJWTSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return []byte(DefaultDevSecret)
	}
	return []byte(secret)
}

// GenerateToken creates and signs a new JWT using HMAC-SHA256 (HS256).
// It populates standard registered claims (Issuer, Subject, ExpiresAt, IssuedAt)
// along with NOVARIS domain claims (MerchantID, Role).
func GenerateToken(merchantID, role string, duration time.Duration) (string, error) {
	now := time.Now().UTC()
	claims := Claims{
		MerchantID: merchantID,
		Role:       role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "novaris-gateway",
			Subject:   merchantID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	// Create a new token object specifying the HS256 signing method
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the shared HMAC secret key
	signedString, err := token.SignedString(GetJWTSecret())
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT: %w", err)
	}

	return signedString, nil
}

// ValidateToken parses, cryptographically verifies, and extracts claims from a raw JWT string.
func ValidateToken(tokenString string) (*Claims, error) {
	// ParseWithClaims parses the token and validates the signature against our keyFunc
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		// CRITICAL SECURITY CHECK (Defense against "alg: none" & algorithm confusion):
		// Verify that the signing algorithm in the header is genuinely HMAC.
		// If an attacker sends an RSA token or "alg: none", reject it immediately!
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing algorithm in header: %v", t.Header["alg"])
		}
		return GetJWTSecret(), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, fmt.Errorf("token has expired: %w", err)
		}
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	// Extract and return typed claims if token is valid
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token claims")
}

// JWTMiddleware wraps an http.HandlerFunc with JWT authentication.
// Requests without a valid 'Authorization: Bearer <token>' header are rejected with HTTP 401.
// Valid requests have their Claims injected into r.Context() and pass to next.
func JWTMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			writeUnauthorized(w, "Missing Authorization header. Expected format: 'Authorization: Bearer <token>'")
			return
		}

		// Split header into parts: expected format is "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			writeUnauthorized(w, "Malformed Authorization header. Format must be 'Bearer <token>'")
			return
		}

		tokenString := strings.TrimSpace(parts[1])
		if tokenString == "" {
			writeUnauthorized(w, "Bearer token is empty")
			return
		}

		// Cryptographically validate the token and check expiration
		claims, err := ValidateToken(tokenString)
		if err != nil {
			writeUnauthorized(w, fmt.Sprintf("Authentication failed: %v", err))
			return
		}

		// Inject validated claims into the request context
		ctx := context.WithValue(r.Context(), ClaimsContextKey, claims)

		// Pass execution to the next handler with the enriched context
		next(w, r.WithContext(ctx))
	}
}

// GetClaimsFromContext extracts validated Claims from an HTTP request context.
// Returns (claims, true) if present, or (nil, false) if not found.
func GetClaimsFromContext(ctx context.Context) (*Claims, bool) {
	claims, ok := ctx.Value(ClaimsContextKey).(*Claims)
	return claims, ok
}

// writeUnauthorized writes a standardized 401 Unauthorized JSON error response.
func writeUnauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)

	resp := ErrorResponse{
		Error:     "Unauthorized",
		Message:   message,
		Timestamp: time.Now().UTC(),
	}

	_ = json.NewEncoder(w).Encode(resp)
}
