<div align="center">

<!-- Top Banner -->
<p align="center">
  <img src="../../../../assets/banner.svg" alt="NOVARIS Banner" width="100%" />
</p>

# 📑 NOVARIS OpenAPI Contract Validator (`/cmd/verifyopenapi`)
### Automated OAS 3.0.3 Conformance Auditing, Status Code Taxonomy & Model Validation

<p align="center">
  <img src="https://img.shields.io/badge/Phase_1_Capstone-COMPLETED_(100%25_Verified)-00C853?style=flat-square&logo=checkbox&logoColor=white" alt="Phase 1 Capstone" />
  <img src="https://img.shields.io/badge/Language-Go_1.22+-00ADD8?style=flat-square&logo=go&logoColor=white" alt="Go" />
  <img src="https://img.shields.io/badge/Standard-OpenAPI_3.0.3-6BA539?style=flat-square&logo=openapiinitiative&logoColor=white" alt="OAS 3.0" />
  <img src="https://img.shields.io/badge/Status_Codes-7_HTTP_Status_Codes-7C4DFF?style=flat-square&logo=cloudflare&logoColor=white" alt="Status Codes" />
  <img src="https://img.shields.io/badge/CI%2FCD-Automated_Gatekeeper-FFA000?style=flat-square&logo=githubactions&logoColor=white" alt="CI/CD" />
</p>

---

</div>

## 📌 Tool Overview

`verifyopenapi` is an automated contract validation gatekeeper introduced in **Day 12**. It statically inspects the machine-readable OpenAPI specification file (`api/openapi.json` or `services/gateway/internal/openapi/openapi.json`) to enforce compliance with NOVARIS API governance standards before deployment.

The utility validates:
1. **OAS Version Specification:** Asserts OpenAPI specification version is `3.0.x`.
2. **Platform Ingress Coverage:** Confirms mandatory declaration of all 5 gateway routes:
   - `/healthz` (Liveness and dependency readiness).
   - `/` (Root service discovery manifest).
   - `/openapi.json` (Raw specification export).
   - `/docs` (Interactive Swagger UI documentation).
   - `/v1/payments` (Core payment settlement ingress).
3. **Status Code Taxonomy:** Enforces complete documentation for all 7 standard gateway HTTP response status codes:
   - `200 OK` (Successful settlement or idempotent replay).
   - `400 Bad Request` (Missing Idempotency-Key or malformed JSON).
   - `401 Unauthorized` (Missing or invalid JWT Bearer token / signature).
   - `409 Conflict` (Optimistic concurrency version race condition).
   - `422 Unprocessable Entity` (Idempotent payload mismatch or domain invalidity).
   - `429 Too Many Requests` (Token-bucket rate limit exceeded).
   - `500 Internal Server Error` (Database failure or unhandled exception).
4. **Schema Component Models:** Verifies data definitions for `PaymentRequest`, `PaymentReceipt`, `HealthResponse`, `RootDiscoveryResponse`, and `ErrorResponse`.

```
          OpenAPI Specification File (api/openapi.json)
                                │
                                ▼
         ┌──────────────────────────────────────────────┐
         │       verifyopenapi Contract Gatekeeper      │
         └──────────────────────┬───────────────────────┘
                                │
         ┌──────────────────────┴──────────────────────┐
         ▼                      ▼                      ▼
  Version & Info         Route Inventory        Status Codes & Models
  - OAS 3.0.3            - /healthz             - 200, 400, 401, 409
  - Title, Version       - /                    - 422, 429, 500
  - Servers, Licenses    - /openapi.json        - PaymentRequest
                         - /docs                - PaymentReceipt
                         - /v1/payments         - ErrorResponse
```

---

## ⚙️ CLI Flags & Parameters

| Flag | Type | Default | Description |
| :--- | :---: | :---: | :--- |
| `-spec` | `string` | `""` | Optional explicit path to an `openapi.json` file. If omitted, checks standard repository paths. |

---

## 🚀 Execution & Usage Examples

### 1. Validate Default Repository OpenAPI Specification
```powershell
Set-Location "services/gateway"
go run ./cmd/verifyopenapi
```

### 2. Validate Custom Specification File
```powershell
go run ./cmd/verifyopenapi -spec ../../api/openapi.json
```

### Sample Output:
```text
================================================================================
   NOVARIS • OPENAPI 3.0 CONTRACT VERIFICATION VALIDATOR                        
================================================================================
Target Spec: api/openapi.json

[1/4] Checking OpenAPI version...
  ✔ Valid OpenAPI version: 3.0.3

[2/4] Verifying required metadata...
  ✔ Found title: NOVARIS Distributed Payment Settlement API
  ✔ Found version: 1.0.0

[3/4] Checking mandatory ingress endpoints...
  ✔ Found: /healthz
  ✔ Found: /
  ✔ Found: /openapi.json
  ✔ Found: /docs
  ✔ Found: /v1/payments

[4/4] Auditing HTTP status code taxonomy & schemas...
  ✔ Found HTTP 200, 400, 401, 409, 422, 429, 500
  ✔ Schema components verified: PaymentRequest, PaymentReceipt, ErrorResponse

================================================================================
   OUTCOME: CONTRACT FULLY COMPLIANT WITH NOVARIS STANDARDS                    
================================================================================
```

---

## 🧪 Automated Testing

Unit tests for `verifyopenapi` ensure the validator correctly flags malformed or incomplete specs:

```powershell
Set-Location "services/gateway"
go test -v ./cmd/verifyopenapi/...
```

---

## 🔗 Related Components

* 📑 **[internal/openapi](../../internal/openapi/README.md):** Gateway package that embeds and serves the verified specification.
* 🌐 **[api/](../../../../api/README.md):** Repository root directory hosting the canonical OpenAPI JSON document.
* 🔍 **[cmd/phase1audit](../phase1audit/README.md):** Capstone audit tool that runs OpenAPI contract checks.

---

<div align="center">

<!-- Bottom Banner -->
<p align="center">
  <img src="../../../../assets/footer.svg" alt="NOVARIS Footer" width="100%" />
</p>

### 🎓 NOVARIS API Gateway Developer Tooling
*OpenAPI 3.0.3 Contract Verification Gatekeeper*

</div>
