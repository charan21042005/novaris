<div align="center">

<!-- Top Banner -->
<p align="center">
  <img src="../../../assets/banner.svg" alt="NOVARIS Banner" width="100%" />
</p>

# 🛠️ NOVARIS Gateway CLI Suite (`/services/gateway/cmd`)
### Developer Utilities, Cryptographic Token Generation & Request Signing Tools

<p align="center">
  <img src="https://img.shields.io/badge/Language-Go_1.22+-00ADD8?style=flat-square&logo=go&logoColor=white" alt="Go" />
  <img src="https://img.shields.io/badge/Pattern-cmd_Standard_Layout-00C853?style=flat-square&logo=gitbook&logoColor=white" alt="Pattern" />
  <img src="https://img.shields.io/badge/Security-JWT_+_HMAC--SHA256-7C4DFF?style=flat-square&logo=jsonwebtokens&logoColor=white" alt="Security" />
  <img src="https://img.shields.io/badge/Environment-Local_Dev_%26_CI-FFA000?style=flat-square&logo=gnubash&logoColor=white" alt="Env" />
</p>

---

</div>

## 📌 Directory Purpose & Go Project Layout

The `cmd/` directory adheres to the **Standard Go Project Layout** (`golang-standards/project-layout`). In idiomatic Go architecture:
* Each subdirectory under `cmd/` represents an **independent executable binary** (package `main`).
* Code inside `cmd/` acts as thin client or CLI wrappers that import reusable business and cryptographic logic from `internal/`.
* Code here is designed for developer testing, automated continuous integration (CI) probes, and operational troubleshooting.

---

## 🧰 Available CLI Utilities

```
services/gateway/cmd/
├── gentoken/         # Mints cryptographically signed JWT Bearer tokens
│   └── main.go
├── signrequest/      # Computes canonical HMAC-SHA256 signatures over request bodies
│   └── main.go
├── verifyschema/     # Verifies database schema, composite indexes & migration compliance
│   ├── main.go
│   └── main_test.go
└── README.md         # This documentation
```

---

## 1. `gentoken` — Developer JWT Minting Engine

Located at [`cmd/gentoken/main.go`](./gentoken/main.go), this utility generates signed HMAC-SHA256 JSON Web Tokens (RFC 7519) on demand.

### Supported CLI Flags:
| Flag | Type | Default | Description |
| :--- | :---: | :---: | :--- |
| `-merchant` | `string` | `merchant_123` | Merchant identifier embedded in the token payload (`merchant_id` and `sub`). |
| `-role` | `string` | `merchant` | Access control role (`merchant`, `admin`, `settlement_worker`). |
| `-mins` | `int` | `60` | Validity duration in minutes. |
| `-expired` | `bool` | `false` | Generates a deliberately expired token (10 minutes in the past) for negative testing. |

### Usage Examples:
```bash
# Generate a standard merchant token valid for 15 minutes
go run cmd/gentoken/main.go -merchant merchant_stripe_live -role merchant -mins 15

# Generate an admin token for audit inspection
go run cmd/gentoken/main.go -merchant merchant_super_admin -role admin

# Generate a deliberately expired token to verify HTTP 401 Unauthorized handling
go run cmd/gentoken/main.go -expired
```

---

## 2. `signrequest` — Developer HMAC-SHA256 Request Signer

Located at [`cmd/signrequest/main.go`](./signrequest/main.go), this utility simulates a merchant client SDK by constructing the canonical payload (`<timestamp>.<raw_body>`) and computing the hex-encoded HMAC-SHA256 digest using the shared secret.

### Supported CLI Flags:
| Flag | Type | Default | Description |
| :--- | :---: | :---: | :--- |
| `-body` | `string` | `{"amount_cents":5000,"recipient":"merchant_456"}` | Raw JSON request body to sign. |
| `-offset` | `int64` | `0` | Offset in seconds applied to the timestamp (e.g. `-600` simulates a 10-minute stale replay attack). |
| `-secret` | `string` | `""` | Custom HMAC secret (defaults to `HMAC_SECRET` env var or dev fallback). |
| `-tamper` | `bool` | `false` | Computes signature over original body, but emits a modified body to verify tamper rejection. |

### Usage Examples:
```bash
# Compute HMAC signature for default payload and print ready-to-run curl command
go run cmd/signrequest/main.go

# Sign a custom settlement payload
go run cmd/signrequest/main.go -body '{"amount_cents":100000,"recipient":"vendor_xyz"}'

# Simulate a Replay Attack (generates timestamp 600s in the past)
go run cmd/signrequest/main.go -offset -600

# Simulate In-Flight Tampering (computes signature on $50, outputs $50,000 payload)
go run cmd/signrequest/main.go -tamper
```

---

## 3. `verifyschema` — Database Schema & Migration Verifier

Located at [`cmd/verifyschema/main.go`](./verifyschema/main.go), this utility validates that the live PostgreSQL instance or offline SQL migration files strictly match the NOVARIS SRS Capstone Blueprint:
* Confirms the existence of all 7 core tables: `accounts`, `transactions`, `journal_entries`, `outbox`, `fraud_reviews`, `feature_snapshots`, and `idempotency_keys` (Day 10).
* Checks column data types, JSONB binary storage format, and CHECK constraints.
* Validates composite B-Tree indexes (`idx_feature_snapshots_account_time`, `idx_idemp_expires_at`).

### Supported CLI Flags:
| Flag | Type | Default | Description |
| :--- | :---: | :---: | :--- |
| `-offline` | `bool` | `false` | Executes deterministic static analysis over SQL migration files without requiring a running database. |
| `-sql` | `string` | `""` | Optional explicit path to an individual SQL migration file to inspect. |

### Usage Examples:
```bash
# Run static migration verification (ideal for CI/CD and offline development)
go run cmd/verifyschema/main.go -offline

# Verify live schema against running PostgreSQL container
go run cmd/verifyschema/main.go

# Run automated Go unit test suite
go test -v ./cmd/verifyschema
```

---

## 🔒 Security Best Practices for Local Development

1. **Fallback Secrets:** If `JWT_SECRET` or `HMAC_SECRET` environment variables are omitted, both utilities fall back to predefined development keys.
2. **Production Isolation:** In staging and production environments, real merchant secrets **must never be committed to source code or hardcoded**. Secrets must be injected via AWS Secrets Manager, HashiCorp Vault, or Kubernetes Secrets.
3. **Canonical Consistency:** Both `signrequest` and the gateway's `HMACMiddleware` compute signatures over the exact byte sequence `<timestamp>.<raw_body>`. Modifying whitespace, key order, or line endings in transit breaks canonical alignment.

---

<div align="center">

<!-- Bottom Banner -->
<p align="center">
  <img src="../../../assets/footer.svg" alt="NOVARIS Footer" width="100%" />
</p>

### 🎓 NOVARIS API Gateway Developer Tooling
*Standard Go Project Layout • Cryptographic Utilities*

</div>
