-- ==============================================================================
-- NOVARIS: Schema Verification & Diagnostic Harness (Day 5)
-- FILE: migrations/verify_day05_schema.sql
-- ==============================================================================
-- Tests:
--   1. Referential integrity across accounts -> transactions -> fraud_reviews & feature_snapshots.
--   2. Strict enforcement of CHECK (verdict IN ('CONFIRMED_FRAUD', 'FALSE_POSITIVE')).
--   3. Querying inside JSONB feature_data using ->, ->>, and @> operators.
--   4. Query plan inspection (EXPLAIN) verifying the composite B-Tree index scan.
-- ==============================================================================

\echo '--- 1. Confirming Table Registration ---'
SELECT table_name 
FROM information_schema.tables 
WHERE table_schema = 'public' 
  AND table_name IN ('accounts', 'transactions', 'journal_entries', 'outbox', 'fraud_reviews', 'feature_snapshots')
ORDER BY table_name;

\echo '--- 2. Seeding Deterministic Test Account & Transaction ---'
DO $$
DECLARE
    v_user_id UUID := 'a0000000-0000-0000-0000-000000000001';
    v_sender_account_id UUID;
    v_recipient_account_id UUID;
    v_tx_id UUID;
    v_analyst_id UUID := 'b0000000-0000-0000-0000-000000000099';
BEGIN
    -- 1. Create or get test sender account
    INSERT INTO accounts (id, user_id, account_type, currency, balance_cents, version)
    VALUES ('c0000000-0000-0000-0000-000000000001', v_user_id, 'WALLET', 'USD', 100000, 1)
    ON CONFLICT (id) DO NOTHING;
    v_sender_account_id := 'c0000000-0000-0000-0000-000000000001';

    -- 2. Create or get test recipient account
    INSERT INTO accounts (id, user_id, account_type, currency, balance_cents, version)
    VALUES ('c0000000-0000-0000-0000-000000000002', v_user_id, 'MERCHANT_PAYOUT', 'USD', 50000, 1)
    ON CONFLICT (id) DO NOTHING;
    v_recipient_account_id := 'c0000000-0000-0000-0000-000000000002';

    -- 3. Create or get test transaction
    INSERT INTO transactions (id, idempotency_key, sender_account_id, recipient_account_id, amount_cents, status, fraud_risk_score, risk_factors)
    VALUES (
        'd0000000-0000-0000-0000-000000000001',
        'idemp_day05_verification_probe_001',
        v_sender_account_id,
        v_recipient_account_id,
        25000,
        'REVIEW',
        65,
        '{"velocity_score": 0.85, "geo_distance_miles": 1200}'::jsonb
    )
    ON CONFLICT (idempotency_key) DO UPDATE SET status = 'REVIEW'
    RETURNING id INTO v_tx_id;

    -- 4. Insert Verified Fraud Review Row
    INSERT INTO fraud_reviews (id, transaction_id, analyst_id, verdict, notes, reviewed_at)
    VALUES (
        'e0000000-0000-0000-0000-000000000001',
        v_tx_id,
        v_analyst_id,
        'CONFIRMED_FRAUD',
        'Abnormal geo-hop velocity: transaction in Tokyo 15 minutes after card swipe in London. Device IP blacklisted.',
        NOW()
    )
    ON CONFLICT (id) DO NOTHING;

    -- 5. Insert Rich Feature Snapshot Row
    INSERT INTO feature_snapshots (id, account_id, transaction_id, feature_data, computed_at)
    VALUES (
        'f0000000-0000-0000-0000-000000000001',
        v_sender_account_id,
        v_tx_id,
        '{
            "velocity_1m": 4,
            "velocity_5m": 14,
            "velocity_1h": 42,
            "amount_ratio_30d": 5.82,
            "geo_hop_speed_mph": 4800.0,
            "ip_country": "JP",
            "device_trust_score": 0.12,
            "card_country": "GB",
            "is_vpn": true
        }'::jsonb,
        NOW()
    )
    ON CONFLICT (id) DO NOTHING;
END $$;

\echo '--- 3. Verifying Seeded fraud_reviews Entry ---'
SELECT id, transaction_id, analyst_id, verdict, notes, reviewed_at 
FROM fraud_reviews 
WHERE id = 'e0000000-0000-0000-0000-000000000001';

\echo '--- 4. Querying Inside JSONB (feature_data Operators ->, ->>, @>) ---'
SELECT 
    id,
    account_id,
    feature_data->>'ip_country' AS extracted_country,
    (feature_data->>'velocity_5m')::int AS extracted_velocity_5m,
    (feature_data->>'is_vpn')::boolean AS is_vpn,
    computed_at
FROM feature_snapshots
WHERE feature_data @> '{"is_vpn": true}'::jsonb;

\echo '--- 5. Inspecting Query Plan for Composite Index (account_id, computed_at DESC) ---'
EXPLAIN (COSTS OFF)
SELECT id, feature_data, computed_at
FROM feature_snapshots
WHERE account_id = 'c0000000-0000-0000-0000-000000000001'
ORDER BY computed_at DESC
LIMIT 10;

\echo '--- 6. Verifying CHECK Constraint Rejection on Invalid Verdict ---'
-- Attempting to insert an invalid verdict must fail with check_violation (23514)
DO $$
BEGIN
    BEGIN
        INSERT INTO fraud_reviews (transaction_id, analyst_id, verdict)
        VALUES ('d0000000-0000-0000-0000-000000000001', 'b0000000-0000-0000-0000-000000000099', 'MAYBE_FRAUD');
        RAISE EXCEPTION 'CHECK constraint failed to reject invalid verdict!';
    EXCEPTION WHEN check_violation THEN
        RAISE NOTICE 'SUCCESS: Database rejected invalid verdict with check_violation as designed.';
    END;
END $$;
