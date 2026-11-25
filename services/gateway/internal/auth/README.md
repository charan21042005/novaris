<div align="center">

<!-- Top Banner -->
<p align="center">
  <img src="../../../../assets/banner.svg" alt="NOVARIS Banner" width="100%" />
</p>

# 🛡️ NOVARIS Cryptographic Security & Auth Engine (`/internal/auth`)
### Dual-Layer Identity Verification, Payload Integrity & Replay Attack Defense

<p align="center">
  <img src="https://img.shields.io/badge/Phase_1_Capstone-COMPLETED_(100%25_Verified)-00C853?style=flat-square&logo=checkbox&logoColor=white" alt="Phase 1 Capstone" />
  <img src="https://img.shields.io/badge/Language-Go_1.22+-00ADD8?style=flat-square&logo=go&logoColor=white" alt="Go" />
  <img src="https://img.shields.io/badge/AuthN-JWT_RFC_7519-00C853?style=flat-square&logo=jsonwebtokens&logoColor=white" alt="JWT" />
  <img src="https://img.shields.io/badge/Integrity-HMAC_RFC_2104-7C4DFF?style=flat-square&logo=shield&logoColor=white" alt="HMAC" />
  <img src="https://img.shields.io/badge/Defense-Constant_Time_Equal-FFA000?style=flat-square&logo=vault&logoColor=white" alt="Timing" />
</p>

---

</div>

## 📌 Package Overview

The `internal/auth` package is the cryptographic core of the NOVARIS API Gateway. It enforces a **Two-Pillar Defense-in-Depth Architecture** on all protected payment endpoints:

```
                          INCOMING PAYMENT REQUEST
                                     │
                                     ▼
        ┌─────────────────────────────────────────────────────────┐
        │             PILLAR 1: IDENTITY (Day 2)                  │
        │              auth.JWTMiddleware (jwt.go)                │
        │                                                         │
        │  * Validates Bearer token signature (HMAC-SHA256)       │
        │  * Enforces algorithm pinning (defeats 'alg: none')     │
        │  * Checks expiration (exp) timestamp                    │
        │  * Injects Claims (MerchantID, Role) into r.Context()   │
        └────────────────────────────┬────────────────────────────┘
                                     │ Passes
                                     ▼
        ┌─────────────────────────────────────────────────────────┐
        │             PILLAR 2: INTEGRITY (Day 3)                 │
        │             auth.HMACMiddleware (hmac.go)               │
        │                                                         │
        │  * Validates X-Novaris-Timestamp within ±300s window    │
        │  * Reads raw wire body bytes (Canonicalization Safety)  │
        │  * Restores r.Body via io.NopCloser                     │
        │  * Recomputes HMAC-SHA256 over <timestamp>.<raw_body>   │
        │  * Compares using constant-time hmac.Equal              │
        └────────────────────────────┬────────────────────────────┘
                                     │ Passes
                                     ▼
        ┌─────────────────────────────────────────────────────────┐
        │                   BUSINESS LOGIC                        │
        │             paymentsHandler (main.go)                   │
        │                                                         │
        │  * Consumes verified claims and restored JSON body      │
        │  * Routes transaction to PostgreSQL Ledger & Kafka      │
        └─────────────────────────────────────────────────────────┘
```

---

## 🔐 Cryptographic Modules

### 1. `jwt.go` — Token Generation, Validation & Context Injection

* **Algorithm:** HMAC-SHA256 (`HS256`).
* **Claims Structure (`Claims`):**
  * `MerchantID` (`string`): Opaque identifier for the merchant account.
  * `Role` (`string`): Access control role (`merchant`, `admin`).
  * `RegisteredClaims` (RFC 7519 standard fields: `Subject`, `Issuer`, `IssuedAt`, `ExpiresAt`).
* **Security Defenses:**
  * **`alg: none` Defense:** In `ValidateToken`, the key callback performs an explicit type assertion:
    ```go
    if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
        return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
    }
    ```
    Tokens claiming `alg: none` or RSA public keys are rejected immediately.
  * **Type-Safe Context Keys:** Uses an unexported package-level type `type contextKey string` to prevent key collisions across independent Go packages.

---

### 2. `hmac.go` — Payload Integrity, Replay Defense & Timing Mitigation

* **Algorithm:** HMAC-SHA256 (RFC 2104 / FIPS PUB 198-1).
* **Headers Used:**
  * `X-Novaris-Signature`: Hex-encoded 64-character HMAC digest.
  * `X-Novaris-Timestamp`: UNIX epoch timestamp (seconds) of request generation.
* **Security Defenses:**
  * **Canonicalization Safety:** Hashes the exact raw byte slice directly from the network stream (`io.ReadAll(r.Body)`). Never parses or reformats JSON before hashing, preventing whitespace and float serialization drift.
  * **Stream Restoration:** Reading `r.Body` drains the `io.ReadCloser` to EOF. `HMACMiddleware` restores the body for downstream business handlers using:
    ```go
    r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
    ```
  * **Side-Channel Timing Attack Defense:** Replaces standard short-circuit string comparisons (`==`) with `hmac.Equal` (which runs `crypto/subtle.ConstantTimeCompare`), ensuring constant CPU execution time.
  * **Replay Attack Defense:** Rejects any request where `abs(now - timestamp) > 300` seconds (5 minutes).

---

## 🧪 Automated Unit Test Suite

The package includes 10 automated unit tests across [`jwt_test.go`](./jwt_test.go) and [`hmac_test.go`](./hmac_test.go):

```bash
go test -v ./...
```

### Test Coverage Matrix:
| Test Function | What It Validates |
| :--- | :--- |
| `TestGenerateAndValidateToken` | Mints a valid token, parses claims, and confirms `MerchantID` and `Role` match. |
| `TestExpiredToken` | Generates a token with negative duration and confirms validation returns an expired error. |
| `TestTamperedToken` | Modifies payload bytes and verifies cryptographic signature mismatch rejection. |
| `TestJWTMiddleware` | Verifies HTTP 401 for missing headers and HTTP 200 with context claims for valid tokens. |
| `TestComputeAndVerifyHMAC_Success` | Generates valid HMAC signature and verifies successful verification. |
| `TestVerifyHMAC_TamperedBody` | Generates signature for original body, sends tampered body, verifies rejection. |
| `TestVerifyHMAC_StaleTimestamp` | Generates request with 10-minute old timestamp, verifies replay rejection. |
| `TestVerifyHMAC_FutureTimestamp` | Generates request with 10-minute future timestamp, verifies clock drift rejection. |
| `TestVerifyHMAC_WrongSecret` | Signs with attacker secret, verifies rejection against server secret. |
| `TestHMACMiddleware` | Verifies end-to-end middleware execution, 401 on tampered/stale requests, and stream restoration. |

---

<div align="center">

<!-- Bottom Banner -->
<p align="center">
  <img src="../../../../assets/footer.svg" alt="NOVARIS Footer" width="100%" />
</p>

### 🎓 NOVARIS Cryptographic Engine
*RFC 7519 • RFC 2104 • Defense in Depth*

</div>
