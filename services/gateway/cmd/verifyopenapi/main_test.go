// ==============================================================================
// NOVARIS: Distributed Payment Settlement & Adaptive Fraud Intelligence Platform
// OPENAPI CONTRACT VALIDATOR TEST SUITE (Day 12)
// FILE: services/gateway/cmd/verifyopenapi/main_test.go
// ==============================================================================

package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateOpenAPISpec_LiveFileSuccess(t *testing.T) {
	// Locate api/openapi.json relative to cmd/verifyopenapi
	possiblePaths := []string{
		filepath.Join("..", "..", "..", "api", "openapi.json"),
		filepath.Join("..", "..", "internal", "openapi", "openapi.json"),
		filepath.Join("C:", "Users", "Tokachichu SriCharan", "OneDrive", "Documents", "GitHub", "novaris", "api", "openapi.json"),
	}

	var targetPath string
	for _, p := range possiblePaths {
		if _, err := os.Stat(p); err == nil {
			targetPath = p
			break
		}
	}

	if targetPath == "" {
		t.Fatalf("could not locate openapi.json for verification test")
	}

	results, err := ValidateOpenAPISpec(targetPath)
	if err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}

	if len(results) != 5 {
		t.Errorf("expected 5 validation checks, got %d", len(results))
	}

	for _, r := range results {
		if !r.Passed {
			t.Errorf("check failed: %s (%s)", r.CheckName, r.Details)
		}
	}
}

func TestValidateOpenAPISpec_NonExistentFileReturnsError(t *testing.T) {
	_, err := ValidateOpenAPISpec("non_existent_file_path.json")
	if err == nil {
		t.Errorf("expected error for non-existent file, got nil")
	}
}
