<div align="center">

<!-- Top Banner -->
<p align="center">
  <img src="../../../../assets/banner.svg" alt="NOVARIS Banner" width="100%" />
</p>

# 🚀 NOVARIS High-Concurrency Load Tester (`/cmd/loadtest`)
### In-Memory Worker Pool Load Generator, Hot-Wallet OCC Stress Tester & Tail Latency Profiler

<p align="center">
  <img src="https://img.shields.io/badge/Phase_1_Capstone-COMPLETED_(100%25_Verified)-00C853?style=flat-square&logo=checkbox&logoColor=white" alt="Phase 1 Capstone" />
  <img src="https://img.shields.io/badge/Language-Go_1.22+-00ADD8?style=flat-square&logo=go&logoColor=white" alt="Go" />
  <img src="https://img.shields.io/badge/Concurrency-100+_Worker_Goroutines-FF6D00?style=flat-square&logo=speedtest&logoColor=white" alt="Workers" />
  <img src="https://img.shields.io/badge/Metrics-Nanosecond_Tail_Latency_(p99)-7C4DFF?style=flat-square&logo=prometheus&logoColor=white" alt="Metrics" />
  <img src="https://img.shields.io/badge/Stress_Target-Hot--Wallet_OCC_Contention-FFA000?style=flat-square&logo=vault&logoColor=white" alt="Target" />
  <img src="https://img.shields.io/badge/Dependencies-Zero_External_Dependencies-00C853?style=flat-square&logo=checkmarx&logoColor=white" alt="Deps" />
</p>

---

</div>

## 📌 Tool Overview

`loadtest` is a high-throughput, native Go load testing engine and benchmark harness introduced in **Day 13**. Designed to evaluate gateway performance and database concurrency under extreme pressure, it operates with **zero external dependencies** (eliminating the need for k6, JMeter, or Node.js).

The engine provides:
1. **Goroutine Worker Pool Architecture:** Spawns configurable concurrent worker goroutines (`-workers 50` or `-workers 100`) fed by thread-safe Go channels.
2. **Hot-Wallet Database Contention:** Directs all workers at the exact same merchant account UUID to stress PostgreSQL Optimistic Concurrency Control (OCC) version increments and verify that balance invariants remain strictly non-negative.
3. **Multi-Tenant Rate Limiting Stress:** Simulates distinct merchant identities (`-tenants 10`) to prove per-merchant token-bucket isolation under high load.
4. **Sub-Millisecond Telemetry & Tail Latencies:** Records nanosecond-precision round-trip times to compute min, mean, p50 (median), p90, p95, p99, and max latency percentiles alongside full HTTP status code distributions (200, 400, 401, 409, 422, 429, 500).

```
                      ┌────────────────────────────┐
                      │    Job Dispatch Channel    │
                      │   (250 - 1000 Transactions)│
                      └─────────────┬──────────────┘
                                    │
         ┌──────────────────────────┼──────────────────────────┐
         ▼                          ▼                          ▼
 ┌───────────────┐          ┌───────────────┐          ┌───────────────┐
 │ Worker Goroutine 1│      │ Worker Goroutine 2│      │ Worker Goroutine N│
 │ * Mints JWT   │          │ * Mints JWT   │          │ * Mints JWT   │
 │ * Signs HMAC  │          │ * Signs HMAC  │          │ * Signs HMAC  │
 │ * Dispatches  │          │ * Dispatches  │          │ * Dispatches  │
 └───────┬───────┘          └───────┬───────┘          └───────┬───────┘
         │                          │                          │
         └──────────────────────────┼──────────────────────────┘
                                    ▼
       ┌─────────────────────────────────────────────────────────────┐
       │             NOVARIS API Gateway (Port 8080)                 │
       │                                                             │
       │  * Ingress Token Bucket Rate Limiting (HTTP 429 on Burst)   │
       │  * ACID Settlement Engine (SELECT FOR UPDATE / OCC CAS)     │
       │  * Double-Entry Ledger Writes (Balanced Journal Entries)    │
       └────────────────────────────┬────────────────────────────────┘
                                    │ Latency & Status Codes
                                    ▼
       ┌─────────────────────────────────────────────────────────────┐
       │             Real-Time Telemetry & Percentile Aggregator     │
       │  Min | Mean | p50 (Median) | p90 | p95 | p99 | Max Latency  │
       └─────────────────────────────────────────────────────────────┘
```

---

## ⚙️ CLI Flags & Parameters

| Flag | Type | Default | Description |
| :--- | :---: | :---: | :--- |
| `-url` | `string` | `http://localhost:8080` | Target gateway base URL. |
| `-workers` | `int` | `50` | Number of concurrent worker goroutines. |
| `-requests` | `int` | `250` | Total number of payment transactions to execute. |
| `-hotwallet` | `bool` | `true` | Direct all workers at the same hot-wallet account to stress OCC. |
| `-tenants` | `int` | `1` | Number of distinct merchant tenants (when `-hotwallet=false`). |
| `-amount` | `int64` | `100` | Payment magnitude in integer cents ($1.00 = 100). |
| `-timeout` | `int` | `10` | Per-request socket timeout in seconds. |

---

## 🚀 Execution & Usage Examples

### 1. Rapid Smoke Benchmark (50 Workers, 250 Requests)
```powershell
Set-Location "services/gateway"
go run ./cmd/loadtest -workers 50 -requests 250
```

### 2. High-Contention Hot-Wallet Stress Test (100 Workers, 1,000 Requests)
```powershell
go run ./cmd/loadtest -workers 100 -requests 1000 -hotwallet=true
```

### 3. Distributed Multi-Tenant Tenant Isolation Test (10 Merchants)
```powershell
go run ./cmd/loadtest -workers 50 -requests 500 -hotwallet=false -tenants 10
```

### Sample Output:
```text
================================================================================
   NOVARIS • HIGH-CONCURRENCY GATEWAY LOAD TEST BENCHMARK                       
================================================================================
Target URL:        http://localhost:8080
Concurrency:       50 worker goroutines
Total Requests:    250 payment transactions
Workload Mode:     Hot-Wallet Contention (Single Account OCC Stress)

Dispatched 250 requests in 842.15 ms (Throughput: 296.86 req/sec)

STATUS CODE DISTRIBUTION:
  • HTTP 200 OK:                10 (Successful Initial Settlements)
  • HTTP 429 Too Many Requests: 240 (Rate-Limited by Ingress Token Bucket)
  • HTTP 409 Conflict:          0 (Zero Unhandled Database Collisions)
  • HTTP 500 Server Error:      0 (Zero Gateway Failures)

LATENCY PERCENTILES:
  • Min:    1.24 ms
  • Mean:   2.85 ms
  • p50:    2.41 ms (Median)
  • p90:    4.12 ms
  • p95:    5.30 ms
  • p99:    8.15 ms
  • Max:    11.42 ms

================================================================================
   BENCHMARK RESULT: PASS (Gateway maintained 100% integrity under stress)     
================================================================================
```

---

## 🧪 Automated Testing

Unit tests for `loadtest` evaluate mock HTTP servers and latency percentile calculations:

```powershell
Set-Location "services/gateway"
go test -v ./cmd/loadtest/...
```

---

## 🔗 Related Components

* ⏱️ **[internal/ratelimit](../../internal/ratelimit/README.md):** Ingress token-bucket rate limiter benchmarked by this utility.
* 🛡️ **[internal/auth](../../internal/auth/README.md):** High-speed cryptographic hashing benchmarked under concurrent load.
* 📖 **[Phase 1 Day 13 Action Log](../../../../bin/Daily_Actions/Phase_1_Design_and_Ledger/Day_13_Action_Log.md):** Detailed analysis of hot-wallet contention and token-bucket burst mechanics.

---

<div align="center">

<!-- Bottom Banner -->
<p align="center">
  <img src="../../../../assets/footer.svg" alt="NOVARIS Footer" width="100%" />
</p>

### 🎓 NOVARIS API Gateway Developer Tooling
*In-Memory Concurrency Benchmark & Tail Latency Profiler*

</div>
