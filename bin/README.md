<div align="center">

<!-- Top Banner -->
<p align="center">
  <img src="../assets/banner.svg" alt="NOVARIS Banner" width="100%" />
</p>

# 📚 NOVARIS • Engineering Reference & Blueprint Vault (`/bin`)
### Central Documentation Archive, Architecture Blueprints, and Daily Action Logs

<p align="center">
  <img src="https://img.shields.io/badge/Vault_Status-Active_Reference-00E5FF?style=flat-square&logo=gitbook&logoColor=white" alt="Status" />
  <img src="https://img.shields.io/badge/Scope-End--to--End_Architecture-00C853?style=flat-square&logo=blueprint&logoColor=white" alt="Scope" />
  <img src="https://img.shields.io/badge/Roadmap-84_Days_Mapped-7C4DFF?style=flat-square&logo=clockify&logoColor=white" alt="Roadmap" />
  <img src="https://img.shields.io/badge/Standard-Production_Grade_SRS-FFA000?style=flat-square&logo=standards&logoColor=white" alt="Standard" />
</p>

---

</div>

## 📌 About This Directory

The **`/bin`** folder serves as the **internal engineering knowledge base and reference repository** for NOVARIS. It centralizes all original architectural proposals, industry-grade Software Requirements Specifications (SRS), technical dependency setup instructions, and day-by-day implementation guides.

> [!NOTE]
> This folder functions as an active development companion while constructing each microservice, data pipeline, and infrastructure mesh. Upon full project delivery, production hardening, and deployment, this temporary reference vault can be cleanly archived.

---

## 🗂️ Directory Taxonomy & Reference Map

```
novaris/bin/
├── 1.Thoughts/             # Initial Capstone proposals, research, & domain architectures
├── 2.Dependencies/         # Prerequisites, toolchain setup, & local dev runtimes
├── 3.Planning/             # Production SRS blueprint, 12-week roadmap, & schema drafts
├── 4.Action_Day_Wise/      # Detailed daily implementation deep-dives & kickoff guides
├── 5.Explanations/         # Architectural summary presentations and system overviews
├── Daily_Actions/          # Live engineering logs & beginner-friendly daily explanations
└── PROJECTS.txt            # Real-time fraud detection problem statements & talking points
```

---

### Detailed Breakdown of Subdirectories

### 1. `1.Thoughts/` — Architectural Foundations & Ideation
Contains the foundational capstone engineering proposals and initial drafts that evolved into NOVARIS:
* **`1.ApexPay_Full_Capstone_Engineering_Proposal.pdf`**: Original problem formulation exploring high-throughput payment settlement.
* **`2.Fraud_Detection_Engine_Capstone_Blueprint.docx`**: Initial streaming fraud detection concepts and feature ingestion patterns.
* **`3.FinalApexPay_Final_Unified_Master_Capstone_Proposal.pdf`**: The synthesized proposal uniting payments with ML fraud scoring.
* **`4.Novaris_SRS_Capstone_Blueprint.docx`**: The core SRS draft establishing the 3,500+ TPS and sub-40ms P99 latency SLA benchmarks.

---

### 2. `2.Dependencies/` — Developer Toolchain & Local Environment
Contains instructions on configuring an isolated local development workstation without cloud dependencies:
* **`1.Dep.pdf`**: Visual dependency matrix and system layer map.
* **`2.Novaris_Setup_and_Prerequisites_Guide.docx`**: Step-by-step setup walkthrough for Go 1.22+, Python 3.11+, Docker Desktop, Minikube, PostgreSQL 16, Redis 7, and Apache Kafka in KRaft mode.

---

### 3. `3.Planning/` — Production Specifications & Master Roadmaps
The core architectural benchmark and implementation schedule guiding the entire project:
* **`Plan.txt`**: High-level problem statement, CQRS architecture, distributed locking strategy, and interview talking points.
* **`final/Novaris_SRS_Capstone_Blueprint.docx`**: The authoritative, publication-ready Software Requirements Specification detailing functional requirements (FR-1 to FR-10), double-entry accounting schemas, and resilience patterns.
* **`final/Novaris_84Day_Roadmap_Overview.docx`**: Phase-by-phase timeline breaking down 84 days of build sessions across 6 core phases.
* **`final/Novaris_84Day_Roadmap_Elaborated.docx`**: Elaborated daily cards including learning goals, concrete labs, verification outcomes, and spaced-recall concepts.

---

### 4. `4.Action_Day_Wise/` — Step-by-Step Implementation Guides
Curated day-by-day kickoff guides authored to guide hands-on engineering from zero to production:
* **`1.Overall_Roadmap/`**: Master sequence tracking dependencies between gateway, ledger, streaming, and ML services.
* **`Day_Wise/`**: Deep-dive build guides covering:
  * `Day 01`: Environment setup, monorepo structure, initial ledger migration, and Go `/healthz` endpoint.
  * `Day 02`: JWT authentication middleware and claims verification.
  * `Day 03`: HMAC-SHA256 digital signature validation for anti-tamper security.
  * `Day 04`: Per-merchant sliding token-bucket distributed rate limiting with Redis.
  * `Day 05`: Full PostgreSQL double-entry ledger schema, composite indexes, and JSONB audit structures.
  * `Day 06`: Go database transaction settlement logic, atomic multi-table writes, and rollback guards.
  * `Day 07`: Table-driven unit testing safety nets, mock isolation, and deterministic test suites.
  * `Day 08`: Concurrency stress testing, goroutine start-gate synchronization, and race detection.
  * `Day 09`: Optimistic Concurrency Control (OCC), monotonic version columns, check-and-increment atomic writes.
  * `Day 10`: Idempotent API design, IETF-compliant header validation, payload fingerprinting, and response caching.

---

### 5. `5.Explanations/` — Conceptual Overviews & Diagrams
* **`1.Novaris.pdf`**: Architectural visual slides and end-to-end transaction journey diagrams suitable for technical presentations and defense reviews.

---

### 6. `Daily_Actions/` — Live Build Journals & Beginner Guides
* **Active Directory:** Dedicated to documenting the actual code written each day during the active build.
* **Format:** Every single day receives its own markdown log (e.g. `Day_01_Action_Log.md`) featuring beginner-friendly explanations, line-by-line code breakdowns, local test commands, and interview defense questions.
* See [Daily_Actions/README.md](./Daily_Actions/README.md) for the live index.

---

<div align="center">

<!-- Bottom Banner -->
<p align="center">
  <img src="../assets/footer.svg" alt="NOVARIS Footer" width="100%" />
</p>

<p align="center">
  <strong>NOVARIS Reference Vault • Maintained by <a href="https://github.com/charan21042005">Tokachichu Sri Charan</a></strong>
</p>

</div>
