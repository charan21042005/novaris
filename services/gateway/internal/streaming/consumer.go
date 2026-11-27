// ==============================================================================
// NOVARIS: Distributed Payment Settlement & Adaptive Fraud Intelligence Platform
// MODULE 2: STREAMING CORE & EVENT DRIVEN INFRASTRUCTURE (Phase 2, Day 17)
// FILE: services/gateway/internal/streaming/consumer.go
// ==============================================================================
//
// 📚 COMPUTER SCIENCE & DISTRIBUTED SYSTEMS MASTERCLASS FOR BEGINNERS:
//
// 1. WHAT IS A KAFKA CONSUMER & CONSUMER GROUP?
//    In Day 16, our API Gateway learned how to PUBLISH events ("Producers").
//    Today in Day 17, we build the other half of the streaming equation: "Consumers".
//    A Kafka Consumer is a client application that subscribes to one or more Kafka
//    topics and reads incoming messages from the broker's append-only commit logs.
//
//    A "Consumer Group" is one of the most powerful inventions in distributed systems:
//    - Suppose your topic `novaris.tx.initiated` has 3 partitions and traffic surges
//      to 10,000 payments/second. A single server might not have enough CPU to process
//      all those transactions.
//    - With Kafka, you can launch 3 separate instances of your settlement service,
//      all configured with the SAME `GroupID` (e.g. "novaris-settlement-group").
//    - Kafka's Group Coordinator will automatically assign Partition 0 to Consumer 1,
//      Partition 1 to Consumer 2, and Partition 2 to Consumer 3!
//    - Work is perfectly load-balanced across your cluster without any manual sharding!
//
// 2. OFFSETS AND CHECKPOINTING (HOW KAFKA REMEMBERS):
//    Unlike traditional message queues (like RabbitMQ or SQS) where messages are
//    deleted from the queue as soon as a worker reads them, Kafka NEVER deletes
//    a message when it is read.
//    Instead, messages remain on disk, and Kafka assigns each record an integer index
//    called an "Offset" (e.g. Offset 0, Offset 1, Offset 2...).
//    The consumer periodically writes back an "Offset Commit" to Kafka's internal
//    `__consumer_offsets` topic, recording:
//      "Group 'novaris-settlement-group' has processed up to Offset 42 on Partition 1."
//    If Consumer 2 crashes and reboots 5 minutes later, it doesn't start over from 0!
//    It queries Kafka for its last committed offset and resumes seamlessly at Offset 43.
//
// 3. DELIVERY GUARANTEES & THE COMMIT TIMING DILEMMA:
//    When should a consumer commit its offset? The timing determines your system's safety:
//    - AT-MOST-ONCE (Auto-commit before processing):
//      If you commit the offset before you finish executing the database transaction,
//      and your server crashes mid-flight, the message is permanently skipped and lost!
//      Money disappears into thin air. (Completely unacceptable for FinTech!).
//    - AT-LEAST-ONCE (Process first, then commit — NOVARIS Production Standard):
//      The consumer reads the event, finishes inserting the ledger journal entry,
//      and ONLY THEN calls `r.CommitMessages()`.
//      If the consumer crashes right before the commit, the replacement worker will
//      re-read the message and process it again. Because NOVARIS uses Idempotency Keys,
//      the second execution is recognized as a duplicate and safely ignored.
//      Combining At-Least-Once delivery with Idempotent execution achieves the holy grail
//      of distributed systems: EFFECTIVELY-ONCE PROCESSING!
//
// 4. PARTITION REBALANCING & THE GROUP COORDINATOR:
//    What happens if you have 3 consumers running and you add a 4th consumer? Or what if
//    Consumer 2 dies due to an out-of-memory crash?
//    Kafka brokers constantly monitor consumers via background "Heartbeat" pings.
//    If a consumer stops heartbeating (or takes longer than `MaxWait` to fetch records),
//    the broker marks it dead and triggers a "Rebalance".
//    Partitions are dynamically reassigned to the surviving healthy consumers.
//
// 5. CONSUMER LAG:
//    Consumer Lag is the difference between the latest offset written by the producer
//    and the latest offset committed by the consumer:
//      Lag = LatestTopicOffset - CommittedConsumerOffset
//    If Lag is 0, your consumers are running in real-time. If Lag starts growing, your
//    consumers are falling behind and need to be scaled up!
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

	"github.com/segmentio/kafka-go"
)

// ==============================================================================
// SENTINEL ERRORS (DEFENSIVE CONSUMER HANDLING)
// ==============================================================================

var (
	// ErrConsumerClosed is returned when attempting operations on a terminated consumer:
	ErrConsumerClosed = errors.New("streaming: consumer is closed")

	// ErrNilHandler is returned when constructing a consumer with an uninitialized callback:
	ErrNilHandler = errors.New("streaming: message handler cannot be nil")

	// ErrEmptyGroupID is returned when the required consumer group identifier is empty:
	ErrEmptyGroupID = errors.New("streaming: group_id is required for consumer groups")

	// ErrEmptyTopic is returned when the target topic name is missing:
	ErrEmptyTopic = errors.New("streaming: topic is required")
)

// ==============================================================================
// CONSUMER CONFIGURATION STRUCT
// ==============================================================================

// ConsumerConfig defines operational parameters for connecting a consumer group to Kafka.
type ConsumerConfig struct {
	// Brokers is the list of Kafka bootstrap endpoints (e.g. ["localhost:9092"]):
	Brokers []string `json:"brokers"`

	// GroupID is the unique consumer group identifier (e.g. "novaris-settlement-group"):
	// Invariant: Multiple workers with the same GroupID cooperate to balance partitions!
	GroupID string `json:"group_id"`

	// Topic is the canonical Kafka topic to consume (e.g. "novaris.tx.initiated"):
	Topic string `json:"topic"`

	// MinBytes is the minimum batch size to fetch from the broker (default: 1 byte):
	MinBytes int `json:"min_bytes"`

	// MaxBytes is the maximum batch size to fetch from the broker (default: 10MB):
	MaxBytes int `json:"max_bytes"`

	// MaxWait is the maximum time the broker will wait for MinBytes before returning:
	MaxWait time.Duration `json:"max_wait"`

	// StartOffset defines where to begin reading if no prior committed offset exists.
	// Use kafka.FirstOffset (-2) to read from beginning of log, or kafka.LastOffset (-1).
	StartOffset int64 `json:"start_offset"`

	// CommitInterval specifies the periodic commit interval.
	// Set to 0 to enforce strict MANUAL offset commits via CommitMessages!
	CommitInterval time.Duration `json:"commit_interval"`

	// ClientID identifies this specific consumer instance during wire handshakes:
	ClientID string `json:"client_id"`
}

// DefaultConsumerConfig returns enterprise production defaults for NOVARIS consumer workers.
func DefaultConsumerConfig(groupID, topic string) ConsumerConfig {
	return ConsumerConfig{
		Brokers:        []string{"localhost:9092"},
		GroupID:        groupID,
		Topic:          topic,
		MinBytes:       1,                      // Return immediately when at least 1 byte is available
		MaxBytes:       10e6,                   // 10MB batch ceiling
		MaxWait:        500 * time.Millisecond, // 500ms max broker poll latency
		StartOffset:    kafka.FirstOffset,      // Scan from beginning if new consumer group
		CommitInterval: 0,                      // 0 = Manual At-Least-Once commits after processing
		ClientID:       "novaris-consumer-worker",
	}
}

// Validate verifies that all mandatory configuration parameters are present.
func (c *ConsumerConfig) Validate() error {
	if len(c.Brokers) == 0 {
		return errors.New("streaming: at least one broker address is required")
	}
	if c.GroupID == "" {
		return ErrEmptyGroupID
	}
	if c.Topic == "" {
		return ErrEmptyTopic
	}
	if c.MinBytes <= 0 {
		c.MinBytes = 1
	}
	if c.MaxBytes <= 0 {
		c.MaxBytes = 10e6
	}
	if c.MaxWait <= 0 {
		c.MaxWait = 500 * time.Millisecond
	}
	return nil
}

// ==============================================================================
// TELEMETRY & CONSUMER RUNTIME METRICS
// ==============================================================================

// ConsumerStats tracks real-time throughput, lag, and reliability metrics.
type ConsumerStats struct {
	// TotalConsumed is the cumulative number of messages read from Kafka:
	TotalConsumed int64 `json:"total_consumed"`

	// TotalCommitted is the cumulative number of offsets successfully committed:
	TotalCommitted int64 `json:"total_committed"`

	// TotalErrors is the cumulative count of handler or network read errors:
	TotalErrors int64 `json:"total_errors"`

	// LastPartition is the partition index of the most recently processed message:
	LastPartition int `json:"last_partition"`

	// LastOffset is the log offset of the most recently processed message:
	LastOffset int64 `json:"last_offset"`

	// LastConsumedAt is the UTC timestamp of the most recent message:
	LastConsumedAt time.Time `json:"last_consumed_at"`
}

// ==============================================================================
// MESSAGE HANDLER FUNCTION TYPE
// ==============================================================================

// MessageHandler is the callback function invoked for every message read from Kafka.
// If handler returns nil, the consumer automatically commits the message offset.
// If handler returns an error, the message is NOT committed (allowing redelivery).
type MessageHandler func(ctx context.Context, msg kafka.Message) error

// WrapTransactionInitiatedHandler returns a standard MessageHandler that automatically
// unmarshals canonical JSON payloads into strongly-typed TransactionInitiatedEvent pointers.
// Invariant: If unmarshaling fails, the message is logged as corrupted and rejected.
func WrapTransactionInitiatedHandler(fn TransactionInitiatedHandler) MessageHandler {
	return func(ctx context.Context, msg kafka.Message) error {
		var evt TransactionInitiatedEvent
		if err := json.Unmarshal(msg.Value, &evt); err != nil {
			return fmt.Errorf("streaming: failed to unmarshal TransactionInitiatedEvent at offset %d: %w", msg.Offset, err)
		}
		return fn(ctx, &evt, msg)
	}
}

// TransactionInitiatedHandler is a helper callback type specialized for deserialized events.
type TransactionInitiatedHandler func(ctx context.Context, event *TransactionInitiatedEvent, msg kafka.Message) error

// ==============================================================================
// CONSUMER GROUP ENGINE INTERFACE
// ==============================================================================

// EventConsumer defines the high-level contract for event consumer workers in NOVARIS.
type EventConsumer interface {
	// Start begins reading messages in a background goroutine until context is canceled:
	Start(ctx context.Context) error

	// Close terminates the consumer, flushes in-flight commits, and leaves the group:
	Close() error

	// Stats returns a thread-safe snapshot of real-time metrics:
	Stats() ConsumerStats
}

// ==============================================================================
// PRODUCTION KAFKA CONSUMER GROUP IMPLEMENTATION
// ==============================================================================

// ConsumerGroup wraps segmentio/kafka-go.Reader to provide production-grade
// consumer group coordination, manual offset commit control, and telemetry.
type ConsumerGroup struct {
	cfg     ConsumerConfig
	reader  *kafka.Reader
	handler MessageHandler
	closed  uint32       // Atomic flag: 0 = active, 1 = closed
	stats   ConsumerStats
	statsMu sync.RWMutex
}

// NewConsumerGroup instantiates and configures a resilient Kafka consumer worker.
func NewConsumerGroup(cfg ConsumerConfig, handler MessageHandler) (*ConsumerGroup, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("streaming: invalid consumer config: %w", err)
	}
	if handler == nil {
		return nil, ErrNilHandler
	}

	// Configure segmentio/kafka-go.Reader with consumer group semantics:
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        cfg.Brokers,
		GroupID:        cfg.GroupID, // Non-empty GroupID enables Kafka Consumer Group protocol!
		Topic:          cfg.Topic,
		MinBytes:       cfg.MinBytes,
		MaxBytes:       cfg.MaxBytes,
		MaxWait:        cfg.MaxWait,
		StartOffset:    cfg.StartOffset,
		CommitInterval: cfg.CommitInterval, // 0 = manual commit via CommitMessages
		// Note: kafka-go automatically manages heartbeats, partition joins, and rebalances
	})

	return &ConsumerGroup{
		cfg:     cfg,
		reader:  reader,
		handler: handler,
	}, nil
}

// Start executes the continuous consumer polling loop.
// It blocks until the provided context is canceled or the consumer is closed.
// Implements strict AT-LEAST-ONCE processing:
//   1. FetchMessage (reads record, does NOT commit offset yet)
//   2. handler(ctx, msg) (executes business logic)
//   3. CommitMessages (commits offset ONLY upon handler success)
func (cg *ConsumerGroup) Start(ctx context.Context) error {
	for {
		// 1. Check if closed or context canceled:
		if atomic.LoadUint32(&cg.closed) == 1 {
			return ErrConsumerClosed
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// 2. Fetch the next message from the assigned partition:
		// Note: FetchMessage retrieves the record WITHOUT committing its offset!
		msg, err := cg.reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || atomic.LoadUint32(&cg.closed) == 1 {
				return nil
			}
			cg.recordError()
			return fmt.Errorf("streaming: fetch error on topic %s: %w", cg.cfg.Topic, err)
		}

		// 3. Update telemetry:
		cg.recordMessage(msg.Partition, msg.Offset)

		// 4. Execute the business logic handler:
		if err := cg.handler(ctx, msg); err != nil {
			cg.recordError()
			// Invariant: Do NOT commit offset if handler failed!
			// Return error or continue based on operational retry policy.
			return fmt.Errorf("streaming: handler execution failed for offset %d on partition %d: %w",
				msg.Offset, msg.Partition, err)
		}

		// 5. Commit offset upon successful handler execution (At-Least-Once Delivery):
		if err := cg.reader.CommitMessages(ctx, msg); err != nil {
			cg.recordError()
			return fmt.Errorf("streaming: commit error for offset %d on partition %d: %w",
				msg.Offset, msg.Partition, err)
		}

		cg.recordCommit()
	}
}

// ReadNext processes exactly ONE message synchronously.
// Highly useful for deterministic testing, verification CLIs, and step-by-step audits.
func (cg *ConsumerGroup) ReadNext(ctx context.Context) (*kafka.Message, error) {
	if atomic.LoadUint32(&cg.closed) == 1 {
		return nil, ErrConsumerClosed
	}

	msg, err := cg.reader.FetchMessage(ctx)
	if err != nil {
		cg.recordError()
		return nil, err
	}

	cg.recordMessage(msg.Partition, msg.Offset)

	// Execute handler:
	if err := cg.handler(ctx, msg); err != nil {
		cg.recordError()
		return &msg, fmt.Errorf("streaming: handler failed: %w", err)
	}

	// Commit offset:
	if err := cg.reader.CommitMessages(ctx, msg); err != nil {
		cg.recordError()
		return &msg, fmt.Errorf("streaming: commit failed: %w", err)
	}

	cg.recordCommit()
	return &msg, nil
}

// Close gracefully closes the reader socket and leaves the consumer group.
// Uses atomic CAS to guarantee safe, idempotent teardown.
func (cg *ConsumerGroup) Close() error {
	if !atomic.CompareAndSwapUint32(&cg.closed, 0, 1) {
		return nil // Already closed
	}
	return cg.reader.Close()
}

// Stats returns a thread-safe snapshot of runtime metrics.
func (cg *ConsumerGroup) Stats() ConsumerStats {
	cg.statsMu.RLock()
	defer cg.statsMu.RUnlock()
	return cg.stats
}

func (cg *ConsumerGroup) recordMessage(partition int, offset int64) {
	cg.statsMu.Lock()
	defer cg.statsMu.Unlock()
	cg.stats.TotalConsumed++
	cg.stats.LastPartition = partition
	cg.stats.LastOffset = offset
	cg.stats.LastConsumedAt = time.Now().UTC()
}

func (cg *ConsumerGroup) recordCommit() {
	cg.statsMu.Lock()
	defer cg.statsMu.Unlock()
	cg.stats.TotalCommitted++
}

func (cg *ConsumerGroup) recordError() {
	cg.statsMu.Lock()
	defer cg.statsMu.Unlock()
	cg.stats.TotalErrors++
}

// ==============================================================================
// IN-MEMORY MOCK CONSUMER (FOR ISOLATED UNIT TESTING & OFFLINE CI/CD)
// ==============================================================================

// MockConsumer provides an in-memory, thread-safe implementation of EventConsumer.
// It allows unit tests to simulate receiving Kafka messages, test handler logic,
// and inspect committed offsets without requiring Docker.
type MockConsumer struct {
	mu           sync.RWMutex
	Queue        []kafka.Message
	Handler      MessageHandler
	Processed    []kafka.Message
	Committed    []int64
	closed       bool
	stats        ConsumerStats
	FailOnOffset int64 // Injects artificial failure when offset matches
}

// NewMockConsumer constructs a ready-to-use in-memory test consumer.
func NewMockConsumer(handler MessageHandler) *MockConsumer {
	return &MockConsumer{
		Queue:        make([]kafka.Message, 0),
		Handler:      handler,
		Processed:    make([]kafka.Message, 0),
		Committed:    make([]int64, 0),
		FailOnOffset: -1, // -1 means no error injection
	}
}

// Enqueue appends a message to the simulated broker log for this mock consumer to process.
func (m *MockConsumer) Enqueue(msg kafka.Message) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Queue = append(m.Queue, msg)
}

// EnqueueEvent serializes a TransactionInitiatedEvent and queues it.
func (m *MockConsumer) EnqueueEvent(event *TransactionInitiatedEvent, partition int, offset int64) error {
	payload, err := event.ToJSON()
	if err != nil {
		return err
	}
	msg := kafka.Message{
		Topic:     TopicTxInitiated,
		Partition: partition,
		Offset:    offset,
		Key:       []byte(event.PayerID),
		Value:     payload,
		Time:      time.Now().UTC(),
	}
	m.Enqueue(msg)
	return nil
}

// Start drains all queued messages through the handler, simulating live consumption.
func (m *MockConsumer) Start(ctx context.Context) error {
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return ErrConsumerClosed
	}
	queue := make([]kafka.Message, len(m.Queue))
	copy(queue, m.Queue)
	m.Queue = m.Queue[:0]
	m.mu.Unlock()

	for _, msg := range queue {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		m.mu.Lock()
		m.stats.TotalConsumed++
		m.stats.LastPartition = msg.Partition
		m.stats.LastOffset = msg.Offset
		m.stats.LastConsumedAt = time.Now().UTC()

		if m.FailOnOffset >= 0 && msg.Offset == m.FailOnOffset {
			m.stats.TotalErrors++
			m.mu.Unlock()
			return errors.New("mock: deliberate simulated handler processing error")
		}
		m.mu.Unlock()

		if err := m.Handler(ctx, msg); err != nil {
			m.mu.Lock()
			m.stats.TotalErrors++
			m.mu.Unlock()
			return err
		}

		m.mu.Lock()
		m.Processed = append(m.Processed, msg)
		m.Committed = append(m.Committed, msg.Offset)
		m.stats.TotalCommitted++
		m.mu.Unlock()
	}

	return nil
}

// Close marks the mock consumer as closed.
func (m *MockConsumer) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	return nil
}

// Stats returns thread-safe mock metrics.
func (m *MockConsumer) Stats() ConsumerStats {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.stats
}

// GetCommittedOffsets returns all offsets committed by this mock consumer.
func (m *MockConsumer) GetCommittedOffsets() []int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	copied := make([]int64, len(m.Committed))
	copy(copied, m.Committed)
	return copied
}
