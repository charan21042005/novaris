// ==============================================================================
// NOVARIS: Distributed Payment Settlement & Adaptive Fraud Intelligence Platform
// MODULE 1: API Gateway Service Entry Point (Day 3 - JWT + HMAC Layered Security)
// ==============================================================================
//
// Day 3 Evolution:
// 1. /healthz remains public for unauthenticated Kubernetes/Docker health probing.
// 2. / (root) provides public service discovery.
// 3. /v1/payments is now protected by a TWO-LAYER DEFENSE IN DEPTH PIPELINE:
//    - Layer 1 (Outermost): auth.JWTMiddleware
//      Answers "WHO ARE YOU?": Validates the Bearer token's cryptographic signature,
//      checks token expiration, and injects merchant identity claims into context.
//    - Layer 2 (Inner): auth.HMACMiddleware
//      Answers "HAS THIS BEEN ALTERED OR REPLAYED?": Reads the raw request body bytes,
//      verifies X-Novaris-Timestamp within a ±300s window (replay defense),
//      recomputes the HMAC-SHA256 signature using the shared secret, and compares
//      using constant-time hmac.Equal (timing attack defense). Restores r.Body.
//    - Layer 3 (Core): paymentsHandler
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
			Day:         3,
			Phase:       "Phase 1 - Design & Ledger",
			Endpoints: map[string]string{
				"health_check":       "/healthz (Public Liveness Probe)",
				"payments_protected": "/v1/payments (Protected - Requires JWT Bearer + HMAC-SHA256 Signature)",
				"transaction_flow":   "/v1/transactions (Coming in Phase 1: Day 4 Rate Limiting & Day 5 Ledger)",
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
		Message:     "Payment request verified successfully: JWT identity confirmed and HMAC body integrity preserved.",
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
		ServiceVersion: "0.3.0-day3",
		Environment:    "development",
	}

	// Create a new ServeMux (HTTP request router)
	mux := http.NewServeMux()

	// 1. Public Routes (No authentication required)
	mux.HandleFunc("/", rootHandler(cfg))
	mux.HandleFunc("/healthz", healthzHandler(cfg))

	// 2. Protected Routes (Stacked Middleware: JWTMiddleware -> HMACMiddleware -> paymentsHandler)
	// Execution flow:
	//   Step 1: JWTMiddleware checks Authorization Bearer token (rejects if missing/forged/expired -> 401)
	//   Step 2: HMACMiddleware checks X-Novaris-Timestamp & X-Novaris-Signature (rejects if tampered/stale -> 401)
	//   Step 3: paymentsHandler processes payment settlement logic (reaches here only if both pass)
	mux.HandleFunc("/v1/payments", auth.JWTMiddleware(auth.HMACMiddleware(paymentsHandler)))

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
		log.Printf(" ⚡ NOVARIS Gateway Service v%s (Day 3 - JWT + HMAC Defense)", cfg.ServiceVersion)
		log.Printf(" Listening on port: :%s", cfg.Port)
		log.Printf(" Public:    http://localhost:%s/healthz", cfg.Port)
		log.Printf(" Protected: http://localhost:%s/v1/payments (JWT + HMAC)", cfg.Port)
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
