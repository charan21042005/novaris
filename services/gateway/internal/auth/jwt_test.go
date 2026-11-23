// ==============================================================================
// NOVARIS: Distributed Payment Settlement & Adaptive Fraud Intelligence Platform
// UNIT TESTS: JWT Authentication & Validation Engine
// ==============================================================================

package auth_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/charan21042005/novaris/services/gateway/internal/auth"
)

// TestGenerateAndValidateToken verifies that a properly signed token is valid and returns expected claims.
func TestGenerateAndValidateToken(t *testing.T) {
	merchantID := "merchant_test_456"
	role := "merchant_admin"
	duration := 15 * time.Minute

	tokenString, err := auth.GenerateToken(merchantID, role, duration)
	if err != nil {
		t.Fatalf("Expected no error generating token, got: %v", err)
	}

	claims, err := auth.ValidateToken(tokenString)
	if err != nil {
		t.Fatalf("Expected token to be valid, got error: %v", err)
	}

	if claims.MerchantID != merchantID {
		t.Errorf("Expected merchantID %s, got: %s", merchantID, claims.MerchantID)
	}
	if claims.Role != role {
		t.Errorf("Expected role %s, got: %s", role, claims.Role)
	}
}

// TestExpiredToken verifies that an expired token is rejected with an expiration error.
func TestExpiredToken(t *testing.T) {
	// Generate an already-expired token (-5 minutes)
	tokenString, err := auth.GenerateToken("merchant_expired", "analyst", -5*time.Minute)
	if err != nil {
		t.Fatalf("Failed to generate test token: %v", err)
	}

	_, err = auth.ValidateToken(tokenString)
	if err == nil {
		t.Fatal("Expected error for expired token, but validation succeeded")
	}
	if !strings.Contains(err.Error(), "expired") {
		t.Errorf("Expected error to mention 'expired', got: %v", err)
	}
}

// TestTamperedToken verifies that altering any character in the token invalidates the cryptographic signature.
func TestTamperedToken(t *testing.T) {
	tokenString, err := auth.GenerateToken("merchant_target", "admin", 10*time.Minute)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Tamper with the token by appending invalid characters to the signature segment
	tampered := tokenString + "tampered_signature"

	_, err = auth.ValidateToken(tampered)
	if err == nil {
		t.Fatal("Expected error for tampered token signature, but validation succeeded")
	}
}

// TestJWTMiddleware verifies the HTTP middleware rejects unauthenticated requests and admits valid ones.
func TestJWTMiddleware(t *testing.T) {
	// Dummy protected handler that echoes claims
	protectedHandler := auth.JWTMiddleware(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := auth.GetClaimsFromContext(r.Context())
		if !ok {
			t.Errorf("Expected claims to be present in context")
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Welcome, " + claims.MerchantID))
	})

	// Case 1: No Authorization Header (Expect 401)
	reqNoAuth := httptest.NewRequest(http.MethodGet, "/v1/payments", nil)
	rrNoAuth := httptest.NewRecorder()
	protectedHandler(rrNoAuth, reqNoAuth)

	if rrNoAuth.Code != http.StatusUnauthorized {
		t.Errorf("Case 1: Expected status 401 Unauthorized, got: %d", rrNoAuth.Code)
	}

	// Case 2: Malformed Header (Expect 401)
	reqMalformed := httptest.NewRequest(http.MethodGet, "/v1/payments", nil)
	reqMalformed.Header.Set("Authorization", "Basic user:pass")
	rrMalformed := httptest.NewRecorder()
	protectedHandler(rrMalformed, reqMalformed)

	if rrMalformed.Code != http.StatusUnauthorized {
		t.Errorf("Case 2: Expected status 401 for non-bearer header, got: %d", rrMalformed.Code)
	}

	// Case 3: Valid Bearer Token (Expect 200 OK)
	validToken, _ := auth.GenerateToken("merchant_verified_99", "admin", 10*time.Minute)
	reqValid := httptest.NewRequest(http.MethodGet, "/v1/payments", nil)
	reqValid.Header.Set("Authorization", "Bearer "+validToken)
	rrValid := httptest.NewRecorder()
	protectedHandler(rrValid, reqValid)

	if rrValid.Code != http.StatusOK {
		t.Errorf("Case 3: Expected status 200 OK, got: %d", rrValid.Code)
	}
	if !strings.Contains(rrValid.Body.String(), "merchant_verified_99") {
		t.Errorf("Expected response body to contain merchant_verified_99, got: %s", rrValid.Body.String())
	}
}
