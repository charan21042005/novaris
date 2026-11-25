// ==============================================================================
// NOVARIS: Distributed Payment Settlement & Adaptive Fraud Intelligence Platform
// LOAD TESTING ENGINE: UNIT TEST SUITE (Day 13)
// FILE: services/gateway/cmd/loadtest/main_test.go
// ==============================================================================
//
// 📚 COMPUTER SCIENCE & TEST HARNESS DESIGN FOR LOAD TOOLS:
//
// 1. ISOLATED MOCK LOAD VERIFICATION:
//    Load test harnesses must themselves be verified to guarantee that:
//      - The goroutine worker pool correctly coordinates over channels without deadlocks.
//      - Status code counters (200, 409, 429, etc.) aggregate accurately under concurrency.
//      - Percentile algorithms (p50, p90, p95, p99) handle varying slice sizes correctly.
//      - HTTP connections are properly closed without socket leaks.
// ==============================================================================

package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

// TestRunLoadTest_MockGatewaySuccess asserts that RunLoadTest dispatches the configured
// number of requests, aggregates metrics correctly, and measures latencies without crashing.
func TestRunLoadTest_MockGatewaySuccess(t *testing.T) {
	var requestCount atomic.Int64

	// Spin up a high-performance in-memory mock gateway server:
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := requestCount.Add(1)

		w.Header().Set("Content-Type", "application/json")

		// Simulate defensive perimeter behavior:
		// First 10 requests get HTTP 200 OK (Settled)
		// Next 5 requests get HTTP 429 Too Many Requests (Rate limited)
		// Remaining requests get HTTP 409 Conflict (OCC Collision)
		switch {
		case count <= 10:
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "settled"})
		case count <= 15:
			w.WriteHeader(http.StatusTooManyRequests)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "rate_limited"})
		default:
			w.WriteHeader(http.StatusConflict)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "occ_conflict"})
		}
	}))
	defer server.Close()

	cfg := LoadTestConfig{
		BaseURL:     server.URL,
		NumWorkers:  5,
		TotalReqs:   25,
		HotWallet:   true,
		NumTenants:  1,
		AmountCents: 100,
		Timeout:     2 * time.Second,
	}

	report := RunLoadTest(cfg)

	// Assertion 1: Total dispatched must match configuration:
	if report.TotalDispatched != 25 {
		t.Errorf("expected 25 total dispatched requests, got %d", report.TotalDispatched)
	}

	// Assertion 2: Verify status code distribution:
	if report.Status200Count != 10 {
		t.Errorf("expected 10 status 200 responses, got %d", report.Status200Count)
	}
	if report.Status429Count != 5 {
		t.Errorf("expected 5 status 429 responses, got %d", report.Status429Count)
	}
	if report.Status409Count != 10 {
		t.Errorf("expected 10 status 409 responses, got %d", report.Status409Count)
	}

	// Assertion 3: Zero internal errors or network drops:
	if report.Status500Count != 0 {
		t.Errorf("expected 0 status 500 errors, got %d", report.Status500Count)
	}
	if report.NetworkErrors != 0 {
		t.Errorf("expected 0 network errors, got %d", report.NetworkErrors)
	}

	// Assertion 4: Verify latency statistics exist and are non-zero:
	if len(report.LatenciesMs) != 25 {
		t.Errorf("expected 25 latency data points, got %d", len(report.LatenciesMs))
	}
	if report.MinMs < 0 || report.P50Ms <= 0 || report.P99Ms <= 0 {
		t.Errorf("expected valid non-negative latency percentiles, got min=%v, p50=%v, p99=%v", report.MinMs, report.P50Ms, report.P99Ms)
	}
	if report.TPS <= 0 {
		t.Errorf("expected positive throughput (TPS), got %v", report.TPS)
	}
}

// TestPct_MathPrecision asserts percentage helper behaves correctly on edge cases.
func TestPct_MathPrecision(t *testing.T) {
	if res := pct(0, 100); res != 0.0 {
		t.Errorf("expected 0.0%%, got %v", res)
	}
	if res := pct(50, 100); res != 50.0 {
		t.Errorf("expected 50.0%%, got %v", res)
	}
	if res := pct(100, 0); res != 0.0 {
		t.Errorf("expected 0.0%% on zero total, got %v", res)
	}
}
