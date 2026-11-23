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

---

## 🗃️ Migration Files Index

| File | Migration Name | Applied In | Core Entities Created |
| :--- | :--- | :---: | :--- |
| **`001_initial_ledger_schema.sql`** | Initial Ledger & Outbox Schema | **Day 1** | `accounts`, `transactions`, `journal_entries`, `outbox` |
| **`002_fraud_review_tables.sql`** | Fraud Reviews & Feature Snapshots | *Day 5* | `fraud_reviews`, `feature_snapshots` |

---

<div align="center">

<p align="center">
  <img src="../assets/footer.svg" alt="NOVARIS Footer" width="100%" />
</p>

</div>
