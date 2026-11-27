<div align="center">

<!-- Top Banner -->
<p align="center">
  <img src="../../../assets/banner.svg" alt="NOVARIS Banner" width="100%" />
</p>

# ⚡ NOVARIS Phase 2 Verification & Testing Runbook
### Unified Copy-Paste Command Cheat Sheet Across Phase 2 Streaming Milestones (Days 15–28)

<p align="center">
  <img src="https://img.shields.io/badge/Runbook-Phase_2_Streaming_Suite-00E5FF?style=flat-square&logo=gitbook&logoColor=white" alt="Runbook" />
  <img src="https://img.shields.io/badge/Days-15_to_28_Active-00C853?style=flat-square&logo=googlecalendar&logoColor=white" alt="Days" />
  <img src="https://img.shields.io/badge/Shell-Bash_%26_PowerShell-7C4DFF?style=flat-square&logo=powershell&logoColor=white" alt="Shell" />
  <img src="https://img.shields.io/badge/Standard-Industrial_FinTech-FFA000?style=flat-square&logo=standards&logoColor=white" alt="Standard" />
</p>

---

</div>

## 📌 Purpose of This Runbook

This document is the **central command cheat sheet** for **Phase 2: Streaming Core & Distributed State** of NOVARIS. Every command required to boot Apache Kafka in KRaft mode, provision and inspect topic partition logs, execute Go verification diagnostics, run automated integration test suites, and audit partition key hashing is cataloged here in copy-pasteable format for both **Bash** (Linux/macOS) and **Windows PowerShell**.

---

## 📑 Table of Contents

1. [Day 15: Apache Kafka Fundamentals, KRaft Cluster & Topic Topology Runbook](#1-day-15-apache-kafka-fundamentals-kraft-cluster--topic-topology-runbook)
   - 1.1 Start Infrastructure Mesh via Docker Compose
   - 1.2 Inspect Kafka Broker Container & KRaft Consensus Status
   - 1.3 Native Kafka CLI Commands (Inside Container)
   - 1.4 Native Go Verification CLI (`cmd/verifykafka`)
   - 1.5 Automated Go Integration & Unit Test Suites
   - 1.6 Low-Level TCP Port & Protocol Health Probes
   - 1.7 Troubleshooting & Diagnostic FAQ
2. [Day 16: Strongly-Typed Go Kafka Producer Runbook (Preview)](#2-day-16-strongly-typed-go-kafka-producer-runbook-preview)
3. [Day 17: Kafka Consumer Groups & Manual Offset Commits Runbook (Preview)](#3-day-17-kafka-consumer-groups--manual-offset-commits-runbook-preview)
4. [Day 18: Transactional Outbox Pattern & CDC Relay Runbook (Preview)](#4-day-18-transactional-outbox-pattern--cdc-relay-runbook-preview)
5. [Day 19: Dead Letter Queues (DLQ) & Staged Retries Runbook (Preview)](#5-day-19-dead-letter-queues-dlq--staged-retries-runbook-preview)
6. [Master Phase 2 Git Multi-Commit & Push Workflow](#6-master-phase-2-git-multi-commit--push-workflow)

---

## 1. Day 15: Apache Kafka Fundamentals, KRaft Cluster & Topic Topology Runbook

### 1.1 Start Infrastructure Mesh via Docker Compose

Before running streaming workloads, ensure the backing services (PostgreSQL 16, Redis 7, Apache Kafka 7.5 KRaft) are active.

#### Bash (Linux / macOS):
```bash
# Navigate to the deployments directory
cd deployments

# Start the container mesh in detached mode
docker compose up -d

# Verify container status (novaris-kafka, novaris-postgres, novaris-redis)
docker compose ps
```

#### Windows PowerShell:
```powershell
# Navigate to the deployments directory
cd deployments

# Start the container mesh in detached mode
docker compose up -d

# Verify container status
docker compose ps
```

---

### 1.2 Inspect Kafka Broker Container & KRaft Consensus Status

Verify that Kafka node ID 1 has successfully booted in dual broker/controller mode without ZooKeeper.

#### Bash (Linux / macOS):
```bash
# View recent Kafka logs to confirm KRaft initialization
docker logs --tail 50 novaris-kafka

# Confirm broker is listening on ports 9092 (host) and 29092 (docker network)
docker exec -it novaris-kafka netstat -tuln | grep -E "9092|29092|29093"
```

#### Windows PowerShell:
```powershell
# View recent Kafka logs to confirm KRaft initialization
docker logs --tail 50 novaris-kafka

# Inspect running containers filtering by novaris
docker ps --filter "name=novaris"
```

---

### 1.3 Native Kafka CLI Commands (Inside Container)

Use Kafka's built-in administration scripts inside `novaris-kafka` to verify topic state and metadata directly.

#### A. List Active Topics:
```bash
docker exec -it novaris-kafka kafka-topics --bootstrap-server localhost:9092 --list
```

#### B. Describe Canonical NOVARIS Topics & Partitions:
```bash
# Describe novaris.tx.initiated (Check partition count, leader broker, ISR)
docker exec -it novaris-kafka kafka-topics --bootstrap-server localhost:9092 --describe --topic novaris.tx.initiated

# Describe Dead Letter Queue
docker exec -it novaris-kafka kafka-topics --bootstrap-server localhost:9092 --describe --topic novaris.tx.dlq
```

#### C. Manually Create a Test Topic:
```bash
docker exec -it novaris-kafka kafka-topics --bootstrap-server localhost:9092 \
  --create --topic test.manual.ping --partitions 3 --replication-factor 1
```

#### D. Delete Test Topic:
```bash
docker exec -it novaris-kafka kafka-topics --bootstrap-server localhost:9092 \
  --delete --topic test.manual.ping
```

---

### 1.4 Native Go Verification CLI (`cmd/verifykafka`)

NOVARIS includes a specialized diagnostic CLI built in pure Go that validates cluster health, queries KRaft metadata, creates all canonical topics idempotently, proves account-partition hashing, and writes a test probe.

#### Bash (Linux / macOS):
```bash
cd services/gateway

# Run full Day 15 audit with ANSI color-coded reporting
go run ./cmd/verifykafka/main.go

# Run audit in machine-readable JSON format (for CI/CD automation)
go run ./cmd/verifykafka/main.go -json
```

#### Windows PowerShell:
```powershell
cd services\gateway

# Run full Day 15 audit with ANSI color-coded reporting
go run .\cmd\verifykafka\main.go

# Run audit in machine-readable JSON format
go run .\cmd\verifykafka\main.go -json
```

#### Custom Broker Flags:
```powershell
# Target a remote or custom cluster
go run .\cmd\verifykafka\main.go -brokers "kafka.internal.novaris.io:9092" -partitions 6 -replication 3
```

---

### 1.5 Automated Go Integration & Unit Test Suites

Execute the comprehensive test suites validating the `streaming` package and end-to-end Kafka functionality.

#### A. Run Streaming Core Package Tests:
```powershell
cd services\gateway
go test -v ./internal/streaming/...
```

#### B. Run Day 15 Master Integration Suite (`kafka_day15_test.go`):
```powershell
cd services\gateway
go test -v -run TestDay15KafkaStreamingCore .
```

#### C. Run All Phase 1 & Phase 2 Gateway Tests:
```powershell
cd services\gateway
go test -v ./...
```

---

### 1.6 Low-Level TCP Port & Protocol Health Probes

Verify network connectivity to the Kafka broker socket before diagnosing higher-level protocol errors.

#### Windows PowerShell:
```powershell
# Test TCP connection to host port 9092
Test-NetConnection -ComputerName localhost -Port 9092
```

#### Bash (Linux / macOS):
```bash
# Test TCP socket reachability using netcat
nc -zv localhost 9092
```

---

### 1.7 Troubleshooting & Diagnostic FAQ

| Issue Encountered | Root Cause | Exact Remediation Command |
| :--- | :--- | :--- |
| `connection refused: localhost:9092` | Docker container `novaris-kafka` is stopped or host port is unmapped. | Run `cd deployments && docker compose up -d kafka` and check `docker ps`. |
| `port 9092 is already allocated` | Another local service (e.g., Redpanda, lingering Kafka, or host service) owns 9092. | In PowerShell: `Get-Process -Id (Get-NetTCPConnection -LocalPort 9092).OwningProcess` then kill conflicting PID. |
| `LEADER_NOT_AVAILABLE` on first write | Kafka creates topic metadata lazily on first publish; leader election takes ~200ms. | Retry request; our `EnsureTopics()` provisions topics upfront to prevent this race condition. |
| `INVALID_REPLICATION_FACTOR` | Requested replication factor exceeds number of active brokers (e.g., requested 3 on single node). | In local development, ensure `ReplicationFactor = 1`. In production, supply 3+ brokers. |
| Docker daemon connection error | Docker Desktop engine service restarted or stopped. | Run `Start-Process "C:\Program Files\Docker\Docker\Docker Desktop.exe"` and wait 15 seconds. |

---

## 2. Day 16: Strongly-Typed Go Kafka Producer & Event Visibility Runbook

*Day 16 introduces the high-performance Go Kafka Producer (`segmentio/kafka-go`) in the API Gateway, emitting `novaris.tx.initiated` events with Leader ACKs and Murmur2 partition routing.*

### 2.1 Interactive Terminal Producer Verification CLI (`cmd/verifyproducer`)

#### Windows PowerShell:
```powershell
Set-Location "services\gateway"

# 1. Run interactive 4-stage producer verification (L2 + L3)
go run ./cmd/verifyproducer

# 2. Output machine-readable JSON certificate for CI/CD
go run ./cmd/verifyproducer -json

# 3. Test with custom amount and payer UUID partition key
go run ./cmd/verifyproducer -amount 50000 -payer "c0000000-0000-0000-0000-000000000088"
```

#### Bash (Linux / macOS):
```bash
cd services/gateway

# 1. Run interactive 4-stage producer verification (L2 + L3)
go run ./cmd/verifyproducer

# 2. Output machine-readable JSON certificate for CI/CD
go run ./cmd/verifyproducer -json

# 3. Test with custom amount and payer UUID partition key
go run ./cmd/verifyproducer -amount 50000 -payer "c0000000-0000-0000-0000-000000000088"
```

---

### 2.2 Automated Day 16 Master Integration Test Suite

#### Windows PowerShell:
```powershell
Set-Location "services\gateway"

# Run master integration test suite against live Docker Kafka container:
go test -v -run TestDay16_KafkaProducerMasterSuite .

# Run producer unit test suite:
go test -v ./internal/streaming/...
```

#### Bash (Linux / macOS):
```bash
cd services/gateway

# Run master integration test suite against live Docker Kafka container:
go test -v -run TestDay16_KafkaProducerMasterSuite .

# Run producer unit test suite:
go test -v ./internal/streaming/...
```

---

### 2.3 Real-Time Topic Inspection (Inside Docker Kafka Container)

Tail the `novaris.tx.initiated` topic to watch financial events arrive in real-time:

#### Windows PowerShell & Bash:
```powershell
docker exec -it novaris-kafka /bin/kafka-console-consumer `
  --bootstrap-server localhost:9092 `
  --topic novaris.tx.initiated `
  --from-beginning `
  --property print.key=true `
  --property print.timestamp=true
```

---

## 3. Day 17: Kafka Consumer Groups & In-Order Settlement Runbook

*Day 17 introduces cooperative consumer groups (`novaris-settlement-workers`), manual offset commit control via `CommitMessages()`, and the decoupled settlement worker service.*

### 3.1 Run Settlement Service Unit Tests
```powershell
Set-Location "services/settlement"
go test -v ./...
```

### 3.2 Run Streaming Consumer Subsystem Unit Tests
```powershell
Set-Location "services/gateway"
go test -v -run "TestConsumer|TestMockConsumer" ./internal/streaming
```

### 3.3 Run Day 17 Consumer Verification CLI Tool (`cmd/verifyconsumer`)
```powershell
# Interactive ANSI Terminal Audit (Proves in-order FIFO consumption & offset commits)
Set-Location "services/gateway"
go run ./cmd/verifyconsumer

# Machine-Readable JSON Output for CI/CD Pipeline Gates
go run ./cmd/verifyconsumer -json

# Custom Event Stream Volume (e.g. 20 Sequential Events)
go run ./cmd/verifyconsumer -events 20 -payer "c0000000-0000-0000-0000-000000000088"
```

### 3.4 Run Master Day 17 Live Cluster Integration Suite
```powershell
Set-Location "services/gateway"
go test -v -run "TestDay17"
```

### 3.5 Launch Settlement Consumer Daemon
```powershell
Set-Location "services/settlement"
go run . -brokers "localhost:9092" -group "novaris-settlement-workers"
```

### 3.6 Inspect Consumer Groups & Consumer Lag Using Docker CLI
```powershell
# List all active consumer groups on the cluster:
docker exec -it novaris-kafka /opt/kafka/bin/kafka-consumer-groups.sh `
    --bootstrap-server localhost:9092 `
    --list

# Describe consumer group offset progress and partition lag:
docker exec -it novaris-kafka /opt/kafka/bin/kafka-consumer-groups.sh `
    --bootstrap-server localhost:9092 `
    --describe `
    --group novaris-settlement-workers
```

---

## 4. Day 18: Transactional Outbox Pattern & Zero-Loss Crash Recovery Runbook

*Day 18 defeats the Dual-Write dilemma by committing outgoing events into an `outbox` table within the same ACID database transaction as ledger writes, published by a non-blocking `SKIP LOCKED` relay daemon.*

### 4.1 Run Outbox Subsystem Unit Tests
```powershell
Set-Location "services/gateway"
go test -v ./internal/outbox/...
```

### 4.2 Run Settlement & OCC with Outbox Unit Tests
```powershell
Set-Location "services/gateway"
go test -v -run "TestSettleTransactionWithOutbox|TestOCC_SettleWithOutbox" .
```

### 4.3 Run Master Day 18 Outbox Integration Test Suite
```powershell
Set-Location "services/gateway"
go test -v -run "TestDay18_OutboxMasterSuite" .
```

### 4.4 Run Day 18 Outbox & Crash Recovery Diagnostic CLI (`cmd/verifyoutbox`)
```powershell
# Interactive 5-Stage Verification (Rollback safety, atomic enlistment, sweep, crash simulation, deduplication)
Set-Location "services/gateway"
go run ./cmd/verifyoutbox -offline -count=5

# Machine-Readable JSON Output for CI/CD Pipeline Gates
go run ./cmd/verifyoutbox -offline -json

# Live Cluster Sweep (Requires Docker PostgreSQL & Kafka active)
go run ./cmd/verifyoutbox -count=10
```

### 4.5 Inspect Outbox Table & Partial Index in PostgreSQL
```powershell
# Connect to PostgreSQL via Docker Compose
docker exec -it novaris-postgres psql -U novaris_admin -d novaris_db

# Inspect outbox table schema
\d outbox;

# Inspect pending records partial index
\d idx_outbox_pending;

# Check pending vs published outbox records
SELECT status, count(*) FROM outbox GROUP BY status;
```

---

## 5. Day 19: Dead Letter Queues (DLQ) & Staged Retries Runbook (Preview)

*Day 19 builds the automated poison-pill triage router and staged backoff queues.*

```powershell
# Upcoming Day 19 test command:
cd services\gateway
go test -v -run TestDay19DeadLetterQueue .
```

---

## 6. Master Phase 2 Git Multi-Commit & Push Workflow

To maintain GitHub velocity and clear commit histories, use semantic commits for every day's deliverables.

```powershell
# Check modified and untracked files
git status

# Stage streaming core changes
git add services/gateway/internal/streaming/
git commit -m "feat(streaming): implement kafka core config, admin client, and partitioner"

# Stage verification CLI
git add services/gateway/cmd/verifykafka/
git commit -m "feat(cli): add Day 15 verifykafka diagnostic CLI"

# Stage action logs and runbook
git add bin/Daily_Actions/Phase_2_Streaming_Core/
git commit -m "docs(phase2): add Day 15 Action Log and Phase 2 testing runbook"

# Push to remote main branch
git push origin main
```

---

<div align="center">

<!-- Bottom Banner -->
<p align="center">
  <img src="../../../assets/footer.svg" alt="NOVARIS Footer" width="100%" />
</p>

<p align="center">
  <strong>NOVARIS Engineering Runbook • Phase 2 Streaming Core • Maintained by <a href="https://github.com/charan21042005">Tokachichu Sri Charan</a></strong>
</p>

</div>
