<div align="center">

<!-- Top Banner -->
<p align="center">
  <img src="../../../../assets/banner.svg" alt="NOVARIS Banner" width="100%" />
</p>

# 🗄️ NOVARIS Database Schema Verifier (`/cmd/verifyschema`)
### PostgreSQL Schema Auditing, Static DDL Parsing & Double-Entry Ledger Validation

<p align="center">
  <img src="https://img.shields.io/badge/Phase_1_Capstone-COMPLETED_(100%25_Verified)-00C853?style=flat-square&logo=checkbox&logoColor=white" alt="Phase 1 Capstone" />
  <img src="https://img.shields.io/badge/Language-Go_1.22+-00ADD8?style=flat-square&logo=go&logoColor=white" alt="Go" />
  <img src="https://img.shields.io/badge/Database-PostgreSQL_16-336791?style=flat-square&logo=postgresql&logoColor=white" alt="PostgreSQL" />
  <img src="https://img.shields.io/badge/Mode-Offline_DDL_%2B_Live_DB-00C853?style=flat-square&logo=git&logoColor=white" alt="Modes" />
  <img src="https://img.shields.io/badge/Coverage-7_Core_Ledger_Tables-7C4DFF?style=flat-square&logo=tablecheck&logoColor=white" alt="Coverage" />
</p>

---

</div>

## 📌 Tool Overview

`verifyschema` is a database compliance and migration auditing utility introduced in **Day 5** and expanded in **Day 10**. It validates that either the active PostgreSQL instance or raw offline SQL migration scripts strictly adhere to the NOVARIS Capstone Schema Blueprint.

The verifier executes dual inspection modes:
1. **Offline Static Analysis (`-offline`):** Parses SQL files directly from disk without requiring a running database daemon. Ideal for CI/CD linting, pre-commit hooks, and air-gapped developer environments.
2. **Live Database Reflection:** Connects to PostgreSQL over TCP, querying `information_schema.tables`, `information_schema.columns`, and `pg_indexes` to ensure physical table schemas match expectations.

```
                  VERIFICATION TARGET
                          │
         ┌────────────────┴────────────────┐
         ▼                                 ▼
   Offline Mode (-offline)            Live DB Mode
 ┌─────────────────────────┐    ┌─────────────────────────┐
 │ Static Lexical Parser   │    │ Live PostgreSQL Conn    │
 │ (migrations/*.sql)      │    │ (information_schema)    │
 └───────────┬─────────────┘    └───────────┬─────────────┘
             │                              │
             └──────────────┬───────────────┘
                            ▼
     ┌──────────────────────────────────────────────┐
     │          Table & Constraint Asserter         │
     │                                              │
     │  * accounts (UUID PK, BIGINT cents, version) │
     │  * transactions (State machine: PENDING...)  │
     │  * journal_entries (DEBIT / CREDIT balance)  │
     │  * outbox (Transactional Outbox pattern)     │
     │  * fraud_reviews (DECISION: APPROVE/REJECT)  │
     │  * feature_snapshots (JSONB feature payload) │
     │  * idempotency_keys (SHA-256 hash, response) │
     │  * Composite Indexes & Foreign Keys          │
     └──────────────────────────────────────────────┘
```

---

## ⚙️ CLI Flags & Parameters

| Flag | Type | Default | Description |
| :--- | :---: | :---: | :--- |
| `-offline` | `bool` | `false` | Executes deterministic static analysis over SQL files without requiring database connections. |
| `-sql` | `string` | `""` | Optional explicit path to an individual SQL migration file. |

---

## 🚀 Execution & Usage Examples

### 1. Offline Static Schema Inspection (CI/CD Ready)
```powershell
Set-Location "services/gateway"
go run ./cmd/verifyschema -offline
```

### 2. Live Database Verification (Against PostgreSQL 16)
```powershell
go run ./cmd/verifyschema
```

### Sample Output:
```text
================================================================================
   NOVARIS • DATABASE SCHEMA COMPLIANCE VERIFIER                                
================================================================================
Mode: Offline Static DDL Analysis
Target Migrations: migrations/*.sql

[1/7] Table: accounts
  ✔ Found: id (UUID PK), balance_cents (BIGINT), currency (VARCHAR(3)), version (BIGINT)
  ✔ Check Constraint: balance_cents >= 0

[2/7] Table: transactions
  ✔ Found: id (UUID PK), payer_id, payee_id, amount_cents, status, created_at

[3/7] Table: journal_entries
  ✔ Found: id, transaction_id, account_id, entry_type (DEBIT/CREDIT), amount_cents
  ✔ Foreign Key: transaction_id -> transactions(id)

[4/7] Table: outbox
  ✔ Found: id, event_type, payload (JSONB), status, published_at

[5/7] Table: fraud_reviews
  ✔ Found: id, transaction_id, risk_score, decision, reviewed_by

[6/7] Table: feature_snapshots
  ✔ Found: account_id, window_start, features (JSONB)
  ✔ Composite Index: idx_feature_snapshots_account_time

[7/7] Table: idempotency_keys
  ✔ Found: key_hash (VARCHAR(64) PK), response_code, response_body, expires_at
  ✔ TTL Index: idx_idemp_expires_at

================================================================================
   OUTCOME: SCHEMA VERIFIED (100% compliant with Phase 1 Double-Entry Blueprint)
================================================================================
```

---

## 🧪 Automated Testing

Unit tests for `verifyschema` validate schema verification routines:

```powershell
Set-Location "services/gateway"
go test -v ./cmd/verifyschema/...
```

---

## 🔗 Related Components

* 📜 **[migrations/](../../../../migrations/README.md):** Directory containing raw SQL DDL files audited by this utility.
* 🏛️ **[Phase 1 Capstone](../../../../bin/Daily_Actions/Phase_1_Design_and_Ledger/README.md):** Double-entry ledger architecture blueprints.
* 🔍 **[cmd/phase1audit](../phase1audit/README.md):** Capstone audit tool that incorporates schema validation.

---

<div align="center">

<!-- Bottom Banner -->
<p align="center">
  <img src="../../../../assets/footer.svg" alt="NOVARIS Footer" width="100%" />
</p>

### 🎓 NOVARIS API Gateway Developer Tooling
*Database Schema Compliance & Double-Entry Ledger Verification*

</div>
