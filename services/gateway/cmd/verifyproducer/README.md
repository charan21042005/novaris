<div align="center">

<!-- Top Banner -->
<p align="center">
  <img src="../../../../assets/banner.svg" alt="NOVARIS Banner" width="100%" />
</p>

# 📤 NOVARIS Kafka Producer Verification CLI (`/cmd/verifyproducer`)
### Phase 2 Day 16 Diagnostic Tool, `tx.initiated` Event Publisher & Topic Visibility Auditor

<p align="center">
  <img src="https://img.shields.io/badge/Phase_2-Streaming_Core-00ADD8?style=flat-square&logo=apachekafka&logoColor=white" alt="Phase 2" />
  <img src="https://img.shields.io/badge/Language-Go_1.22+-00ADD8?style=flat-square&logo=go&logoColor=white" alt="Go" />
  <img src="https://img.shields.io/badge/Milestone-Day_16_Kafka_Producer-00C853?style=flat-square&logo=apachekafka&logoColor=white" alt="Day 16" />
  <img src="https://img.shields.io/badge/Event-novaris.tx.initiated-7C4DFF?style=flat-square&logo=eventstore&logoColor=white" alt="Event" />
  <img src="https://img.shields.io/badge/Outcome-Events_Visible_in_Topic-FFA000?style=flat-square&logo=checkmarx&logoColor=white" alt="Outcome" />
</p>

---

</div>

## 📌 Tool Overview

`verifyproducer` is the dedicated diagnostic, auditing, and certification CLI introduced in **Phase 2 (Day 16)**. It executes an end-to-end verification proving that the NOVARIS Go Kafka Producer successfully converts financial payment requests into durable, immutable events and that these events are **physically visible in the Kafka topic log**.

The tool executes a 4-stage verification loop:
1. **[1/4] Broker Reachability & TCP Ping:** Probes Layer-4 TCP connectivity to the bootstrap Kafka broker (`localhost:9092`), measuring network connect latency.
2. **[2/4] Topic Catalog Verification:** Asserts that `novaris.tx.initiated` is provisioned across 3 partitions with Murmur2 deterministic routing.
3. **[3/4] Direct Event Production:** Instantiates `streaming.KafkaProducer`, constructs an authentic `TransactionInitiatedEvent`, serializes to JSON, routes by Payer UUID partition key, and awaits Leader ACK (`RequiredAcks = RequireOne`).
4. **[4/4] Topic Visibility Audit (Milestone Outcome):** Connects a lightweight Kafka reader to the topic, reads committed messages, deserializes the JSON payload, and proves that the exact `EventID`, `TransactionID`, and `AmountCents` are durable and visible in the Kafka commit log.

```
       ┌─────────────────────────────────────────────────────────────┐
       │             cmd/verifyproducer Proving Loop                 │
       └──────────────────────────────┬──────────────────────────────┘
                                      │
         ┌────────────────────────────┼────────────────────────────┐
         ▼                            ▼                            ▼
  [1/4] TCP Ping               [2/4] Topic Check            [3/4] Publish Event
  - QuickTCPProbe              - novaris.tx.initiated       - TransactionInitiatedEvent
  - RTT Socket Latency         - 3 Partitions Verified      - Murmur2 Key Hashing
  - Broker Inventory           - Topic Specs Confirmed      - Synchronous Leader ACK
         │                            │                            │
         └────────────────────────────┼────────────────────────────┘
                                      │
                                      ▼
       ┌─────────────────────────────────────────────────────────────┐
       │  [4/4] Topic Visibility Audit (Read Back from Kafka Log)    │
       │                                                             │
       │  * Attaches kafka.Reader to novaris.tx.initiated            │
       │  * Fetches committed message at target partition offset     │
       │  * Asserts: EventID, TxID, PayerID, and Amount match 100%   │
       │                                                             │
       │  OUTCOME: "Events visible in the Kafka topic" (100% PASS)   │
       └─────────────────────────────────────────────────────────────┘
```

---

## ⚙️ CLI Flags & Parameters

| Flag | Type | Default | Description |
| :--- | :---: | :---: | :--- |
| `-brokers` | `string` | `localhost:9092` | Comma-separated list of target Kafka brokers. |
| `-client-id` | `string` | `novaris-verifyproducer-cli` | Client identifier string sent during TCP wire handshakes. |
| `-topic` | `string` | `novaris.tx.initiated` | Target event topic for verification. |
| `-amount` | `int64` | `12500` | Payment magnitude in integer cents ($125.00 = 12500). |
| `-payer` | `string` | `c0000000-0000-0000-0000-000000000001` | Payer account UUID (Kafka Partition Key). |
| `-payee` | `string` | `c0000000-0000-0000-0000-000000000002` | Payee recipient account UUID. |
| `-json` | `bool` | `false` | Emits full audit findings as machine-readable JSON for CI/CD pipelines. |

---

## 🚀 Execution & Usage Examples

### 1. Interactive Terminal Audit
```powershell
Set-Location "services/gateway"
go run ./cmd/verifyproducer
```

### 2. Export Machine-Readable JSON Certificate (CI/CD)
```powershell
go run ./cmd/verifyproducer -json
```

### 3. Custom Transfer Amount & Payer Account
```powershell
go run ./cmd/verifyproducer -amount 50000 -payer "c0000000-0000-0000-0000-000000000088"
```

---

## 🔗 Related Components

* ⚡ **[internal/streaming](../../internal/streaming/README.md):** The underlying Kafka producer, partitioner, and admin engine.
* 🛡️ **[services/gateway/main.go](../../main.go):** The API Gateway that emits `tx.initiated` events on payment settlement.
* 📖 **[Phase 2 Day 16 Action Log](../../../../bin/Daily_Actions/Phase_2_Streaming_Core/Day_16_Action_Log.md):** Comprehensive textbook covering Kafka producer internals and delivery semantics.

---

<div align="center">

<!-- Bottom Banner -->
<p align="center">
  <img src="../../../../assets/footer.svg" alt="NOVARIS Footer" width="100%" />
</p>

### 🎓 NOVARIS API Gateway Developer Tooling
*Kafka Producer Verification & Event Visibility Auditor*

</div>
