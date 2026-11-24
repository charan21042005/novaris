<div align="center">

<!-- Floating Top Banner -->
<p align="center">
  <img src="../../assets/banner.svg" alt="NOVARIS Banner" width="100%" />
</p>

# 📖 NOVARIS THEORY TEXTBOOK • DAY 05
## Full Ledger Schema: JSONB Documents, Composite Indexing & Ground-Truth Machine Learning Stores

<p align="center">
  <img src="https://img.shields.io/badge/Volume-01_Foundations_%26_Ledgers-00E5FF?style=flat-square&logo=gitbook&logoColor=white" alt="Volume" />
  <img src="https://img.shields.io/badge/Day-05_of_84-00C853?style=flat-square&logo=googlecalendar&logoColor=white" alt="Day" />
  <img src="https://img.shields.io/badge/Level-First_Principles_to_Production-7C4DFF?style=flat-square&logo=academia&logoColor=white" alt="Level" />
  <img src="https://img.shields.io/badge/Standard-Industrial_FinTech_Grade-FFA000?style=flat-square&logo=standards&logoColor=white" alt="Standard" />
</p>

---

</div>

## 📑 Detailed Table of Contents

1. [Preface: Finishing the Foundation — The Bridge from Ingress to Intelligence](#1-preface-finishing-the-foundation--the-bridge-from-ingress-to-intelligence)
2. [Chapter 0: The Plain-English Database & Schema Engineering Lexicon](#2-chapter-0-the-plain-english-database--schema-engineering-lexicon)
   - 0.1 What is a Database Schema? (The Architectural Blueprint Analogy)
   - 0.2 Relational Foreign Keys & Referential Integrity (The Parent-Child String Analogy)
   - 0.3 Database Migrations & The Immutable Commit History Principle
   - 0.4 Migration Discipline: Why We Never Edit Past Migrations
   - 0.5 JSON vs. JSONB in PostgreSQL (Raw Text vs. Decomposed Binary Byte-Code)
   - 0.6 What is a B-Tree Index? (The Balanced Library Catalog Analogy)
   - 0.7 What is a Composite (Multi-Column) Index? (The Phonebook Analogy)
   - 0.8 Column Ordering in Composite Indexes: The March Birthday Paradox
   - 0.9 CHECK Constraints: Database-Enforced Business Invariants
   - 0.10 PostgreSQL ENUM vs. VARCHAR + CHECK (The Extensibility Trade-Off)
   - 0.11 `EXPLAIN ANALYZE`: The X-Ray of Database Query Execution
   - 0.12 Sequential Scan (`Seq Scan`) vs. Index Scan (`Index Scan`)
   - 0.13 Ground-Truth Labels in Machine Learning (The Teacher's Answer Key)
   - 0.14 Ephemeral vs. Persistent Features (The Speedometer vs. The Flight Black Box)
   - 0.15 Schema Drift: The Danger of Divergent Environments
3. [Chapter 0.5: Complete Software Components, Tooling & Dependency Directory](#3-chapter-05-complete-software-components-tooling--dependency-directory)
   - 0.5.1 Relational Engine & Database Types (PostgreSQL 16, UUID, BIGINT, JSONB)
   - 0.5.2 Indexing & Query Optimization Primitives (B-Tree, Composite, GIN, EXPLAIN)
   - 0.5.3 Go Database Drivers & Schema Verification Stack (`database/sql`, `lib/pq`)
   - 0.5.4 Standards & Specifications (SQL:2016, RFC 8259 JSON, ISO 8601 Timestamps)
4. [Chapter 1: The Theoretical Foundations of Relational Ledger Systems](#4-chapter-1-the-theoretical-foundations-of-relational-ledger-systems)
   - 1.1 The Missing Pieces: Why Days 1–4 Left Two Tables Open
   - 1.2 The Human-in-the-Loop Triage Architecture (The 30–70 Score Band)
   - 1.3 The Temporal Decay of In-Memory State: Why Redis Needs an RDBMS Archive
   - 1.4 How Real-World FinTech Giants Model Fraud Data (Stripe Radar, Adyen Protect)
5. [Chapter 2: The Structural & Mathematical Architecture of JSONB and Composite Indexes](#5-chapter-2-the-structural--mathematical-architecture-of-jsonb-and-composite-indexes)
   - 2.1 The Mathematics of B-Trees: Depth $O(\log N)$ and Hierarchical Sorting
   - 2.2 Hierarchical Column Ordering in Composite B-Trees: Why `(account_id, computed_at DESC)` Succeeds
   - 2.3 The Phonebook Proof: Why `(computed_at, account_id)` Fails for Account History
   - 2.4 Internal Storage Layout of JSONB: Header, Offset Tables, and Key-Value Pointers
   - 2.5 JSONB Containment Mathematics: The Sub-Document Inclusion Operator (`@>`)
6. [Chapter 3: Critical Database Engineering Gotchas & Schema Discipline](#6-chapter-3-critical-database-engineering-gotchas--schema-discipline)
   - 3.1 Migration Discipline: Why Editing `001_initial_ledger_schema.sql` Causes Fatal Drift
   - 3.2 The `docker-entrypoint-initdb.d` Trap: Why New Migrations Don't Run on Existing Volumes
   - 3.3 The ENUM vs. VARCHAR+CHECK Dilemma: Why Altering Enums Requires Table Locks
   - 3.4 Nullable vs. Required Foreign Keys: Why `feature_snapshots.transaction_id` is Nullable
   - 3.5 Cascading Deletions: Why `ON DELETE RESTRICT` Protects Financial Auditability
7. [Chapter 4: The Closed MLOps Feedback Loop: Bridging Relational DBs to AI](#7-chapter-4-the-closed-mlops-feedback-loop-bridging-relational-dbs-to-ai)
   - 4.1 From Analyst Click to Retrained Weights: The Full Data Lifecycle
   - 4.2 How Day 66 Nightly Retraining DAG Queries `fraud_reviews`
   - 4.3 How Day 44–51 Feature Pipelines Join `feature_snapshots` with Labeled Verdicts
   - 4.4 Data Architecture Symmetry Across the 84-Day Roadmap
8. [Chapter 5: The Data Science → Full-Stack Schema Rosetta Stone](#8-chapter-5-the-data-science--full-stack-schema-rosetta-stone)
9. [Chapter 6: Architectural Trade-Off Analysis ("Why This and Not That")](#9-chapter-6-architectural-trade-off-analysis-why-this-and-not-that)
   - 6.1 JSONB Column vs. Fully Normalized Relational Table
   - 6.2 VARCHAR + CHECK vs. Postgres Native ENUM
   - 6.3 Composite Index `(account_id, computed_at)` vs. Two Separate Single-Column Indexes
   - 6.4 Client-Side Validation vs. Database CHECK Constraints (Defense in Depth)
10. [Chapter 7: Line-by-Line Code Masterclass (Complete Dissections)](#10-chapter-7-line-by-line-code-masterclass-complete-dissections)
    - 7.1 `migrations/002_fraud_and_features.sql` (The Complete Day 5 Migration)
    - 7.2 `services/gateway/cmd/verifyschema/main.go` (The Go Schema Verifier CLI)
    - 7.3 `services/gateway/cmd/verifyschema/main_test.go` (Automated Schema Test Suite)
    - 7.4 `services/gateway/cmd/README.md` (Updated CLI Documentation)
11. [Chapter 8: Verification Laboratory: Execution Proof & Complete Runbook](#11-chapter-8-verification-laboratory-execution-proof--complete-runbook)
    - 8.1 Complete Automated Test Suite Output (21/21 Passing Tests across Gateway)
    - 8.2 Live Docker & PostgreSQL Command Matrix (Applying Migrations, Inserts & Index Scans)
    - 8.3 Complete Step-by-Step CLI Runbook (Bash & PowerShell)
12. [Chapter 9: The Interview Defense Compendium (10 High-Stakes Tier-1 Q&As)](#12-chapter-9-the-interview-defense-compendium-10-high-stakes-tier-1-qas)
13. [Chapter 10: The Roadmap Horizon (Day 6 Preview: Atomic Double-Entry Settlement)](#13-chapter-10-the-roadmap-horizon-day-6-preview-atomic-double-entry-settlement)
    - 10.1 Day 6 Architectural Objectives
    - 10.2 Hands-On Mastery Exercises for the Reader

---

## 1. Preface: Finishing the Foundation — The Bridge from Ingress to Intelligence

Over the first four days of NOVARIS, we constructed an unassailable perimeter defense network in front of our Go API Gateway:
* **Day 2:** Cryptographic JWT authentication (Identity verification).
* **Day 3:** HMAC-SHA256 request body signing (Payload integrity & replay defense).
* **Day 4:** In-memory Token-Bucket rate limiting (Velocity & flood policing).

Today on **Day 5**, we deliberately step back from the gateway perimeter to complete the relational database foundation underneath our entire platform.

```
       WHAT WAS BUILT ON DAY 1:                     WHAT WAS DELIBERATELY DEFERRED UNTIL TODAY:
       
┌──────────────────────────────────────┐     ┌────────────────────────────────────────────────────────┐
│ 1. accounts                          │     │ 5. fraud_reviews                                       │
│    (Wallets, Balances, Currencies)   │     │    * Captures human analyst decisions on flagged txs   │
│ 2. transactions                      │     │    * Provides verified labels for ML retraining models │
│    (Payment Lifecycle, Risk Scores)  │     │ 6. feature_snapshots                                   │
│ 3. journal_entries                   │     │    * Point-in-time archive of ephemeral Redis features │
│    (Immutable Double-Entry Ledger)   │     │    * JSONB flexible storage with composite B-Tree      │
│ 4. outbox                            │     │      indexing on (account_id, computed_at DESC)        │
│    (Transactional Kafka Event Relay) │     └────────────────────────────────────────────────────────┘
└──────────────────────────────────────┘
```

### The Concrete Payoff: Why These Two Tables Exist
To a junior engineer, creating database tables for features that won't be actively written until Week 7 or Week 9 might feel like premature optimization. In distributed systems engineering, it is **load-bearing architectural planning**:

1. **`fraud_reviews` is a Machine Learning Dataset in Disguise:**
   When an incoming payment is assigned a risk score between $30$ and $70$ by our machine learning model, it is neither auto-approved nor auto-declined; it is routed to a human fraud analyst. When that human analyst inspects the transaction and clicks *"Confirmed Fraud"* or *"False Positive"*, where does that decision go?
   It is written to `fraud_reviews`.  
   **Crucially, on Day 66, when Apache Airflow triggers our automated nightly model retraining DAG, it will query this exact table.** Without `fraud_reviews`, our AI engine could never learn from human feedback!

2. **`feature_snapshots` Freezes Ephemeral Time:**
   On Day 20, we will build real-time velocity counters in Redis. Redis is blazingly fast, but it is **ephemeral and state-destructive**: Redis can tell you a merchant's transaction count in the *current* 5-minute window; it cannot tell you what that velocity was three weeks ago at 2:14 PM when a disputed charge occurred!
   `feature_snapshots` provides an immutable, point-in-time audit log of the exact feature values present when a decision was made.

### The Physical Analogy: The Medical Chart & The Biopsy Archive
* **`transactions` and `journal_entries`** are like an emergency room admission log. It records who walked in, what time, and what treatment was billed.
* **`feature_snapshots`** is the patient's **vital signs monitor printout** taken at the exact second of cardiac distress (heart rate, blood pressure, oxygen saturation).
* **`fraud_reviews`** is the **pathology biopsy report** signed by the chief surgeon confirming whether the tumor was malignant (Fraud) or benign (False Positive).

---

## 2. Chapter 0: The Plain-English Database & Schema Engineering Lexicon

To ensure engineers from any background (software engineering, computer science, or data science) master these concepts, here is every term explained from first principles:

### 1. What is a Database Schema?
* **Plain English:** The formal blueprint of a relational database. It dictates what tables exist, what columns they contain, what data types are permitted (e.g., integers, text, timestamps), and how tables relate to one another.
* **The Architectural Blueprint Analogy:** Before constructing a skyscraper, architects draft blueprints showing where load-bearing pillars, electrical conduits, and plumbing pipes must run. You cannot pour concrete without a blueprint.

### 2. Relational Foreign Keys & Referential Integrity
* **Plain English:** A constraint that links a column in one table to the primary key of another table. It guarantees that child records cannot reference non-existent parent records.
* **The Parent-Child String Analogy:** Imagine tying a child's wrist to their parent's wrist with a string. A child record (`fraud_reviews.transaction_id`) cannot exist in the playground unless the parent record (`transactions.id`) exists first!
* **`ON DELETE RESTRICT`:** If someone tries to delete the parent transaction, the database refuses, protecting financial audit trails from accidental destruction.

### 3. Database Migrations & The Immutable Commit History Principle
* **Plain English:** Version-controlled scripts (usually numbered SQL files like `001_initial.sql`, `002_features.sql`) that describe incremental changes to the database structure over time.
* **The Git Analogy:** Just as Git tracks changes to code via a chain of immutable commits, migrations track changes to database structure via a chain of immutable scripts.

### 4. Migration Discipline: Why We Never Edit Past Migrations
* **Plain English:** Once a migration script (like `001_initial_ledger_schema.sql`) has been executed on any running environment (development, staging, or production), **it must never be altered**.
* **Why:** If Developer Alice modifies `001` on her machine, but the production database already executed the original `001` three weeks ago, the production database and Alice's database will silently diverge! Any new tables must be added in a brand-new file: `002_fraud_and_features.sql`.

### 5. JSON vs. JSONB in PostgreSQL
* **Plain Text JSON (`json`):** Stores an exact text replica of the JSON string, including whitespace and duplicate keys. Every time you query a field, PostgreSQL must re-parse the raw text from scratch.
* **Binary JSON (`jsonb`):** Decomposes the JSON string into an optimized, pre-parsed binary format with internal lookup tables. It discards whitespace and normalizes keys.
* **Why JSONB Wins:** Querying a nested key inside JSONB is up to **100 times faster** than plain JSON, and JSONB supports indexing via GIN (Generalized Inverted Indexes).

### 6. What is a B-Tree Index?
* **Plain English:** A balanced tree data structure that PostgreSQL builds on disk to allow searches, range queries, and sorted order lookups in logarithmic time $O(\log N)$ instead of scanning every single row in the table.
* **The Library Catalog Analogy:** Imagine walking into a library with 10,000,000 books. Without an index, finding "War and Peace" requires walking down every aisle, looking at every book one by one (Sequential Scan). With a card catalog organized alphabetically (B-Tree Index), you can find the book in 5 steps!

### 7. What is a Composite (Multi-Column) Index?
* **Plain English:** A single B-Tree index built on **two or more columns** simultaneously (e.g., `ON feature_snapshots(account_id, computed_at DESC)`).
* **The Phonebook Analogy:** A printed telephone book is a composite index on `(Last Name, First Name)`. It sorts entries first by Last Name, and then by First Name within each last name group:
  ```
  Smith, Alice
  Smith, Bob
  Smith, Charlie
  Williams, Alice
  ```

### 8. Column Ordering in Composite Indexes: The March Birthday Paradox
* In a composite index on `(Column A, Column B)`, the index can accelerate:
  1. Queries filtering on `Column A` alone.
  2. Queries filtering on `Column A` AND `Column B` together.
* **It CANNOT accelerate queries filtering on `Column B` alone!**
* **The Proof:** If a phonebook is sorted by `(Last Name, First Name)`, can you quickly find everyone whose first name is "Alice"? No! "Alice" is scattered across every single page under Adams, Baker, Clark, Smith, and Williams. You would have to read the entire phonebook from cover to cover!

### 9. CHECK Constraints: Database-Enforced Business Invariants
* **Plain English:** A database rule that validates data before allowing an `INSERT` or `UPDATE`. If the data violates the rule, PostgreSQL aborts the write with an error.
* **The Security Turnstile Analogy:** An office turnstile that only allows employees with blue badges. If someone with an orange badge swipes, the turnstile physically locks.
* **In NOVARIS:** We enforce `CHECK (verdict IN ('CONFIRMED_FRAUD', 'FALSE_POSITIVE'))`. Even if an engineer writes a buggy Go script that attempts to insert `"MAYBE"`, the database itself refuses the write.

### 10. PostgreSQL ENUM vs. VARCHAR + CHECK
* **Native ENUM (`CREATE TYPE verdict_enum AS ENUM (...)`):** Creates a static type in PostgreSQL. Highly compact, but altering it later (e.g., adding a third verdict) requires `ALTER TYPE`, which can trigger table locks in older PostgreSQL versions or break ORMs.
* **VARCHAR(16) + CHECK:** Standard text column guarded by a boolean check constraint. Extremely easy to evolve in future migrations by simply replacing the check constraint. NOVARIS standardizes on VARCHAR + CHECK for maximum architectural agility.

### 11. `EXPLAIN ANALYZE`: The X-Ray of Database Query Execution
* **Plain English:** A PostgreSQL command that takes an SQL query, executes it, and outputs a detailed execution plan showing the exact physical algorithms the database engine used to find the data, how much memory was spent, and how many milliseconds each step took.

### 12. Sequential Scan (`Seq Scan`) vs. Index Scan (`Index Scan`)
* **Sequential Scan (`Seq Scan`):** The database reads every single disk block of the table from start to finish. Efficient for tables with 10 rows; catastrophically slow for tables with 50,000,000 rows.
* **Index Scan (`Index Scan`):** The database navigates the B-Tree index to locate only the exact matching row pointers, fetching only the required disk blocks. Sub-millisecond execution.

### 13. Ground-Truth Labels in Machine Learning
* **Plain English:** In supervised machine learning, "ground truth" represents the verified, objective reality (the true target label $y$) that a model is trained to predict.
* **In NOVARIS:** A model's initial prediction is just a hypothesis (e.g., "78% probability of fraud"). The human analyst's verdict in `fraud_reviews` is the verified **ground truth** used to compute loss functions and update neural weights during retraining.

### 14. Ephemeral vs. Persistent Features
* **Ephemeral Features:** Transient values computed in fast RAM (e.g., Redis counters) that represent the instantaneous present.
* **Persistent Features:** Point-in-time immutable records archived in PostgreSQL JSONB that preserve the exact historical context forever.

### 15. Schema Drift: The Danger of Divergent Environments
* **Plain English:** A critical operational defect where the database schema deployed in production differs from the schema defined in source code or local development.
* **Our Defense:** Numbered SQL migration files and automated Go schema verification CLI tools (`cmd/verifyschema`) executed in CI pipelines.

---

## 3. Chapter 0.5: Complete Software Components, Tooling & Dependency Directory

Every software package, database engine, and technical standard utilized in Day 5 is detailed below:

### 0.5.1 Relational Engine & Database Types

| Component / Type | Engine / Source | Architectural Role in NOVARIS |
| :--- | :--- | :--- |
| **PostgreSQL 16** | Relational RDBMS (Alpine) | Primary ACID storage engine for accounts, transactions, ledgers, outbox, and audit logs. |
| **`UUID` (`gen_random_uuid()`)**| PostgreSQL Built-in | Universally unique, unguessable 128-bit identifiers preventing ID enumeration attacks. |
| **`JSONB`** | PostgreSQL Binary JSON | Decomposed binary JSON storage format enabling fast key lookups and flexible feature vectors. |
| **`TIMESTAMPTZ`** | PostgreSQL Timestamp | UTC timezone-aware microsecond timestamps preventing multi-region clock ambiguity. |
| **`VARCHAR(16)` + `CHECK`**| ANSI SQL Standard | Extensible string type guarded by database-enforced enumeration constraints. |

### 0.5.2 Indexing & Query Optimization Primitives

| Primitive | Mechanism | Architectural Role in NOVARIS |
| :--- | :--- | :--- |
| **Composite B-Tree** | Multi-column B-Tree Index | Hierarchically indexes `(account_id, computed_at DESC)` for sub-millisecond account histories. |
| **Partial Index** | `WHERE transaction_id IS NOT NULL` | Indexes optional foreign keys while saving disk space on null values. |
| **`EXPLAIN ANALYZE`** | PostgreSQL Query Planner | Measures actual runtime execution costs and confirms index utilization over sequential scans. |
| **Containment (`@>`)** | JSONB Binary Operator | Evaluates whether a JSONB document contains a specific sub-structure (e.g. `feature_data @> '{"device_known": true}'`). |

### 0.5.3 Go Database Drivers & Schema Verification Stack

| Package / Module | Version / Source | Architectural Role in NOVARIS |
| :--- | :--- | :--- |
| **`database/sql`** | Go Standard Library | Generic SQL database interface providing connection pooling, statement preparation, and transactions. |
| **`github.com/lib/pq`** | Pure Go Postgres Driver | Low-level wire protocol driver connecting Go microservices to PostgreSQL 16. |
| **`cmd/verifyschema`** | Custom Developer CLI | Automated utility checking `information_schema.tables` and `information_schema.columns` for drift. |
| **`testing`** | Go Standard Library | Automated test harness executing offline SQL schema parsing and constraint assertions. |

### 0.5.4 Standards & Specifications

| Standard | Full Title | Why NOVARIS Adheres to It |
| :--- | :--- | :--- |
| **SQL:2016** | ISO/IEC 9075:2016 Information Technology — Database Languages — SQL | Governs core relational syntax, constraints, foreign keys, and JSON functionality. |
| **RFC 8259** | The JavaScript Object Notation (JSON) Data Interchange Format | Governs structure and data types allowed inside `feature_data` documents. |
| **ISO 8601** | Data elements and interchange formats — Representation of dates and times | Formats `TIMESTAMPTZ` columns to guarantee unambiguous UTC serialization. |

---

## 4. Chapter 1: The Theoretical Foundations of Relational Ledger Systems

### 1.1 The Missing Pieces: Why Days 1–4 Left Two Tables Open
On Day 1, our primary goal was proving that the **core double-entry ledger** worked:
* We needed `accounts` to hold balances.
* We needed `transactions` to represent payment transfers.
* We needed `journal_entries` to enforce balanced debits and credits.
* We needed `outbox` to eliminate dual-write vulnerabilities with Kafka.

At that early stage, adding tables for human fraud review queues and machine learning feature snapshots would have created unnecessary cognitive overload. However, the SRS Capstone Blueprint always required them:

```
┌────────────────────────────────────────────────────────────────────────────────────────┐
│                        NOVARIS FULL RELATIONAL DATABASE BLUEPRINT                      │
├──────────────────────────┬──────────────────────────┬──────────────────────────────────┤
│ TABLE NAME               │ PHASE / DAY ADDED        │ DOMAIN RESPONSIBILITY           │
├──────────────────────────┼──────────────────────────┼──────────────────────────────────┤
│ accounts                 │ Day 01                   │ Financial wallets, balances & OCC│
│ transactions             │ Day 01                   │ Payment requests & risk scores   │
│ journal_entries          │ Day 01                   │ Immutable double-entry ledger    │
│ outbox                   │ Day 01                   │ Transactional Kafka event relay  │
│ fraud_reviews            │ Day 05 (Today)           │ Human analyst review decisions   │
│ feature_snapshots        │ Day 05 (Today)           │ Point-in-time fraud features     │
└──────────────────────────┴──────────────────────────┴──────────────────────────────────┘
```

### 1.2 The Human-in-the-Loop Triage Architecture
In industrial payment processing, automated machine learning models make rapid initial determinations, but high-value financial networks never rely solely on black-box AI:

```
                         TRANSACTION ARRIVES WITH RISK SCORE (0 - 100)
                                              │
                    ┌─────────────────────────┼─────────────────────────┐
                    ▼                         ▼                         ▼
            Score < 30                30 <= Score <= 70             Score > 70
         ┌──────────────┐          ┌──────────────────────┐      ┌──────────────┐
         │ AUTO-APPROVE │          │ HUMAN ANALYST REVIEW │      │ AUTO-DECLINE │
         │ Instant      │          │ (fraud_reviews)      │      │ Immediate    │
         │ Settlement   │          │ Suspended in Escrow  │      │ Rejection    │
         └──────────────┘          └──────────┬───────────┘      └──────────────┘
                                              │
                             ┌────────────────┴────────────────┐
                             ▼                                 ▼
                     CONFIRMED_FRAUD                    FALSE_POSITIVE
                 (Funds Returned to User)          (Funds Settled to Merchant)
```

The table `fraud_reviews` is the permanent, legally audited anchor for the middle band ($30 \le \text{score} \le 70$).

### 1.3 The Temporal Decay of In-Memory State: Why Redis Needs an RDBMS Archive
Redis is an in-memory key-value store optimized for ultra-low latency ($< 1\text{ms}$). During live payment ingestion on Day 20, the gateway reads and writes Redis keys like:
* `account:acc_101:velocity_5m` $\rightarrow$ `14`
* `account:acc_101:last_ip` $\rightarrow$ `"192.168.1.1"`

However, Redis has two fatal limitations for machine learning and auditability:
1. **Destructive Mutation:** When the 5-minute window slides forward, the counter decrements. The historical fact that the account had 14 transactions at 2:14 PM is permanently erased.
2. **Volatile Memory:** Redis is not designed to retain 5 years of historical feature vectors across billions of payment operations.

`feature_snapshots` permanently serializes these ephemeral Redis values into PostgreSQL `JSONB` rows at the exact moment a transaction is evaluated.

---

## 5. Chapter 2: The Structural & Mathematical Architecture of JSONB and Composite Indexes

### 2.1 The Mathematics of B-Trees: Depth $O(\log N)$ and Hierarchical Sorting
A B-Tree (Balanced Tree) of order $m$ satisfies the following properties:
* Every node has at most $m$ children.
* Every non-leaf node (except root) has at least $\lceil m/2 \rceil$ children.
* All leaves appear at the exact same depth.

For a table containing $N = 10,000,000$ rows, the tree height $h$ is:
$$h \le \left\lceil \log_{\lceil m/2 \rceil} \left( \frac{N+1}{2} \right) \right\rceil$$

With typical PostgreSQL index page sizes ($m \approx 100$), the height of the B-Tree is **only 3 or 4 levels**. This means PostgreSQL can locate any single record among 10 million rows in **just 4 disk page reads**!

### 2.2 Hierarchical Column Ordering in Composite B-Trees
When creating a composite index on two columns:
```sql
CREATE INDEX idx_feature_snapshots_account_time 
    ON feature_snapshots(account_id, computed_at DESC);
```
PostgreSQL creates a composite key $(K_1, K_2)$ where:
$$K_1 = \text{account\_id}, \quad K_2 = \text{computed\_at}$$

The sort ordering $\prec$ is defined lexicographically:
$$(a_1, t_1) \prec (a_2, t_2) \iff (a_1 < a_2) \lor (a_1 = a_2 \land t_1 > t_2)$$

```
                                 COMPOSITE B-TREE ROOT NODE
                                             │
             ┌───────────────────────────────┴───────────────────────────────┐
             ▼                                                               ▼
   Entries for Account A                                           Entries for Account B
   (Pre-sorted by Time DESC)                                       (Pre-sorted by Time DESC)
  ┌─────────────────────────────────┐                             ┌─────────────────────────────────┐
  │ (Account_A, 2026-11-24 18:00)   │                             │ (Account_B, 2026-11-24 17:55)   │
  │ (Account_A, 2026-11-24 17:30)   │                             │ (Account_B, 2026-11-24 16:40)   │
  │ (Account_A, 2026-11-24 16:15)   │                             │ (Account_B, 2026-11-24 15:10)   │
  └─────────────────────────────────┘                             └─────────────────────────────────┘
```

### 2.3 The Phonebook Proof: Why `(computed_at, account_id)` Fails
Why can't we place `computed_at` first?
Let us inspect the query executed by our feature retrieval engine:
```sql
SELECT * FROM feature_snapshots 
WHERE account_id = 'c0a80101-0000-0000-0000-000000000001' 
ORDER BY computed_at DESC 
LIMIT 10;
```
1. **With `(account_id, computed_at DESC)` (Our Architecture):**
   * The database performs a B-Tree binary search on `account_id` in $O(\log N)$ time.
   * It lands directly on the first entry for that account.
   * Because entries within that account are already sorted by `computed_at DESC`, the database reads the top 10 items directly off the leaf node.
   * **Result:** Instant `Index Scan` with zero sorting overhead!

2. **With `(computed_at DESC, account_id)` (Bad Architecture):**
   * The tree is sorted primarily by timestamp across all accounts.
   * Account `c0a8...` appears scattered at random intervals throughout the entire tree.
   * The database cannot jump to the account; it must scan the entire index.
   * **Result:** The index is ignored, and PostgreSQL falls back to an expensive, slow `Seq Scan`!

### 2.4 Internal Storage Layout of JSONB
In PostgreSQL storage internals, `JSONB` values are stored in a contiguous binary container:
```
┌──────────────┬──────────────────────────────┬────────────────────────────────────────┐
│ HEADER (4B)  │ JEntry ARRAY (Offsets/Types) │ RAW DATA PAYLOAD                       │
│ 32-bit flags │ [Length & Type for Key 1]    │ "velocity_5m"                          │
│ count of keys│ [Length & Type for Val 1]    │ 12                                     │
│              │ [Length & Type for Key 2]    │ "geo_hop_km"                           │
│              │ [Length & Type for Val 2]    │ 0.0                                    │
└──────────────┴──────────────────────────────┴────────────────────────────────────────┘
```
Because the `JEntry` array stores the exact byte offset and length of every key, PostgreSQL searches for keys using binary search on sorted keys without decoding the entire document.

---

## 6. Chapter 3: Critical Database Engineering Gotchas & Schema Discipline

### 3.1 Migration Discipline: Why Editing `001_initial_ledger_schema.sql` Causes Fatal Drift
A common temptation for junior engineers is to open `migrations/001_initial_ledger_schema.sql`, add the two new tables at the bottom, and re-save the file.
**This is a critical violation of database engineering discipline:**
* In production, `001_initial_ledger_schema.sql` was already executed on Day 1.
* Production databases maintain a migration history table (e.g. `schema_migrations`). Once a migration hash is recorded as applied, the database engine **will never re-read that file**.
* If you edit an old migration, your local database (which might run from scratch) will have the new tables, while the production database will be completely missing them, leading to catastrophic runtime crashes (`relation "fraud_reviews" does not exist`).
* **Rule:** Never edit applied migrations. Always create a new, numbered file: `migrations/002_fraud_and_features.sql`.

### 3.2 The `docker-entrypoint-initdb.d` Trap
PostgreSQL's official Docker image includes a feature: any `.sql` files placed in `/docker-entrypoint-initdb.d/` are automatically executed on container startup.
**HOWEVER, this occurs ONLY ONCE when the database cluster is completely empty:**
* If the Docker container has already created its data volume (in our case, `pgdata`), PostgreSQL detects existing database files and skips `/docker-entrypoint-initdb.d/` entirely!
* Simply dropping `002_fraud_and_features.sql` into the folder does nothing to an existing database.
* The migration must be executed explicitly against the running container via `docker exec -i novaris-postgres psql -U novaris_admin -d novaris_db -f ...`.

### 3.3 The ENUM vs. VARCHAR+CHECK Dilemma
PostgreSQL supports native enumerated types:
```sql
CREATE TYPE review_verdict AS ENUM ('CONFIRMED_FRAUD', 'FALSE_POSITIVE');
```
While visually elegant, native enums carry a hidden operational cost:
* If our compliance team later requires a third status (e.g., `'SUSPECTED_CHARGEBACK'`), altering an enum requires running `ALTER TYPE ... ADD VALUE ...`.
* In older PostgreSQL engines, `ALTER TYPE` cannot be executed inside a multi-statement transaction block, complicating automated rollback tooling.
* By using `VARCHAR(16)` with a `CHECK` constraint:
  ```sql
  verdict VARCHAR(16) NOT NULL CHECK (verdict IN ('CONFIRMED_FRAUD', 'FALSE_POSITIVE'))
  ```
  Updating the rule in a future migration is as simple as:
  ```sql
  ALTER TABLE fraud_reviews DROP CONSTRAINT fraud_reviews_verdict_check;
  ALTER TABLE fraud_reviews ADD CONSTRAINT fraud_reviews_verdict_check 
      CHECK (verdict IN ('CONFIRMED_FRAUD', 'FALSE_POSITIVE', 'SUSPECTED_CHARGEBACK'));
  ```
  It executes within a standard ACID transaction with zero table locks.

### 3.4 Nullable vs. Required Foreign Keys
Notice the design of `feature_snapshots`:
```sql
account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
transaction_id UUID REFERENCES transactions(id) ON DELETE SET NULL,
```
* `account_id` is `NOT NULL`: A feature snapshot is meaningless if we don't know whose account it describes.
* `transaction_id` is **nullable**: Why? Because not all feature calculations are triggered by an immediate payment transaction!
  * For example, a background cron job might compute daily behavioral feature snapshots for high-risk merchants at 00:00:00 UTC. Making `transaction_id` nullable supports both transaction-triggered and periodic snapshots cleanly.

---

## 7. Chapter 4: The Closed MLOps Feedback Loop: Bridging Relational DBs to AI

```
┌────────────────────────────────────────────────────────────────────────────────────────────────────────┐
│                                   THE NOVARIS MLOPS FEEDBACK LOOP                                      │
├────────────────────────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                                        │
│   1. TRANSACTION INGESTION (Gateway & Kafka)                                                           │
│      A payment arrives. Real-time signals are computed in Redis (velocity, IP hops).                   │
│                                      │                                                                 │
│                                      ▼                                                                 │
│   2. POINT-IN-TIME FEATURE ARCHIVAL (Day 25)                                                           │
│      Features are saved to feature_snapshots table (JSONB).                                            │
│      -> {"velocity_5m": 12, "geo_hop_km": 1420.5, "device_known": false}                               │
│                                      │                                                                 │
│                                      ▼                                                                 │
│   3. INFERENCE & TRIAGE (FastAPI / XGBoost - Week 7-8)                                                 │
│      Model predicts risk score = 58. Transaction is routed to review queue.                            │
│                                      │                                                                 │
│                                      ▼                                                                 │
│   4. HUMAN ANALYST VERDICT (React 19 Dashboard - Week 9-10)                                            │
│      Analyst reviews the case, confirms stolen credentials, and writes verdict to fraud_reviews:       │
│      -> verdict: 'CONFIRMED_FRAUD', analyst_id: 'usr_analyst_99'                                       │
│                                      │                                                                 │
│                                      ▼                                                                 │
│   5. NIGHTLY RETRAINING DAG (Apache Airflow / MLflow - Day 66)                                         │
│      Airflow executes training dataset generation SQL:                                                 │
│      SELECT fs.feature_data, fr.verdict                                                                │
│      FROM feature_snapshots fs                                                                         │
│      JOIN fraud_reviews fr ON fs.transaction_id = fr.transaction_id                                    │
│      WHERE fr.reviewed_at >= NOW() - INTERVAL '30 days';                                               │
│                                      │                                                                 │
│                                      ▼                                                                 │
│   6. MODEL RETRAINING & DRIFT MONITORING                                                               │
│      XGBoost retrains on fresh, ground-truth human labels. New ONNX weights deployed!                  │
│                                                                                                        │
└────────────────────────────────────────────────────────────────────────────────────────────────────────┘
```

---

## 8. Chapter 5: The Data Science → Full-Stack Schema Rosetta Stone

For machine learning engineers and data scientists, relational schema design translates directly into familiar dataset concepts:

```
┌───────────────────────────────────────┬───────────────────────────────────────┬───────────────────────────────────────┐
│ CONCEPT                               │ DATA SCIENCE & MACHINE LEARNING       │ NOVARIS RELATIONAL SCHEMA (POSTGRES)  │
├───────────────────────────────────────┼───────────────────────────────────────┼───────────────────────────────────────┤
│ Unstructured / Flexible Features      │ Dict / JSON column in pandas          │ PostgreSQL JSONB Data Type            │
│                                       │ (e.g. metadata dict in DataFrame)     │ (Stores feature vectors flexibly      │
│                                       │                                       │ without rigid column migrations).     │
├───────────────────────────────────────┼───────────────────────────────────────┼───────────────────────────────────────┤
│ Hierarchical Index                    │ pandas MultiIndex                     │ Composite B-Tree Index                │
│                                       │ (df.set_index(['account', 'time']))   │ ON feature_snapshots(account, time)   │
│                                       │ (Sorts rows by multiple keys).        │ (Enables fast multi-key scans).       │
├───────────────────────────────────────┼───────────────────────────────────────┼───────────────────────────────────────┤
│ Ground-Truth Labels                   │ The target array `y_train`            │ The `fraud_reviews` Table             │
│                                       │ (Verified 0/1 outcome labels).        │ (Human verdicts on flagged payments). │
├───────────────────────────────────────┼───────────────────────────────────────┼───────────────────────────────────────┤
│ Offline Feature Store                 │ Feast / Hopsworks Offline Store       │ The `feature_snapshots` Table         │
│                                       │ (Historical feature values for        │ (Preserves point-in-time features     │
│                                       │ point-in-time correct training).      │ for offline training and audit).      │
├───────────────────────────────────────┼───────────────────────────────────────┼───────────────────────────────────────┤
│ Data Validation Schema                │ Pandera / Great Expectations check    │ PostgreSQL CHECK Constraint           │
│                                       │ (schema.validate(df) at pipeline edge)│ (verdict IN ('CONFIRMED_FRAUD', ...)) │
│                                       │                                       │ (Rejects invalid values on disk).     │
├───────────────────────────────────────┼───────────────────────────────────────┼───────────────────────────────────────┤
│ Query Execution Plan                  │ Spark Catalyst physical query plan    │ PostgreSQL EXPLAIN ANALYZE            │
│                                       │ (df.explain() showing physical ops).  │ (Inspects Index Scan vs. Seq Scan).   │
└───────────────────────────────────────┴───────────────────────────────────────┴───────────────────────────────────────┘
```

---

## 9. Chapter 6: Architectural Trade-Off Analysis ("Why This and Not That")

### 6.1 JSONB Column vs. Fully Normalized Relational Table

| Architecture Option | Write Latency | Schema Rigidity | Future Proofing | Best Used For |
| :--- | :--- | :--- | :--- | :--- |
| **JSONB Column (NOVARIS)** | **$< 1\text{ms}$** | **Extremely Flexible** | **High (Zero migrations as features grow)** | **Evolving ML feature sets** |
| Normalized Columns | $< 1\text{ms}$ | Rigid (1 column per feature) | Low (Requires `ALTER TABLE` for every new ML signal) | Stable financial balances |
| Separate Key-Value Table | $5-10\text{ms}$ (Many rows) | Semi-Flexible | Low (Huge row multiplication overhead) | Dynamic user profile tags |

* **Why JSONB Wins for Features:** Our fraud data science team will add 30 new features between Day 20 and Day 60 (device fingerprinting, IP subnet velocity, card BIN anomalies). If each feature required a dedicated SQL column, we would execute 30 table migrations. With `JSONB`, new features are added seamlessly to the payload dictionary.

### 6.2 VARCHAR + CHECK vs. Native PostgreSQL ENUM

* **Native ENUM:** Enforced at the engine type level. However, adding values requires lock-prone `ALTER TYPE` statements that cannot be safely rolled back in transaction blocks.
* **VARCHAR(16) + CHECK (NOVARIS):** Delivers identical database-level rejection of invalid strings while maintaining zero-downtime migration flexibility.

---

## 10. Chapter 7: Line-by-Line Code Masterclass (Complete Dissections)

### 7.1 `migrations/002_fraud_and_features.sql` (The Complete Day 5 Migration)

```sql
-- ==============================================================================
-- NOVARIS: Distributed Payment Settlement & Adaptive Fraud Intelligence Platform
-- MIGRATION 002: Human Fraud Review Logs & Point-in-Time Feature Snapshots (Day 5)
-- FILE: migrations/002_fraud_and_features.sql
-- ==============================================================================

-- ------------------------------------------------------------------------------
-- 1. FRAUD_REVIEWS TABLE
-- Stores human analyst review decisions on transactions scored in the review band (30-70).
-- ------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS fraud_reviews (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Foreign Key: Links directly to the specific transaction evaluated.
    transaction_id UUID NOT NULL REFERENCES transactions(id) ON DELETE RESTRICT,

    -- Identifier of the human fraud analyst who reviewed the flagged transaction.
    analyst_id UUID NOT NULL,

    -- Verdict outcome: Restricted by a strict CHECK constraint.
    verdict VARCHAR(16) NOT NULL
        CHECK (verdict IN ('CONFIRMED_FRAUD', 'FALSE_POSITIVE')),

    -- Free-form text notes detailing the analyst's investigative reasoning.
    notes TEXT,

    -- Timestamp when the human verdict was finalized.
    reviewed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_fraud_reviews_tx_id ON fraud_reviews(transaction_id);
CREATE INDEX IF NOT EXISTS idx_fraud_reviews_analyst_id ON fraud_reviews(analyst_id);
CREATE INDEX IF NOT EXISTS idx_fraud_reviews_verdict_time ON fraud_reviews(verdict, reviewed_at DESC);

-- ------------------------------------------------------------------------------
-- 2. FEATURE_SNAPSHOTS TABLE
-- Point-in-time archive of ephemeral Redis fraud features and real-time signals.
-- ------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS feature_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Foreign Key: Identifies the account/wallet whose features are being snapshotted.
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,

    -- Foreign Key (NULLABLE): The transaction that triggered this feature snapshot.
    transaction_id UUID REFERENCES transactions(id) ON DELETE SET NULL,

    -- Flexible JSONB document containing arbitrary feature vectors.
    feature_data JSONB NOT NULL,

    -- Monotonic timestamp recording exactly when the features were computed.
    computed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ------------------------------------------------------------------------------
-- 3. COMPOSITE INDEX DESIGN: (account_id, computed_at DESC)
-- ------------------------------------------------------------------------------
CREATE INDEX IF NOT EXISTS idx_feature_snapshots_account_time 
    ON feature_snapshots(account_id, computed_at DESC);

CREATE INDEX IF NOT EXISTS idx_feature_snapshots_tx_id 
    ON feature_snapshots(transaction_id) 
    WHERE transaction_id IS NOT NULL;
```

#### Detailed Line-by-Line Breakdown:
1. **`id UUID PRIMARY KEY DEFAULT gen_random_uuid()`:** Generates non-sequential UUIDv4 primary keys natively in PostgreSQL.
2. **`transaction_id UUID NOT NULL REFERENCES transactions(id) ON DELETE RESTRICT`:** Strict foreign key constraint. Prevents transactions from being deleted if a human review record exists.
3. **`verdict VARCHAR(16) NOT NULL CHECK (verdict IN (...))`:** Restricts allowed string values at the database engine level. Rejects any invalid verdicts immediately.
4. **`account_id UUID NOT NULL REFERENCES accounts(id)`:** Foreign key tying the feature vector to a specific financial wallet or merchant account.
5. **`transaction_id UUID REFERENCES transactions(id) ON DELETE SET NULL`:** Nullable foreign key. If a referenced transaction is purged, the feature snapshot remains intact with `transaction_id = NULL`.
6. **`feature_data JSONB NOT NULL`:** High-speed binary JSON column holding arbitrary feature vectors.
7. **`CREATE INDEX ... ON feature_snapshots(account_id, computed_at DESC)`:** High-performance composite B-Tree index supporting sub-millisecond chronological queries for a given account.

---

### 7.2 `services/gateway/cmd/verifyschema/main.go` (The Go Schema Verifier CLI)

```go
package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// RequiredTables lists all 6 load-bearing tables required by the NOVARIS SRS Blueprint.
var RequiredTables = []string{
	"accounts",
	"transactions",
	"journal_entries",
	"outbox",
	"fraud_reviews",
	"feature_snapshots",
}

// RequiredIndexes lists the mission-critical composite and partial indexes.
var RequiredIndexes = []string{
	"idx_feature_snapshots_account_time",
	"idx_fraud_reviews_verdict_time",
	"idx_outbox_pending",
}

func main() {
	offlineMode := flag.Bool("offline", false, "Run static SQL schema analysis without connecting to PostgreSQL")
	sqlFilePath := flag.String("sql", "", "Path to migration SQL file for static parsing")
	flag.Parse()

	log.Println("==================================================================")
	log.Println(" 🔍 NOVARIS Database Schema & Migration Verifier (Day 5)")
	log.Println("==================================================================")

	if *offlineMode || *sqlFilePath != "" {
		runOfflineVerification(*sqlFilePath)
		return
	}

	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://novaris_admin:dev_password_change_me@localhost:5432/novaris_db?sslmode=disable"
	}

	// Mask password in logs
	maskedURL := maskConnectionURL(connStr)
	log.Printf("[DB] Target Connection: %s\n", maskedURL)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Printf("[WARN] Failed to initialize database driver: %v\n", err)
		log.Println("[INFO] Falling back to offline static migration verification...")
		runOfflineVerification("")
		return
	}
	defer db.Close()

	// Verify database connectivity with 2-second timeout
	db.SetConnMaxLifetime(2 * time.Second)
	if err := db.Ping(); err != nil {
		log.Printf("[INFO] Live PostgreSQL connection unavailable (%v)\n", err)
		log.Println("[INFO] Performing static migration file verification instead...")
		runOfflineVerification("")
		return
	}

	// Live database verification
	runLiveVerification(db)
}
```

#### Detailed Line-by-Line Breakdown:
1. **Lines 17–28:** `RequiredTables` and `RequiredIndexes` constants define the target ground-truth schema state mandated by the blueprint.
2. **Lines 34–40:** CLI flags `-offline` and `-sql` allow headless CI/CD pipelines to validate schema integrity without booting database containers.
3. **Lines 44–55:** Connects to PostgreSQL using `DATABASE_URL` with a safety timeout and masks the plaintext database password in log outputs.
4. **Lines 60–66:** Graceful fallback: if a live PostgreSQL container is offline, it automatically switches to deterministic static file analysis.

---

### 7.3 `services/gateway/cmd/verifyschema/main_test.go` (Automated Schema Test Suite)

```go
package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func findMigrationsDir(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	curr := wd
	for i := 0; i < 5; i++ {
		candidate := filepath.Join(curr, "migrations")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
		parent := filepath.Dir(curr)
		if parent == curr {
			break
		}
		curr = parent
	}
	t.Fatalf("could not locate 'migrations' directory starting from %s", wd)
	return ""
}

func TestMigrationFiles_ExistenceAndNaming(t *testing.T) {
	migDir := findMigrationsDir(t)

	// 1. Verify 001_initial_ledger_schema.sql exists
	file001 := filepath.Join(migDir, "001_initial_ledger_schema.sql")
	if _, err := os.Stat(file001); os.IsNotExist(err) {
		t.Fatalf("expected migration file %s to exist, but was missing", file001)
	}

	// 2. Verify 002_fraud_and_features.sql exists
	file002 := filepath.Join(migDir, "002_fraud_and_features.sql")
	if _, err := os.Stat(file002); os.IsNotExist(err) {
		t.Fatalf("expected migration file %s to exist, but was missing", file002)
	}
}

func TestMigration002_SchemaCompleteness(t *testing.T) {
	migDir := findMigrationsDir(t)
	target := filepath.Join(migDir, "002_fraud_and_features.sql")

	content, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("could not read %s: %v", target, err)
	}

	sqlUpper := strings.ToUpper(string(content))

	// Verify required tables
	expectedTables := []string{"FRAUD_REVIEWS", "FEATURE_SNAPSHOTS"}
	for _, table := range expectedTables {
		if !strings.Contains(sqlUpper, "TABLE IF NOT EXISTS "+table) && !strings.Contains(sqlUpper, "TABLE "+table) {
			t.Errorf("migration 002 missing CREATE TABLE statement for %s", table)
		}
	}

	// Verify CHECK constraint on verdict
	if !strings.Contains(sqlUpper, "CHECK (VERDICT IN ('CONFIRMED_FRAUD', 'FALSE_POSITIVE'))") {
		t.Error("missing strict CHECK constraint on fraud_reviews.verdict")
	}

	// Verify JSONB data type for feature_data
	if !strings.Contains(sqlUpper, "FEATURE_DATA JSONB NOT NULL") {
		t.Error("feature_snapshots.feature_data must be defined as JSONB NOT NULL")
	}

	// Verify Composite Index with correct column ordering: account_id first, then computed_at DESC
	if !strings.Contains(sqlUpper, "INDEX IF NOT EXISTS IDX_FEATURE_SNAPSHOTS_ACCOUNT_TIME") {
		t.Error("missing composite index idx_feature_snapshots_account_time")
	}
	if !strings.Contains(sqlUpper, "ON FEATURE_SNAPSHOTS(ACCOUNT_ID, COMPUTED_AT DESC)") {
		t.Error("composite index must place account_id before computed_at DESC")
	}

	// Verify Foreign Key references
	if !strings.Contains(sqlUpper, "REFERENCES TRANSACTIONS(ID)") {
		t.Error("missing foreign key reference to transactions(id)")
	}
	if !strings.Contains(sqlUpper, "REFERENCES ACCOUNTS(ID)") {
		t.Error("missing foreign key reference to accounts(id)")
	}
}

func TestMaskConnectionURL(t *testing.T) {
	raw := "postgres://novaris_admin:super_secret_pwd@localhost:5432/novaris_db?sslmode=disable"
	masked := maskConnectionURL(raw)

	if strings.Contains(masked, "super_secret_pwd") {
		t.Errorf("maskConnectionURL leaked plaintext password in: %s", masked)
	}
	if !strings.Contains(masked, "*****@localhost:5432") {
		t.Errorf("unexpected masked format: %s", masked)
	}
}
```

---

### 7.4 `services/gateway/cmd/README.md` (Updated CLI Documentation)

```markdown
## 3. `verifyschema` — Database Schema & Migration Verifier

Located at `cmd/verifyschema/main.go`, this utility validates that the live PostgreSQL instance or offline SQL migration files strictly match the NOVARIS SRS Capstone Blueprint:
* Confirms the existence of all 6 core tables: `accounts`, `transactions`, `journal_entries`, `outbox`, `fraud_reviews`, `feature_snapshots`.
* Checks column data types, JSONB binary storage format, and CHECK constraints.
* Validates composite B-Tree indexes (`idx_feature_snapshots_account_time`).

### Usage Examples:
```bash
# Run static migration verification (ideal for CI/CD and offline development)
go run cmd/verifyschema/main.go -offline

# Verify live schema against running PostgreSQL container
go run cmd/verifyschema/main.go

# Run automated Go unit test suite
go test -v ./cmd/verifyschema
```
```

---

## 11. Chapter 8: Verification Laboratory: Execution Proof & Complete Runbook

### 8.1 Complete Automated Test Suite Output
Execution of all 21 unit, integration, and schema tests across the gateway:

```bash
$ go test -v ./...
=== RUN   TestPipeline_BurstAllowanceAndRateLimitEnforcement
--- PASS: TestPipeline_BurstAllowanceAndRateLimitEnforcement (0.00s)
=== RUN   TestPipeline_FailFastSecurityOrdering
--- PASS: TestPipeline_FailFastSecurityOrdering (0.00s)
PASS
ok  	github.com/charan21042005/novaris/services/gateway	(cached)

=== RUN   TestMigrationFiles_ExistenceAndNaming
--- PASS: TestMigrationFiles_ExistenceAndNaming (0.00s)
=== RUN   TestMigration002_SchemaCompleteness
--- PASS: TestMigration002_SchemaCompleteness (0.00s)
=== RUN   TestMaskConnectionURL
--- PASS: TestMaskConnectionURL (0.00s)
PASS
ok  	github.com/charan21042005/novaris/services/gateway/cmd/verifyschema	(cached)

=== RUN   TestComputeAndVerifyHMAC_Success
--- PASS: TestComputeAndVerifyHMAC_Success (0.00s)
=== RUN   TestVerifyHMAC_TamperedBody
--- PASS: TestVerifyHMAC_TamperedBody (0.00s)
=== RUN   TestVerifyHMAC_StaleTimestamp
--- PASS: TestVerifyHMAC_StaleTimestamp (0.00s)
=== RUN   TestVerifyHMAC_FutureTimestamp
--- PASS: TestVerifyHMAC_FutureTimestamp (0.00s)
=== RUN   TestVerifyHMAC_WrongSecret
--- PASS: TestVerifyHMAC_WrongSecret (0.00s)
=== RUN   TestHMACMiddleware
--- PASS: TestHMACMiddleware (0.00s)
=== RUN   TestGenerateAndValidateToken
--- PASS: TestGenerateAndValidateToken (0.00s)
=== RUN   TestExpiredToken
--- PASS: TestExpiredToken (0.00s)
=== RUN   TestTamperedToken
--- PASS: TestTamperedToken (0.00s)
=== RUN   TestJWTMiddleware
--- PASS: TestJWTMiddleware (0.00s)
PASS
ok  	github.com/charan21042005/novaris/services/gateway/internal/auth	(cached)

=== RUN   TestTokenBucket_BurstAndExhaustion
--- PASS: TestTokenBucket_BurstAndExhaustion (0.00s)
=== RUN   TestTokenBucket_RefillOverTime
--- PASS: TestTokenBucket_RefillOverTime (0.15s)
=== RUN   TestTokenBucket_CapacityCeiling
--- PASS: TestTokenBucket_CapacityCeiling (0.05s)
=== RUN   TestBucketStore_PerMerchantIsolation
--- PASS: TestBucketStore_PerMerchantIsolation (0.00s)
=== RUN   TestTokenBucket_ConcurrentRaceConditions
--- PASS: TestTokenBucket_ConcurrentRaceConditions (0.00s)
=== RUN   TestRateLimitMiddleware
--- PASS: TestRateLimitMiddleware (0.00s)
PASS
ok  	github.com/charan21042005/novaris/services/gateway/internal/ratelimit	(cached)
```

---

### 8.2 Live Docker & PostgreSQL Command Matrix

To verify that the migration functions seamlessly on live database infrastructure, the following commands execute the migration and assert table creation, index registration, and constraint behavior inside the running Docker container:

#### 1. Copy and Apply Migration to PostgreSQL Container:
```bash
# Copy migration file into the container's temporary filesystem
docker cp migrations/002_fraud_and_features.sql novaris-postgres:/tmp/002_fraud_and_features.sql

# Execute the migration script using psql
docker exec -i novaris-postgres psql -U novaris_admin -d novaris_db -f /tmp/002_fraud_and_features.sql
```

#### 2. Confirm Schema Table Creation:
```bash
# List all relations (all 6 tables must now appear)
docker exec -it novaris-postgres psql -U novaris_admin -d novaris_db -c "\dt"
```
*Expected Output:*
```text
               List of relations
 Schema |       Name        | Type  |     Owner     
--------+-------------------+-------+---------------
 public | accounts          | table | novaris_admin
 public | feature_snapshots | table | novaris_admin
 public | fraud_reviews     | table | novaris_admin
 public | journal_entries   | table | novaris_admin
 public | outbox            | table | novaris_admin
 public | transactions      | table | novaris_admin
(6 rows)
```

#### 3. Inspect Composite Index on `feature_snapshots`:
```bash
docker exec -it novaris-postgres psql -U novaris_admin -d novaris_db -c "\d+ feature_snapshots"
```
*Expected Output:*
```text
                                                 Table "public.feature_snapshots"
     Column     |           Type           | Collation | Nullable |      Default      | Storage  | Compression | Stats target | Description 
----------------+--------------------------+-----------+----------+-------------------+----------+-------------+--------------+-------------
 id             | uuid                     |           | not null | gen_random_uuid() | plain    |             |              | 
 account_id     | uuid                     |           | not null |                   | plain    |             |              | 
 transaction_id | uuid                     |           |          |                   | plain    |             |              | 
 feature_data   | jsonb                    |           | not null |                   | extended |             |              | 
 computed_at    | timestamp with time zone |           | not null | now()             | plain    |             |              | 
Indexes:
    "feature_snapshots_pkey" PRIMARY KEY, btree (id)
    "idx_feature_snapshots_account_time" btree (account_id, computed_at DESC)
    "idx_feature_snapshots_tx_id" btree (transaction_id) WHERE transaction_id IS NOT NULL
Foreign-key constraints:
    "feature_snapshots_account_id_fkey" FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE RESTRICT
    "feature_snapshots_transaction_id_fkey" FOREIGN KEY (transaction_id) REFERENCES transactions(id) ON DELETE SET NULL
```

#### 4. Test Strict CHECK Constraint on `fraud_reviews`:
Attempting to insert an unauthorized verdict string (e.g. `'PENDING_INSPECTION'`):
```bash
docker exec -it novaris-postgres psql -U novaris_admin -d novaris_db -c "
INSERT INTO fraud_reviews (transaction_id, analyst_id, verdict, notes) 
VALUES ('c0a80101-0000-0000-0000-000000000001', 'a0a80101-0000-0000-0000-000000000001', 'SUSPICIOUS_ONLY', 'Analyst flag');"
```
*Expected Database Rejection:*
```text
ERROR:  new row for relation "fraud_reviews" violates check constraint "fraud_reviews_verdict_check"
DETAIL:  Failing row contains (..., SUSPICIOUS_ONLY, Analyst flag, 2026-11-24 ...).
```
The database engine directly defends domain invariants!

#### 5. Verify JSONB Containment (`@>`) & Text Extraction (`->>`):
```bash
# Insert valid test feature snapshot
docker exec -it novaris-postgres psql -U novaris_admin -d novaris_db -c "
INSERT INTO feature_snapshots (account_id, feature_data) 
VALUES ('a1111111-1111-1111-1111-111111111111', '{\"velocity_5m\": 14, \"geo_hop_km\": 1820.4, \"device_known\": false, \"vpn_detected\": true}');"

# Query 1: Extract velocity_5m as text
docker exec -it novaris-postgres psql -U novaris_admin -d novaris_db -c "
SELECT id, account_id, feature_data->>'velocity_5m' AS velocity, feature_data->>'vpn_detected' AS vpn FROM feature_snapshots;"

# Query 2: Filter with JSONB containment operator (@>)
docker exec -it novaris-postgres psql -U novaris_admin -d novaris_db -c "
SELECT id, computed_at FROM feature_snapshots WHERE feature_data @> '{\"vpn_detected\": true}';"
```

#### 6. Confirm B-Tree Index Utilization with `EXPLAIN ANALYZE`:
```bash
docker exec -it novaris-postgres psql -U novaris_admin -d novaris_db -c "
EXPLAIN ANALYZE 
SELECT * FROM feature_snapshots 
WHERE account_id = 'a1111111-1111-1111-1111-111111111111' 
ORDER BY computed_at DESC 
LIMIT 5;"
```
*Expected Output:*
```text
                                                                    QUERY PLAN
--------------------------------------------------------------------------------------------------------------------------------------------------
 Limit  (cost=0.15..8.17 rows=1 width=116) (actual time=0.015..0.016 rows=1 loops=1)
   ->  Index Scan using idx_feature_snapshots_account_time on feature_snapshots  (cost=0.15..8.17 rows=1 width=116) (actual time=0.014..0.014 rows=1 loops=1)
         Index Cond: (account_id = 'a1111111-1111-1111-1111-111111111111'::uuid)
 Planning Time: 0.082 ms
 Execution Time: 0.035 ms
(5 rows)
```
Notice: `Index Scan using idx_feature_snapshots_account_time` executed in **0.035 ms**! Zero disk sort, zero sequential scan.

---

### 8.3 Complete Step-by-Step CLI Runbook

#### For Linux / macOS (Bash):
```bash
# 1. Run automated offline schema tests in Go
cd services/gateway
go test -v ./cmd/verifyschema

# 2. Run schema verifier CLI in offline mode
go run cmd/verifyschema/main.go -offline

# 3. Apply migration to PostgreSQL container
docker cp ../../migrations/002_fraud_and_features.sql novaris-postgres:/tmp/002.sql
docker exec -i novaris-postgres psql -U novaris_admin -d novaris_db -f /tmp/002.sql

# 4. Verify live schema through Go CLI utility
go run cmd/verifyschema/main.go

# 5. Run full gateway test suite
go test -v ./...
```

#### For Windows (PowerShell):
```powershell
# 1. Run automated offline schema tests in Go
Set-Location services\gateway
go test -v ./cmd/verifyschema

# 2. Run schema verifier CLI in offline mode
go run cmd/verifyschema/main.go -offline

# 3. Apply migration to PostgreSQL container
docker cp ..\..\migrations\002_fraud_and_features.sql novaris-postgres:/tmp/002.sql
docker exec -i novaris-postgres psql -U novaris_admin -d novaris_db -f /tmp/002.sql

# 4. Verify live schema through Go CLI utility
go run cmd/verifyschema/main.go

# 5. Run full gateway test suite
go test -v ./...
```

---

## 12. Chapter 9: The Interview Defense Compendium (10 High-Stakes Tier-1 Q&As)

When discussing distributed ledger architecture and database schema design with principal engineers at companies like Stripe, Adyen, Block, or Brex, the following ten technical questions frequently arise:

### Q1: Why did NOVARIS choose PostgreSQL JSONB instead of spinning up MongoDB or a separate NoSQL document store for feature vectors?
> **Candidate Answer:**  
> Introducing a dedicated NoSQL cluster like MongoDB creates substantial operational overhead: dual-database operations, divergent backup lifecycles, and most dangerously, the **dual-write problem**. If a transaction is committed in PostgreSQL but the feature write fails in MongoDB, our ledger state and ML training state become permanently decoupled.  
> PostgreSQL's `JSONB` gives us the exact schema flexibility of document stores inside an ACID-compliant engine. In NOVARIS, the payment record, the double-entry journal entries, and the point-in-time feature snapshot can be written inside a **single atomic transaction (`BEGIN ... COMMIT`)**. If anything fails, the entire transaction rolls back, guaranteeing 100% data consistency between financial ledger records and offline ML training sets without managing extra infrastructure.

---

### Q2: Walk me through the mathematical B-Tree structure and explain why `(account_id, computed_at DESC)` was chosen instead of `(computed_at DESC, account_id)`.
> **Candidate Answer:**  
> A composite B-Tree sorts lexicographically according to its leading key. In our index `(account_id, computed_at DESC)`, `account_id` is the primary partition key, and `computed_at DESC` sorts records within that account's leaf segment.  
> Our feature retrieval pattern is: *"Given account $A$, retrieve its most recent 10 feature snapshots."*  
> With our index, PostgreSQL executes a B-Tree search on `account_id` in $O(\log N)$ steps, lands directly at the head of account $A$'s pre-sorted timeline, and reads the top 10 rows sequentially off the disk page.  
> If we reversed the columns to `(computed_at DESC, account_id)`, the B-Tree would be ordered globally by timestamp. Account $A$'s snapshots would be scattered across millions of different timestamp nodes across the entire tree. To find 10 rows for Account $A$, PostgreSQL would have to scan the entire index or abandon it for a full table `Seq Scan`. The leading column must always match the highest-cardinality equality filter of the query.

---

### Q3: Why did you use `VARCHAR(16) + CHECK` instead of native PostgreSQL `ENUM`?
> **Candidate Answer:**  
> PostgreSQL native `ENUM` types are internally represented as separate engine catalog types (`pg_type`). While compact (4 bytes on disk), they introduce significant operational friction during schema evolution:  
> 1. In standard PostgreSQL, `ALTER TYPE ... ADD VALUE` cannot be executed inside a multi-statement transaction block, meaning automated database migration pipelines (like Liquibase, Flyway, or Golang Migrate) fail if they wrap migrations in `BEGIN ... COMMIT`.  
> 2. Removing or renaming an enum value requires dropping and recreating the type or rewriting the table, which requires an exclusive table lock that blocks all reads and writes.  
> 3. By using `VARCHAR(16)` with a `CHECK (verdict IN ('CONFIRMED_FRAUD', 'FALSE_POSITIVE'))`, we achieve identical data integrity at the storage layer while keeping migrations fully transactional and zero-downtime. Modifying the allowed set in a future migration is an instantaneous metadata update that safely rolls back if aborted.

---

### Q4: Why is `feature_snapshots.transaction_id` nullable while `fraud_reviews.transaction_id` is `NOT NULL`?
> **Candidate Answer:**  
> This directly reflects the real-world domain life-cycle:  
> A `fraud_review` only occurs because a specific payment transaction was flagged by the risk pipeline (scoring 30–70) and queued for a human analyst. A review without an associated transaction is conceptually impossible, so `transaction_id` is strictly `NOT NULL REFERENCES transactions(id) ON DELETE RESTRICT`.  
> Conversely, `feature_snapshots` captures point-in-time state vectors. While many snapshots are triggered by incoming payment transactions, others are computed on a scheduled cadence (e.g. nightly merchant risk aggregations, or pre-computed profile updates upon user login). Setting `transaction_id` to nullable allows the same table to support both transaction-bound snapshots and ambient account snapshots without artificial dummy transactions.

---

### Q5: What is the `docker-entrypoint-initdb.d` trap and how did your migration strategy prevent schema drift?
> **Candidate Answer:**  
> In PostgreSQL Docker images, scripts in `/docker-entrypoint-initdb.d` only execute when the database cluster is initialized for the first time (an empty volume). Once a data directory exists (e.g. from Day 1), Docker skips that directory on subsequent container boots.  
> A common developer mistake is editing `001_initial_ledger_schema.sql` or adding `002_*.sql` to that directory and wondering why the tables don't exist.  
> In NOVARIS, we enforce strict migration discipline:  
> 1. Existing migrations are immutable.  
> 2. New changes are authored in incremental, numbered files (`002_fraud_and_features.sql`).  
> 3. In running environments, migrations are applied explicitly using standard migration tooling or `psql -f`.  
> 4. We built a dedicated Go CLI tool (`cmd/verifyschema`) that interrogates PostgreSQL's `information_schema` directly to verify table and index presence in CI pipelines, catching schema drift before any code is merged.

---

### Q6: Explain how `fraud_reviews` acts as the ground-truth training set for Day 66's nightly XGBoost retraining pipeline.
> **Candidate Answer:**  
> Supervised machine learning algorithms require high-quality labeled data. In payment fraud, the model's output at inference time ($P(\text{fraud}) = 0.62$) is merely a probabilistic estimate, not a ground truth.  
> The true outcome is determined later: either by a cardholder chargeback dispute, or by a human fraud investigator who analyzes user behavior, device logs, and merchant receipts.  
> When an investigator completes an investigation, their verdict (`CONFIRMED_FRAUD` $\rightarrow 1$, `FALSE_POSITIVE` $\rightarrow 0$) is recorded in `fraud_reviews`.  
> On Day 66, our Apache Airflow retraining DAG executes a SQL query joining `feature_snapshots` and `fraud_reviews` on `transaction_id` over a rolling 30-day window. This produces the feature matrix $\mathbf{X}$ and the ground-truth target vector $\mathbf{y}$ used to train the next iteration of our XGBoost model, closing the MLOps feedback loop.

---

### Q7: Why not store features exclusively in Redis where reads are sub-millisecond?
> **Candidate Answer:**  
> Redis is an exceptional in-memory cache for the **instantaneous present**, but it is structurally unsuitable as a feature archive:  
> 1. **Temporal Overwrite:** Redis velocity keys (like a sliding window counter) mutate constantly. Once a 5-minute window passes, the count from 10 minutes ago is overwritten. Redis cannot tell you what a user's velocity was during a specific transaction three weeks ago.  
> 2. **Cost of In-Memory Retention:** Storing billions of historical feature snapshots in RAM would require hundreds of gigabytes of expensive memory.  
> 3. **ACID Joins:** Redis cannot perform relational joins with transaction settlement records or analyst audit tables.  
> In NOVARIS, Redis serves real-time reads during the sub-12ms inference path, but every evaluated feature vector is asynchronously written to `feature_snapshots` in PostgreSQL `JSONB` for permanent storage, auditability, and offline training.

---

### Q8: How does the JSONB containment operator `@>` work internally in PostgreSQL?
> **Candidate Answer:**  
> The containment operator `@>` checks whether the left JSONB document contains the entirety of the right JSONB document (e.g. `feature_data @> '{"vpn_detected": true}'`).  
> Internally, because `JSONB` is stored in decomposed binary format with sorted keys and pre-calculated offset headers, PostgreSQL does not parse text. It performs a binary search on the `JEntry` header array to check if the key `"vpn_detected"` exists. If found, it jumps directly to the value bytes to assert equality with `true`.  
> Furthermore, when indexed with a GIN (Generalized Inverted Index) using `jsonb_path_ops`, queries using `@>` achieve logarithmic lookup speeds across millions of JSON documents.

---

### Q9: What happens if someone tries to delete a transaction that has an existing `fraud_reviews` entry? Explain `ON DELETE RESTRICT`.
> **Candidate Answer:**  
> The foreign key is configured as:  
> `transaction_id UUID NOT NULL REFERENCES transactions(id) ON DELETE RESTRICT`.  
> If an administrator or service attempts to execute `DELETE FROM transactions WHERE id = '...'`, the PostgreSQL query planner checks the referential integrity catalog. Because an active row in `fraud_reviews` references that transaction's primary key, the database engine **aborts the transaction and raises error code 23503 (`foreign_key_violation`)**.  
> This is a mandatory compliance requirement in FinTech: financial transactions and their corresponding fraud investigations must never be deleted. `ON DELETE RESTRICT` guarantees that no application bug or erroneous script can destroy historical financial records.

---

### Q10: How does your Go CLI tool (`cmd/verifyschema`) enforce automated schema governance in CI/CD without requiring a live database connection?
> **Candidate Answer:**  
> In fast continuous integration environments, spinning up a full Docker container mesh for every commit adds latency and consumes runner resources.  
> We designed `cmd/verifyschema` with a dual-mode verification engine:  
> 1. **Live Mode:** Connects to PostgreSQL, checks `information_schema.tables`, validates column data types, and inspects `pg_indexes` for composite index definitions.  
> 2. **Offline Static Mode (`-offline`):** When no database is reachable, the tool automatically parses the repository's `.sql` migration files. It validates that files follow monotonic naming conventions (`001`, `002`), scans the AST/SQL tokens for required `CREATE TABLE` and `CREATE INDEX` statements, verifies that `CHECK` constraints match exact regex patterns, and confirms composite index column ordering.  
> This allows CI pipelines to run schema assertions in **under 5 milliseconds** during pull request checks while maintaining live verification in staging environments.

---

## 13. Chapter 10: The Roadmap Horizon (Day 6 Preview: Atomic Double-Entry Settlement)

With Day 5 complete, our relational foundation is 100% aligned with the NOVARIS Capstone Blueprint:

```
┌──────────────────────────────────────────────────────────────────────────────────┐
│                            NOVARIS SCHEMA COMPLETE (DAY 5)                       │
├────────────────────┬─────────────────────────────────────────────────────────────┤
│ Table              │ Role                                                        │
├────────────────────┼─────────────────────────────────────────────────────────────┤
│ accounts           │ User & merchant balance tracking with version OCC           │
│ transactions       │ Ingress payment requests, risk scores & settlement status   │
│ journal_entries    │ Immutable double-entry ledger (Debits = Credits)            │
│ outbox             │ Transactional Kafka event relay table                       │
│ fraud_reviews      │ Human investigator decisions for middle-band risk scores    │
│ feature_snapshots  │ Point-in-time JSONB features with composite B-Tree indexing │
└────────────────────┴─────────────────────────────────────────────────────────────┘
```

### 10.1 Day 6 Architectural Objectives
Tomorrow, on **Day 6**, we shift our focus from schema definitions to **transactional business logic**:
1. **The Double-Entry Settlement Engine:**
   * Implementing atomic money movement between accounts.
   * Ensuring that for every transaction, $\sum \text{Debits} \equiv \sum \text{Credits}$.
2. **ACID Transaction Boundaries:**
   * Wrapping account balance deductions and journal entry creation inside single atomic PostgreSQL transactions (`BEGIN`, `COMMIT`, `ROLLBACK`).
3. **Concurrency Defense & Negative Balance Rejections:**
   * Using Optimistic Concurrency Control (`version` checks) to reject overdraft attempts and prevent race conditions when two simultaneous charges hit the same wallet.

---

### 10.2 Hands-On Mastery Exercises for the Reader

To solidify your mastery of Day 5 concepts, complete the following three exercises:

1. **Exercise 1: GIN Index Creation on JSONB:**
   * Add a Generalized Inverted Index (GIN) using `jsonb_path_ops` on `feature_snapshots.feature_data`:
     ```sql
     CREATE INDEX idx_feature_snapshots_gin ON feature_snapshots USING GIN (feature_data jsonb_path_ops);
     ```
   * Use `EXPLAIN ANALYZE` to compare query speed between a B-Tree composite query and a GIN containment query (`feature_data @> '{"vpn_detected": true}'`).

2. **Exercise 2: Write a Python Feature Extraction Script:**
   * Using `psycopg2` or `asyncpg`, write a 20-line Python script that connects to `novaris_db`, queries the most recent 100 `feature_snapshots`, and converts the `feature_data` JSONB column directly into a `pandas.DataFrame`.

3. **Exercise 3: Test Cascading Foreign Key Behavior:**
   * Create a dummy transaction and an associated `feature_snapshots` record.
   * Delete the transaction (`DELETE FROM transactions WHERE id = '...'`).
   * Observe how `ON DELETE SET NULL` automatically nulls the `transaction_id` column while preserving the feature vector for historical analytics!

---

<div align="center">

<!-- High-Definition Native Static Footer -->
<p align="center">
  <img src="../../assets/footer.svg" alt="NOVARIS Footer" width="100%" />
</p>

### 🎓 NOVARIS Engineering Curriculum • Volume 01: Foundations & Ledgers
**Authored by Tokachichu Sri Charan**  
*Distributed Systems Engineer & FinTech Architect*

[GitHub Repository](https://github.com/charan21042005/novaris) • [Documentation Hub](https://github.com/charan21042005/novaris/tree/main/bin/Daily_Actions) • [Day 06 Preview](./Day_06_Action_Log.md)

</div>
ok  	github.com/charan21042005/novaris/services/gateway/internal/auth	(cached)

=== RUN   TestTokenBucket_BurstAndExhaustion
--- PASS: TestTokenBucket_BurstAndExhaustion (0.00s)
=== RUN   TestTokenBucket_RefillOverTime
--- PASS: TestTokenBucket_RefillOverTime (0.15s)
=== RUN   TestTokenBucket_CapacityCeiling
--- PASS: TestTokenBucket_CapacityCeiling (0.05s)
=== RUN   TestBucketStore_PerMerchantIsolation
--- PASS: TestBucketStore_PerMerchantIsolation (0.00s)
=== RUN   TestTokenBucket_ConcurrentRaceConditions
--- PASS: TestTokenBucket_ConcurrentRaceConditions (0.00s)
=== RUN   TestRateLimitMiddleware
--- PASS: TestRateLimitMiddleware (0.00s)
PASS
ok  	github.com/charan21042005/novaris/services/gateway/internal/ratelimit	(cached)
```

---

### 8.2 Real-World Database Verification Matrix

The table below catalogs every concrete probe scenario verified for Day 5:

| Scenario | Actions Performed | Database Observation | Result |
| :--- | :--- | :--- | :--- |
| **Probe 1: Migration Execution** | Apply `002_fraud_and_features.sql` to PostgreSQL | `CREATE TABLE` and `CREATE INDEX` executed without errors | ✅ PASS |
| **Probe 2: Table Existence** | Query `information_schema.tables` | All 6 tables (`accounts`, `transactions`, `journal_entries`, `outbox`, `fraud_reviews`, `feature_snapshots`) exist | ✅ PASS |
| **Probe 3: CHECK Constraint** | Attempt inserting verdict `'MAYBE'` into `fraud_reviews` | Rejected by database with check constraint violation | ✅ PASS |
| **Probe 4: Valid Verdict** | Insert row with verdict `'CONFIRMED_FRAUD'` | Row accepted and committed with UUIDv4 primary key | ✅ PASS |
| **Probe 5: JSONB Querying** | Query `feature_snapshots` using `feature_data->>'velocity_5m'` | Extracted integer value without full document parsing | ✅ PASS |
| **Probe 6: Composite Index** | Run `EXPLAIN ANALYZE` filtering by `account_id` | Query plan utilizes `Index Scan using idx_feature_snapshots_account_time` | ✅ PASS |

---

### 8.3 Complete Step-by-Step CLI Runbook

To apply and verify Day 5's database schema locally, run the following commands:

#### Option A: Running Automated Go Schema Tests
```bash
cd services/gateway

# Run automated schema parser and test suite
go test -v ./cmd/verifyschema

# Run full gateway test suite
go test -v ./...

# Run static schema verification CLI
go run cmd/verifyschema/main.go -offline
```

#### Option B: Live PostgreSQL Application (Docker / Bash)
```bash
# 1. Copy migration into the running PostgreSQL container
docker cp migrations/002_fraud_and_features.sql novaris-postgres:/tmp/

# 2. Execute migration inside container
docker exec -it novaris-postgres psql -U novaris_admin -d novaris_db -f /tmp/002_fraud_and_features.sql

# 3. Verify all 6 tables exist
docker exec -it novaris-postgres psql -U novaris_admin -d novaris_db -c "\dt"

# 4. Insert test fraud review
docker exec -it novaris-postgres psql -U novaris_admin -d novaris_db -c "
  INSERT INTO fraud_reviews (transaction_id, analyst_id, verdict, notes)
  SELECT id, gen_random_uuid(), 'FALSE_POSITIVE', 'Verified legitimate via customer call'
  FROM transactions LIMIT 1;
"

# 5. Insert test feature snapshot
docker exec -it novaris-postgres psql -U novaris_admin -d novaris_db -c "
  INSERT INTO feature_snapshots (account_id, feature_data)
  SELECT id, '{\"velocity_5m\": 4, \"geo_hop_km\": 0.0, \"device_known\": true}'::jsonb
  FROM accounts LIMIT 1;
"

# 6. Verify JSONB containment query
docker exec -it novaris-postgres psql -U novaris_admin -d novaris_db -c "
  SELECT id, feature_data->>'velocity_5m' AS velocity 
  FROM feature_snapshots 
  WHERE feature_data @> '{\"device_known\": true}';
"
```

#### Option C: Live PostgreSQL Application (Windows PowerShell)
```powershell
# 1. Copy migration into container
docker cp migrations\002_fraud_and_features.sql novaris-postgres:/tmp/

# 2. Execute migration script
docker exec -it novaris-postgres psql -U novaris_admin -d novaris_db -f /tmp/002_fraud_and_features.sql

# 3. Verify tables
docker exec -it novaris-postgres psql -U novaris_admin -d novaris_db -c "\dt"

# 4. Run Go Schema Verifier against live DB
cd services\gateway
go run cmd\verifyschema\main.go
```

---

## 12. Chapter 9: The Interview Defense Compendium (10 High-Stakes Tier-1 Q&As)

### Q1: Why did you choose JSONB over traditional normalized columns for `feature_snapshots`?
> **Answer:** Machine learning feature sets in fraud detection are highly volatile. Between Days 20 and 60, our data science team will introduce dozens of experimental signals (device hashes, IP velocity, card attempt counters). Normalizing every feature into a dedicated column would require dozens of schema migrations, schema locking, and complex ORM refactors. `JSONB` stores features in pre-parsed binary format, supporting sub-millisecond key lookups and GIN containment indexes while allowing the feature set to evolve dynamically without database downtime.

### Q2: Why does column order in a composite index matter, and why is `account_id` placed before `computed_at`?
> **Answer:** A composite B-Tree index sorts rows lexicographically by the first column, and then by the second column within matching first-column values. In NOVARIS, our query pattern is strictly: *"Fetch the historical feature snapshots for Account X, ordered by time"*. Placing `account_id` first allows PostgreSQL to binary-search directly to that account's branch in $O(\log N)$ time and stream the pre-sorted timestamps. If `computed_at` came first, finding Account X would require scanning every timestamp entry across all accounts, rendering the index useless.

### Q3: Why did you use `VARCHAR(16) + CHECK` instead of a native PostgreSQL `ENUM` for `verdict`?
> **Answer:** Native PostgreSQL ENUMs are static database types. Modifying an enum to add a new verdict (e.g. `SUSPECTED_CHARGEBACK`) requires `ALTER TYPE`, which cannot be executed inside a multi-statement transaction in older engines and can trigger table-level locks. `VARCHAR(16) + CHECK` provides identical database-level type safety while allowing the allowed verdict list to be modified seamlessly within an ACID transaction migration.

### Q4: Why is `transaction_id` nullable in `feature_snapshots`?
> **Answer:** Not all feature snapshots are generated during a payment checkout. Our architecture anticipates scheduled background evaluations (e.g., hourly account risk scoring, nightly fraud batch runs) where feature vectors are recorded when no immediate payment is taking place. Making `transaction_id` nullable supports both transaction-triggered and periodic snapshots cleanly.

### Q5: Why did we create `002_fraud_and_features.sql` instead of editing `001_initial_ledger_schema.sql`?
> **Answer:** It is a core principle of **Migration Discipline**. Once a migration script has executed against any persistent environment, it becomes an immutable historical record. If an engineer edits an already-applied migration file, existing databases will never re-execute it, causing silent schema drift where developer environments and production environments permanently disagree.

### Q6: How does `fraud_reviews` serve as the bridge between human operations and machine learning?
> **Answer:** It represents the **authoritative ground-truth dataset** for supervised model training. When an analyst reviews a transaction in the 30–70 score band and selects `CONFIRMED_FRAUD` or `FALSE_POSITIVE`, that verified human verdict is written to `fraud_reviews`. On Day 66, our Apache Airflow retraining pipeline will join `feature_snapshots` with `fraud_reviews` to extract verified training samples, allowing the XGBoost model to continuously learn from new fraud tactics.

### Q7: Why is `ON DELETE RESTRICT` specified on foreign keys in financial schemas?
> **Answer:** In financial and audit architectures, historical records must be legally preserved. `ON DELETE RESTRICT` prevents foreign key cascades from accidentally deleting transactions or journal entries if a user or account is removed. Any attempt to delete referenced records will be rejected by the relational engine.

### Q8: What is the difference between JSON and JSONB in PostgreSQL?
> **Answer:** Plain `JSON` stores data as an exact raw text representation, preserving whitespace but requiring full string parsing on every read. `JSONB` parses the text upon insertion and stores it in an optimized, decomposed binary format with internal key offset arrays. Reading a key from `JSONB` is dramatically faster, and `JSONB` supports GIN indexes and containment operators (`@>`).

### Q9: Why does `docker-entrypoint-initdb.d` fail to apply new migrations on existing Docker containers?
> **Answer:** PostgreSQL's entrypoint script only executes files in `/docker-entrypoint-initdb.d` during initial database cluster bootstrapping (when the data directory is completely empty). If the persistent volume already exists, initialization is skipped. New migrations must be applied explicitly via client connections or automated migration runners like `golang-migrate`.

### Q10: How do you verify that an index is actually being utilized rather than a sequential scan?
> **Answer:** By running `EXPLAIN ANALYZE <query>`. The output reveals the query planner's execution tree: if an index is active, the node type will read `Index Scan using <index_name>` or `Bitmap Index Scan`. If the table is small, PostgreSQL may choose a `Seq Scan` (because sequential reading is faster for a few pages), but forcing `SET enable_seqscan = off;` verifies that the index path is structurally valid.

---

## 13. Chapter 10: The Roadmap Horizon (Day 6 Preview: Atomic Double-Entry Settlement)

With the relational schema fully aligned with the SRS Capstone Blueprint:
1. `accounts` (Balances, Currencies & OCC Version)
2. `transactions` (Ingress Lifecycles & Risk Scores)
3. `journal_entries` (Double-Entry Debit/Credit Invariants)
4. `outbox` (Atomic Kafka Streaming Relay)
5. `fraud_reviews` (Human Ground-Truth Audit Log)
6. `feature_snapshots` (Point-in-Time JSONB Feature Vectors)

Tomorrow on **Day 6**, we implement the heart of financial computation: **Atomic Double-Entry Settlement Logic**.

```
                UPCOMING MILESTONE: DAY 6 SETTLEMENT CORE
                
 ┌──────────────────────┐         ┌────────────────────────────────────────────────────────┐
 │ Ingress Gateway      │         │ Atomic Settlement Worker (Day 6)                       │
 │ (Days 1 - 4)         │         │                                                        │
 │                      │ ──────> │ * ACID Transaction: BEGIN -> COMMIT / ROLLBACK         │
 │                      │         │ * Verify Balance Sufficiency: sender_balance >= amount │
 │                      │         │ * Create Dual Journal Entries: DEBIT sender, CREDIT rcpt│
 │                      │         │ * Atomically Enforce Invariant: SUM(debit) == SUM(cred)│
 └──────────────────────┘         └────────────────────────────────────────────────────────┘
```

### 10.1 Day 6 Architectural Objectives:
1. **ACID Transaction Orchestration:** Execute multi-table updates inside strict PostgreSQL transactions with automatic rollbacks on error.
2. **Balance Sufficiency Checks:** Guarantee that accounts can never drop below zero cents.
3. **Double-Entry Enforcement:** Ensure that every debit has an equal and opposite credit.

---

### 10.2 Hands-On Mastery Exercises for the Reader

1. **JSONB Containment Querying Exercise:**
   * Insert three rows into `feature_snapshots` with varying values for `device_known` (`true` vs. `false`) and `velocity_5m` ($2$, $15$, $45$).
   * Write an SQL query using the `@>` containment operator to fetch only snapshots where `{"device_known": true}`.
   * Write a query using `feature_data->>'velocity_5m'` casting to integer to filter snapshots where velocity exceeds $10$.

2. **Index Explain Analyze Exercise:**
   * Insert 1,000 synthetic rows into `feature_snapshots` across 5 different `account_id` values.
   * Execute:
     ```sql
     EXPLAIN ANALYZE SELECT * FROM feature_snapshots 
     WHERE account_id = '<your-account-uuid>' 
     ORDER BY computed_at DESC;
     ```
   * Verify that the plan output reports `Index Scan using idx_feature_snapshots_account_time`.

3. **CHECK Constraint Verification Exercise:**
   * Execute an `INSERT` statement into `fraud_reviews` with verdict `'MAYBE_FRAUD'`.
   * Confirm that PostgreSQL returns error code `23514` (`check_violation`).

---

<div align="center">

<p align="center">
  <img src="../../assets/footer.svg" alt="NOVARIS Footer" width="100%" />
</p>

### 🎓 NOVARIS Educational Action Textbook • Day 05 Complete
*Architected with Precision by Tokachichu Sri Charan*

</div>
