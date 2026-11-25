-- ==============================================================================
-- NOVARIS: Distributed Payment Settlement & Adaptive Fraud Intelligence Platform
-- MIGRATION 003: Idempotent API Request Deduplication & Response Caching (Day 10)
-- FILE: migrations/003_idempotency_keys.sql
-- ==============================================================================
--
-- Architectural Purpose & Milestone Context (Day 10 of 84):
-- In distributed financial payment systems, network timeouts and transient disconnects
-- are inevitable. A mobile app or merchant gateway can submit a $500 transfer, the
-- backend processes the payment, but the client's connection drops before receiving
-- the HTTP 200 response. The client client library automatically retries the request!
--
-- Without an idempotency layer:
-- The retried request creates a second debit, charging the customer twice ($1,000)!
--
-- With NOVARIS Day 10 Idempotency:
-- 1. Deterministic Deduplication: Every mutating API request includes an Idempotency-Key.
-- 2. Database-Enforced Uniqueness: The `key` column is the PRIMARY KEY.
-- 3. Canonical Payload Fingerprinting (`request_hash`): Prevents key reuse with mismatched
--    payloads (e.g. attempting to reuse a key from a $10 transfer for a $5,000 transfer).
-- 4. Response Caching: Caches the exact HTTP status code and JSON response body for 24 hours.
-- 5. Safe Network Retries: Duplicate requests return the original cached response with
--    header `Idempotent-Replayed: true` with ZERO secondary charges!
-- ==============================================================================

-- Enable UUID extension if not already present
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ------------------------------------------------------------------------------
-- IDEMPOTENCY_KEYS TABLE
-- Tracks in-flight and completed API requests to guarantee exact-once execution.
-- ------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS idempotency_keys (
    -- Client-supplied unique token (UUIDv4 or cryptographic random token, max 128 chars)
    key VARCHAR(128) PRIMARY KEY,

    -- Account or Merchant UUID that initiated the request
    account_id VARCHAR(64) NOT NULL,

    -- API endpoint path (e.g., '/v1/payments')
    request_path VARCHAR(255) NOT NULL,

    -- SHA-256 cryptographic digest of the raw canonical request body (hex-encoded 64 chars)
    request_hash VARCHAR(64) NOT NULL,

    -- Execution lifecycle state:
    --   IN_PROGRESS : Request is currently being processed by a gateway pod
    --   COMPLETED   : Request completed successfully; response_code and response_body are populated
    --   FAILED      : Request aborted due to validation or internal error
    status VARCHAR(24) NOT NULL CHECK (status IN ('IN_PROGRESS', 'COMPLETED', 'FAILED')),

    -- Cached HTTP status code returned to the original caller (e.g. 200, 201, 400, 422)
    response_code INT,

    -- Cached JSON response body returned to the original caller
    response_body JSONB,

    -- ISO-8601 UTC timestamp of original request ingestion
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Expiration timestamp (default: 24 hours). Expired keys can be pruned by background cleanup
    expires_at TIMESTAMPTZ NOT NULL DEFAULT (NOW() + INTERVAL '24 hours')
);

-- Index for querying idempotency history by account and chronological order
CREATE INDEX IF NOT EXISTS idx_idemp_account_created ON idempotency_keys(account_id, created_at DESC);

-- Index for automated TTL cleanup / pruning of expired keys
CREATE INDEX IF NOT EXISTS idx_idemp_expires_at ON idempotency_keys(expires_at);
