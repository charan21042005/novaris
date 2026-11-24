<div align="center">

<!-- Top Banner -->
<p align="center">
  <img src="../../assets/banner.svg" alt="NOVARIS Banner" width="100%" />
</p>

# ⚡ NOVARIS Master Verification & Testing Runbook
### Unified Copy-Paste Command Cheat Sheet Across All Daily Milestones (Days 1–4+)

<p align="center">
  <img src="https://img.shields.io/badge/Runbook-Master_Testing_Suite-00E5FF?style=flat-square&logo=gitbook&logoColor=white" alt="Runbook" />
  <img src="https://img.shields.io/badge/Days-01_to_04_Verified-00C853?style=flat-square&logo=googlecalendar&logoColor=white" alt="Days" />
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
5. [Master Daily Git Multi-Commit & Push Workflow](#5-master-daily-git-multi-commit--push-workflow)

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
$token = (go run cmd/gentoken/main.go -merchant merchant_stripe_101 -role merchant -mins 15 | Select-String -Pattern 'ey[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+').Matches.Value
$body = '{"amount_cents":5000,"recipient":"merchant_coffee"}'
$secret = "novaris_insecure_dev_hmac_secret_change_in_production_2026"
$ts = [DateTimeOffset]::UtcNow.ToUnixTimeSeconds().ToString()
$canonical = "$ts.$body"

$hmac = New-Object System.Security.Cryptography.HMACSHA256
$hmac.Key = [System.Text.Encoding]::UTF8.GetBytes($secret)
$hash = $hmac.ComputeHash([System.Text.Encoding]::UTF8.GetBytes($canonical))
$sig = [BitConverter]::ToString($hash).Replace("-","").ToLower()

$headers = @{
    "Authorization"       = "Bearer " + $token
    "X-Novaris-Timestamp" = $ts
    "X-Novaris-Signature" = $sig
}

# Send burst of 10 requests (All return 200 OK)
1..10 | ForEach-Object {
    $res = Invoke-RestMethod -Uri "http://localhost:8080/v1/payments" -Method Post -Body $body -ContentType "application/json" -Headers $headers
    Write-Host "Request $_: Success"
}

# 11th request triggers 429
try {
    Invoke-RestMethod -Uri "http://localhost:8080/v1/payments" -Method Post -Body $body -ContentType "application/json" -Headers $headers
} catch {
    Write-Host "11th Request Throttled: HTTP" $_.Exception.Response.StatusCode.value__
}
```

---

## 5. Day 05: Full Ledger Schema, Composite Indexes & JSONB Probes

### 5.1 Run Schema Verification Test Suite:
```bash
cd services/gateway

# Run automated schema test suite (offline AST & SQL checks)
go test -v ./cmd/verifyschema

# Run full gateway test suite (21/21 tests pass)
go test -v ./...
```

### 5.2 Execute Go Schema Verifier CLI:
```bash
cd services/gateway

# Mode 1: Offline static migration validation (Ideal for CI/CD)
go run cmd/verifyschema/main.go -offline

# Mode 2: Live PostgreSQL inspection against running container
go run cmd/verifyschema/main.go
```
*PowerShell Alternative:*
```powershell
Set-Location services\gateway
go test -v ./cmd/verifyschema
go run cmd/verifyschema/main.go -offline
```

### 5.3 Apply Migration 002 to Docker PostgreSQL:
```bash
# Copy migration file to container
docker cp migrations/002_fraud_and_features.sql novaris-postgres:/tmp/002.sql

# Execute SQL file against live database
docker exec -i novaris-postgres psql -U novaris_admin -d novaris_db -f /tmp/002.sql
```
*PowerShell Alternative:*
```powershell
docker cp migrations\002_fraud_and_features.sql novaris-postgres:/tmp/002.sql
docker exec -i novaris-postgres psql -U novaris_admin -d novaris_db -f /tmp/002.sql
```

### 5.4 Inspect Relations & Composite Index via `psql`:
```bash
# List all 6 tables (accounts, transactions, journal_entries, outbox, fraud_reviews, feature_snapshots)
docker exec -it novaris-postgres psql -U novaris_admin -d novaris_db -c "\dt"

# Inspect composite index on feature_snapshots (idx_feature_snapshots_account_time)
docker exec -it novaris-postgres psql -U novaris_admin -d novaris_db -c "\d+ feature_snapshots"

# Inspect CHECK constraint on fraud_reviews (CONFIRMED_FRAUD, FALSE_POSITIVE)
docker exec -it novaris-postgres psql -U novaris_admin -d novaris_db -c "\d+ fraud_reviews"
```

### 5.5 Test CHECK Constraint Enforcement:
```bash
# Attempt to insert an invalid verdict string -> Expect Error 23514 (check constraint violation)
docker exec -it novaris-postgres psql -U novaris_admin -d novaris_db -c "
INSERT INTO fraud_reviews (transaction_id, analyst_id, verdict, notes)
VALUES ('c0a80101-0000-0000-0000-000000000001', 'a0a80101-0000-0000-0000-000000000001', 'SUSPICIOUS_ONLY', 'Analyst flag');"
```

### 5.6 Test JSONB Containment & Extraction Queries:
```bash
# Insert a sample feature snapshot with nested JSONB signals
docker exec -it novaris-postgres psql -U novaris_admin -d novaris_db -c "
INSERT INTO feature_snapshots (account_id, feature_data)
VALUES ('a1111111-1111-1111-1111-111111111111', '{\"velocity_5m\": 14, \"geo_hop_km\": 1820.4, \"device_known\": false, \"vpn_detected\": true}');"

# Query 1: Extract velocity as text using ->> operator
docker exec -it novaris-postgres psql -U novaris_admin -d novaris_db -c "
SELECT id, account_id, feature_data->>'velocity_5m' AS velocity FROM feature_snapshots;"

# Query 2: Filter with JSONB containment operator (@>)
docker exec -it novaris-postgres psql -U novaris_admin -d novaris_db -c "
SELECT id, computed_at FROM feature_snapshots WHERE feature_data @> '{\"vpn_detected\": true}';"
```

### 5.7 Verify B-Tree Index Utilization with `EXPLAIN ANALYZE`:
```bash
docker exec -it novaris-postgres psql -U novaris_admin -d novaris_db -c "
EXPLAIN ANALYZE
SELECT * FROM feature_snapshots
WHERE account_id = 'a1111111-1111-1111-1111-111111111111'
ORDER BY computed_at DESC
LIMIT 5;"
```
*(Verify output confirms `Index Scan using idx_feature_snapshots_account_time` with zero in-memory sort).*

---

## 6. Master Daily Git Multi-Commit & Push Workflow
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

## 5. Master Daily Git Multi-Commit & Push Workflow

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
  <img src="../../assets/footer.svg" alt="NOVARIS Footer" width="100%" />
</p>

### 🎓 NOVARIS Master Runbook
*Deterministic Testing Across All Distributed Subsystems*

</div>
