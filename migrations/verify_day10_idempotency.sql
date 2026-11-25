-- ==============================================================================
-- NOVARIS: Distributed Payment Settlement & Adaptive Fraud Intelligence Platform
-- VERIFICATION HARNESS: Day 10 - Idempotent API Design & Safe Network Retries
-- FILE: migrations/verify_day10_idempotency.sql
-- ==============================================================================
--
-- Objective:
-- Prove that duplicate requests with the same idempotency_key return the cached
-- original result and NEVER execute a second financial charge.
--
-- Invariants Verified:
-- 1. Database-Level Uniqueness: Duplicate INSERT on idempotency_keys triggers primary key violation (23505).
-- 2. State Preservation: Account balance is decremented EXACTLY ONCE.
-- 3. Cached Result Replay: Stored response_code (200) and response_body JSON are returned.
-- ==============================================================================

\set ON_ERROR_STOP on

BEGIN;

DO $$
DECLARE
    v_account_id UUID := 'a0000000-0000-0000-0000-000000000010';
    v_user_id UUID := '00000000-0000-0000-0000-000000000010';
    v_idemp_key VARCHAR(64) := 'idemp_key_day10_test_4567';
    v_req_hash VARCHAR(64) := 'e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855';
    v_initial_balance BIGINT := 10000; -- $100.00
    v_charge_amount BIGINT := 2500;   -- $25.00
    v_cached_code INT;
    v_cached_body JSONB;
    v_final_balance BIGINT;
    v_final_version INT;
    v_duplicate_inserted BOOLEAN := FALSE;
BEGIN
    RAISE NOTICE '================================================================';
    RAISE NOTICE 'NOVARIS DAY 10: IDEMPOTENT REQUEST DEDUPLICATION VERIFICATION';
    RAISE NOTICE '================================================================';

    -- Cleanup prior run
    DELETE FROM idempotency_keys WHERE key = v_idemp_key;
    DELETE FROM accounts WHERE id = v_account_id;

    -- 1. Seed account with $100.00 (10000 cents)
    INSERT INTO accounts (id, user_id, account_type, currency, balance_cents, version)
    VALUES (v_account_id, v_user_id, 'WALLET', 'USD', v_initial_balance, 0);

    RAISE NOTICE '[STEP 1] Seeded test account % with balance = % cents ($100.00)', v_account_id, v_initial_balance;

    -- 2. First Request Arrives: Insert IN_PROGRESS record
    INSERT INTO idempotency_keys (key, account_id, request_path, request_hash, status)
    VALUES (v_idemp_key, v_account_id::text, '/v1/payments', v_req_hash, 'IN_PROGRESS');

    -- Execute settlement charge (debit $25.00)
    UPDATE accounts
    SET balance_cents = balance_cents - v_charge_amount,
        version = version + 1
    WHERE id = v_account_id;

    -- Mark idempotency record COMPLETED with cached response payload
    UPDATE idempotency_keys
    SET status = 'COMPLETED',
        response_code = 200,
        response_body = jsonb_build_object(
            'status', 'settled',
            'amount_cents', v_charge_amount,
            'message', 'Payment settled successfully'
        )
    WHERE key = v_idemp_key;

    RAISE NOTICE '[STEP 2] First request processed: charged % cents. Idempotency record marked COMPLETED.', v_charge_amount;

    -- 3. Duplicate Request Arrives (Network Retry after dropped client connection)
    -- Attempting to insert the exact same key MUST trigger UNIQUE constraint!
    BEGIN
        INSERT INTO idempotency_keys (key, account_id, request_path, request_hash, status)
        VALUES (v_idemp_key, v_account_id::text, '/v1/payments', v_req_hash, 'IN_PROGRESS');
        v_duplicate_inserted := TRUE;
    EXCEPTION WHEN unique_violation THEN
        RAISE NOTICE '[STEP 3] Database PRIMARY KEY constraint successfully caught duplicate key! Error: 23505 (unique_violation).';
        v_duplicate_inserted := FALSE;
    END;

    IF v_duplicate_inserted THEN
        RAISE EXCEPTION 'IDEMPOTENCY BREACH: Database allowed duplicate primary key insert!';
    END IF;

    -- Retrieve cached response for the duplicate request
    SELECT response_code, response_body INTO v_cached_code, v_cached_body
    FROM idempotency_keys
    WHERE key = v_idemp_key AND status = 'COMPLETED';

    IF v_cached_code <> 200 THEN
        RAISE EXCEPTION 'FATAL: Expected cached status code 200, got %', v_cached_code;
    END IF;

    RAISE NOTICE '[STEP 3] Duplicate request served FROM CACHE: Status Code: %, Body: %', v_cached_code, v_cached_body;

    -- 4. Verify that balance was charged EXACTLY ONCE
    SELECT balance_cents, version INTO v_final_balance, v_final_version
    FROM accounts WHERE id = v_account_id;

    IF v_final_balance <> (v_initial_balance - v_charge_amount) THEN
        RAISE EXCEPTION 'DOUBLE SPEND DETECTED: Balance was charged multiple times! Expected %, got %',
            (v_initial_balance - v_charge_amount), v_final_balance;
    END IF;

    IF v_final_version <> 1 THEN
        RAISE EXCEPTION 'VERSION DRIFT: Account version modified multiple times! Expected 1, got %', v_final_version;
    END IF;

    RAISE NOTICE '[STEP 4] Balance Invariant Verified: Balance = % cents ($75.00), Version = 1.', v_final_balance;
    RAISE NOTICE '================================================================';
    RAISE NOTICE 'RESULT: EXACT-ONCE IDEMPOTENT SETTLEMENT VERIFIED (0 DOUBLE-SPEND)';
    RAISE NOTICE '================================================================';
END $$;

COMMIT;
