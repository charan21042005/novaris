// ==============================================================================
// NOVARIS: Distributed Payment Settlement & Adaptive Fraud Intelligence Platform
// UNIT TEST SUITE: End-to-End CLI Probe Verification (Day 11)
// FILE: services/gateway/cmd/e2eprobe/main_test.go
// ==============================================================================

package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProbe_MockGatewayServer(t *testing.T) {
	mux := http.NewServeMux()

	// 1. /healthz
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok", "service": "novaris-gateway"})
	})

	// 2. /v1/payments
	mux.HandleFunc("/v1/payments", func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
			return
		}

		idempKey := r.Header.Get("Idempotency-Key")
		if idempKey == "" {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "bad_request"})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":         "settled",
			"transaction_id": "test-mock-tx-12345",
			"amount_cents":   5000,
		})
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	// Verify healthz probe
	resp, err := http.Get(ts.URL + "/healthz")
	if err != nil {
		t.Fatalf("failed to probe mock healthz: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}
