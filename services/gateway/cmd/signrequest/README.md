<div align="center">

<!-- Top Banner -->
<p align="center">
  <img src="../../../../assets/banner.svg" alt="NOVARIS Banner" width="100%" />
</p>

# ✍️ NOVARIS HMAC Request Signer (`/cmd/signrequest`)
### Canonical Payload Constructor, RFC 2104 HMAC-SHA256 Signer & Replay Attack Simulator

<p align="center">
  <img src="https://img.shields.io/badge/Phase_1_Capstone-COMPLETED_(100%25_Verified)-00C853?style=flat-square&logo=checkbox&logoColor=white" alt="Phase 1 Capstone" />
  <img src="https://img.shields.io/badge/Language-Go_1.22+-00ADD8?style=flat-square&logo=go&logoColor=white" alt="Go" />
  <img src="https://img.shields.io/badge/Standard-RFC_2104_HMAC-7C4DFF?style=flat-square&logo=shield&logoColor=white" alt="HMAC" />
  <img src="https://img.shields.io/badge/Digest-SHA--256_(64_Hex)-00C853?style=flat-square&logo=openssl&logoColor=white" alt="SHA256" />
  <img src="https://img.shields.io/badge/Attack_Defense-Replay_%26_Tamper_Simulation-FFA000?style=flat-square&logo=hackthebox&logoColor=white" alt="Defense" />
</p>

---

</div>

## 📌 Tool Overview

`signrequest` is a developer and security validation utility introduced in **Day 3**. It simulates client payment SDKs by constructing canonical message representations and calculating hex-encoded HMAC-SHA256 digests for request bodies.

The utility is essential for:
1. **Generating Authentic Headers:** Emits `X-Novaris-Signature` and `X-Novaris-Timestamp` headers matching the gateway's strict canonicalization format.
2. **Replay Attack Simulation:** Applies deliberate positive or negative offsets (e.g. `-600` seconds) to simulate stale, intercepted network packets and verify gateway clock-skew rejection ($\pm 300\text{s}$ tolerance).
3. **In-Flight Tampering Simulation:** Computes a valid cryptographic signature over one payload (e.g. \$50.00) but deliberately outputs an altered payload (e.g. \$5,000.00) to confirm constant-time rejection.

```
       ┌─────────────────────────────────────────────────────────────┐
       │             Canonical Payload Construction                  │
       │                                                             │
       │  Timestamp: 1732617600                                      │
       │  Raw Body:  {"amount_cents":5000,"recipient":"acc_beta"}    │
       │                                                             │
       │  Canonical String:                                          │
       │  "1732617600.{\"amount_cents\":5000,\"recipient\":...}"     │
       └──────────────────────────────┬──────────────────────────────┘
                                      │
                                      ▼
       ┌─────────────────────────────────────────────────────────────┐
       │               RFC 2104 HMAC-SHA256 Digest                   │
       │  Key: HMAC_SECRET                                           │
       │  Digest: e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b...     │
       └──────────────────────────────┬──────────────────────────────┘
                                      │
                                      ▼
       ┌─────────────────────────────────────────────────────────────┐
       │                 Emitted HTTP Ingress Headers                │
       │  X-Novaris-Timestamp: 1732617600                            │
       │  X-Novaris-Signature: e3b0c44298fc1c149afbf4c8996fb9242... │
       └─────────────────────────────────────────────────────────────┘
```

---

## ⚙️ CLI Flags & Parameters

| Flag | Type | Default | Description |
| :--- | :---: | :---: | :--- |
| `-body` | `string` | `{"amount_cents":5000,"recipient":"merchant_456"}` | Raw JSON request payload to sign. |
| `-offset` | `int64` | `0` | Timestamp offset in seconds (e.g. `-600` simulates 10-minute old replay attack). |
| `-secret` | `string` | `""` | Custom HMAC secret key (defaults to `HMAC_SECRET` env var or dev key). |
| `-tamper` | `bool` | `false` | Computes signature on original body, but emits modified body to verify tamper rejection. |

---

## 🚀 Execution & Usage Examples

### 1. Sign Standard Settlement Payload
```powershell
Set-Location "services/gateway"
go run ./cmd/signrequest -body '{"amount_cents":10000,"recipient":"c0000000-0000-0000-0000-000000000002"}'
```

### 2. Simulate Stale Replay Attack (10 Minutes in Past)
```powershell
go run ./cmd/signrequest -offset -600
```

### 3. Simulate In-Flight Body Tampering
```powershell
go run ./cmd/signrequest -tamper
```

### Sample Output:
```text
================================================================================
   NOVARIS • HMAC-SHA256 REQUEST SIGNER                                        
================================================================================
Timestamp: 1732617600
Body:      {"amount_cents":10000,"recipient":"c0000000-0000-0000-0000-000000000002"}
Canonical: 1732617600.{"amount_cents":10000,"recipient":"c0000000-0000-0000-0000-000000000002"}
Signature: a9f81d4b68e914041b65e9982463e63db76f272a27ff6b8296eb253b21045239

CURL TEST COMMAND:
curl -X POST http://localhost:8080/v1/payments \
  -H "Content-Type: application/json" \
  -H "X-Novaris-Timestamp: 1732617600" \
  -H "X-Novaris-Signature: a9f81d4b68e914041b65e9982463e63db76f272a27ff6b8296eb253b21045239" \
  -d '{"amount_cents":10000,"recipient":"c0000000-0000-0000-0000-000000000002"}'
```

---

## 🔗 Related Components

* 🛡️ **[internal/auth](../../internal/auth/README.md):** `HMACMiddleware` that verifies signatures generated by this tool.
* 🔑 **[cmd/gentoken](../gentoken/README.md):** JWT minting utility to generate matching Bearer tokens.
* 🔍 **[cmd/e2eprobe](../e2eprobe/README.md):** Probe runner incorporating request signing logic.

---

<div align="center">

<!-- Bottom Banner -->
<p align="center">
  <img src="../../../../assets/footer.svg" alt="NOVARIS Footer" width="100%" />
</p>

### 🎓 NOVARIS API Gateway Developer Tooling
*RFC 2104 HMAC-SHA256 Request Signing Utility*

</div>
