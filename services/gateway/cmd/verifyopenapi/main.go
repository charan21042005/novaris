// ==============================================================================
// NOVARIS: Distributed Payment Settlement & Adaptive Fraud Intelligence Platform
// OPENAPI CONTRACT & SCHEMA VALIDATOR CLI (Day 12)
// FILE: services/gateway/cmd/verifyopenapi/main.go
// ==============================================================================
//
// 📚 COMPUTER SCIENCE & CONTRACT TESTING FOR BEGINNERS:
//
// 1. WHAT IS CONTRACT TESTING?
//    In distributed microservices, APIs act as a legally binding "contract" between
//    producers (backend services) and consumers (frontend web apps, mobile apps,
//    merchant integration code, and third-party payment orchestrators).
//    If a backend engineer changes a field name (e.g. from `amount_cents` to `amount`),
//    consumers immediately break in production!
//
// 2. THE PURPOSE OF THIS CLI TOOL:
//    This automated validator executes in CI/CD build pipelines (GitHub Actions)
//    to statically inspect `api/openapi.json` before any PR can be merged:
//      - Asserts OpenAPI 3.0.x compliance
//      - Asserts service title and descriptive metadata
//      - Asserts mandatory paths (/healthz, /, /openapi.json, /docs, /v1/payments)
//      - Asserts all 7 HTTP status codes (200, 400, 401, 409, 422, 429, 500)
//      - Asserts data model schemas (PaymentRequest, PaymentReceipt, etc.)
// ==============================================================================

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ValidationResult captures the pass/fail outcome and diagnostic details of an assertion.
type ValidationResult struct {
	CheckName string // Human-readable assertion name (e.g., "Mandatory Ingress Routes")
	Passed    bool   // True if assertion succeeded, false if violation detected
	Details   string // Diagnostic explanation of the check outcome
}

// ValidateOpenAPISpec parses the target OpenAPI JSON file and verifies strict
// compliance with NOVARIS Day 12 API Contract Standards.
//
// Returns a slice of individual ValidationResult structs and an error if the file
// cannot be read or unmarshaled as valid JSON.
func ValidateOpenAPISpec(specPath string) ([]ValidationResult, error) {
	// Read the raw file bytes from the specified filesystem path:
	bytes, err := os.ReadFile(specPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read OpenAPI spec at %s: %w", specPath, err)
	}

	// Unmarshal JSON into a generic map[string]interface{}:
	// In Go, dynamic JSON structures with varying fields are parsed into map[string]interface{}.
	var spec map[string]interface{}
	if err := json.Unmarshal(bytes, &spec); err != nil {
		return nil, fmt.Errorf("failed to parse JSON schema: %w", err)
	}

	var results []ValidationResult

	// --------------------------------------------------------------------------
	// Assertion 1: Check OpenAPI Specification Version (OAS 3.0.x)
	// --------------------------------------------------------------------------
	// Use Go type assertion .(string) to safely extract the string value.
	version, _ := spec["openapi"].(string)
	if strings.HasPrefix(version, "3.0") {
		results = append(results, ValidationResult{
			CheckName: "OpenAPI Specification Version",
			Passed:    true,
			Details:   fmt.Sprintf("Compliant OAS version: %s", version),
		})
	} else {
		results = append(results, ValidationResult{
			CheckName: "OpenAPI Specification Version",
			Passed:    false,
			Details:   fmt.Sprintf("Non-compliant version: %s (expected 3.0.x)", version),
		})
	}

	// --------------------------------------------------------------------------
	// Assertion 2: Check API Metadata & Title
	// --------------------------------------------------------------------------
	info, _ := spec["info"].(map[string]interface{})
	title, _ := info["title"].(string)
	if strings.Contains(title, "NOVARIS") {
		results = append(results, ValidationResult{
			CheckName: "API Metadata & Title",
			Passed:    true,
			Details:   fmt.Sprintf("Found valid title: %s", title),
		})
	} else {
		results = append(results, ValidationResult{
			CheckName: "API Metadata & Title",
			Passed:    false,
			Details:   "Missing or invalid title in info block",
		})
	}

	// --------------------------------------------------------------------------
	// Assertion 3: Check Mandatory Ingress Routes
	// --------------------------------------------------------------------------
	paths, _ := spec["paths"].(map[string]interface{})
	requiredRoutes := []string{"/healthz", "/", "/openapi.json", "/docs", "/v1/payments"}
	missingRoutes := []string{}
	for _, r := range requiredRoutes {
		if _, ok := paths[r]; !ok {
			missingRoutes = append(missingRoutes, r)
		}
	}

	if len(missingRoutes) == 0 {
		results = append(results, ValidationResult{
			CheckName: "Mandatory Ingress Routes",
			Passed:    true,
			Details:   fmt.Sprintf("All %d required endpoints present in schema", len(requiredRoutes)),
		})
	} else {
		results = append(results, ValidationResult{
			CheckName: "Mandatory Ingress Routes",
			Passed:    false,
			Details:   fmt.Sprintf("Missing routes: %s", strings.Join(missingRoutes, ", ")),
		})
	}

	// --------------------------------------------------------------------------
	// Assertion 4: Check /v1/payments Method & Status Codes (All 7 Codes)
	// --------------------------------------------------------------------------
	paymentsPath, _ := paths["/v1/payments"].(map[string]interface{})
	postOp, hasPost := paymentsPath["post"].(map[string]interface{})
	if hasPost {
		responses, _ := postOp["responses"].(map[string]interface{})
		requiredCodes := []string{"200", "400", "401", "409", "422", "429", "500"}
		missingCodes := []string{}
		for _, c := range requiredCodes {
			if _, ok := responses[c]; !ok {
				missingCodes = append(missingCodes, c)
			}
		}

		if len(missingCodes) == 0 {
			results = append(results, ValidationResult{
				CheckName: "Payment Contract Status Codes",
				Passed:    true,
				Details:   fmt.Sprintf("All %d status codes documented (200, 400, 401, 409, 422, 429, 500)", len(requiredCodes)),
			})
		} else {
			results = append(results, ValidationResult{
				CheckName: "Payment Contract Status Codes",
				Passed:    false,
				Details:   fmt.Sprintf("Missing status codes: %s", strings.Join(missingCodes, ", ")),
			})
		}
	} else {
		results = append(results, ValidationResult{
			CheckName: "Payment Contract Status Codes",
			Passed:    false,
			Details:   "POST method missing on /v1/payments",
		})
	}

	// --------------------------------------------------------------------------
	// Assertion 5: Check Component Model Schemas Completeness
	// --------------------------------------------------------------------------
	components, _ := spec["components"].(map[string]interface{})
	schemas, _ := components["schemas"].(map[string]interface{})
	requiredSchemas := []string{"PaymentRequest", "PaymentReceipt", "HealthResponse", "RootDiscoveryResponse", "ErrorResponse"}
	missingSchemas := []string{}
	for _, s := range requiredSchemas {
		if _, ok := schemas[s]; !ok {
			missingSchemas = append(missingSchemas, s)
		}
	}

	if len(missingSchemas) == 0 {
		results = append(results, ValidationResult{
			CheckName: "Component Model Schemas",
			Passed:    true,
			Details:   fmt.Sprintf("All %d core data models defined (%s)", len(requiredSchemas), strings.Join(requiredSchemas, ", ")),
		})
	} else {
		results = append(results, ValidationResult{
			CheckName: "Component Model Schemas",
			Passed:    false,
			Details:   fmt.Sprintf("Missing schemas: %s", strings.Join(missingSchemas, ", ")),
		})
	}

	return results, nil
}

// findDefaultOpenAPISpec dynamically checks parent directories to locate api/openapi.json
// regardless of whether the command is run from repo root, services/gateway, or cmd/verifyopenapi.
func findDefaultOpenAPISpec() string {
	candidates := []string{
		filepath.Join("api", "openapi.json"),
		filepath.Join("..", "api", "openapi.json"),
		filepath.Join("..", "..", "api", "openapi.json"),
		filepath.Join("..", "..", "..", "api", "openapi.json"),
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	return filepath.Join("..", "..", "..", "api", "openapi.json")
}

// main executes the CLI command parsing flags, running assertions, and formatting output.
func main() {
	defaultPath := findDefaultOpenAPISpec()
	specPath := flag.String("spec", defaultPath, "Path to openapi.json file to validate")
	flag.Parse()

	fmt.Println("==================================================================")
	fmt.Println(" 🚀 NOVARIS Day 12: OpenAPI 3.0 Contract Verification Validator")
	fmt.Printf(" Target Spec: %s\n", *specPath)
	fmt.Println("==================================================================")

	// Execute all assertions against the target file:
	results, err := ValidateOpenAPISpec(*specPath)
	if err != nil {
		fmt.Printf("❌ CRITICAL ERROR: %v\n", err)
		os.Exit(1)
	}

	allPassed := true
	for i, r := range results {
		icon := "✅ PASS"
		if !r.Passed {
			icon = "❌ FAIL"
			allPassed = false
		}
		fmt.Printf("%s | %d. %-32s | %s\n", icon, i+1, r.CheckName, r.Details)
	}

	fmt.Println("==================================================================")
	if allPassed {
		fmt.Println("🎉 ALL OPENAPI 3.0 CONTRACT CHECKS VERIFIED: 100% Schema Conformance!")
	} else {
		fmt.Println("❌ FAILED: OpenAPI contract contains schema violations.")
		os.Exit(1)
	}
}
