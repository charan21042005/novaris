// ==============================================================================
// NOVARIS: Distributed Payment Settlement & Adaptive Fraud Intelligence Platform
// OPENAPI CONTRACT VALIDATOR TEST SUITE (Day 12)
// FILE: services/gateway/cmd/verifyopenapi/main_test.go
// ==============================================================================
//
// 📚 COMPUTER SCIENCE & TEST ENGINEERING PRINCIPLES FOR BEGINNERS:
//
// 1. TESTING THE TEST INFRASTRUCTURE ("WHO WATCHES THE WATCHMEN?"):
//    In professional software engineering, test tools and CI/CD validation scripts
//    must themselves be thoroughly tested! If your validator CLI has a bug that
//    always returns "exit 0", broken API specs could slip into production unnoticed.
//
// 2. TWO ESSENTIAL TESTING DIMENSIONS:
//    - The Happy Path (`TestValidateOpenAPISpec_LiveFileSuccess`):
//      Asserts that our real `api/openapi.json` file satisfies all 5 checks.
//    - The Defensive Path (`TestValidateOpenAPISpec_NonExistentFileReturnsError`):
//      Asserts that passing a missing file path fails gracefully with a clear error
//      rather than crashing or returning false positives.
// ==============================================================================

package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestValidateOpenAPISpec_LiveFileSuccess asserts that the live repository OpenAPI JSON
// specification passes all 5 validation checks with 100% compliance.
func TestValidateOpenAPISpec_LiveFileSuccess(t *testing.T) {
	// Attempt multiple relative and absolute path locations to locate api/openapi.json
	// regardless of whether `go test` was invoked from the repo root or inside cmd/verifyopenapi:
	possiblePaths := []string{
		filepath.Join("..", "..", "..", "api", "openapi.json"),
		filepath.Join("..", "..", "internal", "openapi", "openapi.json"),
		filepath.Join("C:", "Users", "Tokachichu SriCharan", "OneDrive", "Documents", "GitHub", "novaris", "api", "openapi.json"),
	}

	var targetPath string
	for _, p := range possiblePaths {
		// Use os.Stat to check if file exists:
		if _, err := os.Stat(p); err == nil {
			targetPath = p
			break
		}
	}

	// Fail immediately if the specification file cannot be found in any expected location:
	if targetPath == "" {
		t.Fatalf("could not locate openapi.json for verification test")
	}

	// Execute the validation engine:
	results, err := ValidateOpenAPISpec(targetPath)
	if err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}

	// Assertion 1: Ensure all 5 defined checks executed:
	if len(results) != 5 {
		t.Errorf("expected 5 validation checks, got %d", len(results))
	}

	// Assertion 2: Verify that every single check reported Passed == true:
	for _, r := range results {
		if !r.Passed {
			t.Errorf("check failed: %s (%s)", r.CheckName, r.Details)
		}
	}
}

// TestValidateOpenAPISpec_NonExistentFileReturnsError verifies defensive error handling
// when the CLI is passed an invalid or missing file path.
func TestValidateOpenAPISpec_NonExistentFileReturnsError(t *testing.T) {
	// Execute validator against a non-existent file name:
	_, err := ValidateOpenAPISpec("non_existent_file_path.json")
	
	// Assert that an error is returned:
	if err == nil {
		t.Errorf("expected error for non-existent file, got nil")
	}
}
