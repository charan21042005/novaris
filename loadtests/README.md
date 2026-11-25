# 🚀 NOVARIS Concurrency & Load Testing Engine

Comprehensive load testing and stress-testing suite for the **NOVARIS Distributed Payment Gateway**.

Designed for **Day 13** of the 84-day engineering curriculum to evaluate:
1. **Hot-Wallet Database Contention:** Hundreds of concurrent payment settlement goroutines colliding on the exact same account row, asserting Optimistic Concurrency Control (OCC) version consistency.
2. **Multi-Tenant Rate Limiting:** Enforcing per-merchant token-bucket isolation without cross-tenant interference.
3. **Defense-in-Depth Latency:** Profiling the end-to-end overhead of the 5-layer perimeter (JWT $\to$ Limiter $\to$ HMAC $\to$ Idempotency $\to$ OCC Ledger).

---

## 🛠️ Testing Options: k6 vs. Native Go Load Tester

We provide **two complementary load testing engines**:

| Dimension | Option A: Grafana k6 (`loadtests/*.js`) | Option B: Native Go CLI (`cmd/loadtest`) |
|:---|:---|:---|
| **Language** | Modern JavaScript (ES6) | Pure Go 1.22+ Standard Library |
| **Dependencies** | Requires `k6` CLI or Docker | **Zero external dependencies** (runs anywhere) |
| **Best For** | Complex multi-stage ramping & Grafana dashboards | Instant local verification & CI/CD pipeline runs |
| **Conventions** | Virtual Users (VUs) and iterations | Goroutines, atomic counters, and synchronization gates |

---

## 📦 Option A: Running with Grafana k6

### 1. Install k6

#### Windows:
```powershell
# Via winget:
winget install k6 --source winget

# Via Chocolatey:
choco install k6
```

#### macOS:
```bash
brew install k6
```

#### Linux:
```bash
sudo gpg -k
sudo gpg --no-default-keyring --keyring /usr/share/keyrings/k6-archive-keyring.gpg --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
sudo apt-get update
sudo apt-get install k6
```

#### Docker (Zero Installation):
```bash
docker run --rm -i -v ${PWD}/loadtests:/loadtests --net=host grafana/k6 run /loadtests/k6_hotwallet_load.js
```

---

### 2. Execute Load Scripts

Ensure the gateway is running in another terminal (`cd services/gateway && go run main.go`).

#### Run Hot-Wallet Contention Load Test:
```powershell
k6 run loadtests/k6_hotwallet_load.js
```

#### Run Multi-Merchant Distributed Settlement Load Test:
```powershell
k6 run loadtests/k6_concurrent_settlement.js
```

#### Pass Custom Parameters via Environment Variables:
```powershell
k6 run -e GATEWAY_URL=http://localhost:8080 -e SENDER_ID=c0000000-0000-0000-0000-000000000001 loadtests/k6_hotwallet_load.js
```

---

## ⚡ Option B: Running the Native Go Load Tester CLI

If you don't have `k6` installed, NOVARIS includes a built-in high-throughput Go load generator located at [`services/gateway/cmd/loadtest/`](../services/gateway/cmd/loadtest/):

```powershell
Set-Location "services/gateway"

# 1. Quick verification: 50 concurrent workers, 250 total payment transactions
go run ./cmd/loadtest -workers 50 -requests 250

# 2. Extreme hot-wallet stress test: 100 concurrent workers hitting 1 wallet
go run ./cmd/loadtest -workers 100 -requests 1000 -hotwallet

# 3. High-throughput distributed test across 10 distinct merchants
go run ./cmd/loadtest -workers 50 -requests 500 -merchants 10
```

---

## 📊 Understanding Load Test Metrics

When k6 or the Go load tester finishes, examine the terminal telemetry summary:

```
     ✓ status is 200, 409, or 429
     ✓ no HTTP 500 server crashes
     ✓ has JSON response body

     novaris_settled_transactions...: 142    14.2/s
     novaris_rate_limited_429.......: 458    45.8/s
     novaris_occ_conflicts_409......: 24     2.4/s
     http_req_duration..............: avg=18.4ms min=1.1ms med=14.2ms max=112.5ms p(90)=38.1ms p(95)=64.7ms
```

### Telemetry Breakdown for CS Beginners:
- **`200 OK` (Settled Transactions):** Requests admitted through all 5 layers and successfully committed to the database.
- **`429 Too Many Requests` (Rate Limited):** Token-bucket quota exhaustion. **This is expected and healthy!** It proves the gateway protects the PostgreSQL database from crashing during volumetric traffic spikes.
- **`409 Conflict` (OCC Contention):** Occurs when multiple transactions attempt to update the same wallet version concurrently and exhaust their retry attempts.
- **`p(95)` (95th Percentile Latency):** 95% of all requests completed faster than this duration. In financial platforms, $p(95)$ and $p(99)$ matter much more than average latency because they represent the tail latency experienced by institutional clients.
