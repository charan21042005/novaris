// ==============================================================================
// NOVARIS: Distributed Payment Settlement & Adaptive Fraud Intelligence Platform
// MODULE 2: STREAMING CORE (Phase 2, Day 16: verifyproducer CLI Unit Tests)
// FILE: services/gateway/cmd/verifyproducer/main_test.go
// ==============================================================================

package main

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/charan21042005/novaris/services/gateway/internal/streaming"
)

func TestVerifyProducer_ReportSerialization(t *testing.T) {
	txID := uuid.New().String()
	eventID := uuid.New().String()

	evt := &streaming.TransactionInitiatedEvent{
		EventID:       eventID,
		EventType:     streaming.EventTypeTxInitiated,
		Version:       streaming.EventSchemaVersion1,
		TransactionID: txID,
		PayerID:       "acc_merchant_alpha_001",
		PayeeID:       "acc_user_bob_4920",
		AmountCents:   5000,
		Currency:      "USD",
		Timestamp:     time.Now().UTC(),
		Status:        "INITIATED",
	}

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
	if parsed.PublishedEvent.TransactionID != txID {
		t.Errorf("expected TxID=%s, got %s", txID, parsed.PublishedEvent.TransactionID)
	}
	if !parsed.PayloadMatches {
		t.Error("expected PayloadMatches=true")
	}
}
