// ==============================================================================
// NOVARIS: Distributed Payment Settlement & Adaptive Fraud Intelligence Platform
// MODULE 2: STREAMING CORE (Phase 2, Day 17: Kafka Consumer Unit Tests)
// FILE: services/gateway/internal/streaming/consumer_test.go
// ==============================================================================
//
// 📚 COMPUTER SCIENCE & TESTING METHODOLOGY FOR BEGINNERS:
//
// 1. WHAT DO THESE TESTS PROVE?
//    In event-driven architectures, testing consumers is critical to preventing
//    financial disaster. If a consumer has an off-by-one error in its offset commit,
//    or if it fails to commit after processing a payment, it might re-process the
//    payment upon restart and charge the customer twice!
//    These unit tests verify:
//      - Config Validation: Guard against missing group IDs, topics, or broker lists.
//      - Processing & Commit Fidelity: Assert that the message handler is invoked
//        and that the offset is committed ONLY after handler success.
//      - Defensive Error Handling: Assert that if the handler returns an error,
//        the offset is NEVER committed (enforcing At-Least-Once delivery).
//      - Lifecycle Cleanliness: Assert that Close() flushes resources and is idempotent.
// ==============================================================================

package streaming

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

// TestConsumerConfig_Validation tests configuration defaults and input validation.
func TestConsumerConfig_Validation(t *testing.T) {
	// Sub-test 1: Default config is populated cleanly:
	t.Run("default config has expected defaults", func(t *testing.T) {
		cfg := DefaultConsumerConfig("test-group", "novaris.tx.initiated")
		if err := cfg.Validate(); err != nil {
			t.Fatalf("expected valid default config, got %v", err)
		}
		if cfg.MinBytes != 1 {
			t.Errorf("expected MinBytes=1, got %d", cfg.MinBytes)
		}
		if cfg.MaxBytes != 10e6 {
			t.Errorf("expected MaxBytes=10MB, got %d", cfg.MaxBytes)
		}
		if cfg.CommitInterval != 0 {
			t.Errorf("expected CommitInterval=0 (manual), got %v", cfg.CommitInterval)
		}
	})

	// Sub-test 2: Missing GroupID fails validation:
	t.Run("missing group_id returns ErrEmptyGroupID", func(t *testing.T) {
		cfg := ConsumerConfig{
			Brokers: []string{"localhost:9092"},
			Topic:   "novaris.tx.initiated",
		}
		if err := cfg.Validate(); err != ErrEmptyGroupID {
			t.Fatalf("expected ErrEmptyGroupID, got %v", err)
		}
	})

	// Sub-test 3: Missing Topic returns ErrEmptyTopic:
	t.Run("missing topic returns ErrEmptyTopic", func(t *testing.T) {
		cfg := ConsumerConfig{
			Brokers: []string{"localhost:9092"},
			GroupID: "test-group",
		}
		if err := cfg.Validate(); err != ErrEmptyTopic {
			t.Fatalf("expected ErrEmptyTopic, got %v", err)
		}
	})

	// Sub-test 4: Missing Brokers returns error:
	t.Run("missing brokers returns error", func(t *testing.T) {
		cfg := ConsumerConfig{
			GroupID: "test-group",
			Topic:   "test-topic",
		}
		if err := cfg.Validate(); err == nil {
			t.Fatal("expected error for empty brokers list, got nil")
		}
	})
}

// TestMockConsumer_ProcessAndCommit verifies that MockConsumer drains queued events,
// calls the handler, and commits offsets in exact chronological order.
func TestMockConsumer_ProcessAndCommit(t *testing.T) {
	ctx := context.Background()

	var processedCount int
	var receivedTxIDs []string

	handler := func(ctx context.Context, msg kafka.Message) error {
		processedCount++
		var evt TransactionInitiatedEvent
		if err := json.Unmarshal(msg.Value, &evt); err != nil {
			return err
		}
		receivedTxIDs = append(receivedTxIDs, evt.TransactionID)
		return nil
	}

	mock := NewMockConsumer(handler)
	defer mock.Close()

	// 1. Enqueue 3 sequential mock events:
	tx1 := uuid.New().String()
	tx2 := uuid.New().String()
	tx3 := uuid.New().String()

	evt1 := &TransactionInitiatedEvent{
		EventID:       uuid.New().String(),
		EventType:     EventTypeTxInitiated,
		TransactionID: tx1,
		PayerID:       "acc_payer_001",
		PayeeID:       "acc_payee_002",
		AmountCents:   1000,
		Currency:      "USD",
		Timestamp:     time.Now().UTC(),
	}
	evt2 := &TransactionInitiatedEvent{
		EventID:       uuid.New().String(),
		EventType:     EventTypeTxInitiated,
		TransactionID: tx2,
		PayerID:       "acc_payer_001",
		PayeeID:       "acc_payee_002",
		AmountCents:   2000,
		Currency:      "USD",
		Timestamp:     time.Now().UTC(),
	}
	evt3 := &TransactionInitiatedEvent{
		EventID:       uuid.New().String(),
		EventType:     EventTypeTxInitiated,
		TransactionID: tx3,
		PayerID:       "acc_payer_001",
		PayeeID:       "acc_payee_002",
		AmountCents:   3000,
		Currency:      "USD",
		Timestamp:     time.Now().UTC(),
	}

	_ = mock.EnqueueEvent(evt1, 0, 100)
	_ = mock.EnqueueEvent(evt2, 0, 101)
	_ = mock.EnqueueEvent(evt3, 0, 102)

	// 2. Execute consumer loop:
	if err := mock.Start(ctx); err != nil {
		t.Fatalf("unexpected mock.Start error: %v", err)
	}

	// 3. Assert all 3 events were processed in order:
	if processedCount != 3 {
		t.Fatalf("expected 3 processed events, got %d", processedCount)
	}
	if len(receivedTxIDs) != 3 {
		t.Fatalf("expected 3 received TxIDs, got %d", len(receivedTxIDs))
	}
	if receivedTxIDs[0] != tx1 || receivedTxIDs[1] != tx2 || receivedTxIDs[2] != tx3 {
		t.Errorf("FIFO order mismatch: got %v", receivedTxIDs)
	}

	// 4. Assert offsets were committed:
	committed := mock.GetCommittedOffsets()
	if len(committed) != 3 {
		t.Fatalf("expected 3 committed offsets, got %d", len(committed))
	}
	if committed[0] != 100 || committed[1] != 101 || committed[2] != 102 {
		t.Errorf("unexpected committed offset sequence: %v", committed)
	}

	// 5. Assert telemetry:
	stats := mock.Stats()
	if stats.TotalConsumed != 3 || stats.TotalCommitted != 3 || stats.TotalErrors != 0 {
		t.Errorf("unexpected stats: %+v", stats)
	}
	if stats.LastOffset != 102 {
		t.Errorf("expected LastOffset=102, got %d", stats.LastOffset)
	}
}

// TestMockConsumer_ErrorHandlingAndSkipCommit verifies that if handler fails,
// the offset is NOT committed (At-Least-Once Delivery Invariant).
func TestMockConsumer_ErrorHandlingAndSkipCommit(t *testing.T) {
	ctx := context.Background()

	handlerErr := errors.New("database timeout")
	handler := func(ctx context.Context, msg kafka.Message) error {
		if msg.Offset == 501 {
			return handlerErr
		}
		return nil
	}

	mock := NewMockConsumer(handler)
	defer mock.Close()

	// Enqueue 2 messages: offset 500 should succeed, offset 501 should fail
	msg1 := kafka.Message{Topic: "test", Partition: 0, Offset: 500, Value: []byte("msg1")}
	msg2 := kafka.Message{Topic: "test", Partition: 0, Offset: 501, Value: []byte("msg2")}
	mock.Enqueue(msg1)
	mock.Enqueue(msg2)

	err := mock.Start(ctx)
	if !errors.Is(err, handlerErr) {
		t.Fatalf("expected handler error %v, got %v", handlerErr, err)
	}

	// Assert: ONLY offset 500 was committed! Offset 501 was NOT committed!
	committed := mock.GetCommittedOffsets()
	if len(committed) != 1 {
		t.Fatalf("expected 1 committed offset, got %d", len(committed))
	}
	if committed[0] != 500 {
		t.Errorf("expected committed offset 500, got %d", committed[0])
	}

	stats := mock.Stats()
	if stats.TotalErrors != 1 {
		t.Errorf("expected TotalErrors=1, got %d", stats.TotalErrors)
	}
}

// TestConsumerGroup_ConstructorLifecycle tests constructor guards and graceful termination.
func TestConsumerGroup_ConstructorLifecycle(t *testing.T) {
	cfg := DefaultConsumerConfig("test-group", "test-topic")

	// Nil handler check:
	_, err := NewConsumerGroup(cfg, nil)
	if err != ErrNilHandler {
		t.Fatalf("expected ErrNilHandler, got %v", err)
	}

	// Valid constructor:
	handler := func(ctx context.Context, msg kafka.Message) error { return nil }
	cg, err := NewConsumerGroup(cfg, handler)
	if err != nil {
		t.Fatalf("unexpected NewConsumerGroup error: %v", err)
	}
	if cg == nil {
		t.Fatal("expected non-nil ConsumerGroup")
	}

	// Close cleanly:
	if err := cg.Close(); err != nil {
		t.Errorf("unexpected Close error: %v", err)
	}

	// Idempotent double close:
	if err := cg.Close(); err != nil {
		t.Errorf("expected idempotent double close, got %v", err)
	}
}
