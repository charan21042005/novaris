<div align="center">

<!-- Top Banner -->
<p align="center">
  <img src="../assets/banner.svg" alt="NOVARIS Banner" width="100%" />
</p>

# 🖥️ NOVARIS • Executive & Fraud Analyst Portal (`/frontend`)
### Real-Time Payment Telemetry, Live Fraud Alert Workbench & Human Review Console

<p align="center">
  <img src="https://img.shields.io/badge/Framework-Next.js_14_•_React-000000?style=flat-square&logo=nextdotjs&logoColor=white" alt="Next.js" />
  <img src="https://img.shields.io/badge/Styling-Tailwind_CSS_•_Shadcn_UI-06B6D4?style=flat-square&logo=tailwindcss&logoColor=white" alt="Tailwind" />
  <img src="https://img.shields.io/badge/Streaming-WebSockets_•_SSE-00C853?style=flat-square&logo=socketdotio&logoColor=white" alt="WebSockets" />
  <img src="https://img.shields.io/badge/Roadmap-Phase_5_(Weeks_10--11)-7C4DFF?style=flat-square&logo=target&logoColor=white" alt="Roadmap" />
</p>

---

</div>

## 📌 Portal Overview

The **`/frontend`** workspace provides the operator interface for fraud analysts, compliance officers, and executive engineering teams. It transforms complex high-throughput streaming ledger events into clear, actionable visual dashboards.

### Planned Core Modules (Phase 5):
1. **Live Transaction Telemetry Feed:**
   * Real-time WebSocket stream displaying payment flows, volume metrics, and global throughput rates.
2. **Fraud Analyst Workbench (`fraud_reviews`):**
   * Interactive review queue allowing human risk investigators to inspect transactions flagged by the ML engine.
   * Action buttons to emit `CONFIRMED_FRAUD` or `FALSE_POSITIVE` verdicts which feed back into model retraining.
3. **Ledger Audit & Account Explorer:**
   * Double-entry journal inspector verifying zero financial drift across accounts with interactive debit/credit reconciliation.
4. **System Health & Distributed Tracing:**
   * Real-time monitoring of gateway pod latencies, Redis rate limiter consumption, and Kafka consumer group lag.

---

## 🎨 Visual Preview

```
┌────────────────────────────────────────────────────────────────────────┐
│  NOVARIS Real-Time Platform Status              [Online: 3,520 TPS]   │
├────────────────────────────────────────────────────────────────────────┤
│  Total Settled Volume: $14,892,100.00      P99 Gateway Latency: 22ms   │
├────────────────────────────────────────────────────────────────────────┤
│  ⚡ Active Fraud Review Queue (Requires Attention):                     │
│  [TX-91024] $4,200.00 • Account: acc_8812 • Score: 84 (High Velocity) │
│  Action: [Confirm Fraud]  [Mark False Positive]  [Inspect Features]    │
└────────────────────────────────────────────────────────────────────────┘
```

> [!NOTE]
> Scheduled for full end-to-end development in **Phase 5 (Weeks 10–11)** after the completion of the streaming Kafka backbone (Phase 2), ML inference pipeline (Phase 3), and distributed tracing observability mesh (Phase 4).

---

<div align="center">

<!-- Bottom Banner -->
<p align="center">
  <img src="../assets/footer.svg" alt="NOVARIS Footer" width="100%" />
</p>

### 🎓 NOVARIS Frontend Operations Console
*Next.js Real-Time Settlement & Fraud Management System*

</div>
