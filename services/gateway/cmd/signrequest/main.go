// ==============================================================================
// NOVARIS: Distributed Payment Settlement & Adaptive Fraud Intelligence Platform
// CLI Utility: Developer HMAC Request Signer (cmd/signrequest/main.go)
// ==============================================================================
//
// Purpose:
// Simulates a merchant client library or automated API caller by:
// 1. Building the canonical payload ("<timestamp>.<raw_body>").
// 2. Computing the HMAC-SHA256 signature using the shared secret.
// 3. Emitting formatted headers and ready-to-execute curl commands.
//
// Usage:
//   go run cmd/signrequest/main.go
//   go run cmd/signrequest/main.go -body '{"amount_cents":10000,"recipient":"merchant_xyz"}'
//   go run cmd/signrequest/main.go -offset -600 (generates a stale timestamp 10m ago for 401 testing)
// ==============================================================================

package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/charan21042005/novaris/services/gateway/internal/auth"
)

func main() {
	bodyFlag := flag.String("body", `{"amount_cents":5000,"recipient":"merchant_456"}`, "Raw JSON request body to sign")
	offsetFlag := flag.Int64("offset", 0, "Timestamp offset in seconds (e.g. -600 for a 10m stale request)")
	secretFlag := flag.String("secret", "", "Custom HMAC secret (defaults to HMAC_SECRET env or dev fallback)")
	tamperFlag := flag.Bool("tamper", false, "Simulate tampering by altering the body after computing signature")
	flag.Parse()

	// Determine secret key
	secret := auth.GetHMACSecret()
	if *secretFlag != "" {
		secret = []byte(*secretFlag)
	}

	// Calculate timestamp with requested offset
	targetTime := time.Now().Unix() + *offsetFlag
	timestampStr := strconv.FormatInt(targetTime, 10)

	// Compute HMAC signature on original body
	originalBodyBytes := []byte(*bodyFlag)
	signature := auth.ComputeHMACSignature(originalBodyBytes, targetTime, secret)

	outputBody := *bodyFlag
	if *tamperFlag {
		outputBody = `{"amount_cents":5000000,"recipient":"attacker_tampered_payload"}`
	}

	fmt.Println("==============================================================================")
	fmt.Println("NOVARIS Developer HMAC-SHA256 Request Signing Utility")
	fmt.Println("==============================================================================")
	fmt.Printf("Raw Body:    %s\n", outputBody)
	fmt.Printf("Timestamp:   %s\n", timestampStr)
	fmt.Printf("Signature:   %s\n", signature)
	if *tamperFlag {
		fmt.Println("[WARNING] -tamper active: signature was generated for original body, not tampered body!")
	}
	if *offsetFlag != 0 {
		fmt.Printf("[INFO] -offset active: timestamp is %d seconds offset from current server time.\n", *offsetFlag)
	}
	fmt.Println("------------------------------------------------------------------------------")
	fmt.Println("Example Headers:")
	fmt.Printf("  X-Novaris-Timestamp: %s\n", timestampStr)
	fmt.Printf("  X-Novaris-Signature: %s\n", signature)
	fmt.Println("------------------------------------------------------------------------------")
	fmt.Println("Ready-to-Run curl command (assuming $TOKEN contains valid JWT):")
	fmt.Printf("curl -i -X POST http://localhost:8080/v1/payments \\\n"+
		"  -H \"Authorization: Bearer $TOKEN\" \\\n"+
		"  -H \"X-Novaris-Timestamp: %s\" \\\n"+
		"  -H \"X-Novaris-Signature: %s\" \\\n"+
		"  -H \"Content-Type: application/json\" \\\n"+
		"  -d '%s'\n", timestampStr, signature, outputBody)
	fmt.Println("==============================================================================")

	// If invoked with a pipe or specific automated script, write summary to stdout
	if os.Getenv("NOVARIS_CLI_RAW") == "1" {
		fmt.Printf("%s|%s|%s\n", timestampStr, signature, outputBody)
	}
}
