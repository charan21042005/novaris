<div align="center">

<!-- Top Banner -->
<p align="center">
  <img src="../../../../../assets/banner.svg" alt="NOVARIS Banner" width="100%" />
</p>

# 📬 NOVARIS Outbox & Crash Recovery Diagnostic CLI (`verifyoutbox`)
### 5-Stage Automated Forensic Audit Proving Dual-Write Elimination & Zero-Loss Crash Recovery

<p align="center">
  <img src="https://img.shields.io/badge/Milestone-Phase_2_Day_18_COMPLETED_(100%25_Certified)-00C853?style=flat-square&logo=checkbox&logoColor=white" alt="Milestone" />
  <img src="https://img.shields.io/badge/Diagnostic-5--Stage_Pipeline-7C4DFF?style=flat-square&logo=visualstudiocode&logoColor=white" alt="Diagnostic" />
  <img src="https://img.shields.io/badge/Fault_Injection-Crash_Recovery-FF6D00?style=flat-square&logo=lightning&logoColor=white" alt="Fault Injection" />
  <img src="https://img.shields.io/badge/Delivery-At--Least--Once-00E5FF?style=flat-square&logo=apachekafka&logoColor=white" alt="Delivery" />
  <img src="https://img.shields.io/badge/Outcome-No_Dropped_or_Duplicated_Events-00C853?style=flat-square&logo=target&logoColor=white" alt="Outcome" />
</p>

---

</div>

## 📌 Tool Overview

The **`verifyoutbox`** command-line utility provides deterministic proof that the **Transactional Outbox pattern** (Day 18) eliminates the distributed dual-write dilemma and guarantees that **no event is ever dropped or duplicated across simulated worker crashes**.

```mermaid
flowchart TD
    S1["Stage 1: Infrastructure Connectivity Probe<br/>(PostgreSQL + Kafka TCP Ping)"]
    S2["Stage 2: Atomic Dual-Write Enlistment<br/>(Proves rollback on failure, commits PENDING batch)"]
    S3["Stage 3: Outbox Relay Processing Sweep<br/>(Polls PENDING rows, publishes to Kafka with Murmur2 Key)"]
    S4["Stage 4: Crash Simulation Mid-Flight<br/>(SIGKILL worker post-publish, assert PENDING on reboot)"]
    S5["Stage 5: Idempotency & Zero-Loss Audit<br/>(Proves Dropped Events = 0, Downstream Deduplication)"]

    S1 --> S2 --> S3 --> S4 --> S5
```

---

## 🚀 Execution & Command-Line Flags

```bash
# Run in Hermetic Simulation Mode (zero external dependencies required)
go run cmd/verifyoutbox/main.go -offline

# Run against Live Docker Infrastructure (Postgres 16 + Kafka KRaft)
go run cmd/verifyoutbox/main.go -db="postgres://novaris_admin:novaris_secure_password_2026@localhost:5432/novaris_db?sslmode=disable" -brokers="localhost:9092" -count=5
```

### Supported Flags:
| Flag | Default | Description |
| :--- | :--- | :--- |
| `-db` | `postgres://novaris_admin:...@localhost:5432/novaris_db` | PostgreSQL connection string |
| `-brokers` | `localhost:9092` | Kafka broker host:port addresses |
| `-topic` | `novaris.tx.initiated` | Target Kafka topic for event relaying |
| `-count` | `5` | Number of test transactions in the batch |
| `-offline` | `false` | Force in-memory simulation mode |

---

<div align="center">

<!-- Bottom Banner -->
<p align="center">
  <img src="../../../../../assets/footer.svg" alt="NOVARIS Footer" width="100%" />
</p>

### 🎓 NOVARIS Diagnostic Tooling
*Day 18 Transactional Outbox & Crash Recovery Verifier*

</div>
