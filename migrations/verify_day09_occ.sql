-- ==============================================================================
-- NOVARIS: Distributed Payment Settlement & Adaptive Fraud Intelligence Platform
-- VERIFICATION HARNESS: Day 9 - Optimistic Concurrency Control (OCC)
-- FILE: migrations/verify_day09_occ.sql
-- ==============================================================================
--
-- Objective:
-- Prove that conditional updates (UPDATE ... WHERE id = :id AND version = :version)
-- reject conflicting concurrent writes safely (0 rows affected) instead of silently
-- corrupting financial account balances.
--
-- Mathematical Guarantee:
-- If Session 1 reads version 0, and Session 2 reads version 0:
-- Session 1 updates version 0 -> 1: Affects 1 row (SUCCEEDS).
-- Session 2 attempts update version 0 -> 1: Affects 0 rows (REJECTED SAFELY).
-- The double-spend is neutralized at the PostgreSQL engine level!
-- ==============================================================================

\set ON_ERROR_STOP on

BEGIN;

-- ------------------------------------------------------------------------------
-- STEP 1: Seed a Test Account with Version = 0 and Initial Balance = $100.00
-- ------------------------------------------------------------------------------
DO $$
DECLARE
    v_user_id UUID := '00000000-0000-0000-0000-000000000009';
    v_account_id UUID := 'a0000000-0000-0000-0000-000000000009';
BEGIN
    RAISE NOTICE '================================================================';
    RAISE NOTICE 'NOVARIS DAY 9: OPTIMISTIC CONCURRENCY CONTROL (OCC) VERIFICATION';
    RAISE NOTICE '================================================================';

    -- Remove prior test account if exists
    DELETE FROM accounts WHERE id = v_account_id;

    -- Insert seed account
    INSERT INTO accounts (id, user_id, account_type, currency, balance_cents, version)
    VALUES (v_account_id, v_user_id, 'WALLET', 'USD', 10000, 0);

    RAISE NOTICE '[STEP 1] Seeded test account % with balance = 10000 cents ($100.00), version = 0', v_account_id;
END $$;

-- ------------------------------------------------------------------------------
-- STEP 2: Execute First Mutation (Session 1 Wins Race)
-- ------------------------------------------------------------------------------
DO $$
DECLARE
    v_account_id UUID := 'a0000000-0000-0000-0000-000000000009';
    v_rows_affected INT;
    v_new_balance BIGINT;
    v_new_version INT;
BEGIN
    -- Session 1 updates balance from 10000 -> 8000 (debit 2000 cents) expecting version = 0
    UPDATE accounts
    SET balance_cents = balance_cents - 2000,
        version = version + 1,
        updated_at = NOW()
    WHERE id = v_account_id
      AND version = 0;

    GET DIAGNOSTICS v_rows_affected = ROW_COUNT;

    IF v_rows_affected <> 1 THEN
        RAISE EXCEPTION 'FATAL: Session 1 expected 1 row affected, got %', v_rows_affected;
    END IF;

    SELECT balance_cents, version INTO v_new_balance, v_new_version
    FROM accounts WHERE id = v_account_id;

    RAISE NOTICE '[STEP 2] Session 1 update SUCCEEDED! Rows affected: %. New Balance: %, New Version: %',
        v_rows_affected, v_new_balance, v_new_version;
END $$;

-- ------------------------------------------------------------------------------
-- STEP 3: Execute Conflicting Stale Mutation (Session 2 Loses Race)
-- ------------------------------------------------------------------------------
DO $$
DECLARE
    v_account_id UUID := 'a0000000-0000-0000-0000-000000000009';
    v_rows_affected INT;
    v_final_balance BIGINT;
    v_final_version INT;
BEGIN
    -- Session 2 attempts to debit 2000 cents using STALE expected version = 0
    UPDATE accounts
    SET balance_cents = balance_cents - 2000,
        version = version + 1,
        updated_at = NOW()
    WHERE id = v_account_id
      AND version = 0; -- Stale version! Row version is currently 1!

    GET DIAGNOSTICS v_rows_affected = ROW_COUNT;

    -- The Golden OCC Invariant: rows affected MUST be 0!
    IF v_rows_affected <> 0 THEN
        RAISE EXCEPTION 'OCC BREACH: Session 2 should have updated 0 rows, but updated % rows!', v_rows_affected;
    END IF;

    SELECT balance_cents, version INTO v_final_balance, v_final_version
    FROM accounts WHERE id = v_account_id;

    IF v_final_balance <> 8000 THEN
        RAISE EXCEPTION 'DATA CORRUPTION: Balance was overwritten! Expected 8000, got %', v_final_balance;
    END IF;

    IF v_final_version <> 1 THEN
        RAISE EXCEPTION 'VERSION DRIFT: Version was incremented unexpectedly! Expected 1, got %', v_final_version;
    END IF;

    RAISE NOTICE '[STEP 3] Session 2 conflicting write SAFELY REJECTED! Rows affected: 0.';
    RAISE NOTICE '[STEP 3] Verified Account Balance remained 8000 cents ($80.00), Version remained 1.';
    RAISE NOTICE '================================================================';
    RAISE NOTICE 'RESULT: OPTIMISTIC CONCURRENCY CONTROL (OCC) VERIFIED!';
    RAISE NOTICE '================================================================';
END $$;

COMMIT;
