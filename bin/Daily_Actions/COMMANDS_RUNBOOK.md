<div align="center">

<!-- Top Banner -->
<p align="center">
  <img src="../../assets/banner.svg" alt="NOVARIS Banner" width="100%" />
</p>

# ⚡ NOVARIS Master Verification & Testing Runbook
### Unified Copy-Paste Command Cheat Sheet Across All Daily Milestones (Days 1–3+)

<p align="center">
  <img src="https://img.shields.io/badge/Runbook-Master_Testing_Suite-00E5FF?style=flat-square&logo=gitbook&logoColor=white" alt="Runbook" />
  <img src="https://img.shields.io/badge/Days-01_to_03_Verified-00C853?style=flat-square&logo=googlecalendar&logoColor=white" alt="Days" />
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
4. [Master Daily Git Multi-Commit & Push Workflow](#4-master-daily-git-multi-commit--push-workflow)

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

## 4. Master Daily Git Multi-Commit & Push Workflow

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
