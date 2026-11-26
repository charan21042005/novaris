<div align="center">

<!-- Top Banner -->
<p align="center">
  <img src="../../../../assets/banner.svg" alt="NOVARIS Banner" width="100%" />
</p>

# 🔑 NOVARIS JWT Minting Utility (`/cmd/gentoken`)
### Developer Token Generator, RFC 7519 HMAC-SHA256 Signer & Security Test Harness

<p align="center">
  <img src="https://img.shields.io/badge/Phase_1_Capstone-COMPLETED_(100%25_Verified)-00C853?style=flat-square&logo=checkbox&logoColor=white" alt="Phase 1 Capstone" />
  <img src="https://img.shields.io/badge/Language-Go_1.22+-00ADD8?style=flat-square&logo=go&logoColor=white" alt="Go" />
  <img src="https://img.shields.io/badge/Standard-RFC_7519_JWT-00C853?style=flat-square&logo=jsonwebtokens&logoColor=white" alt="JWT" />
  <img src="https://img.shields.io/badge/Algorithm-HMAC--SHA256_(HS256)-7C4DFF?style=flat-square&logo=shield&logoColor=white" alt="Algorithm" />
  <img src="https://img.shields.io/badge/Testing-Negative_Expired_Tokens-FFA000?style=flat-square&logo=snyk&logoColor=white" alt="Testing" />
</p>

---

</div>

## 📌 Tool Overview

`gentoken` is a developer and operator command-line utility introduced in **Day 2**. It mints cryptographically signed JSON Web Tokens (RFC 7519) on demand using HMAC-SHA256 (`HS256`).

The utility generates:
* **Standard Merchant Tokens:** Authenticates payment ingress requests with embedded `merchant_id` and role claims.
* **Privileged Admin Tokens:** Generates credentials for administrative endpoints, audit reviews, and manual settlements.
* **Deliberately Expired Tokens:** Mints tokens expired in the past (e.g. 10 minutes prior) for negative security testing and verifying HTTP 401 Unauthorized gateway rejections.

```
       ┌─────────────────────────────────────────────────────────────┐
       │              cmd/gentoken Minting Execution                 │
       │                                                             │
       │  Inputs:                                                    │
       │    -merchant merchant_alpha_001                             │
       │    -role merchant                                           │
       │    -mins 60                                                 │
       │    -secret (from JWT_SECRET env var)                        │
       └──────────────────────────────┬──────────────────────────────┘
                                      │
                                      ▼
       ┌─────────────────────────────────────────────────────────────┐
       │                 JWT Payload (RFC 7519)                      │
       │  {                                                          │
       │    "sub": "merchant_alpha_001",                             │
       │    "merchant_id": "merchant_alpha_001",                     │
       │    "role": "merchant",                                      │
       │    "iss": "novaris-gateway",                                │
       │    "iat": 1732617600,                                       │
       │    "exp": 1732621200                                        │
       │  }                                                          │
       └──────────────────────────────┬──────────────────────────────┘
                                      │ Signs via HMAC-SHA256
                                      ▼
       eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOi... (Bearer Token)
```

---

## ⚙️ CLI Flags & Parameters

| Flag | Type | Default | Description |
| :--- | :---: | :---: | :--- |
| `-merchant` | `string` | `merchant_123` | Merchant identifier embedded into `merchant_id` and `sub` claims. |
| `-role` | `string` | `merchant` | Access control role (`merchant`, `admin`, `auditor`). |
| `-mins` | `int` | `60` | Validity duration in minutes. |
| `-expired` | `bool` | `false` | Generates a deliberately expired token (10 minutes in the past) for negative testing. |

---

## 🚀 Execution & Usage Examples

### 1. Generate Standard Merchant Token (1-Hour TTL)
```powershell
Set-Location "services/gateway"
go run ./cmd/gentoken -merchant merchant_stripe_alpha -role merchant -mins 60
```

### 2. Generate Admin Token for Operations
```powershell
go run ./cmd/gentoken -merchant admin_sec_ops -role admin -mins 120
```

### 3. Generate Expired Token for Security Perimeter Verification
```powershell
go run ./cmd/gentoken -merchant merchant_test -expired
```

### Sample Output:
```text
================================================================================
   NOVARIS • JWT BEARER TOKEN GENERATOR                                         
================================================================================
Merchant ID: merchant_stripe_alpha
Role:        merchant
Expires In:  60 minute(s)
Algorithm:   HS256 (HMAC-SHA256)

TOKEN:
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJtZXJjaGFudF9pZCI6Im1lcmNoYW50X3N0cmlwZV9hbHBoYSIsInJvbGUiOiJtZXJjaGFudCIsImlzcyI6Im5vdmFyaXMtZ2F0ZXdheSIsInN1YiI6Im1lcmNoYW50X3N0cmlwZV9hbHBoYSIsImV4cCI6MTczMjYyMTIwMCwiaWF0IjoxNzMyNjE3NjAwfQ.d7K...

CURL AUTHORIZATION HEADER:
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

---

## 🔒 Security Best Practices

1. **Environment Secret Injection:** `gentoken` reads the signing key from the `JWT_SECRET` environment variable. If unset, it uses the development default key (`novaris_dev_secret_key_change_in_production_min32b`).
2. **Production Safety:** Real production signing keys must never be hardcoded or logged to disk. Always generate tokens inside secure bastion hosts or automated CI/CD runners.

---

## 🔗 Related Components

* 🛡️ **[internal/auth](../../internal/auth/README.md):** Gateway authentication package that verifies tokens minted by this utility.
* 🛠️ **[cmd/signrequest](../signrequest/README.md):** Companion utility for computing payload HMAC signatures.
* 🔍 **[cmd/e2eprobe](../e2eprobe/README.md):** Automated probe that uses minted tokens for end-to-end integration tests.

---

<div align="center">

<!-- Bottom Banner -->
<p align="center">
  <img src="../../../../assets/footer.svg" alt="NOVARIS Footer" width="100%" />
</p>

### 🎓 NOVARIS API Gateway Developer Tooling
*RFC 7519 JWT Minting Utility • HS256 Authentication*

</div>
