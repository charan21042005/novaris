// ==============================================================================
// NOVARIS: Distributed Payment Settlement & Adaptive Fraud Intelligence Platform
// MODULE 1: API Gateway Service Entry Point (Day 2 - JWT Protected)
// ==============================================================================
//
// Day 2 Evolution:
// 1. /healthz remains public for unauthenticated Kubernetes/Docker health probing.
// 2. / (root) provides public service discovery.
// 3. /v1/payments is now a PROTECTED ROUTE wrapped with auth.JWTMiddleware:
//    - Rejects unauthenticated requests with HTTP 401 Unauthorized.
//    - Rejects expired or cryptographically tampered tokens.
//    - Injects verified claims (merchant_id, role) into request context.
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

// InfoResponse provides public service metadata for developers querying the root route.
type InfoResponse struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Status      string            `json:"status"`
	Version     string            `json:"version"`
	Endpoints   map[string]string `json:"endpoints"`
}

// PaymentResponse defines the response returned by the protected /v1/payments route.
type PaymentResponse struct {
	Status     string    `json:"status"`
	MerchantID string    `json:"merchant_id"`
	Role       string    `json:"role"`
	Message    string    `json:"message"`
	Timestamp  time.Time `json:"timestamp"`
}

// loggingMiddleware intercepts incoming HTTP requests, records the processing time,
// and outputs a structured log line for full observability.
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
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
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
			Endpoints: map[string]string{
				"health_check":       "/healthz (Public Liveness Probe)",
				"payments_protected": "/v1/payments (Protected - Requires 'Authorization: Bearer <token>')",
				"transaction_flow":   "/v1/transactions (Coming in Phase 1: Day 3 HMAC & Day 4 Rate Limiting)",
			},
		}

		_ = json.NewEncoder(w).Encode(payload)
	}
}

// paymentsHandler is a PROTECTED route stub.
// It retrieves the pre-verified JWT claims from the request context and returns an authorized response.
// Real double-entry ledger logic and Kafka streaming will be hooked into this pipeline in later days.
func paymentsHandler(w http.ResponseWriter, r *http.Request) {
	// Retrieve pre-verified claims that were attached to the context by auth.JWTMiddleware
	claims, ok := auth.GetClaimsFromContext(r.Context())
	if !ok {
		// Defensive programming: this should never happen because the middleware guarantees presence
		http.Error(w, "Internal Server Error: Missing claims in context", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	payload := PaymentResponse{
		Status:     "authorized",
		MerchantID: claims.MerchantID,
		Role:       claims.Role,
		Message:    "Access granted to protected payments route. Identity cryptographically verified via JWT.",
		Timestamp:  time.Now().UTC(),
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
		ServiceVersion: "1.0.0",
		Environment:    "development",
	}

	// Create a new ServeMux (HTTP request router)
	mux := http.NewServeMux()

	// 1. Public Routes (No authentication required)
	mux.HandleFunc("/", rootHandler(cfg))
	mux.HandleFunc("/healthz", healthzHandler(cfg))

	// 2. Protected Routes (Wrapped with auth.JWTMiddleware)
	// Any request arriving here without a valid Bearer token is rejected with HTTP 401
	mux.HandleFunc("/v1/payments", auth.JWTMiddleware(paymentsHandler))

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
		log.Printf("=======================================================")
		log.Printf(" ⚡ NOVARIS Gateway Service v%s (Day 2 - JWT Security)", cfg.ServiceVersion)
		log.Printf(" Listening on port: :%s", cfg.Port)
		log.Printf(" Public:    http://localhost:%s/healthz", cfg.Port)
		log.Printf(" Protected: http://localhost:%s/v1/payments", cfg.Port)
		log.Printf("=======================================================")

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
