<div align="center">

<!-- Top Banner -->
<p align="center">
  <img src="../../../assets/banner.svg" alt="NOVARIS Banner" width="100%" />
</p>

# ⚡ NOVARIS Phase 1 Verification & Testing Runbook
### Unified Copy-Paste Command Cheat Sheet Across Phase 1 Milestones (Days 01–14)

<p align="center">
  <img src="https://img.shields.io/badge/Runbook-Phase_1_Testing_Suite-00E5FF?style=flat-square&logo=gitbook&logoColor=white" alt="Runbook" />
  <img src="https://img.shields.io/badge/Days-01_to_11_Verified-00C853?style=flat-square&logo=googlecalendar&logoColor=white" alt="Days" />
  <img src="https://img.shields.io/badge/Shell-Bash_%26_PowerShell-7C4DFF?style=flat-square&logo=powershell&logoColor=white" alt="Shell" />
  <img src="https://img.shields.io/badge/Standard-Industrial_FinTech-FFA000?style=flat-square&logo=standards&logoColor=white" alt="Standard" />
</p>

---

</div>

## 📌 Purpose of This Runbook

This document is the **central command cheat sheet** for NOVARIS. Every command required to boot infrastructure, run migrations, compile microservices, mint test credentials, and execute automated and manual HTTP verification probes is cataloged here in copy-pasteable format for both **Bash** (Linux/macOS) and **Windows PowerShell**.

---

## 📑 Table of Contents

1. [Day 01: Docker Infrastructure, PostgreSQL Ledger & Health Probes](#1-day-01-docker-infrastructure-postgresql-ledger--health-probes)
2. [Day 02: JWT Authentication, Token CLI & Identity Probes](#2-day-02-jwt-authentication-token-cli--identity-probes)
3. [Day 03: HMAC-SHA256 Request Signing, Signer CLI & Replay Probes](#3-day-03-hmac-sha256-request-signing-signer-cli--replay-probes)
4. [Day 04: Token-Bucket Rate Limiter, Concurrency & 429 Probes](#4-day-04-token-bucket-rate-limiter-concurrency--429-probes)
5. [Day 05: Database Schema Migration & Verification Runbook (Composite Indexes & JSONB)](#5-day-05-database-schema-migration--verification-runbook-composite-indexes--jsonb)
6. [Day 06: Atomic Settlement & Double-Entry Ledger Verification Runbook](#6-day-06-atomic-settlement--double-entry-ledger-verification-runbook)
7. [Day 07: Table-Driven Testing, Code Coverage & Benchmark Runbook](#7-day-07-table-driven-testing-code-coverage--benchmark-runbook)
8. [Day 08: Concurrency Testing, Goroutine Collision & Race Detection Runbook](#8-day-08-concurrency-testing-goroutine-collision--race-detection-runbook)
9. [Day 09: Optimistic Concurrency Control (OCC) & Version Invariant Runbook](#9-day-09-optimistic-concurrency-control-occ--version-invariant-runbook)
10. [Day 10: Idempotent API Design & Safe Network Retries Runbook](#10-day-10-idempotent-api-design--safe-network-retries-runbook)
11. [Day 11: Integration Wiring, API Contracts & E2E Verification Runbook](#11-day-11-integration-wiring-api-contracts--e2e-verification-runbook)
12. [Master Daily Git Multi-Commit & Push Workflow](#12-master-daily-git-multi-commit--push-workflow)

---

## 1. Day 01: Docker Infrastructure, PostgreSQL Ledger & Health Probes

### 1.1 Start Container Mesh:
```bash
# Navigate to deployment directory
cd deployments

# Start PostgreSQL 16, Redis 7, and Kafka KRaft in background
docker compose up -d

# Check status of containers (all should report Up or Healthy)
docker compose ps
```

### 1.2 Verify Database Schema:
```bash
# Confirm all ledger tables were automatically migrated
docker exec -it novaris-postgres psql -U novaris -d novaris_ledger -c "\dt"

# Inspect the accounts table
docker exec -it novaris-postgres psql -U novaris -d novaris_ledger -c "\d accounts"

# Inspect the double-entry journal entries table
docker exec -it novaris-postgres psql -U novaris -d novaris_ledger -c "\d journal_entries"
```

### 1.3 Run Gateway Service (Terminal 1):
```bash
cd services/gateway
go run main.go
```

### 1.4 Test Healthcheck & Service Discovery (Terminal 2):
```bash
# Test healthcheck probe (Expect HTTP 200 OK)
curl -i http://localhost:8080/healthz

# Test service discovery (Expect HTTP 200 OK)
curl -i http://localhost:8080/
```
*PowerShell Alternative:*
```powershell
Invoke-RestMethod -Uri "http://localhost:8080/healthz"
Invoke-RestMethod -Uri "http://localhost:8080/"
```

---

## 2. Day 02: JWT Authentication, Token CLI & Identity Probes

### 2.1 Run Automated Unit Tests:
```bash
cd services/gateway
go test -v ./internal/auth -run Test.*Token|Test.*JWT
```

### 2.2 Mint Test JWT Tokens via CLI:
```bash
cd services/gateway

# Mint standard valid merchant token (15-minute lifespan)
TOKEN=$(go run cmd/gentoken/main.go -merchant merchant_123 -role merchant -mins 15)
echo "Token: $TOKEN"

# Mint deliberately expired token
EXPIRED=$(go run cmd/gentoken/main.go -merchant merchant_test -expired)
```
*PowerShell Alternative:*
```powershell
Set-Location services\gateway
$token = (go run cmd/gentoken/main.go -merchant merchant_123 -role merchant -mins 15)
$expired = (go run cmd/gentoken/main.go -merchant merchant_test -expired)
```

### 2.3 Execute 5 Identity Probes via `curl`:
```bash
# Case 1: Missing Token -> Expect 401 Unauthorized
curl -i http://localhost:8080/v1/payments

# Case 2: Valid Bearer Token -> Expect 200 OK with claims echoed back
curl -i http://localhost:8080/v1/payments \
  -H "Authorization: Bearer $TOKEN"

# Case 3: Tampered Token (modified payload) -> Expect 401 Unauthorized
curl -i http://localhost:8080/v1/payments \
  -H "Authorization: Bearer ${TOKEN}x"

# Case 4: Expired Token -> Expect 401 Unauthorized ("token is expired")
curl -i http://localhost:8080/v1/payments \
  -H "Authorization: Bearer $EXPIRED"

# Case 5: Public Route (/healthz) -> Expect 200 OK
curl -i http://localhost:8080/healthz
```

---

## 3. Day 03: HMAC-SHA256 Request Signing, Signer CLI & Replay Probes

### 3.1 Run All Automated Tests (10/10 PASS):
```bash
cd services/gateway
go test -v ./...
```

### 3.2 Compute HMAC Signatures via Developer CLI:
```bash
cd services/gateway

# Standard interactive signing output
go run cmd/signrequest/main.go

# Custom JSON body signing
go run cmd/signrequest/main.go -body '{"amount_cents":100000,"recipient":"vendor_xyz"}'

# Stale timestamp generation (10m in past for replay testing)
go run cmd/signrequest/main.go -offset -600

# Simulated in-flight tampering
go run cmd/signrequest/main.go -tamper
```

### 3.3 Execute 7 Integrity & Replay Probes via `curl`:
```bash
# Setup environment variables
TOKEN=$(go run cmd/gentoken/main.go -merchant merchant_stripe_live -role merchant -mins 15)
BODY='{"amount_cents":5000,"recipient":"merchant_456"}'
TS=$(date +%s)
SECRET="novaris_insecure_dev_hmac_secret_change_in_production_2026"
CANONICAL="$TS.$BODY"
SIG=$(echo -n "$CANONICAL" | openssl dgst -sha256 -hmac "$SECRET" | sed 's/^.* //')

# Probe 1: Public Healthcheck -> Expect 200 OK
curl -i http://localhost:8080/healthz

# Probe 2: Missing All Auth Headers -> Expect 401 Unauthorized
curl -i -X POST http://localhost:8080/v1/payments \
  -H "Content-Type: application/json" \
  -d "$BODY"

# Probe 3: JWT Only, Missing HMAC Headers -> Expect 401 Unauthorized
curl -i -X POST http://localhost:8080/v1/payments \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "$BODY"

# Probe 4: Fully Valid Signed Request -> Expect 200 OK (claims & body restored)
curl -i -X POST http://localhost:8080/v1/payments \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Novaris-Timestamp: $TS" \
  -H "X-Novaris-Signature: $SIG" \
  -H "Content-Type: application/json" \
  -d "$BODY"

# Probe 5: Tampered Request Body ($50 changed to $50,000) -> Expect 401 Unauthorized
curl -i -X POST http://localhost:8080/v1/payments \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Novaris-Timestamp: $TS" \
  -H "X-Novaris-Signature: $SIG" \
  -H "Content-Type: application/json" \
  -d '{"amount_cents":5000000,"recipient":"attacker"}'

# Probe 6: Stale Replay Timestamp (-10m in past) -> Expect 401 Unauthorized
STALE_TS=$(($TS - 600))
STALE_CANONICAL="$STALE_TS.$BODY"
STALE_SIG=$(echo -n "$STALE_CANONICAL" | openssl dgst -sha256 -hmac "$SECRET" | sed 's/^.* //')
curl -i -X POST http://localhost:8080/v1/payments \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Novaris-Timestamp: $STALE_TS" \
  -H "X-Novaris-Signature: $STALE_SIG" \
  -H "Content-Type: application/json" \
  -d "$BODY"

# Probe 7: Future Timestamp (+10m ahead) -> Expect 401 Unauthorized
FUTURE_TS=$(($TS + 600))
FUTURE_CANONICAL="$FUTURE_TS.$BODY"
FUTURE_SIG=$(echo -n "$FUTURE_CANONICAL" | openssl dgst -sha256 -hmac "$SECRET" | sed 's/^.* //')
curl -i -X POST http://localhost:8080/v1/payments \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Novaris-Timestamp: $FUTURE_TS" \
  -H "X-Novaris-Signature: $FUTURE_SIG" \
  -H "Content-Type: application/json" \
  -d "$BODY"
```

### 3.4 Execute Verification via Windows PowerShell:
```powershell
$headersValid = @{
    "Authorization"       = "Bearer " + $token
    "X-Novaris-Timestamp" = $ts
    "X-Novaris-Signature" = $sig
}

# Valid request (200 OK)
Invoke-RestMethod -Uri "http://localhost:8080/v1/payments" -Method Post -Body $body -ContentType "application/json" -Headers $headersValid

# Tampered request (401 Unauthorized)
try { Invoke-RestMethod -Uri "http://localhost:8080/v1/payments" -Method Post -Body '{"amount_cents":9999999}' -ContentType "application/json" -Headers $headersValid } catch { $_.Exception.Response.StatusCode.value__ }
```

---

## 4. Day 04: Token-Bucket Rate Limiter, Concurrency & 429 Probes

### 4.1 Run Rate Limiter Automated Tests:
```bash
cd services/gateway

# Run rate limiter unit and concurrency test suite
go test -v ./internal/ratelimit

# Run full gateway test suite (auth + ratelimit + integration)
go test -v ./...
```

### 4.2 Start Gateway with Rate Limiting Enabled:
```bash
# Start the gateway server
go run main.go
```

### 4.3 Burst & 429 Verification via Bash / cURL:
```bash
# Generate JWT token for merchant_stripe_101
TOKEN=$(go run cmd/gentoken/main.go -merchant merchant_stripe_101 | grep -o 'ey[A-Za-z0-9_-]*\.[A-Za-z0-9_-]*\.[A-Za-z0-9_-]*' | head -n 1)

# Payload & cryptographic HMAC signature
BODY='{"amount_cents":5000,"recipient":"merchant_coffee"}'
SECRET="novaris_insecure_dev_hmac_secret_change_in_production_2026"
TS=$(date +%s)
CANONICAL="$TS.$BODY"
SIG=$(echo -n "$CANONICAL" | openssl dgst -sha256 -hmac "$SECRET" | sed 's/^.* //')

# Probe 1: Send burst of 10 requests (All return HTTP 200 OK)
for i in {1..10}; do
  curl -s -o /dev/null -w "Request $i: HTTP %{http_code}\n" -X POST http://localhost:8080/v1/payments \
    -H "Authorization: Bearer $TOKEN" \
    -H "X-Novaris-Timestamp: $TS" \
    -H "X-Novaris-Signature: $SIG" \
    -H "Content-Type: application/json" \
    -d "$BODY"
done

# Probe 2: Send 11th request immediately (Returns HTTP 429 Too Many Requests + Retry-After: 1)
curl -i -X POST http://localhost:8080/v1/payments \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Novaris-Timestamp: $TS" \
  -H "X-Novaris-Signature: $SIG" \
  -H "Content-Type: application/json" \
  -d "$BODY"

# Probe 3: Send request from Merchant B during Merchant A throttling (Returns 200 OK - Noisy Neighbor Isolation)
TOKEN_B=$(go run cmd/gentoken/main.go -merchant merchant_isolated_beta | grep -o 'ey[A-Za-z0-9_-]*\.[A-Za-z0-9_-]*\.[A-Za-z0-9_-]*' | head -n 1)
curl -s -o /dev/null -w "Merchant B: HTTP %{http_code}\n" -X POST http://localhost:8080/v1/payments \
  -H "Authorization: Bearer $TOKEN_B" \
  -H "X-Novaris-Timestamp: $TS" \
  -H "X-Novaris-Signature: $SIG" \
  -H "Content-Type: application/json" \
  -d "$BODY"
```

### 4.4 Burst & 429 Verification via Windows PowerShell:
```powershell
cd services\gateway

# 1. Mint JWT token
$token = (go run cmd/gentoken/main.go -merchant merchant_stripe_101 | Select-String -Pattern "ey[A-Za-z0-9_-]+" | Select-Object -First 1).Matches.Value

# 2. Compute HMAC signature
$body = '{"amount_cents":5000,"recipient":"merchant_coffee"}'
$ts = [DateTimeOffset]::UtcNow.ToUnixTimeSeconds()
$key = [System.Text.Encoding]::UTF8.GetBytes("novaris_insecure_dev_hmac_secret_change_in_production_2026")
$hmac = New-Object System.Security.Cryptography.HMACSHA256
$hmac.Key = $key
$canonical = [System.Text.Encoding]::UTF8.GetBytes("$ts.$body")
$sig = [BitConverter]::ToString($hmac.ComputeHash($canonical)).Replace("-", "").ToLower()

$headers = @{
    "Authorization"       = "Bearer " + $token
    "X-Novaris-Timestamp" = $ts
    "X-Novaris-Signature" = $sig
}

# 3. Burst of 10 requests (All return 200 OK)
1..10 | ForEach-Object {
    $res = Invoke-RestMethod -Uri "http://localhost:8080/v1/payments" -Method Post -Body $body -ContentType "application/json" -Headers $headers
    Write-Host "Request $_ : Status $($res.status) for $($res.merchant_id)"
}

# 4. Immediate 11th request (Catches 429 Too Many Requests)
try {
    Invoke-RestMethod -Uri "http://localhost:8080/v1/payments" -Method Post -Body $body -ContentType "application/json" -Headers $headers
} catch {
    Write-Host "Throttled with HTTP $($_.Exception.Response.StatusCode.value__)"
    Write-Host "Response Payload: $($_.ErrorDetails.Message)"
}
```

---

## 5. Day 05: Database Schema Migration & Verification Runbook (Composite Indexes & JSONB)

### 5.1 Apply Migration 002 via Docker:
```bash
# Copy migration 002 to the running PostgreSQL container
docker cp migrations/002_fraud_and_features.sql novaris-postgres:/tmp/002_fraud_and_features.sql

# Execute the migration script inside the container
docker exec -it novaris-postgres psql -U novaris_admin -d novaris_db -f /tmp/002_fraud_and_features.sql

# Verify all 6 tables are registered in the public schema
docker exec -it novaris-postgres psql -U novaris_admin -d novaris_db -c "\dt"
```

### 5.2 Execute Deterministic Test & Diagnostic Harness:
```bash
# Copy verification harness into container
docker cp migrations/verify_day05_schema.sql novaris-postgres:/tmp/verify_day05_schema.sql

# Run diagnostic verification suite
docker exec -it novaris-postgres psql -U novaris_admin -d novaris_db -f /tmp/verify_day05_schema.sql
```

### 5.3 Windows PowerShell One-Liners:
```powershell
# 1. Apply Migration 002
docker cp migrations/002_fraud_and_features.sql novaris-postgres:/tmp/002_fraud_and_features.sql
docker exec -i novaris-postgres psql -U novaris_admin -d novaris_db -f /tmp/002_fraud_and_features.sql

# 2. Verify registered tables
docker exec -i novaris-postgres psql -U novaris_admin -d novaris_db -c "\dt"

# 3. Run verification harness
docker cp migrations/verify_day05_schema.sql novaris-postgres:/tmp/verify_day05_schema.sql
docker exec -i novaris-postgres psql -U novaris_admin -d novaris_db -f /tmp/verify_day05_schema.sql

# 4. Direct JSONB query probe
docker exec -i novaris-postgres psql -U novaris_admin -d novaris_db -c "SELECT id, feature_data->>'ip_country' AS country, (feature_data->>'velocity_5m')::int AS velocity FROM feature_snapshots WHERE feature_data @> '{\"is_vpn\": true}'::jsonb;"

# 5. Composite Index EXPLAIN plan probe (Confirms Index Scan)
docker exec -i novaris-postgres psql -U novaris_admin -d novaris_db -c "EXPLAIN (COSTS OFF) SELECT id, feature_data, computed_at FROM feature_snapshots WHERE account_id = 'c0000000-0000-0000-0000-000000000001' ORDER BY computed_at DESC LIMIT 10;"
```

### 5.4 Direct Host Connection (Alternative to Docker Exec):
```bash
# Connect directly to exposed host port 5432
PGPASSWORD='novaris_secure_password_2026' psql -h localhost -p 5432 -U novaris_admin -d novaris_db -f migrations/002_fraud_and_features.sql
PGPASSWORD='novaris_secure_password_2026' psql -h localhost -p 5432 -U novaris_admin -d novaris_db -f migrations/verify_day05_schema.sql
```

---

## 6. Day 06: Atomic Settlement & Double-Entry Ledger Verification Runbook

### 6.1 Run Automated Go Settlement Unit & Mock Tests:
```powershell
# Navigate to the API Gateway service directory
cd services\gateway

# Run all tests in the package with verbose output
go test -v ./...

# Run only settlement-specific tests (mock rollback and atomicity)
go test -v -run "TestSettleTransaction" ./...
```

### 6.2 Execute Deterministic SQL Verification Harness (Docker CLI):
```bash
# 1. Copy verification harness into the running PostgreSQL container
docker cp migrations/verify_day06_settlement.sql novaris-postgres:/tmp/verify_day06_settlement.sql

# 2. Execute verification script in psql
docker exec -it novaris-postgres psql -U novaris_admin -d novaris_db -f /tmp/verify_day06_settlement.sql
```

### 6.3 Windows PowerShell Alternative (Non-Interactive Pipe):
```powershell
docker cp migrations/verify_day06_settlement.sql novaris-postgres:/tmp/verify_day06_settlement.sql
docker exec -i novaris-postgres psql -U novaris_admin -d novaris_db -f /tmp/verify_day06_settlement.sql
```

### 6.4 Direct Ledger Balance Audit Query in PostgreSQL:
```sql
-- Audit all settled transactions to verify double-entry invariant (Sum(Debits) == Sum(Credits))
SELECT 
    transaction_id,
    SUM(CASE WHEN entry_type = 'DEBIT' THEN amount_cents ELSE 0 END) AS debits,
    SUM(CASE WHEN entry_type = 'CREDIT' THEN amount_cents ELSE 0 END) AS credits,
    CASE 
        WHEN SUM(CASE WHEN entry_type = 'DEBIT' THEN amount_cents ELSE 0 END) = 
             SUM(CASE WHEN entry_type = 'CREDIT' THEN amount_cents ELSE 0 END) 
        THEN 'BALANCED' 
        ELSE 'CORRUPTED' 
    END AS status
FROM journal_entries
GROUP BY transaction_id;
```

---

## 7. Day 07: Table-Driven Testing, Code Coverage & Benchmark Runbook

### 7.1 Run the Full Day 7 Table-Driven Unit Test Suite:
```powershell
# Navigate to the gateway service directory
cd services\gateway

# Run all 15 subtests with verbose per-test reporting
go test -v -run "TestSettleTransaction_TableDriven" ./...
```

### 7.2 Run High-Throughput In-Memory Settlement Benchmarks:
```powershell
# Profile settlement execution duration and memory allocations
go test -v -run "^$" -bench "BenchmarkSettleTransaction_Mock" -benchmem ./...
```

### 7.3 Generate Statement Code Coverage Metrics:
```powershell
# Generate coverage profile for the gateway package
go test -coverprofile=coverage.out .

# Print per-function statement coverage (asserting settleTransaction is 100%)
go tool cover -func=coverage.out | Select-String "settleTransaction"
```

### 7.4 Generate and View Interactive HTML Heatmap:
```powershell
# Render interactive HTML heatmap (green = covered, red = uncovered)
go tool cover -html=coverage.out -o coverage.html

# Open in default browser
Start-Process coverage.html
```

### 7.5 Run Static Diagnostics (Zero Warnings Verification):
```powershell
& "C:\Users\Tokachichu SriCharan\go\bin\gopls.exe" check settlement.go settlement_test.go main.go
```

---

## 8. Day 08: Concurrency Testing, Goroutine Collision & Race Detection Runbook

### 8.1 Run the Full Day 8 Concurrency Test Suite (50–100 Goroutines):
```powershell
# Navigate to the API gateway microservice directory
cd services\gateway

# Run only the Day 8 concurrency stress test suites with verbose output
go test -v -run "TestConcurrent" ./...
```

### 8.2 Run Concurrency Tests in High-Iteration Loops (Exposing Rare Timing Races):
```powershell
# Execute concurrency test suites 10 times consecutively
go test -v -count=10 -run "TestConcurrent" ./...
```

### 8.3 Run Full Gateway Test Suite (All 25 Tests in Parallel):
```powershell
# Run all unit, integration, and concurrency tests
go test -v ./...
```

### 8.4 Run Race Detector with ThreadSanitizer (Linux / macOS / WSL / CI):
```bash
# In Linux or GitHub Actions CI where 64-bit GCC is native:
go test -race -v -run "TestConcurrent" ./...

# Run concurrency tests 50 times under ThreadSanitizer
go test -race -count=50 -cpu=8 -run "TestConcurrent" ./...
```

### 8.5 Verify Zero Static Diagnostics Across All Gateway Code:
```powershell
& "C:\Users\Tokachichu SriCharan\go\bin\gopls.exe" check concurrency_test.go settlement.go settlement_test.go main.go
```

### 8.6 Benchmark In-Memory Concurrency & Allocation Efficiency:
```powershell
go test -bench=. -benchmem ./...
```

---

## 9. Day 09: Optimistic Concurrency Control (OCC) & Version Invariant Runbook

### 9.1 Run All 8 Automated Day 9 OCC Test Suites:
```powershell
# Navigate to the API Gateway directory
cd services\gateway

# Run only the OCC test suites with verbose per-test reporting
go test -v -run "TestOCC" ./...
```

### 9.2 Run OCC Tests Under Multi-Iteration Loop:
```powershell
# Verify that 50 concurrent goroutines consistently produce exactly 1 winner across multiple iterations
go test -v -count=10 -run "TestOCC" ./...
```

### 9.3 Run All 33 Gateway Unit, Integration, and Concurrency Tests:
```powershell
go test -v ./...
```

### 9.4 Verify Zero Static Diagnostics Across All Gateway Code:
```powershell
& "C:\Users\Tokachichu SriCharan\go\bin\gopls.exe" check occ.go occ_test.go settlement.go concurrency_test.go main.go
```

### 9.5 Execute Deterministic SQL OCC Verification Harness (Docker CLI):
```bash
# Copy verification harness into PostgreSQL container
docker cp migrations/verify_day09_occ.sql novaris-postgres:/tmp/verify_day09_occ.sql

# Execute in psql (asserting Session 1 succeeds and Session 2 stale write modifies 0 rows)
docker exec -it novaris-postgres psql -U novaris_admin -d novaris_db -f /tmp/verify_day09_occ.sql
```

### 9.6 Windows PowerShell Alternative (Non-Interactive Pipe):
```powershell
docker cp migrations/verify_day09_occ.sql novaris-postgres:/tmp/verify_day09_occ.sql
docker exec -i novaris-postgres psql -U novaris_admin -d novaris_db -f /tmp/verify_day09_occ.sql
```

---

## 10. Day 10: Idempotent API Design & Safe Network Retries Runbook

### 10.1 Run All 7 Day 10 Automated Idempotency Test Suites:
```powershell
# Navigate to the API Gateway directory
cd services\gateway

# Run only the Idempotency test suites with verbose per-test reporting
go test -v -run "TestIdempotency" ./...
```

### 10.2 Run 50-Worker Concurrent Idempotency Race Under Multi-Iteration Loop:
```powershell
# Verify that 50 concurrent goroutines submitting identical keys execute handler EXACTLY ONCE
go test -v -count=10 -run "TestIdempotency_50ConcurrentWorkers" ./...
```

### 10.3 Run All 40 Gateway Unit, Integration, and Concurrency Tests:
```powershell
go test -v ./...
```

### 10.4 Verify Zero Static Diagnostics Across All Gateway Code:
```powershell
& "C:\Users\Tokachichu SriCharan\go\bin\gopls.exe" check idempotency.go idempotency_test.go occ.go occ_test.go settlement.go concurrency_test.go main.go
```

### 10.5 Apply Day 10 Database DDL Migration (Docker CLI):
```bash
# Copy and execute migration 003 into PostgreSQL container
docker cp migrations/003_idempotency_keys.sql novaris-postgres:/tmp/003_idempotency_keys.sql
docker exec -it novaris-postgres psql -U novaris_admin -d novaris_db -f /tmp/003_idempotency_keys.sql
```

### 10.6 Execute Deterministic SQL Idempotency Verification Harness:
```bash
# Copy and execute verification harness (asserting primary key deduplication and zero double-spend)
docker cp migrations/verify_day10_idempotency.sql novaris-postgres:/tmp/verify_day10_idempotency.sql
docker exec -it novaris-postgres psql -U novaris_admin -d novaris_db -f /tmp/verify_day10_idempotency.sql
```

---

## 11. Day 11: Integration Wiring, API Contracts & E2E Verification Runbook

### 11.1 Run All 9 Day 11 Automated Integration Test Scenarios:
```powershell
# Navigate to the API Gateway directory
cd services\gateway

# Run only the End-to-End Integration test suites with verbose per-test reporting
go test -v -run "TestIntegration" ./...
```

### 11.2 Execute Multi-Merchant Concurrency Stress Test:
```powershell
# Verify that 25 simultaneous requests across 5 distinct merchants complete with zero data races
go test -v -run "TestIntegration_Scenario9_MultiMerchantConcurrentStress" ./...
```

### 11.3 Run Live End-to-End CLI Contract Probe (`cmd/e2eprobe`):
```powershell
# 1. In Terminal 1: Start the gateway service
cd services\gateway
go run main.go

# 2. In Terminal 2: Run the automated contract verification probe
cd services\gateway
go run cmd/e2eprobe/main.go
```

### 11.4 Execute Automated Postman Collection via Newman CLI:
```bash
# Execute full 10-scenario contract suite headlessly
newman run deployments/postman/NOVARIS_API_Contracts.postman_collection.json \
  -e deployments/postman/NOVARIS_Local_Dev.postman_environment.json \
  --reporters cli
```

### 11.5 Apply Day 11 Deterministic Seed Accounts Migration (Docker CLI):
```bash
# Seed deterministic test accounts (Alpha, Beta, Gamma) into PostgreSQL container
docker cp migrations/004_seed_test_accounts.sql novaris-postgres:/tmp/004_seed_test_accounts.sql
docker exec -it novaris-postgres psql -U novaris_admin -d novaris_db -f /tmp/004_seed_test_accounts.sql
```

### 11.6 Execute SQL Integration & Total Monetary Conservation Proof Harness:
```bash
# Verify balance conservation, atomic OCC updates, and outbox event staging
docker cp migrations/verify_day11_integration.sql novaris-postgres:/tmp/verify_day11_integration.sql
docker exec -it novaris-postgres psql -U novaris_admin -d novaris_db -f /tmp/verify_day11_integration.sql
```

---

## 12. Day 12 — OpenAPI 3.0 & Architecture Decision Records (ADRs)

### 12.1 Run OpenAPI Specification Validator CLI (`cmd/verifyopenapi`):
```powershell
# Navigate to the gateway directory
cd services\gateway

# Validate canonical OpenAPI JSON specification
go run ./cmd/verifyopenapi/main.go -spec ../../api/openapi.json

# Run unit tests for the validator
go test -v ./cmd/verifyopenapi/...
```

### 12.2 Run Gateway OpenAPI & Swagger UI Serving Integration Tests:
```powershell
# Run the OpenAPI router integration suite
go test -v -run "TestGateway_OpenAPI|TestGateway_Swagger|TestGateway_Root" .
```

### 12.3 Run Embedded OpenAPI Package Unit Tests:
```powershell
# Verify embedded //go:embed spec and swagger handlers
go test -v ./internal/openapi/...
```

### 12.4 Preview Interactive Swagger UI in Local Browser:
```powershell
# 1. Start the API Gateway
go run main.go idempotency.go occ.go settlement.go

# 2. Open interactive Swagger UI documentation in browser:
# http://localhost:8080/docs
# Or fetch raw OpenAPI JSON specification:
# http://localhost:8080/openapi.json
```

### 12.5 Preview Documentation with Redocly CLI & Spin Up Prism Mock Server:
```bash
# Preview via Redocly live development server
npx @redocly/cli preview-docs ../../api/openapi.yaml --port 8081

# Spin up a Prism mock server that validates against openapi.yaml
npx @stoplight/prism-cli mock ../../api/openapi.yaml -p 4010
```

### 12.6 Inspect Published Architecture Decision Records (ADRs):
```bash
# View Master ADR Repository Index
cat docs/adr/README.md

# Inspect ADR-0001 (OCC vs Pessimistic Row Locking)
cat docs/adr/0001-optimistic-concurrency-control-for-payment-settlement.md

# Inspect ADR-0002 (Five-Layer Pipeline Ordering)
cat docs/adr/0002-five-layer-defense-in-depth-gateway-pipeline.md
```

---

## 13. Day 13: Concurrency Load Scripting (k6 & Hot-Wallet Contention) Runbook

### 13.1 Run 100-Worker Simultaneous Concurrency Stress Test Suite:
```powershell
Set-Location "services/gateway"

# Run all 3 Day 13 100-worker stress tests (Hot-Wallet OCC, Pipeline Spike, Idempotency Race)
go test -v -run "TestLoadStress_" .

# Verify zero race conditions over multiple iterations
go test -v -count=5 -run "TestLoadStress_100ConcurrentWorkers_HotWalletOCC" .
```

### 13.2 Execute Native Go High-Throughput Load Testing CLI:
```powershell
Set-Location "services/gateway"

# 1. Hot-wallet contention test: 50 concurrent workers, 250 payment transactions
go run ./cmd/loadtest -workers 50 -requests 250 -hotwallet=true

# 2. Extreme 100-worker burst test: 100 concurrent workers hitting 1 wallet
go run ./cmd/loadtest -workers 100 -requests 1000 -hotwallet=true

# 3. Distributed multi-tenant test: 50 workers across 10 distinct merchants
go run ./cmd/loadtest -workers 50 -requests 500 -hotwallet=false -tenants 10

# 4. Run loadtest CLI unit test suite
go test -v ./cmd/loadtest/...
```

### 13.3 Execute Grafana k6 Load Scripts (When k6 is Installed or via Docker):
```powershell
# Run hot-wallet concurrency load test via local k6 CLI:
k6 run loadtests/k6_hotwallet_load.js

# Run distributed multi-merchant settlement load test:
k6 run loadtests/k6_concurrent_settlement.js

# Run k6 via Docker (Zero local installation):
docker run --rm -i -v "${PWD}/loadtests:/loadtests" --net=host grafana/k6 run /loadtests/k6_hotwallet_load.js
```

### 13.4 Run PostgreSQL Concurrency & Balance Conservation Audit:
```bash
# Copy Day 13 verification SQL into PostgreSQL container:
docker cp migrations/verify_day13_concurrency.sql novaris-postgres:/tmp/verify_day13_concurrency.sql

# Execute SQL audit (proves SUM(Debit) == SUM(Credit) and CAS version increments):
docker exec -it novaris-postgres psql -U novaris_admin -d novaris_db -f /tmp/verify_day13_concurrency.sql
```

### 13.5 Run Full Platform Test Suite Across All Subsystems:
```powershell
Set-Location "services/gateway"
go test -v ./...
```

---

## 14. Master Daily Git Multi-Commit & Push Workflow

To maintain a professional, atomic Git history and continuously build your public GitHub contribution profile:

```bash
# 1. Inspect changes
git status

# 2. Stage and commit atomically by concern:
git add <core-code-files>
git commit -m "feat(<subsystem>): <descriptive message>"

git add <test-files>
git commit -m "test(<subsystem>): <descriptive message>"

git add <cli-utility-files>
git commit -m "feat(cli): <descriptive message>"

git add <documentation-files>
git commit -m "docs(<subsystem>): <descriptive message>"

# 3. Push to remote main branch
git push origin main
```

---

<div align="center">

<!-- Bottom Banner -->
<p align="center">
  <img src="../../../assets/footer.svg" alt="NOVARIS Footer" width="100%" />
</p>

### 🎓 NOVARIS Phase 1 Runbook
*Deterministic Testing Across Phase 1 Subsystems*

</div>
