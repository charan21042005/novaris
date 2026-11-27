<div align="center">

<!-- Top Banner -->
<p align="center">
  <img src="../../../../assets/banner.svg" alt="NOVARIS Banner" width="100%" />
</p>

# 📥 NOVARIS Kafka Consumer Verification CLI (`/cmd/verifyconsumer`)
### Phase 2 Day 17 Diagnostic Tool, In-Order Consumption Auditor & Consumer Group Prover

<p align="center">
  <img src="https://img.shields.io/badge/Phase_2-Streaming_Core-00ADD8?style=flat-square&logo=apachekafka&logoColor=white" alt="Phase 2" />
  <img src="https://img.shields.io/badge/Language-Go_1.22+-00ADD8?style=flat-square&logo=go&logoColor=white" alt="Go" />
  <img src="https://img.shields.io/badge/Milestone-Day_17_Kafka_Consumer-00C853?style=flat-square&logo=apachekafka&logoColor=white" alt="Day 17" />
  <img src="https://img.shields.io/badge/Consumer_Group-Dynamic_Audit_Group-7C4DFF?style=flat-square&logo=target&logoColor=white" alt="Consumer Group" />
  <img src="https://img.shields.io/badge/Outcome-Consumer_Reads_Every_Event-FFA000?style=flat-square&logo=checkmarx&logoColor=white" alt="Outcome" />
</p>

---

</div>

## 📌 Tool Overview

`verifyconsumer` is the dedicated diagnostic, auditing, and certification CLI introduced in **Phase 2 (Day 17)**. It executes an end-to-end verification proving that NOVARIS Consumer Groups reliably consume events from Apache Kafka in strict chronological sequence, process them without message loss, and commit offsets back to Kafka.

The tool executes a 4-stage verification loop:
1. **[1/4] Broker Reachability & TCP Ping:** Probes Layer-4 TCP connectivity to the bootstrap Kafka broker (`localhost:9092`), measuring network connect latency.
2. **[2/4] Topic & Consumer Group Initialization:** Asserts that `novaris.tx.initiated` exists and initializes a dynamic test consumer group.
3. **[3/4] Sequential Batch Production:** Generates and publishes $N$ sequentially numbered financial events ($Seq 1, Seq 2, Seq 3...$) using a fixed Payer Account UUID to ensure all events land on the identical partition.
4. **[4/4] In-Order Consumption & Offset Commit Audit (Milestone Outcome):** Boots `streaming.ConsumerGroup`, reads the event batch, asserts that all $N$ events arrive in strict monotonic order ($1 \rightarrow 2 \rightarrow 3 \rightarrow 4 \rightarrow 5$), verifies zero message drops, and confirms offset commits.

```
       ┌─────────────────────────────────────────────────────────────┐
       │             cmd/verifyconsumer Proving Loop                 │
       └──────────────────────────────┬──────────────────────────────┘
                                      │
         ┌────────────────────────────┼────────────────────────────┐
         ▼                            ▼                            ▼
  [1/4] TCP Ping               [2/4] Topic & Group          [3/4] Publish Batch
  - QuickTCPProbe              - novaris.tx.initiated       - N Sequenced Events
  - Socket Latency RTT         - Dynamic Group UUID         - Fixed Payer UUID Key
  - Broker Inventory           - Partition Count: 3         - Murmur2 Single Partition
         │                            │                            │
         └────────────────────────────┼────────────────────────────┘
                                      │
                                      ▼
       ┌─────────────────────────────────────────────────────────────┐
       │  [4/4] In-Order Consumption & Offset Commit Audit           │
       │                                                             │
       │  * ConsumerGroup reads batch from FirstOffset               │
       │  * Asserts: Every event read without loss (Zero Drops)      │
       │  * Asserts: Sequence is strictly monotonic (FIFO preserved) │
       │  * Asserts: Offsets committed to __consumer_offsets         │
       │                                                             │
       │  OUTCOME: "Consumer reads every published event, in order"  │
       └─────────────────────────────────────────────────────────────┘
```

---

## ⚙️ CLI Flags & Parameters

| Flag | Type | Default | Description |
| :--- | :---: | :---: | :--- |
| `-brokers` | `string` | `localhost:9092` | Comma-separated list of target Kafka brokers. |
| `-topic` | `string` | `novaris.tx.initiated` | Target event topic for verification. |
| `-events` | `int` | `5` | Number of sequential events to publish and verify. |
| `-payer` | `string` | `c0000000-0000-0000-0000-000000000088` | Payer account UUID (Kafka Partition Key). |
| `-json` | `bool` | `false` | Emits full audit findings as machine-readable JSON for CI/CD pipelines. |

---

## 🚀 Execution & Usage Examples

### 1. Interactive Terminal Audit
```powershell
Set-Location "services/gateway"
go run ./cmd/verifyconsumer
```

### 2. Export Machine-Readable JSON Certificate (CI/CD)
```powershell
go run ./cmd/verifyconsumer -json
```

### 3. Verify High-Volume Event Stream (e.g. 20 Sequential Events)
```powershell
go run ./cmd/verifyconsumer -events 20
```

---

## 🔗 Related Components

* ⚡ **[internal/streaming](../../internal/streaming/README.md):** The underlying Kafka consumer group, reader engine, and partitioner.
* 💳 **[services/settlement](../../../settlement/README.md):** The decoupled settlement consumer daemon that processes `tx.initiated` events.
* 📖 **[Phase 2 Day 17 Action Log](../../../../bin/Daily_Actions/Phase_2_Streaming_Core/Day_17_Action_Log.md):** Comprehensive textbook covering consumer groups, offset commits, rebalancing, and delivery semantics.

---

<div align="center">

<!-- Bottom Banner -->
<p align="center">
  <img src="../../../../assets/footer.svg" alt="NOVARIS Footer" width="100%" />
</p>

### 🎓 NOVARIS Consumer Verification Engine
*Strict FIFO Stream Validation & Offset Audit Infrastructure*

</div>
