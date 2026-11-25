# ADR-0002: Five-Layer Defense-in-Depth Middleware Pipeline Ordering

* **Status:** Accepted
* **Date:** 2026-11-25
* **Deciders:** Tokachichu Sri Charan (Core Systems Architect)
* **Technical Area:** API Gateway, Ingress Defense, Middleware Pipeline, Computational Cost Minimization
* **PR / Milestone:** Day 04, Day 10 & Day 11 Architecture Milestone

---

## 1. Context & Problem Statement

The NOVARIS API Gateway handles sensitive mutating financial requests at `POST /v1/payments`. To defend against identity fraud, distributed denial-of-service (DDoS) floods, payload tampering, man-in-the-middle replay attacks, and duplicate double-charging, the gateway incorporates five distinct verification subsystems:
1. RFC 7519 JWT Bearer Token Authentication
2. Per-Merchant Token-Bucket Rate Limiting
3. RFC 2104 HMAC-SHA256 Payload Signature Verification
4. IETF `Idempotency-Key` Deduplication & Cached Response Replay
5. Database-Level OCC Double-Entry Settlement Consensus

Each subsystem incurs a distinct computational, cryptographic, and I/O cost. If middleware layers are misordered, attackers can exploit inexpensive endpoints to inflict massive CPU or database load on downstream systems. 

We must establish a formal, immutable ordering policy for all ingress middleware based on first-principles computational cost minimization and attack surface reduction.

---

## 2. Alternatives Considered

### Option A: Cryptographic Integrity First (HMAC -> Rate Limiting -> JWT)
* **Mechanism:** Verify that the request body is unaltered before checking authentication or rate limits.
* **Why Rejected:**
  - HMAC computation requires reading the entire HTTP body into memory and calculating a cryptographic SHA-256 digest.
  - An unauthenticated attacker sending 100,000 requests per second with 10MB garbage payloads would force the server to allocate gigabytes of RAM and saturate all CPU cores calculating hashes, causing complete gateway outage before the rate limiter or auth validator is reached!

### Option B: Idempotency First (Idempotency -> JWT -> Rate Limiting)
* **Mechanism:** Check the database `idempotency_keys` table immediately to return cached responses as quickly as possible.
* **Why Rejected:**
  - An attacker could flood the gateway with random UUID `Idempotency-Key` headers, forcing the database connection pool to execute millions of `SELECT` queries without ever proving who they are!
  - Unauthenticated callers could probe existing transaction IDs and poison cached state.

### Option C: Strict Five-Layer Cost-Ascending Pipeline (CHOSEN)
* **Mechanism:**
  ```
  Layer 1: JWT Authentication        (Microseconds / Header Only / Stateless)
     │
     ▼
  Layer 2: Token-Bucket Limiter      (Nanoseconds / RAM Math / Throttles DDoS)
     │
     ▼
  Layer 3: HMAC-SHA256 Anti-Tamper   (Milliseconds / Cryptographic Body Digest)
     │
     ▼
  Layer 4: Idempotency Deduplication (Sub-millisecond RAM / Cached DB Replay)
     │
     ▼
  Layer 5: OCC Ledger Settlement     (ACID Multi-Table Consensus / PostgreSQL)
  ```

---

## 3. Decision Outcome

We formally adopt **Option C: Strict Cost-Ascending Five-Layer Defense Pipeline**.

### Layer Specifications & Responsibilities:

1. **Layer 1 — JWT Authentication (`internal/auth/jwt.go`):**
   - *Why First:* Cryptographically establishes caller identity (`sub`, `merchant_id`). Inspects only the `Authorization` header. Does not touch the request body or database. Rejects forged or expired tokens in microseconds (`HTTP 401 Unauthorized`).
2. **Layer 2 — Token-Bucket Rate Limiter (`internal/ratelimit/limiter.go`):**
   - *Why Second:* Now that caller identity is proven, velocity quotas (10 burst, 2 req/s) are evaluated using lock-free mathematical elapsed time in RAM. Volumetric attacks are dropped with `HTTP 429 Too Many Requests` *before* memory is allocated to read the request body.
3. **Layer 3 — HMAC-SHA256 Verification (`internal/auth/hmac.go`):**
   - *Why Third:* Enforces payload integrity and timestamp freshness ($\pm 300\text{s}$) in constant time (`hmac.Equal`). Safely buffers and restores `r.Body` using `io.NopCloser` so downstream layers can re-read the stream.
4. **Layer 4 — Idempotency Deduplication (`services/gateway/idempotency.go`):**
   - *Why Fourth:* Evaluates `Idempotency-Key` and SHA-256 canonical body hash. If the key was previously processed, it intercepts the call and serves the cached receipt with `Idempotent-Replayed: true`. Layer 5 (the ledger database) is never contacted!
5. **Layer 5 — OCC Ledger Settlement (`services/gateway/occ.go`):**
   - *Why Fifth:* Executes the atomic double-entry balance transfer, journal entries, and transactional outbox event. Protected from malicious, throttled, altered, or duplicate requests.

---

## 4. Consequences & Trade-offs

### Positive Consequences:
* **Fail-Fast Efficiency:** Computational exhaustion attacks are neutralized at the earliest, cheapest layer.
* **Zero Database Connection Flooding:** Attackers cannot trigger database lookups without passing both cryptographic identity and velocity gates.
* **Exact-Once Guarantees:** Legitimate retries never incur double-debiting.
* **Separation of Concerns:** Each middleware is isolated, self-contained, and unit-testable.

### Negative Consequences & Mitigations:
* **Stream Buffering Overhead:** Layers 3 and 4 must read `r.Body` into a byte slice.
  * *Mitigation:* Because Layer 2 rate-limits traffic first, memory allocation is strictly bounded by admitted burst quotas.

---

## 5. Compliance & Invariant Verification

1. **Stream Cursor Invariant:**
   `r.Body` must remain readable across all layers. Restored via `io.NopCloser(bytes.NewReader(bodyBytes))`.
2. **Timing Attack Invariant:**
   All HMAC and token signature evaluations MUST use constant-time comparisons (`crypto/subtle.ConstantTimeCompare` or `hmac.Equal`).
3. **Test Proofs:** Formally verified in `services/gateway/integration_test.go` across all 9 contract scenarios.
