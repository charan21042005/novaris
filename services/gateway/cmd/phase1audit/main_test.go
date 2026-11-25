// ==============================================================================
// NOVARIS: Distributed Payment Settlement & Adaptive Fraud Intelligence Platform
// PHASE 1 CAPSTONE AUDIT RUNNER TEST HARNESS (Day 14)
// FILE: services/gateway/cmd/phase1audit/main_test.go
// ==============================================================================
//
// 📚 COMPUTER SCIENCE & TEST DESIGN FOR AUDIT TOOLS:
//
// 1. TESTING THE AUDIT RUNNER:
//    An audit tool must be verified to ensure:
//      - It discovers all candidate paths across multiple repository depths.
//      - It accurately parses SQL keywords and invariants without false negatives.
//      - It aggregates totals and sets `Certified: true` when all 5 criteria are met.
//      - The JSON serialization format is schema-compliant for CI/CD ingestion.
// ==============================================================================

package main

import (
	"encoding/json"
	"testing"
)

// TestRunPhase1Audit_AllCriteriaPass asserts that RunPhase1Audit evaluates all 5 criteria
// and produces a certified audit report.
func TestRunPhase1Audit_AllCriteriaPass(t *testing.T) {
	report := RunPhase1Audit()

	if report.TotalChecks != 5 {
		t.Fatalf("Expected 5 total criteria checks, got %d", report.TotalChecks)
	}

	for _, check := range report.DetailedChecks {
		if !check.Passed {
			t.Errorf("Criterion %s (%s) failed! Details: %s", check.CriterionID, check.Title, check.Details)
		}
	}

	if !report.Certified {
		t.Fatalf("Expected Phase 1 to be fully Certified, but Certified=false (Passed=%d, Failed=%d)",
			report.PassedChecks, report.FailedChecks)
	}

	// Verify JSON marshal round-trip
	data, err := json.Marshal(report)
	if err != nil {
		t.Fatalf("Failed to marshal audit report to JSON: %v", err)
	}

	var decoded Phase1AuditReport
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal audit report from JSON: %v", err)
	}

	if decoded.PassedChecks != 5 || !decoded.Certified {
		t.Fatalf("Decoded JSON report does not match original report")
	}

	t.Logf("✅ AUDIT RUNNER UNIT TEST PASS: All %d checks passed and certified!", report.PassedChecks)
}
