# ADR-0001: Optimistic Concurrency Control (OCC) vs. Pessimistic Row Locking for Ledger Settlement

* **Status:** Accepted
* **Date:** 2026-11-25
* **Deciders:** Tokachichu Sri Charan (Core Systems Architect)
* **Technical Area:** Distributed Database Ledger, ACID Transactions, High-Concurrency Settlement
* **PR / Milestone:** Day 09 & Day 12 Architecture Milestone

---

## 1. Context & Problem Statement

Modern payment platforms process thousands of transactions per second (TPS). In double-entry financial bookkeeping, a settlement operation requires debiting a sender account, crediting a recipient account, and appending immutable journal entries to an audit log.

Under high concurrency, multiple simultaneous transactions attempt to update the same financial accounts (e.g. popular merchant hot-wallets, marketplace escrow accounts, payment aggregators). 

If two transactions concurrently read Account A's balance ($B = \$100$), compute a debit (\$40 and \$70 respectively), and write the result back unconditionally, a **Lost Update anomaly** occurs. Account A would be debited by only one transaction while both recipients are credited, creating money out of thin air and causing balance drift ($\sum \text{Debits} \ne \sum \text{Credits}$).

We must choose a concurrency control strategy that guarantees strict balance isolation while maximizing throughput and preventing database connection pool exhaustion.

---

## 2. Alternatives Considered

We evaluated four candidate concurrency architectures:

### Option A: Pessimistic Row-Level Locking (`SELECT FOR UPDATE`)
* **Mechanism:** The transaction locks the account rows upon initial read (`SELECT balance_cents FROM accounts WHERE id = $1 FOR UPDATE`).
* **Why Rejected:**
  1. *Connection Pool Exhaustion:* Locks are held across the entire lifetime of the transaction, including network round-trips, outbox event staging, and potential external fraud evaluations. Connection pools quickly starve.
  2. *Deadlock Hazards:* When two bidirectional transfers occur simultaneously (Transfer 1 locks A then attempts B; Transfer 2 locks B then attempts A), a classic AB-BA circular wait deadlock occurs, forcing the database engine to abort transactions.
  3. *Latency Spikes:* Serializing hot-wallet updates behind exclusive row locks introduces catastrophic tail latency ($p99 > 2.5\text{s}$) under burst traffic.

### Option B: Distributed Locks via Redis (Redlock)
* **Mechanism:** Acquire a distributed mutex key in Redis before touching the PostgreSQL database.
* **Why Rejected:**
  1. *Martin Kleppmann's GC Pause Anomaly:* A process that acquires a Redis lock may experience a garbage collection pause or network blip that outlasts the lock TTL. The lock expires, a second process acquires it, and both write concurrently to the database.
  2. *Dual-State Inconsistency:* Adds an external state dependency that must be coordinated with database transactions.

### Option C: Single-Threaded In-Memory Partitioning (LMAX Disruptor Pattern)
* **Mechanism:** Route all transactions for a given account to a single dedicated CPU core/goroutine queue, eliminating locks entirely.
* **Why Rejected:** While blazingly fast, it requires complex sharding, cross-partition coordination for transfers spanning two shards, and high operational overhead unsuitable for Phase 1.

### Option D: Database-Level Optimistic Concurrency Control (OCC) with Version Columns (CHOSEN)
* **Mechanism:**
  1. Accounts carry an integer `version` column initialized to `1`.
  2. The gateway reads account balances and versions without acquiring row locks.
  3. The update statement uses an atomic conditional check-and-increment:
     ```sql
     UPDATE accounts
     SET balance_cents = balance_cents - $1,
         version = version + 1,
         updated_at = NOW()
     WHERE id = $2 AND version = $3 AND balance_cents >= $1;
     ```
  4. If `RowsAffected == 0`, a collision occurred (another transaction updated the account first). The gateway catches the zero-row return, safely rolls back the transaction, and retries with randomized exponential backoff.

---

## 3. Decision Outcome

We formally adopt **Option D: Database-Level Optimistic Concurrency Control (OCC) with Jittered Exponential Backoff**.

### Key Implementation Principles:
1. **Schema Requirement:** The `accounts` table mandates an `INTEGER NOT NULL DEFAULT 1` column named `version`.
2. **Atomic Write Invariant:** Every mutating SQL statement affecting an account balance MUST include `AND version = $expected_version`.
3. **Collision Detection:** Code must inspect the driver's `result.RowsAffected()`. If zero rows were updated, it is treated as a transient concurrency conflict.
4. **Retry Discipline:** OCC collisions are retried up to `MaxRetries = 3` with exponential backoff:
   $$\text{Backoff} = \text{BaseBackoff} \times 2^{\text{attempt}} + \text{rand}(0, \text{Jitter})$$
5. **Rollback Traps:** Every transaction begins with `defer tx.Rollback()`. If all retry attempts are exhausted, the transaction safely unwinds and returns `HTTP 409 Conflict`.

---

## 4. Consequences & Trade-offs

### Positive Consequences:
* **Zero Database Deadlocks:** Transactions do not hold long-lived exclusive locks across multiple accounts. Circular wait conditions cannot form.
* **High Read & Write Concurrency:** Uncontested transactions execute at maximum database speed with minimal latency ($p99 < 5\text{ms}$).
* **Connection Pool Protection:** Database connections are acquired, utilized for atomic microsecond writes, and immediately returned to the pool.
* **Mathematical Invariant Preservation:** Balance floors (`balance_cents >= $1`) and monotonic version increases ($V \to V+1$) are guaranteed at the SQL engine level.

### Negative Consequences & Mitigations:
* **Retry Overhead Under Extreme Hot-Spotting:** If 100 concurrent requests target the exact same merchant account, 99 will collide on the first attempt and enter the retry loop.
  * *Mitigation 1:* Jittered exponential backoff distributes retries across time, breaking lockstep synchronization.
  * *Mitigation 2:* In Phase 4, hot accounts will leverage asynchronous batching and queue aggregation.
* **Client Retry Visibility:** If maximum retries (3) are exceeded, the client receives `HTTP 409 Conflict`.
  * *Mitigation:* The API contract explicitly documents that clients must retry `409 Conflict` using the same `Idempotency-Key`.

---

## 5. Compliance & Invariant Verification

1. **Balance Conservation Invariant:**
   $$\sum \text{Debits} = \sum \text{Credits} \quad \implies \quad \Delta \text{SystemBalance} = 0$$
2. **Version Monotonicity Invariant:**
   $$V_{\text{after}} = V_{\text{before}} + 1$$
3. **Test Proofs:** Verified by `services/gateway/occ_test.go`, `concurrency_test.go` (50 concurrent goroutines with 0 data races), and `migrations/verify_day11_integration.sql`.
