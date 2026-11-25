// ==============================================================================
// NOVARIS: Distributed Payment Settlement & Adaptive Fraud Intelligence Platform
// LOAD TESTING ENGINE: Hot-Wallet Concurrency & Contention Script (Day 13)
// FILE: loadtests/k6_hotwallet_load.js
// ==============================================================================
//
// 📚 COMPUTER SCIENCE & LOAD TESTING PRINCIPLES FOR BEGINNERS:
//
// 1. WHAT IS LOAD TESTING?
//    Load testing is the practice of simulating real-world and peak-capacity traffic
//    against a software system to observe how it behaves under stress.
//    We seek to answer critical questions:
//      - Does the system maintain data integrity (zero race conditions)?
//      - How does latency degrade as concurrent users (VUs) increase?
//      - What is the maximum throughput (Transactions Per Second - TPS) before saturation?
//      - Does the system fail gracefully (e.g. HTTP 429 / 409) or crash (HTTP 500 / OOM)?
//
// 2. THE "HOT WALLET" CONTENTION PROBLEM:
//    In high-volume payment platforms (e.g. Uber driver payouts, Black Friday flash sales,
//    or central exchange clearing accounts), thousands of transactions hit the EXACT
//    SAME account row in the database simultaneously.
//    - Under naive locking (`SELECT FOR UPDATE`), 100 threads wait in a lock queue,
//      causing connection pool exhaustion and multi-second latency spikes.
//    - Under NOVARIS Optimistic Concurrency Control (OCC), threads check `version = version + 1`.
//      Colliding threads detect version mismatches, abort instantly with zero lock wait,
//      and retry with randomized exponential jitter!
//
// 3. WHAT IS k6?
//    k6 is a modern, developer-centric open-source load testing tool developed by Grafana.
//    It executes load tests written in ES6 JavaScript with a high-performance Go runtime engine,
//    allowing a single machine to simulate thousands of concurrent virtual users (VUs).
// ==============================================================================

import http from 'k6/http';
import { check, sleep } from 'k6';
import crypto from 'k6/crypto';
import { Counter, Rate, Trend } from 'k6/metrics';

// ==============================================================================
// CUSTOM TELEMETRY METRICS
// ==============================================================================
// Custom metrics give granular visibility into financial system behavior:
const settledTransactions = new Counter('novaris_settled_transactions');
const rateLimitedRequests = new Counter('novaris_rate_limited_429');
const occConflicts        = new Counter('novaris_occ_conflicts_409');
const idempotentReplays   = new Counter('novaris_idempotent_replays_200');
const settlementLatency   = new Trend('novaris_settlement_latency_ms');
const successRate         = new Rate('novaris_success_rate');

// ==============================================================================
// LOAD TEST CONFIGURATION & SCENARIOS
// ==============================================================================
export const options = {
  // Define staged ramp-up and ramp-down execution phases:
  stages: [
    { duration: '5s',  target: 10 },  // Phase 1: Warm-up (ramp up to 10 VUs)
    { duration: '15s', target: 50 },  // Phase 2: High Contention (ramp up to 50 concurrent VUs)
    { duration: '10s', target: 100 }, // Phase 3: Extreme Burst (spike to 100 concurrent VUs)
    { duration: '5s',  target: 0 },   // Phase 4: Cool-down (ramp down to 0)
  ],

  // Service Level Objectives (SLOs) & Quality Gates:
  // k6 automatically fails the test run (exit code 1) if any threshold is violated.
  thresholds: {
    // 95% of all successful HTTP requests must complete in under 350ms:
    'http_req_duration{status:200}': ['p(95)<350'],
    // HTTP 500 Internal Server Errors must be strictly 0% (zero tolerance for unhandled crashes):
    'http_req_failed{status:500}': ['rate==0'],
  },
};

// ==============================================================================
// ENVIRONMENT VARIABLES & DEFAULTS
// ==============================================================================
const BASE_URL    = __ENV.GATEWAY_URL || 'http://localhost:8080';
const HMAC_SECRET = __ENV.HMAC_SECRET  || 'novaris_dev_hmac_secret_key_2026_super_secure';
// Shared Hot-Wallet Accounts (Deterministic Day 11 test seed accounts):
const SENDER_ID   = __ENV.SENDER_ID    || 'c0000000-0000-0000-0000-000000000001';
const RECIPIENT_ID = __ENV.RECIPIENT_ID || 'c0000000-0000-0000-0000-000000000002';

// Minimal JWT generation helper for k6:
// Creates an HMAC-SHA256 signed JWT conforming to NOVARIS Layer 1 requirements.
function generateTestJWT(merchantId, role, secret) {
  const header = { alg: 'HS256', typ: 'JWT' };
  const now = Math.floor(Date.now() / 1000);
  const payload = {
    sub: merchantId,
    merchant_id: merchantId,
    role: role,
    iat: now,
    exp: now + 3600, // Valid for 1 hour
    iss: 'novaris-k6-load-engine'
  };

  const encodeBase64Url = (obj) => {
    const jsonStr = JSON.stringify(obj);
    return encoding.b64encode(jsonStr, 'rawurl');
  };

  // Minimal base64url string builder for k6 environment:
  const b64Header = crypto.b64encode(JSON.stringify(header), 'rawurl');
  const b64Payload = crypto.b64encode(JSON.stringify(payload), 'rawurl');
  const unsignedToken = `${b64Header}.${b64Payload}`;
  const signature = crypto.hmac('sha256', secret, unsignedToken, 'rawurl');

  return `${unsignedToken}.${signature}`;
}

// Generate random RFC 4122 UUIDv4 for Idempotency-Key header:
function generateUUIDv4() {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
    const r = Math.random() * 16 | 0;
    const v = c === 'x' ? r : (r & 0x3 | 0x8);
    return v.toString(16);
  });
}

// Static pre-minted dev token fallback matching services/gateway/internal/auth default:
const DEV_JWT_SECRET = 'novaris_dev_super_secret_signing_key_2026_do_not_use_in_prod';

// ==============================================================================
// VIRTUAL USER (VU) EXECUTION ITERATION LOOP
// ==============================================================================
export default function () {
  // 1. Construct the financial settlement request body:
  // Every request transfers 100 cents ($1.00) from the shared hot-wallet account.
  const payloadObject = {
    sender_id: SENDER_ID,
    recipient_id: RECIPIENT_ID,
    amount_cents: 100,
    currency: 'USD'
  };
  const rawBody = JSON.stringify(payloadObject);

  // 2. Generate canonical headers for the 5-layer perimeter:
  const timestamp = Math.floor(Date.now() / 1000);
  const idempotencyKey = generateUUIDv4();

  // Compute canonical HMAC-SHA256 signature over: "<timestamp>.<raw_body>"
  const canonicalString = `${timestamp}.${rawBody}`;
  const hmacSignature = crypto.hmac('sha256', HMAC_SECRET, canonicalString, 'hex');

  // Mint or compute valid Bearer JWT token:
  const authToken = generateTestJWT(SENDER_ID, 'merchant_admin', DEV_JWT_SECRET);

  const headers = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${authToken}`,
    'Idempotency-Key': idempotencyKey,
    'X-Novaris-Timestamp': timestamp.toString(),
    'X-Novaris-Signature': hmacSignature,
  };

  // 3. Dispatch HTTP POST /v1/payments:
  const start = Date.now();
  const res = http.post(`${BASE_URL}/v1/payments`, rawBody, { headers: headers });
  const latency = Date.now() - start;

  settlementLatency.add(latency);

  // 4. Assert responses and track metrics based on outcome:
  if (res.status === 200) {
    // Check if this was a fresh settlement or an idempotent replay
    if (res.headers['Idempotent-Replayed'] === 'true') {
      idempotentReplays.add(1);
    } else {
      settledTransactions.add(1);
    }
    successRate.add(1);
  } else if (res.status === 429) {
    // Layer 2 Token-Bucket throttling kicked in (expected under sudden burst!)
    rateLimitedRequests.add(1);
    successRate.add(1); // Throttling is defensive success (not a crash!)
  } else if (res.status === 409) {
    // Layer 5 OCC version contention (concurrency collision handled gracefully)
    occConflicts.add(1);
    successRate.add(1); // Conflict rejection is an intentional CAS defense!
  } else {
    // Unexpected error (e.g. 500, 401, 422)
    successRate.add(0);
  }

  // Verify the system behaved predictably:
  check(res, {
    'status is 200, 409, or 429': (r) => r.status === 200 || r.status === 409 || r.status === 429,
    'no HTTP 500 server crashes': (r) => r.status !== 500,
    'has JSON response body': (r) => r.headers['Content-Type'] && r.headers['Content-Type'].includes('application/json'),
  });

  // Small random think time (10ms - 50ms) between iterations to simulate real user cadence:
  sleep(0.01 + Math.random() * 0.04);
}
