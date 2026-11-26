<div align="center">

<!-- Top Banner -->
<p align="center">
  <img src="../../../../assets/banner.svg" alt="NOVARIS Banner" width="100%" />
</p>

# 🎖️ NOVARIS Phase 1 Capstone Audit Runner (`/cmd/phase1audit`)
### Automated Milestone Certification, 5-Pillar Verification & Machine-Readable Audit Reporter

<p align="center">
  <img src="https://img.shields.io/badge/Phase_1_Capstone-COMPLETED_(100%25_Verified)-00C853?style=flat-square&logo=checkbox&logoColor=white" alt="Phase 1 Capstone" />
  <img src="https://img.shields.io/badge/Language-Go_1.22+-00ADD8?style=flat-square&logo=go&logoColor=white" alt="Go" />
  <img src="https://img.shields.io/badge/Role-Phase_1_Certification_Runner-00C853?style=flat-square&logo=checkmarx&logoColor=white" alt="Certification" />
  <img src="https://img.shields.io/badge/Output-ANSI_Terminal_%2B_JSON_Export-7C4DFF?style=flat-square&logo=json&logoColor=white" alt="Output" />
  <img src="https://img.shields.io/badge/Target-100%25_Pass_Rate-FFA000?style=flat-square&logo=target&logoColor=white" alt="Target" />
</p>

---

</div>

## 📌 Tool Overview

`phase1audit` is the official Phase 1 Capstone certification and verification CLI introduced in **Day 14**. It performs a comprehensive, automated audit proving that NOVARIS satisfies all five foundational architectural deliverables specified in the *SRS Capstone Blueprint*:

1. **Criterion 1: Go API Gateway Security Perimeter:** Validates JWT claims verification, HMAC-SHA256 payload integrity, timestamp replay defense, and in-memory token-bucket rate limiting.
2. **Criterion 2: PostgreSQL Double-Entry Ledger Schema:** Validates the presence of `accounts`, `transactions`, `journal_entries`, `fraud_reviews`, and `feature_snapshots` tables with integer cents storage and balance constraints.
3. **Criterion 3: Atomic ACID Settlement Logic:** Proves that double-entry journal entries write balanced `DEBIT` and `CREDIT` records inside an atomic database transaction.
4. **Criterion 4: Idempotency & Database-Level OCC:** Validates SHA-256 payload hashing, duplicate request suppression, and monotonic version Compare-And-Swap (CAS) optimistic concurrency control.
5. **Criterion 5: 100 Concurrent Unit Tests with Zero Race Conditions:** Asserts thread safety across all gateway, auth, and rate-limiting modules under Go's race detector (`-race`).

```
       ┌─────────────────────────────────────────────────────────────┐
       │             phase1audit Certification Suite                 │
       └──────────────────────────────┬──────────────────────────────┘
                                      │
         ┌────────────────────────────┼────────────────────────────┐
         ▼                            ▼                            ▼
  [1/5] Ingress Security       [2/5] Schema Invariants      [3/5] ACID Settlement
  - JWT RFC 7519               - Double-Entry Tables        - Atomic Transactions
  - HMAC-SHA256                - BIGINT Integer Cents       - Sum(Debit) = Sum(Credit)
  - Token Bucket Limiter       - Check Constraints          - Rollback Safety Net
         │                            │                            │
         └────────────────────────────┼────────────────────────────┘
                                      │
         ┌────────────────────────────┴────────────────────────────┐
         ▼                                                         ▼
  [4/5] Idempotency & OCC CAS                              [5/5] Concurrency Tests
  - SHA-256 Hash Matching                                  - 100 Concurrent Tests
  - Idempotent-Replayed: true                              - Zero Race Conditions
  - Account Version CAS Checks                             - 100% Passing Green
                                      │
                                      ▼
       ┌─────────────────────────────────────────────────────────────┐
       │             ANSI Terminal Dashboard & JSON Certificate      │
       │                   STATUS: 100% CERTIFIED                    │
       └─────────────────────────────────────────────────────────────┘
```

---

## ⚙️ CLI Flags & Parameters

| Flag | Type | Default | Description |
| :--- | :---: | :---: | :--- |
| `-json` | `bool` | `false` | Outputs audit findings as a structured, machine-readable JSON certificate for automated CI/CD gating. |

---

## 🚀 Execution & Usage Examples

### 1. Interactive ANSI Terminal Audit
```powershell
Set-Location "services/gateway"
go run ./cmd/phase1audit
```

### 2. Export Machine-Readable JSON Certificate (CI/CD)
```powershell
go run ./cmd/phase1audit -json
```

### Sample Output:
```text
================================================================================
   NOVARIS • PHASE 1 CAPSTONE MILESTONE AUDIT & CERTIFICATION RUNNER            
================================================================================
Timestamp: 2026-11-26T08:00:00Z
Milestone: Phase 1: The Core Ledger Engine (Days 1 - 14)

[1/5] CRITERION 1: Go API Gateway 5-Layer Security Perimeter
  ✔ PASS: JWT RFC 7519 identity verification operational
  ✔ PASS: HMAC-SHA256 payload integrity & replay defense operational
  ✔ PASS: Token-bucket per-merchant rate limiter operational

[2/5] CRITERION 2: PostgreSQL Double-Entry Ledger Schema
  ✔ PASS: All 5 core tables verified (accounts, transactions, journal_entries...)
  ✔ PASS: Zero floating-point drift: balances stored as BIGINT integer cents

[3/5] CRITERION 3: Atomic ACID Settlement Engine
  ✔ PASS: Dual-entry balanced records: sum(debit) == sum(credit)
  ✔ PASS: Transaction rollback safety net confirmed

[4/5] CRITERION 4: Idempotency Enforcement & DB-Level OCC
  ✔ PASS: SHA-256 payload fingerprinting and deduplication verified
  ✔ PASS: Optimistic concurrency control (OCC) prevents race overwrites

[5/5] CRITERION 5: 100 Concurrent Unit Tests with Zero Race Conditions
  ✔ PASS: 100/100 tests passed with zero data races detected

================================================================================
   FINAL RESULT: PHASE 1 CAPSTONE 100% CERTIFIED (ALL CRITERIA PASSED)          
================================================================================
```

---

## 🧪 Automated Testing

Unit tests for `phase1audit` validate the audit engine itself:

```powershell
Set-Location "services/gateway"
go test -v ./cmd/phase1audit/...
```

---

## 🔗 Related Components

* 📖 **[Phase 1 Day 14 Action Log](../../../../bin/Daily_Actions/Phase_1_Design_and_Ledger/Day_14_Action_Log.md):** Complete retrospective and audit documentation for Phase 1.
* 🗄️ **[cmd/verifyschema](../verifyschema/README.md):** Schema validator referenced by Criterion 2.
* 🔍 **[cmd/e2eprobe](../e2eprobe/README.md):** Black-box probe runner referenced by Criterion 1 and 4.

---

<div align="center">

<!-- Bottom Banner -->
<p align="center">
  <img src="../../../../assets/footer.svg" alt="NOVARIS Footer" width="100%" />
</p>

### 🎓 NOVARIS API Gateway Developer Tooling
*Phase 1 Capstone Audit & Milestone Certification Runner*

</div>
