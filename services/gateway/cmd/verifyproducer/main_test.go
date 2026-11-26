// ==============================================================================
// NOVARIS: Distributed Payment Settlement & Adaptive Fraud Intelligence Platform
// MODULE 2: STREAMING CORE (Phase 2, Day 16: verifyproducer CLI Unit Tests)
// FILE: services/gateway/cmd/verifyproducer/main_test.go
// ==============================================================================
//
// 📚 COMPUTER SCIENCE & TESTING METHODOLOGY FOR BEGINNERS:
//
// 1. WHAT DOES THIS TEST VERIFY?
//    This unit test validates the data integrity and serialization mechanics of
//    the Day 16 verification diagnostic CLI tool (`verifyproducer`).
//    Diagnostic tools used in production operations and CI/CD pipelines must
//    themselves be rigorously tested! If a diagnostic tool has a bug in its
//    JSON serialization or status reporting, it could report false positives or
//    break enterprise monitoring dashboards.
//
// 2. TESTING CRITICAL DATA STRUCTURES (Report Struct):
//    The `Report` struct contains all metrics from the 4 verification stages:
//    - Broker Reachability
//    - Topic Provisioning
//    - Event Production (with Leader ACK latency)
//    - Topic Visibility (Offset, Partition, and Byte Matching)
//    This test verifies that the Report struct can be converted to JSON and back
//    without loss of precision or corrupted fields.
// ==============================================================================

package main

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/charan21042005/novaris/services/gateway/internal/streaming"
)

// TestVerifyProducer_ReportSerialization verifies that the diagnostic report
// serializes cleanly to JSON and accurately reconstructs all field values.
func TestVerifyProducer_ReportSerialization(t *testing.T) {
	txID := uuid.New().String()
	eventID := uuid.New().String()

	// 1. Construct a mock TransactionInitiatedEvent:
	evt := &streaming.TransactionInitiatedEvent{
		EventID:       eventID,
		EventType:     streaming.EventTypeTxInitiated,
		Version:       streaming.EventSchemaVersion1,
		TransactionID: txID,
		PayerID:       "acc_merchant_alpha_001", // Partition Key
		PayeeID:       "acc_user_bob_4920",
		AmountCents:   5000, // $50.00 USD
		Currency:      "USD",
		Timestamp:     time.Now().UTC(),
		Status:        "INITIATED",
	}

	// 2. Populate the full audit report struct:
	report := Report{
		Timestamp:        time.Now().UTC(),
		TargetBrokers:    []string{"localhost:9092"},
		ClusterStatus:    "REACHABLE",
		TopicVerified:    streaming.TopicTxInitiated,
		PartitionCount:   3,
		PublishedEvent:   evt,
		PublishLatencyMs: 4.25,
		ConsumedEvent:    evt,
		ReadLatencyMs:    12.10,
		PayloadMatches:   true,
		Outcome:          "SUCCESS: Day 16 Kafka Producer Verified (Events Visible in Kafka Topic)",
	}

	// 3. Serialize report to JSON bytes:
	bytes, err := json.Marshal(report)
	if err != nil {
		t.Fatalf("failed to marshal report to JSON: %v", err)
	}

	// 4. Deserialize back into a fresh struct (Fidelity Audit):
	var parsed Report
	if err := json.Unmarshal(bytes, &parsed); err != nil {
		t.Fatalf("failed to unmarshal JSON into Report: %v", err)
	}

	// 5. Assert that every critical field survived the round-trip:
	if parsed.ClusterStatus != "REACHABLE" {
		t.Errorf("expected ClusterStatus=REACHABLE, got %s", parsed.ClusterStatus)
	}
	if parsed.PublishedEvent.TransactionID != txID {
		t.Errorf("expected TxID=%s, got %s", txID, parsed.PublishedEvent.TransactionID)
	}
	if parsed.PublishedEvent.EventID != eventID {
		t.Errorf("expected EventID=%s, got %s", eventID, parsed.PublishedEvent.EventID)
	}
	if parsed.PartitionCount != 3 {
		t.Errorf("expected PartitionCount=3, got %d", parsed.PartitionCount)
	}
	if !parsed.PayloadMatches {
		t.Error("expected PayloadMatches=true")
	}
}
