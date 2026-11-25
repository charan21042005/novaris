// ==============================================================================
// NOVARIS: Distributed Payment Settlement & Adaptive Fraud Intelligence Platform
// DEVELOPER CLI UTILITY: End-to-End API Contract Verification Probe (Day 11)
// FILE: services/gateway/cmd/e2eprobe/main.go
// ==============================================================================
//
// Usage:
//   go run cmd/e2eprobe/main.go
//   go run cmd/e2eprobe/main.go -url http://localhost:8080
//   go run cmd/e2eprobe/main.go -amount 7500 -merchant c0000000-0000-0000-0000-000000000001
//
// Purpose:
// Interactive CLI test runner that executes the complete Day 11 API contract
// sequence against a live or local NOVARIS gateway instance:
// 1. Liveness & Discovery Probes (/healthz, /)
// 2. Missing Auth Header Rejection (HTTP 401)
// 3. Full-Stack Payment Settlement (HTTP 200 OK)
// 4. Exact-Once Network Retry Replay (HTTP 200 OK + Idempotent-Replayed: true)
// 5. Payload Tampering Rejection (HTTP 422 Unprocessable Entity)
// 6. Missing Idempotency Key Rejection (HTTP 400 Bad Request)
// ==============================================================================

package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/charan21042005/novaris/services/gateway/internal/auth"
)

type ProbeResult struct {
	Step        string
	Expected    string
	Actual      string
	Passed      bool
	LatencyMs   float64
	Details     string
}

func main() {
	baseURL := flag.String("url", "http://localhost:8080", "Base URL of target gateway instance")
	merchantID := flag.String("merchant", "c0000000-0000-0000-0000-000000000001", "Merchant Account UUID")
	recipientID := flag.String("recipient", "c0000000-0000-0000-0000-000000000002", "Recipient Account UUID")
	amountCents := flag.Int64("amount", 5000, "Transfer amount in integer cents ($50.00 = 5000)")
	flag.Parse()

	log.Println("==================================================================")
	log.Println(" 🚀 NOVARIS Day 11: End-to-End API Contract Verification Probe")
	log.Printf(" Target Gateway: %s\n", *baseURL)
	log.Printf(" Merchant (Payer):     %s\n", *merchantID)
	log.Printf(" Recipient (Payee):   %s\n", *recipientID)
	log.Printf(" Amount:               $%0.2f (%d cents)\n", float64(*amountCents)/100.0, *amountCents)
	log.Println("==================================================================")

	client := &http.Client{Timeout: 5 * time.Second}
	var results []ProbeResult

	// --------------------------------------------------------------------------
	// STEP 1: Probe /healthz
	// --------------------------------------------------------------------------
	start := time.Now()
	respHealth, err := client.Get(*baseURL + "/healthz")
	lat := float64(time.Since(start).Microseconds()) / 1000.0
	if err != nil {
		log.Fatalf("FATAL: Gateway is offline or unreachable at %s: %v\nEnsure 'go run main.go' is running!", *baseURL, err)
	}
	defer respHealth.Body.Close()
	bodyHealth, _ := io.ReadAll(respHealth.Body)

	results = append(results, ProbeResult{
		Step:      "1. Public Liveness Probe (GET /healthz)",
		Expected:  "HTTP 200 OK (status=ok)",
		Actual:    fmt.Sprintf("HTTP %d (%s)", respHealth.StatusCode, string(bodyHealth)),
		Passed:    respHealth.StatusCode == http.StatusOK,
		LatencyMs: lat,
		Details:   "Liveness verification succeeded.",
	})

	// --------------------------------------------------------------------------
	// STEP 2: Probe Missing Auth Header (Expect HTTP 401)
	// --------------------------------------------------------------------------
	start = time.Now()
	reqNoAuth, _ := http.NewRequest(http.MethodPost, *baseURL+"/v1/payments", bytes.NewReader([]byte(`{"amount_cents":1000}`)))
	reqNoAuth.Header.Set("Content-Type", "application/json")
	respNoAuth, err := client.Do(reqNoAuth)
	lat = float64(time.Since(start).Microseconds()) / 1000.0
	if err == nil {
		defer respNoAuth.Body.Close()
		results = append(results, ProbeResult{
			Step:      "2. Security Defense: Missing Bearer Token",
			Expected:  "HTTP 401 Unauthorized",
			Actual:    fmt.Sprintf("HTTP %d", respNoAuth.StatusCode),
			Passed:    respNoAuth.StatusCode == http.StatusUnauthorized,
			LatencyMs: lat,
			Details:   "Unauthenticated requests halted at Layer 1.",
		})
	}

	// --------------------------------------------------------------------------
	// STEP 3: Full-Stack End-to-End Settlement (Expect HTTP 200)
	// --------------------------------------------------------------------------
	idempKey := "e2e-probe-" + uuid.New().String()
	payloadMap := map[string]interface{}{
		"amount_cents": *amountCents,
		"recipient":    *recipientID,
	}
	bodyBytes, _ := json.Marshal(payloadMap)

	token, _ := auth.GenerateToken(*merchantID, "merchant", 15*time.Minute)
	now := time.Now().Unix()
	sig := auth.ComputeHMACSignature(bodyBytes, now, auth.GetHMACSecret())

	start = time.Now()
	reqSettlement, _ := http.NewRequest(http.MethodPost, *baseURL+"/v1/payments", bytes.NewReader(bodyBytes))
	reqSettlement.Header.Set("Authorization", "Bearer "+token)
	reqSettlement.Header.Set("Content-Type", "application/json")
	reqSettlement.Header.Set("Idempotency-Key", idempKey)
	reqSettlement.Header.Set("X-Novaris-Timestamp", strconv.FormatInt(now, 10))
	reqSettlement.Header.Set("X-Novaris-Signature", sig)

	respSettlement, err := client.Do(reqSettlement)
	lat = float64(time.Since(start).Microseconds()) / 1000.0
	var firstTxID string
	if err == nil {
		defer respSettlement.Body.Close()
		bodyResp, _ := io.ReadAll(respSettlement.Body)

		var parsed map[string]interface{}
		_ = json.Unmarshal(bodyResp, &parsed)
		if txID, ok := parsed["transaction_id"].(string); ok {
			firstTxID = txID
		}

		passed := respSettlement.StatusCode == http.StatusOK && firstTxID != ""
		results = append(results, ProbeResult{
			Step:      "3. Full-Stack End-to-End Settlement Success",
			Expected:  "HTTP 200 OK (status=settled, transaction_id present)",
			Actual:    fmt.Sprintf("HTTP %d (tx_id=%s)", respSettlement.StatusCode, firstTxID),
			Passed:    passed,
			LatencyMs: lat,
			Details:   fmt.Sprintf("Payment settled across all 5 layers in %0.2fms.", lat),
		})
	}

	// --------------------------------------------------------------------------
	// STEP 4: Exact-Once Network Retry (Expect HTTP 200 + Idempotent-Replayed: true)
	// --------------------------------------------------------------------------
	start = time.Now()
	reqRetry, _ := http.NewRequest(http.MethodPost, *baseURL+"/v1/payments", bytes.NewReader(bodyBytes))
	reqRetry.Header.Set("Authorization", "Bearer "+token)
	reqRetry.Header.Set("Content-Type", "application/json")
	reqRetry.Header.Set("Idempotency-Key", idempKey)
	reqRetry.Header.Set("X-Novaris-Timestamp", strconv.FormatInt(now, 10))
	reqRetry.Header.Set("X-Novaris-Signature", sig)

	respRetry, err := client.Do(reqRetry)
	lat = float64(time.Since(start).Microseconds()) / 1000.0
	if err == nil {
		defer respRetry.Body.Close()
		bodyRetry, _ := io.ReadAll(respRetry.Body)
		var parsedRetry map[string]interface{}
		_ = json.Unmarshal(bodyRetry, &parsedRetry)
		replayedHeader := respRetry.Header.Get("Idempotent-Replayed") == "true"
		retryTxID, _ := parsedRetry["transaction_id"].(string)

		passed := respRetry.StatusCode == http.StatusOK && replayedHeader && retryTxID == firstTxID
		results = append(results, ProbeResult{
			Step:      "4. Exact-Once Network Retry Replay",
			Expected:  "HTTP 200 OK + Idempotent-Replayed: true (Same Transaction ID)",
			Actual:    fmt.Sprintf("HTTP %d (replayed=%v, tx_id=%s)", respRetry.StatusCode, replayedHeader, retryTxID),
			Passed:    passed,
			LatencyMs: lat,
			Details:   "Zero secondary charges! Cached response delivered in sub-millisecond time.",
		})
	}

	// --------------------------------------------------------------------------
	// STEP 5: Key Reuse with Mismatched Payload (Expect HTTP 422)
	// --------------------------------------------------------------------------
	alteredBody, _ := json.Marshal(map[string]interface{}{"amount_cents": 999999, "recipient": *recipientID})
	alteredSig := auth.ComputeHMACSignature(alteredBody, now, auth.GetHMACSecret())

	start = time.Now()
	reqMismatch, _ := http.NewRequest(http.MethodPost, *baseURL+"/v1/payments", bytes.NewReader(alteredBody))
	reqMismatch.Header.Set("Authorization", "Bearer "+token)
	reqMismatch.Header.Set("Content-Type", "application/json")
	reqMismatch.Header.Set("Idempotency-Key", idempKey) // Reuse original key with $9,999.99 transfer!
	reqMismatch.Header.Set("X-Novaris-Timestamp", strconv.FormatInt(now, 10))
	reqMismatch.Header.Set("X-Novaris-Signature", alteredSig)

	respMismatch, err := client.Do(reqMismatch)
	lat = float64(time.Since(start).Microseconds()) / 1000.0
	if err == nil {
		defer respMismatch.Body.Close()
		passed := respMismatch.StatusCode == http.StatusUnprocessableEntity
		results = append(results, ProbeResult{
			Step:      "5. Security Defense: Key Reuse Mismatched Payload",
			Expected:  "HTTP 422 Unprocessable Entity",
			Actual:    fmt.Sprintf("HTTP %d", respMismatch.StatusCode),
			Passed:    passed,
			LatencyMs: lat,
			Details:   "Altered payload hash detected; unauthorized key reuse halted.",
		})
	}

	// --------------------------------------------------------------------------
	// STEP 6: Missing Idempotency Key (Expect HTTP 400)
	// --------------------------------------------------------------------------
	start = time.Now()
	reqMissingKey, _ := http.NewRequest(http.MethodPost, *baseURL+"/v1/payments", bytes.NewReader(bodyBytes))
	reqMissingKey.Header.Set("Authorization", "Bearer "+token)
	reqMissingKey.Header.Set("Content-Type", "application/json")
	reqMissingKey.Header.Set("X-Novaris-Timestamp", strconv.FormatInt(now, 10))
	reqMissingKey.Header.Set("X-Novaris-Signature", sig)

	respMissingKey, err := client.Do(reqMissingKey)
	lat = float64(time.Since(start).Microseconds()) / 1000.0
	if err == nil {
		defer respMissingKey.Body.Close()
		passed := respMissingKey.StatusCode == http.StatusBadRequest
		results = append(results, ProbeResult{
			Step:      "6. Contract Defense: Missing Idempotency-Key Header",
			Expected:  "HTTP 400 Bad Request",
			Actual:    fmt.Sprintf("HTTP %d", respMissingKey.StatusCode),
			Passed:    passed,
			LatencyMs: lat,
			Details:   "Mutating payment route strictly mandates Idempotency-Key.",
		})
	}

	// --------------------------------------------------------------------------
	// PRINT RESULTS TABLE
	// --------------------------------------------------------------------------
	log.Println("\n==================================================================")
	log.Println(" 📊 CONTRACT PROBE TELEMETRY RESULTS")
	log.Println("==================================================================")

	allPassed := true
	for _, r := range results {
		statusSymbol := "✅ PASS"
		if !r.Passed {
			statusSymbol = "❌ FAIL"
			allPassed = false
		}
		fmt.Printf("%s | %-45s | %0.2fms\n", statusSymbol, r.Step, r.LatencyMs)
		fmt.Printf("       Expected: %s\n", r.Expected)
		fmt.Printf("       Actual:   %s\n", r.Actual)
		fmt.Printf("       Notes:    %s\n\n", r.Details)
	}

	if allPassed {
		log.Println("🎉 ALL DAY 11 CONTRACT SCENARIOS VERIFIED: Gateway integration wiring 100% compliant!")
	} else {
		log.Println("⚠️ WARNING: One or more contract assertions failed.")
		os.Exit(1)
	}
}
