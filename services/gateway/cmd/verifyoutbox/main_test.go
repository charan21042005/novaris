// ==============================================================================
// NOVARIS: Distributed Payment Settlement & Adaptive Fraud Intelligence Platform
// MODULE 1: API Gateway Service - verifyoutbox CLI Unit Test Suite (Day 18)
// FILE: services/gateway/cmd/verifyoutbox/main_test.go
// ==============================================================================

package main

import (
	"encoding/json"
	"testing"
	"time"
)

// TestReport_Serialization verifies Report JSON marshaling and data fidelity.
func TestReport_Serialization(t *testing.T) {
	rep := Report{
		Timestamp:            time.Now().UTC(),
		Mode:                 "HERMETIC_SIMULATION",
		DatabaseReachable:    true,
		KafkaReachable:       true,
		DualWriteAtomic:      true,
		EventsEnlisted:       10,
		EventsRelayed:        10,
		CrashSimulated:       true,
		DroppedEventsOnCrash: 0,
		DuplicateDebits:      0,
		OutcomeCertified:     true,
	}

	data, err := json.Marshal(rep)
	if err != nil {
		t.Fatalf("failed marshaling report to JSON: %v", err)
	}

	var decoded Report
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed unmarshaling report JSON: %v", err)
	}

	if decoded.Mode != "HERMETIC_SIMULATION" {
		t.Errorf("expected mode HERMETIC_SIMULATION, got %s", decoded.Mode)
	}
	if decoded.DroppedEventsOnCrash != 0 {
		t.Errorf("expected 0 dropped events, got %d", decoded.DroppedEventsOnCrash)
	}
	if !decoded.OutcomeCertified {
		t.Errorf("expected outcome certified true")
	}
}
