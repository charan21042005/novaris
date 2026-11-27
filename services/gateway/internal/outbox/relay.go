// ==============================================================================
// NOVARIS: Distributed Payment Settlement & Adaptive Fraud Intelligence Platform
// MODULE 1: API Gateway Service - Asynchronous Outbox Relay Daemon (Day 18)
// FILE: services/gateway/internal/outbox/relay.go
// ==============================================================================
//
// Architectural Principles & Invariants:
// 1. Decoupled Asynchronous Polling Relay:
//    The OutboxRelay is an autonomous background engine that bridges PostgreSQL
//    relational transactions and Apache Kafka event streaming. It decouples perimeter
//    HTTP response times from broker latency and provides guaranteed At-Least-Once
//    event propagation.
//
// 2. Crash Safety & Recovery Semantics:
//    - The relay fetches a batch of PENDING records from the outbox table.
//    - For each record, it publishes the payload to the Kafka cluster using the
//      aggregate_id (e.g. Payer Account UUID) as the partition key.
//    - ONLY upon receiving broker delivery confirmation (ACK), the relay updates
//      the record's status to PUBLISHED in PostgreSQL.
//    - If the relay worker crashes mid-batch, unacknowledged records remain PENDING
//      and are safely re-published upon reboot.
//
// 3. Effectively-Once Invariant:
//    Combined with the downstream Settlement Consumer's idempotency deduplication
//    (Day 17), network re-deliveries resulting from relay restarts never cause
//    duplicate financial settlement or balance double-debiting.
// ==============================================================================

package outbox

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

// EventPublisher abstracts the message broker destination for the outbox relay.
// This interface allows seamless integration with segmentio/kafka-go or in-memory mocks.
type EventPublisher interface {
	// PublishEvent sends a serialized event payload to a Kafka topic partitioned by key.
	PublishEvent(ctx context.Context, topic string, key string, payload []byte) error
}

// RelayConfig configures the polling and publishing parameters of the OutboxRelay.
type RelayConfig struct {
	// PollInterval is the idle duration between polling sweeps when no records are found.
	PollInterval time.Duration

	// BatchSize is the maximum number of pending records fetched per sweep.
	BatchSize int

	// MaxRetries is the threshold before a persistently failing record is marked FAILED.
	MaxRetries int

	// DefaultTopic is the fallback topic name if a record does not specify event_type.
	DefaultTopic string
}

// DefaultRelayConfig provides robust defaults for production deployments.
func DefaultRelayConfig() RelayConfig {
	return RelayConfig{
		PollInterval: 100 * time.Millisecond,
		BatchSize:    50,
		MaxRetries:   5,
		DefaultTopic: "novaris.tx.initiated",
	}
}

// RelayStats tracks real-time operational telemetry for the outbox relay.
type RelayStats struct {
	TotalPolled     int64     `json:"total_polled"`
	TotalPublished  int64     `json:"total_published"`
	TotalFailed     int64     `json:"total_failed"`
	BatchesRun      int64     `json:"batches_run"`
	LastPollTime    time.Time `json:"last_poll_time"`
	IsRunning       bool      `json:"is_running"`
}

// CrashHookFunc is a callback invoked during testing to simulate worker crashes.
type CrashHookFunc func(rec *OutboxRecord) error

// OutboxRelay is the active background poller daemon.
type OutboxRelay struct {
	repo      Repository
	publisher EventPublisher
	config    RelayConfig

	// Atomic telemetry counters
	totalPolled    atomic.Int64
	totalPublished atomic.Int64
	totalFailed    atomic.Int64
	batchesRun     atomic.Int64

	mu           sync.RWMutex
	lastPollTime time.Time
	running      bool
	stopChan     chan struct{}
	doneChan     chan struct{}

	// Testing / Fault-Injection hooks
	postPublishCrashHook CrashHookFunc
}

// NewOutboxRelay constructs and initializes a new OutboxRelay daemon.
func NewOutboxRelay(repo Repository, publisher EventPublisher, cfg RelayConfig) (*OutboxRelay, error) {
	if repo == nil {
		return nil, errors.New("outbox relay: repository cannot be nil")
	}
	if publisher == nil {
		return nil, errors.New("outbox relay: publisher cannot be nil")
	}

	if cfg.PollInterval <= 0 {
		cfg.PollInterval = 100 * time.Millisecond
	}
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 50
	}
	if cfg.MaxRetries <= 0 {
		cfg.MaxRetries = 5
	}
	if cfg.DefaultTopic == "" {
		cfg.DefaultTopic = "novaris.tx.initiated"
	}

	return &OutboxRelay{
		repo:      repo,
		publisher: publisher,
		config:    cfg,
		stopChan:  make(chan struct{}),
		doneChan:  make(chan struct{}),
	}, nil
}

// SetPostPublishCrashHook injects a fault hook that fires immediately after a record
// is successfully published to Kafka, but BEFORE the database status is updated to PUBLISHED.
// This is used in Lab L3 to verify zero-loss recovery across simulated crashes!
func (r *OutboxRelay) SetPostPublishCrashHook(hook CrashHookFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.postPublishCrashHook = hook
}

// Start begins the asynchronous polling worker loop in a dedicated goroutine.
func (r *OutboxRelay) Start(ctx context.Context) {
	r.mu.Lock()
	if r.running {
		r.mu.Unlock()
		return
	}
	r.running = true
	r.stopChan = make(chan struct{})
	r.doneChan = make(chan struct{})
	r.mu.Unlock()

	go r.run(ctx)
}

// Stop gracefully shuts down the outbox relay, waiting for the in-flight batch to drain.
func (r *OutboxRelay) Stop() {
	r.mu.Lock()
	if !r.running {
		r.mu.Unlock()
		return
	}
	r.running = false
	close(r.stopChan)
	r.mu.Unlock()

	<-r.doneChan
}

// Stats returns a thread-safe snapshot of the relay's current telemetry counters.
func (r *OutboxRelay) Stats() RelayStats {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return RelayStats{
		TotalPolled:    r.totalPolled.Load(),
		TotalPublished: r.totalPublished.Load(),
		TotalFailed:    r.totalFailed.Load(),
		BatchesRun:     r.batchesRun.Load(),
		LastPollTime:   r.lastPollTime,
		IsRunning:      r.running,
	}
}

// run is the internal polling loop executed by the background goroutine.
func (r *OutboxRelay) run(ctx context.Context) {
	defer close(r.doneChan)

	ticker := time.NewTicker(r.config.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Printf("[OUTBOX_RELAY] Shutdown signaled via context. Exiting relay worker.")
			return
		case <-r.stopChan:
			log.Printf("[OUTBOX_RELAY] Stop requested. Exiting relay worker.")
			return
		case <-ticker.C:
			// Process a single batch sweep
			count, err := r.ProcessBatch(ctx)
			if err != nil {
				log.Printf("[OUTBOX_RELAY_WARN] Batch sweep encountered error: %v", err)
			}
			// If we processed a full batch, poll again immediately without waiting for ticker
			if count >= r.config.BatchSize {
				for count >= r.config.BatchSize {
					select {
					case <-ctx.Done():
						return
					case <-r.stopChan:
						return
					default:
						count, _ = r.ProcessBatch(ctx)
					}
				}
			}
		}
	}
}

// ProcessBatch polls a single batch of pending records, publishes them to Kafka,
// and marks them as PUBLISHED in the outbox table. Returns the number of records processed.
func (r *OutboxRelay) ProcessBatch(ctx context.Context) (int, error) {
	r.mu.Lock()
	r.lastPollTime = time.Now().UTC()
	r.mu.Unlock()
	r.batchesRun.Add(1)

	records, err := r.repo.FetchPending(ctx, r.config.BatchSize)
	if err != nil {
		return 0, fmt.Errorf("fetching pending outbox records: %w", err)
	}

	if len(records) == 0 {
		return 0, nil
	}

	r.totalPolled.Add(int64(len(records)))
	processedCount := 0

	for _, rec := range records {
		select {
		case <-ctx.Done():
			return processedCount, ctx.Err()
		default:
		}

		targetTopic := rec.EventType
		if targetTopic == "" {
			targetTopic = r.config.DefaultTopic
		}

		// Partition key is aggregate_id (e.g. Payer Account UUID) guaranteeing strict FIFO!
		partitionKey := rec.AggregateID

		// 1. Publish to Apache Kafka
		pubErr := r.publisher.PublishEvent(ctx, targetTopic, partitionKey, rec.Payload)
		if pubErr != nil {
			r.totalFailed.Add(1)
			newRetry := rec.RetryCount + 1
			log.Printf("[OUTBOX_RELAY_WARN] Failed publishing outbox record %s (tx: %s): %v. Retry count: %d",
				rec.ID, rec.AggregateID, pubErr, newRetry)

			_ = r.repo.MarkFailed(ctx, rec.ID, newRetry)
			continue
		}

		// 2. Simulated Crash Hook (Fault Injection for Lab L3)
		r.mu.RLock()
		crashHook := r.postPublishCrashHook
		r.mu.RUnlock()
		if crashHook != nil {
			if hookErr := crashHook(rec); hookErr != nil {
				log.Printf("[OUTBOX_FAULT_INJECTION] ⚡ Simulated worker crash triggered after publish for record %s: %v",
					rec.ID, hookErr)
				// Abort batch processing immediately to simulate crash before MarkPublished commits!
				return processedCount, hookErr
			}
		}

		// 3. Mark record as PUBLISHED in PostgreSQL
		if markErr := r.repo.MarkPublished(ctx, rec.ID); markErr != nil {
			log.Printf("[OUTBOX_RELAY_ERROR] Published event %s to Kafka, but failed marking DB status: %v",
				rec.ID, markErr)
			// Notice: Record remains PENDING! Upon restart, it will be re-sent (At-Least-Once).
		} else {
			r.totalPublished.Add(1)
			processedCount++
		}
	}

	return processedCount, nil
}

// ------------------------------------------------------------------------------
// In-Memory Mock Event Publisher Implementation (For Testing & Crash Verification)
// ------------------------------------------------------------------------------

// PublishedMessage records an event emitted through EventPublisher for test assertions.
type PublishedMessage struct {
	Topic     string
	Key       string
	Payload   []byte
	Timestamp time.Time
}

// MockPublisher is a thread-safe implementation of EventPublisher for testing.
type MockPublisher struct {
	mu           sync.RWMutex
	messages     []PublishedMessage
	PublishError error
}

// NewMockPublisher constructs a clean MockPublisher.
func NewMockPublisher() *MockPublisher {
	return &MockPublisher{
		messages: make([]PublishedMessage, 0),
	}
}

// PublishEvent appends the published message to the in-memory log.
func (p *MockPublisher) PublishEvent(ctx context.Context, topic string, key string, payload []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.PublishError != nil {
		return p.PublishError
	}

	payloadClone := append([]byte(nil), payload...)
	p.messages = append(p.messages, PublishedMessage{
		Topic:     topic,
		Key:       key,
		Payload:   payloadClone,
		Timestamp: time.Now().UTC(),
	})

	return nil
}

// GetMessages returns a snapshot copy of all published messages.
func (p *MockPublisher) GetMessages() []PublishedMessage {
	p.mu.RLock()
	defer p.mu.RUnlock()

	result := make([]PublishedMessage, len(p.messages))
	copy(result, p.messages)
	return result
}

// MessageCount returns the total number of published messages.
func (p *MockPublisher) MessageCount() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.messages)
}

// Reset clears recorded messages.
func (p *MockPublisher) Reset() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.messages = make([]PublishedMessage, 0)
	p.PublishError = nil
}
