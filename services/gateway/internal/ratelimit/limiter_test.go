// ==============================================================================
// NOVARIS: Distributed Payment Settlement & Adaptive Fraud Intelligence Platform
// UNIT TEST SUITE: Token-Bucket Rate Limiter & Concurrency Race Detection (Day 4)
// PACKAGE: internal/ratelimit/limiter_test.go
// ==============================================================================

package ratelimit

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/charan21042005/novaris/services/gateway/internal/auth"
)

func TestTokenBucket_BurstAndExhaustion(t *testing.T) {
	// Capacity = 5, Refill = 1 token/sec
	bucket := NewTokenBucket(5.0, 1.0)

	// Consume 5 tokens in immediate succession (burst)
	for i := 1; i <= 5; i++ {
		allowed, _ := bucket.TryConsume()
		if !allowed {
			t.Fatalf("request %d within burst capacity was unexpectedly rate limited", i)
		}
	}

	// 6th immediate request must be throttled
	allowed, retryAfter := bucket.TryConsume()
	if allowed {
		t.Fatal("expected 6th request to be rejected, but it was allowed")
	}
	if retryAfter < 1 {
		t.Errorf("expected Retry-After >= 1 second, got %d", retryAfter)
	}
}

func TestTokenBucket_RefillOverTime(t *testing.T) {
	// Capacity = 2, Refill = 10 tokens/sec (fast refill for testing)
	bucket := NewTokenBucket(2.0, 10.0)

	// Exhaust all tokens
	bucket.TryConsume()
	bucket.TryConsume()

	allowed, _ := bucket.TryConsume()
	if allowed {
		t.Fatal("expected bucket to be exhausted")
	}

	// Wait 150ms (refills 1.5 tokens)
	time.Sleep(150 * time.Millisecond)

	allowed, _ = bucket.TryConsume()
	if !allowed {
		t.Fatal("expected request to succeed after refill interval")
	}
}

func TestTokenBucket_CapacityCeiling(t *testing.T) {
	// Capacity = 3, Refill = 100 tokens/sec
	bucket := NewTokenBucket(3.0, 100.0)

	// Wait 50ms - mathematically would refill 5 tokens, but should cap at capacity = 3.0
	time.Sleep(50 * time.Millisecond)

	avail := bucket.AvailableTokens()
	if avail > 3 {
		t.Fatalf("expected available tokens capped at 3, got %d", avail)
	}
}

func TestBucketStore_PerMerchantIsolation(t *testing.T) {
	// Capacity = 2, Refill = 0.1 tokens/sec
	store := NewBucketStore(2.0, 0.1)

	// Merchant A exhausts their quota
	bucketA := store.GetBucket("merchant_a")
	bucketA.TryConsume()
	bucketA.TryConsume()

	allowedA, _ := bucketA.TryConsume()
	if allowedA {
		t.Fatal("expected merchant_a to be throttled")
	}

	// Merchant B should have an independent full bucket
	bucketB := store.GetBucket("merchant_b")
	allowedB, _ := bucketB.TryConsume()
	if !allowedB {
		t.Fatal("expected merchant_b to succeed despite merchant_a being throttled (Noisy Neighbor isolation failure)")
	}
}

func TestTokenBucket_ConcurrentRaceConditions(t *testing.T) {
	// Capacity = 25, Refill = 0.0 (no refill during test to measure exact burst count)
	bucket := NewTokenBucket(25.0, 0.0)

	var allowedCount int64
	var rejectedCount int64
	var wg sync.WaitGroup

	goroutines := 100

	// Launch 100 concurrent goroutines racing to consume from the same bucket
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			allowed, _ := bucket.TryConsume()
			if allowed {
				atomic.AddInt64(&allowedCount, 1)
			} else {
				atomic.AddInt64(&rejectedCount, 1)
			}
		}()
	}

	wg.Wait()

	// Exactly 25 requests must have succeeded, and exactly 75 must have been rejected
	if allowedCount != 25 {
		t.Errorf("expected exactly 25 allowed requests under concurrent race, got %d", allowedCount)
	}
	if rejectedCount != 75 {
		t.Errorf("expected exactly 75 rejected requests under concurrent race, got %d", rejectedCount)
	}
}

func TestRateLimitMiddleware(t *testing.T) {
	store := NewBucketStore(3.0, 1.0)
	handlerRanCount := 0

	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerRanCount++
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"success"}`))
	})

	middleware := RateLimitMiddleware(store)(dummyHandler)

	// Create request with injected claims for merchant_xyz
	makeReq := func() *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/v1/payments", nil)
		claims := &auth.Claims{
			MerchantID: "merchant_xyz",
			Role:       "merchant",
		}
		ctx := context.WithValue(req.Context(), auth.ClaimsContextKey, claims)
		rec := httptest.NewRecorder()
		middleware.ServeHTTP(rec, req.WithContext(ctx))
		return rec
	}

	// 3 requests within burst capacity should return 200 OK
	for i := 1; i <= 3; i++ {
		rec := makeReq()
		if rec.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200 OK, got %d", i, rec.Code)
		}
		if rec.Header().Get("X-RateLimit-Limit") != "3" {
			t.Errorf("expected X-RateLimit-Limit 3, got %s", rec.Header().Get("X-RateLimit-Limit"))
		}
	}

	// 4th request must be rejected with 429 Too Many Requests
	rec4 := makeReq()
	if rec4.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 4th request to return 429 Too Many Requests, got %d", rec4.Code)
	}
	if rec4.Header().Get("Retry-After") == "" {
		t.Error("expected Retry-After header on 429 response")
	}

	if handlerRanCount != 3 {
		t.Errorf("expected inner handler to run exactly 3 times, got %d", handlerRanCount)
	}
}
