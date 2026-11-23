// ==============================================================================
// NOVARIS: Distributed Payment Settlement & Adaptive Fraud Intelligence Platform
// UNIT TEST SUITE: HMAC Cryptographic Verification & Middleware (Day 3)
// PACKAGE: internal/auth/hmac_test.go
// ==============================================================================

package auth

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

func TestComputeAndVerifyHMAC_Success(t *testing.T) {
	secret := []byte("test_secret_key_12345")
	body := []byte(`{"amount_cents":5000,"recipient":"merchant_456"}`)
	now := time.Now().Unix()

	sig := ComputeHMACSignature(body, now, secret)
	if sig == "" {
		t.Fatal("expected non-empty signature string")
	}

	valid, err := VerifyHMACSignature(body, strconv.FormatInt(now, 10), sig, secret)
	if err != nil {
		t.Fatalf("expected valid signature, got error: %v", err)
	}
	if !valid {
		t.Fatal("expected verification to succeed, got false")
	}
}

func TestVerifyHMAC_TamperedBody(t *testing.T) {
	secret := []byte("test_secret_key_12345")
	originalBody := []byte(`{"amount_cents":5000,"recipient":"merchant_456"}`)
	tamperedBody := []byte(`{"amount_cents":5000000,"recipient":"attacker"}`)
	now := time.Now().Unix()

	// Compute signature on original body
	sig := ComputeHMACSignature(originalBody, now, secret)

	// Verify using tampered body
	valid, err := VerifyHMACSignature(tamperedBody, strconv.FormatInt(now, 10), sig, secret)
	if err == nil {
		t.Fatal("expected error validating tampered body, got nil")
	}
	if valid {
		t.Fatal("expected validation to fail for tampered body, got true")
	}
}

func TestVerifyHMAC_StaleTimestamp(t *testing.T) {
	secret := []byte("test_secret_key_12345")
	body := []byte(`{"amount_cents":100}`)
	// Stale timestamp 600 seconds (10 minutes) in the past
	staleTime := time.Now().Unix() - 600

	sig := ComputeHMACSignature(body, staleTime, secret)

	valid, err := VerifyHMACSignature(body, strconv.FormatInt(staleTime, 10), sig, secret)
	if err == nil {
		t.Fatal("expected error validating stale timestamp, got nil")
	}
	if valid {
		t.Fatal("expected validation to fail for stale timestamp, got true")
	}
}

func TestVerifyHMAC_FutureTimestamp(t *testing.T) {
	secret := []byte("test_secret_key_12345")
	body := []byte(`{"amount_cents":100}`)
	// Future timestamp 600 seconds (10 minutes) ahead
	futureTime := time.Now().Unix() + 600

	sig := ComputeHMACSignature(body, futureTime, secret)

	valid, err := VerifyHMACSignature(body, strconv.FormatInt(futureTime, 10), sig, secret)
	if err == nil {
		t.Fatal("expected error validating future timestamp, got nil")
	}
	if valid {
		t.Fatal("expected validation to fail for future timestamp, got true")
	}
}

func TestVerifyHMAC_WrongSecret(t *testing.T) {
	correctSecret := []byte("correct_shared_secret")
	wrongSecret := []byte("wrong_attacker_secret")
	body := []byte(`{"amount_cents":100}`)
	now := time.Now().Unix()

	sig := ComputeHMACSignature(body, now, wrongSecret)

	valid, err := VerifyHMACSignature(body, strconv.FormatInt(now, 10), sig, correctSecret)
	if err == nil {
		t.Fatal("expected error validating signature with wrong secret, got nil")
	}
	if valid {
		t.Fatal("expected validation to fail for wrong secret, got true")
	}
}

func TestHMACMiddleware(t *testing.T) {
	secret := GetHMACSecret()
	body := []byte(`{"amount_cents":7500,"currency":"USD"}`)

	var receivedBody []byte
	handlerRan := false

	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerRan = true
		// Verify body is intact and can still be read
		b, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("inner handler failed to read body: %v", err)
		}
		receivedBody = b
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"success"}`))
	})

	middleware := HMACMiddleware(dummyHandler)

	// Case 1: Missing headers
	reqMissing := httptest.NewRequest(http.MethodPost, "/v1/payments", bytes.NewReader(body))
	recMissing := httptest.NewRecorder()
	middleware.ServeHTTP(recMissing, reqMissing)
	if recMissing.Code != http.StatusUnauthorized {
		t.Errorf("Case 1: expected 401 for missing headers, got %d", recMissing.Code)
	}

	// Case 2: Tampered body
	now := time.Now().Unix()
	nowStr := strconv.FormatInt(now, 10)
	validSig := ComputeHMACSignature(body, now, secret)

	tamperedBody := []byte(`{"amount_cents":9999999,"currency":"USD"}`)
	reqTampered := httptest.NewRequest(http.MethodPost, "/v1/payments", bytes.NewReader(tamperedBody))
	reqTampered.Header.Set(HeaderNovarisTimestamp, nowStr)
	reqTampered.Header.Set(HeaderNovarisSignature, validSig)
	recTampered := httptest.NewRecorder()
	middleware.ServeHTTP(recTampered, reqTampered)
	if recTampered.Code != http.StatusUnauthorized {
		t.Errorf("Case 2: expected 401 for tampered body, got %d", recTampered.Code)
	}

	// Case 3: Stale timestamp
	staleTS := strconv.FormatInt(now-400, 10)
	staleSig := ComputeHMACSignature(body, now-400, secret)
	reqStale := httptest.NewRequest(http.MethodPost, "/v1/payments", bytes.NewReader(body))
	reqStale.Header.Set(HeaderNovarisTimestamp, staleTS)
	reqStale.Header.Set(HeaderNovarisSignature, staleSig)
	recStale := httptest.NewRecorder()
	middleware.ServeHTTP(recStale, reqStale)
	if recStale.Code != http.StatusUnauthorized {
		t.Errorf("Case 3: expected 401 for stale timestamp, got %d", recStale.Code)
	}

	// Case 4: Valid request
	reqValid := httptest.NewRequest(http.MethodPost, "/v1/payments", bytes.NewReader(body))
	reqValid.Header.Set(HeaderNovarisTimestamp, nowStr)
	reqValid.Header.Set(HeaderNovarisSignature, validSig)
	recValid := httptest.NewRecorder()
	middleware.ServeHTTP(recValid, reqValid)

	if recValid.Code != http.StatusOK {
		t.Errorf("Case 4: expected 200 for valid request, got %d", recValid.Code)
	}
	if !handlerRan {
		t.Error("Case 4: expected inner handler to be invoked")
	}
	if string(receivedBody) != string(body) {
		t.Errorf("Case 4: expected inner handler to receive %q, got %q", string(body), string(receivedBody))
	}
}
