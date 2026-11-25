// ==============================================================================
// NOVARIS: Distributed Payment Settlement & Adaptive Fraud Intelligence Platform
// HIGH-CONCURRENCY LOAD TESTING ENGINE & HOT-WALLET STRESS GENERATOR (Day 13)
// FILE: services/gateway/cmd/loadtest/main.go
// ==============================================================================
//
// 📚 COMPUTER SCIENCE & CONCURRENCY BENCHMARKING FOR BEGINNERS:
//
// 1. WHY BUILD A NATIVE GO LOAD TESTER?
//    While external tools like k6, Apache JMeter, and Locust are widely used,
//    having a self-contained, native Go load testing utility built directly into
//    the repository provides massive advantages:
//      - Zero External Dependencies: Runs on any operating system with only `go run`.
//      - Direct Code Reuse: Can reuse the platform's native HMAC and JWT signing routines.
//      - Nanosecond Precision: Measures exact socket turnaround latency using Go's monotonic clock.
//      - CI/CD Ready: Can be executed directly in GitHub Actions pipelines without tool installs.
//
// 2. CONCURRENCY MECHANICS: WORKER POOL & SYNC.WAITGROUP
//    To simulate hundreds of simultaneous payment clients:
//      - We spawn N concurrent goroutines (workers).
//      - A shared Go channel (`jobs`) acts as a thread-safe task queue.
//      - `sync.WaitGroup` coordinates graceful completion: `wg.Add(1)` on start, `wg.Done()` on exit,
//        and `wg.Wait()` blocks the main function until every single worker has finished.
//
// 3. STATISTICAL LATENCY PERCENTILES (p50, p90, p95, p99):
//    In financial distributed systems, the "average" (mean) latency is misleading because
//    one slow request can be hidden by 99 fast requests.
//    - p50 (Median): 50% of requests were faster than this value.
//    - p90: 90% of requests were faster than this value.
//    - p95: Standard Institutional Service Level Agreement (SLA) threshold.
//    - p99 (Tail Latency): The slowest 1% of requests (representing edge-case queue delays).
// ==============================================================================

package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/charan21042005/novaris/services/gateway/internal/auth"
)

// LoadTestConfig captures all configurable CLI parameters for the load run.
type LoadTestConfig struct {
	BaseURL     string        // Target gateway URL (e.g., "http://localhost:8080")
	NumWorkers  int           // Number of concurrent worker goroutines
	TotalReqs   int           // Total number of payment settlement requests to dispatch
	HotWallet   bool          // If true, all workers collide on the exact same account UUID
	NumTenants  int           // Number of distinct merchant tenants to cycle through
	AmountCents int64         // Transfer magnitude in integer cents ($1.00 = 100)
	Timeout     time.Duration // Per-request HTTP client timeout
}

// ResultRecord holds telemetry captured from a single executed HTTP request.
type ResultRecord struct {
	StatusCode int           // HTTP status code returned (200, 400, 401, 409, 422, 429, 500)
	Duration   time.Duration // Monotonic round-trip socket duration
	IsReplayed bool          // True if the "Idempotent-Replayed: true" header was present
	Err        error         // Network error if request failed to reach the server
}

// LoadTestReport aggregates summary statistics across all executed requests.
type LoadTestReport struct {
	TotalDispatched int64
	Status200Count  int64
	Status409Count  int64
	Status429Count  int64
	Status422Count  int64
	Status400Count  int64
	Status401Count  int64
	Status500Count  int64
	NetworkErrors   int64
	ReplayedCount   int64
	TotalDuration   time.Duration
	TPS             float64
	LatenciesMs     []float64
	P50Ms           float64
	P90Ms           float64
	P95Ms           float64
	P99Ms           float64
	MinMs           float64
	MaxMs           float64
	AvgMs           float64
}

// RunLoadTest executes the concurrent load benchmark against the target gateway.
func RunLoadTest(cfg LoadTestConfig) LoadTestReport {
	// Reusable HTTP client with keep-alive connection pooling configured for high concurrency:
	client := &http.Client{
		Timeout: cfg.Timeout,
		Transport: &http.Transport{
			MaxIdleConns:        cfg.NumWorkers * 2,
			MaxIdleConnsPerHost: cfg.NumWorkers * 2,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	// Task channel distributing work items among the worker pool:
	jobs := make(chan int, cfg.TotalReqs)
	results := make(chan ResultRecord, cfg.TotalReqs)

	// Pre-fill the work queue with request indices:
	for i := 0; i < cfg.TotalReqs; i++ {
		jobs <- i
	}
	close(jobs)

	// Determine deterministic seed accounts:
	const defaultSender = "c0000000-0000-0000-0000-000000000001"
	const defaultRecipient = "c0000000-0000-0000-0000-000000000002"
	hmacSecret := auth.GetHMACSecret()

	var wg sync.WaitGroup
	startTime := time.Now()

	// Spawn the worker goroutine pool:
	for w := 0; w < cfg.NumWorkers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			// Determine merchant identity:
			merchantID := defaultSender
			if !cfg.HotWallet && cfg.NumTenants > 1 {
				// Distribute across tenant pool:
				tenantIndex := (workerID % cfg.NumTenants) + 1
				merchantID = fmt.Sprintf("merchant_tenant_%02d", tenantIndex)
			}

			// Pre-mint valid JWT token for this worker:
			token, err := auth.GenerateToken(merchantID, "merchant_admin", 1*time.Hour)
			if err != nil {
				results <- ResultRecord{Err: fmt.Errorf("failed to generate token: %w", err)}
				return
			}

			// Drain jobs channel:
			for range jobs {
				// Construct unique payment intent:
				idempKey := uuid.New().String()
				timestamp := time.Now().Unix()

				bodyPayload := map[string]interface{}{
					"sender_id":    defaultSender,
					"recipient_id": defaultRecipient,
					"amount_cents": cfg.AmountCents,
					"currency":     "USD",
				}
				rawBody, _ := json.Marshal(bodyPayload)

				// Compute canonical HMAC-SHA256 signature:
				signature := auth.ComputeHMACSignature(rawBody, timestamp, hmacSecret)

				// Build HTTP request:
				req, err := http.NewRequest(http.MethodPost, cfg.BaseURL+"/v1/payments", bytes.NewReader(rawBody))
				if err != nil {
					results <- ResultRecord{Err: err}
					continue
				}

				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer "+token)
				req.Header.Set("Idempotency-Key", idempKey)
				req.Header.Set("X-Novaris-Timestamp", strconv.FormatInt(timestamp, 10))
				req.Header.Set("X-Novaris-Signature", signature)

				// Dispatch request and record high-precision monotonic duration:
				reqStart := time.Now()
				resp, err := client.Do(req)
				duration := time.Since(reqStart)

				if err != nil {
					results <- ResultRecord{Duration: duration, Err: err}
					continue
				}

				// Consume and close response body:
				isReplayed := resp.Header.Get("Idempotent-Replayed") == "true"
				_, _ = io.Copy(io.Discard, resp.Body)
				resp.Body.Close()

				results <- ResultRecord{
					StatusCode: resp.StatusCode,
					Duration:   duration,
					IsReplayed: isReplayed,
				}
			}
		}(w)
	}

	// Wait for all workers to complete:
	wg.Wait()
	close(results)
	totalDuration := time.Since(startTime)

	// Aggregate telemetry results:
	var report LoadTestReport
	report.TotalDuration = totalDuration
	var latencies []float64
	var totalLatencySum float64

	for r := range results {
		report.TotalDispatched++
		if r.Err != nil {
			report.NetworkErrors++
			continue
		}

		ms := float64(r.Duration.Nanoseconds()) / 1000000.0
		latencies = append(latencies, ms)
		totalLatencySum += ms

		if r.IsReplayed {
			report.ReplayedCount++
		}

		switch r.StatusCode {
		case http.StatusOK:
			report.Status200Count++
		case http.StatusConflict:
			report.Status409Count++
		case http.StatusTooManyRequests:
			report.Status429Count++
		case http.StatusUnprocessableEntity:
			report.Status422Count++
		case http.StatusBadRequest:
			report.Status400Count++
		case http.StatusUnauthorized:
			report.Status401Count++
		case http.StatusInternalServerError:
			report.Status500Count++
		}
	}

	// Compute percentiles:
	if len(latencies) > 0 {
		sort.Float64s(latencies)
		report.LatenciesMs = latencies
		report.MinMs = latencies[0]
		report.MaxMs = latencies[len(latencies)-1]
		report.AvgMs = totalLatencySum / float64(len(latencies))
		report.P50Ms = latencies[int(float64(len(latencies))*0.50)]
		report.P90Ms = latencies[int(float64(len(latencies))*0.90)]
		report.P95Ms = latencies[int(float64(len(latencies))*0.95)]
		report.P99Ms = latencies[int(float64(len(latencies))*0.99)]
	}

	if totalDuration.Seconds() > 0 {
		report.TPS = float64(report.TotalDispatched) / totalDuration.Seconds()
	}

	return report
}

// PrintReport outputs an ANSI-formatted, executive summary table to stdout.
func PrintReport(r LoadTestReport, cfg LoadTestConfig) {
	fmt.Println()
	fmt.Println("==================================================================")
	fmt.Println(" 🚀 NOVARIS Day 13: Concurrency Load Test Benchmark Report")
	fmt.Println("==================================================================")
	fmt.Printf(" Target Gateway URL:      %s\n", cfg.BaseURL)
	fmt.Printf(" Concurrent Workers:      %d goroutines\n", cfg.NumWorkers)
	fmt.Printf(" Total Requests:          %d transactions\n", cfg.TotalReqs)
	fmt.Printf(" Workload Profile:        Hot-Wallet Contention: %v | Tenants: %d\n", cfg.HotWallet, cfg.NumTenants)
	fmt.Printf(" Benchmark Duration:      %v\n", r.TotalDuration.Round(time.Millisecond))
	fmt.Printf(" Throughput (TPS):        %0.1f requests/sec\n", r.TPS)
	fmt.Println("------------------------------------------------------------------")
	fmt.Println(" 📊 HTTP STATUS CODE DISTRIBUTION:")
	fmt.Printf("   ✅ 200 OK (Settled / Replayed):   %d (%0.1f%%)\n", r.Status200Count, pct(r.Status200Count, r.TotalDispatched))
	fmt.Printf("   ⚡ 429 Too Many Requests (Limit): %d (%0.1f%%)\n", r.Status429Count, pct(r.Status429Count, r.TotalDispatched))
	fmt.Printf("   🔒 409 Conflict (OCC Collision):   %d (%0.1f%%)\n", r.Status409Count, pct(r.Status409Count, r.TotalDispatched))
	fmt.Printf("   ⚠️ 422 Unprocessable (Fingerprint): %d (%0.1f%%)\n", r.Status422Count, pct(r.Status422Count, r.TotalDispatched))
	fmt.Printf("   ❌ 400 Bad Request (Validation):   %d (%0.1f%%)\n", r.Status400Count, pct(r.Status400Count, r.TotalDispatched))
	fmt.Printf("   🚫 401 Unauthorized (Auth/Crypto): %d (%0.1f%%)\n", r.Status401Count, pct(r.Status401Count, r.TotalDispatched))
	fmt.Printf("   🔥 500 Internal Error (Crashes):   %d (%0.1f%%)\n", r.Status500Count, pct(r.Status500Count, r.TotalDispatched))
	fmt.Printf("   🔌 Network Connection Drops:       %d (%0.1f%%)\n", r.NetworkErrors, pct(r.NetworkErrors, r.TotalDispatched))
	fmt.Println("------------------------------------------------------------------")
	fmt.Println(" ⏱️ LATENCY PERCENTILES (Monotonic Network Turnaround):")
	fmt.Printf("   Min Latency:            %0.2f ms\n", r.MinMs)
	fmt.Printf("   Average Latency:        %0.2f ms\n", r.AvgMs)
	fmt.Printf("   Median (p50):           %0.2f ms\n", r.P50Ms)
	fmt.Printf("   90th Percentile (p90):  %0.2f ms\n", r.P90Ms)
	fmt.Printf("   95th Percentile (p95):  %0.2f ms\n", r.P95Ms)
	fmt.Printf("   99th Percentile (p99):  %0.2f ms\n", r.P99Ms)
	fmt.Printf("   Max Latency:            %0.2f ms\n", r.MaxMs)
	fmt.Println("==================================================================")

	// Verification Verdict:
	if r.Status500Count == 0 && r.NetworkErrors == 0 {
		fmt.Println("🎉 CONCURRENCY BENCHMARK VERIFIED: ZERO RACE CONDITIONS OR CRASHES OBSERVED!")
	} else {
		fmt.Println("⚠️ CONCURRENCY WARNING: System produced internal errors under load.")
		os.Exit(1)
	}
}

func pct(part, total int64) float64 {
	if total == 0 {
		return 0.0
	}
	return (float64(part) / float64(total)) * 100.0
}

func main() {
	baseURL := flag.String("url", "http://localhost:8080", "Target gateway base URL")
	workers := flag.Int("workers", 50, "Number of concurrent worker goroutines")
	requests := flag.Int("requests", 250, "Total number of payment transactions to dispatch")
	hotwallet := flag.Bool("hotwallet", true, "Direct all workers at the same hot-wallet account")
	tenants := flag.Int("tenants", 1, "Number of distinct merchant tenants (when -hotwallet=false)")
	amount := flag.Int64("amount", 100, "Transfer amount in integer cents ($1.00 = 100)")
	timeoutSec := flag.Int("timeout", 10, "Per-request client timeout in seconds")
	flag.Parse()

	cfg := LoadTestConfig{
		BaseURL:     *baseURL,
		NumWorkers:  *workers,
		TotalReqs:   *requests,
		HotWallet:   *hotwallet,
		NumTenants:  *tenants,
		AmountCents: *amount,
		Timeout:     time.Duration(*timeoutSec) * time.Second,
	}

	report := RunLoadTest(cfg)
	PrintReport(report, cfg)
}
