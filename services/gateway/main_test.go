// ==============================================================================
// NOVARIS: Distributed Payment Settlement & Adaptive Fraud Intelligence Platform
// INTEGRATION TEST SUITE: Full Perimeter Ingress Pipeline (Day 4)
// PACKAGE: main
// ==============================================================================
//
// Tests the full Defense-in-Depth pipeline:
//   Layer 1: auth.JWTMiddleware (Identity Authentication)
//   Layer 2: ratelimit.RateLimitMiddleware (Token-Bucket Volume/Burst Throttling)
//   Layer 3: auth.HMACMiddleware (Payload Integrity & Replay Defense)
//   Core:    paymentsHandler (Authorized Financial Settlement)
// ==============================================================================

package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/charan21042005/novaris/services/gateway/internal/auth"
	"github.com/charan21042005/novaris/services/gateway/internal/ratelimit"
)

// setupTestPipeline initializes a test gateway pipeline with custom rate limit settings.
func setupTestPipeline(burstCapacity, refillRate float64) http.Handler {
	bucketStore := ratelimit.NewBucketStore(burstCapacity, refillRate)

	pipeline := auth.JWTMiddleware(
		ratelimit.RateLimitMiddleware(bucketStore)(
			auth.HMACMiddleware(paymentsHandler),
		),
	)

	return pipeline
}

// buildSignedPaymentRequest creates an HTTP request with valid JWT and HMAC-SHA256 headers.
func buildSignedPaymentRequest(t *testing.T, merchantID string, bodyBytes []byte) *http.Request {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, "/v1/payments", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	// 1. Generate valid JWT token for the merchant
	token, err := auth.GenerateToken(merchantID, "merchant", 15*time.Minute)
	if err != nil {
		t.Fatalf("failed to generate test JWT token: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	// 2. Compute valid HMAC signature
	now := time.Now().Unix()
	timestampStr := strconv.FormatInt(now, 10)
	sig := auth.ComputeHMACSignature(bodyBytes, now, auth.GetHMACSecret())
	req.Header.Set("X-Novaris-Timestamp", timestampStr)
	req.Header.Set("X-Novaris-Signature", sig)

	return req
}

// TestPipeline_BurstAllowanceAndRateLimitEnforcement verifies that:
// 1. Exactly burstCapacity requests succeed with HTTP 200 OK.
// 2. The (burstCapacity + 1)-th request is throttled with HTTP 429 and Retry-After header.
// 3. A second merchant is unaffected (Noisy Neighbor isolation).
func TestPipeline_BurstAllowanceAndRateLimitEnforcement(t *testing.T) {
	// Configure test limiter with burst capacity = 5.0, refill = 1.0 token/sec
	pipeline := setupTestPipeline(5.0, 1.0)
	body := []byte(`{"amount_cents":5000,"recipient":"merchant_test_receiver"}`)

	merchantA := "merchant_burst_alpha"
	merchantB := "merchant_isolated_beta"

	// 1. Send 5 burst requests from Merchant A (all must succeed with 200 OK)
	for i := 1; i <= 5; i++ {
		req := buildSignedPaymentRequest(t, merchantA, body)
		rec := httptest.NewRecorder()

		pipeline.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("burst request %d of 5 for merchant %s failed with status %d: %s",
				i, merchantA, rec.Code, rec.Body.String())
		}

		// Verify structured response
		var resp PaymentResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode payment response: %v", err)
		}
		if resp.Status != "authorized" || resp.MerchantID != merchantA {
			t.Errorf("unexpected response content: %+v", resp)
		}
	}

	// 2. The 6th request from Merchant A must be rejected with HTTP 429 Too Many Requests
	{
		req := buildSignedPaymentRequest(t, merchantA, body)
		rec := httptest.NewRecorder()

		pipeline.ServeHTTP(rec, req)

		if rec.Code != http.StatusTooManyRequests {
			t.Fatalf("expected 6th request to be 429 Too Many Requests, got status %d: %s",
				rec.Code, rec.Body.String())
		}

		// Verify Retry-After header exists and is valid
		retryAfterHeader := rec.Header().Get("Retry-After")
		if retryAfterHeader == "" {
			t.Error("expected Retry-After header on 429 response, but none was found")
		}
		retryVal, err := strconv.Atoi(retryAfterHeader)
		if err != nil || retryVal < 1 {
			t.Errorf("expected Retry-After >= 1, got '%s'", retryAfterHeader)
		}
	}

	// 3. Merchant B sends a request while Merchant A is throttled (Noisy Neighbor Isolation)
	// Merchant B must succeed with HTTP 200 OK because they have their own independent bucket
	{
		req := buildSignedPaymentRequest(t, merchantB, body)
		rec := httptest.NewRecorder()

		pipeline.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("merchant B was unfairly throttled by merchant A's flood! Status: %d", rec.Code)
		}
	}
}

// TestPipeline_FailFastSecurityOrdering verifies that unauthenticated requests
// are blocked at Layer 1 (JWT) without touching the rate limiter or HMAC layers.
func TestPipeline_FailFastSecurityOrdering(t *testing.T) {
	pipeline := setupTestPipeline(10.0, 2.0)
	body := []byte(`{"amount_cents":2500,"recipient":"merchant_test"}`)

	// Send request with no JWT token
	req := httptest.NewRequest(http.MethodPost, "/v1/payments", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	pipeline.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized for missing JWT, got %d", rec.Code)
	}
}
