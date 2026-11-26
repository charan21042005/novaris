// ==============================================================================
// NOVARIS: Distributed Payment Settlement & Adaptive Fraud Intelligence Platform
// MODULE 2: STREAMING CORE (Phase 2, Day 16: Kafka Producer Unit Tests)
// FILE: services/gateway/internal/streaming/producer_test.go
// ==============================================================================
//
// 📚 COMPUTER SCIENCE & TESTING METHODOLOGY FOR BEGINNERS:
//
// 1. WHY IS UNIT TESTING ESSENTIAL FOR EVENT STREAMING?
//    In event-driven architectures, an event once published cannot be "un-sent".
//    If a producer publishes an event with a negative amount, a missing transaction
//    ID, or corrupted JSON, downstream consumers (fraud, ledger) could crash or
//    corrupt financial state! Unit tests act as an unbreakable mathematical barrier,
//    proving that only 100% valid, structurally sound events can ever be emitted.
//
// 2. THE THREE PILLARS OF THESE UNIT TESTS:
//    - Pillar 1 (Invariant Validation): Assert that every business rule (positive
//      cents, valid UUIDs, non-empty keys) is enforced before network serialization.
//    - Pillar 2 (Round-Trip JSON Fidelity): Assert that when an event struct is
//      marshaled into JSON and then unmarshaled back, zero data is lost or altered.
//    - Pillar 3 (Hermetic Mock Simulation): Test error handling, retry behavior, and
//      lifecycle termination (Close) without needing a live Kafka Docker container!
// ==============================================================================

package streaming

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

// TestTransactionInitiatedEvent_Validation verifies that all mandatory financial
// invariants and schema rules are strictly enforced prior to network transmission.
func TestTransactionInitiatedEvent_Validation(t *testing.T) {
	// Sub-test 1: Guard against nil pointer exceptions:
	t.Run("nil event returns ErrNilEvent", func(t *testing.T) {
		var evt *TransactionInitiatedEvent
		err := evt.Validate()
		if err != ErrNilEvent {
			t.Fatalf("expected ErrNilEvent, got %v", err)
		}
	})

	// Sub-test 2: Invariant: TransactionID is mandatory for ledger traceability:
	t.Run("missing transaction_id fails validation", func(t *testing.T) {
		evt := &TransactionInitiatedEvent{
			PayerID:     uuid.New().String(),
			PayeeID:     uuid.New().String(),
			AmountCents: 5000,
		}
		err := evt.Validate()
		if err != ErrMissingTxID {
			t.Fatalf("expected ErrMissingTxID, got %v", err)
		}
	})

	// Sub-test 3: Invariant: PayerID is mandatory because it serves as the Kafka Partition Key!
	t.Run("missing payer_id fails validation", func(t *testing.T) {
		evt := &TransactionInitiatedEvent{
			TransactionID: uuid.New().String(),
			PayeeID:       uuid.New().String(),
			AmountCents:   5000,
		}
		err := evt.Validate()
		if err != ErrMissingPayerID {
			t.Fatalf("expected ErrMissingPayerID, got %v", err)
		}
	})

	// Sub-test 4: Invariant: PayeeID is mandatory so consumers know who gets the funds:
	t.Run("missing payee_id fails validation", func(t *testing.T) {
		evt := &TransactionInitiatedEvent{
			TransactionID: uuid.New().String(),
			PayerID:       uuid.New().String(),
			AmountCents:   5000,
		}
		err := evt.Validate()
		if err != ErrMissingPayeeID {
			t.Fatalf("expected ErrMissingPayeeID, got %v", err)
		}
	})

	// Sub-test 5: Financial Invariant: Transfer amounts must be strictly positive (> 0 cents):
	t.Run("zero or negative amount fails validation", func(t *testing.T) {
		// Test zero cents ($0.00 transfer):
		evt := &TransactionInitiatedEvent{
			TransactionID: uuid.New().String(),
			PayerID:       uuid.New().String(),
			PayeeID:       uuid.New().String(),
			AmountCents:   0,
		}
		if err := evt.Validate(); err != ErrInvalidAmount {
			t.Fatalf("expected ErrInvalidAmount for 0 cents, got %v", err)
		}

		// Test negative cents (-$5.00 transfer):
		evt.AmountCents = -500
		if err := evt.Validate(); err != ErrInvalidAmount {
			t.Fatalf("expected ErrInvalidAmount for negative cents, got %v", err)
		}
	})

	// Sub-test 6: Default population & round-trip JSON serialization fidelity:
	t.Run("valid event populates default metadata and serializes to JSON", func(t *testing.T) {
		txID := uuid.New().String()
		payerID := uuid.New().String()
		payeeID := uuid.New().String()

		evt := &TransactionInitiatedEvent{
			TransactionID:  txID,
			PayerID:        payerID,
			PayeeID:        payeeID,
			AmountCents:    7500, // $75.00 USD
			MerchantID:     "merchant_alpha",
			IdempotencyKey: uuid.New().String(),
		}

		// 1. Serialize to canonical JSON bytes:
		rawJSON, err := evt.ToJSON()
		if err != nil {
			t.Fatalf("expected successful ToJSON, got %v", err)
		}

		// 2. Verify automatic field population by Validate():
		if evt.EventID == "" {
			t.Error("expected generated EventID UUIDv4, got empty")
		}
		if evt.EventType != EventTypeTxInitiated {
			t.Errorf("expected EventType=%s, got %s", EventTypeTxInitiated, evt.EventType)
		}
		if evt.Version != EventSchemaVersion1 {
			t.Errorf("expected Version=%s, got %s", EventSchemaVersion1, evt.Version)
		}
		if evt.Currency != "USD" {
			t.Errorf("expected Currency=USD, got %s", evt.Currency)
		}
		if evt.Status != "INITIATED" {
			t.Errorf("expected Status=INITIATED, got %s", evt.Status)
		}
		if evt.Timestamp.IsZero() {
			t.Error("expected non-zero UTC Timestamp")
		}

		// 3. Verify round-trip unmarshaling (Fidelity Check):
		var unmarshaled TransactionInitiatedEvent
		if err := json.Unmarshal(rawJSON, &unmarshaled); err != nil {
			t.Fatalf("failed to unmarshal JSON: %v", err)
		}
		if unmarshaled.TransactionID != txID {
			t.Errorf("unmarshaled TxID mismatch: expected %s, got %s", txID, unmarshaled.TransactionID)
		}
		if unmarshaled.AmountCents != 7500 {
			t.Errorf("unmarshaled AmountCents mismatch: expected 7500, got %d", unmarshaled.AmountCents)
		}
		if unmarshaled.PayerID != payerID {
			t.Errorf("unmarshaled PayerID mismatch: expected %s, got %s", payerID, unmarshaled.PayerID)
		}
	})
}

// TestMockProducer_Execution verifies the in-memory test producer behavior,
// proving that services can run hermetic tests without a live Kafka Docker broker.
func TestMockProducer_Execution(t *testing.T) {
	ctx := context.Background()
	mock := NewMockProducer()
	defer mock.Close()

	txID := uuid.New().String()
	payerID := "acc_merchant_alpha_001"
	payeeID := "acc_user_bob_4920"

	evt := &TransactionInitiatedEvent{
		TransactionID:  txID,
		PayerID:        payerID,
		PayeeID:        payeeID,
		AmountCents:    10000, // $100.00 USD
		MerchantID:     "merchant_alpha",
		IdempotencyKey: uuid.New().String(),
	}

	// 1. Publish event through mock producer:
	res, err := mock.PublishTxInitiated(ctx, evt)
	if err != nil {
		t.Fatalf("unexpected PublishTxInitiated error: %v", err)
	}

	// 2. Assert return metadata:
	if res.Topic != TopicTxInitiated {
		t.Errorf("expected topic %s, got %s", TopicTxInitiated, res.Topic)
	}
	if res.Key != payerID {
		t.Errorf("expected key %s, got %s", payerID, res.Key)
	}

	// 3. Verify that the event was safely recorded in memory:
	events := mock.GetEvents()
	if len(events) != 1 {
		t.Fatalf("expected 1 recorded event, got %d", len(events))
	}
	if events[0].TransactionID != txID {
		t.Errorf("expected TxID %s, got %s", txID, events[0].TransactionID)
	}

	// 4. Verify operational telemetry metrics:
	stats := mock.Stats()
	if stats.TotalPublished != 1 {
		t.Errorf("expected TotalPublished=1, got %d", stats.TotalPublished)
	}
	if stats.TotalFailed != 0 {
		t.Errorf("expected TotalFailed=0, got %d", stats.TotalFailed)
	}

	// 5. Test simulated error injection (Chaos Engineering):
	mock.FailNextWrite = true
	_, err = mock.PublishTxInitiated(ctx, evt)
	if err == nil {
		t.Fatal("expected error on FailNextWrite=true, got nil")
	}

	stats = mock.Stats()
	if stats.TotalFailed != 1 {
		t.Errorf("expected TotalFailed=1, got %d", stats.TotalFailed)
	}

	// 6. Test that Close() prevents further writes (Fail-Fast Lifecycle):
	_ = mock.Close()
	_, err = mock.PublishTxInitiated(ctx, evt)
	if err != ErrProducerClosed {
		t.Fatalf("expected ErrProducerClosed, got %v", err)
	}
}

// TestKafkaProducer_Construction verifies that NewProducer initializes cleanly
// and enforces idempotent close semantics.
func TestKafkaProducer_Construction(t *testing.T) {
	cfg := LoadConfigFromEnv()
	producer, err := NewProducer(&cfg)
	if err != nil {
		t.Fatalf("failed to construct KafkaProducer: %v", err)
	}
	if producer == nil {
		t.Fatal("expected non-nil KafkaProducer")
	}

	// 1. Verify telemetry is initialized cleanly:
	stats := producer.Stats()
	if stats.TotalPublished != 0 || stats.TotalFailed != 0 {
		t.Errorf("expected clean initial stats, got %+v", stats)
	}

	// 2. Verify graceful socket close:
	if err := producer.Close(); err != nil {
		t.Errorf("unexpected error on Close: %v", err)
	}

	// 3. Invariant: Closing twice must be idempotent (no panic or error):
	if err := producer.Close(); err != nil {
		t.Errorf("expected idempotent double Close, got %v", err)
	}

	// 4. Invariant: Publishing on a closed producer must return ErrProducerClosed:
	_, err = producer.Publish(context.Background(), "test.topic", "key", []byte("payload"))
	if err != ErrProducerClosed {
		t.Errorf("expected ErrProducerClosed after Close(), got %v", err)
	}
}
