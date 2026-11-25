<div align="center">

<!-- Top Banner -->
<p align="center">
  <img src="../../assets/banner.svg" alt="NOVARIS Banner" width="100%" />
</p>

# 🏛️ NOVARIS Architecture Decision Records (ADRs)
### Systematic Historical Documentation of Key Architectural Decisions, Invariants, and Trade-Offs

<p align="center">
  <img src="https://img.shields.io/badge/Documentation-ADR_Standard-00E5FF?style=flat-square&logo=gitbook&logoColor=white" alt="Standard" />
  <img src="https://img.shields.io/badge/Format-Michael_Nygard_RFC-00C853?style=flat-square&logo=markdown&logoColor=white" alt="Format" />
  <img src="https://img.shields.io/badge/Status-Living_Repository-7C4DFF?style=flat-square&logo=target&logoColor=white" alt="Status" />
  <img src="https://img.shields.io/badge/Audit-FinTech_Grade-FFA000?style=flat-square&logo=securityscorecard&logoColor=white" alt="Audit" />
</p>

---

</div>

## 📌 Purpose of Architecture Decision Records (ADRs)

In distributed financial platforms handling millions of dollars in transaction throughput, **architecture drift and tribal knowledge are critical operational liabilities**. When architectural decisions are made verbally, in transient chat threads, or inside pull request comment histories, new engineers and auditors cannot answer the fundamental question:

> *"Why was this system built this way instead of the obvious alternative?"*

An **Architecture Decision Record (ADR)** is a lightweight, immutable document that captures a significant architectural decision along with its context, considered alternatives, and evaluated consequences.

### The ADR Lifecycle:
```
  ┌──────────┐     Accepted     ┌──────────┐     Superceded By ADR-YYYY     ┌────────────┐
  │ Proposed │ ───────────────► │ Accepted │ ─────────────────────────────► │ Superseded │
  └──────────┘                  └──────────┘                                └────────────┘
       │                              │
       │ Rejected                     │ Deprecated
       ▼                              ▼
  ┌──────────┐                  ┌────────────┐
  │ Rejected │                  │ Deprecated │
  └──────────┘                  └────────────┘
```

---

## 🗺️ Master Index of Architecture Decisions

| ADR Number | Decision Title | Decision Date | Status | Impact Area |
| :---: | :--- | :---: | :---: | :--- |
| **[ADR-0001](./0001-optimistic-concurrency-control-for-payment-settlement.md)** | **Optimistic Concurrency Control (OCC) vs. Pessimistic Row Locking for Ledger Settlement** | 2026-11-25 | ✅ **Accepted** | Database Ledger, Concurrency, Settlement Core |
| **[ADR-0002](./0002-five-layer-defense-in-depth-gateway-pipeline.md)** | **Five-Layer Defense-in-Depth Middleware Pipeline Ordering and Ingress Gateways** | 2026-11-25 | ✅ **Accepted** | API Gateway, Security, Rate Limiting, Idempotency |

---

## 📋 ADR Authoring Standards (Michael Nygard Pattern)

Every ADR in NOVARIS adheres strictly to the following structural sections:
1. **Title:** Sequential number and concise imperative verb phrase (`0001-optimistic-concurrency-control...`).
2. **Status:** `Proposed`, `Accepted`, `Rejected`, `Deprecated`, or `Superseded by ADR-XXXX`.
3. **Context:** The technical, financial, or operational problem forcing the decision.
4. **Decision:** The chosen architecture or technology and the explicit commitment.
5. **Alternatives Considered:** What other designs were evaluated and why they were rejected.
6. **Consequences:** Positive benefits, negative trade-offs, and mitigation strategies.
7. **Compliance & Invariants:** Specific mathematical, cryptographic, or relational proofs guaranteed by the decision.

---

<div align="center">

<!-- Bottom Banner -->
<p align="center">
  <img src="../../assets/footer.svg" alt="NOVARIS Footer" width="100%" />
</p>

<p align="center">
  <strong>NOVARIS Architecture Vault • Maintained by <a href="https://github.com/charan21042005">Tokachichu Sri Charan</a></strong>
</p>

</div>
