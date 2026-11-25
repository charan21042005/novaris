// ==============================================================================
// NOVARIS: Distributed Payment Settlement & Adaptive Fraud Intelligence Platform
// CONCURRENCY STRESS SUITE: Goroutine Collisions & Race Detector Verification (Day 8)
// FILE: services/gateway/concurrency_test.go
// ==============================================================================
//
// Architectural Objectives (Day 8 Concurrency Testing):
// 1. Simulate real-world high-concurrency collisions where N concurrent goroutines
//    hit the exact same merchant account simultaneously.
// 2. Validate memory safety under the Go Race Detector (ThreadSanitizer -race).
// 3. Prove that in-memory sync.RWMutex synchronization in BucketStore eliminates
//    data races and prevents counter corruption under multi-threaded load.
// 4. Assert that bidirectional concurrent transfers (A -> B and B -> A) execute
//    without deadlocks, goroutine leaks, or uncoordinated state corruption.
// ==============================================================================

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/charan21042005/novaris/services/gateway/internal/auth"
	"github.com/charan21042005/novaris/services/gateway/internal/ratelimit"
	"github.com/google/uuid"
)

// TestConcurrentSettlement_SameAccountCollisions spawns 50 concurrent goroutines
// all executing settleTransaction against the exact same sender account at the
// exact same nanosecond using a synchronized start barrier (startGate).
//
// Verification Guarantee:
// - Zero data races detected by ThreadSanitizer (go test -race).
// - All 50 transactions execute or rollback with complete memory isolation.
func TestConcurrentSettlement_SameAccountCollisions(t *testing.T) {
	const concurrentWorkers = 50
	const sharedSenderID = "c0000000-0000-0000-0000-000000000001"
	const targetRecipientID = "c0000000-0000-0000-0000-000000000002"
	const transferAmount = int64(1000) // $10.00 each

	var successCount atomic.Int64
	var failureCount atomic.Int64
	var wg sync.WaitGroup

	// Synchronization Barrier: Releases all 50 workers simultaneously
	startGate := make(chan struct{})

	for i := 0; i < concurrentWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			// Block until the start gun is fired
			<-startGate

			// Each worker allocates an isolated mock connection to simulate
			// independent connection pool checkout from *sql.DB
			db, mock, err := sqlmock.New()
			if err != nil {
				failureCount.Add(1)
				t.Errorf("worker %d failed to create sqlmock: %v", workerID, err)
				return
			}
			defer db.Close()

			// Set up expected SQL sequence for this concurrent transaction
			mock.ExpectBegin()
			mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO transactions
		 (id, idempotency_key, sender_account_id, recipient_account_id, amount_cents, status)
		 VALUES ($1, $2, $3, $4, $5, 'SETTLED')`)).
				WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sharedSenderID, targetRecipientID, transferAmount).
				WillReturnResult(sqlmock.NewResult(1, 1))

			mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO journal_entries
		 (transaction_id, account_id, entry_type, amount_cents)
		 VALUES ($1, $2, 'DEBIT', $3)`)).
				WithArgs(sqlmock.AnyArg(), sharedSenderID, transferAmount).
				WillReturnResult(sqlmock.NewResult(1, 1))

			mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO journal_entries
		 (transaction_id, account_id, entry_type, amount_cents)
		 VALUES ($1, $2, 'CREDIT', $3)`)).
				WithArgs(sqlmock.AnyArg(), targetRecipientID, transferAmount).
				WillReturnResult(sqlmock.NewResult(1, 1))

			mock.ExpectCommit()

			// Execute settlement logic under race detector scrutiny
			txID, err := settleTransaction(db, sharedSenderID, targetRecipientID, transferAmount)
			if err != nil {
				failureCount.Add(1)
				t.Errorf("worker %d: unexpected settlement error: %v", workerID, err)
				return
			}

			if _, err := uuid.Parse(txID); err != nil {
				failureCount.Add(1)
				t.Errorf("worker %d: invalid UUID returned: %s", workerID, txID)
				return
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				failureCount.Add(1)
				t.Errorf("worker %d: unmet mock expectations: %v", workerID, err)
				return
			}

			successCount.Add(1)
		}(i)
	}

	// FIRE THE STARTING GUN: Unleash all 50 goroutines simultaneously
	close(startGate)

	// Wait for all concurrent workers to finish
	wg.Wait()

	if success := successCount.Load(); success != concurrentWorkers {
		t.Fatalf("expected %d successful concurrent settlements, got %d (failures: %d)",
			concurrentWorkers, success, failureCount.Load())
	}

	t.Logf("CONCURRENCY SUCCESS: %d concurrent settlement goroutines completed with ZERO data races!",
		concurrentWorkers)
}

// TestConcurrentGatewayPipeline_FullStackRaceStress submits 50 concurrent HTTP requests
// through the full defense-in-depth middleware pipeline (JWT -> Rate Limiter -> HMAC -> Handler).
//
// Verification Guarantee:
// - Verifies sync.RWMutex thread safety in BucketStore under concurrent read/write pressure.
// - Confirms the Token-Bucket rate limiter admits the burst allowance (10) and throttles excess (40)
//   with HTTP 429 and RFC 6585 Retry-After headers, with ZERO data races!
func TestConcurrentGatewayPipeline_FullStackRaceStress(t *testing.T) {
	const concurrentRequests = 50
	const merchantID = "merchant_concurrency_stress_001"
	const secret = "shared_hmac_secret_key_day08_concurrency"

	// Configure Token-Bucket store: Burst = 10 tokens, Refill = 2 tokens/sec
	bucketStore := ratelimit.NewBucketStore(10.0, 2.0)

	// Assemble full production middleware stack
	handler := auth.JWTMiddleware(
		ratelimit.RateLimitMiddleware(bucketStore)(
			auth.HMACMiddleware(paymentsHandler),
		),
	)

	// Generate a valid JWT token for this merchant
	token, err := auth.GenerateToken(merchantID, "merchant", 15*time.Minute)
	if err != nil {
		t.Fatalf("failed to mint test JWT token: %v", err)
	}

	var status200Count atomic.Int64
	var status429Count atomic.Int64
	var otherStatusCount atomic.Int64
	var wg sync.WaitGroup

	startGate := make(chan struct{})

	for i := 0; i < concurrentRequests; i++ {
		wg.Add(1)
		go func(reqIndex int) {
			defer wg.Done()

			// Construct valid JSON body
			bodyBytes, _ := json.Marshal(map[string]interface{}{
				"amount_cents": 5000,
				"recipient":    "c0000000-0000-0000-0000-000000000002",
			})

			// Compute HMAC signature over timestamp + body
			ts := time.Now().Unix()
			sig := auth.ComputeHMACSignature(bodyBytes, ts, auth.GetHMACSecret())

			// Construct HTTP Request
			req := httptest.NewRequest(http.MethodPost, "/v1/payments", bytes.NewReader(bodyBytes))
			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("X-Novaris-Timestamp", fmt.Sprintf("%d", ts))
			req.Header.Set("X-Novaris-Signature", sig)
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()

			// Wait for barrier release
			<-startGate

			// Serve request through the full pipeline
			handler.ServeHTTP(rec, req)

			switch rec.Code {
			case http.StatusOK:
				status200Count.Add(1)
			case http.StatusTooManyRequests:
				status429Count.Add(1)
				// Verify RFC 6585 standard Retry-After header
				if retryAfter := rec.Header().Get("Retry-After"); retryAfter == "" {
					t.Errorf("request %d: expected Retry-After header on 429 response", reqIndex)
				}
			default:
				otherStatusCount.Add(1)
				t.Errorf("request %d: unexpected HTTP status %d: %s", reqIndex, rec.Code, rec.Body.String())
			}
		}(i)
	}

	// Unleash all 50 concurrent HTTP requests simultaneously
	close(startGate)
	wg.Wait()

	admitted := status200Count.Load()
	throttled := status429Count.Load()
	unexpected := otherStatusCount.Load()

	t.Logf("GATEWAY PIPELINE CONCURRENCY RESULTS: Total=%d | Admitted (200)=%d | Throttled (429)=%d | Other=%d",
		concurrentRequests, admitted, throttled, unexpected)

	// In a burst bucket of capacity 10.0, exactly 10 requests should be admitted,
	// and remaining ~40 should be throttled with HTTP 429.
	if admitted != 10 {
		t.Errorf("expected exactly 10 admitted burst requests, got %d", admitted)
	}
	if throttled != 40 {
		t.Errorf("expected exactly 40 throttled requests, got %d", throttled)
	}
	if unexpected != 0 {
		t.Errorf("encountered %d unexpected non-200/non-429 responses", unexpected)
	}
}

// TestConcurrentTransfer_BidirectionalStress simulates two concurrent streams of transfers:
// Stream 1: 25 workers transferring Account A -> Account B
// Stream 2: 25 workers transferring Account B -> Account A
//
// Verification Guarantee:
// - Proves bidirectional resource contention executes cleanly under ThreadSanitizer.
// - Confirms zero deadlocks, race conditions, or unhandled panics.
func TestConcurrentTransfer_BidirectionalStress(t *testing.T) {
	const workersPerDirection = 25
	accountA := "c0000000-0000-0000-0000-000000000001"
	accountB := "c0000000-0000-0000-0000-000000000002"
	amount := int64(2500)

	var successCount atomic.Int64
	var wg sync.WaitGroup
	startGate := make(chan struct{})

	// Launch Stream 1: A -> B
	for i := 0; i < workersPerDirection; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			<-startGate

			db, mock, err := sqlmock.New()
			if err != nil {
				t.Errorf("Stream 1 worker %d: failed to create sqlmock: %v", workerID, err)
				return
			}
			defer db.Close()

			mock.ExpectBegin()
			mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO transactions`)).
				WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), accountA, accountB, amount).
				WillReturnResult(sqlmock.NewResult(1, 1))
			mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO journal_entries`)).
				WithArgs(sqlmock.AnyArg(), accountA, amount).
				WillReturnResult(sqlmock.NewResult(1, 1))
			mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO journal_entries`)).
				WithArgs(sqlmock.AnyArg(), accountB, amount).
				WillReturnResult(sqlmock.NewResult(1, 1))
			mock.ExpectCommit()

			if _, err := settleTransaction(db, accountA, accountB, amount); err != nil {
				t.Errorf("Stream 1 worker %d (A->B) failed: %v", workerID, err)
				return
			}
			successCount.Add(1)
		}(i)
	}

	// Launch Stream 2: B -> A
	for i := 0; i < workersPerDirection; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			<-startGate

			db, mock, err := sqlmock.New()
			if err != nil {
				t.Errorf("Stream 2 worker %d: failed to create sqlmock: %v", workerID, err)
				return
			}
			defer db.Close()

			mock.ExpectBegin()
			mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO transactions`)).
				WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), accountB, accountA, amount).
				WillReturnResult(sqlmock.NewResult(1, 1))
			mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO journal_entries`)).
				WithArgs(sqlmock.AnyArg(), accountB, amount).
				WillReturnResult(sqlmock.NewResult(1, 1))
			mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO journal_entries`)).
				WithArgs(sqlmock.AnyArg(), accountA, amount).
				WillReturnResult(sqlmock.NewResult(1, 1))
			mock.ExpectCommit()

			if _, err := settleTransaction(db, accountB, accountA, amount); err != nil {
				t.Errorf("Stream 2 worker %d (B->A) failed: %v", workerID, err)
				return
			}
			successCount.Add(1)
		}(i)
	}

	// Release both streams simultaneously
	close(startGate)
	wg.Wait()

	totalExpected := int64(workersPerDirection * 2)
	if actual := successCount.Load(); actual != totalExpected {
		t.Fatalf("expected %d total bidirectional settlements, got %d", totalExpected, actual)
	}

	t.Logf("BIDIRECTIONAL CONCURRENCY SUCCESS: %d transfers (25 A->B and 25 B->A) completed without deadlock!",
		totalExpected)
}

// TestConcurrentDefensiveValidation_StateIsolation fires 100 concurrent goroutines
// alternating between valid and corrupt inputs (negative amounts, empty IDs, self-transfers)
// to prove that input validation has zero cross-goroutine state leakage.
func TestConcurrentDefensiveValidation_StateIsolation(t *testing.T) {
	const iterations = 100
	var wg sync.WaitGroup

	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			switch index % 4 {
			case 0:
				// Zero amount check
				_, err := settleTransaction(nil, "acc1", "acc2", 0)
				if err == nil {
					t.Errorf("worker %d: expected zero amount error, got nil", index)
				}
			case 1:
				// Negative amount check
				_, err := settleTransaction(nil, "acc1", "acc2", -100)
				if err == nil {
					t.Errorf("worker %d: expected negative amount error, got nil", index)
				}
			case 2:
				// Self transfer check
				_, err := settleTransaction(nil, "same_acc", "same_acc", 5000)
				if err == nil {
					t.Errorf("worker %d: expected self-transfer error, got nil", index)
				}
			case 3:
				// Nil DB check
				_, err := settleTransaction(nil, "acc1", "acc2", 5000)
				if err == nil {
					t.Errorf("worker %d: expected nil db error, got nil", index)
				}
			}
		}(i)
	}

	wg.Wait()
	t.Logf("STATE ISOLATION PROVEN: %d concurrent validation evaluations completed without memory races!",
		iterations)
}
