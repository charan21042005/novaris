<div align="center">

<!-- Top Banner -->
<p align="center">
  <img src="../../assets/banner.svg" alt="NOVARIS Banner" width="100%" />
</p>

# 💳 NOVARIS • Distributed Settlement & Ledger Engine (`/services/settlement`)
### High-Throughput Double-Entry Accounting, Outbox Relay & SAGA Coordination

<p align="center">
  <img src="https://img.shields.io/badge/Phase_1_Capstone-COMPLETED_(100%25_Verified)-00C853?style=flat-square&logo=checkbox&logoColor=white" alt="Phase 1 Capstone" />
  <img src="https://img.shields.io/badge/Language-Go_1.22+-00ADD8?style=flat-square&logo=go&logoColor=white" alt="Go" />
  <img src="https://img.shields.io/badge/Pattern-Double--Entry_Accounting-00C853?style=flat-square&logo=cashapp&logoColor=white" alt="Accounting" />
  <img src="https://img.shields.io/badge/Concurrency-OCC_Compare--And--Swap-00E5FF?style=flat-square&logo=affinitylock&logoColor=white" alt="OCC" />
  <img src="https://img.shields.io/badge/Coordination-Redlock_Distributed_Mutex-DC382D?style=flat-square&logo=redis&logoColor=white" alt="Redlock" />
  <img src="https://img.shields.io/badge/Roadmap-Phase_2_(Weeks_3--6)-7C4DFF?style=flat-square&logo=target&logoColor=white" alt="Roadmap" />
</p>

---

</div>

## 📌 Service Overview

The **Settlement Engine** is the financial core of NOVARIS. When fully decoupled in Phase 2, this worker service consumes `tx.initiated` payment events from Apache Kafka and executes high-velocity, mathematically proven ledger state changes against PostgreSQL.

### Core Architectural Pillars:
1. **Mathematical Invariance ($\sum \text{Debits} = \sum \text{Credits}$):**
   * Every balance transfer generates a balanced pair of rows in `journal_entries`. Money is never modified by raw destructive updates.
2. **Optimistic Concurrency Control (OCC):**
   * Uses the monotonic `version INT` column on `accounts` to prevent lost updates when concurrent transfers target the same merchant account.
3. **Transactional Outbox Publisher:**
   * Automatically polls the PostgreSQL `outbox` table and publishes guaranteed delivery events (`tx.settled`) to downstream Kafka partitions without dual-write race conditions.
4. **Redlock Distributed Mutex Coordination:**
   * Employs multi-node Redis locking (`SETNX` with leasing) for coordinated cross-datacenter settlement runs.

---

## 🏗️ Architectural Transition (Phase 1 to Phase 2)

```
Phase 1 (Days 1–10, In-Gateway Prototype):
  [Client] ──> [API Gateway (main.go)] ──> [settlement.go + occ.go] ──> [PostgreSQL]

Phase 2 (Weeks 3–6, Decoupled Microservice Mesh):
  [Client] ──> [API Gateway] ──(Kafka: tx.initiated)──> [Settlement Service Worker]
                                                                  │
                                                ┌─────────────────┴─────────────────┐
                                                ▼                                   ▼
                                      [PostgreSQL Ledger]                 [Redis Distributed Lock]
```

> [!NOTE]
> During Phase 1 (Days 6–10), the atomic double-entry settlement logic is prototyped and verified inside [`services/gateway/settlement.go`](../gateway/settlement.go) and [`services/gateway/occ.go`](../gateway/occ.go). In Phase 2, this logic migrates into this dedicated microservice directory to run as independent worker pods under Kubernetes Horizontal Pod Autoscaling (HPA).

---

<div align="center">

<!-- Bottom Banner -->
<p align="center">
  <img src="../../assets/footer.svg" alt="NOVARIS Footer" width="100%" />
</p>

### 🎓 NOVARIS Settlement Service
*Mathematical Zero-Drift Financial Accounting Engine*

</div>
