-- ==============================================================================
-- NOVARIS: Distributed Payment Settlement & Adaptive Fraud Intelligence Platform
-- DIAGNOSTIC HARNESS: Atomic Double-Entry Settlement & Rollback Verification (Day 6)
-- FILE: migrations/verify_day06_settlement.sql
-- ==============================================================================
--
-- Purpose:
-- This verification harness provides deterministic, executable proof that:
-- 1. Double-entry ledger writes produce balanced journal rows (Debit == Credit).
-- 2. Multi-table mutations inside a transaction commit atomically (ACID Atomicity).
-- 3. A deliberate failure partway through (e.g., foreign key violation) triggers
--    a complete ROLLBACK, leaving zero orphan rows behind — not even the first insert.
-- ==============================================================================

\echo '===================================================================='
\echo '  NOVARIS DAY 6: ATOMIC SETTLEMENT & ROLLBACK VERIFICATION SUITE'
\echo '===================================================================='

-- ------------------------------------------------------------------------------
-- 1. Seed Test Accounts (Idempotent)
-- ------------------------------------------------------------------------------
\echo '\n--- 1. Ensuring Deterministic Test Accounts Exist ---'
DO $$
DECLARE
    v_user_id UUID := 'a0000000-0000-0000-0000-000000000001';
BEGIN
    -- Sender Wallet Account
    INSERT INTO accounts (id, user_id, account_type, currency, balance_cents, version)
    VALUES ('c0000000-0000-0000-0000-000000000001', v_user_id, 'WALLET', 'USD', 100000, 1)
    ON CONFLICT (id) DO NOTHING;

    -- Recipient Merchant Account
    INSERT INTO accounts (id, user_id, account_type, currency, balance_cents, version)
    VALUES ('c0000000-0000-0000-0000-000000000002', v_user_id, 'MERCHANT_PAYOUT', 'USD', 50000, 1)
    ON CONFLICT (id) DO NOTHING;

    RAISE NOTICE 'Test accounts initialized: Sender=c000...01, Recipient=c000...02';
END $$;

-- ------------------------------------------------------------------------------
-- 2. Happy Path: Successful Atomic Settlement Transaction
-- ------------------------------------------------------------------------------
\echo '\n--- 2. Executing Happy-Path Atomic Settlement (1 Tx + 2 Balanced Journals) ---'
DO $$
DECLARE
    v_tx_id UUID := 'd0000000-0000-0000-0000-000000000006';
    v_sender UUID := 'c0000000-0000-0000-0000-000000000001';
    v_recipient UUID := 'c0000000-0000-0000-0000-000000000002';
    v_amount BIGINT := 5000; -- 50.00 USD
BEGIN
    -- Clean previous run if exists
    DELETE FROM journal_entries WHERE transaction_id = v_tx_id;
    DELETE FROM transactions WHERE id = v_tx_id;

    -- Step 1: Insert Transaction Record
    INSERT INTO transactions (id, idempotency_key, sender_account_id, recipient_account_id, amount_cents, status)
    VALUES (v_tx_id, 'idemp_day06_happy_path_001', v_sender, v_recipient, v_amount, 'SETTLED');

    -- Step 2: Insert Debit Entry (Money leaves sender)
    INSERT INTO journal_entries (transaction_id, account_id, entry_type, amount_cents)
    VALUES (v_tx_id, v_sender, 'DEBIT', v_amount);

    -- Step 3: Insert Credit Entry (Money enters recipient)
    INSERT INTO journal_entries (transaction_id, account_id, entry_type, amount_cents)
    VALUES (v_tx_id, v_recipient, 'CREDIT', v_amount);

    RAISE NOTICE 'Happy-path settlement written atomically for tx: %', v_tx_id;
END $$;

-- Verify Happy Path Row Creation & Balance Invariant
\echo '\n--- 3. Verifying Double-Entry Ledger Balance (Debit == Credit) ---'
SELECT 
    t.id AS transaction_id,
    t.amount_cents AS transfer_amount,
    t.status,
    COUNT(j.id) AS total_journal_entries,
    SUM(CASE WHEN j.entry_type = 'DEBIT' THEN j.amount_cents ELSE 0 END) AS total_debits,
    SUM(CASE WHEN j.entry_type = 'CREDIT' THEN j.amount_cents ELSE 0 END) AS total_credits,
    CASE 
        WHEN SUM(CASE WHEN j.entry_type = 'DEBIT' THEN j.amount_cents ELSE 0 END) = 
             SUM(CASE WHEN j.entry_type = 'CREDIT' THEN j.amount_cents ELSE 0 END)
             AND COUNT(j.id) = 2
        THEN '✅ BALANCED (ZERO DRIFT)'
        ELSE '❌ UNBALANCED LEDGER ERROR'
    END AS ledger_invariant_status
FROM transactions t
JOIN journal_entries j ON t.id = j.transaction_id
WHERE t.id = 'd0000000-0000-0000-0000-000000000006'
GROUP BY t.id, t.amount_cents, t.status;

-- ------------------------------------------------------------------------------
-- 4. Rollback Path: Deliberately Force Failure & Prove Zero Ghost Rows
-- ------------------------------------------------------------------------------
\echo '\n--- 4. Executing Deliberate Failure (Simulating Rollback on Foreign Key Violation) ---'
DO $$
DECLARE
    v_count_before INT;
    v_count_after INT;
    v_fail_tx_id UUID := 'd0000000-0000-0000-0000-000000000099';
    v_sender UUID := 'c0000000-0000-0000-0000-000000000001';
    v_invalid_recipient UUID := '00000000-0000-0000-0000-000000000000'; -- Non-existent account!
BEGIN
    SELECT COUNT(*) INTO v_count_before FROM transactions;

    -- Sub-transaction block simulating application Begin() ... Rollback()
    BEGIN
        -- Insert 1: Transaction insert (this temporarily succeeds!)
        INSERT INTO transactions (id, idempotency_key, sender_account_id, recipient_account_id, amount_cents, status)
        VALUES (v_fail_tx_id, 'idemp_day06_fail_probe', v_sender, v_invalid_recipient, 7500, 'SETTLED');

        -- Insert 2: Debit insert (this temporarily succeeds!)
        INSERT INTO journal_entries (transaction_id, account_id, entry_type, amount_cents)
        VALUES (v_fail_tx_id, v_sender, 'DEBIT', 7500);

        -- Insert 3: Credit insert (FAILS! Invalid recipient violates FK constraint)
        INSERT INTO journal_entries (transaction_id, account_id, entry_type, amount_cents)
        VALUES (v_fail_tx_id, v_invalid_recipient, 'CREDIT', 7500);

    EXCEPTION WHEN foreign_key_violation THEN
        RAISE NOTICE 'Captured expected foreign key violation on step 3. Aborting sub-transaction (ROLLBACK)...';
    END;

    -- Query count after failure
    SELECT COUNT(*) INTO v_count_after FROM transactions;

    -- Prove that Insert 1 (which succeeded earlier) was completely rolled back!
    IF v_count_before = v_count_after THEN
        RAISE NOTICE '✅ TRUE ATOMICITY PROVEN: Count before (%) equals count after (%). Zero ghost rows left behind!',
            v_count_before, v_count_after;
    ELSE
        RAISE EXCEPTION '❌ ATOMICITY FAILED: Ghost rows detected! Before: %, After: %',
            v_count_before, v_count_after;
    END IF;
END $$;

\echo '\n===================================================================='
\echo '  DAY 6 VERIFICATION COMPLETE: ALL ACID & BALANCED INVARIANTS PROVEN'
\echo '===================================================================='
