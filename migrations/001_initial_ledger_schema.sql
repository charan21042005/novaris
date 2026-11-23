-- ==============================================================================
-- NOVARIS: Distributed Payment Settlement & Adaptive Fraud Intelligence Platform
-- MIGRATION 001: Initial Double-Entry Ledger & Transactional Outbox Schema
-- ==============================================================================
--
-- Architectural Principles Behind This Schema:
-- 1. Financial Precision: All monetary figures are stored in BIGINT cents (e.g., $10.50 = 1050 cents).
--    NEVER use FLOAT or DOUBLE in financial ledgers; IEEE-754 binary floating-point representation
--    causes rounding errors (e.g. 0.1 + 0.2 = 0.30000000000000004) that silently corrupt money balances.
--
-- 2. Mathematical Balance (Double-Entry Bookkeeping):
--    Account balances are never simply updated with blind `balance = balance - amount` queries.
--    Every financial transfer produces an immutable set of balanced rows in `journal_entries`:
--    sum(DEBIT) must exactly equal sum(CREDIT) for every single settled transaction.
--
-- 3. Optimistic Concurrency Control (OCC):
--    The `version` integer in the `accounts` table protects against race conditions when two
--    concurrent requests attempt to settle against the same wallet simultaneously.
--
-- 4. Dual-Write Elimination (Transactional Outbox):
--    The `outbox` table guarantees atomic event delivery to Apache Kafka by committing
--    both the business transaction and the outgoing Kafka event inside the EXACT SAME ACID transaction.
-- ==============================================================================

-- Enable UUID generation extension if not already present
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ------------------------------------------------------------------------------
-- 1. ACCOUNTS TABLE
-- Represents wallets, merchant balances, and internal escrow accounts.
-- ------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    -- Account types: WALLET (user funds), ESCROW (clearing), MERCHANT_PAYOUT, FEE_REVENUE
    account_type VARCHAR(32) NOT NULL,
    -- ISO 4217 Currency Code (default: USD)
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    -- Stored in integer cents to prevent floating point drift
    balance_cents BIGINT NOT NULL DEFAULT 0,
    -- Optimistic Concurrency Control (OCC) version column.
    -- Incremented with every balance mutation. Fails stale writes atomically.
    version INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index for querying all accounts owned by a specific user
CREATE INDEX IF NOT EXISTS idx_accounts_user_id ON accounts(user_id);
-- Index on currency and account type for aggregation and reporting
CREATE INDEX IF NOT EXISTS idx_accounts_type_currency ON accounts(account_type, currency);

-- ------------------------------------------------------------------------------
-- 2. TRANSACTIONS TABLE
-- Records incoming payment requests and their lifecycle state.
-- ------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    -- Deterministic client-supplied key for 24-hour idempotent deduplication
    idempotency_key VARCHAR(64) UNIQUE NOT NULL,
    sender_account_id UUID REFERENCES accounts(id) ON DELETE RESTRICT,
    recipient_account_id UUID REFERENCES accounts(id) ON DELETE RESTRICT,
    -- Transfer amount in integer cents
    amount_cents BIGINT NOT NULL CHECK (amount_cents > 0),
    -- Lifecycle: PENDING -> REVIEW -> APPROVED -> SETTLED (or DECLINED / FAILED)
    status VARCHAR(24) NOT NULL DEFAULT 'PENDING',
    -- Real-time fraud risk score computed by the ML Engine (0 to 100)
    fraud_risk_score INT CHECK (fraud_risk_score >= 0 AND fraud_risk_score <= 100),
    -- Explainable AI risk factors (SHAP attribution vector and rule triggers)
    risk_factors JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Fast lookup index for client-supplied idempotency key
CREATE INDEX IF NOT EXISTS idx_tx_idempotency ON transactions(idempotency_key);
-- Index for querying transactions by status (e.g. pending review queue)
CREATE INDEX IF NOT EXISTS idx_tx_status ON transactions(status);
-- Index for sender and recipient transaction history
CREATE INDEX IF NOT EXISTS idx_tx_sender ON transactions(sender_account_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_tx_recipient ON transactions(recipient_account_id, created_at DESC);

-- ------------------------------------------------------------------------------
-- 3. JOURNAL_ENTRIES TABLE
-- The immutable double-entry ledger. Every movement of money is logged here.
-- Total Debits MUST equal Total Credits for any settled transaction.
-- ------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS journal_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id UUID NOT NULL REFERENCES transactions(id) ON DELETE RESTRICT,
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
    -- Entry type: 'DEBIT' (money leaving an account) or 'CREDIT' (money arriving)
    entry_type VARCHAR(6) NOT NULL CHECK (entry_type IN ('DEBIT', 'CREDIT')),
    -- Absolute value of the movement in cents (always positive)
    amount_cents BIGINT NOT NULL CHECK (amount_cents > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Primary ledger audit index: retrieve all chronological entries for an account
CREATE INDEX IF NOT EXISTS idx_journal_account ON journal_entries(account_id, created_at DESC);
-- Index to verify that debits equal credits for a given transaction
CREATE INDEX IF NOT EXISTS idx_journal_tx ON journal_entries(transaction_id);

-- ------------------------------------------------------------------------------
-- 4. OUTBOX TABLE
-- Implements the Transactional Outbox Pattern to eliminate dual-write vulnerabilities.
-- Business writes and outbox records commit in the same database transaction.
-- ------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS outbox (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    -- Domain aggregate: e.g., 'TRANSACTION', 'ACCOUNT', 'FRAUD_ALERT'
    aggregate_type VARCHAR(64) NOT NULL,
    aggregate_id UUID NOT NULL,
    -- Kafka topic target: e.g., 'tx.initiated', 'tx.settled', 'tx.declined'
    event_type VARCHAR(64) NOT NULL,
    -- Complete serialized event payload
    payload JSONB NOT NULL,
    -- Outbox relay status: 'PENDING', 'PUBLISHED', 'FAILED'
    status VARCHAR(16) NOT NULL DEFAULT 'PENDING',
    retry_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    published_at TIMESTAMPTZ
);

-- Partial index for the outbox relay poller to quickly fetch un-published events
CREATE INDEX IF NOT EXISTS idx_outbox_pending ON outbox(created_at ASC) WHERE status = 'PENDING';
