<div align="center">

<!-- Top Banner -->
<p align="center">
  <img src="../assets/banner.svg" alt="NOVARIS Banner" width="100%" />
</p>

# 📑 NOVARIS • Database Migrations (`/migrations`)
### Immutable Double-Entry Ledger Schema & Transactional Outbox DDL

<p align="center">
  <img src="https://img.shields.io/badge/Database-PostgreSQL_16-4169E1?style=flat-square&logo=postgresql&logoColor=white" alt="PostgreSQL" />
  <img src="https://img.shields.io/badge/Accounting-Double--Entry_Zero--Drift-00C853?style=flat-square&logo=cashapp&logoColor=white" alt="Accounting" />
  <img src="https://img.shields.io/badge/Concurrency-Optimistic_Locking_(OCC)-00E5FF?style=flat-square&logo=affinitylock&logoColor=white" alt="Concurrency" />
  <img src="https://img.shields.io/badge/Pattern-Transactional_Outbox-7C4DFF?style=flat-square&logo=apachekafka&logoColor=white" alt="Outbox" />
  <img src="https://img.shields.io/badge/Indexing-Composite_B--Tree_+_GIN-FF6D00?style=flat-square&logo=speedtest&logoColor=white" alt="Indexing" />
</p>

---

</div>

## 📌 Financial Design Principles

This directory houses the version-controlled SQL migrations defining NOVARIS's persistent data models. The schema strictly enforces banking-grade financial correctness:

### 1. Integer Cents vs. Floating Point
All monetary columns (`balance_cents`, `amount_cents`) use `BIGINT`.
* **The Problem:** In IEEE 754 floating-point arithmetic (used by `FLOAT` and `DOUBLE`), binary fractions cannot accurately represent base-10 decimals (e.g. `0.1 + 0.2 = 0.30000000000000004`). In a system processing 3,500 transactions per second, fractional penny drift compounds into thousands of dollars of unreconciled discrepancy.
* **The Solution:** Every dollar is represented as 100 integer cents ($50.25 is stored as `5025`).

### 2. Double-Entry Bookkeeping
Money cannot simply be created or destroyed via an `UPDATE accounts SET balance = balance - 100` query. Every single transaction generates balanced pairs of entries in `journal_entries`:
$$\sum \text{Debits} = \sum \text{Credits}$$
Every movement of money is permanently reconstructable and auditable.

### 3. Optimistic Concurrency Control (OCC)
The `version INT NOT NULL DEFAULT 0` column on `accounts` prevents lost updates during high-concurrency race conditions:
```sql
UPDATE accounts 
SET balance_cents = balance_cents - :amount, version = version + 1 
WHERE id = :id AND version = :expected_version;
```
If another transaction commits first, the version check fails atomically without corrupting balances.

### 4. Transactional Outbox Pattern
The `outbox` table solves the dual-write failure mode where an application commits a database record but crashes before publishing to Kafka. Both operations commit atomically within the same PostgreSQL ACID transaction.

### 5. Semi-Structured JSONB for Evolving ML Features (Day 5)
The `feature_snapshots.feature_data` column uses binary JSON (`JSONB`). This allows the machine learning team to add new fraud signals (e.g. IP velocity, geo-hop speed, device trust score) dynamically across Weeks 7-8 without executing disruptive DDL table alters.

### 6. Left-to-Right Composite Index Ordering (Day 5)
The composite index `idx_feature_snapshots_account_time` is ordered as `(account_id, computed_at DESC)`:
* **The Rule:** A composite index on `(A, B)` speeds up queries on `A` alone or `A AND B`, but is useless for queries filtering on `B` alone.
* **Access Pattern:** Queries always fetch a single account's feature history over time (`WHERE account_id = ? ORDER BY computed_at DESC`), making `account_id` the essential leading column.

---

## 🗃️ Migration Files Index

| File | Migration Name | Applied In | Core Entities Created | Status |
| :--- | :--- | :---: | :--- | :---: |
| **[`001_initial_ledger_schema.sql`](./001_initial_ledger_schema.sql)** | Initial Ledger & Outbox Schema | **Day 1** | `accounts`, `transactions`, `journal_entries`, `outbox` | ✅ Active |
| **[`002_fraud_and_features.sql`](./002_fraud_and_features.sql)** | Human Reviews & Feature Snapshots | **Day 5** | `fraud_reviews`, `feature_snapshots` | ✅ Active |
| **[`003_idempotency_keys.sql`](./003_idempotency_keys.sql)** | Idempotent Request Deduplication & Caching | **Day 10** | `idempotency_keys` | ✅ Active |
| **[`verify_day05_schema.sql`](./verify_day05_schema.sql)** | Day 5 Verification & Diagnostics | **Day 5** | Deterministic Seed Data, Constraint & Index Tests | 🧪 Test Suite |
| **[`verify_day06_settlement.sql`](./verify_day06_settlement.sql)** | Day 6 Settlement & Atomicity Proof | **Day 6** | Balanced Double-Entry Settlement & Rollback Trap | 🧪 Test Suite |
| **[`verify_day09_occ.sql`](./verify_day09_occ.sql)** | Day 9 Optimistic Concurrency Proof | **Day 9** | Monotonic Version Increment & Conflicting Write Rejection | 🧪 Test Suite |
| **[`verify_day10_idempotency.sql`](./verify_day10_idempotency.sql)** | Day 10 Idempotency & Replay Proof | **Day 10** | Primary Key Deduplication, Error 23505 & Replay Invariant | 🧪 Test Suite |

---

## 🚀 Applying Migrations to an Existing Database Container

Because PostgreSQL only executes scripts in `/docker-entrypoint-initdb.d/` on the very first container initialization, subsequent migrations on an existing volume must be applied manually:

```bash
# 1. Copy the migration script into the running container
docker cp migrations/002_fraud_and_features.sql novaris-postgres:/tmp/002_fraud_and_features.sql

# 2. Execute migration with psql
docker exec -it novaris-postgres psql -U novaris_admin -d novaris_db -f /tmp/002_fraud_and_features.sql

# 3. Execute Day 5 deterministic schema verification harness
docker cp migrations/verify_day05_schema.sql novaris-postgres:/tmp/verify_day05_schema.sql
docker exec -it novaris-postgres psql -U novaris_admin -d novaris_db -f /tmp/verify_day05_schema.sql

# 4. Execute Day 6 deterministic atomic settlement & rollback proof harness
docker cp migrations/verify_day06_settlement.sql novaris-postgres:/tmp/verify_day06_settlement.sql
docker exec -it novaris-postgres psql -U novaris_admin -d novaris_db -f /tmp/verify_day06_settlement.sql

# 5. Execute Day 9 deterministic Optimistic Concurrency Control (OCC) harness
docker cp migrations/verify_day09_occ.sql novaris-postgres:/tmp/verify_day09_occ.sql
docker exec -it novaris-postgres psql -U novaris_admin -d novaris_db -f /tmp/verify_day09_occ.sql
```

---

<div align="center">

<p align="center">
  <img src="../assets/footer.svg" alt="NOVARIS Footer" width="100%" />
</p>

</div>
