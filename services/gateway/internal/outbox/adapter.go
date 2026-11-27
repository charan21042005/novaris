// ==============================================================================
// NOVARIS: Distributed Payment Settlement & Adaptive Fraud Intelligence Platform
// MODULE 1: API Gateway Service - Kafka Producer Adapter for Outbox Relay (Day 18)
// FILE: services/gateway/internal/outbox/adapter.go
// ==============================================================================
//
// Architectural Role:
// This adapter connects the streaming.KafkaProducer (Day 16) to the outbox.EventPublisher
// interface without creating circular dependencies. It enables the OutboxRelay to
// publish events to Kafka partitions using Murmur2 deterministic routing.
// ==============================================================================

package outbox

import (
	"context"
	"errors"

	"github.com/charan21042005/novaris/services/gateway/internal/streaming"
)

// RawKafkaPublisher is any producer capable of emitting a raw key/payload to a topic.
type RawKafkaPublisher interface {
	Publish(ctx context.Context, topic string, key string, payload []byte) (*streaming.PublishResult, error)
}

// KafkaAdapter implements EventPublisher by delegating to a RawKafkaPublisher.
type KafkaAdapter struct {
	producer RawKafkaPublisher
}

// NewKafkaAdapter wraps a live or mock Kafka producer as an EventPublisher.
func NewKafkaAdapter(producer RawKafkaPublisher) (*KafkaAdapter, error) {
	if producer == nil {
		return nil, errors.New("outbox: kafka producer cannot be nil")
	}
	return &KafkaAdapter{producer: producer}, nil
}

// PublishEvent sends the event payload through the adapted Kafka producer.
func (a *KafkaAdapter) PublishEvent(ctx context.Context, topic string, key string, payload []byte) error {
	_, err := a.producer.Publish(ctx, topic, key, payload)
	return err
}
