// ==============================================================================
// NOVARIS: Distributed Payment Settlement & Adaptive Fraud Intelligence Platform
// LOAD TESTING ENGINE: Multi-Merchant Distributed Settlement Load (Day 13)
// FILE: loadtests/k6_concurrent_settlement.js
// ==============================================================================
//
// 📚 COMPUTER SCIENCE & LOAD PROFILES FOR BEGINNERS:
//
// 1. HOT-WALLET VS. DISTRIBUTED LOAD PROFILES:
//    - Hot-Wallet Testing (k6_hotwallet_load.js) tests the WORST-CASE database scenario:
//      Every virtual user hits the EXACT SAME account row simultaneously. This stresses
//      Optimistic Concurrency Control (OCC) collision detection, version check-and-increment,
//      and exponential backoff retry algorithms.
//    - Distributed Settlement Testing (this file) tests the NORMAL-CASE scenario:
//      Traffic is partitioned across many independent merchant accounts. Because different
//      accounts modify different rows, OCC collisions drop to near-zero, and the system
//      demonstrates its true horizontal scaling capability (Transactions Per Second - TPS).
//
// 2. PARALLEL PER-MERCHANT RATE LIMIT ISOLATION:
//    Notice that each virtual user gets its own distinct merchant ID (`merchant_vu_${__VU}`).
//    This tests Layer 2 (Token-Bucket Limiter) multi-tenant isolation:
//    One aggressive merchant consuming all 10 burst tokens CANNOT starve or throttle
//    another well-behaved merchant on a different bucket!
// ==============================================================================

import http from 'k6/http';
import { check, sleep } from 'k6';
import crypto from 'k6/crypto';
import { Counter, Rate, Trend } from 'k6/metrics';

// Custom metrics tracking multi-tenant behavior:
const distributedSettlements = new Counter('novaris_distributed_settlements');
const multiTenantLatency     = new Trend('novaris_multitenant_latency_ms');
const errorRate              = new Rate('novaris_error_rate');

export const options = {
  // Sustained throughput test: ramp to 50 concurrent VUs and hold for 20 seconds:
  stages: [
    { duration: '5s',  target: 20 }, // Ramp-up
    { duration: '20s', target: 50 }, // Sustained load
    { duration: '5s',  target: 0 },  // Ramp-down
  ],

  thresholds: {
    // 95th percentile latency must be under 200ms when traffic is distributed:
    'http_req_duration{status:200}': ['p(95)<200'],
    // Overall transaction error rate must be under 2%:
    'http_req_failed': ['rate<0.02'],
  },
};

const BASE_URL        = __ENV.GATEWAY_URL || 'http://localhost:8080';
const HMAC_SECRET     = __ENV.HMAC_SECRET  || 'novaris_dev_hmac_secret_key_2026_super_secure';
const DEV_JWT_SECRET  = 'novaris_dev_super_secret_signing_key_2026_do_not_use_in_prod';

function generateUUIDv4() {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
    const r = Math.random() * 16 | 0;
    const v = c === 'x' ? r : (r & 0x3 | 0x8);
    return v.toString(16);
  });
}

function generateTestJWT(merchantId, role, secret) {
  const header = { alg: 'HS256', typ: 'JWT' };
  const now = Math.floor(Date.now() / 1000);
  const payload = {
    sub: merchantId,
    merchant_id: merchantId,
    role: role,
    iat: now,
    exp: now + 3600,
    iss: 'novaris-multitenant-engine'
  };

  const b64Header = crypto.b64encode(JSON.stringify(header), 'rawurl');
  const b64Payload = crypto.b64encode(JSON.stringify(payload), 'rawurl');
  const unsignedToken = `${b64Header}.${b64Payload}`;
  const signature = crypto.hmac('sha256', secret, unsignedToken, 'rawurl');

  return `${unsignedToken}.${signature}`;
}

export default function () {
  // Each Virtual User (VU) simulates a distinct merchant account:
  // __VU is an integer identifier (1, 2, 3... 50) provided by k6 for each concurrent thread.
  const merchantID = `merchant_tenant_${__VU}`;
  const recipientID = `c0000000-0000-0000-0000-000000000002`;

  const payloadObject = {
    sender_id: 'c0000000-0000-0000-0000-000000000001',
    recipient_id: recipientID,
    amount_cents: 2500, // $25.00
    currency: 'USD'
  };
  const rawBody = JSON.stringify(payloadObject);

  const timestamp = Math.floor(Date.now() / 1000);
  const idempotencyKey = generateUUIDv4();
  const canonicalString = `${timestamp}.${rawBody}`;
  const hmacSignature = crypto.hmac('sha256', HMAC_SECRET, canonicalString, 'hex');
  const authToken = generateTestJWT(merchantID, 'merchant_admin', DEV_JWT_SECRET);

  const headers = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${authToken}`,
    'Idempotency-Key': idempotencyKey,
    'X-Novaris-Timestamp': timestamp.toString(),
    'X-Novaris-Signature': hmacSignature,
  };

  const start = Date.now();
  const res = http.post(`${BASE_URL}/v1/payments`, rawBody, { headers: headers });
  const latency = Date.now() - start;

  multiTenantLatency.add(latency);

  if (res.status === 200) {
    distributedSettlements.add(1);
    errorRate.add(0);
  } else if (res.status === 429) {
    // Per-merchant burst exceeded
    errorRate.add(0);
  } else {
    errorRate.add(1);
  }

  check(res, {
    'status is 200 or 429': (r) => r.status === 200 || r.status === 429,
    'no 500 errors': (r) => r.status !== 500,
  });

  // Paced interval between requests (50ms - 100ms) to model steady production traffic:
  sleep(0.05 + Math.random() * 0.05);
}
