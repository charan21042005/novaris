<div align="center">

<!-- Top Banner -->
<p align="center">
  <img src="../assets/banner.svg" alt="NOVARIS Banner" width="100%" />
</p>

# 🐳 NOVARIS • Local Development Infrastructure Mesh
### Containerized Storage, In-Memory Caching & Distributed Streaming

<p align="center">
  <img src="https://img.shields.io/badge/Phase_1_Capstone-COMPLETED_(100%25_Verified)-00C853?style=flat-square&logo=checkbox&logoColor=white" alt="Phase 1 Capstone" />
  <img src="https://img.shields.io/badge/Docker-24%2B-2496ED?style=flat-square&logo=docker&logoColor=white" alt="Docker" />
  <img src="https://img.shields.io/badge/PostgreSQL-16_Alpine-4169E1?style=flat-square&logo=postgresql&logoColor=white" alt="PostgreSQL" />
  <img src="https://img.shields.io/badge/Redis-7_Alpine-DC382D?style=flat-square&logo=redis&logoColor=white" alt="Redis" />
  <img src="https://img.shields.io/badge/Apache_Kafka-KRaft_7.5.0-231F20?style=flat-square&logo=apachekafka&logoColor=white" alt="Kafka" />
</p>

---

</div>

## 📌 Architecture & Overview

This directory contains the Docker Compose environment for local development and integration testing. Rather than installing disparate databases on the host operating system, NOVARIS orchestrates three core infrastructure dependencies in isolated, reproducible Docker containers:

1. **PostgreSQL 16 Alpine (`novaris-postgres`):**
   * Primary relational store for user accounts, transactions, double-entry journal entries, and outbox records.
   * Automatically executes SQL scripts in `migrations/` during initial container boot.
2. **Redis 7 Alpine (`novaris-redis`):**
   * High-speed in-memory store configured with AOF (`--appendonly yes`) for distributed idempotency locks (`SETNX`), rate limiting, and Redlock mutex coordination.
3. **Apache Kafka KRaft (`novaris-kafka`):**
   * High-throughput event streaming backbone running in ZooKeeper-less KRaft mode, partitioned by `account_id` to guarantee per-account event ordering.

---

## 🔌 Port Mapping & Credentials

| Service | Container Name | Host Port | Internal Port | Default Credentials | Purpose |
| :--- | :--- | :---: | :---: | :--- | :--- |
| **PostgreSQL** | `novaris-postgres` | `5432` | `5432` | `novaris_admin` / `dev_password_change_me` (DB: `novaris_db`) | Financial ledger & outbox persistence |
| **Redis** | `novaris-redis` | `6379` | `6379` | *None (Development Mode)* | Distributed locking & rate limiting |
| **Apache Kafka** | `novaris-kafka` | `9092` | `9092` | *PLAINTEXT (Unauthenticated dev)* | Event bus (`tx.initiated`, `tx.settled`) |

---

## 🚀 Quick Commands

```bash
# 1. Start all infrastructure containers in detached mode
docker compose -f deployments/docker-compose.yml up -d

# 2. View streaming logs across all containers
docker compose -f deployments/docker-compose.yml logs -f

# 3. Check health status of all containers
docker compose -f deployments/docker-compose.yml ps

# 4. Connect to PostgreSQL via psql inside the container
docker exec -it novaris-postgres psql -U novaris_admin -d novaris_db

# 5. Connect to Redis CLI
docker exec -it novaris-redis redis-cli ping

# 6. Stop all containers (preserving persistent volume data)
docker compose -f deployments/docker-compose.yml down

# 7. Complete reset (wipes all database & Kafka volumes)
docker compose -f deployments/docker-compose.yml down -v
```

---

<div align="center">

<p align="center">
  <img src="../assets/footer.svg" alt="NOVARIS Footer" width="100%" />
</p>

</div>
