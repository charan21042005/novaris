<div align="center">

<!-- Top Banner -->
<p align="center">
  <img src="../assets/banner.svg" alt="NOVARIS Banner" width="100%" />
</p>

# 🚀 NOVARIS • Concurrency & Load Testing Engine (`/loadtests`)
### Industrial Load Generation, Hot-Wallet OCC Stress Testing & Tail Latency Profiling

<p align="center">
  <img src="https://img.shields.io/badge/Milestone-Day_13_of_84-00E5FF?style=flat-square&logo=target&logoColor=white" alt="Milestone" />
  <img src="https://img.shields.io/badge/Phase-1:_Design,_Ledger_%26_Perimeter-00C853?style=flat-square&logo=shield&logoColor=white" alt="Phase" />
  <img src="https://img.shields.io/badge/Engine-Grafana_k6_%7C_Go_CLI-7C4DFF?style=flat-square&logo=k6&logoColor=white" alt="Engine" />
  <img src="https://img.shields.io/badge/Concurrency-100_Worker_Simultaneous_Race-FFA000?style=flat-square&logo=speedtest&logoColor=white" alt="Concurrency" />
  <img src="https://img.shields.io/badge/Contention-Hot--Wallet_CAS_Convergence-00C853?style=flat-square&logo=affinitylock&logoColor=white" alt="Contention" />
  <img src="https://img.shields.io/badge/Audit-Zero_Race_Conditions_Verified-FF6D00?style=flat-square&logo=securityscorecard&logoColor=white" alt="Audit" />
</p>

---

</div>

## 📌 Directory Purpose & Architectural Objectives

This directory houses the comprehensive concurrency load testing and stress-testing infrastructure for the **NOVARIS Distributed Payment Platform**.

Introduced in **Day 13** of the 84-day engineering curriculum, this subsystem directly validates the Phase 1 capstone criterion: **"100 concurrent unit tests passing with zero race conditions."**

### Core Evaluation Scenarios:
1. **Hot-Wallet Database Contention:** Hundreds of concurrent payment settlement goroutines colliding on the exact same account row, asserting Optimistic Concurrency Control (OCC) version consistency and zero balance drift.
2. **Multi-Tenant Rate Limiting:** Enforcing per-merchant token-bucket isolation without cross-tenant interference or noisy-neighbor starvation.
3. **Defense-in-Depth Latency:** Profiling the end-to-end overhead of the 5-layer perimeter (JWT Auth $\to$ Token-Bucket Limiter $\to$ HMAC-SHA256 Signature $\to$ Idempotency Engine $\to$ OCC Ledger).
4. **Sudden Burst Shedding:** Proving that vertical traffic spikes (0 to 100 workers in 1 nanosecond) are cleanly absorbed by the edge limiter with HTTP 429 and **0 server crashes (zero HTTP 500s)**.

---

## 🛠️ Testing Options: k6 vs. Native Go Load Tester

We provide **two complementary load testing engines**:

| Dimension | Option A: Grafana k6 (`loadtests/*.js`) | Option B: Native Go CLI (`cmd/loadtest`) |
|:---|:---|:---|
| **Language** | Modern JavaScript (ES6) | Pure Go 1.22+ Standard Library |
| **Dependencies** | Requires `k6` CLI or Docker | **Zero external dependencies** (runs anywhere) |
| **Best For** | Multi-stage ramping profiles & Grafana Cloud dashboards | Instant local verification & CI/CD pipeline runs |
| **Execution Model** | Virtual Users (VUs) running inside Goja JS engines | Fixed worker pool of goroutines reading job channels |
| **SLO Quality Gates** | Programmatic thresholds (`p(95) < 350ms`, error rate $< 5\%$) | High-precision monotonic socket turnaround percentiles ($p50, p90, p95, p99$) |

---

## 📦 Option A: Running with Grafana k6

### 1. Install k6

#### Windows:
```powershell
# Via winget (recommended):
winget install k6 --source winget

# Via Chocolatey:
choco install k6
```

#### macOS:
```bash
brew install k6
```

#### Linux (Debian / Ubuntu):
```bash
sudo gpg -k
sudo gpg --no-default-keyring --keyring /usr/share/keyrings/k6-archive-keyring.gpg --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
sudo apt-get update
sudo apt-get install k6
```

#### Docker (Zero Local Installation):
```bash
docker run --rm -i -v "${PWD}/loadtests:/loadtests" --net=host grafana/k6 run /loadtests/k6_hotwallet_load.js
```

---

### 2. Available Load Testing Scripts

```
loadtests/
├── k6_hotwallet_load.js           # 100-VU phased hot-wallet contention test
├── k6_concurrent_settlement.js    # Distributed multi-tenant settlement load test
└── README.md                      # Architecture guide & execution runbook
```

#### 1. Hot-Wallet Contention Load Test (`k6_hotwallet_load.js`):
Models the worst-case scenario: 100 Virtual Users hammer the exact same wallet account simultaneously.
```powershell
k6 run loadtests/k6_hotwallet_load.js
```

#### 2. Multi-Merchant Distributed Settlement Load Test (`k6_concurrent_settlement.js`):
Partitions virtual users across distinct merchant tenants (`merchant_tenant_${__VU}`), proving horizontal scaling without noisy-neighbor contention:
```powershell
k6 run loadtests/k6_concurrent_settlement.js
```

#### 3. Custom Runtime Parameters via Environment Variables:
```powershell
k6 run -e GATEWAY_URL=http://localhost:8080 -e SENDER_ID=c0000000-0000-0000-0000-000000000001 loadtests/k6_hotwallet_load.js
```

---

## ⚡ Option B: Running the Native Go Load Tester CLI

If you do not have `k6` installed, NOVARIS provides a zero-dependency high-throughput Go load generator located at [`services/gateway/cmd/loadtest/`](../services/gateway/cmd/loadtest/):

```powershell
Set-Location "services/gateway"

# 1. Quick verification: 50 concurrent workers, 250 total payment transactions
go run ./cmd/loadtest -workers 50 -requests 250 -hotwallet=true

# 2. Extreme hot-wallet stress test: 100 concurrent workers, 1,000 requests
go run ./cmd/loadtest -workers 100 -requests 1000 -hotwallet=true

# 3. High-throughput distributed test across 10 distinct merchants
go run ./cmd/loadtest -workers 50 -requests 500 -hotwallet=false -tenants 10

# 4. Run automated unit test suite for the CLI itself
go test -v ./cmd/loadtest/...
```

---

## 📊 Understanding Load Test Metrics

When k6 or the native Go load tester finishes, examine the terminal telemetry summary:

```
================================================================================
🚀 NOVARIS HIGH-CONCURRENCY LOAD TEST EXECUTION REPORT
================================================================================
Target URL:          http://127.0.0.1:8080/v1/payments
Concurrency Workers: 50
Total Requests:      250
Elapsed Test Time:   242.15 ms
Throughput:          1032.42 requests/sec
--------------------------------------------------------------------------------
HTTP STATUS CODE DISTRIBUTION:
  [200 OK / Settled]       10 (  4.0%)
  [409 Conflict / OCC]    148 ( 59.2%)
  [429 Rate Limited]       92 ( 36.8%)
  [500 Server Crashes]      0 (  0.0%)  <-- ZERO SERVER CRASHES!
--------------------------------------------------------------------------------
LATENCY PERCENTILES (Monotonic Socket Turnaround):
  Min Latency:       0.45 ms
  Avg Latency:       2.14 ms
  p50 (Median):      1.82 ms
  p90:               3.95 ms
  p95:               5.12 ms
  p99:               9.84 ms
  Max Latency:      14.20 ms
================================================================================
```

### Telemetry Breakdown for CS Beginners:
- **`200 OK` (Settled Transactions):** Requests admitted through all 5 layers and successfully committed to the database.
- **`429 Too Many Requests` (Rate Limited):** Token-bucket quota exhaustion. **This is expected and healthy!** It proves the gateway protects the PostgreSQL database from crashing during volumetric traffic spikes.
- **`409 Conflict` (OCC Contention):** Occurs when multiple transactions attempt to update the same wallet version concurrently and exhaust their retry attempts.
- **`500 Internal Server Error` (Server Crashes):** Must be strictly **0**. Any non-zero count indicates unhandled panics, deadlocks, or connection pool exhaustion.
- **`p(95)` (95th Percentile Latency):** 95% of all requests completed faster than this duration. In financial platforms, $p(95)$ and $p(99)$ matter much more than average latency because they represent the tail latency experienced by institutional clients.

---

## 🔒 Security & Concurrency Invariants Enforced Under Load

1. **Double-Entry Balance Conservation:** $\sum \text{Debits} - \sum \text{Credits} = 0$. Money is neither created nor destroyed during high-concurrency races.
2. **Fail-Fast Defense-in-Depth:** Rate limiting and JWT authentication run before any database connection is acquired, preventing connection pool exhaustion under volumetric attacks.
3. **Monotonic Version Progression:** Account row version increments strictly match the count of settled transactions ($V_{\text{final}} = V_{\text{initial}} + \text{Settled}$).
4. **Exact-Once Idempotency:** Duplicate payment submissions with the identical `Idempotency-Key` return cached responses without charging accounts twice.

---

<div align="center">

<!-- Bottom Banner -->
<p align="center">
  <img src="../assets/footer.svg" alt="NOVARIS Footer" width="100%" />
</p>

### 🎓 NOVARIS Performance & Load Engineering
*Industrial Load Scripting • Hot-Wallet Contention • Verified Zero Race Conditions*

</div>
