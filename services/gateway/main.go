// ==============================================================================
// NOVARIS: Distributed Payment Settlement & Adaptive Fraud Intelligence Platform
// MODULE 1: API Gateway Service Entry Point (Day 4 - JWT + Rate Limiting + HMAC Defense)
// ==============================================================================
//
// Day 4 Evolution:
// 1. /healthz remains public for unauthenticated Kubernetes/Docker health probing.
// 2. / (root) provides public service discovery with updated Day 4 capabilities.
// 3. /v1/payments is now protected by a THREE-LAYER DEFENSE IN DEPTH PIPELINE:
//    - Layer 1 (Identity): auth.JWTMiddleware
//      Answers "WHO ARE YOU?": Validates the Bearer token's cryptographic signature,
//      checks token expiration, and injects merchant identity claims into context.
//    - Layer 2 (Velocity / Volume): ratelimit.RateLimitMiddleware
//      Answers "HOW MUCH / HOW FAST?": Enforces per-merchant token-bucket rate limiting
//      (Capacity = 10 burst, Refill = 2 tokens/sec). Evaluated via lazy math refill
//      without background goroutines. Returns HTTP 429 Too Many Requests with Retry-After.
//      *CRITICAL FAIL-FAST OPTIMIZATION*: Drops abusive traffic BEFORE costly HMAC hashing!
//    - Layer 3 (Integrity / Replay): auth.HMACMiddleware
//      Answers "HAS THIS BEEN ALTERED OR REPLAYED?": Reads request body bytes, verifies
//      X-Novaris-Timestamp within ±300s window, recomputes HMAC-SHA256 signature using
//      constant-time comparison, and restores r.Body for downstream handler consumption.
//    - Core (Settlement): paymentsHandler
//      Reads verified claims and payment payload, returning authorized settlement JSON.
// ==============================================================================

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/charan21042005/novaris/services/gateway/internal/auth"
	"github.com/charan21042005/novaris/services/gateway/internal/ratelimit"
)

// Config holds runtime configuration parameters for the gateway.
type Config struct {
	Port           string
	ServiceName    string
	ServiceVersion string
	Environment    string
}

// HealthResponse defines the structured JSON payload returned by the /healthz probe.
type HealthResponse struct {
	Status      string    `json:"status"`
	Service     string    `json:"service"`
	Version     string    `json:"version"`
	Environment string    `json:"environment"`
	Timestamp   time.Time `json:"timestamp"`
}

// InfoResponse provides discoverability metadata for public and protected routes.
type InfoResponse struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Status      string            `json:"status"`
	Version     string            `json:"version"`
	Day         int               `json:"day"`
	Phase       string            `json:"phase"`
	Endpoints   map[string]string `json:"endpoints"`
}

// PaymentRequest defines the expected payload for initiating a settlement transaction.
type PaymentRequest struct {
	AmountCents int64  `json:"amount_cents"`
	Recipient   string `json:"recipient"`
	Currency    string `json:"currency,omitempty"`
}

// PaymentResponse is returned by /v1/payments upon successful multi-layer verification.
type PaymentResponse struct {
	Status      string    `json:"status"`
	MerchantID  string    `json:"merchant_id"`
	Role        string    `json:"role"`
	AmountCents int64     `json:"amount_cents,omitempty"`
	Recipient   string    `json:"recipient,omitempty"`
	Message     string    `json:"message"`
	Timestamp   time.Time `json:"timestamp"`
}

// loggingMiddleware intercepts incoming HTTP requests, records processing latency,
// and outputs a structured log line for observability.
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		duration := time.Since(start)
		log.Printf("[HTTP] %s %s from %s in %v", r.Method, r.URL.Path, r.RemoteAddr, duration)
	})
}

// healthzHandler responds with HTTP 200 and system health metadata. (Public route)
func healthzHandler(cfg Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		payload := HealthResponse{
			Status:      "ok",
			Service:     cfg.ServiceName,
			Version:     cfg.ServiceVersion,
			Environment: cfg.Environment,
			Timestamp:   time.Now().UTC(),
		}

		_ = json.NewEncoder(w).Encode(payload)
	}
}

// rootHandler provides introductory information and lists available routes. (Public route)
func rootHandler(cfg Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		payload := InfoResponse{
			Name:        "NOVARIS API Gateway",
			Description: "High-Throughput Distributed Payment Settlement & Fraud Intelligence Ingestion Gateway",
			Status:      "online",
			Version:     cfg.ServiceVersion,
			Day:         4,
			Phase:       "Phase 1 - Design, Ledger & Ingestion Security",
			Endpoints: map[string]string{
				"health_check":       "/healthz (Public Liveness Probe)",
				"payments_protected": "/v1/payments (Protected - Requires JWT Identity + Token-Bucket Rate Limiter [10 burst, 2/s] + HMAC-SHA256 Integrity)",
				"transaction_flow":   "/v1/transactions (Coming in Phase 1: Day 5 Ledger Schema & Double-Entry Ingestion)",
			},
		}

		_ = json.NewEncoder(w).Encode(payload)
	}
}

// paymentsHandler is the protected business settlement endpoint.
// It retrieves verified JWT claims from the context and decodes the intact body.
func paymentsHandler(w http.ResponseWriter, r *http.Request) {
	// Retrieve pre-verified claims attached to context by auth.JWTMiddleware
	claims, ok := auth.GetClaimsFromContext(r.Context())
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintln(w, `{"error":"unauthorized","message":"missing context claims"}`)
		return
	}

	// Read and decode the restored request body (verified intact by auth.HMACMiddleware)
	var req PaymentRequest
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	payload := PaymentResponse{
		Status:      "authorized",
		MerchantID:  claims.MerchantID,
		Role:        claims.Role,
		AmountCents: req.AmountCents,
		Recipient:   req.Recipient,
		Message:     "Payment request verified successfully: JWT identity confirmed, rate limit quota allocated, and HMAC body integrity preserved.",
		Timestamp:   time.Now().UTC(),
	}

	_ = json.NewEncoder(w).Encode(payload)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	cfg := Config{
		Port:           port,
		ServiceName:    "novaris-gateway",
		ServiceVersion: "0.4.0-day4",
		Environment:    "development",
	}

	// Initialize the In-Memory Token-Bucket Rate Limiter Store
	// Default: Burst capacity = 10.0 tokens, Refill rate = 2.0 tokens/second per merchant
	bucketStore := ratelimit.NewBucketStore(
		ratelimit.DefaultBucketCapacity,
		ratelimit.DefaultRefillRate,
	)

	// Create a new ServeMux (HTTP request router)
	mux := http.NewServeMux()

	// 1. Public Routes (No authentication required)
	mux.HandleFunc("/", rootHandler(cfg))
	mux.HandleFunc("/healthz", healthzHandler(cfg))

	// 2. Protected Routes (Stacked Defense-in-Depth Middleware: JWT -> RateLimit -> HMAC -> paymentsHandler)
	// Execution order:
	//   Step 1: JWTMiddleware checks Authorization Bearer token (rejects if missing/forged/expired -> 401).
	//           Injects verified Claims (including MerchantID) into request context.
	//   Step 2: RateLimitMiddleware reads claims.MerchantID and evaluates token bucket quota.
	//           If tokens are exhausted, rejects immediately with HTTP 429 Too Many Requests and Retry-After header.
	//           *Fail-Fast Principle*: Eliminates abusive traffic BEFORE doing body read, hashing & crypto checks!
	//   Step 3: HMACMiddleware checks X-Novaris-Timestamp & X-Novaris-Signature (rejects if tampered/stale -> 401).
	//           Restores r.Body for downstream consumption.
	//   Step 4: paymentsHandler processes payment settlement logic (reaches here only if all three layers pass).
	mux.HandleFunc("/v1/payments",
		auth.JWTMiddleware(
			ratelimit.RateLimitMiddleware(bucketStore)(
				auth.HMACMiddleware(paymentsHandler),
			),
		),
	)

	// Wrap router with logging middleware for latency & request tracing
	handlerWithMiddleware := loggingMiddleware(mux)

	// Configure server with production-grade timeouts to prevent Slowloris attacks
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      handlerWithMiddleware,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Channel for OS termination signals (SIGINT, SIGTERM)
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	// Launch HTTP listener in a background goroutine
	go func() {
		log.Printf("=================================================================")
		log.Printf(" ⚡ NOVARIS Gateway Service v%s (Day 4 - Rate Limiter Defense)", cfg.ServiceVersion)
		log.Printf(" Listening on port: :%s", cfg.Port)
		log.Printf(" Public:    http://localhost:%s/healthz", cfg.Port)
		log.Printf(" Protected: http://localhost:%s/v1/payments (JWT + RateLimit + HMAC)", cfg.Port)
		log.Printf(" Rate Limit: Capacity=%.0f tokens, Refill=%.1f tokens/sec per merchant",
			ratelimit.DefaultBucketCapacity, ratelimit.DefaultRefillRate)
		log.Printf("=================================================================")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("FATAL: Server crashed: %v", err)
		}
	}()

	// Block until OS termination signal arrives
	<-stopChan
	log.Println("\n[SHUTDOWN] OS termination signal received. Draining connections...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("FATAL: Forced shutdown: %v", err)
	}

	log.Println("[SHUTDOWN] NOVARIS Gateway stopped cleanly.")
}
