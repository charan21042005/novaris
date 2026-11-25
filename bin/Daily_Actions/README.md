<div align="center">

<!-- Top Banner -->
<p align="center">
  <img src="../../assets/banner.svg" alt="NOVARIS Banner" width="100%" />
</p>

# 📅 NOVARIS • Daily Engineering Action Logs
### Comprehensive Day-by-Day Implementation, Explanations & Learning Records Across the 84-Day Master Curriculum

<p align="center">
  <img src="https://img.shields.io/badge/Phase_1_Capstone-COMPLETED_(100%25_Verified)-00C853?style=flat-square&logo=checkbox&logoColor=white" alt="Phase 1 Capstone" />
  <img src="https://img.shields.io/badge/Progress-Day_1_to_84-00E5FF?style=flat-square&logo=gitbook&logoColor=white" alt="Progress" />
  <img src="https://img.shields.io/badge/Target-Zero_to_Production-00C853?style=flat-square&logo=target&logoColor=white" alt="Target" />
  <img src="https://img.shields.io/badge/Architecture-Modular_Phase_Vaults-7C4DFF?style=flat-square&logo=visualstudiocode&logoColor=white" alt="Architecture" />
  <img src="https://img.shields.io/badge/Documentation-Textbook_Standard-FFA000?style=flat-square&logo=readthedocs&logoColor=white" alt="Docs" />
</p>

---

</div>

## 🎯 Purpose of This Directory

This directory acts as the **central day-by-day learning journal and technical breakdown** of everything built in NOVARIS. To maintain clean ergonomics across the 12-week, 84-day master build, all daily logs are partitioned into **Phase-specific vaults**:

1. **Plain-English Explanations:** Every architectural concept explained so clearly that even a beginner or non-technical stakeholder can follow.
2. **Line-by-Line Code Walkthroughs:** Why each line of Go, Python, SQL, or Docker config was written and what alternative approaches were rejected.
3. **Local Testing & Verification:** Concrete CLI commands and expected outputs proving the day's milestone was met.
4. **Interview Talking Points:** What questions hiring managers ask about that day's work and how to answer them with confidence.

---

## ⚡ Phase 1 Testing & CLI Verification Runbook

Looking for copy-pasteable commands to run or probe any Phase 1 milestone? Check the centralized cheat sheet:  
👉 **[`Phase_1_Design_and_Ledger/COMMANDS_RUNBOOK.md`](./Phase_1_Design_and_Ledger/COMMANDS_RUNBOOK.md)** — Master copy-paste runbook containing all Docker setup commands, `curl` probes, CLI generators, Postman collections, and PowerShell scripts for Phase 1.

---

## 📂 Phase Vaults Directory Structure

```
bin/Daily_Actions/
├── README.md                                   # Root directory index (this file)
│
├── Phase_1_Design_and_Ledger/                 # 🏛️ Phase 1: Weeks 1-2 (Days 01–14) — [ACTIVE VAULT]
│   ├── README.md                               # Phase 1 master index, architecture overview & badges
│   ├── COMMANDS_RUNBOOK.md                     # Centralized test & run cheat sheet for Phase 1 milestones
│   ├── Day_01_Action_Log.md                    # Day 01: Monorepo, Docker Compose & Ledger Schema Foundation
│   ├── Day_02_Action_Log.md                    # Day 02: JWT Authentication & Middleware Pipeline
│   ├── Day_03_Action_Log.md                    # Day 03: HMAC-SHA256 Request Signing & Replay Defense
│   ├── Day_04_Action_Log.md                    # Day 04: Token-Bucket Rate Limiter & Concurrency Defense
│   ├── Day_05_Action_Log.md                    # Day 05: Ledger Schema, Human Fraud Reviews & Composite Indexes
│   ├── Day_06_Action_Log.md                    # Day 06: Atomic Settlement Logic, Double-Entry Writes & Rollbacks
│   ├── Day_07_Action_Log.md                    # Day 07: Testing Discipline, Table-Driven Tests & Mock Isolation
│   ├── Day_08_Action_Log.md                    # Day 08: Concurrency Stress Testing & Race Detection
│   ├── Day_09_Action_Log.md                    # Day 09: Optimistic Concurrency Control (OCC) & Version Columns
│   ├── Day_10_Action_Log.md                    # Day 10: Idempotent API Design, Request Hashing & Cached Replays
│   ├── Day_11_Action_Log.md                    # Day 11: Integration Wiring, End-to-End Contracts & Postman Suite
│   ├── Day_12_Action_Log.md                    # Day 12: Documentation Discipline: OpenAPI 3.0 & ADR Repository
│   ├── Day_13_Action_Log.md                    # Day 13: Concurrency Load Scripting (k6 & Hot-Wallet Contention)
│   └── Day_14_Action_Log.md                    # Day 14: Phase 1 Capstone Review & Final Sign-Off (COMPLETED)
│
├── Phase_2_Streaming_Core/                     # 🚀 Phase 2: Weeks 3-4 (Days 15–28) — [UPCOMING]
├── Phase_3_Fraud_Engine/                       # 🧠 Phase 3: Weeks 5-6 (Days 29–42) — [UPCOMING]
├── Phase_4_High_Throughput/                    # ⚡ Phase 4: Weeks 7-8 (Days 43–56) — [UPCOMING]
├── Phase_5_Observability/                      # 📊 Phase 5: Weeks 9-10 (Days 57–70) — [UPCOMING]
└── Phase_6_Capstone/                           # 🎓 Phase 6: Weeks 11-12 (Days 71–84) — [UPCOMING]
```

---

## 🗺️ Master Daily Logs Index

| Phase | Day | Focus / Milestone | Document Link | Status |
| :---: | :---: | :--- | :--- | :---: |
| **Phase 1** | **Day 01** | **Monorepo Layout, Docker Compose & Ledger Schema Foundation** | [Day_01_Action_Log.md](./Phase_1_Design_and_Ledger/Day_01_Action_Log.md) | ✅ *COMPLETED* |
| **Phase 1** | **Day 02** | **JWT Authentication & Security Middleware Pipeline** | [Day_02_Action_Log.md](./Phase_1_Design_and_Ledger/Day_02_Action_Log.md) | ✅ *COMPLETED* |
| **Phase 1** | **Day 03** | **Cryptographic HMAC-SHA256 Request Signing & Replay Defense** | [Day_03_Action_Log.md](./Phase_1_Design_and_Ledger/Day_03_Action_Log.md) | ✅ *COMPLETED* |
| **Phase 1** | **Day 04** | **Token-Bucket Rate Limiting, Concurrency Safety & Fail-Fast Gateway** | [Day_04_Action_Log.md](./Phase_1_Design_and_Ledger/Day_04_Action_Log.md) | ✅ *COMPLETED* |
| **Phase 1** | **Day 05** | **Full Ledger Schema: Human Fraud Reviews, Feature Snapshots & Composite JSONB Indexing** | [Day_05_Action_Log.md](./Phase_1_Design_and_Ledger/Day_05_Action_Log.md) | ✅ *COMPLETED* |
| **Phase 1** | **Day 06** | **Atomic Settlement Logic: Double-Entry Ledger Writes, Go Transactions & Rollbacks** | [Day_06_Action_Log.md](./Phase_1_Design_and_Ledger/Day_06_Action_Log.md) | ✅ *COMPLETED* |
| **Phase 1** | **Day 07** | **Testing Discipline: Table-Driven Testing Safety Nets, Mock Isolation & Benchmarking** | [Day_07_Action_Log.md](./Phase_1_Design_and_Ledger/Day_07_Action_Log.md) | ✅ *COMPLETED* |
| **Phase 1** | **Day 08** | **Concurrency Testing: Goroutines, Start-Gate Synchronization & The Race Detector** | [Day_08_Action_Log.md](./Phase_1_Design_and_Ledger/Day_08_Action_Log.md) | ✅ *COMPLETED* |
| **Phase 1** | **Day 09** | **Optimistic Concurrency Control: Version Columns & Conditional Database Updates** | [Day_09_Action_Log.md](./Phase_1_Design_and_Ledger/Day_09_Action_Log.md) | ✅ *COMPLETED* |
| **Phase 1** | **Day 10** | **Idempotent API Design: Network Retry Deduplication & Response Caching** | [Day_10_Action_Log.md](./Phase_1_Design_and_Ledger/Day_10_Action_Log.md) | ✅ *COMPLETED* |
| **Phase 1** | **Day 11** | **Integration Wiring & End-to-End API Contracts (Postman & Full-Stack Gateway Pipeline)** | [Day_11_Action_Log.md](./Phase_1_Design_and_Ledger/Day_11_Action_Log.md) | ✅ *COMPLETED* |
| **Phase 1** | **Day 12** | **Documentation Discipline: OpenAPI Specifications & Architecture Decision Records (ADR-001)** | [Day_12_Action_Log.md](./Phase_1_Design_and_Ledger/Day_12_Action_Log.md) | ✅ *COMPLETED* |
| **Phase 1** | **Day 13** | **Concurrency Load Scripting (k6 & Hot-Wallet Contention)** | [Day_13_Action_Log.md](./Phase_1_Design_and_Ledger/Day_13_Action_Log.md) | ✅ *COMPLETED* |
| **Phase 1** | **Day 14** | **Phase 1 Capstone Milestone Review & Comprehensive Production Verification** | [Day_14_Action_Log.md](./Phase_1_Design_and_Ledger/Day_14_Action_Log.md) | ✅ *COMPLETED* |
| ... | ... | *(Continuously updated across the 12-week roadmap)* | ... | ... |

---

<div align="center">

<p align="center">
  <img src="../../assets/footer.svg" alt="NOVARIS Footer" width="100%" />
</p>

</div>
