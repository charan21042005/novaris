// ==============================================================================
// NOVARIS: Distributed Payment Settlement & Adaptive Fraud Intelligence Platform
// MODULE 2: STREAMING CORE (Phase 2, Day 17: Settlement Consumer Unit Tests)
// FILE: services/settlement/consumer_test.go
// ==============================================================================
//
// 📚 COMPUTER SCIENCE & TESTING METHODOLOGY FOR BEGINNERS:
//
// 1. WHAT DOES THIS TEST SUITE VERIFY?
//    This unit test validates the core business processing logic of the Settlement
//    Consumer Service:
//      - Scenario 1 (Valid Settlement): An authentic `tx.initiated` message is decoded,
//        debit and credit volumes balance out ($ \sum \text{Debits} = \sum \text{Credits} $),
//        and an audit record is stored.
//      - Scenario 2 (Idempotent Deduplication): If network retries deliver the SAME
//        transaction twice, the processor detects the duplicate, logs a warning, skips
//        double-debiting the account, and returns nil so the offset advances!
//      - Scenario 3 (Poison-Pill Defense): Malformed JSON bytes are rejected with a clear
//        error, preventing corrupted data from entering the financial ledger.
//      - Scenario 4 (Invariant Enforcement): Zero-cent transfers are rejected.
// ==============================================================================

package main

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

// TestSettlementProcessor_ValidEventProcessing verifies successful event decoding and accounting.
func TestSettlementProcessor_ValidEventProcessing(t *testing.T) {
	ctx := context.Background()
	proc := NewSettlementProcessor()

	txID := uuid.New().String()
	payerID := "c0000000-0000-0000-0000-000000000001"
	payeeID := "c0000000-0000-0000-0000-000000000002"

	event := &TransactionInitiatedEvent{
		EventID:        uuid.New().String(),
		EventType:      EventTypeTxInitiated,
		Version:        EventSchemaVersion1,
		TransactionID:  txID,
		PayerID:        payerID,
		PayeeID:        payeeID,
		AmountCents:    15000, // $150.00 USD
		Currency:       "USD",
		MerchantID:     "merchant_unit_test",
		IdempotencyKey: uuid.New().String(),
		Timestamp:      time.Now().UTC(),
		Status:         "INITIATED",
	}

	payload, err := event.ToJSON()
	if err != nil {
		t.Fatalf("failed to serialize event: %v", err)
	}

	msg := kafka.Message{
		Topic:     TopicTxInitiated,
		Partition: 1,
		Offset:    42,
		Key:       []byte(payerID),
		Value:     payload,
		Time:      time.Now().UTC(),
	}

	// Execute handler:
	if err := proc.HandleMessage(ctx, msg); err != nil {
		t.Fatalf("unexpected HandleMessage error: %v", err)
	}

	// Assertions:
	if proc.GetProcessedCount() != 1 {
		t.Errorf("expected 1 processed record, got %d", proc.GetProcessedCount())
	}

	debits, credits := proc.TotalVolumes()
	if debits != 15000 || credits != 15000 {
		t.Errorf("ledger balance mismatch: expected debits=credits=15000, got debits=%d, credits=%d",
			debits, credits)
	}

	logEntries := proc.GetAuditLog()
	if len(logEntries) != 1 {
		t.Fatalf("expected 1 audit log entry, got %d", len(logEntries))
	}
	if logEntries[0].TransactionID != txID {
		t.Errorf("audit log TxID mismatch: expected %s, got %s", txID, logEntries[0].TransactionID)
	}
	if logEntries[0].Partition != 1 || logEntries[0].Offset != 42 {
		t.Errorf("audit log partition/offset mismatch: got part=%d, off=%d", logEntries[0].Partition, logEntries[0].Offset)
	}
}

// TestSettlementProcessor_IdempotentDeduplication proves duplicate protection.
func TestSettlementProcessor_IdempotentDeduplication(t *testing.T) {
	ctx := context.Background()
	proc := NewSettlementProcessor()

	txID := uuid.New().String()
	event := &TransactionInitiatedEvent{
		EventID:       uuid.New().String(),
		EventType:     EventTypeTxInitiated,
		TransactionID: txID,
		PayerID:       "acc_001",
		PayeeID:       "acc_002",
		AmountCents:   5000,
		Currency:      "USD",
		Timestamp:     time.Now().UTC(),
	}

	payload, _ := event.ToJSON()
	msg := kafka.Message{Topic: "novaris.tx.initiated", Partition: 0, Offset: 10, Value: payload}

	// First delivery: should succeed and record $50:
	if err := proc.HandleMessage(ctx, msg); err != nil {
		t.Fatalf("first delivery failed: %v", err)
	}

	// Second delivery (e.g. network retry after crash): should return nil (commit) but skip double debit:
	if err := proc.HandleMessage(ctx, msg); err != nil {
		t.Fatalf("duplicate delivery should succeed idempotently, got error: %v", err)
	}

	// Cumulative ledger volume should still be exactly 5000 cents ($50.00), NOT 10000 cents:
	debits, credits := proc.TotalVolumes()
	if debits != 5000 || credits != 5000 {
		t.Errorf("idempotency violation: double debit occurred! Expected 5000, got debits=%d, credits=%d",
			debits, credits)
	}
	if proc.GetProcessedCount() != 1 {
		t.Errorf("expected processed count=1, got %d", proc.GetProcessedCount())
	}
}

// TestSettlementProcessor_CorruptedMessageReturnsError tests defensive handling of bad JSON.
func TestSettlementProcessor_CorruptedMessageReturnsError(t *testing.T) {
	ctx := context.Background()
	proc := NewSettlementProcessor()

	badMsg := kafka.Message{
		Topic:     "novaris.tx.initiated",
		Partition: 0,
		Offset:    99,
		Value:     []byte("this is not valid json!"),
	}

	err := proc.HandleMessage(ctx, badMsg)
	if err == nil {
		t.Fatal("expected error for corrupted JSON payload, got nil")
	}
}

// TestSettlementProcessor_InvalidInvariantsReturnsError tests domain rules.
func TestSettlementProcessor_InvalidInvariantsReturnsError(t *testing.T) {
	ctx := context.Background()
	proc := NewSettlementProcessor()

	// Zero amount cents:
	invalidEvent := &TransactionInitiatedEvent{
		TransactionID: uuid.New().String(),
		PayerID:       "acc_001",
		PayeeID:       "acc_002",
		AmountCents:   0, // Invalid!
	}
	rawJSON := []byte(`{"transaction_id":"` + invalidEvent.TransactionID + `","payer_id":"acc_1","payee_id":"acc_2","amount_cents":0}`)

	msg := kafka.Message{Topic: "novaris.tx.initiated", Partition: 0, Offset: 100, Value: rawJSON}
	err := proc.HandleMessage(ctx, msg)
	if err == nil {
		t.Fatal("expected error for zero amount cents, got nil")
	}
}
