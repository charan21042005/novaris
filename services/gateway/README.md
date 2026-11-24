<div align="center">

<!-- Top Banner -->
<p align="center">
  <img src="../../assets/banner.svg" alt="NOVARIS Banner" width="100%" />
</p>

# 🚪 NOVARIS • API Gateway Service (`/services/gateway`)
### High-Throughput Request Ingestion, Rate Limiting & Cryptographic Verification

<p align="center">
  <img src="https://img.shields.io/badge/Language-Go_1.22+-00ADD8?style=flat-square&logo=go&logoColor=white" alt="Go" />
  <img src="https://img.shields.io/badge/Security-JWT_+_HMAC--SHA256-00C853?style=flat-square&logo=jsonwebtokens&logoColor=white" alt="Security" />
  <img src="https://img.shields.io/badge/Rate_Limiter-Token_Bucket_[10_burst,_2/s]-FF6D00?style=flat-square&logo=speedtest&logoColor=white" alt="Rate Limiter" />
  <img src="https://img.shields.io/badge/Port-8080-7C4DFF?style=flat-square&logo=gnubash&logoColor=white" alt="Port" />
  <img src="https://img.shields.io/badge/Resilience-Graceful_Shutdown-FFA000?style=flat-square&logo=shield&logoColor=white" alt="Resilience" />
</p>

---

</div>

## 📌 Service Overview

The **API Gateway** is the single ingress point for all external merchants, payment flows, and client applications interacting with NOVARIS. It sits on the network perimeter and guarantees that no invalid, duplicate, unauthenticated, or tampered traffic touches downstream event streams or financial ledgers.

### Multi-Layer Security Architecture (Defense in Depth):
1. **Liveness & Readiness Probing:** Exposes standardized `/healthz` endpoints for Kubernetes container orchestration.
2. **Cryptographic Identity Verification (Day 2):** Enforces RFC 7519 JSON Web Token (JWT) validation with defenses against the `alg: none` exploit and algorithm confusion attacks.
3. **In-Memory Token-Bucket Rate Limiter (Day 4):** Enforces per-merchant rate quotas (Burst: 10 requests, Sustained: 2 req/sec) using zero-goroutine lazy mathematical refill. Responds with HTTP `429 Too Many Requests` + `Retry-After: <sec>`. *Fail-Fast Optimization:* Drops floods before body hashing.
4. **Payload Integrity & Replay Defense (Day 3):** Enforces HMAC-SHA256 request body signing with `X-Novaris-Timestamp` (±300s replay window) and constant-time signature comparison (`hmac.Equal`) to neutralize side-channel timing attacks.
5. **Structured Observability:** Middleware tracking request method, path, remote IP, and microsecond processing latency.
6. **Graceful Termination:** Intercepts OS signals (`SIGINT`, `SIGTERM`) and allows in-flight connections to drain cleanly with a 5-second deadline.
7. *(Upcoming Day 5):* Immutable double-entry financial ledger schema & database migrations.

---

## 📡 HTTP API Specification

### 1. `GET /healthz` — Service Health & Probing (Public)
Returns the operational health and environment metadata of the gateway.

* **Response (`200 OK`):**
```json
{
  "status": "ok",
  "service": "novaris-gateway",
  "version": "0.3.0-day3",
  "environment": "development",
  "timestamp": "2026-11-23T06:45:11.717Z"
}
```

### 2. `GET /` — API Discovery (Public)
Provides service discovery details and listed routes.

### 3. `POST /v1/payments` — Payment Ingestion (Protected - Layered Defense)
Requires valid credentials across both security layers:
1. **Identity:** `Authorization: Bearer <valid_jwt>` (Proves WHO is calling).
2. **Integrity & Freshness:** `X-Novaris-Timestamp` + `X-Novaris-Signature` (Proves WHAT was sent hasn't been altered or replayed).

* **Headers Required:**
  ```http
  Authorization: Bearer <jwt_token>
  X-Novaris-Timestamp: <unix_epoch_seconds>
  X-Novaris-Signature: <hmac_sha256_hex_digest>
  Content-Type: application/json
  ```

* **Example JSON Request Body:**
  ```json
  {
    "amount_cents": 5000,
    "recipient": "merchant_456"
  }
  ```

* **Success Response (`200 OK`):**
```json
{
  "status": "authorized",
  "merchant_id": "merchant_123",
  "role": "merchant",
  "amount_cents": 5000,
  "recipient": "merchant_456",
  "message": "Payment request verified successfully: JWT identity confirmed, rate limit quota allocated, and HMAC body integrity preserved.",
  "timestamp": "2026-11-24T12:50:11.608Z"
}
```

* **Rate Limit Exceeded Response (`429 Too Many Requests`):**
  * Emitted when a merchant exceeds their burst capacity ($c=10.0$) and sustains traffic above refill ($r=2.0\text{ tokens/sec}$).
  * Includes RFC 6585 standard `Retry-After` header indicating minimum seconds to wait before retrying.
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

* **Authentication & Integrity Failure Responses (`401 Unauthorized`):**
  * Missing Bearer token: `{"error": "unauthorized", "message": "missing Authorization header"}`
  * Missing HMAC headers: `{"error": "unauthorized", "auth_type": "hmac_sha256", "message": "missing required signature headers..."}`
  * Tampered request body: `{"error": "unauthorized", "auth_type": "hmac_sha256", "message": "HMAC signature verification failed: invalid signature: payload hash mismatch"}`
  * Stale/Replayed timestamp: `{"error": "unauthorized", "auth_type": "hmac_sha256", "message": "request timestamp outside allowed window..."}`

---

## 🛠️ Developer CLI Utilities

### 1. JWT Minting Utility (`cmd/gentoken`)
```bash
# Generate a standard valid token (15-minute lifespan)
go run cmd/gentoken/main.go -merchant merchant_123 -role merchant -mins 15

# Generate an expired token (for testing 401 rejection)
go run cmd/gentoken/main.go -merchant merchant_test -expired
```

### 2. HMAC Request Signing Utility (`cmd/signrequest`)
```bash
# Compute HMAC signature for default body and output curl command
go run cmd/signrequest/main.go

# Compute HMAC signature for a custom JSON body
go run cmd/signrequest/main.go -body '{"amount_cents":10000,"recipient":"merchant_xyz"}'

# Generate a stale timestamp (10m in the past) to test replay rejection
go run cmd/signrequest/main.go -offset -600

# Simulate payload tampering (computes signature on original, emits tampered body)
go run cmd/signrequest/main.go -tamper
```

### 3. Automated Unit Test Suite
```bash
go test -v ./...
```

---

## 📂 Subdirectory Architecture & Technical Documentation

* 👉 **[`cmd/` CLI Documentation](./cmd/README.md):** Developer CLI suite (`gentoken` and `signrequest`) for local token minting and signature testing.
* 👉 **[`internal/` Engine Documentation](./internal/README.md):** Go compiler-enforced private package boundary.
* 👉 **[`internal/auth/` Cryptographic Specs](./internal/auth/README.md):** Cryptographic JWT & HMAC-SHA256 engine specifications and test matrices.
* 👉 **[`internal/ratelimit/` Rate Limiter Specs](./internal/ratelimit/README.md):** Token-bucket rate limiter mathematical model, concurrency architecture, and test suites.

---

<div align="center">

<p align="center">
  <img src="../../assets/footer.svg" alt="NOVARIS Footer" width="100%" />
</p>

</div>
