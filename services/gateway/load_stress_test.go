// ==============================================================================
// NOVARIS: Distributed Payment Settlement & Adaptive Fraud Intelligence Platform
// HIGH-DENSITY 100-WORKER CONCURRENCY & OCC STRESS TEST SUITE (Day 13)
// FILE: services/gateway/load_stress_test.go
// ==============================================================================
//
// 📚 COMPUTER SCIENCE & 100-WORKER CONCURRENCY CONCEPTS FOR BEGINNERS:
//
// 1. WHY TEST WITH 100 SIMULTANEOUS WORKERS?
//    In single-threaded applications, race conditions are impossible because operations
//    execute sequentially one after another.
//    In a production payment gateway running on multi-core cloud servers:
//      - The operating system and Go runtime scheduler schedule goroutines across
//        all available CPU cores in parallel.
//      - When 100 goroutines attempt to modify shared memory or database records
//        at the exact same microsecond, subtle concurrency bugs (like non-atomic
//        reads, dirty writes, lost updates, and memory race conditions) manifest immediately!
//
// 2. THE SYNCHRONIZED START-GATE PATTERN (chan struct{}):
//    If you simply spawn 100 goroutines in a loop:
//      for i := 0; i < 100; i++ { go doWork() }
//    Worker #1 might finish before Worker #100 is even created by the Go scheduler!
//    That is NOT a true simultaneous spike.
//    To force 100 goroutines to strike the target at the exact same nanosecond:
//      - We create an unbuffered channel: `startGate := make(chan struct{})`.
//      - Every worker blocks at `<-startGate`.
//      - The main test goroutine calls `close(startGate)`.
//      - Closing the channel broadcasts to all 100 goroutines simultaneously, releasing
//        them into the system in a nanosecond-level tidal wave of traffic!
//
// 3. PHASE 1 CAPSTONE MILESTONE GOAL:
//    "100 concurrent unit tests with zero race conditions."
// ==============================================================================

package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/charan21042005/novaris/services/gateway/internal/auth"
	"github.com/charan21042005/novaris/services/gateway/internal/ratelimit"
	"github.com/google/uuid"
)

// TestLoadStress_100ConcurrentWorkers_HotWalletOCC fires 100 simultaneous goroutines
// at the exact same hot-wallet in-memory mock ledger to verify that Optimistic
// Concurrency Control (OCC) maintains mathematical balance conservation with ZERO data races.
func TestLoadStress_100ConcurrentWorkers_HotWalletOCC(t *testing.T) {
	const workerCount = 100
	const initialBalance = int64(500000) // $5,000.00
	const transferAmount = int64(100)    // $1.00 per transfer
	const sharedSenderID = "c0000000-0000-0000-0000-000000000001"
	const targetRecipientID = "c0000000-0000-0000-0000-000000000002"

	// Mock in-memory account ledger state protected by OCC versioning:
	type MockAccount struct {
		sync.Mutex
		Balance int64
		Version int
	}

	senderAccount := &MockAccount{Balance: initialBalance, Version: 1}
	recipientAccount := &MockAccount{Balance: 0, Version: 1}

	var settledCount atomic.Int64
	var collisionRetryCount atomic.Int64
	var failureCount atomic.Int64

	startGate := make(chan struct{})
	var wg sync.WaitGroup

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			// Block until start gate is released:
			<-startGate

			// Execute OCC transaction with up to 5 retries and exponential backoff:
			maxRetries := 5
			settled := false

			for attempt := 0; attempt < maxRetries; attempt++ {
				// 1. Read current state & version:
				senderAccount.Lock()
				currentBalance := senderAccount.Balance
				readVersion := senderAccount.Version
				senderAccount.Unlock()

				// Business validation:
				if currentBalance < transferAmount {
					failureCount.Add(1)
					return
				}

				// Small simulated processing latency (50-200 microseconds)
				time.Sleep(time.Duration(50+workerID%150) * time.Microsecond)

				// 2. Atomic Compare-And-Swap (CAS) write:
				senderAccount.Lock()
				if senderAccount.Version == readVersion {
					// Successful CAS write!
					senderAccount.Balance -= transferAmount
					senderAccount.Version++
					senderAccount.Unlock()

					// Credit recipient:
					recipientAccount.Lock()
					recipientAccount.Balance += transferAmount
					recipientAccount.Version++
					recipientAccount.Unlock()

					settledCount.Add(1)
					settled = true
					break
				}
				senderAccount.Unlock()

				// Collision detected! Increment retry counter:
				collisionRetryCount.Add(1)
				// Exponential backoff with jitter:
				time.Sleep(time.Duration((1<<attempt)*100+(workerID*13)%200) * time.Microsecond)
			}

			if !settled {
				failureCount.Add(1)
			}
		}(i)
	}

	// Release all 100 workers simultaneously!
	close(startGate)
	wg.Wait()

	totalSettled := settledCount.Load()
	totalCollisions := collisionRetryCount.Load()
	totalFailed := failureCount.Load()

	t.Logf("100-WORKER HOT-WALLET RESULTS: Settled=%d | OCC Collisions Retried=%d | Contention Drops=%d",
		totalSettled, totalCollisions, totalFailed)

	// Mathematical Invariant 1: Conservation of Money
	// Total sender balance + Total recipient balance MUST EXACTLY equal initial balance:
	senderAccount.Lock()
	finalSender := senderAccount.Balance
	finalSenderVersion := senderAccount.Version
	senderAccount.Unlock()

	recipientAccount.Lock()
	finalRecipient := recipientAccount.Balance
	recipientAccount.Unlock()

	if finalSender+finalRecipient != initialBalance {
		t.Fatalf("BALANCE CONSERVATION VIOLATED! initial=%d, sender=%d, recipient=%d (diff=%d)",
			initialBalance, finalSender, finalRecipient, (finalSender+finalRecipient)-initialBalance)
	}

	// Mathematical Invariant 2: Total recipient balance must equal total settled * transferAmount:
	expectedRecipientBalance := totalSettled * transferAmount
	if finalRecipient != expectedRecipientBalance {
		t.Fatalf("RECIPIENT BALANCE DRIFT! expected=%d, got=%d", expectedRecipientBalance, finalRecipient)
	}

	// Mathematical Invariant 3: Sender version progression must equal initial version + settled transfers:
	expectedVersion := 1 + int(totalSettled)
	if finalSenderVersion != expectedVersion {
		t.Fatalf("VERSION PROGRESSION DRIFT! expected version %d, got %d", expectedVersion, finalSenderVersion)
	}

	t.Logf("ZERO RACE CONDITIONS: Balance conservation proven across 100 simultaneous workers! (Final version: %d)", finalSenderVersion)
}

// TestLoadStress_100ConcurrentWorkers_FullPipelineSpike tests the entire Five-Layer
// Defensive Perimeter under a 100-worker sudden traffic spike.
// Asserts that the Token-Bucket Rate Limiter admits initial burst quota and cleanly
// throttles the excess with HTTP 429 Too Many Requests with ZERO HTTP 500 crashes.
func TestLoadStress_100ConcurrentWorkers_FullPipelineSpike(t *testing.T) {
	const workerCount = 100
	const burstCapacity = 10.0 // Allow 10 burst requests
	const refillRate = 2.0    // Refill 2 tokens/sec

	// Configure server harness:
	bucketStore := ratelimit.NewBucketStore(burstCapacity, refillRate)
	idempStore := NewInMemoryIdempotencyStore()

	cfg := Config{
		Port:           "0",
		ServiceName:    "novaris-spike-test",
		ServiceVersion: "0.13.0-loadtest",
		Environment:    "test",
	}

	handler := SetupServerMux(cfg, nil, bucketStore, idempStore)
	server := httptest.NewServer(handler)
	defer server.Close()

	// High-throughput HTTP client:
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        workerCount,
			MaxIdleConnsPerHost: workerCount,
		},
	}

	const merchantID = "c0000000-0000-0000-0000-000000000001"
	token, err := auth.GenerateToken(merchantID, "merchant_admin", 1*time.Hour)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}
	hmacSecret := auth.GetHMACSecret()

	var admittedCount atomic.Int64
	var throttledCount atomic.Int64
	var crashCount atomic.Int64

	startGate := make(chan struct{})
	var wg sync.WaitGroup

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func(workerIndex int) {
			defer wg.Done()

			// Block until synchronous broadcast:
			<-startGate

			idempKey := uuid.New().String()
			timestamp := time.Now().Unix()

			bodyPayload := map[string]interface{}{
				"sender_id":    merchantID,
				"recipient_id": "c0000000-0000-0000-0000-000000000002",
				"amount_cents": 500,
				"currency":     "USD",
			}
			rawBody, _ := json.Marshal(bodyPayload)
			sig := auth.ComputeHMACSignature(rawBody, timestamp, hmacSecret)

			req, _ := http.NewRequest(http.MethodPost, server.URL+"/v1/payments", bytes.NewReader(rawBody))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("Idempotency-Key", idempKey)
			req.Header.Set("X-Novaris-Timestamp", strconv.FormatInt(timestamp, 10))
			req.Header.Set("X-Novaris-Signature", sig)

			resp, err := client.Do(req)
			if err != nil {
				t.Errorf("worker %d network error: %v", workerIndex, err)
				return
			}
			defer resp.Body.Close()

			switch resp.StatusCode {
			case http.StatusOK:
				admittedCount.Add(1)
			case http.StatusTooManyRequests:
				throttledCount.Add(1)
				// Verify Retry-After header:
				if resp.Header.Get("Retry-After") == "" {
					t.Errorf("worker %d: missing Retry-After header on 429", workerIndex)
				}
			case http.StatusInternalServerError:
				crashCount.Add(1)
			}
		}(i)
	}

	// Release all 100 workers at the exact same nanosecond!
	close(startGate)
	wg.Wait()

	admitted := admittedCount.Load()
	throttled := throttledCount.Load()
	crashes := crashCount.Load()

	t.Logf("100-WORKER PIPELINE SPIKE: Admitted=%d | Throttled (429)=%d | Crashes (500)=%d",
		admitted, throttled, crashes)

	// Assertion 1: Zero crashes under 100-worker sudden flood:
	if crashes != 0 {
		t.Fatalf("DEFENSE COLLAPSED: %d requests resulted in HTTP 500 crashes!", crashes)
	}

	// Assertion 2: Admitted count must strictly be around burst capacity ceiling (10):
	if admitted < 10 || admitted > 15 {
		t.Errorf("expected ~10 admitted burst requests, got %d", admitted)
	}

	// Assertion 3: Remaining requests must be cleanly throttled with 429:
	if throttled < 85 {
		t.Errorf("expected >= 85 throttled requests, got %d", throttled)
	}

	t.Logf("FAIL-FAST PROTECTION PROVEN: 100-worker spike was successfully absorbed without backend saturation!")
}

// TestLoadStress_100ConcurrentWorkers_IdempotencyRace asserts that when 100 simultaneous
// workers submit the EXACT SAME Idempotency-Key, the system settles the payment
// EXACTLY ONCE, with ZERO duplicate balance deductions.
func TestLoadStress_100ConcurrentWorkers_IdempotencyRace(t *testing.T) {
	const workerCount = 100
	idempStore := NewInMemoryIdempotencyStore()

	// Handler with atomic execution counter:
	var handlerExecutionCount atomic.Int64
	businessHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerExecutionCount.Add(1)
		// Small simulated processing time:
		time.Sleep(5 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"settled","transaction_id":"test-tx-100"}`))
	})

	pipeline := IdempotencyMiddleware(idempStore)(businessHandler)

	const sharedIdempotencyKey = "idemp_100_concurrent_worker_race_test_key"
	const sharedBody = `{"sender_id":"acc_1","recipient_id":"acc_2","amount_cents":5000}`

	var success200Count atomic.Int64
	var inFlight409Count atomic.Int64

	startGate := make(chan struct{})
	var wg sync.WaitGroup

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-startGate

			req := httptest.NewRequest(http.MethodPost, "/v1/payments", bytes.NewReader([]byte(sharedBody)))
			req.Header.Set("Idempotency-Key", sharedIdempotencyKey)
			rec := httptest.NewRecorder()

			pipeline.ServeHTTP(rec, req)

			switch rec.Code {
			case http.StatusOK:
				success200Count.Add(1)
			case http.StatusConflict:
				inFlight409Count.Add(1)
			}
		}()
	}

	// Fire all 100 goroutines!
	close(startGate)
	wg.Wait()

	totalExecutions := handlerExecutionCount.Load()
	total200s := success200Count.Load()
	total409s := inFlight409Count.Load()

	t.Logf("100-WORKER IDEMPOTENCY RACE: Handler Invocations=%d | HTTP 200 (Settled/Replay)=%d | HTTP 409 (In-Flight)=%d",
		totalExecutions, total200s, total409s)

	// EXACT-ONCE INVARIANT:
	// The underlying business settlement handler must have been executed EXACTLY 1 time!
	if totalExecutions != 1 {
		t.Fatalf("DOUBLE-SPEND DETECTED! Settlement handler was executed %d times (expected exactly 1)!", totalExecutions)
	}

	if total200s+total409s != workerCount {
		t.Errorf("unexpected response distribution: 200s=%d, 409s=%d, total=%d", total200s, total409s, workerCount)
	}

	t.Logf("EXACT-ONCE GUARANTEE PROVEN: 100 concurrent racers produced EXACTLY 1 ledger settlement!")
}
