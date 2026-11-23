<div align="center">

<!-- Top Banner -->
<p align="center">
  <img src="../../assets/banner.svg" alt="NOVARIS Banner" width="100%" />
</p>

# 🚪 NOVARIS • API Gateway Service (`/services/gateway`)
### High-Throughput Request Ingestion, Rate Limiting & Cryptographic Verification

<p align="center">
  <img src="https://img.shields.io/badge/Language-Go_1.22+-00ADD8?style=flat-square&logo=go&logoColor=white" alt="Go" />
  <img src="https://img.shields.io/badge/Role-Edge_Ingestion-00C853?style=flat-square&logo=fastapi&logoColor=white" alt="Role" />
  <img src="https://img.shields.io/badge/Port-8080-7C4DFF?style=flat-square&logo=gnubash&logoColor=white" alt="Port" />
  <img src="https://img.shields.io/badge/Resilience-Graceful_Shutdown-FFA000?style=flat-square&logo=shield&logoColor=white" alt="Resilience" />
</p>

---

</div>

## 📌 Service Overview

The **API Gateway** is the single entry point for all external merchants, payment flows, and client applications interacting with NOVARIS. It sits on the network perimeter and guarantees that no invalid, duplicate, or tampered traffic touches downstream event streams or financial ledgers.

### Core Responsibilities:
* **Liveness & Readiness Probing:** Exposes standardized `/healthz` endpoints for Kubernetes container orchestration.
* **Structured Observability:** Middleware tracking request method, path, remote IP, and microsecond processing latency.
* **Graceful Termination:** Intercepts OS signals (`SIGINT`, `SIGTERM`) and allows in-flight connections to drain cleanly with a 5-second deadline.
* *(Upcoming Day 2–4):* HMAC-SHA256 digital signature validation, JWT bearer token claims verification, and distributed sliding token-bucket rate limiting via Redis.

---

## 📡 HTTP API Specification

### 1. `GET /healthz` — Service Health & Probing
Returns the operational health and environment metadata of the gateway.

* **Response (`200 OK`):**
```json
{
  "status": "ok",
  "service": "novaris-gateway",
  "version": "1.0.0",
  "environment": "development",
  "timestamp": "2026-11-23T11:26:00Z"
}
```

### 2. `GET /` — API Information
Provides service discovery details and listed endpoints.

* **Response (`200 OK`):**
```json
{
  "name": "NOVARIS API Gateway",
  "description": "High-Throughput Distributed Payment Settlement & Fraud Intelligence Ingestion Gateway",
  "status": "online",
  "version": "1.0.0",
  "endpoints": {
    "health_check": "/healthz",
    "transaction_flow": "/v1/transactions (Coming in Phase 1: Day 2-4)"
  }
}
```

---

## 🛠️ Local Development & Running

```bash
# Navigate to gateway directory
cd services/gateway

# Run directly using Go
go run main.go

# Verify using curl or browser
curl http://localhost:8080/healthz
```

---

<div align="center">

<p align="center">
  <img src="../../assets/footer.svg" alt="NOVARIS Footer" width="100%" />
</p>

</div>
