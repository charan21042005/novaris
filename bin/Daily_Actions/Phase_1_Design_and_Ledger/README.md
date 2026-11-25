<div align="center">

<!-- Top Banner -->
<p align="center">
  <img src="../../../assets/banner.svg" alt="NOVARIS Banner" width="100%" />
</p>

# 🏛️ Phase 1: Design, Ledger & Perimeter Defense
### Master Engineering Action Logs • Days 01 through 14
#### Establishing the Relational Foundation, Cryptographic Security, Idempotent Settlement, and OCC Consensus

<p align="center">
  <img src="https://img.shields.io/badge/Phase-1_of_6:_Design_%26_Ledger-00E5FF?style=flat-square&logo=gitbook&logoColor=white" alt="Phase" />
  <img src="https://img.shields.io/badge/Milestones-Days_01_to_14-00C853?style=flat-square&logo=target&logoColor=white" alt="Milestones" />
  <img src="https://img.shields.io/badge/Completed-Days_01_to_11-7C4DFF?style=flat-square&logo=checkmarx&logoColor=white" alt="Completed" />
  <img src="https://img.shields.io/badge/Standards-RFC_7519_%7C_RFC_2104_%7C_IETF_Idempotency-FFA000?style=flat-square&logo=securityscorecard&logoColor=white" alt="Standards" />
</p>

---

</div>

## 📌 Phase 1 Overview: Building the Unshakeable Financial Core

In distributed fintech systems, speed without correctness is negligence. **Phase 1 (Weeks 1 & 2)** of the NOVARIS engineering roadmap establishes the zero-drift, tamper-evident foundation upon which all downstream asynchronous streaming (Kafka), fraud scoring (XGBoost/ONNX), and distributed consensus (Raft/CockroachDB) are anchored.

### Core Architectural Guarantees Established in Phase 1:
1. **Double-Entry Bookkeeping:** Money is never created or destroyed out of thin air. Every credit requires an equal and offsetting debit ($\sum \text{Debits} = \sum \text{Credits}$).
2. **Stateless Identity Verification:** RFC 7519 JSON Web Tokens (JWT) verified at Layer 1 with strict `alg: HS256` validation to prevent algorithm confusion attacks.
3. **Cryptographic Anti-Tamper & Anti-Replay:** RFC 2104 HMAC-SHA256 request signatures verified in constant-time with a sliding timestamp tolerance window ($\pm 300\text{s}$).
4. **Fail-Fast Velocity Throttling:** Lazy mathematical Token-Bucket rate limiting (10 burst capacity, 2 req/s refill) in RAM, dropping volumetric floods before costly payload parsing.
5. **Optimistic Concurrency Control (OCC):** High-throughput settlement scaling without row-level lock contention (`SELECT FOR UPDATE`), enforced via atomic SQL `WHERE id = $1 AND version = $2` conditional writes.
6. **Exact-Once Payment Execution:** IETF-compliant `Idempotency-Key` deduplication, SHA-256 canonical payload hashing, and sub-millisecond cached response replays.
7. **Unified Defense-in-Depth Pipeline:** Complete five-layer middleware integration connecting ingress routing, perimeter defense, and database settlement consensus into an unbroken execution path.

---

## ⚡ Master Testing & CLI Verification Runbook

All copy-pasteable testing commands, Docker deployment guides, Newman automated runners, and Go test invocations across all Phase 1 milestones are centralized in the Master Runbook:
👉 **[`../COMMANDS_RUNBOOK.md`](../COMMANDS_RUNBOOK.md)**

---

## 🗺️ Phase 1 Daily Action Logs Index (Days 01–14)

| Day | Milestone Focus | Core CS & Architectural Concept | Action Log Document | Status |
| :---: | :--- | :--- | :---: | :---: |
| **Day 01** | **Monorepo Layout & Ledger Schema** | PostgreSQL 16 double-entry tables, Docker Compose Alpine mesh, public `/healthz` probe | [Day 01 Action Log](./Day_01_Action_Log.md) | ✅ *COMPLETED* |
| **Day 02** | **JWT Authentication & Middleware** | RFC 7519 token minting, claim extraction, context propagation, and `alg: none` defenses | [Day 02 Action Log](./Day_02_Action_Log.md) | ✅ *COMPLETED* |
| **Day 03** | **HMAC Request Signing & Replay Defense** | RFC 2104 canonical digest computation, constant-time `hmac.Equal`, and $\pm 300\text{s}$ timestamp drift window | [Day 03 Action Log](./Day_03_Action_Log.md) | ✅ *COMPLETED* |
| **Day 04** | **Token-Bucket Rate Limiter & Concurrency** | In-memory lazy mathematical token refill, `sync.RWMutex` thread safety, and HTTP 429 throttling | [Day 04 Action Log](./Day_04_Action_Log.md) | ✅ *COMPLETED* |
| **Day 05** | **Full Ledger Schema & Composite Indexing** | Human fraud review queues, feature snapshots, GIN indexing, and B-Tree composite access paths | [Day 05 Action Log](./Day_05_Action_Log.md) | ✅ *COMPLETED* |
| **Day 06** | **Atomic Settlement Logic & Go Transactions** | ACID multi-table double-entry writes, `defer tx.Rollback()`, and balance conservation invariants | [Day 06 Action Log](./Day_06_Action_Log.md) | ✅ *COMPLETED* |
| **Day 07** | **Testing Discipline & Table-Driven Tests** | Table-driven testing patterns, `go-sqlmock` database isolation, and subtest parallelism | [Day 07 Action Log](./Day_07_Action_Log.md) | ✅ *COMPLETED* |
| **Day 08** | **Concurrency Stress Testing & Race Detection** | High-density goroutines, nanosecond start-gate barriers (`chan struct{}`), and ThreadSanitizer | [Day 08 Action Log](./Day_08_Action_Log.md) | ✅ *COMPLETED* |
| **Day 09** | **Optimistic Concurrency Control (OCC)** | Monotonic version columns, conditional check-and-increment writes, and jittered backoff retries | [Day 09 Action Log](./Day_09_Action_Log.md) | ✅ *COMPLETED* |
| **Day 10** | **Idempotent API Design & Network Retries** | IETF `Idempotency-Key` headers, SHA-256 payload hashing, response buffer capture, and cached replays | [Day 10 Action Log](./Day_10_Action_Log.md) | ✅ *COMPLETED* |
| **Day 11** | **Integration Wiring & End-to-End API Contracts** | Five-layer pipeline assembly, Postman API contract test suite, deterministic seed accounts, and live E2E probe | [Day 11 Action Log](./Day_11_Action_Log.md) | ✅ *COMPLETED* |
| **Day 12** | **Documentation Discipline: OpenAPI & ADR-001** | OpenAPI 3.0 specification, Swagger contract enforcement, and Architecture Decision Record (OCC vs Pessimistic) | *Pending* | ⏳ *Up Next* |
| **Day 13** | **Edge-Case Hardening & Payload Fuzzing** | Extreme boundary testing, system clock skew simulations, malicious payload fuzzing, and security audits | *Pending* | 📅 *Upcoming* |
| **Day 14** | **Phase 1 Capstone Review & Certification** | Full Phase 1 retrospective, end-to-end load benchmarking, and formal architectural sign-off | *Pending* | 📅 *Upcoming* |

---

## 🏛️ Phase 1 Five-Layer Defensive Perimeter Architecture

```
[ Inbound HTTP Payment Request ]
               │
               ▼
┌────────────────────────────────────────┐
│ Layer 1: JWT Authentication (Day 02)   │ ──► [❌ 401 Unauthorized: Invalid/Expired Token]
└────────────────────────────────────────┘
               │ (Identity Verified: MerchantID injected into Context)
               ▼
┌────────────────────────────────────────┐
│ Layer 2: Token-Bucket Limiter (Day 04) │ ──► [❌ 429 Too Many Requests: Quota Exhausted]
└────────────────────────────────────────┘
               │ (Burst & Velocity Admitted in Nanoseconds)
               ▼
┌────────────────────────────────────────┐
│ Layer 3: HMAC-SHA256 Anti-Tamper (D03) │ ──► [❌ 401 Unauthorized: Tampered Payload or Stale Replay]
└────────────────────────────────────────┘
               │ (Payload Integrity & Freshness Verified)
               ▼
┌────────────────────────────────────────┐
│ Layer 4: Idempotency Deduplication (D10)│ ──► [🔁 200 OK: Cached Replay (Idempotent-Replayed: true)]
└────────────────────────────────────────┘ ──► [❌ 422 Unprocessable: Key Reuse Mismatch]
               │ (Fresh Mutating Transaction)
               ▼
┌────────────────────────────────────────┐
│ Layer 5: OCC Ledger Settlement (D06,09)│ ──► [❌ 409 Conflict: High Contention Retry Exhausted]
└────────────────────────────────────────┘
               │ (Committed: Money Conserved, Version +1, Outbox Event Staged)
               ▼
[ Response Buffered & Cached in idempotency_keys ──► 200 OK Streamed to Merchant ]
```

---

<div align="center">

<!-- Bottom Banner -->
<p align="center">
  <img src="../../../assets/footer.svg" alt="NOVARIS Footer" width="100%" />
</p>

<p align="center">
  <strong>NOVARIS Engineering Action Logs • Phase 1 Vault • Maintained by <a href="https://github.com/charan21042005">Tokachichu Sri Charan</a></strong>
</p>

</div>
