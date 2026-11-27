// ==============================================================================
// NOVARIS: Distributed Payment Settlement & Adaptive Fraud Intelligence Platform
// MODULE 2: STREAMING CORE (Phase 2, Day 17: Canonical Settlement Event Model)
// FILE: services/settlement/event.go
// ==============================================================================
//
// 📚 COMPUTER SCIENCE & EVENT CONTRACT DISCIPLINE FOR BEGINNERS:
//
// 1. WHAT IS AN EVENT SCHEMA CONTRACT?
//    In microservice ecosystems, independent services (Gateway, Settlement, Fraud)
//    do NOT share database tables or memory. Instead, they agree on a standardized
//    data contract (the "Event Schema").
//    The Gateway produces this JSON envelope, and the Settlement Service consumes it.
//    Because services must be loosely coupled, the Settlement Service defines its
//    own domain representation of the contract without importing internal gateway packages!
//
// 2. THE CANONICAL FINANCIAL EVENT INVARIANTS:
//    - EventID: Globally unique UUIDv4 identifying this specific message instance.
//    - EventType: Canonical domain classification ("novaris.tx.initiated").
//    - TransactionID: The canonical UUID of the ledger transaction.
//    - PayerID: Sender account UUID (used as Kafka Partition Key for FIFO ordering).
//    - PayeeID: Recipient account UUID.
//    - AmountCents: Strictly positive magnitude in integer cents ($100.00 = 10000).
// ==============================================================================

package main

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

// Standard topic and event identifiers:
const (
	TopicTxInitiated     = "novaris.tx.initiated"
	EventTypeTxInitiated = "novaris.tx.initiated"
	EventSchemaVersion1  = "1.0"
)

// Sentinel validation errors:
var (
	ErrNilEvent       = errors.New("settlement: event cannot be nil")
	ErrMissingTxID    = errors.New("settlement: transaction_id is required")
	ErrMissingPayerID = errors.New("settlement: payer_id is required")
	ErrMissingPayeeID = errors.New("settlement: payee_id is required")
	ErrInvalidAmount  = errors.New("settlement: amount_cents must be strictly positive")
)

// TransactionInitiatedEvent models the canonical financial event consumed by the settlement engine.
type TransactionInitiatedEvent struct {
	EventID        string            `json:"event_id"`
	EventType      string            `json:"event_type"`
	Version        string            `json:"version"`
	TransactionID  string            `json:"transaction_id"`
	PayerID        string            `json:"payer_id"`
	PayeeID        string            `json:"payee_id"`
	AmountCents    int64             `json:"amount_cents"`
	Currency       string            `json:"currency"`
	MerchantID     string            `json:"merchant_id"`
	IdempotencyKey string            `json:"idempotency_key,omitempty"`
	Timestamp      time.Time         `json:"timestamp"`
	Status         string            `json:"status"`
	Metadata       map[string]string `json:"metadata,omitempty"`
}

// Validate verifies that all mandatory invariants are satisfied before settlement.
func (e *TransactionInitiatedEvent) Validate() error {
	if e == nil {
		return ErrNilEvent
	}
	if e.TransactionID == "" {
		return ErrMissingTxID
	}
	if e.PayerID == "" {
		return ErrMissingPayerID
	}
	if e.PayeeID == "" {
		return ErrMissingPayeeID
	}
	if e.AmountCents <= 0 {
		return ErrInvalidAmount
	}
	if e.EventID == "" {
		e.EventID = uuid.New().String()
	}
	if e.EventType == "" {
		e.EventType = EventTypeTxInitiated
	}
	if e.Version == "" {
		e.Version = EventSchemaVersion1
	}
	if e.Currency == "" {
		e.Currency = "USD"
	}
	return nil
}

// ToJSON serializes the event struct into canonical JSON bytes.
func (e *TransactionInitiatedEvent) ToJSON() ([]byte, error) {
	if err := e.Validate(); err != nil {
		return nil, err
	}
	return json.Marshal(e)
}
