// ==============================================================================
// NOVARIS: Distributed Payment Settlement & Adaptive Fraud Intelligence Platform
// MODULE 2: STREAMING CORE (Phase 2, Day 17: Settlement Consumer Engine)
// FILE: services/settlement/consumer.go
// ==============================================================================
//
// 📚 COMPUTER SCIENCE & DISTRIBUTED SYSTEMS MASTERCLASS FOR BEGINNERS:
//
// 1. WHAT DOES THE SETTLEMENT CONSUMER DO?
//    In Day 16, the API Gateway emitted `novaris.tx.initiated` events to Kafka.
//    Now in Day 17, this Settlement Consumer Service acts as the first decoupled
//    asynchronous worker in the NOVARIS ecosystem!
//    Instead of the gateway calling the settlement service over a slow, fragile
//    HTTP REST connection, this service listens directly to Kafka:
//      [Kafka: novaris.tx.initiated] ──> [Settlement Consumer] ──> [Ledger DB]
//
// 2. WHY ASYNCHRONOUS DECOUPLING MATTERS:
//    - Fault Tolerance: If the settlement database is undergoing maintenance or
//      momentarily locks up, the API Gateway NEVER notices! Payments are still
//      accepted at the gateway and safely buffered on Kafka's disk logs.
//      When the settlement worker comes back online, it resumes right where it left off.
//    - Independent Scaling: If 50,000 transactions flood in during Black Friday,
//      we can spin up 10 settlement worker pods in Kubernetes under consumer group
//      "novaris-settlement-workers". Kafka will automatically divide the partitions
//      among them, multiplying our processing throughput!
//
// 3. AT-LEAST-ONCE PROCESSING + IDEMPOTENCY = EFFECTIVELY-ONCE:
//    Every financial transaction event carries an `IdempotencyKey` and `TransactionID`.
//    If a network hiccup or crash occurs right after we update account balances but
//    before Kafka receives the offset commit, Kafka will re-deliver the message.
//    Our `ProcessedCache` catches this duplicate `TransactionID`, skips double-debiting,
//    and immediately commits the offset safely.
// ==============================================================================

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
)

// ProcessedRecord holds audit telemetry for an event processed by the settlement worker.
type ProcessedRecord struct {
	EventID       string    `json:"event_id"`
	TransactionID string    `json:"transaction_id"`
	PayerID       string    `json:"payer_id"`
	PayeeID       string    `json:"payee_id"`
	AmountCents   int64     `json:"amount_cents"`
	Currency      string    `json:"currency"`
	Partition     int       `json:"partition"`
	Offset        int64     `json:"offset"`
	ProcessedAt   time.Time `json:"processed_at"`
}

// SettlementProcessor coordinates business logic execution for incoming payment events.
type SettlementProcessor struct {
	mu          sync.RWMutex
	processed   map[string]ProcessedRecord // In-memory deduplication index (tx_id -> record)
	auditLog    []ProcessedRecord          // Append-only in-memory audit ledger
	totalDebits int64                      // Cumulative cents debited
	totalCredit int64                      // Cumulative cents credited
}

// NewSettlementProcessor instantiates a ready-to-use settlement business processor.
func NewSettlementProcessor() *SettlementProcessor {
	return &SettlementProcessor{
		processed: make(map[string]ProcessedRecord),
		auditLog:  make([]ProcessedRecord, 0),
	}
}

// HandleMessage decodes a raw Kafka message, validates the event schema, checks
// for idempotency duplicates, performs double-entry bookkeeping, and logs the outcome.
// Invariant: Returning nil triggers the consumer group to commit the offset to Kafka.
// Returning an error aborts the commit, allowing Kafka to redeliver the event.
func (p *SettlementProcessor) HandleMessage(ctx context.Context, msg kafka.Message) error {
	// 1. Deserialize message value into canonical TransactionInitiatedEvent:
	var evt TransactionInitiatedEvent
	if err := json.Unmarshal(msg.Value, &evt); err != nil {
		log.Printf("[SETTLEMENT_WARN] ❌ Poison-pill message at partition %d, offset %d: failed to deserialize JSON: %v",
			msg.Partition, msg.Offset, err)
		// Note: In Day 19 we route unparsable poison-pills to a Dead-Letter Queue (DLQ).
		// For Day 17, return error so the corrupt message is not silently ignored.
		return fmt.Errorf("settlement: unmarshal error: %w", err)
	}

	// 2. Enforce domain invariants:
	if err := evt.Validate(); err != nil {
		log.Printf("[SETTLEMENT_WARN] ❌ Invalid financial invariants for event %s: %v", evt.EventID, err)
		return fmt.Errorf("settlement: invariant validation failed: %w", err)
	}

	// 3. Thread-safe Idempotency Deduplication Check:
	p.mu.Lock()
	defer p.mu.Unlock()

	if prev, exists := p.processed[evt.TransactionID]; exists {
		log.Printf("[SETTLEMENT_IDEMP] ⚠️ Duplicate transaction %s detected (first seen at %v). Skipping double-settlement.",
			evt.TransactionID, prev.ProcessedAt.Format(time.RFC3339))
		// Return nil so the offset is committed and the consumer advances forward!
		return nil
	}

	// 4. Double-Entry Accounting Ledger Invariant:
	// Every payment transfer debits the payer and credits the payee by the identical magnitude:
	// sum(Debits) == sum(Credits)
	p.totalDebits += evt.AmountCents
	p.totalCredit += evt.AmountCents

	record := ProcessedRecord{
		EventID:       evt.EventID,
		TransactionID: evt.TransactionID,
		PayerID:       evt.PayerID,
		PayeeID:       evt.PayeeID,
		AmountCents:   evt.AmountCents,
		Currency:      evt.Currency,
		Partition:     msg.Partition,
		Offset:        msg.Offset,
		ProcessedAt:   time.Now().UTC(),
	}

	p.processed[evt.TransactionID] = record
	p.auditLog = append(p.auditLog, record)

	// 5. Structured Financial Logging (Fulfills Lab L2):
	// "Build a first consumer that logs every event it reads."
	dollars := float64(evt.AmountCents) / 100.0
	log.Printf("[SETTLEMENT_WORKER] 📥 Consumed tx.initiated event %s | TxID: %s | Amount: $%.2f %s | Payer: %s -> Payee: %s | Partition: %d | Offset: %d",
		evt.EventID, evt.TransactionID, dollars, evt.Currency, evt.PayerID, evt.PayeeID, msg.Partition, msg.Offset)

	return nil
}

// GetProcessedCount returns the total number of unique transactions processed.
func (p *SettlementProcessor) GetProcessedCount() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.processed)
}

// GetAuditLog returns a thread-safe copy of the processed records for testing assertions.
func (p *SettlementProcessor) GetAuditLog() []ProcessedRecord {
	p.mu.RLock()
	defer p.mu.RUnlock()
	copied := make([]ProcessedRecord, len(p.auditLog))
	copy(copied, p.auditLog)
	return copied
}

// TotalVolumes returns the cumulative debit and credit ledger magnitudes in cents.
func (p *SettlementProcessor) TotalVolumes() (int64, int64) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.totalDebits, p.totalCredit
}
