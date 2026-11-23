// ==============================================================================
// NOVARIS: Distributed Payment Settlement & Adaptive Fraud Intelligence Platform
// CLI Utility: Test JWT Token Generator (cmd/gentoken/main.go)
// ==============================================================================
//
// Usage:
//   go run cmd/gentoken/main.go
//   go run cmd/gentoken/main.go -merchant merchant_acme_corp -role admin
//   go run cmd/gentoken/main.go -expired (generates an already-expired token for testing 401s)
// ==============================================================================

package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/charan21042005/novaris/services/gateway/internal/auth"
)

func main() {
	// Parse CLI flags
	merchantID := flag.String("merchant", "merchant_stripe_test_101", "Merchant ID to embed in token claims")
	role := flag.String("role", "merchant_admin", "RBAC role to embed in token claims")
	minutes := flag.Int("mins", 60, "Duration in minutes before token expires")
	expired := flag.Bool("expired", false, "Generate an intentionally expired token (for testing 401 rejections)")
	flag.Parse()

	duration := time.Duration(*minutes) * time.Minute
	if *expired {
		// Set negative duration so the token is already expired upon generation
		duration = -10 * time.Minute
	}

	token, err := auth.GenerateToken(*merchantID, *role, duration)
	if err != nil {
		log.Fatalf("FATAL: Failed to generate token: %v", err)
	}

	// Print bare token for easy piping (e.g. TOKEN=$(go run cmd/gentoken/main.go))
	fmt.Println(token)
}
