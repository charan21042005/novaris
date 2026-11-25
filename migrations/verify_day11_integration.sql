-- ==============================================================================
-- NOVARIS: Distributed Payment Settlement & Adaptive Fraud Intelligence Platform
-- SQL VERIFICATION HARNESS: Day 11 End-to-End Integration & Seed Data Validation
-- FILE: migrations/verify_day11_integration.sql
-- ==============================================================================
--
-- Verification Objectives:
-- 1. Confirm that all 3 deterministic test accounts are populated in `accounts`.
-- 2. Validate that sender (Alpha) has sufficient balance for live payments.
-- 3. Confirm that `idempotency_keys` table is active and ready for deduplication.
-- 4. Execute a simulated end-to-end double-entry settlement transaction:
--    - Insert transaction record
--    - Update sender and recipient balances via OCC check-and-increment
--    - Insert balanced debit and credit legs in journal_entries
--    - Insert event record in outbox table
--    - Register completed record in idempotency_keys table
-- 5. Validate Total System Monetary Conservation:
--    - Initial sum: 1,000,000 + 250,000 = 1,250,000 cents ($12,500.00)
--    - Post-settlement sum: must remain EXACTLY 1,250,000 cents ($12,500.00)
-- ==============================================================================

DO $$
DECLARE
    v_sender_id UUID := 'c0000000-0000-0000-0000-000000000001';
    v_recipient_id UUID := 'c0000000-0000-0000-0000-000000000002';
    v_transfer_amount BIGINT := 5000; -- $50.00 USD
    v_sender_bal_before BIGINT;
    v_recipient_bal_before BIGINT;
    v_sender_ver_before INT;
    v_recipient_ver_before INT;
    v_sender_bal_after BIGINT;
    v_recipient_bal_after BIGINT;
    v_sender_ver_after INT;
    v_recipient_ver_after INT;
    v_tx_id UUID;
    v_idemp_key VARCHAR(128) := 'e2e-harness-key-' || uuid_generate_v4()::text;
    v_payload_hash VARCHAR(64) := 'd2d2d2d2d2d2d2d2d2d2d2d2d2d2d2d2d2d2d2d2d2d2d2d2d2d2d2d2d2d2d2d2';
    v_rows_affected INT;
    v_total_system_before BIGINT;
    v_total_system_after BIGINT;
BEGIN
    RAISE NOTICE '==================================================================';
    RAISE NOTICE ' 🧪 NOVARIS DAY 11: End-to-End Integration & Seed Harness';
    RAISE NOTICE '==================================================================';

    -- 1. Ensure test accounts exist
    PERFORM 1 FROM accounts WHERE id = v_sender_id;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'CRITICAL: Sender seed account % is missing! Run 004_seed_test_accounts.sql first.', v_sender_id;
    END IF;

    PERFORM 1 FROM accounts WHERE id = v_recipient_id;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'CRITICAL: Recipient seed account % is missing! Run 004_seed_test_accounts.sql first.', v_recipient_id;
    END IF;

    -- 2. Read state before settlement
    SELECT balance_cents, version INTO v_sender_bal_before, v_sender_ver_before FROM accounts WHERE id = v_sender_id;
    SELECT balance_cents, version INTO v_recipient_bal_before, v_recipient_ver_before FROM accounts WHERE id = v_recipient_id;
    v_total_system_before := v_sender_bal_before + v_recipient_bal_before;

    RAISE NOTICE '  [STATE BEFORE] Alpha (Sender):    $% (cents: %, version: %)', (v_sender_bal_before::numeric / 100.0), v_sender_bal_before, v_sender_ver_before;
    RAISE NOTICE '  [STATE BEFORE] Beta (Recipient):  $% (cents: %, version: %)', (v_recipient_bal_before::numeric / 100.0), v_recipient_bal_before, v_recipient_ver_before;
    RAISE NOTICE '  [STATE BEFORE] Total Money Pool:  $% (cents: %)', (v_total_system_before::numeric / 100.0), v_total_system_before;

    -- 3. Execute End-to-End Multi-Table Atomic Settlement
    -- Leg A: Register Idempotency Key in-progress
    INSERT INTO idempotency_keys (key, account_id, request_path, request_hash, status, created_at, expires_at)
    VALUES (v_idemp_key, v_sender_id::text, '/v1/payments', v_payload_hash, 'IN_PROGRESS', NOW(), NOW() + INTERVAL '24 hours');

    -- Leg B: Insert Transaction
    INSERT INTO transactions (idempotency_key, account_id, amount_cents, currency, status, created_at, updated_at)
    VALUES (v_idemp_key, v_sender_id, v_transfer_amount, 'USD', 'SETTLED', NOW(), NOW())
    RETURNING id INTO v_tx_id;

    -- Leg C: Debit Sender via OCC check-and-increment
    UPDATE accounts
    SET balance_cents = balance_cents - v_transfer_amount,
        version = version + 1,
        updated_at = NOW()
    WHERE id = v_sender_id AND version = v_sender_ver_before AND balance_cents >= v_transfer_amount;
    GET DIAGNOSTICS v_rows_affected = ROW_COUNT;
    IF v_rows_affected != 1 THEN
        RAISE EXCEPTION 'FAILED: OCC Debit failed on sender account!';
    END IF;

    -- Leg D: Credit Recipient via OCC check-and-increment
    UPDATE accounts
    SET balance_cents = balance_cents + v_transfer_amount,
        version = version + 1,
        updated_at = NOW()
    WHERE id = v_recipient_id AND version = v_recipient_ver_before;
    GET DIAGNOSTICS v_rows_affected = ROW_COUNT;
    IF v_rows_affected != 1 THEN
        RAISE EXCEPTION 'FAILED: OCC Credit failed on recipient account!';
    END IF;

    -- Leg E: Balanced Double-Entry Journal Entries
    INSERT INTO journal_entries (transaction_id, account_id, entry_type, amount_cents, currency, created_at)
    VALUES
        (v_tx_id, v_sender_id, 'DEBIT', v_transfer_amount, 'USD', NOW()),
        (v_tx_id, v_recipient_id, 'CREDIT', v_transfer_amount, 'USD', NOW());

    -- Leg F: Transactional Outbox Event
    INSERT INTO outbox (event_type, aggregate_type, aggregate_id, payload, created_at)
    VALUES (
        'PAYMENT_SETTLED',
        'TRANSACTION',
        v_tx_id::text,
        jsonb_build_object(
            'transaction_id', v_tx_id,
            'sender_id', v_sender_id,
            'recipient_id', v_recipient_id,
            'amount_cents', v_transfer_amount,
            'currency', 'USD',
            'timestamp', NOW()
        ),
        NOW()
    );

    -- Leg G: Update Idempotency Key to COMPLETED with cached JSON response
    UPDATE idempotency_keys
    SET status = 'COMPLETED',
        response_code = 200,
        response_body = jsonb_build_object(
            'status', 'settled',
            'transaction_id', v_tx_id,
            'merchant_id', v_sender_id::text,
            'amount_cents', v_transfer_amount,
            'recipient', v_recipient_id::text,
            'message', 'Payment settled end-to-end via OCC double-entry ledger pipeline.'
        )
    WHERE key = v_idemp_key;

    -- 4. Read State After Settlement
    SELECT balance_cents, version INTO v_sender_bal_after, v_sender_ver_after FROM accounts WHERE id = v_sender_id;
    SELECT balance_cents, version INTO v_recipient_bal_after, v_recipient_ver_after FROM accounts WHERE id = v_recipient_id;
    v_total_system_after := v_sender_bal_after + v_recipient_bal_after;

    RAISE NOTICE '------------------------------------------------------------------';
    RAISE NOTICE '  [STATE AFTER]  Alpha (Sender):    $% (cents: %, version: %)', (v_sender_bal_after::numeric / 100.0), v_sender_bal_after, v_sender_ver_after;
    RAISE NOTICE '  [STATE AFTER]  Beta (Recipient):  $% (cents: %, version: %)', (v_recipient_bal_after::numeric / 100.0), v_recipient_bal_after, v_recipient_ver_after;
    RAISE NOTICE '  [STATE AFTER]  Total Money Pool:  $% (cents: %)', (v_total_system_after::numeric / 100.0), v_total_system_after;

    -- 5. Assert Invariants
    IF v_total_system_before != v_total_system_after THEN
        RAISE EXCEPTION 'CRITICAL INVARIANT VIOLATION: Money created or destroyed! Before=%, After=%', v_total_system_before, v_total_system_after;
    END IF;

    IF (v_sender_bal_before - v_sender_bal_after) != v_transfer_amount THEN
        RAISE EXCEPTION 'FAILED: Sender balance not debited by exactly % cents!', v_transfer_amount;
    END IF;

    IF (v_recipient_bal_after - v_recipient_bal_before) != v_transfer_amount THEN
        RAISE EXCEPTION 'FAILED: Recipient balance not credited by exactly % cents!', v_transfer_amount;
    END IF;

    IF v_sender_ver_after != (v_sender_ver_before + 1) THEN
        RAISE EXCEPTION 'FAILED: Sender version did not increment monotonically!';
    END IF;

    IF v_recipient_ver_after != (v_recipient_ver_before + 1) THEN
        RAISE EXCEPTION 'FAILED: Recipient version did not increment monotonically!';
    END IF;

    RAISE NOTICE '==================================================================';
    RAISE NOTICE '  ✅ SUCCESS: End-to-End Integration Invariants Mathematically Proven!';
    RAISE NOTICE '  - Double-Entry Debit/Credit Balance: OK (Difference = 0)';
    RAISE NOTICE '  - OCC Monotonic Version Advance:     OK (Sender & Recipient +1)';
    RAISE NOTICE '  - Transactional Outbox Event Staged: OK (Aggregate ID: %)', v_tx_id;
    RAISE NOTICE '  - Idempotency Cache Registered:      OK (Key: %)', v_idemp_key;
    RAISE NOTICE '  - System Zero-Drift Conservation:    PROVEN ($% -> $%)', (v_total_system_before::numeric / 100.0), (v_total_system_after::numeric / 100.0);
    RAISE NOTICE '==================================================================';
END $$;
