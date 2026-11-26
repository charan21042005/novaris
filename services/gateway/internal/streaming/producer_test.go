// ==============================================================================
// NOVARIS: Distributed Payment Settlement & Adaptive Fraud Intelligence Platform
// MODULE 2: STREAMING CORE (Day 16)
// FILE: services/gateway/internal/streaming/producer_test.go
// ==============================================================================

package streaming

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

// TestTransactionInitiatedEvent_Validation tests that all required domain invariants are enforced.
func TestTransactionInitiatedEvent_Validation(t *testing.T) {
	t.Run("nil event returns ErrNilEvent", func(t *testing.T) {
		var evt *TransactionInitiatedEvent
		err := evt.Validate()
		if err != ErrNilEvent {
			t.Fatalf("expected ErrNilEvent, got %v", err)
		}
	})

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

	t.Run("zero or negative amount fails validation", func(t *testing.T) {
		evt := &TransactionInitiatedEvent{
			TransactionID: uuid.New().String(),
			PayerID:       uuid.New().String(),
			PayeeID:       uuid.New().String(),
			AmountCents:   0,
		}
		if err := evt.Validate(); err != ErrInvalidAmount {
			t.Fatalf("expected ErrInvalidAmount for 0 cents, got %v", err)
		}

		evt.AmountCents = -500
		if err := evt.Validate(); err != ErrInvalidAmount {
			t.Fatalf("expected ErrInvalidAmount for negative cents, got %v", err)
		}
	})

	t.Run("valid event populates default metadata and serializes to JSON", func(t *testing.T) {
		txID := uuid.New().String()
		payerID := uuid.New().String()
		payeeID := uuid.New().String()

		evt := &TransactionInitiatedEvent{
			TransactionID:  txID,
			PayerID:        payerID,
			PayeeID:        payeeID,
			AmountCents:    7500,
			MerchantID:     "merchant_alpha",
			IdempotencyKey: uuid.New().String(),
		}

		rawJSON, err := evt.ToJSON()
		if err != nil {
			t.Fatalf("expected successful ToJSON, got %v", err)
		}

		// Verify fields populated by validation:
		if evt.EventID == "" {
			t.Error("expected generated EventID, got empty")
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
			t.Error("expected non-zero Timestamp")
		}

		// Verify round-trip unmarshaling:
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
	})
}

// TestMockProducer_Execution verifies the in-memory test producer behavior.
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
		AmountCents:    10000,
		MerchantID:     "merchant_alpha",
		IdempotencyKey: uuid.New().String(),
	}

	// 1. Publish event:
	res, err := mock.PublishTxInitiated(ctx, evt)
	if err != nil {
		t.Fatalf("unexpected PublishTxInitiated error: %v", err)
	}

	if res.Topic != TopicTxInitiated {
		t.Errorf("expected topic %s, got %s", TopicTxInitiated, res.Topic)
	}
	if res.Key != payerID {
		t.Errorf("expected key %s, got %s", payerID, res.Key)
	}

	// 2. Verify stored events:
	events := mock.GetEvents()
	if len(events) != 1 {
		t.Fatalf("expected 1 recorded event, got %d", len(events))
	}
	if events[0].TransactionID != txID {
		t.Errorf("expected TxID %s, got %s", txID, events[0].TransactionID)
	}

	// 3. Verify statistics:
	stats := mock.Stats()
	if stats.TotalPublished != 1 {
		t.Errorf("expected TotalPublished=1, got %d", stats.TotalPublished)
	}
	if stats.TotalFailed != 0 {
		t.Errorf("expected TotalFailed=0, got %d", stats.TotalFailed)
	}

	// 4. Test simulated failure:
	mock.FailNextWrite = true
	_, err = mock.PublishTxInitiated(ctx, evt)
	if err == nil {
		t.Fatal("expected error on FailNextWrite=true, got nil")
	}

	stats = mock.Stats()
	if stats.TotalFailed != 1 {
		t.Errorf("expected TotalFailed=1, got %d", stats.TotalFailed)
	}

	// 5. Test close prevents further writes:
	_ = mock.Close()
	_, err = mock.PublishTxInitiated(ctx, evt)
	if err != ErrProducerClosed {
		t.Fatalf("expected ErrProducerClosed, got %v", err)
	}
}

// TestKafkaProducer_Construction verifies that NewProducer initializes cleanly with defaults.
func TestKafkaProducer_Construction(t *testing.T) {
	cfg := LoadConfigFromEnv()
	producer, err := NewProducer(&cfg)
	if err != nil {
		t.Fatalf("failed to construct KafkaProducer: %v", err)
	}
	if producer == nil {
		t.Fatal("expected non-nil KafkaProducer")
	}

	// Verify stats initially zero:
	stats := producer.Stats()
	if stats.TotalPublished != 0 || stats.TotalFailed != 0 {
		t.Errorf("expected clean initial stats, got %+v", stats)
	}

	// Verify close:
	if err := producer.Close(); err != nil {
		t.Errorf("unexpected error on Close: %v", err)
	}

	// Double close should be idempotent:
	if err := producer.Close(); err != nil {
		t.Errorf("expected idempotent double Close, got %v", err)
	}

	// Publish on closed producer returns ErrProducerClosed:
	_, err = producer.Publish(context.Background(), "test.topic", "key", []byte("payload"))
	if err != ErrProducerClosed {
		t.Errorf("expected ErrProducerClosed after Close(), got %v", err)
	}
}
