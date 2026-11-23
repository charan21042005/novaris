<div align="center">

<!-- Top Banner -->
<p align="center">
  <img src="../../assets/banner.svg" alt="NOVARIS Banner" width="100%" />
</p>

# 🚪 NOVARIS • API Gateway Service (`/services/gateway`)
### High-Throughput Request Ingestion, Rate Limiting & Cryptographic Verification

<p align="center">
  <img src="https://img.shields.io/badge/Language-Go_1.22+-00ADD8?style=flat-square&logo=go&logoColor=white" alt="Go" />
  <img src="https://img.shields.io/badge/Security-JWT_HMAC--SHA256-00C853?style=flat-square&logo=jsonwebtokens&logoColor=white" alt="Security" />
  <img src="https://img.shields.io/badge/Port-8080-7C4DFF?style=flat-square&logo=gnubash&logoColor=white" alt="Port" />
  <img src="https://img.shields.io/badge/Resilience-Graceful_Shutdown-FFA000?style=flat-square&logo=shield&logoColor=white" alt="Resilience" />
</p>

---

</div>

## 📌 Service Overview

The **API Gateway** is the single entry point for all external merchants, payment flows, and client applications interacting with NOVARIS. It sits on the network perimeter and guarantees that no invalid, duplicate, unauthenticated, or tampered traffic touches downstream event streams or financial ledgers.

### Core Responsibilities:
* **Liveness & Readiness Probing:** Exposes standardized `/healthz` endpoints for Kubernetes container orchestration.
* **Cryptographic Identity Verification (Day 2):** Enforces RFC 7519 JSON Web Token (JWT) validation with defenses against the `alg: none` exploit and algorithm confusion attacks.
* **Structured Observability:** Middleware tracking request method, path, remote IP, and microsecond processing latency.
* **Graceful Termination:** Intercepts OS signals (`SIGINT`, `SIGTERM`) and allows in-flight connections to drain cleanly with a 5-second deadline.
* *(Upcoming Day 3–4):* HMAC-SHA256 request body signing (payload integrity) and distributed sliding token-bucket rate limiting via Redis.

---

## 📡 HTTP API Specification

### 1. `GET /healthz` — Service Health & Probing (Public)
Returns the operational health and environment metadata of the gateway.

* **Response (`200 OK`):**
```json
{
  "status": "ok",
  "service": "novaris-gateway",
  "version": "1.0.0",
  "environment": "development",
  "timestamp": "2026-11-23T06:33:11.717Z"
}
```

### 2. `GET /` — API Discovery (Public)
Provides service discovery details and listed routes.

### 3. `GET /v1/payments` — Payment Ingestion (Protected)
Requires a valid, signed Bearer token in the `Authorization` header.

* **Headers Required:**
  ```http
  Authorization: Bearer <valid_jwt_token>
  ```
* **Success Response (`200 OK`):**
```json
{
  "status": "authorized",
  "merchant_id": "merchant_acme_corp",
  "role": "merchant_admin",
  "message": "Access granted to protected payments route. Identity cryptographically verified via JWT.",
  "timestamp": "2026-11-23T06:33:11.608Z"
}
```
* **Failure Responses (`401 Unauthorized`):**
  * Missing header: `{"error": "Unauthorized", "message": "Missing Authorization header..."}`
  * Tampered token: `{"error": "Unauthorized", "message": "Authentication failed: invalid token..."}`
  * Expired token: `{"error": "Unauthorized", "message": "Authentication failed: token has expired..."}`

---

## 🛠️ Testing & Generating Tokens

A CLI token generator is provided in `cmd/gentoken` to create test tokens on demand:

```bash
# Generate a standard valid token (1 hour duration)
go run cmd/gentoken/main.go -merchant merchant_acme -role admin

# Generate an expired token (for testing 401 rejection)
go run cmd/gentoken/main.go -merchant merchant_test -expired

# Run automated unit test suite
go test -v ./...
```

---

<div align="center">

<p align="center">
  <img src="../../assets/footer.svg" alt="NOVARIS Footer" width="100%" />
</p>

</div>
