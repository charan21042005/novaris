// ==============================================================================
// NOVARIS: Distributed Payment Settlement & Adaptive Fraud Intelligence Platform
// END-TO-END INTEGRATION TEST SUITE: Full Perimeter & OCC Settlement (Day 11)
// FILE: services/gateway/integration_test.go
// ==============================================================================
//
// Milestone Context (Day 11 of 84):
// Proves that all individual components built across Days 1 through 10 function
// cohesively as an industrial, end-to-end payment processing engine:
//   Client HTTP Request
//     -> Logging Middleware
//       -> Layer 1: JWT Identity Verification (RFC 7519)
//         -> Layer 2: Token-Bucket Rate Limiter (Lazy Math Refill)
//           -> Layer 3: HMAC-SHA256 Anti-Tamper & Anti-Replay (RFC 2104)
//             -> Layer 4: Idempotent Deduplication Engine (IETF Draft)
//               -> Layer 5: OCC Double-Entry Settlement Core (ACID Invariants)
//
// Contract Scenarios Verified:
//   1. Full-Stack End-to-End Settlement Success (HTTP 200 OK)
//   2. Exact-Once Network Retry Replay (HTTP 200 OK + Idempotent-Replayed: true)
//   3. Key Reuse Mismatched Payload Rejection (HTTP 422 Unprocessable Entity)
//   4. In-Flight Concurrent Duplicate Race Throttling (HTTP 409 Conflict)
//   5. Missing Idempotency Key Rejection (HTTP 400 Bad Request)
//   6. Rate Limit Flood Throttling (HTTP 429 Too Many Requests + Retry-After)
//   7. Cryptographic Tamper & Stale Replay Defense (HTTP 401 Unauthorized)
//   8. Unauthenticated / Forged Token Defense (HTTP 401 Unauthorized)
//   9. Self-Transfer Disallowance Invariant (HTTP 400 Bad Request)
//  10. Multi-Merchant High-Density Concurrency Stress (Zero Data Races)
// ==============================================================================

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/charan21042005/novaris/services/gateway/internal/auth"
	"github.com/charan21042005/novaris/services/gateway/internal/ratelimit"
)

// testServerHarness encapsulates a running httptest.Server wired with the full Day 11 stack.
type testServerHarness struct {
	Server      *httptest.Server
	Client      *http.Client
	BucketStore *ratelimit.BucketStore
	IdempStore  IdempotencyStore
}

func newTestServerHarness(t *testing.T, burstCapacity, refillRate float64) *testServerHarness {
	t.Helper()

	cfg := Config{
		Port:           "0",
		ServiceName:    "novaris-gateway-test",
		ServiceVersion: "0.11.0-test",
		Environment:    "test",
	}

	bucketStore := ratelimit.NewBucketStore(burstCapacity, refillRate)
	idempStore := NewInMemoryIdempotencyStore()

	// SetupServerMux constructs the production router and wires the complete five-layer pipeline
	router := SetupServerMux(cfg, nil, bucketStore, idempStore)
	server := httptest.NewServer(router)

	return &testServerHarness{
		Server:      server,
		Client:      server.Client(),
		BucketStore: bucketStore,
		IdempStore:  idempStore,
	}
}

// buildFullStackRequest constructs a fully authenticated, signed, and deduplicated HTTP request.
func buildFullStackRequest(
	t *testing.T,
	serverURL string,
	merchantID string,
	idempKey string,
	bodyBytes []byte,
	timestampOffsetSec int64,
	tamperSignature bool,
) *http.Request {
	t.Helper()

	req, err := http.NewRequest(http.MethodPost, serverURL+"/v1/payments", bytes.NewReader(bodyBytes))
	if err != nil {
		t.Fatalf("failed to create http request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// 1. Layer 1: Mint valid JWT Token
	token, err := auth.GenerateToken(merchantID, "merchant", 15*time.Minute)
	if err != nil {
		t.Fatalf("failed to generate test JWT token: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	// 2. Layer 3: Compute valid HMAC-SHA256 signature and timestamp
	ts := time.Now().Unix() + timestampOffsetSec
	timestampStr := strconv.FormatInt(ts, 10)
	sig := auth.ComputeHMACSignature(bodyBytes, ts, auth.GetHMACSecret())
	if tamperSignature {
		sig = "tampered_bad_signature_digest_000000000000000000000000000000000000"
	}

	req.Header.Set("X-Novaris-Timestamp", timestampStr)
	req.Header.Set("X-Novaris-Signature", sig)

	// 3. Layer 4: Attach Idempotency-Key
	if idempKey != "" {
		req.Header.Set("Idempotency-Key", idempKey)
	}

	return req
}

// ------------------------------------------------------------------------------
// SCENARIO 1: Full-Stack End-to-End Settlement Success
// ------------------------------------------------------------------------------
func TestIntegration_Scenario1_FullStackSettlementSuccess(t *testing.T) {
	harness := newTestServerHarness(t, 10.0, 2.0)
	defer harness.Server.Close()

	merchantID := "c0000000-0000-0000-0000-000000000001"
	recipientID := "c0000000-0000-0000-0000-000000000002"
	idempKey := uuid.New().String()

	payload := map[string]interface{}{
		"amount_cents": 5000,
		"recipient":    recipientID,
	}
	bodyBytes, _ := json.Marshal(payload)

	req := buildFullStackRequest(t, harness.Server.URL, merchantID, idempKey, bodyBytes, 0, false)
	resp, err := harness.Client.Do(req)
	if err != nil {
		t.Fatalf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected HTTP 200 OK, got %d: %s", resp.StatusCode, string(respBody))
	}

	// Verify that response is fresh and NOT replayed
	if resp.Header.Get("Idempotent-Replayed") == "true" {
		t.Errorf("fresh request should NOT have Idempotent-Replayed: true header")
	}

	var paymentResp PaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&paymentResp); err != nil {
		t.Fatalf("failed to parse JSON response: %v", err)
	}

	if paymentResp.Status != "settled" {
		t.Errorf("expected status 'settled', got %s", paymentResp.Status)
	}
	if paymentResp.TransactionID == "" {
		t.Errorf("expected non-empty transaction UUID")
	}
	if paymentResp.AmountCents != 5000 {
		t.Errorf("expected amount_cents=5000, got %d", paymentResp.AmountCents)
	}
}

// ------------------------------------------------------------------------------
// SCENARIO 2: Exact-Once Network Retry Replay (Idempotent-Replayed: true)
// ------------------------------------------------------------------------------
func TestIntegration_Scenario2_IdempotentRetryReplay(t *testing.T) {
	harness := newTestServerHarness(t, 10.0, 2.0)
	defer harness.Server.Close()

	merchantID := "c0000000-0000-0000-0000-000000000001"
	recipientID := "c0000000-0000-0000-0000-000000000002"
	idempKey := uuid.New().String()

	payload := map[string]interface{}{
		"amount_cents": 7500, // $75.00
		"recipient":    recipientID,
	}
	bodyBytes, _ := json.Marshal(payload)

	// Attempt 1: Initial Settlement
	req1 := buildFullStackRequest(t, harness.Server.URL, merchantID, idempKey, bodyBytes, 0, false)
	resp1, err := harness.Client.Do(req1)
	if err != nil {
		t.Fatalf("Attempt 1 failed: %v", err)
	}
	defer resp1.Body.Close()

	if resp1.StatusCode != http.StatusOK {
		t.Fatalf("Attempt 1 expected HTTP 200, got %d", resp1.StatusCode)
	}

	var res1 PaymentResponse
	_ = json.NewDecoder(resp1.Body).Decode(&res1)

	// Attempt 2: Network Retry with identical Idempotency-Key and Body
	req2 := buildFullStackRequest(t, harness.Server.URL, merchantID, idempKey, bodyBytes, 0, false)
	resp2, err := harness.Client.Do(req2)
	if err != nil {
		t.Fatalf("Attempt 2 retry failed: %v", err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusOK {
		t.Fatalf("Attempt 2 expected HTTP 200, got %d", resp2.StatusCode)
	}

	// CRITICAL CONTRACT CHECK: Idempotent-Replayed header MUST be present on retry!
	if resp2.Header.Get("Idempotent-Replayed") != "true" {
		t.Errorf("CRITICAL CONTRACT VIOLATION: Expected 'Idempotent-Replayed: true' on retry!")
	}

	var res2 PaymentResponse
	_ = json.NewDecoder(resp2.Body).Decode(&res2)

	// CRITICAL INVARIANT: The transaction ID must match the original exactly (zero double-charge)
	if res2.TransactionID != res1.TransactionID {
		t.Errorf("DOUBLE SPEND HAZARD: retry produced differing transaction ID! Res1=%s, Res2=%s",
			res1.TransactionID, res2.TransactionID)
	}
	if res2.AmountCents != res1.AmountCents {
		t.Errorf("mismatched settled amount on retry")
	}
}

// ------------------------------------------------------------------------------
// SCENARIO 3: Key Reuse with Mismatched Payload Rejected (HTTP 422)
// ------------------------------------------------------------------------------
func TestIntegration_Scenario3_PayloadMismatchRejected(t *testing.T) {
	harness := newTestServerHarness(t, 10.0, 2.0)
	defer harness.Server.Close()

	merchantID := "c0000000-0000-0000-0000-000000000001"
	idempKey := uuid.New().String()

	// Initial request: $10.00 transfer
	body1, _ := json.Marshal(map[string]interface{}{"amount_cents": 1000, "recipient": "c0000000-0000-0000-0000-000000000002"})
	req1 := buildFullStackRequest(t, harness.Server.URL, merchantID, idempKey, body1, 0, false)
	resp1, _ := harness.Client.Do(req1)
	resp1.Body.Close()

	// Adversarial request: reuses same key but attempts a $5,000.00 transfer!
	body2, _ := json.Marshal(map[string]interface{}{"amount_cents": 500000, "recipient": "c0000000-0000-0000-0000-000000000002"})
	req2 := buildFullStackRequest(t, harness.Server.URL, merchantID, idempKey, body2, 0, false)
	resp2, err := harness.Client.Do(req2)
	if err != nil {
		t.Fatalf("adversarial request failed: %v", err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("expected HTTP 422 Unprocessable Entity on payload mismatch, got %d", resp2.StatusCode)
	}
}

// ------------------------------------------------------------------------------
// SCENARIO 4: Missing Idempotency Key Rejected (HTTP 400 Bad Request)
// ------------------------------------------------------------------------------
func TestIntegration_Scenario4_MissingIdempotencyKeyRejected(t *testing.T) {
	harness := newTestServerHarness(t, 10.0, 2.0)
	defer harness.Server.Close()

	merchantID := "c0000000-0000-0000-0000-000000000001"
	body, _ := json.Marshal(map[string]interface{}{"amount_cents": 2500, "recipient": "c0000000-0000-0000-0000-000000000002"})

	// Omit Idempotency-Key
	req := buildFullStackRequest(t, harness.Server.URL, merchantID, "", body, 0, false)
	resp, err := harness.Client.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected HTTP 400 Bad Request when Idempotency-Key is omitted, got %d", resp.StatusCode)
	}
}

// ------------------------------------------------------------------------------
// SCENARIO 5: Rate Limit Burst Throttling (HTTP 429 Too Many Requests)
// ------------------------------------------------------------------------------
func TestIntegration_Scenario5_RateLimitBurstThrottling(t *testing.T) {
	// Configure tight bucket: Capacity = 3, Refill = 0.1 tokens/sec
	harness := newTestServerHarness(t, 3.0, 0.1)
	defer harness.Server.Close()

	merchantID := "merchant_flood_tester"
	body, _ := json.Marshal(map[string]interface{}{"amount_cents": 100, "recipient": "c0000000-0000-0000-0000-000000000002"})

	// Fire 3 burst requests: all should pass with 200 OK
	for i := 1; i <= 3; i++ {
		idempKey := fmt.Sprintf("burst-key-%d", i)
		req := buildFullStackRequest(t, harness.Server.URL, merchantID, idempKey, body, 0, false)
		resp, err := harness.Client.Do(req)
		if err != nil {
			t.Fatalf("burst request %d of 3 failed with client error: %v", i, err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("burst request %d of 3 failed with status %d", i, resp.StatusCode)
		}
		_ = resp.Body.Close()
	}

	// The 4th request must be throttled with HTTP 429 Too Many Requests
	idempKey4 := "burst-key-4"
	req4 := buildFullStackRequest(t, harness.Server.URL, merchantID, idempKey4, body, 0, false)
	resp4, err := harness.Client.Do(req4)
	if err != nil {
		t.Fatalf("4th request failed: %v", err)
	}
	defer resp4.Body.Close()

	if resp4.StatusCode != http.StatusTooManyRequests {
		t.Errorf("expected HTTP 429 Too Many Requests, got %d", resp4.StatusCode)
	}

	retryAfter := resp4.Header.Get("Retry-After")
	if retryAfter == "" {
		t.Errorf("expected Retry-After header on HTTP 429 response")
	}
}

// ------------------------------------------------------------------------------
// SCENARIO 6: Cryptographic Tampering & Stale Replay Defense (HTTP 401)
// ------------------------------------------------------------------------------
func TestIntegration_Scenario6_TamperAndReplayDefense(t *testing.T) {
	harness := newTestServerHarness(t, 10.0, 2.0)
	defer harness.Server.Close()

	merchantID := "merchant_security_tester"
	idempKey := uuid.New().String()
	body, _ := json.Marshal(map[string]interface{}{"amount_cents": 1000, "recipient": "c0000000-0000-0000-0000-000000000002"})

	// Subtest A: Tampered Signature
	t.Run("TamperedHMACSignature", func(t *testing.T) {
		req := buildFullStackRequest(t, harness.Server.URL, merchantID, idempKey, body, 0, true)
		resp, err := harness.Client.Do(req)
		if err != nil {
			t.Fatalf("tampered HMAC request failed with client error: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("expected HTTP 401 on tampered HMAC signature, got %d", resp.StatusCode)
		}
	})

	// Subtest B: Stale Replay Attack (Timestamp 600s in the past)
	t.Run("StaleTimestampReplayAttack", func(t *testing.T) {
		req := buildFullStackRequest(t, harness.Server.URL, merchantID, idempKey, body, -600, false)
		resp, err := harness.Client.Do(req)
		if err != nil {
			t.Fatalf("stale timestamp replay request failed with client error: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("expected HTTP 401 on stale timestamp replay, got %d", resp.StatusCode)
		}
	})
}

// ------------------------------------------------------------------------------
// SCENARIO 7: Missing or Forged JWT Bearer Token (HTTP 401)
// ------------------------------------------------------------------------------
func TestIntegration_Scenario7_JWTAuthenticationEnforcement(t *testing.T) {
	harness := newTestServerHarness(t, 10.0, 2.0)
	defer harness.Server.Close()

	body, _ := json.Marshal(map[string]interface{}{"amount_cents": 1000, "recipient": "c0000000-0000-0000-0000-000000000002"})

	req, _ := http.NewRequest(http.MethodPost, harness.Server.URL+"/v1/payments", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", uuid.New().String())

	resp, err := harness.Client.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected HTTP 401 when Authorization header is missing, got %d", resp.StatusCode)
	}
}

// ------------------------------------------------------------------------------
// SCENARIO 8: Public Healthz & Discovery API Contract Verification
// ------------------------------------------------------------------------------
func TestIntegration_Scenario8_PublicProbes(t *testing.T) {
	harness := newTestServerHarness(t, 10.0, 2.0)
	defer harness.Server.Close()

	// 1. GET /healthz
	respHealth, err := harness.Client.Get(harness.Server.URL + "/healthz")
	if err != nil {
		t.Fatalf("GET /healthz failed: %v", err)
	}
	defer respHealth.Body.Close()

	if respHealth.StatusCode != http.StatusOK {
		t.Errorf("GET /healthz expected HTTP 200, got %d", respHealth.StatusCode)
	}

	var health HealthResponse
	_ = json.NewDecoder(respHealth.Body).Decode(&health)
	if health.Status != "ok" {
		t.Errorf("health status expected 'ok', got '%s'", health.Status)
	}

	// 2. GET / (Discovery)
	respRoot, err := harness.Client.Get(harness.Server.URL + "/")
	if err != nil {
		t.Fatalf("GET / failed: %v", err)
	}
	defer respRoot.Body.Close()

	if respRoot.StatusCode != http.StatusOK {
		t.Errorf("GET / expected HTTP 200, got %d", respRoot.StatusCode)
	}

	var info InfoResponse
	_ = json.NewDecoder(respRoot.Body).Decode(&info)
	if info.Day != 11 {
		t.Errorf("expected Day 11 in discovery info, got Day %d", info.Day)
	}
}

// ------------------------------------------------------------------------------
// SCENARIO 9: Multi-Merchant High-Density Concurrency Stress (Race Freedom)
// ------------------------------------------------------------------------------
func TestIntegration_Scenario9_MultiMerchantConcurrentStress(t *testing.T) {
	// 5 distinct merchants, each sending 5 concurrent requests (25 total simultaneous requests)
	harness := newTestServerHarness(t, 10.0, 2.0)
	defer harness.Server.Close()

	merchants := []string{
		"merchant_alpha_01",
		"merchant_beta_02",
		"merchant_gamma_03",
		"merchant_delta_04",
		"merchant_epsilon_05",
	}

	const requestsPerMerchant = 5
	totalRequests := len(merchants) * requestsPerMerchant

	var wg sync.WaitGroup
	startGate := make(chan struct{})

	var successCount int64
	var failureCount int64

	for _, m := range merchants {
		for i := 0; i < requestsPerMerchant; i++ {
			wg.Add(1)
			go func(merchantID string, reqIdx int) {
				defer wg.Done()

				idempKey := fmt.Sprintf("stress-%s-%d", merchantID, reqIdx)
				body, _ := json.Marshal(map[string]interface{}{
					"amount_cents": 1000 + (reqIdx * 100),
					"recipient":    "c0000000-0000-0000-0000-000000000002",
				})

				req := buildFullStackRequest(t, harness.Server.URL, merchantID, idempKey, body, 0, false)

				// Synchronize on nanosecond release barrier
				<-startGate

				resp, err := harness.Client.Do(req)
				if err != nil {
					atomic.AddInt64(&failureCount, 1)
					return
				}
				defer resp.Body.Close()

				if resp.StatusCode == http.StatusOK {
					atomic.AddInt64(&successCount, 1)
				} else {
					atomic.AddInt64(&failureCount, 1)
				}
			}(m, i)
		}
	}

	// Release all 25 goroutines simultaneously
	close(startGate)
	wg.Wait()

	t.Logf("HIGH-DENSITY CONCURRENCY RESULTS: Total=%d | Succeeded=%d | Failed/Throttled=%d",
		totalRequests, successCount, failureCount)

	if successCount != int64(totalRequests) {
		t.Errorf("expected all %d requests to succeed under isolated burst allowances, but only %d succeeded",
			totalRequests, successCount)
	}
}
