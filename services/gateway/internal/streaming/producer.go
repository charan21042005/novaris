// ==============================================================================
// NOVARIS: Distributed Payment Settlement & Adaptive Fraud Intelligence Platform
// MODULE 2: STREAMING CORE & EVENT DRIVEN INFRASTRUCTURE (Phase 2, Day 16)
// FILE: services/gateway/internal/streaming/producer.go
// ==============================================================================
//
// 📚 COMPUTER SCIENCE & DISTRIBUTED SYSTEMS MASTERCLASS FOR BEGINNERS:
//
// 1. WHAT IS AN EVENT PRODUCER?
//    In traditional monoliths, when a payment succeeds, the system directly calls
//    the fraud engine, the email notification service, and the analytics database
//    via synchronous HTTP or RPC calls. If the email service is down or slow, the
//    entire payment request hangs or fails!
//    In an Event-Driven Architecture (EDA), an "Event Producer" (our Go Gateway)
//    does something revolutionary: it writes an immutable, append-only record called
//    an "Event" into a high-throughput commit log (Apache Kafka).
//    Once the event is safely recorded on disk by Kafka, the gateway immediately
//    returns HTTP 201 Created to the user.
//    Other microservices (Fraud, Settlement, Audit, Notifications) act as "Consumers",
//    reading this event independently at their own speed. If the email service is
//    down for 10 minutes, zero payments are affected—the email service simply catches
//    up when it reboots!
//
// 2. THE ANATOMY OF A FINANCIAL EVENT:
//    Financial events represent facts that have occurred in the past. Because the
//    past cannot be rewritten, events are 100% immutable:
//    - EventID: Unique UUIDv4 identifying this specific message envelope.
//    - EventType: Canonical domain identifier (e.g. "novaris.tx.initiated").
//    - Partition Key: The sender account UUID ("PayerID"). In distributed systems,
//      events with the SAME partition key are guaranteed to land on the SAME
//      partition on the SAME broker in exact chronological First-In, First-Out (FIFO)
//      order. This prevents balance race conditions (e.g. spending before depositing)!
//    - Payload: The complete financial facts (Amount, Currency, TransactionID,
//      MerchantID, IdempotencyKey, and UTC Timestamp).
//
// 3. UNDERSTANDING PRODUCER ACKNOWLEDGMENT LEVELS (RequiredAcks):
//    Kafka lets you choose your exact balance between latency and data durability:
//    - acks=0 (Fire & Forget / RequireNone):
//      The producer sends the TCP packet and immediately continues without waiting
//      for a response from the broker.
//      Throughput is maximum, but if the broker crashes before writing to disk,
//      the event is permanently lost. (Unacceptable for money!).
//    - acks=1 (Leader ACK / RequireOne — NOVARIS Day 16 Production Standard):
//      The producer sends the message and waits until the Partition Leader broker
//      writes the record to its local memory/log.
//      Latency is ultra-fast (<15ms), and provides rock-solid defense against network
//      packet drops or transient disconnections.
//    - acks=-1 / acks=all (Quorum ACK / RequireAll):
//      The producer waits until the Leader AND every In-Sync Replica (ISR) follower
//      have mirrored the record.
//      Provides mathematical zero-data-loss even if the physical server hosting the
//      leader broker literally loses power.
//
// 4. MICRO-BATCHING & RECORD ACCUMULATION:
//    Making an individual operating system syscall (`sendmsg`) over TCP for every
//    200-byte JSON message creates massive CPU overhead and network congestion.
//    Instead, the Go producer utilizes a "Record Accumulator" memory buffer:
//    - BatchSize (100 records): If 100 payments arrive simultaneously, they are packed
//      into a single TCP packet.
//    - BatchTimeout (10ms): If traffic is light, the producer waits at most 10ms for
//      more records to arrive before transmitting the batch anyway.
//    This micro-batching pattern delivers the best of both worlds: extreme throughput
//    (tens of thousands of msgs/sec) alongside sub-15ms single-payment latency!
//
// 5. DETERMINISTIC PARTITIONING (MURMUR2 HASHING):
//    How does Kafka decide which partition (0, 1, or 2) receives an event?
//    Kafka runs the Murmur2 32-bit hash algorithm over the Partition Key (PayerID):
//      Partition Index = |Murmur2(PayerID)| % NumberOfPartitions
//    Because Murmur2 is a deterministic mathematical function, the same PayerID
//    will ALWAYS produce the exact same partition index, ensuring account-level FIFO!
// ==============================================================================

package streaming

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

// ==============================================================================
// DOMAIN CONSTANTS & CANONICAL EVENT SCHEMAS
// ==============================================================================

const (
	// EventTypeTxInitiated is emitted immediately when the API Gateway accepts a valid
	// payment request and writes it to the ledger database:
	EventTypeTxInitiated = "novaris.tx.initiated"

	// EventTypeTxSettled is emitted when funds are permanently transferred between accounts:
	EventTypeTxSettled = "novaris.tx.settled"

	// EventTypeTxFailed is emitted when a payment is rejected (e.g. insufficient funds, fraud):
	EventTypeTxFailed = "novaris.tx.failed"

	// EventSchemaVersion1 specifies version 1.0 of the canonical event JSON contract:
	EventSchemaVersion1 = "1.0"
)

// ==============================================================================
// SENTINEL ERROR DEFINITIONS (DEFENSIVE ERROR HANDLING)
// ==============================================================================

var (
	// ErrNilEvent is returned when attempting to publish or validate a nil pointer:
	ErrNilEvent = errors.New("streaming: event cannot be nil")

	// ErrMissingTxID is returned when the required transaction UUID is empty:
	ErrMissingTxID = errors.New("streaming: transaction_id is required")

	// ErrMissingPayerID is returned when the sender account UUID (Partition Key) is empty:
	ErrMissingPayerID = errors.New("streaming: payer_id (partition key) is required")

	// ErrMissingPayeeID is returned when the recipient account UUID is empty:
	ErrMissingPayeeID = errors.New("streaming: payee_id is required")

	// ErrInvalidAmount is returned when amount_cents is zero or negative:
	ErrInvalidAmount = errors.New("streaming: amount_cents must be strictly positive")

	// ErrProducerClosed is returned when attempting to publish through a terminated producer:
	ErrProducerClosed = errors.New("streaming: producer is closed")

	// ErrBrokerUnreachable is returned when TCP connection to all Kafka brokers fails:
	ErrBrokerUnreachable = errors.New("streaming: kafka broker unreachable")
)

// ==============================================================================
// CANONICAL EVENT MODEL (TransactionInitiatedEvent)
// ==============================================================================

// TransactionInitiatedEvent models the canonical immutable financial event
// emitted by the API Gateway immediately upon accepting a valid payment request.
// Every downstream consumer in NOVARIS (Fraud Engine, Settlement, Audit) consumes
// this strongly typed event model.
type TransactionInitiatedEvent struct {
	// EventID is a globally unique UUIDv4 identifying this specific message envelope:
	EventID string `json:"event_id"`

	// EventType identifies the canonical domain classification ("novaris.tx.initiated"):
	EventType string `json:"event_type"`

	// Version identifies the schema contract version ("1.0") for backward compatibility:
	Version string `json:"version"`

	// TransactionID is the canonical UUID of the underlying ledger transaction:
	TransactionID string `json:"transaction_id"`

	// PayerID is the sender account UUID, serving as the Kafka Partition Key.
	// Invariant: Murmur2(PayerID) routes all events for this account to the same partition!
	PayerID string `json:"payer_id"`

	// PayeeID is the recipient account UUID receiving the funds:
	PayeeID string `json:"payee_id"`

	// AmountCents represents the transfer magnitude in integer cents ($50.00 = 5000):
	// Note: We always use integer cents to prevent floating-point rounding errors!
	AmountCents int64 `json:"amount_cents"`

	// Currency is the standard ISO 4217 three-letter currency code (e.g. "USD"):
	Currency string `json:"currency"`

	// MerchantID is the authenticated tenant or client submitting the transaction:
	MerchantID string `json:"merchant_id"`

	// IdempotencyKey is the client-supplied deduplication identifier:
	IdempotencyKey string `json:"idempotency_key,omitempty"`

	// Timestamp is the exact UTC timestamp when the gateway accepted the request:
	Timestamp time.Time `json:"timestamp"`

	// Status reflects the initial transaction lifecycle state ("INITIATED"):
	Status string `json:"status"`

	// Metadata carries optional tracing contexts, correlation IDs, or client user-agents:
	Metadata map[string]string `json:"metadata,omitempty"`
}

// Validate inspects the event struct to ensure all mandatory financial invariants
// are strictly satisfied before the message is allowed into the Kafka network pipeline.
// If any optional fields are omitted, sensible defaults are automatically populated.
func (e *TransactionInitiatedEvent) Validate() error {
	// 1. Guard against nil pointer dereference:
	if e == nil {
		return ErrNilEvent
	}

	// 2. Validate mandatory business keys:
	if e.TransactionID == "" {
		return ErrMissingTxID
	}
	if e.PayerID == "" {
		return ErrMissingPayerID
	}
	if e.PayeeID == "" {
		return ErrMissingPayeeID
	}

	// 3. Validate financial magnitude: Money cannot be zero or negative!
	if e.AmountCents <= 0 {
		return ErrInvalidAmount
	}

	// 4. Populate default envelope fields if missing:
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
	if e.Status == "" {
		e.Status = "INITIATED"
	}
	if e.Timestamp.IsZero() {
		e.Timestamp = time.Now().UTC()
	}

	return nil
}

// ToJSON validates the event and serializes it into canonical JSON byte format
// ready for wire transmission over the Kafka TCP protocol.
func (e *TransactionInitiatedEvent) ToJSON() ([]byte, error) {
	if err := e.Validate(); err != nil {
		return nil, err
	}
	return json.Marshal(e)
}

// ==============================================================================
// TELEMETRY & RUNTIME METRICS STRUCTS
// ==============================================================================

// PublishResult encapsulates telemetry and broker feedback from a successful write.
type PublishResult struct {
	// Topic is the name of the Kafka topic where the message was written:
	Topic string `json:"topic"`

	// Key is the partition key string (e.g. Payer Account UUID):
	Key string `json:"key"`

	// Partition is the specific partition index (e.g. 0, 1, or 2) chosen by Murmur2:
	Partition int `json:"partition"`

	// Offset is the append-only log index assigned by the Kafka broker:
	Offset int64 `json:"offset"`

	// Timestamp is the UTC timestamp assigned to the message:
	Timestamp time.Time `json:"timestamp"`

	// Latency is the round-trip duration from TCP write to Leader ACK confirmation:
	Latency time.Duration `json:"latency_ns"`
}

// ProducerStats tracks real-time operational metrics for gateway telemetry dashboards.
type ProducerStats struct {
	// TotalPublished is the cumulative count of successfully acknowledged events:
	TotalPublished int64 `json:"total_published"`

	// TotalFailed is the cumulative count of write errors:
	TotalFailed int64 `json:"total_failed"`

	// LastLatencyNs is the round-trip write latency in nanoseconds of the most recent event:
	LastLatencyNs int64 `json:"last_latency_ns"`

	// LastPublished is the UTC timestamp of the most recent successful publication:
	LastPublished time.Time `json:"last_published_at"`
}

// ==============================================================================
// EVENT PRODUCER CONTRACT (INTERFACE)
// ==============================================================================

// EventProducer specifies the contractual interface for publishing domain events.
// Decoupling via an interface allows the API Gateway to seamlessly switch between:
//  1. KafkaProducer: The real production client speaking TCP to live Kafka brokers.
//  2. MockProducer: An in-memory thread-safe mock for offline unit testing.
type EventProducer interface {
	// PublishTxInitiated serializes and publishes a TransactionInitiatedEvent:
	PublishTxInitiated(ctx context.Context, event *TransactionInitiatedEvent) (*PublishResult, error)

	// Publish writes a raw keyed payload to a specified target topic:
	Publish(ctx context.Context, topic string, key string, payload []byte) (*PublishResult, error)

	// Close flushes in-flight memory buffers and terminates TCP connections:
	Close() error

	// Stats returns real-time publication telemetry:
	Stats() ProducerStats
}

// ==============================================================================
// PRODUCTION KAFKA PRODUCER IMPLEMENTATION
// ==============================================================================

// KafkaProducer is the enterprise-grade implementation of EventProducer wrapping
// segmentio/kafka-go. It manages socket connections, record batching, and metrics.
type KafkaProducer struct {
	cfg     *Config
	writer  *kafka.Writer
	closed  uint32       // Atomic flag: 0 = open, 1 = closed (prevents race conditions)
	stats   ProducerStats
	statsMu sync.RWMutex // Protects stats struct during concurrent reads and writes
}

// NewProducer instantiates and configures a high-performance, resilient Kafka Producer.
// It applies enterprise defaults: Murmur2 balancing, Leader ACKs, and micro-batching.
func NewProducer(cfg *Config) (*KafkaProducer, error) {
	if cfg == nil {
		c := LoadConfigFromEnv()
		cfg = &c
	}

	// Configure segmentio/kafka-go Writer with enterprise streaming defaults:
	// Architectural Decision: The Topic field is intentionally left unset at the
	// Writer level so individual messages can dynamically target different topics
	// (e.g. "novaris.tx.initiated", "novaris.tx.settled", "novaris.dlq").
	writer := &kafka.Writer{
		Addr:         kafka.TCP(cfg.Brokers...),
		Balancer:     &kafka.Murmur2Balancer{}, // 100% binary-compatible with Apache Kafka Java client
		RequiredAcks: kafka.RequireOne,        // Leader ACK confirmation for sub-15ms durability
		MaxAttempts:  3,                       // Up to 3 automatic retries on transient network hiccups
		BatchSize:    100,                     // Micro-batch buffer: up to 100 records per TCP packet
		BatchTimeout: 10 * time.Millisecond,   // Flush incomplete batches every 10ms to maintain low latency
		WriteTimeout: cfg.DialTimeout,         // Fail-fast timeout on network partitions
		ReadTimeout:  cfg.DialTimeout,
		Async:        false,                   // Synchronous mode: wait for Leader ACK before returning
	}

	return &KafkaProducer{
		cfg:    cfg,
		writer: writer,
	}, nil
}

// Publish writes a raw keyed record to the target Kafka topic and awaits Leader ACK.
// It records latency telemetry and updates operational statistics thread-safely.
func (p *KafkaProducer) Publish(ctx context.Context, topic string, key string, payload []byte) (*PublishResult, error) {
	// 1. Guard against writes to a closed producer:
	if atomic.LoadUint32(&p.closed) == 1 {
		return nil, ErrProducerClosed
	}

	start := time.Now()

	// 2. Construct the Kafka message envelope with trace headers:
	msg := kafka.Message{
		Topic: topic,
		Key:   []byte(key),   // Murmur2 hash uses this key to determine partition
		Value: payload,       // Canonical JSON payload
		Time:  time.Now().UTC(),
		Headers: []kafka.Header{
			{Key: "client_id", Value: []byte(p.cfg.ClientID)},
			{Key: "version", Value: []byte(EventSchemaVersion1)},
		},
	}

	// 3. Transmit message over TCP to the Kafka partition leader:
	err := p.writer.WriteMessages(ctx, msg)
	latency := time.Since(start)

	// 4. Update operational statistics under mutex protection:
	p.statsMu.Lock()
	if err != nil {
		p.stats.TotalFailed++
		p.statsMu.Unlock()
		return nil, fmt.Errorf("streaming: failed to write message to topic %q: %w", topic, err)
	}

	p.stats.TotalPublished++
	p.stats.LastLatencyNs = latency.Nanoseconds()
	p.stats.LastPublished = time.Now().UTC()
	p.statsMu.Unlock()

	// 5. Return success telemetry to caller:
	return &PublishResult{
		Topic:     topic,
		Key:       key,
		Timestamp: msg.Time,
		Latency:   latency,
	}, nil
}

// PublishTxInitiated validates, serializes, and publishes a transaction event.
// Uses event.PayerID as the partition key to guarantee per-account FIFO ordering.
func (p *KafkaProducer) PublishTxInitiated(ctx context.Context, event *TransactionInitiatedEvent) (*PublishResult, error) {
	// 1. Enforce business rules and domain invariants:
	if err := event.Validate(); err != nil {
		return nil, fmt.Errorf("streaming: invalid transaction event: %w", err)
	}

	// 2. Serialize event struct to canonical JSON:
	payload, err := event.ToJSON()
	if err != nil {
		return nil, fmt.Errorf("streaming: serialization error: %w", err)
	}

	// 3. Publish to canonical topic using PayerID as partition key:
	// Invariant: All transactions initiated by the same account will land on the exact
	// same partition, guaranteeing chronological First-In, First-Out (FIFO) ordering!
	return p.Publish(ctx, TopicTxInitiated, event.PayerID, payload)
}

// Close gracefully drains in-flight message batches and closes all open TCP sockets.
// Uses atomic CompareAndSwapUint32 to guarantee idempotent, race-free termination.
func (p *KafkaProducer) Close() error {
	if !atomic.CompareAndSwapUint32(&p.closed, 0, 1) {
		return nil // Already closed by another goroutine
	}
	return p.writer.Close()
}

// Stats returns a thread-safe snapshot of real-time publication metrics.
func (p *KafkaProducer) Stats() ProducerStats {
	p.statsMu.RLock()
	defer p.statsMu.RUnlock()
	return p.stats
}

// ==============================================================================
// IN-MEMORY MOCK PRODUCER (FOR ISOLATED UNIT TESTING & OFFLINE CI/CD)
// ==============================================================================

// MockProducer provides an in-memory thread-safe implementation of EventProducer.
// It allows unit tests across the entire NOVARIS repository to verify event emission,
// inspect published envelopes, and inject failure scenarios without requiring Docker.
type MockProducer struct {
	mu            sync.RWMutex
	Events        []*TransactionInitiatedEvent
	RawMessages   []kafka.Message
	FailNextWrite bool
	closed        bool
	stats         ProducerStats
}

// NewMockProducer instantiates a ready-to-use in-memory test producer.
func NewMockProducer() *MockProducer {
	return &MockProducer{
		Events:      make([]*TransactionInitiatedEvent, 0),
		RawMessages: make([]kafka.Message, 0),
	}
}

// PublishTxInitiated records the event in memory after running validation.
func (m *MockProducer) PublishTxInitiated(ctx context.Context, event *TransactionInitiatedEvent) (*PublishResult, error) {
	if err := event.Validate(); err != nil {
		return nil, err
	}

	payload, err := event.ToJSON()
	if err != nil {
		return nil, err
	}

	return m.Publish(ctx, TopicTxInitiated, event.PayerID, payload)
}

// Publish simulates writing to Kafka by appending to thread-safe in-memory slices.
func (m *MockProducer) Publish(ctx context.Context, topic string, key string, payload []byte) (*PublishResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return nil, ErrProducerClosed
	}

	// Simulate error injection for resilience testing:
	if m.FailNextWrite {
		m.stats.TotalFailed++
		m.FailNextWrite = false
		return nil, errors.New("mock: deliberate simulated write failure")
	}

	msg := kafka.Message{
		Topic: topic,
		Key:   []byte(key),
		Value: payload,
		Time:  time.Now().UTC(),
	}
	m.RawMessages = append(m.RawMessages, msg)

	// If the payload is a TransactionInitiatedEvent, decode and record it:
	var evt TransactionInitiatedEvent
	if err := json.Unmarshal(payload, &evt); err == nil {
		m.Events = append(m.Events, &evt)
	}

	m.stats.TotalPublished++
	m.stats.LastLatencyNs = 100_000 // 0.1ms mock latency
	m.stats.LastPublished = msg.Time

	return &PublishResult{
		Topic:     topic,
		Key:       key,
		Timestamp: msg.Time,
		Latency:   100 * time.Microsecond,
	}, nil
}

// Close marks the mock producer as closed.
func (m *MockProducer) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	return nil
}

// Stats returns thread-safe mock metrics.
func (m *MockProducer) Stats() ProducerStats {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.stats
}

// GetEvents returns a thread-safe copy of all published events for test assertions.
func (m *MockProducer) GetEvents() []*TransactionInitiatedEvent {
	m.mu.RLock()
	defer m.mu.RUnlock()
	copied := make([]*TransactionInitiatedEvent, len(m.Events))
	copy(copied, m.Events)
	return copied
}

// Clear resets recorded events between test iterations.
func (m *MockProducer) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Events = m.Events[:0]
	m.RawMessages = m.RawMessages[:0]
}
