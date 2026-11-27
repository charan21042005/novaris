<div align="center">

<!-- Top Banner -->
<p align="center">
  <img src="../assets/banner.svg" alt="NOVARIS Banner" width="100%" />
</p>

# ⚙️ NOVARIS • Distributed Microservices Mesh (`/services`)
### Polyglot Service Architecture for High-Throughput Payment Ingestion, Settlement & Fraud Detection

<p align="center">
  <img src="https://img.shields.io/badge/Phase_1_Capstone-COMPLETED_(100%25_Verified)-00C853?style=flat-square&logo=checkbox&logoColor=white" alt="Phase 1 Capstone" />
  <img src="https://img.shields.io/badge/Architecture-Distributed_Microservices-00ADD8?style=flat-square&logo=microservices&logoColor=white" alt="Microservices" />
  <img src="https://img.shields.io/badge/Languages-Go_1.22+_•_Python_3.11+-00C853?style=flat-square&logo=polyglot&logoColor=white" alt="Polyglot" />
  <img src="https://img.shields.io/badge/Throughput_Target-3%2C500+_TPS-FF6D00?style=flat-square&logo=speedtest&logoColor=white" alt="Throughput" />
  <img src="https://img.shields.io/badge/P99_Latency-<40ms-7C4DFF?style=flat-square&logo=clockify&logoColor=white" alt="Latency" />
</p>

---

</div>

## 📌 Microservice Topology & Domain Decomposition

The **`/services`** directory houses the backend microservices comprising the NOVARIS core processing plane. Designed around domain-driven boundaries and event-driven choreography, NOVARIS separates public ingress, ledger consensus, and artificial intelligence into independently deployable and scalable services:

```mermaid
flowchart TD
    subgraph IngressLayer ["Ingress & Perimeter Protection"]
        Client["Merchant / Client App"] -->|"HTTPS / POST /v1/payments"| Gateway["services/gateway<br/>(Go 1.22+)"]
    end

    subgraph MessagingLayer ["Distributed Event Streaming Mesh"]
        Gateway -->|"Kafka Topic: tx.initiated"| Kafka["Apache Kafka (KRaft Mode)"]
    end

    subgraph ProcessingLayer ["Decoupled Microservice Workers"]
        Kafka -->|"Consumer Group: settlement-workers"| Settlement["services/settlement<br/>(Go 1.22+)"]
        Kafka -->|"Consumer Group: fraud-scorers"| FraudML["services/fraud-ml<br/>(Python 3.11+)"]
    end

    subgraph DataStorageLayer ["Data Persistence & Cache"]
        Settlement -->|"Double-Entry ACID Writes"| Postgres[("PostgreSQL 16 Ledger")]
        Settlement -->|"Distributed Locks"| Redis[("Redis 7 Cache")]
        FraudML -->|"Dynamic JSONB Feature Reads"| Postgres
    end
```

---

## 🗂️ Microservices Directory Map

| Service Directory | Core Runtime | Architectural Responsibilities | Roadmap Phase |
| :--- | :---: | :--- | :---: |
| **[`gateway/`](./gateway/)** | **Go 1.22+** | Public edge ingress, RFC 7519 JWT auth, RFC 2104 HMAC-SHA256 signature verification, Token-Bucket rate limiting, IETF idempotency deduplication (`idempotency_keys`), and health probing. | **Phase 1** *(Active)* |
| **[`settlement/`](./settlement/)** | **Go 1.22+** | High-density double-entry ledger settlement, transactional outbox event relaying, Redlock distributed locking, and multi-partition Kafka consumer groups. | **Phase 2 (Day 17 - Active Verified)** |
| **[`fraud-ml/`](./fraud-ml/)** | **Python 3.11+** | Real-time adaptive machine learning scoring (<15ms SLA), dynamic XGBoost/LightGBM inference, streaming feature extraction, and human-review triage. | **Phase 3** *(Weeks 7–8)* |

---

## 💡 Why a Polyglot Architecture?

In enterprise financial engineering, no single programming language is optimal for all tasks:

1. **Go for Gateway & Settlement:**
   * **Ultra-Low Latency & High Concurrency:** Go's lightweight goroutines (2KB stack vs. 1MB OS thread) and non-blocking network runtime allow a single gateway pod to handle tens of thousands of concurrent client connections with sub-millisecond overhead.
   * **Predictable Garbage Collection:** Go's sub-millisecond GC pauses prevent latency spikes that could cause payment timeouts.
   * **Type Safety & Binary Portability:** Compiles into single static binaries with zero external runtime dependencies.

2. **Python for Fraud Intelligence:**
   * **World-Class ML Ecosystem:** Native access to NumPy, Scikit-Learn, XGBoost, and PyTorch for trained decision trees and neural networks.
   * **Fast Streaming Frameworks:** Powered by `FastStream` and `asyncio` for asynchronous Kafka consumption and batch vector scoring.

---

<div align="center">

<!-- Bottom Banner -->
<p align="center">
  <img src="../assets/footer.svg" alt="NOVARIS Footer" width="100%" />
</p>

### 🎓 NOVARIS Microservices Mesh
*Production-Grade Distributed System Architecture*

</div>
