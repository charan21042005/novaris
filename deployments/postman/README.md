<div align="center">

<!-- Top Banner -->
<p align="center">
  <img src="../../assets/banner.svg" alt="NOVARIS Banner" width="100%" />
</p>

# 📮 NOVARIS • Postman API Contracts & Integration Suite (`/deployments/postman`)
### Industry-Standard Automated API Contract Collections & Environment Definitions

<p align="center">
  <img src="https://img.shields.io/badge/Postman-v2.1.0_Format-FF6C37?style=flat-square&logo=postman&logoColor=white" alt="Postman" />
  <img src="https://img.shields.io/badge/Contracts-10_Scenarios_Verified-00C853?style=flat-square&logo=checkmarx&logoColor=white" alt="Scenarios" />
  <img src="https://img.shields.io/badge/Crypto_Pre--Request-CryptoJS_HMAC_+_JWT-7C4DFF?style=flat-square&logo=javascript&logoColor=white" alt="CryptoJS" />
  <img src="https://img.shields.io/badge/CLI_Runner-Newman_Ready-00E5FF?style=flat-square&logo=gnubash&logoColor=white" alt="Newman" />
</p>

---

</div>

## 📌 Directory Purpose & Industry Context (Day 11)

On **Day 11**, NOVARIS reaches a critical architectural milestone: **Integration Wiring**.
Up through Day 10, individual components (JWT validation, token-bucket rate limiting, HMAC digital signatures, optimistic concurrency control, and idempotency deduplication) were built and verified in unit isolation.

Day 11 connects these layers into a single, cohesive, live request path:
$$\text{Client HTTP} \longrightarrow \text{JWT Auth} \longrightarrow \text{Rate Limiter} \longrightarrow \text{HMAC Validator} \longrightarrow \text{Idempotency Engine} \longrightarrow \text{OCC Settlement} \longrightarrow \text{PostgreSQL Ledger}$$

To prove this entire chain operates cleanly under manual and automated testing without requiring custom test code, NOVARIS ships with a complete, production-grade **Postman Collection & Environment Suite**.

---

## 🗂️ Suite Files Inventory

| File | Type | Purpose |
| :--- | :---: | :--- |
| **[`NOVARIS_API_Contracts.postman_collection.json`](./NOVARIS_API_Contracts.postman_collection.json)** | Collection (v2.1) | 10 executable API contract requests with embedded JavaScript pre-request crypto signing and test assertion scripts. |
| **[`NOVARIS_Local_Dev.postman_environment.json`](./NOVARIS_Local_Dev.postman_environment.json)** | Environment | Parameterized variables: `base_url`, `merchant_id`, `recipient_id`, `jwt_secret`, and `hmac_secret`. |

---

## 🧪 Included Contract Test Scenarios

```
┌────┬──────────────────────────────────────────┬─────────────────┬──────────────────────────────────────────────┐
│ #  │ SCENARIO NAME                            │ EXPECTED STATUS │ ARCHITECTURAL GUARANTEE                      │
├────┼──────────────────────────────────────────┼─────────────────┼──────────────────────────────────────────────┤
│ 01 │ Public Liveness Probe                    │ 200 OK          │ Health probe responds without authentication.│
│ 02 │ Public Service Discovery Catalog         │ 200 OK          │ Exposes Day 11 versioning and capabilities.  │
│ 03 │ Missing Bearer Token Rejection           │ 401 Unauthorized│ Layer 1 Identity defense halts request.      │
│ 04 │ Missing HMAC Signature Rejection         │ 401 Unauthorized│ Layer 3 Integrity defense halts request.     │
│ 05 │ Missing Idempotency-Key Rejection        │ 400 Bad Request │ Layer 4 Deduplication mandates key header.   │
│ 06 │ Full Pipeline End-to-End Settlement      │ 200 OK          │ Real payment settles; transaction UUID given.│
│ 07 │ Exact-Once Network Retry Replay          │ 200 OK          │ Replays cached result; Idempotent-Replayed!  │
│ 08 │ Key Reuse Mismatched Payload Rejection   │ 422 Unproc. Ent.│ Altered body hash triggers security abort.   │
│ 09 │ Validation Defense: Self-Transfer Error  │ 400 Bad Request │ Disallows transfers to identical account ID. │
└────┴──────────────────────────────────────────┴─────────────────┴──────────────────────────────────────────────┘
```

---

## 🚀 How to Run the Postman Collection

### Option A: Using the Postman Desktop App (GUI)
1. Open the Postman Desktop application.
2. Click **Import** (top left).
3. Drag and drop both JSON files:
   * `NOVARIS_API_Contracts.postman_collection.json`
   * `NOVARIS_Local_Dev.postman_environment.json`
4. Select the environment **"NOVARIS Local Development Mesh"** in the top-right dropdown.
5. Click on the collection name and click **Run Collection**. All 10 tests will execute and turn green!

### Option B: Automated Headless CLI Execution via Newman
You can run the entire contract suite directly from your terminal using **Newman** (Postman's official CLI runner):

```bash
# 1. Install Newman globally (if not already installed)
npm install -g newman

# 2. Run the collection with the local environment
newman run deployments/postman/NOVARIS_API_Contracts.postman_collection.json \
  -e deployments/postman/NOVARIS_Local_Dev.postman_environment.json \
  --reporters cli
```

---

<div align="center">

<!-- Bottom Banner -->
<p align="center">
  <img src="../../assets/footer.svg" alt="NOVARIS Footer" width="100%" />
</p>

### 🎓 NOVARIS API Contracts & Quality Engineering
*Standardized Postman Collections • Continuous Integration Testing*

</div>
