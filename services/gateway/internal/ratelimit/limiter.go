// ==============================================================================
// NOVARIS: Distributed Payment Settlement & Adaptive Fraud Intelligence Platform
// MODULE: In-Memory Token-Bucket Rate Limiter & Concurrency Defense (Day 4)
// PACKAGE: internal/ratelimit/limiter.go
// ==============================================================================
//
// Theoretical Architecture & Purpose:
// 1. Day 2 (JWT) proved WHO is calling (Merchant Authentication / Identity).
// 2. Day 3 (HMAC) proved WHAT was sent was untampered (Payload Integrity).
// 3. Day 4 (Rate Limiting) proves HOW FAST / HOW MUCH:
//    - Solves the "Noisy Neighbor" problem and protects downstream databases/brokers.
//    - Defends against runaway client retry loops, accidental script floods, and DoS.
// 4. Token-Bucket Algorithm Mechanics:
//    - Capacity (b.capacity = 10.0): Maximum burst tolerance.
//    - Refill Rate (b.refillRate = 2.0 tokens/sec): Sustained long-run throughput ceiling.
//    - Each HTTP request consumes exactly 1.0 token.
// 5. Lazy Refill (Zero-Goroutine Architecture):
//    - Computes elapsed time mathematically on each request: tokens += elapsed * refillRate.
//    - Eliminates background ticker goroutines, avoiding timer leaks and idle CPU waste.
// 6. Concurrency Safety:
//    - Guarded by sync.RWMutex on the store and sync.Mutex on each individual bucket.
//    - Prevents fatal Go runtime concurrent map read/write panics and race condition oversells.
// 7. Standard RFC 6585 Compliance:
//    - Emits HTTP 429 Too Many Requests when tokens are exhausted.
//    - Supplies standard "Retry-After" header specifying seconds until next token availability.
// ==============================================================================

package ratelimit

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/charan21042005/novaris/services/gateway/internal/auth"
)

// Default limiter tuning parameters for the NOVARIS API Gateway.
const (
	// DefaultBucketCapacity defines the maximum number of tokens a bucket can hold.
	// This represents the burst allowance (e.g., 10 requests allowed in a sudden spike).
	DefaultBucketCapacity float64 = 10.0

	// DefaultRefillRate defines how many tokens are added back to the bucket per second.
	// 2.0 tokens/sec = sustained rate of 2 requests/sec per merchant.
	DefaultRefillRate float64 = 2.0
)

// TokenBucket represents an individual tenant's rate-limiting bucket.
// All field mutations are synchronized via the embedded mutex.
type TokenBucket struct {
	mu           sync.Mutex
	tokens       float64   // Current available tokens (can be fractional during refill)
	capacity     float64   // Maximum burst capacity
	refillRate   float64   // Tokens replenished per second
	lastRefilled time.Time // Exact timestamp of the most recent consumption or refill
}

// NewTokenBucket instantiates a freshly initialized TokenBucket at full capacity.
func NewTokenBucket(capacity, refillRate float64) *TokenBucket {
	return &TokenBucket{
		tokens:       capacity,
		capacity:     capacity,
		refillRate:   refillRate,
		lastRefilled: time.Now(),
	}
}

// refillLocked mathematically recalculates the current token balance based on elapsed time.
// CALLER MUST HOLD b.mu LOCK BEFORE CALLING THIS METHOD.
//
// Mathematical Formula:
//   elapsed = now - lastRefilled
//   new_tokens = min(capacity, current_tokens + (elapsed * refillRate))
func (b *TokenBucket) refillLocked(now time.Time) {
	elapsed := now.Sub(b.lastRefilled).Seconds()
	if elapsed > 0 {
		b.tokens = math.Min(b.capacity, b.tokens+(elapsed*b.refillRate))
		b.lastRefilled = now
	}
}

// TryConsume attempts to withdraw 1 token from the bucket.
// Returns:
//   - allowed (bool): true if a token was available and consumed; false if rate-limited.
//   - retryAfter (int): seconds the caller must wait before at least 1 token is replenished.
func (b *TokenBucket) TryConsume() (bool, int) {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	b.refillLocked(now)

	// If at least 1.0 full token is available, consume it and grant access
	if b.tokens >= 1.0 {
		b.tokens -= 1.0
		return true, 0
	}

	// Calculate wait time until next token becomes available:
	// deficit = 1.0 - tokens
	// waitSeconds = ceil(deficit / refillRate)
	deficit := 1.0 - b.tokens
	waitSeconds := 1
	if b.refillRate > 0 {
		waitSeconds = int(math.Ceil(deficit / b.refillRate))
		if waitSeconds < 1 {
			waitSeconds = 1
		}
	}

	return false, waitSeconds
}

// AvailableTokens returns the current integer floor of tokens available in the bucket.
func (b *TokenBucket) AvailableTokens() int {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.refillLocked(time.Now())
	if b.tokens < 0 {
		return 0
	}
	return int(math.Floor(b.tokens))
}

// BucketStore manages isolated, per-merchant token buckets in memory.
// Thread-safe for concurrent read and write operations across thousands of goroutines.
type BucketStore struct {
	mu         sync.RWMutex
	buckets    map[string]*TokenBucket
	capacity   float64
	refillRate float64
}

// NewBucketStore creates a thread-safe registry of per-merchant token buckets.
func NewBucketStore(capacity, refillRate float64) *BucketStore {
	return &BucketStore{
		buckets:    make(map[string]*TokenBucket),
		capacity:   capacity,
		refillRate: refillRate,
	}
}

// GetBucket retrieves an existing merchant bucket or lazily instantiates a new one.
// Uses a double-checked locking pattern (RLock first, upgrade to Lock if absent)
// to maximize concurrent throughput under heavy read traffic.
func (s *BucketStore) GetBucket(merchantID string) *TokenBucket {
	// Fast path: concurrent read lock
	s.mu.RLock()
	bucket, exists := s.buckets[merchantID]
	s.mu.RUnlock()
	if exists {
		return bucket
	}

	// Slow path: write lock to register new tenant bucket
	s.mu.Lock()
	defer s.mu.Unlock()

	// Re-check existence in case another goroutine created it during lock upgrade
	if bucket, exists = s.buckets[merchantID]; exists {
		return bucket
	}

	bucket = NewTokenBucket(s.capacity, s.refillRate)
	s.buckets[merchantID] = bucket
	return bucket
}

// RateLimitMiddleware returns an HTTP middleware adapter that enforces per-merchant rate limits.
// Positioning:
//   Must be positioned AFTER auth.JWTMiddleware (so merchant claims are present in r.Context()).
//   Must be positioned BEFORE auth.HMACMiddleware (so rate-limited floods are rejected before
//   incurring expensive HMAC body hashing and socket reads - Fail-Fast Principle).
func RateLimitMiddleware(store *BucketStore) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Extract merchant identity from validated JWT claims
			var merchantKey string
			if claims, ok := auth.GetClaimsFromContext(r.Context()); ok && claims.MerchantID != "" {
				merchantKey = claims.MerchantID
			} else {
				// Fallback to client IP if unauthenticated or missing claims (defensive)
				merchantKey = extractClientIP(r)
			}

			bucket := store.GetBucket(merchantKey)
			allowed, retryAfter := bucket.TryConsume()

			// Expose standard rate limit telemetry headers (RFC 6585 & Stripe convention)
			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%.0f", store.capacity))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(bucket.AvailableTokens()))

			if !allowed {
				// 429 Too Many Requests
				w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.WriteHeader(http.StatusTooManyRequests)
				fmt.Fprintf(w, `{"error":"too_many_requests","message":"rate limit exceeded, please retry after specified delay","retry_after_seconds":%d}`+"\n", retryAfter)
				return // Short-circuit: next handler is never invoked
			}

			// Token successfully consumed. Forward request downstream.
			next.ServeHTTP(w, r)
		}
	}
}

// extractClientIP extracts the remote IP address, stripping any ephemeral port suffix.
func extractClientIP(r *http.Request) string {
	// Inspect standard reverse proxy headers first
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	// Fallback to direct socket address
	remoteAddr := r.RemoteAddr
	if idx := strings.LastIndex(remoteAddr, ":"); idx != -1 {
		return remoteAddr[:idx]
	}
	return remoteAddr
}
