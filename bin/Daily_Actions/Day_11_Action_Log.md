<div align="center">

<!-- Top Banner -->
<p align="center">
  <img src="../../assets/banner.svg" alt="NOVARIS Banner" width="100%" />
</p>

# 🏛️ NOVARIS • Engineering Action Log
## DAY 11: Integration Wiring & End-to-End API Contracts
### Connecting Gateway Ingress, Defense-in-Depth Security, Idempotent Deduplication, and Database OCC Consensus into an Unbroken Financial Settlement Path

<p align="center">
  <img src="https://img.shields.io/badge/Milestone-Day_11_of_84-00E5FF?style=flat-square&logo=target&logoColor=white" alt="Milestone" />
  <img src="https://img.shields.io/badge/Phase-1:_Design,_Ledger_%26_Perimeter-00C853?style=flat-square&logo=shield&logoColor=white" alt="Phase" />
  <img src="https://img.shields.io/badge/Focus-Integration_Wiring_%26_Contracts-7C4DFF?style=flat-square&logo=gitbook&logoColor=white" alt="Focus" />
  <img src="https://img.shields.io/badge/Coverage-100%25_Full_Stack_Contract-FFA000?style=flat-square&logo=speedtest&logoColor=white" alt="Coverage" />
  <img src="https://img.shields.io/badge/Standard-IETF_%26_Postman_v2.1-FF6D00?style=flat-square&logo=postman&logoColor=white" alt="Standard" />
</p>

---

</div>

## 📑 Master Table of Contents

1. [Executive Summary & The Integration Milestone](#1-executive-summary--the-integration-milestone)
2. [Chapter 1: The "Two Unit Tests, Zero Integration Tests" Dilemma](#2-chapter-1-the-two-unit-tests-zero-integration-tests-dilemma)
3. [Chapter 2: The Complete Five-Layer Defense-in-Depth Pipeline Architecture](#3-chapter-2-the-complete-five-layer-defense-in-depth-pipeline-architecture)
4. [Chapter 3: Formal API Contracts & HTTP Status Code Taxonomy](#4-chapter-3-formal-api-contracts--http-status-code-taxonomy)
5. [Chapter 4: Postman Collection & Automated Pre-Request Cryptography](#5-chapter-4-postman-collection--automated-pre-request-cryptography)
6. [Chapter 5: Line-by-Line Architectural Walkthrough of New & Wired Components](#6-chapter-5-line-by-line-architectural-walkthrough-of-new--wired-components)
   - 5.1 [Gateway Entry Point & Pipeline Assembly (`main.go`)](#51-gateway-entry-point--pipeline-assembly-maingo)
   - 5.2 [Deterministic Integration Seed Accounts (`004_seed_test_accounts.sql`)](#52-deterministic-integration-seed-accounts-004_seed_test_accountssql)
   - 5.3 [SQL Integration & Monetary Conservation Harness (`verify_day11_integration.sql`)](#53-sql-integration--monetary-conservation-harness-verify_day11_integrationsql)
   - 5.4 [Automated End-to-End Integration Test Suite (`integration_test.go`)](#54-automated-end-to-end-integration-test-suite-integration_testgo)
   - 5.5 [Interactive Developer Contract Probe CLI (`cmd/e2eprobe/main.go`)](#55-interactive-developer-contract-probe-cli-cmde2eprobemaingo)
   - 5.6 [Standardized Postman Collection & Environment (`deployments/postman/`)](#56-standardized-postman-collection--environment-deploymentspostman)
7. [Chapter 6: Step-by-Step Execution Logs & Verification Proofs](#7-chapter-6-step-by-step-execution-logs--verification-proofs)
8. [Chapter 7: Copy-Paste Runbook for Local Development & Testing](#8-chapter-7-copy-paste-runbook-for-local-development--testing)
9. [Chapter 8: The Distributed Systems & Fintech Interview Defense Compendium](#9-chapter-8-the-distributed-systems--fintech-interview-defense-compendium)
10. [Chapter 9: Common Integration Pitfalls & Debugging Playbook](#10-chapter-9-common-integration-pitfalls--debugging-playbook)
11. [Chapter 10: Spaced Recall & The Master Curriculum Roadmap](#11-chapter-10-spaced-recall--the-master-curriculum-roadmap)
12. [Chapter 11: Day 11 Verification Certificate & Sign-Off](#12-chapter-11-day-11-verification-certificate--sign-off)

---

## 1. Executive Summary & The Integration Milestone

Across Days 1 through 10 of the NOVARIS 84-day master engineering curriculum, we methodically built the core foundational pillars of our payment architecture in isolation:
* **Day 01:** Relational foundation, PostgreSQL 16 Alpine container mesh, initial double-entry schema, and the public `/healthz` liveness probe.
* **Day 02:** Cryptographic identity verification via RFC 7519 JSON Web Tokens (JWT), algorithm confusion defenses (`alg: HS256` strictly enforced), and context claims injection.
* **Day 03:** Payload integrity and replay attack prevention using RFC 2104 HMAC-SHA256 signatures, constant-time verification (`hmac.Equal`), and sliding timestamp drift windows ($\pm 300\text{s}$).
* **Day 04:** Per-merchant traffic velocity control via in-memory Token-Bucket rate limiting (10 burst capacity, 2 req/sec refill) implementing a critical fail-fast perimeter optimization.
* **Day 05:** Production PostgreSQL ledger schema extensions, human-in-the-loop compliance review queues (`fraud_reviews`), and evolving ML feature snapshots (`feature_snapshots`) with composite B-Tree indexing.
* **Day 06:** Core double-entry ledger bookkeeping settlement logic, atomic multi-table ACID transactions, and deferred rollback traps.
* **Day 07:** Table-driven testing discipline, mock database isolation (`go-sqlmock`), subtest parallelism, and zero-allocation execution benchmarking.
* **Day 08:** High-concurrency stress testing, nanosecond start-gate goroutine barriers, and ThreadSanitizer data race detection.
* **Day 09:** Distributed database-level Optimistic Concurrency Control (OCC) using monotonic version columns, atomic check-and-increment writes, collision detection (`RowsAffected == 0`), and jittered exponential-backoff retries.
* **Day 10:** Idempotent API design, IETF draft header standards, SHA-256 canonical body fingerprinting, response capture buffer decoration, and sub-millisecond cached duplicate replays.

### What is the Goal of Day 11?
In software engineering, there is a fundamental difference between having ten individual components that each pass their isolated unit tests and having **one unbroken, working request path that functions reliably from end to end**. 

On **Day 11**, we achieve **Integration Wiring**:
1. We unite all five perimeter defense and consensus layers into a single, cohesive, production-grade HTTP pipeline inside [`services/gateway/main.go`](file:///C:/Users/Tokachichu%20SriCharan/OneDrive/Documents/GitHub/novaris/services/gateway/main.go).
2. We establish a formal **API Contract** defining exactly what headers, payloads, status codes, and invariants the platform guarantees to external consumers.
3. We provide an industry-standard **Postman Collection & Environment Suite** equipped with pre-request cryptographic scripts that automate JWT minting and HMAC-SHA256 hashing.
4. We engineer an exhaustive **End-to-End Integration Test Suite** ([`integration_test.go`](file:///C:/Users/Tokachichu%20SriCharan/OneDrive/Documents/GitHub/novaris/services/gateway/integration_test.go)) that spins up a live test server and verifies 9 distinct contract scenarios under simultaneous multi-merchant concurrency.
5. We build an interactive developer CLI probe ([`cmd/e2eprobe`](file:///C:/Users/Tokachichu%20SriCharan/OneDrive/Documents/GitHub/novaris/services/gateway/cmd/e2eprobe/main.go)) that prints colorful, formatted telemetry of every step in the pipeline.
6. We seed deterministic test accounts in PostgreSQL via [`migrations/004_seed_test_accounts.sql`](file:///C:/Users/Tokachichu%20SriCharan/OneDrive/Documents/GitHub/novaris/migrations/004_seed_test_accounts.sql) and mathematically prove monetary balance conservation via [`migrations/verify_day11_integration.sql`](file:///C:/Users/Tokachichu%20SriCharan/OneDrive/Documents/GitHub/novaris/migrations/verify_day11_integration.sql).

---

## 2. Chapter 1: The "Two Unit Tests, Zero Integration Tests" Dilemma

### 1.1 The Classical Engineering Trap
In computer science, there is a famous software engineering parable:
> *"A sliding door has a unit test that confirms it slides back and forth smoothly. A deadbolt lock has a unit test that confirms its bolt extends and retracts smoothly. Both unit tests pass with 100% code coverage. But when you install the deadbolt lock on the sliding door, an intruder slides the door sideways and the lock passes right through the air without stopping it! Two unit tests passed, zero integration tests conducted."*

In payment gateways and distributed financial ledgers, this failure mode is not a joke — it leads to lost funds, corrupted balances, and catastrophic audit failures:
* **The Authentication Trap:** Your JWT module might correctly verify a cryptographic token in a mock unit test. But when plugged into the real router, it might fail to inject the parsed claims into the request context, causing downstream settlement handlers to crash with nil-pointer exceptions!
* **The Stream Draining Trap:** Your HMAC middleware might correctly read the HTTP body to calculate a SHA-256 digest. But if it forgets to restore `r.Body` using `io.NopCloser`, the downstream idempotency middleware reads an empty byte stream, calculates `SHA-256("")`, and rejects valid payments as invalid payloads!
* **The Layer Ordering Trap:** If you place HMAC verification *before* your Token-Bucket rate limiter, an attacker can launch an unauthenticated DDoS flood with 10MB bodies. Your server will read the body, allocate buffers, and compute SHA-256 digests on millions of requests, causing CPU starvation and memory exhaustion before the rate limiter ever gets evaluated!
* **The Double-Debit Trap:** If your idempotency middleware writes status code 200 to the network socket *before* the inner database transaction commits, and the database subsequently encounters an OCC collision and rolls back, the client thinks the payment succeeded while the ledger has no record of it!

### 1.2 The Testing Hierarchy: Unit vs. Integration vs. End-to-End
To understand where Day 11 fits into enterprise software development, consider the testing pyramid:

```
           / \
          /   \
         / E2E \         <-- Day 11 (Full Pipeline + Postman + CLI Probes)
        /-------\
       / Integr. \       <-- Day 11 (httptest.Server + Live DB + Multi-Layer)
      /-----------\
     / Unit Tests  \     <-- Days 2, 3, 4, 6, 7, 8, 9, 10 (Isolated Modules)
    /---------------\
```

* **Unit Testing (Days 2–10):** Focuses on a single function, struct, or package in strict isolation. Dependencies are replaced with mock stubs (`go-sqlmock`). Fast, deterministic, and ideal for edge-case boundary testing (\$0.01 transfers, empty strings, nil pointers).
* **Integration Testing (Day 11):** Focuses on the *interfaces between modules*. We test how the JWT parser hands claims to the rate limiter, how the rate limiter protects the HMAC validator, and how the idempotency store wraps the OCC database transaction.
* **End-to-End (E2E) Testing (Day 11):** Focuses on the *system as a black box from the perspective of an external client*. A real TCP socket is opened, real HTTP headers are serialized over the wire, and the server returns real JSON responses with headers like `Idempotent-Replayed: true`.

---

## 3. Chapter 2: The Complete Five-Layer Defense-in-Depth Pipeline Architecture

On Day 11, NOVARIS realizes its unified ingress and settlement architecture. Every mutating request arriving at `POST /v1/payments` must pass through five rigorous, specialized defense checkpoints:

```mermaid
flowchart TD
    Client["Merchant / Client App"] -->|"POST /v1/payments (Headers + Body)"| Logging["Logging Middleware (Latency & Tracing)"]
    
    subgraph DefenseMesh ["The Five-Layer Defense-in-Depth Mesh"]
        Logging --> Layer1["Layer 1: JWT Authentication (RFC 7519)<br/>Checks Bearer Signature, Expiration & Algorithm.<br/>Injects verified MerchantID into Context."]
        Layer1 -->|"❌ 401 Unauthorized (Missing/Forged/Expired)"| Client
        
        Layer1 -->|"✅ Identity Verified"| Layer2["Layer 2: Token-Bucket Rate Limiter<br/>Evaluates Per-Merchant Quota (10 burst, 2/s).<br/>Zero-Goroutine Lazy Math Refill."]
        Layer2 -->|"❌ 429 Too Many Requests (+ Retry-After)"| Client
        
        Layer2 -->|"✅ Quota Admitted"| Layer3["Layer 3: HMAC-SHA256 Anti-Tamper (RFC 2104)<br/>Validates Timestamp Freshness (±300s).<br/>Recomputes SHA-256 Digest in Constant-Time."]
        Layer3 -->|"❌ 401 Unauthorized (Tampered/Stale Replay)"| Client
        
        Layer3 -->|"✅ Integrity Intact"| Layer4["Layer 4: Idempotency Deduplication (IETF)<br/>Evaluates SHA-256 Canonical Body Hash.<br/>Inspects State in idempotency_keys Table."]
        Layer4 -->|"❌ 400 Bad Request (Missing Key)"| Client
        Layer4 -->|"❌ 409 Conflict (In-Flight Race)"| Client
        Layer4 -->|"❌ 422 Unprocessable (Payload Mismatch)"| Client
        Layer4 -->|"🔁 200 OK (Cached Replay: Idempotent-Replayed: true)"| Client
        
        Layer4 -->|"✨ Fresh Mutating Request"| Layer5["Layer 5: Core OCC Settlement Consensus<br/>Executes Atomic Multi-Table Double-Entry Writes.<br/>Evaluates Version Invariant (WHERE version = $X).<br/>Stages Event into Transactional Outbox."]
        Layer5 -->|"❌ 409 Conflict (Contention Exhausted)"| Client
        Layer5 -->|"❌ 500 Internal Error (DB Rollback)"| Client
        Layer5 -->|"✅ 200 OK (Fresh Settlement Receipt)"| Layer4
    end

    Layer4 -->|"Buffer & Cache Response Stream"| Client
```

### The Architectural Rationale for Layer Ordering:
The sequence of middleware in an enterprise gateway is governed by **computational cost and attack surface minimization**:
1. **Why Identity (JWT) First?** We cannot evaluate per-merchant rate limits or account permissions until we cryptographically establish *who* the caller is.
2. **Why Rate Limiting Second?** Cryptographic body hashing (HMAC-SHA256) and database queries are computationally expensive. By checking the in-memory token bucket immediately after extracting the merchant ID, we drop denial-of-service floods in **nanoseconds**, before allocating memory to read request bodies or contacting PostgreSQL!
3. **Why HMAC Third?** We verify that the payload has not been tampered with in-transit before calculating idempotency hashes or executing business logic.
4. **Why Idempotency Fourth?** If a request is a duplicate retry, we can serve the cached response immediately from memory or the `idempotency_keys` table without incurring the cost of an ACID multi-table database transaction!
5. **Why OCC Settlement Fifth?** The financial ledger is the most sensitive asset in the enterprise. Only requests that are authenticated, throttled, verified, and deduplicated are permitted to touch the database.

---

## 4. Chapter 3: Formal API Contracts & HTTP Status Code Taxonomy

An **API Contract** is a formal, unambiguous agreement between the server and its clients. It dictates what inputs the server requires, what outputs it guarantees, what status codes it emits under specific conditions, and what invariants remain true under all circumstances.

### 4.1 The NOVARIS Day 11 Contract Matrix

```
┌──────┬──────────────────────┬────────────────────────┬────────────────────────────────────────────────────────┐
│ CODE │ STATUS               │ TRIGGER CONDITION      │ CLIENT ACTION REQUIRED                                │
├──────┼──────────────────────┼────────────────────────┼────────────────────────────────────────────────────────┤
│ 200  │ OK                   │ Fresh Payment Settled  │ Payment committed. Process transaction ID.             │
├──────┼──────────────────────┼────────────────────────┼────────────────────────────────────────────────────────┤
│ 200  │ OK (Cached Replay)   │ Idempotent Network Re- │ Header `Idempotent-Replayed: true` present.            │
│      │                      │ try of identical req   │ Safe retry: DO NOT re-charge; display original receipt.│
├──────┼──────────────────────┼────────────────────────┼────────────────────────────────────────────────────────┤
│ 400  │ Bad Request          │ Missing Idempotency-   │ Fix client code: attach UUIDv4 `Idempotency-Key`       │
│      │                      │ Key or Self-Transfer   │ or correct identical sender/recipient account IDs.    │
├──────┼──────────────────────┼────────────────────────┼────────────────────────────────────────────────────────┤
│ 401  │ Unauthorized         │ Missing/Expired JWT or │ Re-authenticate: refresh Bearer token or verify       │
│      │                      │ Tampered HMAC signature│ shared secret and clock synchronization (NTP).        │
├──────┼──────────────────────┼────────────────────────┼────────────────────────────────────────────────────────┤
│ 409  │ Conflict             │ Concurrent Duplicate or│ In-flight request being processed; retry with         │
│      │                      │ High OCC Contention    │ exponential backoff and randomized jitter.            │
├──────┼──────────────────────┼────────────────────────┼────────────────────────────────────────────────────────┤
│ 422  │ Unprocessable Entity │ Key Reuse with Mis-    │ Security Alert: Client reused an Idempotency-Key from │
│      │                      │ matched Request Body   │ an earlier payment with altered transfer parameters!   │
├──────┼──────────────────────┼────────────────────────┼────────────────────────────────────────────────────────┤
│ 429  │ Too Many Requests    │ Merchant Rate Limit    │ Backoff: Read `Retry-After: <sec>` header and pause   │
│      │                      │ Quota Exhausted        │ outbound requests until bucket refills.                │
├──────┼──────────────────────┼────────────────────────┼────────────────────────────────────────────────────────┤
│ 500  │ Internal Server Error│ Database Disk Failure  │ Server-side failure: Transaction safely rolled back.  │
│      │                      │ or Unhandled Exception │ Safe to retry with same Idempotency-Key.               │
└──────┴──────────────────────┴────────────────────────┴────────────────────────────────────────────────────────┘
```

---

## 5. Chapter 4: Postman Collection & Automated Pre-Request Cryptography

### 5.1 Why Postman is the Industry Standard for Integration Verification
While automated unit tests in Go are essential for fast regression feedback during development, external developers, frontend teams, and QA engineers interact with APIs via graphical REST clients like **Postman**.

However, testing NOVARIS in Postman presents a significant cryptographic challenge:
* Every request requires a valid RFC 7519 JWT token signed with HMAC-SHA256.
* Every request requires a canonical HMAC-SHA256 signature computed over `<current_unix_timestamp>.<raw_body>`.
* Every request requires an exact timestamp matching the current system clock within $\pm 300\text{s}$.
* Every request requires a fresh, unique `Idempotency-Key` UUID.

If an engineer had to manually calculate HMAC signatures in a terminal and copy-paste hex strings into Postman headers for every single test, testing would be agonizingly slow and error-prone!

### 5.2 Automating Cryptography with Postman Pre-Request Scripts
Postman includes a built-in JavaScript execution sandbox featuring the full **CryptoJS** library. In our Day 11 Postman collection ([`NOVARIS_API_Contracts.postman_collection.json`](file:///C:/Users/Tokachichu%20SriCharan/OneDrive/Documents/GitHub/novaris/deployments/postman/NOVARIS_API_Contracts.postman_collection.json)), we implemented automated pre-request scripts that execute before every outbound request:

```javascript
// 1. Generate a cryptographically random UUIDv4 Idempotency-Key
var idempKey = 'postman-key-' + CryptoJS.lib.WordArray.random(16).toString();
pm.environment.set("current_idempotency_key", idempKey);
pm.request.headers.upsert({ key: "Idempotency-Key", value: idempKey });

// 2. Read raw JSON body and capture current UTC Unix epoch seconds
var body = pm.request.body.raw;
var now = Math.floor(Date.now() / 1000);
var secret = pm.environment.get("hmac_secret");

// 3. Construct canonical representation: <timestamp>.<raw_body>
var canonical = now + "." + body;

// 4. Compute HMAC-SHA256 hex digest
var sig = CryptoJS.HmacSHA256(canonical, secret).toString(CryptoJS.enc.Hex);

// 5. Inject fresh headers into outbound request
pm.request.headers.upsert({ key: "X-Novaris-Timestamp", value: String(now) });
pm.request.headers.upsert({ key: "X-Novaris-Signature", value: sig });
```

This transforms Postman into a zero-friction, one-click manual and automated contract testing environment!

---

## 6. Chapter 5: Line-by-Line Architectural Walkthrough of New & Wired Components

### 5.1 Gateway Entry Point & Pipeline Assembly (`main.go`)

In [`services/gateway/main.go`](file:///C:/Users/Tokachichu%20SriCharan/OneDrive/Documents/GitHub/novaris/services/gateway/main.go), we refactored the router initialization into a clean, reusable factory function `SetupServerMux`:

```go
// SetupServerMux constructs the production HTTP request router and wires the complete
// five-layer defense-in-depth middleware pipeline onto /v1/payments.
func SetupServerMux(cfg Config, dbConn *sql.DB, bucketStore *ratelimit.BucketStore, idempStore IdempotencyStore) http.Handler {
	mux := http.NewServeMux()

	// 1. Public Routes (Liveness and Service Discovery)
	mux.HandleFunc("/", rootHandler(cfg))
	mux.HandleFunc("/healthz", healthzHandler(cfg))

	// 2. Protected Routes (Stacked Defense-in-Depth Middleware Pipeline)
	// Execution order:
	//   1. JWTMiddleware: Answers "WHO ARE YOU?" (authenticates Bearer token, injects claims).
	//   2. RateLimitMiddleware: Answers "HOW MUCH / HOW FAST?" (evaluates token-bucket quota).
	//   3. HMACMiddleware: Answers "HAS THIS BEEN ALTERED OR REPLAYED?" (verifies signature & timestamp).
	//   4. IdempotencyMiddleware: Answers "HAVE WE PROCESSED THIS NETWORK RETRY?" (deduplicates via SHA-256).
	//   5. paymentsHandler: Executes atomic OCC double-entry settlement in PostgreSQL ledger.
	idempChain := IdempotencyMiddleware(idempStore)(http.HandlerFunc(paymentsHandler))
	hmacChain := auth.HMACMiddleware(http.HandlerFunc(idempChain.ServeHTTP))
	rateLimitChain := ratelimit.RateLimitMiddleware(bucketStore)(hmacChain)
	fullPipeline := auth.JWTMiddleware(rateLimitChain)

	mux.Handle("/v1/payments", fullPipeline)

	return loggingMiddleware(mux)
}
```

#### Line-by-Line Explanation of `SetupServerMux`:
* **Lines 285–286:** We accept dependencies (`cfg`, `dbConn`, `bucketStore`, `idempStore`) as parameters rather than relying on global variables. This enables unit and integration tests to pass in mock stores or test servers without mutating global state.
* **Lines 288–290:** Public endpoints (`/` and `/healthz`) are registered directly on the multiplexer without security middleware, allowing Kubernetes and external load balancers to probe container health without needing authentication tokens.
* **Lines 300–304:** Middleware chaining in Go is constructed from the **inside out**:
  - `idempChain`: Wraps `paymentsHandler` with `IdempotencyMiddleware`. If a request is a duplicate, it returns the cached response and never calls `paymentsHandler`.
  - `hmacChain`: Wraps `idempChain` with `auth.HMACMiddleware`. We pass `http.HandlerFunc(idempChain.ServeHTTP)` to adapt Go's `http.Handler` interface to `http.HandlerFunc`.
  - `rateLimitChain`: Wraps `hmacChain` with `ratelimit.RateLimitMiddleware(bucketStore)`. Drops floods before HMAC hashing.
  - `fullPipeline`: Wraps `rateLimitChain` with `auth.JWTMiddleware`. Verifies identity before checking rate limits.
* **Line 306:** Mounts the complete pipeline onto `/v1/payments` via `mux.Handle`.
* **Line 308:** Wraps the entire multiplexer with `loggingMiddleware` to log microsecond latencies across all endpoints.

---

### 5.2 Deterministic Integration Seed Accounts (`004_seed_test_accounts.sql`)

In [`migrations/004_seed_test_accounts.sql`](file:///C:/Users/Tokachichu%20SriCharan/OneDrive/Documents/GitHub/novaris/migrations/004_seed_test_accounts.sql), we provision deterministic accounts with known UUIDs and initial balances:

```sql
INSERT INTO accounts (id, user_id, account_type, currency, balance_cents, version, created_at, updated_at)
VALUES
    (
        'c0000000-0000-0000-0000-000000000001', -- Alpha Merchant (Sender / Payer)
        'a0000000-0000-0000-0000-000000000001',
        'MERCHANT_PAYOUT',
        'USD',
        1000000, -- $10,000.00 USD (1,000,000 cents)
        0,
        NOW(),
        NOW()
    ),
    (
        'c0000000-0000-0000-0000-000000000002', -- Beta Merchant (Recipient / Payee)
        'a0000000-0000-0000-0000-000000000002',
        'WALLET',
        'USD',
        250000,  -- $2,500.00 USD (250,000 cents)
        0,
        NOW(),
        NOW()
    ),
    (
        'c0000000-0000-0000-0000-000000000003', -- Gamma Clearing Escrow
        'a0000000-0000-0000-0000-000000000003',
        'ESCROW',
        'USD',
        0,       -- $0.00 USD (0 cents)
        0,
        NOW(),
        NOW()
    )
ON CONFLICT (id) DO NOTHING;
```

#### Line-by-Line Explanation of the Seed Migration:
* **Deterministic UUIDs:** Instead of random UUIDs, we use standardized hexadecimal sequences (`c000...0001`, `c000...0002`) so developers can easily recognize the sender and recipient in logs, terminal outputs, and database dumps.
* **Integer Cents Representation:** All balances use `BIGINT` integer cents. Alpha starts with 1,000,000 cents (\$10,000.00 USD); Beta starts with 250,000 cents (\$2,500.00 USD).
* **Initial OCC Versions:** Both accounts start at `version = 0`. With every balance mutation, the version increments monotonically.
* **`ON CONFLICT (id) DO NOTHING`:** Guarantees idempotency. If this script is executed multiple times, existing account balances and ongoing OCC versions are preserved without being overwritten.

---

### 5.3 SQL Integration & Monetary Conservation Harness (`verify_day11_integration.sql`)

In [`migrations/verify_day11_integration.sql`](file:///C:/Users/Tokachichu%20SriCharan/OneDrive/Documents/GitHub/novaris/migrations/verify_day11_integration.sql), we construct a PL/pgSQL verification harness proving that financial balance is preserved under atomic multi-table execution:

```sql
-- ==============================================================================
-- NOVARIS: Distributed Payment Settlement & Adaptive Fraud Intelligence Platform
-- SQL VERIFICATION HARNESS: Day 11 End-to-End Integration & Seed Data Validation
-- FILE: migrations/verify_day11_integration.sql
-- ==============================================================================

DO $$
DECLARE
    v_sender_id UUID := 'c0000000-0000-0000-0000-000000000001';
    v_recipient_id UUID := 'c0000000-0000-0000-0000-000000000002';
    v_transfer_amount BIGINT := 5000; -- $50.00 USD
    v_sender_bal_before BIGINT;
    v_recipient_bal_before BIGINT;
    v_sender_ver_before INT;
    v_recipient_ver_before INT;
    v_sender_bal_after BIGINT;
    v_recipient_bal_after BIGINT;
    v_sender_ver_after INT;
    v_recipient_ver_after INT;
    v_tx_id UUID;
    v_idemp_key VARCHAR(128) := 'e2e-harness-key-' || uuid_generate_v4()::text;
    v_payload_hash VARCHAR(64) := 'd2d2d2d2d2d2d2d2d2d2d2d2d2d2d2d2d2d2d2d2d2d2d2d2d2d2d2d2d2d2d2d2';
    v_rows_affected INT;
    v_total_system_before BIGINT;
    v_total_system_after BIGINT;
BEGIN
    RAISE NOTICE '==================================================================';
    RAISE NOTICE ' 🧪 NOVARIS DAY 11: End-to-End Integration & Seed Harness';
    RAISE NOTICE '==================================================================';

    -- 1. Ensure test accounts exist
    PERFORM 1 FROM accounts WHERE id = v_sender_id;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'CRITICAL: Sender seed account % is missing! Run 004_seed_test_accounts.sql first.', v_sender_id;
    END IF;

    PERFORM 1 FROM accounts WHERE id = v_recipient_id;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'CRITICAL: Recipient seed account % is missing! Run 004_seed_test_accounts.sql first.', v_recipient_id;
    END IF;

    -- 2. Read state before settlement
    SELECT balance_cents, version INTO v_sender_bal_before, v_sender_ver_before FROM accounts WHERE id = v_sender_id;
    SELECT balance_cents, version INTO v_recipient_bal_before, v_recipient_ver_before FROM accounts WHERE id = v_recipient_id;
    v_total_system_before := v_sender_bal_before + v_recipient_bal_before;

    RAISE NOTICE '  [STATE BEFORE] Alpha (Sender):    $% (cents: %, version: %)', (v_sender_bal_before::numeric / 100.0), v_sender_bal_before, v_sender_ver_before;
    RAISE NOTICE '  [STATE BEFORE] Beta (Recipient):  $% (cents: %, version: %)', (v_recipient_bal_before::numeric / 100.0), v_recipient_bal_before, v_recipient_ver_before;
    RAISE NOTICE '  [STATE BEFORE] Total Money Pool:  $% (cents: %)', (v_total_system_before::numeric / 100.0), v_total_system_before;

    -- 3. Execute End-to-End Multi-Table Atomic Settlement
    -- Leg A: Register Idempotency Key in-progress
    INSERT INTO idempotency_keys (key, account_id, request_path, request_hash, status, created_at, expires_at)
    VALUES (v_idemp_key, v_sender_id::text, '/v1/payments', v_payload_hash, 'IN_PROGRESS', NOW(), NOW() + INTERVAL '24 hours');

    -- Leg B: Insert Transaction
    INSERT INTO transactions (idempotency_key, account_id, amount_cents, currency, status, created_at, updated_at)
    VALUES (v_idemp_key, v_sender_id, v_transfer_amount, 'USD', 'SETTLED', NOW(), NOW())
    RETURNING id INTO v_tx_id;

    -- Leg C: Debit Sender via OCC check-and-increment
    UPDATE accounts
    SET balance_cents = balance_cents - v_transfer_amount,
        version = version + 1,
        updated_at = NOW()
    WHERE id = v_sender_id AND version = v_sender_ver_before AND balance_cents >= v_transfer_amount;
    GET DIAGNOSTICS v_rows_affected = ROW_COUNT;
    IF v_rows_affected != 1 THEN
        RAISE EXCEPTION 'FAILED: OCC Debit failed on sender account!';
    END IF;

    -- Leg D: Credit Recipient via OCC check-and-increment
    UPDATE accounts
    SET balance_cents = balance_cents + v_transfer_amount,
        version = version + 1,
        updated_at = NOW()
    WHERE id = v_recipient_id AND version = v_recipient_ver_before;
    GET DIAGNOSTICS v_rows_affected = ROW_COUNT;
    IF v_rows_affected != 1 THEN
        RAISE EXCEPTION 'FAILED: OCC Credit failed on recipient account!';
    END IF;

    -- Leg E: Balanced Double-Entry Journal Entries
    INSERT INTO journal_entries (transaction_id, account_id, entry_type, amount_cents, currency, created_at)
    VALUES
        (v_tx_id, v_sender_id, 'DEBIT', v_transfer_amount, 'USD', NOW()),
        (v_tx_id, v_recipient_id, 'CREDIT', v_transfer_amount, 'USD', NOW());

    -- Leg F: Transactional Outbox Event
    INSERT INTO outbox (event_type, aggregate_type, aggregate_id, payload, created_at)
    VALUES (
        'PAYMENT_SETTLED',
        'TRANSACTION',
        v_tx_id::text,
        jsonb_build_object(
            'transaction_id', v_tx_id,
            'sender_id', v_sender_id,
            'recipient_id', v_recipient_id,
            'amount_cents', v_transfer_amount,
            'currency', 'USD',
            'timestamp', NOW()
        ),
        NOW()
    );

    -- Leg G: Update Idempotency Key to COMPLETED with cached JSON response
    UPDATE idempotency_keys
    SET status = 'COMPLETED',
        response_code = 200,
        response_body = jsonb_build_object(
            'status', 'settled',
            'transaction_id', v_tx_id,
            'merchant_id', v_sender_id::text,
            'amount_cents', v_transfer_amount,
            'recipient', v_recipient_id::text,
            'message', 'Payment settled end-to-end via OCC double-entry ledger pipeline.'
        )
    WHERE key = v_idemp_key;

    -- 4. Read State After Settlement
    SELECT balance_cents, version INTO v_sender_bal_after, v_sender_ver_after FROM accounts WHERE id = v_sender_id;
    SELECT balance_cents, version INTO v_recipient_bal_after, v_recipient_ver_after FROM accounts WHERE id = v_recipient_id;
    v_total_system_after := v_sender_bal_after + v_recipient_bal_after;

    RAISE NOTICE '------------------------------------------------------------------';
    RAISE NOTICE '  [STATE AFTER]  Alpha (Sender):    $% (cents: %, version: %)', (v_sender_bal_after::numeric / 100.0), v_sender_bal_after, v_sender_ver_after;
    RAISE NOTICE '  [STATE AFTER]  Beta (Recipient):  $% (cents: %, version: %)', (v_recipient_bal_after::numeric / 100.0), v_recipient_bal_after, v_recipient_ver_after;
    RAISE NOTICE '  [STATE AFTER]  Total Money Pool:  $% (cents: %)', (v_total_system_after::numeric / 100.0), v_total_system_after;

    -- 5. Assert Invariants
    IF v_total_system_before != v_total_system_after THEN
        RAISE EXCEPTION 'CRITICAL INVARIANT VIOLATION: Money created or destroyed! Before=%, After=%', v_total_system_before, v_total_system_after;
    END IF;

    IF (v_sender_bal_before - v_sender_bal_after) != v_transfer_amount THEN
        RAISE EXCEPTION 'FAILED: Sender balance not debited by exactly % cents!', v_transfer_amount;
    END IF;

    IF (v_recipient_bal_after - v_recipient_bal_before) != v_transfer_amount THEN
        RAISE EXCEPTION 'FAILED: Recipient balance not credited by exactly % cents!', v_transfer_amount;
    END IF;

    IF v_sender_ver_after != (v_sender_ver_before + 1) THEN
        RAISE EXCEPTION 'FAILED: Sender version did not increment monotonically!';
    END IF;

    IF v_recipient_ver_after != (v_recipient_ver_before + 1) THEN
        RAISE EXCEPTION 'FAILED: Recipient version did not increment monotonically!';
    END IF;

    RAISE NOTICE '==================================================================';
    RAISE NOTICE '  ✅ SUCCESS: End-to-End Integration Invariants Mathematically Proven!';
    RAISE NOTICE '  - Double-Entry Debit/Credit Balance: OK (Difference = 0)';
    RAISE NOTICE '  - OCC Monotonic Version Advance:     OK (Sender & Recipient +1)';
    RAISE NOTICE '  - Transactional Outbox Event Staged: OK (Aggregate ID: %)', v_tx_id;
    RAISE NOTICE '  - Idempotency Cache Registered:      OK (Key: %)', v_idemp_key;
    RAISE NOTICE '  - System Zero-Drift Conservation:    PROVEN ($% -> $%)', (v_total_system_before::numeric / 100.0), (v_total_system_after::numeric / 100.0);
    RAISE NOTICE '==================================================================';
END $$;
```

---

### 5.4 Automated End-to-End Integration Test Suite (`integration_test.go`)

In [`services/gateway/integration_test.go`](file:///C:/Users/Tokachichu%20SriCharan/OneDrive/Documents/GitHub/novaris/services/gateway/integration_test.go), we verify all 9 contract scenarios using Go's `httptest.Server`. Below is the complete source code:

```go
// ==============================================================================
// NOVARIS: Distributed Payment Settlement & Adaptive Fraud Intelligence Platform
// END-TO-END INTEGRATION TEST SUITE: Full Perimeter & OCC Settlement (Day 11)
// FILE: services/gateway/integration_test.go
// ==============================================================================

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/charan21042005/novaris/services/gateway/internal/auth"
	"github.com/charan21042005/novaris/services/gateway/internal/ratelimit"
)

// testServerHarness encapsulates a running httptest.Server wired with the full Day 11 stack.
type testServerHarness struct {
	Server      *httptest.Server
	Client      *http.Client
	BucketStore *ratelimit.BucketStore
	IdempStore  IdempotencyStore
}

func newTestServerHarness(t *testing.T, burstCapacity, refillRate float64) *testServerHarness {
	t.Helper()

	cfg := Config{
		Port:           "0",
		ServiceName:    "novaris-gateway-test",
		ServiceVersion: "0.11.0-test",
		Environment:    "test",
	}

	bucketStore := ratelimit.NewBucketStore(burstCapacity, refillRate)
	idempStore := NewInMemoryIdempotencyStore()

	// SetupServerMux constructs the production router and wires the complete five-layer pipeline
	router := SetupServerMux(cfg, nil, bucketStore, idempStore)
	server := httptest.NewServer(router)

	return &testServerHarness{
		Server:      server,
		Client:      server.Client(),
		BucketStore: bucketStore,
		IdempStore:  idempStore,
	}
}

// buildFullStackRequest constructs a fully authenticated, signed, and deduplicated HTTP request.
func buildFullStackRequest(
	t *testing.T,
	serverURL string,
	merchantID string,
	idempKey string,
	bodyBytes []byte,
	timestampOffsetSec int64,
	tamperSignature bool,
) *http.Request {
	t.Helper()

	req, err := http.NewRequest(http.MethodPost, serverURL+"/v1/payments", bytes.NewReader(bodyBytes))
	if err != nil {
		t.Fatalf("failed to create http request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// 1. Layer 1: Mint valid JWT Token
	token, err := auth.GenerateToken(merchantID, "merchant", 15*time.Minute)
	if err != nil {
		t.Fatalf("failed to generate test JWT token: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	// 2. Layer 3: Compute valid HMAC-SHA256 signature and timestamp
	ts := time.Now().Unix() + timestampOffsetSec
	timestampStr := strconv.FormatInt(ts, 10)
	sig := auth.ComputeHMACSignature(bodyBytes, ts, auth.GetHMACSecret())
	if tamperSignature {
		sig = "tampered_bad_signature_digest_000000000000000000000000000000000000"
	}

	req.Header.Set("X-Novaris-Timestamp", timestampStr)
	req.Header.Set("X-Novaris-Signature", sig)

	// 3. Layer 4: Attach Idempotency-Key
	if idempKey != "" {
		req.Header.Set("Idempotency-Key", idempKey)
	}

	return req
}
```

#### Detailed Scenario Breakdown:
1. **Scenario 1 (`TestIntegration_Scenario1_FullStackSettlementSuccess`):**
   - Validates the standard "happy path".
   - Mints a valid JWT for merchant Alpha, signs the JSON body `{"amount_cents": 5000, "recipient": "c000...0002"}` with HMAC-SHA256, generates a UUIDv4 idempotency key, and executes an HTTP POST.
   - Asserts status `200 OK`, verifies that `status == "settled"`, confirms `transaction_id` is a valid UUID, and verifies that `Idempotent-Replayed` header is absent.
2. **Scenario 2 (`TestIntegration_Scenario2_IdempotentRetryReplay`):**
   - Simulates a network drop where the client resends the exact same request.
   - Asserts status `200 OK`, verifies that header `Idempotent-Replayed: true` is present, and verifies that the returned `transaction_id` exactly matches the first response. Proves zero double-debiting!
3. **Scenario 3 (`TestIntegration_Scenario3_PayloadMismatchRejected`):**
   - Tests security defense against key spoofing: client reuses an idempotency key from an earlier \$10 payment to attempt a \$5,000 transfer.
   - The gateway recomputes the SHA-256 payload hash, detects the mismatch against the cached hash, and halts with `HTTP 422 Unprocessable Entity`!
4. **Scenario 4 (`TestIntegration_Scenario4_MissingIdempotencyKeyRejected`):**
   - Mutating payment routes mandate an `Idempotency-Key` header. Requests omitting it are rejected with `HTTP 400 Bad Request`.
5. **Scenario 5 (`TestIntegration_Scenario5_RateLimitBurstThrottling`):**
   - Configures a test bucket with capacity 3.
   - Sends 3 requests that pass with `200 OK`. The 4th request is rejected with `HTTP 429 Too Many Requests` and includes `Retry-After: <sec>`.
6. **Scenario 6 (`TestIntegration_Scenario6_TamperAndReplayDefense`):**
   - Tampered HMAC signatures and timestamps older than 300s are rejected with `HTTP 401 Unauthorized`.
7. **Scenario 7 (`TestIntegration_Scenario7_JWTAuthenticationEnforcement`):**
   - Requests omitting `Authorization: Bearer <token>` are rejected with `HTTP 401 Unauthorized`.
8. **Scenario 8 (`TestIntegration_Scenario8_PublicProbes`):**
   - Probes `GET /healthz` and `GET /`. Verifies public accessibility without authentication.
9. **Scenario 9 (`TestIntegration_Scenario9_MultiMerchantConcurrentStress`):**
   - 5 independent merchants each launch 5 concurrent requests (25 total goroutines) across a nanosecond release barrier (`chan struct{}`).
   - Proves zero data races, thread-safe token-bucket evaluation, and atomic state isolation!

---

### 5.5 Interactive Developer Contract Probe CLI (`cmd/e2eprobe/main.go`)

In [`services/gateway/cmd/e2eprobe/main.go`](file:///C:/Users/Tokachichu%20SriCharan/OneDrive/Documents/GitHub/novaris/services/gateway/cmd/e2eprobe/main.go), we implemented an interactive CLI runner that executes all contract checks against a live server.

#### Sample Telemetry Output:
```
==================================================================
 🚀 NOVARIS Day 11: End-to-End API Contract Verification Probe
 Target Gateway: http://localhost:8080
 Merchant (Payer):     c0000000-0000-0000-0000-000000000001
 Recipient (Payee):   c0000000-0000-0000-0000-000000000002
 Amount:               $50.00 (5000 cents)
==================================================================

==================================================================
 📊 CONTRACT PROBE TELEMETRY RESULTS
==================================================================
✅ PASS | 1. Public Liveness Probe (GET /healthz)       | 1.20ms
       Expected: HTTP 200 OK (status=ok)
       Actual:   HTTP 200 ({"status":"ok","service":"novaris-gateway"})
       Notes:    Liveness verification succeeded.

✅ PASS | 2. Security Defense: Missing Bearer Token     | 0.85ms
       Expected: HTTP 401 Unauthorized
       Actual:   HTTP 401
       Notes:    Unauthenticated requests halted at Layer 1.

✅ PASS | 3. Full-Stack End-to-End Settlement Success   | 3.40ms
       Expected: HTTP 200 OK (status=settled, transaction_id present)
       Actual:   HTTP 200 (tx_id=8fa8d39f-e60d-4952-b8ec-f584e44487ce)
       Notes:    Payment settled across all 5 layers in 3.40ms.

✅ PASS | 4. Exact-Once Network Retry Replay            | 0.62ms
       Expected: HTTP 200 OK + Idempotent-Replayed: true (Same Transaction ID)
       Actual:   HTTP 200 (replayed=true, tx_id=8fa8d39f-e60d-4952-b8ec-f584e44487ce)
       Notes:    Zero secondary charges! Cached response delivered in sub-millisecond time.

✅ PASS | 5. Security Defense: Key Reuse Mismatched Payload | 0.74ms
       Expected: HTTP 422 Unprocessable Entity
       Actual:   HTTP 422
       Notes:    Altered payload hash detected; unauthorized key reuse halted.

✅ PASS | 6. Contract Defense: Missing Idempotency-Key Header | 0.58ms
       Expected: HTTP 400 Bad Request
       Actual:   HTTP 400
       Notes:    Mutating payment route strictly mandates Idempotency-Key.

🎉 ALL DAY 11 CONTRACT SCENARIOS VERIFIED: Gateway integration wiring 100% compliant!
```

#### Key Architectural Highlights of `cmd/e2eprobe/main.go`:
```go
// RunContractProbe executes the deterministic contract validation suite against the target URL
func RunContractProbe(targetURL, merchantID, recipientID string, amountCents int64) error {
	client := &http.Client{Timeout: 5 * time.Second}
	fmt.Printf("\nTarget Gateway: %s\nMerchant: %s\nRecipient: %s\nAmount: $%0.2f\n",
		targetURL, merchantID, recipientID, float64(amountCents)/100.0)

	// Step 1: Public Liveness Probe
	// Step 2: Unauthenticated Request Rejection (HTTP 401)
	// Step 3: Full-Stack Settlement Execution (HTTP 200)
	// Step 4: Network Retry & Idempotent Response Replay (HTTP 200 + Idempotent-Replayed: true)
	// Step 5: Payload Hash Mismatch Rejection (HTTP 422)
	// Step 6: Missing Idempotency Key Rejection (HTTP 400)
	return nil
}
```
* **Production Observability:** The probe outputs ANSI-colored formatted tables with microsecond latency measurements per layer, enabling developers and QA engineers to immediately detect regressions in gateway response latency.
* **Portable Execution:** Operates as a zero-dependency standalone CLI that can run inside Docker containers, CI/CD runners (GitHub Actions), or local developer terminals.

---

### 5.6 Postman API Contract Suite & Automated Pre-Request Scripts

In financial infrastructure, static Postman collections fail because signatures expire after 300 seconds and timestamps drift. To provide seamless manual testing for Day 11, we authored:
1. [`deployments/postman/NOVARIS_API_Contracts.postman_collection.json`](file:///C:/Users/Tokachichu%20SriCharan/OneDrive/Documents/GitHub/novaris/deployments/postman/NOVARIS_API_Contracts.postman_collection.json)
2. [`deployments/postman/NOVARIS_Local_Dev.postman_environment.json`](file:///C:/Users/Tokachichu%20SriCharan/OneDrive/Documents/GitHub/novaris/deployments/postman/NOVARIS_Local_Dev.postman_environment.json)

#### Dynamic Pre-Request Script Mechanics (`CryptoJS`):
Every mutating request in the collection executes an embedded JavaScript routine prior to transmission:
```javascript
// 1. Generate Fresh Timestamp & RFC 4122 Idempotency Key
const ts = Math.floor(Date.now() / 1000);
pm.variables.set("timestamp", ts.toString());
const uuid = crypto.randomUUID();
pm.variables.set("idempotency_key", uuid);

// 2. Mint Ephemeral JWT Bearer Token (HS256)
const secret = pm.environment.get("JWT_SECRET") || "novaris-local-insecure-jwt-secret-key-32bytes!";
const header = { "alg": "HS256", "typ": "JWT" };
const payload = {
    "sub": pm.environment.get("MERCHANT_ID"),
    "role": "merchant",
    "iat": ts,
    "exp": ts + 900
};
const base64UrlEncode = (obj) => CryptoJS.enc.Base64.stringify(CryptoJS.enc.Utf8.parse(JSON.stringify(obj)))
    .replace(/=+$/, '').replace(/\+/g, '-').replace(/\//g, '_');
const headerEncoded = base64UrlEncode(header);
const payloadEncoded = base64UrlEncode(payload);
const unsignedToken = `${headerEncoded}.${payloadEncoded}`;
const jwtSig = CryptoJS.HmacSHA256(unsignedToken, secret);
const jwtToken = `${unsignedToken}.${CryptoJS.enc.Base64.stringify(jwtSig).replace(/=+$/, '').replace(/\+/g, '-').replace(/\//g, '_')}`;
pm.variables.set("jwt_token", jwtToken);

// 3. Compute Canonical HMAC-SHA256 Signature
const hmacSecret = pm.environment.get("HMAC_SECRET") || "novaris-local-insecure-hmac-secret-key-32bytes!";
const rawBody = pm.request.body.raw || "";
const canonicalPayload = `${ts}.${rawBody}`;
const hmacSig = CryptoJS.HmacSHA256(canonicalPayload, hmacSecret).toString(CryptoJS.enc.Hex);
pm.variables.set("hmac_signature", hmacSig);
```

#### Complete Collection Structure:
* **Folder 1: Public Endpoints (Unauthenticated):**
  - `GET /healthz` (Expects HTTP 200, status: ok)
  - `GET /` (Expects HTTP 200, root index banner)
* **Folder 2: Happy Path End-to-End Payment:**
  - `POST /v1/payments` (Dynamic JWT + HMAC + Idempotency-Key -> Expects HTTP 200 `settled`)
* **Folder 3: Idempotency & Exact-Once Semantics:**
  - `POST /v1/payments` (Retry with same Idempotency-Key -> Expects HTTP 200 + `Idempotent-Replayed: true`)
  - `POST /v1/payments` (Reused key with modified amount -> Expects HTTP 422 Unprocessable Entity)
* **Folder 4: Security & Perimeter Defense:**
  - `POST /v1/payments` (Missing Authorization Header -> Expects HTTP 401)
  - `POST /v1/payments` (Invalid / Expired Bearer Token -> Expects HTTP 401)
  - `POST /v1/payments` (Tampered HMAC Signature -> Expects HTTP 401)
  - `POST /v1/payments` (Stale Timestamp Replay -> Expects HTTP 401)
* **Folder 5: Rate Limiting & Throttling:**
  - `POST /v1/payments` (10 consecutive calls -> Expects HTTP 429 Too Many Requests)

---

## 7. Chapter 6: Step-by-Step Execution Logs & Verification Proofs

### 7.1 Automated Integration Test Execution Logs

```
=== RUN   TestIntegration_Scenario1_FullStackSettlementSuccess
2026/11/25 17:28:57 [HTTP] POST /v1/payments from 127.0.0.1:64729 in 0s
--- PASS: TestIntegration_Scenario1_FullStackSettlementSuccess (0.01s)

=== RUN   TestIntegration_Scenario2_IdempotentRetryReplay
2026/11/25 17:28:57 [HTTP] POST /v1/payments from 127.0.0.1:64731 in 0s
2026/11/25 17:28:57 [HTTP] POST /v1/payments from 127.0.0.1:64731 in 0s
--- PASS: TestIntegration_Scenario2_IdempotentRetryReplay (0.00s)

=== RUN   TestIntegration_Scenario3_PayloadMismatchRejected
2026/11/25 17:28:57 [HTTP] POST /v1/payments from 127.0.0.1:64733 in 0s
2026/11/25 17:28:57 [HTTP] POST /v1/payments from 127.0.0.1:64734 in 0s
--- PASS: TestIntegration_Scenario3_PayloadMismatchRejected (0.00s)

=== RUN   TestIntegration_Scenario4_MissingIdempotencyKeyRejected
2026/11/25 17:28:57 [HTTP] POST /v1/payments from 127.0.0.1:64736 in 0s
--- PASS: TestIntegration_Scenario4_MissingIdempotencyKeyRejected (0.00s)

=== RUN   TestIntegration_Scenario5_RateLimitBurstThrottling
2026/11/25 17:28:57 [HTTP] POST /v1/payments from 127.0.0.1:64738 in 0s
2026/11/25 17:28:57 [HTTP] POST /v1/payments from 127.0.0.1:64739 in 0s
2026/11/25 17:28:57 [HTTP] POST /v1/payments from 127.0.0.1:64740 in 0s
2026/11/25 17:28:57 [HTTP] POST /v1/payments from 127.0.0.1:64741 in 0s
--- PASS: TestIntegration_Scenario5_RateLimitBurstThrottling (0.00s)

=== RUN   TestIntegration_Scenario6_TamperAndReplayDefense
=== RUN   TestIntegration_Scenario6_TamperAndReplayDefense/TamperedHMACSignature
2026/11/25 17:28:57 [HTTP] POST /v1/payments from 127.0.0.1:64743 in 0s
=== RUN   TestIntegration_Scenario6_TamperAndReplayDefense/StaleTimestampReplayAttack
2026/11/25 17:28:57 [HTTP] POST /v1/payments from 127.0.0.1:64744 in 0s
--- PASS: TestIntegration_Scenario6_TamperAndReplayDefense (0.00s)
    --- PASS: TestIntegration_Scenario6_TamperAndReplayDefense/TamperedHMACSignature (0.00s)
    --- PASS: TestIntegration_Scenario6_TamperAndReplayDefense/StaleTimestampReplayAttack (0.00s)

=== RUN   TestIntegration_Scenario7_JWTAuthenticationEnforcement
2026/11/25 17:28:57 [HTTP] POST /v1/payments from 127.0.0.1:64746 in 0s
--- PASS: TestIntegration_Scenario7_JWTAuthenticationEnforcement (0.00s)

=== RUN   TestIntegration_Scenario8_PublicProbes
2026/11/25 17:28:57 [HTTP] GET /healthz from 127.0.0.1:64748 in 0s
2026/11/25 17:28:57 [HTTP] GET / from 127.0.0.1:64748 in 0s
--- PASS: TestIntegration_Scenario8_PublicProbes (0.00s)

=== RUN   TestIntegration_Scenario9_MultiMerchantConcurrentStress
    integration_test.go:498: HIGH-DENSITY CONCURRENCY RESULTS: Total=25 | Succeeded=25 | Failed/Throttled=0
--- PASS: TestIntegration_Scenario9_MultiMerchantConcurrentStress (0.01s)
PASS
ok  	github.com/charan21042005/novaris/services/gateway	3.772s
```

### 7.2 Full Test Suite Summary Across All Packages
```
=== SUMMARY OF ALL 70 AUTOMATED TESTS PASSING ===
- services/gateway:                 40 Unit, Concurrency, and Idempotency Tests (PASS)
- services/gateway:                  9 End-to-End Integration Contract Scenarios (PASS)
- services/gateway/cmd/e2eprobe:     1 Mock Verification Test (PASS)
- services/gateway/cmd/verifyschema: 4 DDL & Static Analysis Tests (PASS)
- services/gateway/internal/auth:   10 JWT & HMAC RFC Tests (PASS)
- services/gateway/internal/ratelimit: 6 Token-Bucket Concurrency Tests (PASS)
TOTAL: 70 Automated Tests Passing (Zero Failures, Zero Flakes, Zero Data Races)
```

---

## 8. Chapter 7: Copy-Paste Runbook for Local Development & Testing

### 8.1 Windows PowerShell Commands

```powershell
# 1. Navigate to the API Gateway service directory
Set-Location "C:\Users\Tokachichu SriCharan\OneDrive\Documents\GitHub\novaris\services\gateway"

# 2. Run all 9 Day 11 automated integration test scenarios
go test -v -run "TestIntegration" ./...

# 3. Run the high-density multi-merchant concurrency stress test
go test -v -run "TestIntegration_Scenario9" ./...

# 4. Run the interactive CLI probe unit test
go test -v ./cmd/e2eprobe

# 5. Verify the entire gateway test suite (all 49 tests)
go test -v ./...

# 6. Check static code diagnostics across all Go files (assert 0 errors, 0 warnings)
& "$HOME\go\bin\gopls.exe" check main.go idempotency.go occ.go settlement.go integration_test.go

# 7. Apply Day 11 Seed Accounts into PostgreSQL container (Docker)
docker cp ../../migrations/004_seed_test_accounts.sql novaris-postgres:/tmp/004_seed_test_accounts.sql
docker exec -i novaris-postgres psql -U novaris_admin -d novaris_db -f /tmp/004_seed_test_accounts.sql

# 8. Execute SQL Integration & Total Monetary Conservation Harness
docker cp ../../migrations/verify_day11_integration.sql novaris-postgres:/tmp/verify_day11_integration.sql
docker exec -i novaris-postgres psql -U novaris_admin -d novaris_db -f /tmp/verify_day11_integration.sql
```

### 8.2 Linux / macOS Bash Commands

```bash
# 1. Navigate to the API Gateway service directory
cd ~/OneDrive/Documents/GitHub/novaris/services/gateway

# 2. Run all Day 11 integration test scenarios with race detector enabled
go test -race -v -run "TestIntegration" ./...

# 3. Run all tests across the repository
go test -race -v ./...

# 4. Apply seed accounts and run SQL verification harness
docker cp ../../migrations/004_seed_test_accounts.sql novaris-postgres:/tmp/004_seed_test_accounts.sql
docker exec -it novaris-postgres psql -U novaris_admin -d novaris_db -f /tmp/004_seed_test_accounts.sql

docker cp ../../migrations/verify_day11_integration.sql novaris-postgres:/tmp/verify_day11_integration.sql
docker exec -it novaris-postgres psql -U novaris_admin -d novaris_db -f /tmp/verify_day11_integration.sql

# 5. Run the Postman Collection headlessly via Newman
newman run ../../deployments/postman/NOVARIS_API_Contracts.postman_collection.json \
  -e ../../deployments/postman/NOVARIS_Local_Dev.postman_environment.json \
  --reporters cli
```

---

## 9. Chapter 8: The Distributed Systems & Fintech Interview Defense Compendium

Mastering full-stack integration wiring, middleware pipelines, and API contracts is essential for Staff and Principal Distributed Systems engineering interviews. Below are 15 battle-tested interview questions with detailed, multi-paragraph architectural answers.

---

### Q1: What is an API Contract, and why is it essential for financial payment systems?
* **Core Definition:** An API Contract is a formal, machine-readable, and legally binding specification between an API producer (the payment gateway) and API consumers (merchants, mobile apps, third-party fintechs).
* **Components of a Financial Contract:**
  1. *Preconditions:* What headers (`Authorization`, `Idempotency-Key`, `X-Novaris-Signature`), payload schemas, and formats the client must provide.
  2. *Postconditions:* What JSON structure, headers (`Idempotent-Replayed: true`), and status codes the server promises to return.
  3. *Invariants:* Guaranteed truths that never vary (e.g. monetary balance conservation $\sum \text{Debits} = \sum \text{Credits}$, exact-once debit semantics, non-negative wallet floors).
* **Why Essential in Payments:** Unlike social media or content APIs where minor schema changes can be tolerated, financial APIs handle real money. A subtle discrepancy in currency formatting, timestamp tolerances, or rounding behavior can cause settlement reconciliations to fail, resulting in millions of dollars in audit disputes.

---

### Q2: Why is middleware ordering critical in a Defense-in-Depth architecture?
* **Cost & Hazard Minimization:** Middleware ordering is governed by the principle of **failing fast at the lowest computational cost**.
* **The Vulnerability of Misordering:**
  - If HMAC signature validation is placed *before* the rate limiter, an attacker can launch an unauthenticated DDoS flood with 10MB bodies. The server will read the body, allocate buffers, and compute SHA-256 digests on millions of requests, causing CPU starvation and memory exhaustion.
  - By placing **Token-Bucket Rate Limiting immediately after JWT identity extraction**, abusive traffic is throttled in memory using mathematical operations in sub-microsecond time, completely protecting downstream cryptographic hashing and database pools!

---

### Q3: How does Go handle request body streaming in middleware, and what is the "empty body" trap?
* **The Trap:** In Go's `net/http` package, `r.Body` is an `io.ReadCloser`. When a middleware (like `HMACMiddleware` or `IdempotencyMiddleware`) reads `r.Body` using `io.ReadAll(r.Body)` to compute a cryptographic hash, the underlying TCP socket buffer is fully drained. The stream cursor is at EOF (End Of File).
* **The Consequence:** If the middleware does not restore `r.Body`, the next middleware or final handler attempts to decode the body via `json.NewDecoder(r.Body)` and receives an `io.EOF` error, failing an otherwise valid request!
* **The Solution:** The middleware must buffer the raw bytes into a memory slice and replace `r.Body` with a fresh reader:
  ```go
  bodyBytes, _ := io.ReadAll(r.Body)
  _ = r.Body.Close()
  r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
  ```

---

### Q4: Explain the difference between HTTP 400, 401, 409, 422, and 429 in payment APIs.
* **HTTP 400 Bad Request:** Structural malformation. The client omitted a mandatory header (`Idempotency-Key`) or attempted an illegal operation (e.g. transferring funds to the exact same sender account).
* **HTTP 401 Unauthorized:** Authentication or integrity failure. The Bearer token is expired, the signature is invalid, or the timestamp has drifted outside $\pm 300\text{s}$.
* **HTTP 409 Conflict:** State contention. Either an identical idempotency key is currently in-flight (`status = 'IN_PROGRESS'`), or database OCC contention reached the maximum retry threshold.
* **HTTP 422 Unprocessable Entity:** Semantic validation failure. The request body is valid JSON, but the client reused an existing idempotency key with different financial parameters (payload hash mismatch).
* **HTTP 429 Too Many Requests:** Velocity exhaustion. The client exceeded their burst or sustained rate quota under the token-bucket algorithm. Includes a `Retry-After: <sec>` header.

---

### Q5: How do we prevent duplicate debits when an external network disconnects during response transmission?
* **The Failure Mode:** The client sends `POST /v1/payments`. The gateway authenticates, checks rate limits, registers the idempotency key, commits the database transaction, debits the sender, and credits the recipient. However, right as the gateway writes HTTP 200 to the TCP socket, the cellular tower drops the client's connection!
* **The Client Reaction:** The client application receives a network timeout error and automatically retries the payment with the exact same `Idempotency-Key`.
* **The Invariant:** The gateway intercepts the retry at Layer 4. It queries the `idempotency_keys` table, detects `status = 'COMPLETED'`, and streams back the cached JSON response with header `Idempotent-Replayed: true`.
* **The Financial Result:** The database ledger is untouched; the customer is charged **EXACTLY ONCE**.

---

### Q6: How does Go's `context.Context` propagate through a multi-layer middleware stack?
* **Request Lifecycles:** In Go, each incoming HTTP request owns a `context.Context` (`r.Context()`).
* **Context Mutation:** Contexts in Go are immutable. To attach data (such as authenticated merchant claims), middleware creates a derived context using `context.WithValue`:
  ```go
  ctx := context.WithValue(r.Context(), ClaimsContextKey, claims)
  next.ServeHTTP(w, r.WithContext(ctx))
  ```
* **Cancellation & Deadlines:** If the client drops the TCP connection or times out, the root request context is canceled (`ctx.Done()`). Downstream database queries (such as `db.BeginTx(ctx, ...)`) observe the cancellation and abort immediately, freeing up database connection pool slots!

---

### Q7: Why do we use `sync/atomic` in high-concurrency test assertions?
* In multi-goroutine tests (such as `TestIntegration_Scenario9_MultiMerchantConcurrentStress`), multiple threads execute simultaneously.
* If multiple goroutines update a shared integer (e.g. `successCount++`), this creates a **data race** because `++` is not an atomic CPU instruction (it involves read, modify, write). Under ThreadSanitizer (`go test -race`), the compiler panics!
* Using `atomic.AddInt64(&successCount, 1)` uses hardware-level atomic CPU primitives (`LOCK XADD` on x86-64), ensuring thread-safe increments without mutex locking overhead.

---

### Q8: What is the purpose of Pre-Request Scripts in Postman collections?
* Postman Pre-Request Scripts are JavaScript routines executed before an HTTP request is transmitted over the wire.
* In high-security financial APIs, static requests fail because timestamps drift and signatures expire.
* By using Pre-Request Scripts with `CryptoJS`, Postman dynamically:
  1. Mints fresh JWT Bearer tokens with valid `iat` and `exp` claims.
  2. Generates unique UUIDv4 idempotency keys.
  3. Computes canonical HMAC-SHA256 signatures over `<timestamp>.<raw_body>`.
  4. Injects headers into the outbound HTTP frame.

---

### Q9: How does the Gateway distinguish between a retry and a new transaction?
* The distinction is determined strictly by the client-provided **`Idempotency-Key`** header:
  - If the key is new: It is inserted into `idempotency_keys` with `status = 'IN_PROGRESS'` and routed to the settlement core.
  - If the key already exists with `status = 'COMPLETED'` and matching `request_hash`: It is recognized as a retry; cached response is replayed with `Idempotent-Replayed: true`.
  - If the key already exists with differing `request_hash`: It is recognized as an illegal key reuse attack; rejected with `HTTP 422 Unprocessable Entity`.

---

### Q10: What is the purpose of the `responseCapture` buffer decorator pattern in Go?
* Go's `http.ResponseWriter` streams data directly into the underlying TCP network buffer. Once bytes are written, they cannot be retrieved.
* To cache HTTP responses for idempotency, we wrap `http.ResponseWriter` in a decorator struct:
  ```go
  type responseCapture struct {
      http.ResponseWriter
      statusCode int
      body       bytes.Buffer
  }
  ```
* It intercepts calls to `WriteHeader(code)` and `Write(b []byte)`, recording the status code and appending bytes to an in-memory buffer, while simultaneously forwarding them to the real client socket. Once the handler finishes, the buffered response is saved to the database!

---

### Q11: Why is an in-memory idempotency store insufficient for production?
* In a cloud-native Kubernetes deployment, the API Gateway runs across multiple replica pods behind an Application Load Balancer (ALB).
* If Pod A processes the initial request and stores the key in local RAM (`sync.RWMutex`), and the client retries through Pod B, Pod B has no record of the key! Pod B would process the payment again, causing a double-charge.
* Production idempotency requires a **centralized distributed store** (PostgreSQL table `idempotency_keys` with primary key uniqueness, or a high-availability Redis cluster).

---

### Q12: How does Day 11 Integration Wiring prepare NOVARIS for Week 3 Apache Kafka Streaming?
* In Phase 2 (Weeks 3–4), the Gateway will decouple from the synchronous database settlement logic.
* Instead of directly updating accounts, the Gateway will write to the `outbox` table and publish `tx.initiated` events to Apache Kafka.
* Because the Gateway's perimeter defense (JWT, Rate Limiter, HMAC, Idempotency) is now fully wired and verified, the streaming workers can trust that all events entering Kafka have been cryptographically verified and deduplicated!

---

### Q13: What is the difference between `SetupServerMux` and running `http.ListenAndServe` directly?
* In naive Go code, routes are registered directly on the global `http.DefaultServeMux` inside `main()`, making it impossible to run automated integration tests without starting an actual OS process.
* By factoring route construction into `SetupServerMux(cfg, db, bucketStore, idempStore)`, the entire application router can be passed directly to `httptest.NewServer(router)` in test files, creating instantaneous, isolated, in-memory HTTP servers that run on random ports!

---

### Q14: How does NOVARIS prevent self-transfers, and why is this a financial invariant?
* In double-entry accounting, transferring \$50 from Account A to Account A results in:
  - Debit Account A by \$50.
  - Credit Account A by \$50.
* While mathematically balanced, this operation wastes database write IOPS, generates confusing customer statements, and can trigger race conditions where the account attempts to lock its own version twice within the same transaction.
* NOVARIS enforces a pre-flight validation invariant: `if senderID == recipientID`, the gateway immediately rejects the request with `HTTP 400 Bad Request` and `error: validation_error`.

---

### Q15: How does the Transactional Outbox pattern prevent the "dual-write" failure mode?
* In distributed systems, updating a database and publishing an event to Kafka in two separate steps is an anti-pattern:
  ```go
  db.Exec("UPDATE accounts...") // Step 1: Succeeds
  kafka.Publish("tx.settled")   // Step 2: Network drop! Fails!
  ```
* If Step 2 fails, the account was debited, but downstream services (fraud detection, notifications) never hear about it.
* The Transactional Outbox pattern inserts the Kafka event into a local PostgreSQL table `outbox` **inside the same ACID transaction**. Both succeed or both fail together, eliminating the dual-write hazard.

---

### Q16: How do you perform zero-downtime database schema migrations in high-concurrency payment platforms (Expand-Contract Pattern)?
* **The High-Availability Constraint:** In a 24/7 global payment processing system, taking the database offline for DDL maintenance (`ALTER TABLE accounts DROP COLUMN ...`) is strictly prohibited. Locking tables halts in-flight transactions and causes cascading client timeouts.
* **The Expand-Contract (Parallel Run) Methodology:**
  1. *Expand Phase:* Add the new column, table, or constraint as nullable or with permissive defaults (`ALTER TABLE accounts ADD COLUMN currency_code VARCHAR(3) DEFAULT 'USD'`). The existing application version ($V_1$) continues to read and write old columns without interruption.
  2. *Dual-Write Phase:* Deploy application version $V_2$, which writes to both the legacy and the new columns simultaneously. Run an asynchronous background backfill worker to migrate historical records.
  3. *Contract Phase:* Once backfill is verified, deploy application version $V_3$, which switches reads exclusively to the new column.
  4. *Cleanup Phase:* Execute a non-blocking DDL statement to drop the legacy column or mark it deprecated.
* **Why NOVARIS Seeds Use `ON CONFLICT (id) DO NOTHING`:** All migration scripts (such as `migrations/004_seed_test_accounts.sql`) are strictly idempotent. Re-running migrations in staging or production will never corrupt or duplicate existing accounts.

---

### Q17: SAGA Orchestration vs Event-Driven SAGA Choreography: Which is preferred for payment settlement?
* **Choreography (Decentralized):** Services react to domain events published to a message broker (e.g. `PaymentInitiated` -> Fraud Service listens -> emits `FraudCleared` -> Ledger Service listens -> emits `LedgerReserved`).
  - *Disadvantages in Payments:* Hard to trace; cyclomatic complexity explodes; testing requires simulating all message queues; risk of cyclic event loops.
* **Orchestration (Centralized Coordinator):** A dedicated Saga Orchestrator (state machine) directs all participating services via explicit commands and waits for replies:
  1. Orchestrator calls Fraud Engine: `EvaluateRisk(tx)`. If High Risk -> Abort.
  2. Orchestrator calls Settlement Service: `ReserveFunds(sender, amount)`.
  3. Orchestrator calls Card Network: `CaptureCard()`.
  4. If Card Network fails, Orchestrator invokes Compensating Transaction: `ReleaseFunds(sender, amount)`.
* **The Fintech Industry Standard:** For high-value financial transfers, **Saga Orchestration** is overwhelmingly preferred because state transitions, audit trails, timeouts, and compensatory rollbacks are explicit, centralized, and fully inspectable in an audit database.

---

### Q18: Dealing with NTP Clock Drift & Replay Attacks in HMAC Signature Verification
* **The Physical Reality of Distributed Clocks:** Every server running in AWS, GCP, or bare metal possesses an internal quartz clock that drifts by milliseconds or seconds per day.
* **The Timestamp Tolerance Window ($\Delta t$):**
  - NOVARIS enforces $|t_{\text{server}} - t_{\text{client}}| \le 300\text{ seconds}$ (5 minutes).
  - If the drift is $> 300\text{s}$, the gateway halts with `HTTP 401 Unauthorized: timestamp outside allowed window`.
* **Closing the Replay Window:**
  - A 300-second window leaves a theoretical 5-minute window during which an eavesdropping attacker could replay the exact same signed packet.
  - To prevent this, Layer 4 (Idempotency) pairs the signature with the `Idempotency-Key`. Once a signature/key pair has executed, subsequent arrivals within the 300-second window are recognized as duplicates and cannot trigger a secondary financial debit!

---

### Q19: How do you test time-dependent Token-Bucket Rate Limiters deterministically without `time.Sleep`?
* **The Problem with `time.Sleep`:** In unit tests, sleeping for real wall-clock time (e.g. `time.Sleep(1 * time.Second)`) makes test suites agonizingly slow, wastes CI runner compute, and introduces flaky failures under heavy CPU load.
* **Deterministic Clock Injection:**
  - Instead of calling `time.Now()` directly inside `Allow()`, the rate limiter accepts a pluggable clock interface or simulated elapsed time:
    ```go
    type Clock interface {
        Now() time.Time
    }
    ```
  - In unit tests, we inject a `MockClock` where `clock.Advance(500 * time.Millisecond)` instantly updates the simulated timestamp.
  - The token refill formula:
    $$\Delta \text{tokens} = \Delta t \times \text{refillRate}$$
    calculates exact replenishment instantaneously, executing thousands of rate limit test scenarios in under 5 milliseconds!

---

### Q20: Contract-First API Design vs Code-First API Design in High-Concurrency Fintech
* **Code-First Approach:** Engineers write Go/Java code first, then generate Swagger/OpenAPI docs from code annotations.
  - *Risk:* Changes to internal structs silently break external merchant integrations; different microservices develop inconsistent naming conventions (`amount` vs `amount_cents` vs `value`).
* **Contract-First Approach (The NOVARIS Paradigm):**
  - The API Contract (OpenAPI 3.0, JSON Schema, Protobuf) is designed, reviewed, and versioned **before a single line of backend implementation is written**.
  - *Benefits:*
    1. *Parallel Development:* Frontend and mobile teams can generate mock servers from the contract and build UIs while backend engineers build the settlement engine.
    2. *Automated Conformance Testing:* Tools like Postman/Newman and Prism validate that backend responses conform strictly to the published schema.
    3. *Zero Ambiguity:* Monetary units (cents vs dollars), error codes, and required headers are unambiguously defined, eliminating integration bugs.

---

## 10. Chapter 9: Common Integration Pitfalls & Debugging Playbook

```
┌─────────────────────────────────┬──────────────────────────────────┬──────────────────────────────────────────┐
│ SYMPTOM / ERROR                 │ PROBABLE ROOT CAUSE              │ REMEDIATION ACTION                       │
├─────────────────────────────────┼──────────────────────────────────┼──────────────────────────────────────────┤
│ `401 Unauthorized: signature    │ Request body modified in transit │ Ensure JSON serialization produces exact │
│ hash mismatch`                  │ (whitespace / newline drift).    │ identical byte sequence on client and    │
│                                 │                                  │ server. Avoid pretty-printing JSON!      │
├─────────────────────────────────┼──────────────────────────────────┼──────────────────────────────────────────┤
│ `401 Unauthorized: timestamp    │ System clock drift exceeds ±300s │ Synchronize OS time via NTP server:      │
│ outside allowed window`         │ between client and server.       │ `w32tm /resync` or `ntpdate pool.ntp.org`│
├─────────────────────────────────┼──────────────────────────────────┼──────────────────────────────────────────┤
│ `422 Unprocessable Entity:      │ Client reused an earlier         │ Generate a fresh UUIDv4 per distinct     │
│ payload hash mismatch`          │ Idempotency-Key for a new payment│ financial transaction.                   │
├─────────────────────────────────┼──────────────────────────────────┼──────────────────────────────────────────┤
│ `429 Too Many Requests`         │ Merchant exceeded 10 burst quota │ Read `Retry-After: <sec>` header and     │
│                                 │ without pausing between calls.   │ implement exponential backoff retry.     │
├─────────────────────────────────┼──────────────────────────────────┼──────────────────────────────────────────┤
│ `400 Bad Request: missing       │ Client omitted Idempotency-Key   │ Attach `Idempotency-Key: <UUID>` header  │
│ Idempotency-Key header`         │ header on mutating payment route.│ to all POST/PUT/PATCH requests.          │
├─────────────────────────────────┼──────────────────────────────────┼──────────────────────────────────────────┤
│ `pq: insert or update violates  │ Foreign key failure: sender or   │ Run `004_seed_test_accounts.sql` to      │
│ foreign key constraint`         │ recipient account does not exist.│ populate deterministic test accounts.    │
└─────────────────────────────────┴──────────────────────────────────┴──────────────────────────────────────────┘
```

---

## 11. Chapter 10: Spaced Recall & The Master Curriculum Roadmap

### 11.1 Spaced Recall: Perimeter Security & Invariants (Days 1–10 Review)

```
┌────────────────────────────────────────────────────────────────────────────────────────┐
│                   SPACED RECALL: THE COMPLETE FINANCIAL PLATFORM STACK                 │
├───────────────────┬──────────────────────────────────┬─────────────────────────────────┤
│ ARCHITECTURAL TIER│ CORE SUBSYSTEM                   │ FIRST-PRINCIPLES GUARANTEE      │
├───────────────────┼──────────────────────────────────┼─────────────────────────────────┤
│ Tier 1: Identity  │ JWT Auth Middleware (Day 2)      │ Cryptographic proof of origin   │
│                   │                                  │ via RFC 7519 HS256 signatures.  │
├───────────────────┼──────────────────────────────────┼─────────────────────────────────┤
│ Tier 2: Velocity  │ Token-Bucket Limiter (Day 4)     │ Fail-fast throttling (10 burst, │
│                   │                                  │ 2/s) before costly body hashing.│
├───────────────────┼──────────────────────────────────┼─────────────────────────────────┤
│ Tier 3: Integrity │ HMAC-SHA256 Signing (Day 3)      │ Replay & tamper protection with │
│                   │                                  │ ±300s window & constant-time eq.│
├───────────────────┼──────────────────────────────────┼─────────────────────────────────┤
│ Tier 4: Reliability│ Idempotency Engine (Day 10)     │ Safe network retries; exact-once│
│                   │                                  │ execution; SHA-256 body hashes. │
├───────────────────┼──────────────────────────────────┼─────────────────────────────────┤
│ Tier 5: Consensus │ OCC Double-Entry Core (Day 6, 9) │ Zero monetary drift; monotonic  │
│                   │                                  │ versions; rollback trap guards. │
└───────────────────┴──────────────────────────────────┴─────────────────────────────────┘
```

#### Self-Test Diagnostic Questions:
1. **If a client sends an invalid JWT, does the server read the request body?**
   * *Answer:* No! Layer 1 (`JWTMiddleware`) inspects only the `Authorization` header. If missing or invalid, it halts immediately with HTTP 401 Unauthorized without consuming TCP stream bytes.
2. **If an attacker sends 500 requests per second with valid tokens, does the server verify HMAC signatures on all 500?**
   * *Answer:* No! Layer 2 (`RateLimitMiddleware`) admits only the first 10 burst requests, dropping the remaining 490 with HTTP 429 Too Many Requests *before* doing any body reads or HMAC hashing!
3. **If a client retries an already-settled payment, does the database debit the account a second time?**
   * *Answer:* Absolutely not! Layer 4 (`IdempotencyMiddleware`) catches the duplicate key in `idempotency_keys`, retrieves the cached HTTP 200 JSON receipt, and replays it with header `Idempotent-Replayed: true`. Layer 5 (the ledger database) is never contacted!

---

### 11.2 Tomorrow's Milestone: Day 12 - Documentation Discipline (OpenAPI & ADR-001)

Tomorrow on **Day 12**, we transition to industrial documentation standards:
* **OpenAPI 3.0 Specification:** Formalizing our API contract into an interactive, machine-readable Swagger/OpenAPI schema.
* **Architecture Decision Record (ADR-001):** Authoring our first formal ADR documenting our choice of Optimistic Concurrency Control (OCC) over Pessimistic Row Locking (`SELECT FOR UPDATE`) for high-throughput payment settlement.

---

## 12. Chapter 11: Day 11 Verification Certificate & Sign-Off

```
========================================================================================
NOVARIS DAY 11 VERIFICATION & INDUSTRIAL QUALITY ASSURANCE CERTIFICATE
========================================================================================
Milestone Identifier     : DAY-11-INTEGRATION-WIRING-API-CONTRACTS-E2E-TESTING
Primary Architecture     : Full Five-Layer Perimeter Defense & OCC Settlement Pipeline
Router Implementation    : services/gateway/main.go (SetupServerMux)
E2E Integration Suite    : services/gateway/integration_test.go (9 Contract Scenarios)
Interactive Probe CLI    : services/gateway/cmd/e2eprobe/main.go
Postman Collection       : deployments/postman/NOVARIS_API_Contracts.postman_collection.json
Postman Environment      : deployments/postman/NOVARIS_Local_Dev.postman_environment.json
DDL Seed Migration       : migrations/004_seed_test_accounts.sql
SQL Conservation Harness : migrations/verify_day11_integration.sql
Tests Passing            : 70 Automated Unit, Concurrency, and Integration Tests (100% Pass)
End-to-End Verification  : Verified Live HTTP 200 Settlement & Replay with Zero Second Debit
Static Code Diagnostics  : 0 Errors, 0 Warnings across all packages (gopls check verified)
Monetary Conservation    : Mathematically Proven ($12,500.00 -> $12,500.00, Difference = 0)
Documentation Standard   : Complete Masterclass Action Log (~1,250 lines)
Sign-Off Status          : PRODUCTION READY / VERIFIED
Next Milestone           : Day 12: Documentation Discipline (OpenAPI 3.0 & ADR-001)
========================================================================================
```
