-- ==============================================================================
-- NOVARIS: Distributed Payment Settlement & Adaptive Fraud Intelligence Platform
-- MIGRATION 002: Human-in-the-Loop Fraud Reviews & Ephemeral Feature Snapshots
-- ==============================================================================
--
-- Architectural Purpose & Milestone Context (Day 5 of 84):
-- Days 2 through 4 fortified the API Gateway perimeter (JWT -> Rate Limiting -> HMAC).
-- Day 5 shifts inward to complete the persistence foundation defined in the SRS Blueprint:
--
-- 1. Human-in-the-Loop Governance (`fraud_reviews`):
--    When transactions fall into the ambiguous risk band (Risk Score 30 - 70), they are
--    flagged for human analyst triage (built in Weeks 9-10). This table permanently
--    records human verdicts ('CONFIRMED_FRAUD' vs 'FALSE_POSITIVE').
--    *Critical Downstream Dependency*: This table forms the verified ground-truth labeled
--    training dataset queried by Day 66's automated nightly Airflow retraining DAG!
--
-- 2. Temporal Machine Learning Feature Freezing (`feature_snapshots`):
--    In Week 3 (Day 20), we build sub-millisecond Redis velocity counters. However,
--    Redis is an ephemeral in-memory cache—it only knows an account's velocity *right now*,
--    never what it was three weeks ago during a disputed transaction.
--    `feature_snapshots` acts as the permanent offline feature store (Week 7-8 & Feast Day 51),
--    preserving historical model inputs at the exact microsecond of transaction scoring.
--
-- 3. Semi-Structured JSONB Flexibility:
--    The `feature_data` column uses binary JSON (JSONB). As feature engineering evolves
--    across Weeks 7-8 (adding device fingerprints, geo-hops, graph embeddings), new
--    signals are added without requiring continuous schema migrations.
--
-- 4. High-Performance Composite Indexing:
--    Implements a composite B-Tree index on `(account_id, computed_at DESC)`.
--    The leading column (`account_id`) narrows the search space to a single tenant, while
--    the second column (`computed_at DESC`) returns chronological history with zero sort overhead.
-- ==============================================================================

-- ------------------------------------------------------------------------------
-- 1. FRAUD_REVIEWS TABLE
-- Stores human analyst review decisions on flagged transactions (Score 30-70).
-- ------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS fraud_reviews (
    -- Unique primary identifier for this review record
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Foreign Key binding this review directly to the underlying financial transaction.
    -- ON DELETE RESTRICT: Prevents accidental deletion of a transaction with an active audit log.
    transaction_id UUID NOT NULL REFERENCES transactions(id) ON DELETE RESTRICT,

    -- Identifier of the security analyst or automated compliance officer who made the determination
    analyst_id UUID NOT NULL,

    -- The human verdict. Strictly constrained at the database engine level via CHECK constraint.
    -- VARCHAR + CHECK is chosen over native Postgres ENUM for consistency with Day 1's
    -- `transactions.status` and to allow friction-free future additions via ALTER TABLE.
    verdict VARCHAR(16) NOT NULL CHECK (verdict IN ('CONFIRMED_FRAUD', 'FALSE_POSITIVE')),

    -- Free-form textual reasoning entered by the analyst explaining why the flag was confirmed or overturned
    notes TEXT,

    -- ISO-8601 UTC timestamp of when the analyst submitted their final determination
    reviewed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index for looking up all reviews associated with a specific transaction
CREATE INDEX IF NOT EXISTS idx_fraud_reviews_tx 
    ON fraud_reviews(transaction_id);

-- Index for auditing analyst workload and productivity over time
CREATE INDEX IF NOT EXISTS idx_fraud_reviews_analyst 
    ON fraud_reviews(analyst_id, reviewed_at DESC);

-- Composite index for Day 66's retraining pipeline: quickly extract training batches by verdict and date
CREATE INDEX IF NOT EXISTS idx_fraud_reviews_verdict_time 
    ON fraud_reviews(verdict, reviewed_at DESC);


-- ------------------------------------------------------------------------------
-- 2. FEATURE_SNAPSHOTS TABLE
-- Permanent historical archive of ephemeral Redis fraud signals captured at scoring time.
-- ------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS feature_snapshots (
    -- Unique primary identifier for this feature snapshot
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- The account whose behavioral profile and velocity signals are being snapshotted
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,

    -- The specific transaction that triggered this feature extraction.
    -- Nullable by architectural design: allows non-transactional periodic background snapshots.
    transaction_id UUID REFERENCES transactions(id) ON DELETE SET NULL,

    -- De-normalized, semi-structured feature vector stored as binary JSON (JSONB).
    -- Example payload:
    -- {
    --   "velocity_1m": 3,
    --   "velocity_5m": 12,
    --   "amount_ratio_30d": 4.25,
    --   "geo_hop_speed_mph": 842.1,
    --   "ip_country": "US",
    --   "device_trust_score": 0.94
    -- }
    feature_data JSONB NOT NULL,

    -- Exact microsecond timestamp when the feature vector was computed by the enrichment worker
    computed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index for looking up features captured for a specific transaction
CREATE INDEX IF NOT EXISTS idx_feature_snapshots_tx 
    ON feature_snapshots(transaction_id) 
    WHERE transaction_id IS NOT NULL;

-- ------------------------------------------------------------------------------
-- COMPOSITE INDEX: (account_id, computed_at DESC)
-- ------------------------------------------------------------------------------
-- WHY THIS EXACT COLUMN ORDER MATTERS (The Left-to-Right Index Rule):
-- A B-Tree composite index physically sorts data hierarchically:
--   1. First by `account_id` (Equality filter)
--   2. Within each `account_id`, by `computed_at DESC` (Range/Sort order)
--
-- Anticipated Access Pattern:
--   "Fetch the last 50 feature snapshots for Account X ordered by computed_at DESC"
--   Query: SELECT * FROM feature_snapshots WHERE account_id = '...' ORDER BY computed_at DESC LIMIT 50;
--
-- Why NOT (computed_at DESC, account_id)?
-- If `computed_at` were the leading column, searching for Account X across all time would
-- require scanning through every single timestamp bucket on earth! Leading with `account_id`
-- instantly isolates the B-Tree search to that merchant's sub-tree.
-- ------------------------------------------------------------------------------
CREATE INDEX IF NOT EXISTS idx_feature_snapshots_account_time 
    ON feature_snapshots(account_id, computed_at DESC);

-- GIN (Generalized Inverted Index) on the JSONB payload:
-- Accelerates containment queries (@>) and JSON path lookups when filtering by internal feature keys
-- Example: SELECT * FROM feature_snapshots WHERE feature_data @> '{"ip_country": "US"}';
CREATE INDEX IF NOT EXISTS idx_feature_snapshots_data_gin 
    ON feature_snapshots USING GIN (feature_data);
