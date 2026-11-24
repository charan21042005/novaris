<div align="center">

<!-- Top Banner -->
<p align="center">
  <img src="../../../../assets/banner.svg" alt="NOVARIS Banner" width="100%" />
</p>

# ⏱️ NOVARIS Rate Limiting Engine (`/internal/ratelimit`)
### In-Memory Token-Bucket Algorithm, Concurrency Safety & Fail-Fast Gateway Ingress Protection

<p align="center">
  <img src="https://img.shields.io/badge/Language-Go_1.22+-00ADD8?style=flat-square&logo=go&logoColor=white" alt="Go" />
  <img src="https://img.shields.io/badge/Algorithm-Token_Bucket-FF6D00?style=flat-square&logo=speedtest&logoColor=white" alt="Algorithm" />
  <img src="https://img.shields.io/badge/Refill-Zero--Goroutine_Lazy_Math-00C853?style=flat-square&logo=hyperledger&logoColor=white" alt="Refill" />
  <img src="https://img.shields.io/badge/Standard-RFC_6585_HTTP_429-7C4DFF?style=flat-square&logo=shield&logoColor=white" alt="RFC 6585" />
  <img src="https://img.shields.io/badge/Concurrency-sync.RWMutex_Safe-FFA000?style=flat-square&logo=vault&logoColor=white" alt="Concurrency" />
</p>

---

</div>

## 📌 Package Overview

The `internal/ratelimit` package is the traffic control engine of the NOVARIS API Gateway. It enforces rate limits on incoming payment requests on a **per-merchant basis**, preventing rogue clients, buggy retry loops, and malicious Denial of Service (DoS) floods from exhausting gateway memory, database connections, and event queues.

```
                          INCOMING PAYMENT REQUEST
                                     │
                                     ▼
        ┌─────────────────────────────────────────────────────────┐
        │             LAYER 1: IDENTITY (Day 2)                   │
        │              auth.JWTMiddleware (jwt.go)                │
        │                                                         │
        │  * Validates Bearer token cryptographic signature       │
        │  * Checks token expiration (exp)                        │
        │  * Injects Claims (MerchantID, Role) into context       │
        └────────────────────────────┬────────────────────────────┘
                                     │ Passes
                                     ▼
        ┌─────────────────────────────────────────────────────────┐
        │          LAYER 2: TRAFFIC CONTROL (Day 4)               │
        │       ratelimit.RateLimitMiddleware (limiter.go)        │
        │                                                         │
        │  * Reads claims.MerchantID from context                 │
        │  * Fetches or creates per-merchant TokenBucket          │
        │  * Computes lazy mathematical refill: Δt × r            │
        │  * If tokens < 1.0: REJECTS with HTTP 429 + Retry-After │
        │  * CRITICAL: Drops floods BEFORE costly body hashing!   │
        └────────────────────────────┬────────────────────────────┘
                                     │ Passes (1 token consumed)
                                     ▼
        ┌─────────────────────────────────────────────────────────┐
        │             LAYER 3: INTEGRITY (Day 3)                  │
        │             auth.HMACMiddleware (hmac.go)               │
        │                                                         │
        │  * Validates X-Novaris-Timestamp within ±300s window    │
        │  * Reads body bytes & restores r.Body via io.NopCloser  │
        │  * Recomputes HMAC-SHA256 & compares constant-time      │
        └────────────────────────────┬────────────────────────────┘
                                     │ Passes
                                     ▼
        ┌─────────────────────────────────────────────────────────┐
        │                   CORE SETTLEMENT                       │
        │             paymentsHandler (main.go)                   │
        │                                                         │
        │  * Consumes verified claims and restored JSON body      │
        │  * Dispatches transaction to Ledger & Kafka pipeline    │
        └─────────────────────────────────────────────────────────┘
```

---

## 📐 Mathematical Formulation: The Token-Bucket Algorithm

The Token Bucket algorithm models rate limiting as a bucket with finite capacity $c$ that continuously fills with tokens at rate $r$ (tokens per second):

* **Capacity ($c = 10.0$):** Defines the **burst allowance**. A merchant can submit up to 10 requests instantaneously.
* **Refill Rate ($r = 2.0$):** Defines the **sustained throughput ceiling**. Long-term throughput cannot exceed 2.0 requests/second.
* **Cost:** Each HTTP request consumes exactly $1.0$ token.

### Zero-Goroutine Lazy Refill Formula

Traditional rate limiters spawn a background ticker goroutine (`time.NewTicker`) for every active merchant. In high-density multi-tenant gateways hosting 100,000 merchants, this creates 100,000 idle goroutines, triggering memory bloat and timer contention.

NOVARIS eliminates background timers entirely using **Lazy Mathematical Refill**: tokens are recomputed purely on request arrival based on elapsed wall-clock time:

$$\Delta t = t_{\text{current}} - t_{\text{last\_refill}}$$

$$\text{tokens}_{\text{new}} = \min\Big(c,\; \text{tokens}_{\text{current}} + \Delta t \times r\Big)$$

$$t_{\text{last\_refill}} = t_{\text{current}}$$

If $\text{tokens}_{\text{new}} \ge 1.0$, the request is permitted and $1.0$ token is deducted:
$$\text{tokens}_{\text{after}} = \text{tokens}_{\text{new}} - 1.0$$

If $\text{tokens}_{\text{new}} < 1.0$, the request is rejected with HTTP 429 and the `Retry-After` duration is computed:
$$\text{Retry-After} = \left\lceil \frac{1.0 - \text{tokens}_{\text{new}}}{r} \right\rceil \text{ seconds}$$

---

## 🔒 Concurrency & Thread Safety

In Go, concurrent reads and writes to a standard `map` result in a fatal, unrecoverable runtime crash:
`fatal error: concurrent map read and map write`

The NOVARIS limiter implements a two-tier locking hierarchy to guarantee $100\%$ thread safety under high concurrency:

1. **`BucketStore.mu` (`sync.RWMutex`):**
   * Uses an optimistic read lock (`RLock()`) to look up existing merchant buckets with zero contention.
   * Promotes to an exclusive write lock (`Lock()`) with double-checked locking only when allocating a bucket for a newly observed merchant.

2. **`TokenBucket.mu` (`sync.Mutex`):**
   * Guards the individual bucket's `tokens` balance and `lastRefill` timestamp.
   * Lock contention is isolated strictly to goroutines serving the *same* merchant, ensuring independent merchants never block each other.

---

## 🌐 RFC 6585 HTTP 429 & Retry-After Specification

When a merchant exhausts their token balance, the middleware intercepts the call and emits:

```http
HTTP/1.1 429 Too Many Requests
Content-Type: application/json
Retry-After: 1
```

```json
{
  "error": "too_many_requests",
  "merchant_id": "merchant_123",
  "message": "rate limit quota exhausted. Please retry after the specified backoff interval.",
  "retry_after_seconds": 1
}
```

---

## 🧪 Verification & Test Suite

The package includes a comprehensive unit and concurrent test suite (`limiter_test.go`):

| Test Case | Description | Expected Outcome |
| :--- | :--- | :--- |
| `TestTokenBucket_BurstAndExhaustion` | Consumes burst capacity ($c=5.0$) then sends immediate 6th request | First 5 return 200 OK; 6th returns 429 + `Retry-After >= 1` |
| `TestTokenBucket_RefillOverTime` | Exhausts bucket, sleeps 150ms with $r=10.0$ tokens/sec | Correctly refills 1 token and allows next request |
| `TestTokenBucket_CapacityCeiling` | Sleeps after bucket is full | Token count is strictly clamped to $c$ and never overflows |
| `TestBucketStore_PerMerchantIsolation` | Floods Merchant A's bucket to exhaustion; queries Merchant B | Merchant A is throttled (429); Merchant B succeeds (200 OK) |
| `TestTokenBucket_ConcurrentRaceConditions` | Spawns 100 concurrent goroutines competing for 20 tokens | Exactly 20 requests succeed; exactly 80 requests are throttled; zero race panics |
| `TestRateLimitMiddleware` | Tests end-to-end HTTP middleware handler wrapping | Validates HTTP 429 status code, JSON response body, and `Retry-After` header |

```bash
# Run ratelimit test suite
go test -v ./internal/ratelimit
```

---

<div align="center">

<p align="center">
  <img src="../../../../assets/footer.svg" alt="NOVARIS Footer" width="100%" />
</p>

</div>
