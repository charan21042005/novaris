-- ==============================================================================
-- NOVARIS: Distributed Payment Settlement & Adaptive Fraud Intelligence Platform
-- MIGRATION 004: Deterministic Test Accounts for Integration Wiring (Day 11)
-- FILE: migrations/004_seed_test_accounts.sql
-- ==============================================================================
--
-- Milestone Purpose (Day 11 of 84: Integration Wiring):
-- In Day 11, the entire perimeter-to-database chain is connected into a single
-- working request path:
--   JWT Auth -> Rate Limiter -> HMAC Verification -> Idempotency Cache -> OCC Settlement -> PostgreSQL Ledger
--
-- To execute end-to-end integration tests without manual database setup, this migration
-- seeds deterministic, idempotent merchant test accounts into the `accounts` table.
-- ==============================================================================

-- ------------------------------------------------------------------------------
-- 1. SEED TEST ACCOUNTS
-- Using ON CONFLICT (id) DO NOTHING guarantees idempotency on container reboots.
-- ------------------------------------------------------------------------------

-- Sender Account: Merchant Alpha (Funded with $10,000.00 USD = 1,000,000 cents)
INSERT INTO accounts (id, user_id, account_type, currency, balance_cents, version)
VALUES (
    'c0000000-0000-0000-0000-000000000001',
    'a0000000-0000-0000-0000-000000000001',
    'WALLET',
    'USD',
    1000000,
    0
) ON CONFLICT (id) DO NOTHING;

-- Recipient Account: Merchant Beta (Funded with $10,000.00 USD = 1,000,000 cents)
INSERT INTO accounts (id, user_id, account_type, currency, balance_cents, version)
VALUES (
    'c0000000-0000-0000-0000-000000000002',
    'a0000000-0000-0000-0000-000000000002',
    'WALLET',
    'USD',
    1000000,
    0
) ON CONFLICT (id) DO NOTHING;

-- Low-Balance Account: Merchant Gamma (Funded with $10.00 USD = 1,000 cents, for overdraft tests)
INSERT INTO accounts (id, user_id, account_type, currency, balance_cents, version)
VALUES (
    'c0000000-0000-0000-0000-000000000003',
    'a0000000-0000-0000-0000-000000000003',
    'WALLET',
    'USD',
    1000,
    0
) ON CONFLICT (id) DO NOTHING;
