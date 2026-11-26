<div align="center">

<!-- Top Banner -->
<p align="center">
  <img src="../../../../assets/banner.svg" alt="NOVARIS Banner" width="100%" />
</p>

# 🔍 NOVARIS End-to-End API Probe (`/cmd/e2eprobe`)
### Automated Black-Box Contract Verification, Cryptographic Handshakes & Idempotency Audit

<p align="center">
  <img src="https://img.shields.io/badge/Phase_1_Capstone-COMPLETED_(100%25_Verified)-00C853?style=flat-square&logo=checkbox&logoColor=white" alt="Phase 1 Capstone" />
  <img src="https://img.shields.io/badge/Language-Go_1.22+-00ADD8?style=flat-square&logo=go&logoColor=white" alt="Go" />
  <img src="https://img.shields.io/badge/Role-E2E_Black--Box_Runner-00ADD8?style=flat-square&logo=curl&logoColor=white" alt="Role" />
  <img src="https://img.shields.io/badge/Security-Dual--Pillar_JWT_+_HMAC-7C4DFF?style=flat-square&logo=shield&logoColor=white" alt="Security" />
  <img src="https://img.shields.io/badge/Coverage-7_Stage_Contract_Suite-FFA000?style=flat-square&logo=checkmarx&logoColor=white" alt="Coverage" />
</p>

---

</div>

## 📌 Tool Overview

`e2eprobe` is an automated black-box test harness and diagnostic runner introduced in **Day 11**. It acts as a realistic client SDK, firing a sequential battery of 7 HTTP transactions against a live NOVARIS API Gateway instance.

The tool tests:
1. **Unauthenticated Public Endpoints:** Verifies `/healthz` and `/` root discovery.
2. **Negative Authentication Defense:** Ensures requests without authorization headers are terminated with `HTTP 401 Unauthorized`.
3. **Cryptographic Settlement Path:** Mints a fresh RFC 7519 JWT, calculates a canonical RFC 2104 HMAC-SHA256 signature, generates an `Idempotency-Key` UUID, and executes a real settlement payment (`HTTP 200 OK`).
4. **Idempotent Network Retry:** Re-submits the exact same transaction payload and `Idempotency-Key`, verifying that the gateway detects deduplication, avoids balance double-debiting, and returns the cached receipt with `Idempotent-Replayed: true`.
5. **Payload Tampering Detection:** Submits a modified transaction payload with a recycled `Idempotency-Key`, ensuring the gateway terminates the attack with `HTTP 422 Unprocessable Entity`.
6. **Missing Header Validation:** Submits a payment without an `Idempotency-Key`, asserting fail-fast termination with `HTTP 400 Bad Request`.

```
       ┌─────────────────────────────────────────────────────────────┐
       │                e2eprobe Automated Test Runner               │
       └──────────────────────────────┬──────────────────────────────┘
                                      │
     [1/7] GET /healthz               │ ──> HTTP 200 (Liveness Check)
     [2/7] GET /                      │ ──> HTTP 200 (Root Service Discovery)
     [3/7] POST /v1/payments (No Auth)│ ──> HTTP 401 (Defense Perimeter Check)
     [4/7] POST /v1/payments (Valid)  │ ──> HTTP 200 (Full Settlement Path)
     [5/7] POST /v1/payments (Retry)  │ ──> HTTP 200 (Idempotent-Replayed: true)
     [6/7] POST /v1/payments (Mismatch│ ──> HTTP 422 (Key Conflict Rejection)
     [7/7] POST /v1/payments (No Key) │ ──> HTTP 400 (Bad Request Rejection)
                                      │
                                      ▼
       ┌─────────────────────────────────────────────────────────────┐
       │                Live NOVARIS Gateway (Port 8080)             │
       └─────────────────────────────────────────────────────────────┘
```

---

## ⚙️ CLI Flags & Parameters

| Flag | Type | Default | Description |
| :--- | :---: | :---: | :--- |
| `-url` | `string` | `http://localhost:8080` | Target gateway base URL to probe. |
| `-merchant` | `string` | `c0000000-0000-0000-0000-000000000001` | Payer merchant account UUID. |
| `-recipient` | `string` | `c0000000-0000-0000-0000-000000000002` | Payee recipient account UUID. |
| `-amount` | `int64` | `5000` | Payment magnitude in integer cents ($50.00 = 5000). |

---

## 🚀 Execution & Usage Examples

### 1. Running Probe Against Local Gateway
```powershell
Set-Location "services/gateway"
go run ./cmd/e2eprobe
```

### Sample Output:
```text
================================================================================
   NOVARIS • DAY 11: END-TO-END GATEWAY CONTRACT VERIFICATION PROBE            
================================================================================
Target URL:  http://localhost:8080
Merchant ID: c0000000-0000-0000-0000-000000000001
Recipient:   c0000000-0000-0000-0000-000000000002
Amount:      5000 cents ($50.00)

[1/7] Probing /healthz endpoint...
  ✔ PASS: /healthz responded HTTP 200 (status: healthy, db: reachable, kafka: ready)

[2/7] Probing / discovery endpoint...
  ✔ PASS: / responded HTTP 200 with valid service manifest

[3/7] Probing security perimeter: POST /v1/payments without credentials...
  ✔ PASS: Unauthenticated request rejected with HTTP 401 Unauthorized

[4/7] Submitting authentic payment request (JWT + HMAC signed)...
  ✔ PASS: Payment accepted with HTTP 200 OK (Tx ID: 8f9b..., Status: SETTLED)

[5/7] Submitting idempotent retry with identical key and payload...
  ✔ PASS: Idempotent replay returned HTTP 200 OK (Idempotent-Replayed: true)

[6/7] Submitting conflicting payload with reused idempotency key...
  ✔ PASS: Conflict rejected with HTTP 422 Unprocessable Entity

[7/7] Submitting payment without Idempotency-Key header...
  ✔ PASS: Missing key rejected with HTTP 400 Bad Request

================================================================================
   OUTCOME: ALL 7 PROBES PASSED (Gateway contract is 100% compliant)           
================================================================================
```

---

## 🧪 Automated Testing

Unit tests for the probe runner verify mock HTTP interactions without requiring a running gateway:

```powershell
Set-Location "services/gateway"
go test -v ./cmd/e2eprobe/...
```

---

## 🔗 Related Components

* 🛡️ **[internal/auth](../../internal/auth/README.md):** Cryptographic auth middleware validated by probes 3, 4, and 5.
* 🛠️ **[cmd/gentoken](../gentoken/README.md):** Token generation utility mirrored internally by the probe runner.
* 🛠️ **[cmd/signrequest](../signrequest/README.md):** HMAC request signing utility mirrored internally by the probe runner.

---

<div align="center">

<!-- Bottom Banner -->
<p align="center">
  <img src="../../../../assets/footer.svg" alt="NOVARIS Footer" width="100%" />
</p>

### 🎓 NOVARIS API Gateway Developer Tooling
*Automated Black-Box Probe Runner • Contract Compliance*

</div>
