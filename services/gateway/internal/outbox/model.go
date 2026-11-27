// ==============================================================================
// NOVARIS: Distributed Payment Settlement & Adaptive Fraud Intelligence Platform
// MODULE 1: API Gateway Service - Transactional Outbox Domain Models (Day 18)
// FILE: services/gateway/internal/outbox/model.go
// ==============================================================================
//
// Architectural Principles & Invariants:
// 1. Dual-Write Elimination:
//    In a distributed microservice architecture, writing to a database and publishing
//    to a message broker in two independent network calls creates the classic
//    "Dual-Write Problem": if the service crashes between the two steps, or if the
//    broker is temporarily unreachable, state becomes permanently desynchronized.
//    The Transactional Outbox pattern eliminates this vulnerability by saving the
//    event inside the EXACT SAME ACID database transaction as the business write.
//
// 2. Immutability & Envelope Specification:
//    Each OutboxRecord captures the target Kafka topic, partition key, and serialized
//    JSON payload. Once written by the business transaction, its payload is immutable.
//
// 3. State Transition Lifecycle:
//    - PENDING:   The event is durably committed to the relational store and awaits relay.
//    - PUBLISHED: The event has been confirmed acknowledged by the Kafka cluster.
//    - FAILED:    The event exceeded maximum retry attempts (e.g. malformed or poisoned)
//                 and is flagged for operator inspection / Dead-Letter Queue (DLQ).
// ==============================================================================

package outbox

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Sentinel domain validation errors for outbox records.
var (
	// ErrNilRecord is returned when attempting to operate on a nil outbox pointer.
	ErrNilRecord = errors.New("outbox: record cannot be nil")

	// ErrEmptyAggregateID is returned when an outbox record lacks an aggregate identifier.
	ErrEmptyAggregateID = errors.New("outbox: aggregate_id cannot be empty")

	// ErrEmptyAggregateType is returned when an outbox record lacks an aggregate type.
	ErrEmptyAggregateType = errors.New("outbox: aggregate_type cannot be empty")

	// ErrEmptyEventType is returned when an outbox record lacks an event type / topic.
	ErrEmptyEventType = errors.New("outbox: event_type cannot be empty")

	// ErrEmptyPayload is returned when an outbox record has no event payload.
	ErrEmptyPayload = errors.New("outbox: payload cannot be empty")
)

// OutboxStatus represents the lifecycle state of an outbox event.
type OutboxStatus string

const (
	// StatusPending indicates the record is written to the database and waiting to be published.
	StatusPending OutboxStatus = "PENDING"

	// StatusPublished indicates the record has been successfully published to Kafka.
	StatusPublished OutboxStatus = "PUBLISHED"

	// StatusFailed indicates the record failed to publish after exhausting retry attempts.
	StatusFailed OutboxStatus = "FAILED"
)

// Aggregate types supported in the NOVARIS domain.
const (
	AggregateTransaction = "TRANSACTION"
	AggregateAccount     = "ACCOUNT"
	AggregateFraudReview = "FRAUD_REVIEW"
)

// OutboxRecord models the relational outbox table schema (PostgreSQL migration 001).
// Matches table `outbox` with columns:
//   id (UUID), aggregate_type (VARCHAR), aggregate_id (UUID), event_type (VARCHAR),
//   payload (JSONB), status (VARCHAR), retry_count (INT), created_at (TIMESTAMPTZ), published_at (TIMESTAMPTZ).
type OutboxRecord struct {
	// ID is the unique identifier (UUIDv4) of the outbox record.
	ID string `json:"id" db:"id"`

	// AggregateType is the domain aggregate boundary, e.g. "TRANSACTION".
	AggregateType string `json:"aggregate_type" db:"aggregate_type"`

	// AggregateID is the primary key of the aggregate (e.g. transaction UUID or account UUID).
	// Crucially, this is used as the Kafka Partition Key to guarantee FIFO order per entity!
	AggregateID string `json:"aggregate_id" db:"aggregate_id"`

	// EventType specifies the Kafka topic or event classification, e.g. "novaris.tx.initiated".
	EventType string `json:"event_type" db:"event_type"`

	// Payload contains the raw JSON-serialized canonical event bytes.
	Payload json.RawMessage `json:"payload" db:"payload"`

	// Status tracks the relay lifecycle: PENDING, PUBLISHED, or FAILED.
	Status OutboxStatus `json:"status" db:"status"`

	// RetryCount tracks the number of failed publishing attempts.
	RetryCount int `json:"retry_count" db:"retry_count"`

	// CreatedAt is the UTC timestamp when the record committed with the business transaction.
	CreatedAt time.Time `json:"created_at" db:"created_at"`

	// PublishedAt is the UTC timestamp when the Kafka broker acknowledged the event.
	PublishedAt *time.Time `json:"published_at,omitempty" db:"published_at"`
}

// Validate checks all structural and domain invariants of the outbox record.
func (r *OutboxRecord) Validate() error {
	if r == nil {
		return ErrNilRecord
	}

	if strings.TrimSpace(r.ID) == "" {
		r.ID = uuid.New().String()
	}

	if strings.TrimSpace(r.AggregateType) == "" {
		return ErrEmptyAggregateType
	}

	if strings.TrimSpace(r.AggregateID) == "" {
		return ErrEmptyAggregateID
	}

	if strings.TrimSpace(r.EventType) == "" {
		return ErrEmptyEventType
	}

	if len(r.Payload) == 0 || string(r.Payload) == "null" {
		return ErrEmptyPayload
	}

	if r.Status == "" {
		r.Status = StatusPending
	}

	if r.CreatedAt.IsZero() {
		r.CreatedAt = time.Now().UTC()
	}

	return nil
}

// NewOutboxRecord constructs and validates a new OutboxRecord from an arbitrary serializable payload.
func NewOutboxRecord(aggregateType, aggregateID, eventType string, eventData interface{}) (*OutboxRecord, error) {
	if strings.TrimSpace(aggregateType) == "" {
		return nil, ErrEmptyAggregateType
	}
	if strings.TrimSpace(aggregateID) == "" {
		return nil, ErrEmptyAggregateID
	}
	if strings.TrimSpace(eventType) == "" {
		return nil, ErrEmptyEventType
	}

	payloadBytes, err := json.Marshal(eventData)
	if err != nil {
		return nil, fmt.Errorf("outbox: marshaling event payload to JSON: %w", err)
	}

	record := &OutboxRecord{
		ID:            uuid.New().String(),
		AggregateType: aggregateType,
		AggregateID:   aggregateID,
		EventType:     eventType,
		Payload:       payloadBytes,
		Status:        StatusPending,
		RetryCount:    0,
		CreatedAt:     time.Now().UTC(),
	}

	if err := record.Validate(); err != nil {
		return nil, err
	}

	return record, nil
}
