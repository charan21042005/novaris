<div align="center">

<!-- Top Banner -->
<p align="center">
  <img src="../../../../assets/banner.svg" alt="NOVARIS Banner" width="100%" />
</p>

# ⚡ NOVARIS Kafka Cluster Verification CLI (`/cmd/verifykafka`)
### Phase 2 Streaming Core Diagnostic Tool, KRaft Quorum Auditor & Partition Routing Prover

<p align="center">
  <img src="https://img.shields.io/badge/Phase_2-Streaming_Core-00ADD8?style=flat-square&logo=apachekafka&logoColor=white" alt="Phase 2" />
  <img src="https://img.shields.io/badge/Language-Go_1.22+-00ADD8?style=flat-square&logo=go&logoColor=white" alt="Go" />
  <img src="https://img.shields.io/badge/Engine-KRaft_Raft_Quorum-FF6D00?style=flat-square&logo=apachekafka&logoColor=white" alt="KRaft" />
  <img src="https://img.shields.io/badge/Routing-Deterministic_Murmur2_Proof-00C853?style=flat-square&logo=hyperledger&logoColor=white" alt="Routing" />
  <img src="https://img.shields.io/badge/Probing-Synchronous_Leader_ACK-7C4DFF?style=flat-square&logo=prometheus&logoColor=white" alt="ACK" />
</p>

---

</div>

## 📌 Tool Overview

`verifykafka` is the official diagnostic, auditing, and health certification CLI introduced in **Phase 2 (Day 15)**. It connects to the local or remote Apache Kafka cluster to inspect broker topology, verify KRaft quorum health, idempotently provision canonical Novaris event topics, prove partition routing determinism, and test the end-to-end write path.

The tool executes a 5-stage verification sequence:
1. **KRaft Reachability & Metadata Discovery:** Probes TCP socket connectivity and queries broker node IDs, roles, and the elected KRaft controller node.
2. **Canonical Topic Provisioning:** Ensures all 6 core Novaris streaming topics (`novaris.tx.initiated`, `settled`, `failed`, `dlq`, `retry.1`, `retry.2`) exist with 3 partitions and a replication factor of 1.
3. **Active Cluster Topic Catalog Inspection:** Audits live broker topic registries, displaying partition counts and classifying user streams versus internal metadata topics.
4. **Deterministic Partition Routing Proof:** Simulates high-velocity transactions across distinct merchant accounts, mathematically proving that any given account UUID routes strictly to the exact same partition every time ($P = |\text{Murmur2}(\text{Key})| \pmod 3$).
5. **Diagnostic Probe Write Path:** Dispatches an authentic diagnostic probe message to `novaris.tx.initiated` and measures nanosecond round-trip latency waiting for broker Leader ACK (`RequiredAcks = RequireOne`).

```
       ┌─────────────────────────────────────────────────────────────┐
       │             cmd/verifykafka 5-Stage Verification            │
       └──────────────────────────────┬──────────────────────────────┘
                                      │
         ┌────────────────────────────┼────────────────────────────┐
         ▼                            ▼                            ▼
  [1/5] KRaft Metadata         [2/5] Topic Provisioning     [3/5] Topic Catalog
  - TCP Ping Latency           - 6 Canonical Topics         - novaris.tx.initiated
  - KRaft Controller Node      - 3 Partitions Each          - novaris.tx.settled
  - Active Broker Inventory    - Idempotent Verification    - novaris.tx.dlq ...
         │                            │                            │
         └────────────────────────────┼────────────────────────────┘
                                      │
         ┌────────────────────────────┴────────────────────────────┐
         ▼                                                         ▼
  [4/5] Partition Routing Proof                            [5/5] Probe Write Path
  - Murmur2 Hash Modulo 3                                  - Ingress Test Message
  - Per-Account FIFO Invariant                             - Synchronous Leader ACK
  - Uniform Key Distribution                               - Socket Latency Metric
                                      │
                                      ▼
       ┌─────────────────────────────────────────────────────────────┐
       │             ANSI Terminal Dashboard & JSON Certificate      │
       │               STATUS: DAY 15 MILESTONE CERTIFIED            │
       └─────────────────────────────────────────────────────────────┘
```

---

## ⚙️ CLI Flags & Parameters

| Flag | Type | Default | Description |
| :--- | :---: | :---: | :--- |
| `-brokers` | `string` | `localhost:9092` | Comma-separated list of Kafka broker endpoints. |
| `-client-id` | `string` | `novaris-verifykafka-cli` | Client identifier string sent during TCP wire handshakes. |
| `-json` | `bool` | `false` | Emits full audit findings as machine-readable JSON for CI/CD pipelines. |

---

## 🚀 Execution & Usage Examples

### 1. Interactive Terminal Audit
```powershell
Set-Location "services/gateway"
go run ./cmd/verifykafka
```

### 2. Export JSON Certificate
```powershell
go run ./cmd/verifykafka -json
```

### 3. Target Custom Kafka Broker
```powershell
go run ./cmd/verifykafka -brokers "kafka-prod.novaris.internal:9092"
```

### Sample Output:
```text
================================================================================
   NOVARIS • PHASE 2 DAY 15: KAFKA CLUSTER & STREAMING VERIFICATION CLI        
================================================================================
Timestamp:           2026-11-26T08:33:05Z
Target Brokers:     localhost:9092
Milestone Focus:    Kafka Fundamentals (Topics, Partitions, Brokers)

[1/5] KAFKA CLUSTER REACHABILITY & KRAFT METADATA:
  ✔ Status:             REACHABLE (Protocol Round-Trip Ping: 17.36 ms)
  ✔ Connected Broker:   localhost:9092
  ✔ KRaft Controller:   Node ID 1 (localhost:9092)
  ✔ Active Brokers:     1 broker(s) registered in Raft metadata quorum
      - Node 1 at localhost:9092 [Broker + KRaft Controller (Combined Role)]

[2/5] NOVARIS CANONICAL TOPIC PROVISIONING:
  ✔ Provisioning Summary: 6 requested, 0 newly created, 6 active/verified (10.4449ms)
      • novaris.tx.initiated       [ALREADY_EXISTS] Topic active with 3 partition(s)
      • novaris.tx.settled         [ALREADY_EXISTS] Topic active with 3 partition(s)
      • novaris.tx.failed          [ALREADY_EXISTS] Topic active with 3 partition(s)
      • novaris.tx.dlq             [ALREADY_EXISTS] Topic active with 3 partition(s)
      • novaris.tx.retry.1         [ALREADY_EXISTS] Topic active with 3 partition(s)
      • novaris.tx.retry.2         [ALREADY_EXISTS] Topic active with 3 partition(s)

[3/5] ACTIVE CLUSTER TOPIC CATALOG (Including KRaft Metadata):
      - novaris.tx.settled         | 3 partition(s) | User Stream
      - novaris.tx.retry.2         | 3 partition(s) | User Stream
      - novaris.tx.failed          | 3 partition(s) | User Stream
      - novaris.tx.dlq             | 3 partition(s) | User Stream
      - novaris.tx.retry.1         | 3 partition(s) | User Stream
      - novaris.tx.initiated       | 3 partition(s) | User Stream

[4/5] DETERMINISTIC PARTITION KEY ROUTING PROOF (Per-Account FIFO):
      • acc_merchant_alpha_001 -> Partition 0 (Invariant: all events for acc_merchant_alpha_001 land here)
      • acc_user_bob_4920 -> Partition 1 (Invariant: all events for acc_user_bob_4920 land here)
      • acc_merchant_strata_88 -> Partition 0 (Invariant: all events for acc_merchant_strata_88 land here)
      • acc_treasury_escrow -> Partition 0 (Invariant: all events for acc_treasury_escrow land here)
      • acc_fee_collection -> Partition 2 (Invariant: all events for acc_fee_collection land here)

[5/5] DIAGNOSTIC PROBE WRITE PATH VERIFICATION:
  ✔ Target Topic:         novaris.tx.initiated
  ✔ Partition Routed:     Partition 0
  ✔ Routing Key:          acc_merchant_alpha_001
  ✔ Payload Size:         143 bytes
  ✔ Write Latency:        1.0118688s (Leader ACK Confirmed)

================================================================================
   FINAL OUTCOME: SUCCESS: All Day 15 Kafka Milestones Verified & Reachable   
================================================================================
```

---

## 🔗 Related Components

* ⚡ **[internal/streaming](../../internal/streaming/README.md):** The underlying Kafka client, partitioner, and administration engine.
* 📖 **[Phase 2 Day 15 Action Log](../../../../bin/Daily_Actions/Phase_2_Streaming_Core/Day_15_Action_Log.md):** Full systems theory on KRaft consensus, Zero-Copy OS PageCache, and partitioner mathematics.
* 📋 **[Phase 2 Commands Runbook](../../../../bin/Daily_Actions/Phase_2_Streaming_Core/COMMANDS_RUNBOOK.md):** Quick reference for Docker Compose Kafka commands.

---

<div align="center">

<!-- Bottom Banner -->
<p align="center">
  <img src="../../../../assets/footer.svg" alt="NOVARIS Footer" width="100%" />
</p>

### 🎓 NOVARIS API Gateway Developer Tooling
*Kafka KRaft Cluster Health & Streaming Verification Tool*

</div>
