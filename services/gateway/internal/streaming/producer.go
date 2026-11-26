// ==============================================================================
// NOVARIS: Distributed Payment Settlement & Adaptive Fraud Intelligence Platform
// MODULE 2: STREAMING CORE & EVENT DRIVEN INFRASTRUCTURE (Day 16)
// FILE: services/gateway/internal/streaming/producer.go
// ==============================================================================
//
// 📚 COMPUTER SCIENCE & DISTRIBUTED SYSTEMS MASTERCLASS FOR BEGINNERS:
//
// 1. WHAT IS AN EVENT PRODUCER?
//    In distributed architectures, an "Event Producer" is a client application
//    (like our Go API Gateway) that creates and publishes immutable records
//    ("events") to a distributed log (Apache Kafka).
//    Unlike synchronous REST or RPC calls—where the caller blocks waiting for
//    downstream services to complete—the producer publishes the event and
//    receives an immediate durability acknowledgment (ACK) from the broker.
//    Downstream consumers (fraud detection, audit ledger, notifications) react
//    to the event asynchronously at their own decoupled pace.
//
// 2. THE ANATOMY OF A PAYMENT EVENT:
//    Financial events must be self-contained and immutable:
//    - EventID: Unique UUIDv4 identifying this specific message instance.
//    - EventType: Canonical domain name (e.g., "novaris.tx.initiated").
//    - Partition Key: The sender account UUID ("PayerID"). This guarantees that
//      all events for the same account route to the same Kafka partition,
//      ensuring strict First-In, First-Out (FIFO) chronological ordering!
//    - Payload: The complete financial facts (Amount, Currency, TransactionID,
//      MerchantID, IdempotencyKey, and UTC Timestamp).
//
// 3. UNDERSTANDING PRODUCER ACKNOWLEDGMENT LEVELS (RequiredAcks):
//    Kafka offers three distinct durability trade-offs via ACKs:
//    - acks=0 (Fire & Forget): Producer never waits for broker response.
//      Throughput is maximum, but data loss is high if the broker crashes.
//    - acks=1 (Leader ACK - NOVARIS Day 16 Default): Producer blocks until the
//      partition leader broker writes the record to its local pagecache.
//      Provides excellent latency (<5ms) with high durability against network blips.
//    - acks=-1 / acks=all (Quorum ACK): Producer waits until all In-Sync
//      Replicas (ISR) have replicated the record to their local logs.
//      Provides absolute zero-data-loss durability for multi-broker clusters.
//
// 4. BATCHING AND MICRO-BUFFERING:
//    Network system calls (`sendmsg`, `write`) have high overhead. If a producer
//    makes a separate TCP round-trip for every 200-byte message, throughput
//    collapses. `kafka.Writer` groups messages into memory batches:
//    - BatchSize: Maximum records buffered before an immediate TCP flush.
//    - BatchTimeout: Maximum time (e.g., 10ms) to wait for more records
//      before flushing an incomplete batch.
//    This balances high throughput (thousands of msgs/sec) with low tail latency.
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

// Standard domain event types emitted by the NOVARIS platform:
const (
	EventTypeTxInitiated = "novaris.tx.initiated"
	EventTypeTxSettled   = "novaris.tx.settled"
	EventTypeTxFailed    = "novaris.tx.failed"
	EventSchemaVersion1  = "1.0"
)

// Common validation errors for streaming events:
var (
	ErrNilEvent            = errors.New("streaming: event cannot be nil")
	ErrMissingTxID         = errors.New("streaming: transaction_id is required")
	ErrMissingPayerID      = errors.New("streaming: payer_id (partition key) is required")
	ErrMissingPayeeID      = errors.New("streaming: payee_id is required")
	ErrInvalidAmount       = errors.New("streaming: amount_cents must be strictly positive")
	ErrProducerClosed      = errors.New("streaming: producer is closed")
	ErrBrokerUnreachable   = errors.New("streaming: kafka broker unreachable")
)

// TransactionInitiatedEvent models the canonical immutable financial event
// emitted by the API Gateway immediately upon accepting a valid payment request.
type TransactionInitiatedEvent struct {
	// EventID is a globally unique UUIDv4 identifying this specific event record:
	EventID string `json:"event_id"`

	// EventType identifies the domain classification ("novaris.tx.initiated"):
	EventType string `json:"event_type"`

	// Version identifies the event schema contract version ("1.0"):
	Version string `json:"version"`

	// TransactionID is the canonical UUID of the ledger transaction:
	TransactionID string `json:"transaction_id"`

	// PayerID is the sender account UUID, serving as the Kafka Partition Key:
	PayerID string `json:"payer_id"`

	// PayeeID is the recipient account UUID:
	PayeeID string `json:"payee_id"`

	// AmountCents represents the transfer magnitude in integer cents ($50.00 = 5000):
	AmountCents int64 `json:"amount_cents"`

	// Currency is the standard ISO 4217 three-letter currency code (e.g. "USD"):
	Currency string `json:"currency"`

	// MerchantID is the authenticated tenant or client submitting the transaction:
	MerchantID string `json:"merchant_id"`

	// IdempotencyKey is the client-supplied deduplication identifier:
	IdempotencyKey string `json:"idempotency_key,omitempty"`

	// Timestamp is the exact UTC epoch timestamp when the gateway accepted the request:
	Timestamp time.Time `json:"timestamp"`

	// Status reflects the initial transaction lifecycle state ("INITIATED"):
	Status string `json:"status"`

	// Metadata carries optional tracing contexts, correlation IDs, or client user-agents:
	Metadata map[string]string `json:"metadata,omitempty"`
}

// Validate verifies that all mandatory financial invariants are satisfied before publishing.
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
	if e.Status == "" {
		e.Status = "INITIATED"
	}
	if e.Timestamp.IsZero() {
		e.Timestamp = time.Now().UTC()
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

// PublishResult encapsulates telemetry and broker feedback from a successful write.
type PublishResult struct {
	Topic     string        `json:"topic"`
	Key       string        `json:"key"`
	Partition int           `json:"partition"`
	Offset    int64         `json:"offset"`
	Timestamp time.Time     `json:"timestamp"`
	Latency   time.Duration `json:"latency_ns"`
}

// ProducerStats tracks real-time operational metrics for gateway telemetry.
type ProducerStats struct {
	TotalPublished int64         `json:"total_published"`
	TotalFailed    int64         `json:"total_failed"`
	LastLatencyNs  int64         `json:"last_latency_ns"`
	LastPublished  time.Time     `json:"last_published_at"`
}

// EventProducer specifies the contractual interface for publishing domain events.
// Decouples the gateway from raw Kafka primitives, allowing trivial test mocking.
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

// KafkaProducer is the production implementation of EventProducer using kafka-go.
type KafkaProducer struct {
	cfg        *Config
	writer     *kafka.Writer
	closed     uint32
	stats      ProducerStats
	statsMu    sync.RWMutex
}

// NewProducer instantiates and configures a high-performance, resilient Kafka Producer.
func NewProducer(cfg *Config) (*KafkaProducer, error) {
	if cfg == nil {
		c := LoadConfigFromEnv()
		cfg = &c
	}

	// Configure segmentio/kafka-go Writer with enterprise streaming defaults:
	// Note: Topic is intentionally left unset at the Writer level so individual
	// messages can dynamically target different topics (e.g., tx.initiated, tx.settled, tx.failed).
	writer := &kafka.Writer{
		Addr:         kafka.TCP(cfg.Brokers...),
		Balancer:     &kafka.Murmur2Balancer{}, // Guarantees compatibility with Apache Kafka Java client
		RequiredAcks: kafka.RequireOne,        // Leader ACK confirmation for durability
		MaxAttempts:  3,                       // Up to 3 automatic retries on transient network errors
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
func (p *KafkaProducer) Publish(ctx context.Context, topic string, key string, payload []byte) (*PublishResult, error) {
	if atomic.LoadUint32(&p.closed) == 1 {
		return nil, ErrProducerClosed
	}

	start := time.Now()

	// Construct the Kafka message record:
	msg := kafka.Message{
		Topic: topic,
		Key:   []byte(key),
		Value: payload,
		Time:  time.Now().UTC(),
		Headers: []kafka.Header{
			{Key: "client_id", Value: []byte(p.cfg.ClientID)},
			{Key: "version", Value: []byte(EventSchemaVersion1)},
		},
	}

	// Write message to the Kafka broker cluster:
	err := p.writer.WriteMessages(ctx, msg)
	latency := time.Since(start)

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

	// Note: In segmentio/kafka-go synchronous mode, the writer returns once the message
	// is acknowledged by the broker leader.
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
	if err := event.Validate(); err != nil {
		return nil, fmt.Errorf("streaming: invalid transaction event: %w", err)
	}

	payload, err := event.ToJSON()
	if err != nil {
		return nil, fmt.Errorf("streaming: serialization error: %w", err)
	}

	// PayerID is the partition key: all events for the same payer land in the same partition!
	return p.Publish(ctx, TopicTxInitiated, event.PayerID, payload)
}

// Close gracefully drains in-flight message batches and closes all open TCP sockets.
func (p *KafkaProducer) Close() error {
	if !atomic.CompareAndSwapUint32(&p.closed, 0, 1) {
		return nil // Already closed
	}
	return p.writer.Close()
}

// Stats returns a snapshot of runtime publication metrics.
func (p *KafkaProducer) Stats() ProducerStats {
	p.statsMu.RLock()
	defer p.statsMu.RUnlock()
	return p.stats
}

// ==============================================================================
// IN-MEMORY MOCK PRODUCER (FOR ISOLATED UNIT TESTING & DECOUPLED MODE)
// ==============================================================================

// MockProducer provides an in-memory thread-safe implementation of EventProducer.
// Allows unit tests to verify event emission without requiring a running Kafka broker.
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

// PublishTxInitiated records the event in memory.
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

// Publish simulates writing to Kafka by appending to in-memory slices.
func (m *MockProducer) Publish(ctx context.Context, topic string, key string, payload []byte) (*PublishResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return nil, ErrProducerClosed
	}

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

	// If payload is a TransactionInitiatedEvent, decode and record it:
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

// Stats returns mock metrics.
func (m *MockProducer) Stats() ProducerStats {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.stats
}

// GetEvents returns a thread-safe copy of all published events.
func (m *MockProducer) GetEvents() []*TransactionInitiatedEvent {
	m.mu.RLock()
	defer m.mu.RUnlock()
	copied := make([]*TransactionInitiatedEvent, len(m.Events))
	copy(copied, m.Events)
	return copied
}

// Clear resets recorded events.
func (m *MockProducer) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Events = m.Events[:0]
	m.RawMessages = m.RawMessages[:0]
}
