<div align="center">

<!-- Top Banner -->
<p align="center">
  <img src="../../assets/banner.svg" alt="NOVARIS Banner" width="100%" />
</p>

# 🧠 NOVARIS • Adaptive Fraud Intelligence Engine (`/services/fraud-ml`)
### Real-Time Streaming Feature Extraction, Vector Scoring & Human Review Triage

<p align="center">
  <img src="https://img.shields.io/badge/Phase_1_Capstone-COMPLETED_(100%25_Verified)-00C853?style=flat-square&logo=checkbox&logoColor=white" alt="Phase 1 Capstone" />
  <img src="https://img.shields.io/badge/Language-Python_3.11+-3776AB?style=flat-square&logo=python&logoColor=white" alt="Python" />
  <img src="https://img.shields.io/badge/Framework-FastStream_+_AsyncIO-00ADD8?style=flat-square&logo=fastapi&logoColor=white" alt="FastStream" />
  <img src="https://img.shields.io/badge/ML_Engine-XGBoost_•_LightGBM-FF6D00?style=flat-square&logo=scikitlearn&logoColor=white" alt="ML" />
  <img src="https://img.shields.io/badge/Inference_SLA-<15ms-00C853?style=flat-square&logo=speedtest&logoColor=white" alt="SLA" />
  <img src="https://img.shields.io/badge/Roadmap-Phase_3_(Weeks_7--8)-7C4DFF?style=flat-square&logo=target&logoColor=white" alt="Roadmap" />
</p>

---

</div>

## 📌 Service Overview

The **Fraud Intelligence Service** operates as a real-time predictive firewall embedded in the transaction stream. For every transaction ingested by the API Gateway, this service computes dynamic behavioral features, evaluates risk using gradient-boosted decision tree ensembles, and issues verdicts before settlement finality.

### Core Architectural Responsibilities:
1. **Sub-15ms Real-Time Inference Budget:**
   * High-speed vectorized scoring allowing transactions to be triaged without exceeding the platform's sub-40ms P99 latency SLA.
2. **Dynamic Sliding-Window Features:**
   * Aggregates rolling features from Redis and PostgreSQL: 1-hour transaction velocity, geo-velocity hops (impossible travel speed), and device fingerprint drift.
3. **Point-in-Time Feature Snapshots (Day 5 Schema):**
   * Persists extracted feature vectors into `feature_snapshots.feature_data` using PostgreSQL `JSONB` for offline model re-training and drift analysis.
4. **Human-in-the-Loop Review Triage:**
   * Ambiguous transactions (e.g. risk score 70–89) automatically route to `fraud_reviews` for compliance analyst evaluation, while scores $\ge 90$ trigger instantaneous automated declines.

---

## 🌊 Streaming Architecture

```mermaid
flowchart LR
    Kafka["Kafka Topic:<br/>tx.initiated"] -->|"FastStream Consumer"| FeatureStore["Sliding Window Aggregator<br/>(Redis 7)"]
    FeatureStore -->|"Feature Vector"| XGBoost["XGBoost Classifier<br/>Model (.onnx / .json)"]
    XGBoost --> Verdict{"Risk Score"}

    Verdict -->|"Score < 70 (Low)"| Allow["Approved<br/>(tx.fraud_cleared)"]
    Verdict -->|"70 <= Score < 90 (Medium)"| Review["Hold for Human Review<br/>(INSERT INTO fraud_reviews)"]
    Verdict -->|"Score >= 90 (High)"| Decline["Immediate Decline<br/>(tx.declined)"]
```

> [!NOTE]
> Database preparation for this service was implemented on **Day 5** via [`migrations/002_fraud_and_features.sql`](../../migrations/002_fraud_and_features.sql). The full Python inference service and model deployment pipeline will be constructed in **Phase 3 (Weeks 7–8)** of the NOVARIS 84-day master engineering roadmap.

---

<div align="center">

<!-- Bottom Banner -->
<p align="center">
  <img src="../../assets/footer.svg" alt="NOVARIS Footer" width="100%" />
</p>

### 🎓 NOVARIS Adaptive Fraud Intelligence
*Real-Time Predictive Risk Scoring & Compliance Automation*

</div>
