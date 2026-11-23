// ==============================================================================
// NOVARIS: Distributed Payment Settlement & Adaptive Fraud Intelligence Platform
// MODULE: Cryptographic HMAC-SHA256 Request Signing & Replay Defense (Day 3)
// PACKAGE: internal/auth/hmac.go
// ==============================================================================
//
// Theoretical Architecture & Purpose:
// 1. Day 2's JWT verified WHO is calling (Merchant Authentication / Identity).
// 2. Day 3's HMAC-SHA256 verifies WHAT was sent (Payload Integrity & Non-Repudiation):
//    - Proves that the HTTP request body was not altered or tampered with in transit.
//    - Concretely: If an attacker intercepts a legitimate JWT-authenticated payment
//      for $50.00 and changes the amount to $50,000.00, JWT verification alone passes!
//      HMAC-SHA256 catches the tampering immediately because the hash doesn't match.
// 3. Defense against Replay Attacks:
//    - Incorporates an X-Novaris-Timestamp header into the canonical signed payload.
//    - Enforces a strict ±300-second (5-minute) validation window against server time.
//    - Stale or future-dated requests are rejected before heavy processing occurs.
// 4. Defense against Side-Channel Timing Attacks:
//    - Uses hmac.Equal (constant-time byte comparison) instead of naive '==' comparison.
//    - Prevents attackers from iteratively deducing signature bytes via nanosecond latency.
// 5. Canonicalization Safety:
//    - Signs the RAW, untouched request body bytes directly from the network wire.
//    - Avoids re-serializing JSON, preventing whitespace, key-order, or float format drift.
// ==============================================================================

package auth

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

// Header constants defined by the NOVARIS API Gateway specification.
const (
	// HeaderNovarisSignature carries the hex-encoded HMAC-SHA256 signature.
	HeaderNovarisSignature = "X-Novaris-Signature"

	// HeaderNovarisTimestamp carries the UNIX epoch timestamp (in seconds) of request generation.
	HeaderNovarisTimestamp = "X-Novaris-Timestamp"
)

// ReplayWindowSeconds defines the maximum allowed clock drift/age for request timestamps.
// 300 seconds = 5 minutes.
const ReplayWindowSeconds int64 = 300

// DefaultHMACSecret provides a fallback secret key for local development.
// In production, HMAC_SECRET must be supplied via a secure environment variable or KMS.
const DefaultHMACSecret = "novaris_insecure_dev_hmac_secret_change_in_production_2026"

// GetHMACSecret retrieves the HMAC shared secret key from the HMAC_SECRET environment variable,
// falling back to DefaultHMACSecret if unset.
func GetHMACSecret() []byte {
	secret := os.Getenv("HMAC_SECRET")
	if secret == "" {
		return []byte(DefaultHMACSecret)
	}
	return []byte(secret)
}

// BuildCanonicalPayload constructs the exact canonical string format required for signing.
// Canonical format: "<timestamp>.<raw_body>"
//
// CANONICALIZATION GUARANTEE:
// By binding the timestamp directly with the raw byte sequence of the body, we guarantee:
// 1. Neither the timestamp nor the body can be tampered with independently.
// 2. The exact byte sequence received over the wire is what gets validated.
func BuildCanonicalPayload(timestamp int64, rawBody []byte) []byte {
	tsStr := strconv.FormatInt(timestamp, 10)
	buf := make([]byte, 0, len(tsStr)+1+len(rawBody))
	buf = append(buf, tsStr...)
	buf = append(buf, '.')
	buf = append(buf, rawBody...)
	return buf
}

// ComputeHMACSignature calculates a cryptographic HMAC-SHA256 signature over the canonical payload.
// Returns the resulting signature as a lowercase hexadecimal string.
func ComputeHMACSignature(rawBody []byte, timestamp int64, secret []byte) string {
	mac := hmac.New(sha256.New, secret)
	canonical := BuildCanonicalPayload(timestamp, rawBody)
	mac.Write(canonical)
	return hex.EncodeToString(mac.Sum(nil))
}

// VerifyHMACSignature validates whether a provided signature matches the recomputed HMAC
// for a given body, timestamp, and secret key, including replay window checks.
func VerifyHMACSignature(rawBody []byte, timestampStr string, providedSig string, secret []byte) (bool, error) {
	if timestampStr == "" {
		return false, errors.New("missing timestamp header")
	}
	if providedSig == "" {
		return false, errors.New("missing signature header")
	}

	// 1. Parse and validate the timestamp integer.
	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return false, fmt.Errorf("invalid timestamp format: %w", err)
	}

	// 2. REPLAY PROTECTION CHECK:
	// Verify that the request was generated within our allowed clock drift window (±300 seconds).
	now := time.Now().Unix()
	age := now - timestamp
	if age > ReplayWindowSeconds || age < -ReplayWindowSeconds {
		return false, fmt.Errorf("request timestamp outside allowed window (age: %d seconds, allowed: ±%d)", age, ReplayWindowSeconds)
	}

	// 3. CANONICALIZATION & RECOMPUTATION:
	// Recompute expected HMAC-SHA256 signature using the exact raw bytes and timestamp.
	expectedSig := ComputeHMACSignature(rawBody, timestamp, secret)

	// 4. TIMING-ATTACK DEFENSE:
	// Use hmac.Equal (constant-time byte comparison) instead of naive string '=='.
	// Standard '==' returns false on the first differing byte, allowing attackers
	// to measure microsecond/nanosecond response latency to iteratively guess valid signatures.
	// hmac.Equal takes constant time regardless of where or whether a mismatch occurs.
	if !hmac.Equal([]byte(expectedSig), []byte(providedSig)) {
		return false, errors.New("invalid signature: payload hash mismatch")
	}

	return true, nil
}

// HMACMiddleware returns an HTTP middleware that intercepts requests, reads the raw body,
// checks timestamp age, validates the HMAC-SHA256 signature, and restores the request body.
// If verification fails, it responds immediately with HTTP 401 Unauthorized.
func HMACMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only enforce HMAC signing on state-changing or authenticated mutation requests.
		// For Day 3, this protects payment endpoints (e.g., POST/PUT/PATCH).
		sig := r.Header.Get(HeaderNovarisSignature)
		tsStr := r.Header.Get(HeaderNovarisTimestamp)

		if sig == "" || tsStr == "" {
			respondHMACUnauthorized(w, "missing required signature headers (X-Novaris-Signature, X-Novaris-Timestamp)")
			return
		}

		// CHEAP FAIL-FAST CHECK:
		// Validate timestamp format and window FIRST before reading or hashing body bytes.
		// Prevents CPU exhaustion from computing hashes on obviously expired/replayed requests.
		timestamp, err := strconv.ParseInt(tsStr, 10, 64)
		if err != nil {
			respondHMACUnauthorized(w, "invalid X-Novaris-Timestamp format: must be integer UNIX epoch")
			return
		}

		now := time.Now().Unix()
		age := now - timestamp
		if age > ReplayWindowSeconds || age < -ReplayWindowSeconds {
			respondHMACUnauthorized(w, fmt.Sprintf("request timestamp outside allowed window (age: %ds, max: ±%ds)", age, ReplayWindowSeconds))
			return
		}

		// CANONICALIZATION: Read raw untouched body bytes directly from the network stream.
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, `{"error":"bad_request","message":"unable to read request body: %v"}`+"\n", err)
			return
		}

		// STREAM RESTORATION:
		// In Go, reading r.Body consumes the io.Reader stream to EOF.
		// If we don't replace it, downstream business handlers (like paymentsHandler)
		// will encounter an empty body when they try to read or parse JSON!
		// We restore r.Body with an io.NopCloser wrapping a fresh bytes.Reader.
		r.Body = io.NopCloser(bytes.NewReader(bodyBytes))

		// Cryptographic HMAC Verification
		secret := GetHMACSecret()
		valid, err := VerifyHMACSignature(bodyBytes, tsStr, sig, secret)
		if err != nil || !valid {
			respondHMACUnauthorized(w, fmt.Sprintf("HMAC signature verification failed: %v", err))
			return
		}

		// HMAC verification successful. Pass execution to the next handler in the pipeline.
		next.ServeHTTP(w, r)
	}
}

// respondHMACUnauthorized writes a standardized RFC 7807 compatible JSON 401 response.
func respondHMACUnauthorized(w http.ResponseWriter, reason string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusUnauthorized)
	fmt.Fprintf(w, `{"error":"unauthorized","auth_type":"hmac_sha256","message":%q}`+"\n", reason)
}
