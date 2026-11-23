// ==============================================================================
// NOVARIS: Distributed Payment Settlement & Adaptive Fraud Intelligence Platform
// MODULE 1: API Gateway Service Entry Point
// ==============================================================================
//
// Purpose of the API Gateway:
// The Gateway is the single public "front door" for external merchants and clients.
// On Day 1, this service establishes:
//   1. A clean, idiomatic Go HTTP server using the standard library (net/http).
//   2. Structured request logging middleware (method, path, latency).
//   3. A standard Kubernetes-compatible `/healthz` liveness/readiness probe.
//   4. Graceful shutdown to handle OS termination signals (SIGINT, SIGTERM) cleanly.
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
)

// Config holds runtime configuration parameters for the gateway.
// In later phases, these will be populated from environment variables.
type Config struct {
	Port            string
	ServiceName     string
	ServiceVersion string
	Environment     string
}

// HealthResponse defines the structured JSON payload returned by the /healthz probe.
// This matches standard container orchestration (Kubernetes) health expectations.
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

// loggingMiddleware intercepts incoming HTTP requests, records the processing time,
// and outputs a structured log line. In production, this provides request observability.
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Call the next handler in the chain
		next.ServeHTTP(w, r)

		// Log method, path, remote client IP, and elapsed duration
		duration := time.Since(start)
		log.Printf("[HTTP] %s %s from %s in %v", r.Method, r.URL.Path, r.RemoteAddr, duration)
	})
}

// healthzHandler responds with HTTP 200 and system health metadata.
// Used by Docker and Kubernetes to verify the service is alive and healthy.
func healthzHandler(cfg Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only allow GET requests on the health endpoint
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

		if err := json.NewEncoder(w).Encode(payload); err != nil {
			log.Printf("ERROR: Failed to serialize health check payload: %v", err)
		}
	}
}

// rootHandler provides introductory information and lists available routes.
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
				"health_check":     "/healthz",
				"transaction_flow": "/v1/transactions (Coming in Phase 1: Day 2-4)",
			},
		}

		_ = json.NewEncoder(w).Encode(payload)
	}
}

func main() {
	// Initialize default configuration
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	cfg := Config{
		Port:            port,
		ServiceName:     "novaris-gateway",
		ServiceVersion: "1.0.0",
		Environment:     "development",
	}

	// Create a new ServeMux (HTTP request router)
	mux := http.NewServeMux()

	// Register route handlers
	mux.HandleFunc("/", rootHandler(cfg))
	mux.HandleFunc("/healthz", healthzHandler(cfg))

	// Wrap router with logging middleware
	handlerWithMiddleware := loggingMiddleware(mux)

	// Configure the HTTP server with production-grade timeouts to prevent slowloris attacks
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      handlerWithMiddleware,
		ReadTimeout:  10 * time.Second, // Max duration reading the entire request
		WriteTimeout: 10 * time.Second, // Max duration writing the response
		IdleTimeout:  60 * time.Second, // Max time to keep idle keep-alive connections
	}

	// Channel to listen for OS termination signals (CTRL+C, SIGTERM from Docker/K8s)
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	// Start server in a background goroutine so it does not block the main signal listener
	go func() {
		log.Printf("=======================================================")
		log.Printf(" ⚡ NOVARIS Gateway Service v%s started", cfg.ServiceVersion)
		log.Printf(" Listening on port: :%s", cfg.Port)
		log.Printf(" Healthcheck: http://localhost:%s/healthz", cfg.Port)
		log.Printf("=======================================================")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("FATAL: Server crashed: %v", err)
		}
	}()

	// Block here until a termination signal is received
	<-stopChan
	log.Println("\n[SHUTDOWN] OS termination signal received. Initiating graceful shutdown...")

	// Create a shutdown context with a 5-second deadline
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("FATAL: Server forced to shutdown: %v", err)
	}

	log.Println("[SHUTDOWN] NOVARIS Gateway stopped cleanly.")
}
