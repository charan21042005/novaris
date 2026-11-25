// ==============================================================================
// NOVARIS: Distributed Payment Settlement & Adaptive Fraud Intelligence Platform
// UNIT & CONCURRENCY TEST SUITE: Idempotent API Request Engine (Day 10)
// FILE: services/gateway/idempotency_test.go
// ==============================================================================

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

// TestIdempotency_NewRequestProcessing verifies that a brand new request claims the key,
// invokes the downstream handler, and caches the response.
func TestIdempotency_NewRequestProcessing(t *testing.T) {
	store := NewInMemoryIdempotencyStore()
	idempKey := "idemp-test-001"
	payload := []byte(`{"amount_cents": 5000, "recipient": "merchant_abc"}`)

	var handlerInvocations atomic.Int64
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerInvocations.Add(1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status":         "settled",
			"transaction_id": "tx-12345",
		})
	})

	wrapped := IdempotencyMiddleware(store)(handler)

	req := httptest.NewRequest(http.MethodPost, "/v1/payments", bytes.NewReader(payload))
	req.Header.Set(HeaderIdempotencyKey, idempKey)
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201 Created, got %d", w.Code)
	}

	if replayed := w.Header().Get(HeaderIdempotentReplayed); replayed != "false" {
		t.Errorf("expected Idempotent-Replayed: false, got %s", replayed)
	}

	if handlerInvocations.Load() != 1 {
		t.Errorf("expected handler to be invoked 1 time, got %d", handlerInvocations.Load())
	}
}

// TestIdempotency_DuplicateRequestReturnsCachedResponse verifies that a duplicate request
// with the exact same key and payload returns the original cached result without calling the handler again.
func TestIdempotency_DuplicateRequestReturnsCachedResponse(t *testing.T) {
	store := NewInMemoryIdempotencyStore()
	idempKey := "idemp-test-002"
	payload := []byte(`{"amount_cents": 5000, "recipient": "merchant_abc"}`)

	var handlerInvocations atomic.Int64
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerInvocations.Add(1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":         "settled",
			"transaction_id": "tx-unique-999",
			"amount_cents":   5000,
		})
	})

	wrapped := IdempotencyMiddleware(store)(handler)

	// Call 1: Original Request
	req1 := httptest.NewRequest(http.MethodPost, "/v1/payments", bytes.NewReader(payload))
	req1.Header.Set(HeaderIdempotencyKey, idempKey)
	w1 := httptest.NewRecorder()
	wrapped.ServeHTTP(w1, req1)

	if w1.Code != http.StatusOK {
		t.Fatalf("call 1: expected 200 OK, got %d", w1.Code)
	}
	if w1.Header().Get(HeaderIdempotentReplayed) != "false" {
		t.Errorf("call 1: expected Idempotent-Replayed: false")
	}

	// Call 2: Duplicate Network Retry
	req2 := httptest.NewRequest(http.MethodPost, "/v1/payments", bytes.NewReader(payload))
	req2.Header.Set(HeaderIdempotencyKey, idempKey)
	w2 := httptest.NewRecorder()
	wrapped.ServeHTTP(w2, req2)

	// Golden Invariant Assertions:
	if w2.Code != http.StatusOK {
		t.Fatalf("call 2: expected 200 OK, got %d", w2.Code)
	}
	if replayed := w2.Header().Get(HeaderIdempotentReplayed); replayed != "true" {
		t.Errorf("call 2: expected Idempotent-Replayed: true, got %s", replayed)
	}

	// Assert handler was NOT invoked a second time!
	if handlerInvocations.Load() != 1 {
		t.Fatalf("DOUBLE CHARGE RISK: handler was invoked %d times!", handlerInvocations.Load())
	}

	// Assert response body matches original perfectly
	if w1.Body.String() != w2.Body.String() {
		t.Errorf("expected cached response body to match original.\nGot: %s\nExpected: %s",
			w2.Body.String(), w1.Body.String())
	}

	t.Logf("DUPLICATE RETRY SUCCESS: Cached response returned with Idempotent-Replayed: true!")
}

// TestIdempotency_PayloadMismatchRejectedWith422 verifies Stripe-compliant security behavior:
// If a client reuses an idempotency key with a DIFFERENT payload, it is rejected with HTTP 422.
func TestIdempotency_PayloadMismatchRejectedWith422(t *testing.T) {
	store := NewInMemoryIdempotencyStore()
	idempKey := "idemp-test-003"

	payloadOriginal := []byte(`{"amount_cents": 1000, "recipient": "merchant_1"}`)
	payloadMismatched := []byte(`{"amount_cents": 9000, "recipient": "merchant_2"}`) // Different amount!

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"settled"}`))
	})

	wrapped := IdempotencyMiddleware(store)(handler)

	// Request 1: Original payload
	req1 := httptest.NewRequest(http.MethodPost, "/v1/payments", bytes.NewReader(payloadOriginal))
	req1.Header.Set(HeaderIdempotencyKey, idempKey)
	w1 := httptest.NewRecorder()
	wrapped.ServeHTTP(w1, req1)
	if w1.Code != http.StatusOK {
		t.Fatalf("request 1 failed: %d", w1.Code)
	}

	// Request 2: Mismatched payload reusing the same key!
	req2 := httptest.NewRequest(http.MethodPost, "/v1/payments", bytes.NewReader(payloadMismatched))
	req2.Header.Set(HeaderIdempotencyKey, idempKey)
	w2 := httptest.NewRecorder()
	wrapped.ServeHTTP(w2, req2)

	if w2.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected HTTP 422 Unprocessable Entity on payload mismatch, got %d", w2.Code)
	}

	var resp map[string]interface{}
	_ = json.Unmarshal(w2.Body.Bytes(), &resp)
	if resp["error"] != "idempotency_conflict" {
		t.Errorf("expected error 'idempotency_conflict', got %v", resp["error"])
	}
}

// TestIdempotency_ConcurrentDuplicateReturns409 verifies that a duplicate request
// arriving while the original is still IN_PROGRESS returns HTTP 409 Conflict.
func TestIdempotency_ConcurrentDuplicateReturns409(t *testing.T) {
	store := NewInMemoryIdempotencyStore()
	idempKey := "idemp-test-004"
	payload := []byte(`{"amount_cents": 2500}`)

	// Pre-seed the record as IN_PROGRESS
	hash := ComputePayloadHash(payload)
	_, isNew, err := store.LockOrGet(context.Background(), idempKey, "acc-test", "/v1/payments", hash)
	if err != nil || !isNew {
		t.Fatalf("failed to seed in-progress record: %v", err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called when request is in progress")
	})

	wrapped := IdempotencyMiddleware(store)(handler)

	req := httptest.NewRequest(http.MethodPost, "/v1/payments", bytes.NewReader(payload))
	req.Header.Set(HeaderIdempotencyKey, idempKey)
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected HTTP 409 Conflict on in-progress duplicate, got %d", w.Code)
	}
}

// TestIdempotency_MissingKeyRejectedWith400 verifies that mutating requests without
// an Idempotency-Key header are rejected with HTTP 400 Bad Request.
func TestIdempotency_MissingKeyRejectedWith400(t *testing.T) {
	store := NewInMemoryIdempotencyStore()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called when idempotency key is missing")
	})

	wrapped := IdempotencyMiddleware(store)(handler)

	req := httptest.NewRequest(http.MethodPost, "/v1/payments", bytes.NewReader([]byte(`{}`)))
	// No Idempotency-Key header set!
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected HTTP 400 Bad Request on missing key, got %d", w.Code)
	}
}

// TestIdempotency_50ConcurrentWorkersSimultaneousRace simulates 50 concurrent goroutines
// submitting the EXACT SAME request with the identical Idempotency-Key simultaneously.
// INVARIANT PROOF:
// The underlying settlement handler MUST BE INVOKED EXACTLY ONCE!
// The account is charged EXACTLY ONCE!
func TestIdempotency_50ConcurrentWorkersSimultaneousRace(t *testing.T) {
	const workers = 50
	store := NewInMemoryIdempotencyStore()
	idempKey := "idemp-race-test-" + uuid.New().String()
	payload := []byte(`{"amount_cents": 1000, "recipient": "merchant_xyz"}`)

	var handlerInvocations atomic.Int64
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerInvocations.Add(1)
		time.Sleep(10 * time.Millisecond) // Simulate DB processing
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status": "settled",
		})
	})

	wrapped := IdempotencyMiddleware(store)(handler)

	var wg sync.WaitGroup
	startGate := make(chan struct{})

	var successCount atomic.Int64
	var inProgressConflictCount atomic.Int64

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			<-startGate // Synchronized release

			req := httptest.NewRequest(http.MethodPost, "/v1/payments", bytes.NewReader(payload))
			req.Header.Set(HeaderIdempotencyKey, idempKey)
			w := httptest.NewRecorder()

			wrapped.ServeHTTP(w, req)

			switch w.Code {
			case http.StatusOK:
				successCount.Add(1)
			case http.StatusConflict:
				inProgressConflictCount.Add(1)
			}
		}(i)
	}

	close(startGate)
	wg.Wait()

	// THE GOLDEN IDEMPOTENCY INVARIANT:
	if handlerInvocations.Load() != 1 {
		t.Fatalf("CRITICAL DOUBLE SPEND: handler was invoked %d times under concurrency!",
			handlerInvocations.Load())
	}

	t.Logf("IDEMPOTENCY RACE PROVEN: Handler invoked EXACTLY 1 time across 50 concurrent callers! (Success: %d, InProgress: %d)",
		successCount.Load(), inProgressConflictCount.Load())
}

// TestIdempotency_PostgresStoreMock tests the PostgreSQL database implementation of IdempotencyStore.
func TestIdempotency_PostgresStoreMock(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to initialize sqlmock: %v", err)
	}
	defer db.Close()

	store := NewPostgresIdempotencyStore(db)
	ctx := context.Background()
	key := "idemp-pg-test-01"
	payloadHash := ComputePayloadHash([]byte(`{"amount": 100}`))

	// 1. First Call: INSERT succeeds (claims key)
	mock.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO idempotency_keys (key, account_id, request_path, request_hash, status)
		VALUES ($1, $2, $3, $4, 'IN_PROGRESS')
		ON CONFLICT (key) DO NOTHING;
	`)).
		WithArgs(key, "acc-1", "/v1/payments", payloadHash).
		WillReturnResult(sqlmock.NewResult(1, 1))

	rec, isNew, err := store.LockOrGet(ctx, key, "acc-1", "/v1/payments", payloadHash)
	if err != nil || !isNew {
		t.Fatalf("expected new record claimed, got isNew=%v, err=%v", isNew, err)
	}
	if rec.Status != StatusInProgress {
		t.Errorf("expected status IN_PROGRESS, got %s", rec.Status)
	}

	// 2. Save Response: UPDATE marks COMPLETED
	mock.ExpectExec(regexp.QuoteMeta(`
		UPDATE idempotency_keys
		SET status = 'COMPLETED',
		    response_code = $1,
		    response_body = $2::jsonb
		WHERE key = $3;
	`)).
		WithArgs(200, `{"status":"settled"}`, key).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = store.SaveResponse(ctx, key, 200, []byte(`{"status":"settled"}`))
	if err != nil {
		t.Fatalf("failed to save response: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled mock expectations: %v", err)
	}
}
