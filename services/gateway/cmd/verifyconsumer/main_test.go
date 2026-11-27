// ==============================================================================
// NOVARIS: Distributed Payment Settlement & Adaptive Fraud Intelligence Platform
// MODULE 2: STREAMING CORE (Phase 2, Day 17: verifyconsumer CLI Unit Tests)
// FILE: services/gateway/cmd/verifyconsumer/main_test.go
// ==============================================================================
//
// 📚 COMPUTER SCIENCE & TESTING METHODOLOGY FOR BEGINNERS:
//
// 1. WHAT DOES THIS TEST VERIFY?
//    This unit test validates the data integrity and serialization mechanics of
//    the Day 17 verification diagnostic CLI tool (`verifyconsumer`).
//    Diagnostic tools used in production operations and CI/CD pipelines must
//    themselves be rigorously tested! If a diagnostic tool has a bug in its
//    JSON serialization or status reporting, it could report false positives or
//    break enterprise monitoring dashboards.
// ==============================================================================

package main

import (
	"encoding/json"
	"testing"
	"time"
)

func TestVerifyConsumer_ReportSerialization(t *testing.T) {
	report := Report{
		Timestamp:          time.Now().UTC(),
		TargetBrokers:      []string{"localhost:9092"},
		ClusterStatus:      "REACHABLE",
		TopicVerified:      "novaris.tx.initiated",
		ConsumerGroupID:    "novaris-verifyconsumer-test",
		EventsPublished:    5,
		EventsConsumed:     5,
		InOrderPreserved:   true,
		ZeroLossVerified:   true,
		OffsetsCommitted:   5,
		BatchPublishMs:     15.42,
		ConsumeDurationMs:  45.88,
		Outcome:            "SUCCESS: Day 17 Kafka Consumer Verified (Consumer reads every published event, in order)",
	}

	bytes, err := json.Marshal(report)
	if err != nil {
		t.Fatalf("failed to marshal report: %v", err)
	}

	var parsed Report
	if err := json.Unmarshal(bytes, &parsed); err != nil {
		t.Fatalf("failed to unmarshal report: %v", err)
	}

	if parsed.ClusterStatus != "REACHABLE" {
		t.Errorf("expected ClusterStatus=REACHABLE, got %s", parsed.ClusterStatus)
	}
	if parsed.EventsPublished != 5 || parsed.EventsConsumed != 5 {
		t.Errorf("mismatch in events: pub=%d, con=%d", parsed.EventsPublished, parsed.EventsConsumed)
	}
	if !parsed.InOrderPreserved {
		t.Error("expected InOrderPreserved=true")
	}
	if !parsed.ZeroLossVerified {
		t.Error("expected ZeroLossVerified=true")
	}
	if parsed.OffsetsCommitted != 5 {
		t.Errorf("expected 5 offsets committed, got %d", parsed.OffsetsCommitted)
	}
}
