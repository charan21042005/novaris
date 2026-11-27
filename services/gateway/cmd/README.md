<div align="center">

<!-- Top Banner -->
<p align="center">
  <img src="../../../assets/banner.svg" alt="NOVARIS Banner" width="100%" />
</p>

# 🛠️ NOVARIS Gateway CLI Suite (`/services/gateway/cmd`)
### Developer Utilities, Cryptographic Token Generation & Request Signing Tools

<p align="center">
  <img src="https://img.shields.io/badge/Phase_1_Capstone-COMPLETED_(100%25_Verified)-00C853?style=flat-square&logo=checkbox&logoColor=white" alt="Phase 1 Capstone" />
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
├── e2eprobe/         # End-to-end API contract verification runner (Day 11)
│   ├── main.go
│   ├── main_test.go
│   └── README.md
├── gentoken/         # Mints cryptographically signed JWT Bearer tokens (Day 2)
│   ├── main.go
│   └── README.md
├── signrequest/      # Computes canonical HMAC-SHA256 signatures over request bodies (Day 3)
│   ├── main.go
│   └── README.md
├── verifyopenapi/    # Validates OpenAPI 3.0 schema conformance & status codes (Day 12)
│   ├── main.go
│   ├── main_test.go
│   └── README.md
├── verifyschema/     # Verifies database schema, composite indexes & migration compliance (Day 5, 10)
│   ├── main.go
│   ├── main_test.go
│   └── README.md
├── loadtest/         # High-concurrency load generator & tail latency benchmark (Day 13)
│   ├── main.go
│   ├── main_test.go
│   └── README.md
├── phase1audit/      # Phase 1 Capstone automated audit & certification runner (Day 14)
│   ├── main.go
│   ├── main_test.go
│   └── README.md
├── verifykafka/      # Phase 2 Day 15 Kafka cluster & streaming core verification CLI
│   ├── main.go
│   └── README.md
├── verifyproducer/   # Phase 2 Day 16 Kafka producer & event visibility verification CLI
│   ├── main.go
│   ├── main_test.go
│   └── README.md
├── verifyconsumer/   # Phase 2 Day 17 Kafka consumer & in-order audit verification CLI
│   ├── main.go
│   ├── main_test.go
│   └── README.md
├── verifyoutbox/     # Phase 2 Day 18 Transactional Outbox & simulated crash recovery verifier
│   ├── main.go
│   ├── main_test.go
│   └── README.md
└── README.md         # This master CLI documentation
```

### 🧭 Subfolder Documentation Directory:
* 🔍 **[cmd/e2eprobe](./e2eprobe/README.md):** Automated black-box API contract probe runner.
* 🔑 **[cmd/gentoken](./gentoken/README.md):** RFC 7519 HMAC-SHA256 JWT minting utility.
* ✍️ **[cmd/signrequest](./signrequest/README.md):** RFC 2104 HMAC-SHA256 request signing and replay test tool.
* 🗄️ **[cmd/verifyschema](./verifyschema/README.md):** PostgreSQL schema compliance and static DDL validator.
* 📑 **[cmd/verifyopenapi](./verifyopenapi/README.md):** OpenAPI 3.0.3 contract gatekeeper and status code auditor.
* 🚀 **[cmd/loadtest](./loadtest/README.md):** High-concurrency worker pool load tester and OCC benchmark.
* 🎖️ **[cmd/phase1audit](./phase1audit/README.md):** Phase 1 Capstone automated audit and milestone certifier.
* ⚡ **[cmd/verifykafka](./verifykafka/README.md):** Phase 2 Kafka KRaft cluster diagnostic and partition routing tool.
* 📤 **[cmd/verifyproducer](./verifyproducer/README.md):** Phase 2 Kafka Producer diagnostic, event publisher and visibility audit tool.
* 📥 **[cmd/verifyconsumer](./verifyconsumer/README.md):** Phase 2 Kafka Consumer diagnostic, in-order FIFO audit and manual offset verifier.
* 📬 **[cmd/verifyoutbox](./verifyoutbox/README.md):** Phase 2 Transactional Outbox diagnostic, atomic dual-write and crash recovery verifier.
* 📥 **[cmd/verifyconsumer](./verifyconsumer/README.md):** Phase 2 Kafka Consumer diagnostic, in-order auditor and offset commit prover.

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

## 4. `e2eprobe` — End-to-End API Contract Verification Probe (Day 11)

Located at [`cmd/e2eprobe/main.go`](./e2eprobe/main.go), this interactive CLI test runner validates that a running gateway instance fulfills all Day 11 API contract specifications in sequence:
* Probes `/healthz` liveness and `/` service discovery.
* Verifies missing authentication defense (HTTP 401).
* Mints valid JWT, signs canonical body with HMAC-SHA256, and executes end-to-end payment settlement (HTTP 200).
* Fires immediate network retry with identical `Idempotency-Key` and verifies cached response with `Idempotent-Replayed: true` (HTTP 200).
* Attempts key reuse with mismatched payload and confirms security halt (HTTP 422).
* Asserts missing idempotency key rejection (HTTP 400).

### Supported CLI Flags:
| Flag | Type | Default | Description |
| :--- | :---: | :---: | :--- |
| `-url` | `string` | `http://localhost:8080` | Target gateway base URL. |
| `-merchant` | `string` | `c0000000-0000-0000-0000-000000000001` | Payer merchant account UUID. |
| `-recipient` | `string` | `c0000000-0000-0000-0000-000000000002` | Payee recipient account UUID. |
| `-amount` | `int64` | `5000` | Transfer magnitude in integer cents ($50.00 = 5000). |

### Usage Examples:
```bash
# Run probe sequence against local gateway on port 8080
go run cmd/e2eprobe/main.go

# Probe custom staging environment
go run cmd/e2eprobe/main.go -url https://staging-gateway.novaris.io

# Run automated Go unit test suite
go test -v ./cmd/e2eprobe
```

---

## 5. `verifyopenapi` — OpenAPI 3.0 Contract Verification Validator

Validates machine-readable OpenAPI schemas (`api/openapi.json`) statically against NOVARIS platform standards before deployment.

### Features:
1. **OAS Version Conformance:** Asserts specification version `3.0.x`.
2. **Metadata Integrity:** Verifies title, contact details, and server specifications.
3. **Mandatory Ingress Coverage:** Asserts presence of `/healthz`, `/`, `/openapi.json`, `/docs`, and `/v1/payments`.
4. **Complete Status Code Taxonomy:** Enforces documentation for all 7 standard HTTP response codes (`200`, `400`, `401`, `409`, `422`, `429`, `500`).
5. **Component Model Auditing:** Verifies presence of `PaymentRequest`, `PaymentReceipt`, `HealthResponse`, `RootDiscoveryResponse`, and `ErrorResponse`.

### Usage Examples:
```bash
# Validate default OpenAPI JSON specification
go run cmd/verifyopenapi/main.go

# Validate custom spec path
go run cmd/verifyopenapi/main.go -spec ../../api/openapi.json

# Run unit tests
go test -v ./cmd/verifyopenapi/...
```

---

## 6. `loadtest` — High-Concurrency Load Testing Engine & Hot-Wallet Stress Generator (Day 13)

Located at [`cmd/loadtest/main.go`](./loadtest/main.go), this high-throughput, native Go load testing utility executes concurrent stress tests against a live gateway instance:
* **Zero External Dependencies:** Built with pure Go standard library; runs immediately without installing k6 or Node.js.
* **Worker Pool Concurrency:** Spawns configurable worker goroutines (`-workers 50` or `-workers 100`) fed by thread-safe job channels.
* **Hot-Wallet Database Contention:** Stresses Optimistic Concurrency Control (OCC) version increments by directing all workers at the exact same account UUID.
* **Multi-Tenant Rate Limiting:** Simulates distinct merchant identities to evaluate per-merchant token-bucket isolation.
* **High-Precision Telemetry:** Measures nanosecond socket turnaround times to report min, average, p50 (median), p90, p95, p99, and max latency percentiles alongside full HTTP status code distributions (200, 400, 401, 409, 422, 429, 500).

### Supported CLI Flags:
| Flag | Type | Default | Description |
| :--- | :---: | :---: | :--- |
| `-url` | `string` | `http://localhost:8080` | Target gateway base URL. |
| `-workers` | `int` | `50` | Number of concurrent worker goroutines. |
| `-requests` | `int` | `250` | Total number of payment transactions to dispatch. |
| `-hotwallet` | `bool` | `true` | Direct all workers at the same hot-wallet account. |
| `-tenants` | `int` | `1` | Number of distinct merchant tenants (when `-hotwallet=false`). |
| `-amount` | `int64` | `100` | Transfer magnitude in integer cents ($1.00 = 100). |
| `-timeout` | `int` | `10` | Per-request client timeout in seconds. |

### Usage Examples:
```bash
# Quick verification: 50 concurrent workers, 250 total payment transactions
go run cmd/loadtest/main.go -workers 50 -requests 250

# Extreme hot-wallet stress test: 100 concurrent workers hitting 1 wallet
go run cmd/loadtest/main.go -workers 100 -requests 1000 -hotwallet

# Multi-tenant distributed test across 10 distinct merchants
go run cmd/loadtest/main.go -workers 50 -requests 500 -hotwallet=false -tenants 10

# Run automated unit test suite
go test -v ./cmd/loadtest/...
```

---

## 7. `cmd/phase1audit`: Phase 1 Capstone Milestone Certification Runner (Day 14)

### Purpose:
The `phase1audit` utility programmatically validates that NOVARIS satisfies all five core deliverables of **Phase 1: The Core Ledger Engine**:
1. **Criterion 1:** Go Gateway 5-Layer Security Perimeter (JWT Auth, HMAC Signatures, Token-Bucket Rate Limiter).
2. **Criterion 2:** Full PostgreSQL Double-Entry Ledger Schema (5 Core Tables, Integer Cents, Constraints).
3. **Criterion 3:** Atomic ACID Settlement Engine (Balanced Journal Entries, Rollback Safety Nets).
4. **Criterion 4:** Idempotency Deduplication & DB-Level OCC (SHA-256 Hashes, Monotonic Version CAS).
5. **Criterion 5:** 100 Concurrent Unit Tests with Zero Race Conditions.

### Supported CLI Flags:
| Flag | Type | Default | Description |
| :--- | :---: | :---: | :--- |
| `-json` | `bool` | `false` | Output audit report as machine-readable JSON for CI/CD pipelines. |

### Usage Examples:
```bash
# Run interactive ANSI terminal audit
go run cmd/phase1audit/main.go

# Output machine-readable JSON certificate for CI/CD
go run cmd/phase1audit/main.go -json

# Run unit test harness for the audit CLI
go test -v ./cmd/phase1audit/...
```

---

## 8. `verifykafka` — Kafka Cluster & Streaming Verification CLI (Phase 2, Day 15)

Located at [`cmd/verifykafka/main.go`](./verifykafka/main.go), this diagnostic tool audits the Phase 2 Kafka event backbone:
* **KRaft Quorum Health:** Validates TCP connectivity, queries controller node IDs, and audits registered active brokers.
* **Canonical Topic Provisioning:** Idempotently ensures all 6 core Novaris streaming topics (`novaris.tx.initiated`, `settled`, `failed`, `dlq`, `retry.1`, `retry.2`) exist with 3 partitions.
* **Deterministic Partition Routing Proof:** Mathematically verifies that transactions for a given account key consistently map to the identical partition ($P = |\text{Murmur2}(\text{Key})| \pmod 3$) preserving per-account FIFO invariants.
* **Diagnostic Probe Write Path:** Publishes an authentic probe event and validates synchronous broker Leader ACK.

### Supported CLI Flags:
| Flag | Type | Default | Description |
| :--- | :---: | :---: | :--- |
| `-brokers` | `string` | `localhost:9092` | Comma-separated list of target Kafka brokers. |
| `-client-id` | `string` | `novaris-verifykafka-cli` | Client identifier string sent during TCP wire handshakes. |
| `-json` | `bool` | `false` | Emits full audit findings as machine-readable JSON for CI/CD pipelines. |

### Usage Examples:
```bash
# Run interactive ANSI terminal audit against local KRaft cluster
go run cmd/verifykafka/main.go

# Export machine-readable JSON audit report
go run cmd/verifykafka/main.go -json
```

---

## 9. `verifyproducer` — Kafka Producer & Event Visibility Verification CLI (Phase 2, Day 16)

Located at [`cmd/verifyproducer/main.go`](./verifyproducer/main.go), this diagnostic tool certifies the Day 16 milestone by turning a payment request into a durable event on Kafka and confirming that events are **visible in the Kafka topic**:
* **Direct Event Production:** Serializes and publishes strongly-typed `TransactionInitiatedEvent` records to `novaris.tx.initiated` with Murmur2 deterministic partition keys.
* **Topic Visibility Audit:** Reads committed events directly back from the physical partition log using a lightweight reader to prove end-to-end event visibility.
* **Schema & Financial Invariant Validation:** Asserts 100% field parity across EventID, TransactionID, AmountCents, and Payer Account UUID.

### Supported CLI Flags:
| Flag | Type | Default | Description |
| :--- | :---: | :---: | :--- |
| `-brokers` | `string` | `localhost:9092` | Comma-separated list of target Kafka brokers. |
| `-topic` | `string` | `novaris.tx.initiated` | Target event topic for verification. |
| `-amount` | `int64` | `12500` | Payment magnitude in integer cents ($125.00 = 12500). |
| `-payer` | `string` | `c0000000-0000-0000-0000-000000000001` | Payer account UUID (Kafka Partition Key). |
| `-json` | `bool` | `false` | Emits full audit findings as machine-readable JSON for CI/CD pipelines. |

### Usage Examples:
```bash
# Run interactive ANSI terminal audit
go run cmd/verifyproducer/main.go

# Export machine-readable JSON audit report
go run cmd/verifyproducer/main.go -json
```

---

## 10. `verifyconsumer` — Kafka Consumer & In-Order Audit Engine (Day 17)

The `verifyconsumer` CLI is the dedicated diagnostic tool developed for **Phase 2 Day 17 (Kafka Consumer)**. It executes an end-to-end verification proving that consumer groups reliably read every published event in strict FIFO chronological sequence without message loss and commit offsets back to Kafka.

### Key Capabilities:
* **Broker & Group Handshake:** Probes TCP reachability and initializes dynamic consumer groups (`novaris-verifyconsumer-<uuid>`).
* **Sequential Batch Production:** Emits $N$ sequentially numbered events using a fixed Payer Account UUID, guaranteeing single-partition alignment.
* **In-Order FIFO Audit (Day 17 Outcome):** Reads the event batch, asserts that sequence numbers are strictly monotonic ($1 \rightarrow 2 \rightarrow 3 \rightarrow 4 \rightarrow 5$), and asserts zero message loss.
* **Offset Commit Proving:** Confirms offset checkpoints are durably registered in Kafka's `__consumer_offsets`.

### Supported CLI Flags:
| Flag | Type | Default | Description |
| :--- | :---: | :---: | :--- |
| `-brokers` | `string` | `localhost:9092` | Comma-separated list of target Kafka brokers. |
| `-topic` | `string` | `novaris.tx.initiated` | Target event topic for verification. |
| `-events` | `int` | `5` | Number of sequential events to publish and verify. |
| `-payer` | `string` | `c0000000-0000-0000-0000-000000000088` | Payer account UUID (Kafka Partition Key). |
| `-json` | `bool` | `false` | Emits full audit findings as machine-readable JSON for CI/CD pipelines. |

### Usage Examples:
```bash
# Run interactive ANSI terminal consumer audit
go run cmd/verifyconsumer/main.go

# Export machine-readable JSON audit report
go run cmd/verifyconsumer/main.go -json

# Verify 20-event high-volume stream
go run cmd/verifyconsumer/main.go -events 20
```

---

## 11. Transactional Outbox & Simulated Crash Recovery Verifier (`cmd/verifyoutbox`)

The `verifyoutbox` CLI is the dedicated diagnostic tool developed for **Phase 2 Day 18 (Transactional Outbox)**. It executes an end-to-end 5-stage automated audit proving that events are atomically enlisted within database transactions, safely polled by background relays using `FOR UPDATE SKIP LOCKED`, and that **zero events are dropped or duplicated** even when a worker process suffers a catastrophic mid-flight crash.

### Key Capabilities:
* **Infrastructure Health Check:** Validates live PostgreSQL and Kafka reachability, with automatic fallback to deterministic memory simulation (`-offline`).
* **Atomic Dual-Write Isolation:** Proves that an aborted transaction rolls back both ledger mutations and outbox records simultaneously.
* **Relay Sweep Verification:** Verifies background polling, Murmur2 partition key alignment, and `status = 'PUBLISHED'` transitions.
* **Simulated Crash Fault Injection (Day 18 Outcome):** Injects a mid-flight crash hook between Kafka publish and database status commit, asserting that a replacement worker recovers and completes the stream with **zero lost events**.
* **Downstream Deduplication Audit:** Confirms that re-transmitted messages are recognized and safely deduplicated by downstream idempotent consumers.

### Supported CLI Flags:
| Flag | Type | Default | Description |
| :--- | :---: | :---: | :--- |
| `-brokers` | `string` | `localhost:9092` | Comma-separated list of target Kafka brokers. |
| `-topic` | `string` | `novaris.tx.initiated` | Target event topic for verification. |
| `-count` | `int` | `5` | Number of transactions to test. |
| `-offline` | `bool` | `true` | Runs hermetic in-memory simulation without external infrastructure. |
| `-json` | `bool` | `false` | Emits full audit findings as machine-readable JSON for CI/CD pipelines. |

### Usage Examples:
```bash
# Run 5-stage automated crash recovery diagnostic
go run cmd/verifyoutbox/main.go -offline -count=5

# Emit machine-readable JSON report
go run cmd/verifyoutbox/main.go -offline -json
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
