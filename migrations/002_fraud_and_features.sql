-- ==============================================================================
-- NOVARIS: Distributed Payment Settlement & Adaptive Fraud Intelligence Platform
-- MIGRATION 002: Human Fraud Review Logs & Point-in-Time Feature Snapshots (Day 5)
-- FILE: migrations/002_fraud_and_features.sql
-- ==============================================================================
--
-- Architectural Purpose & Domain Significance:
-- Days 1-4 established perimeter ingress defenses (JWT, HMAC, Token-Bucket Rate Limiter)
-- and core transactional bookkeeping (accounts, transactions, journal_entries, outbox).
--
-- Today on Day 5, we complete the SRS Capstone relational database blueprint by
-- adding the two missing load-bearing tables required for future ML and Operations:
--
-- 1. `fraud_reviews`:
--    - Represents: A human fraud analyst's final verdict on an escalated transaction.
--    - Written By: The Human-in-the-Loop Analyst Dashboard (React 19 / WebSocket - Week 9-10).
--    - Read By: The Nightly Retraining DAG (Airflow / MLflow - Day 66 / Week 7-8).
--    - Machine Learning Role: THIS IS NOT JUST AN AUDIT LOG. This table is the literal,
--      authoritative source of human-verified GROUND-TRUTH LABELS used to retrain
--      the XGBoost and Graph Neural Network fraud detection models.
--
-- 2. `feature_snapshots`:
--    - Represents: A point-in-time immutable archive of ephemeral fraud signals.
--    - Written By: The Feature Enrichment & Ingestion Service (Kafka / Redis - Day 25).
--    - Read By: The Offline ML Training Pipeline (Day 44-46) and Feast Feature Store (Day 51).
--    - Machine Learning Role: Redis sliding counters only store the "current" state
--      (e.g., velocity in the last 5 minutes right now). Redis CANNOT tell you what
--      the velocity was 3 weeks ago when a specific fraudulent charge occurred.
--      This table archives computed features permanently in JSONB format.
--
-- 3. Composite Indexing:
--    - An index on (account_id, computed_at DESC) allows instant chronological retrieval
--      of an account's behavioral history without expensive full-table sequential scans.
--
-- MIGRATION DISCIPLINE NOTE:
-- We NEVER modify `001_initial_ledger_schema.sql` after it has been executed on a database.
-- Numbered, immutable migrations ensure every staging, dev, and production database
-- executes the exact same sequence of deterministic schema transformations.
-- ==============================================================================

-- ------------------------------------------------------------------------------
-- 1. FRAUD_REVIEWS TABLE
-- Stores human analyst review decisions on transactions scored in the review band (30-70).
-- ------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS fraud_reviews (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Foreign Key: Links directly to the specific transaction evaluated.
    -- If a transaction is deleted, CASCADE or RESTRICT prevents dangling audit trails.
    transaction_id UUID NOT NULL REFERENCES transactions(id) ON DELETE RESTRICT,

    -- Identifier of the human fraud analyst who reviewed the flagged transaction.
    -- Guarantees auditability, regulatory compliance, and analyst performance tracking.
    analyst_id UUID NOT NULL,

    -- Verdict outcome: Restricted by a strict CHECK constraint to prevent typos.
    -- Pattern: VARCHAR(16) + CHECK constraint is chosen over native Postgres ENUM
    -- because it easily accommodates future verdict additions (e.g. 'ESCALATED_L2')
    -- via simple ALTER TABLE migrations without lock-heavy ALTER TYPE statements.
    verdict VARCHAR(16) NOT NULL
        CHECK (verdict IN ('CONFIRMED_FRAUD', 'FALSE_POSITIVE')),

    -- Free-form text notes detailing the analyst's investigative reasoning
    -- (e.g. "Customer confirmed card was stolen via phone call").
    notes TEXT,

    -- Timestamp when the human verdict was finalized.
    reviewed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index on transaction_id for instant retrieval of review status during settlement checks
CREATE INDEX IF NOT EXISTS idx_fraud_reviews_tx_id ON fraud_reviews(transaction_id);

-- Index on analyst_id for supervisor audit logs and analyst productivity metrics
CREATE INDEX IF NOT EXISTS idx_fraud_reviews_analyst_id ON fraud_reviews(analyst_id);

-- Index on verdict and reviewed_at for training dataset extraction queries (e.g. Day 66 DAG)
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
    -- Nullable by design: Enables periodic scheduled snapshots (e.g., hourly account health)
    -- even when no immediate transaction is currently occurring.
    transaction_id UUID REFERENCES transactions(id) ON DELETE SET NULL,

    -- Flexible JSONB document containing arbitrary feature vectors:
    -- e.g. {"velocity_5m": 12, "geo_hop_km": 1420.5, "device_known": false, "card_attempts": 3}
    -- JSONB (Binary JSON) is selected over plain JSON because:
    -- 1. It is stored in a decomposed binary format, eliminating text parsing on reads.
    -- 2. It supports indexing via Generalized Inverted Indexes (GIN) in future phases.
    -- 3. It supports fast existence and containment operators (@>, ?, ?&).
    feature_data JSONB NOT NULL,

    -- Monotonic timestamp recording exactly when the features were computed.
    computed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ------------------------------------------------------------------------------
-- 3. COMPOSITE INDEX DESIGN: WHY COLUMN ORDER MATTERS
-- ------------------------------------------------------------------------------
-- A composite B-Tree index physically sorts data hierarchically:
-- It sorts FIRST by the 1st column, and THEN by the 2nd column within identical 1st values.
--
-- CRITICAL QUERY ACCESS PATTERN:
-- In fraud intelligence and feature retrieval, the query pattern is ALWAYS:
--   "Fetch all feature snapshots for Account X, ordered by most recent first"
--   -> SELECT * FROM feature_snapshots WHERE account_id = $1 ORDER BY computed_at DESC;
--
-- COLUMN ORDERING JUSTIFICATION:
-- 1. (account_id, computed_at DESC):
--    - The database can jump directly to the account's branch in the B-Tree in O(log N)
--      and scan forward/backward along computed_at in pre-sorted order!
-- 2. If the order were reversed to (computed_at, account_id):
--    - The index would sort all accounts by timestamp. To find Account X, the database
--      would have to scan across thousands of different timestamps—rendering the index
--      virtually USELESS for account-level history queries.
--
-- Therefore, account_id MUST come first, followed by computed_at DESC.
-- ------------------------------------------------------------------------------
CREATE INDEX IF NOT EXISTS idx_feature_snapshots_account_time 
    ON feature_snapshots(account_id, computed_at DESC);

-- Auxiliary index for looking up snapshots triggered by a specific transaction
CREATE INDEX IF NOT EXISTS idx_feature_snapshots_tx_id 
    ON feature_snapshots(transaction_id) 
    WHERE transaction_id IS NOT NULL;
