// ==============================================================================
// NOVARIS: Distributed Payment Settlement & Adaptive Fraud Intelligence Platform
// MODULE 2: STREAMING CORE (Phase 2, Day 17: Kafka Consumer Integration Tests)
// FILE: services/gateway/kafka_day17_test.go
// ==============================================================================
//
// 📚 COMPUTER SCIENCE & TEST METHODOLOGY FOR BEGINNERS:
//
// 1. WHAT DOES THIS SUITE VALIDATE?
//    This test suite verifies Day 17 deliverables against the live Apache Kafka
//    cluster running via Docker Compose:
//    - Sub-test 1: Consumer Constructor & Config Validation.
//    - Sub-test 2: Single Consumer Group Message Processing & Offset Commit (L2 / L3).
//    - Sub-test 3: Strict FIFO In-Order Reading Invariant (Per-Account Sequencing).
//    - Sub-test 4: Multi-Consumer Group Fan-Out (Independent Settlement & Fraud Groups).
//    - Sub-test 5: Full Gateway HTTP Ingress -> Kafka Producer -> Consumer Pipeline.
//
// 2. HERMETIC ENVIRONMENT DEFENSE:
//    If the Kafka container is stopped or unreachable, these integration tests
//    cleanly skip with a helpful diagnostic hint rather than failing CI/CD builds.
// ==============================================================================

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"

	"github.com/charan21042005/novaris/services/gateway/internal/auth"
	"github.com/charan21042005/novaris/services/gateway/internal/ratelimit"
	"github.com/charan21042005/novaris/services/gateway/internal/streaming"
)

func TestDay17_KafkaConsumerMasterSuite(t *testing.T) {
	kafkaCfg := streaming.LoadConfigFromEnv()

	// 1. Defensive Guard: Verify broker reachability before attempting live tests:
	if _, err := streaming.QuickTCPProbe(kafkaCfg.Brokers[0], 2*time.Second); err != nil {
		t.Skipf("Skipping Day 17 live Kafka tests: broker %s unreachable (%v)", kafkaCfg.Brokers[0], err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 2. Ensure canonical topics exist on the broker:
	admin := streaming.NewAdminClient(kafkaCfg)
	specs := streaming.GetStandardTopics(kafkaCfg.DefaultPartitions, kafkaCfg.DefaultReplicationFactor)
	_, err := admin.EnsureTopics(ctx, specs)
	if err != nil {
		t.Fatalf("Failed to provision canonical topics: %v", err)
	}

	// --------------------------------------------------------------------------
	// SUB-TEST 1: Consumer Constructor & Config Validation
	// --------------------------------------------------------------------------
	t.Run("SubTest_1_ConsumerConstructorAndConfig", func(t *testing.T) {
		cfg := streaming.DefaultConsumerConfig("test-group-subtest1", streaming.TopicTxInitiated)
		cfg.Brokers = kafkaCfg.Brokers

		handler := func(ctx context.Context, msg kafka.Message) error { return nil }
		consumer, err := streaming.NewConsumerGroup(cfg, handler)
		if err != nil {
			t.Fatalf("Failed to construct ConsumerGroup: %v", err)
		}
		if consumer == nil {
			t.Fatal("Expected non-nil ConsumerGroup")
		}

		// Verify initial metrics:
		stats := consumer.Stats()
		if stats.TotalConsumed != 0 || stats.TotalCommitted != 0 {
			t.Errorf("Expected initial clean stats, got %+v", stats)
		}

		if err := consumer.Close(); err != nil {
			t.Fatalf("Failed to close consumer cleanly: %v", err)
		}
	})

	// --------------------------------------------------------------------------
	// SUB-TEST 2: Single Consumer Group Message Processing & Offset Commit
	// --------------------------------------------------------------------------
	t.Run("SubTest_2_SingleConsumerGroupProcessingAndCommit", func(t *testing.T) {
		producer, err := streaming.NewProducer(&kafkaCfg)
		if err != nil {
			t.Fatalf("Failed to construct producer: %v", err)
		}
		defer producer.Close()

		txID := uuid.New().String()
		eventID := uuid.New().String()
		payerID := "c0000000-0000-0000-0000-000000000017"

		event := &streaming.TransactionInitiatedEvent{
			EventID:        eventID,
			EventType:      streaming.EventTypeTxInitiated,
			Version:        streaming.EventSchemaVersion1,
			TransactionID:  txID,
			PayerID:        payerID,
			PayeeID:        "c0000000-0000-0000-0000-000000000018",
			AmountCents:    17500, // $175.00 USD
			Currency:       "USD",
			MerchantID:     "merchant_day17_subtest2",
			IdempotencyKey: uuid.New().String(),
			Timestamp:      time.Now().UTC(),
			Status:         "INITIATED",
		}

		// 1. Publish event to Kafka:
		pubCtx, pubCancel := context.WithTimeout(context.Background(), 5*time.Second)
		_, err = producer.PublishTxInitiated(pubCtx, event)
		pubCancel()
		if err != nil {
			t.Fatalf("Failed to publish test event: %v", err)
		}

		// 2. Start consumer in dedicated test group:
		uniqueGroupID := fmt.Sprintf("subtest2-group-%s", uuid.New().String()[:8])
		consumerCfg := streaming.DefaultConsumerConfig(uniqueGroupID, streaming.TopicTxInitiated)
		consumerCfg.Brokers = kafkaCfg.Brokers
		consumerCfg.StartOffset = kafka.FirstOffset

		var receivedEvent *streaming.TransactionInitiatedEvent
		var receivedOffset int64

		handler := func(ctx context.Context, msg kafka.Message) error {
			var cand streaming.TransactionInitiatedEvent
			if err := json.Unmarshal(msg.Value, &cand); err == nil {
				if cand.EventID == eventID {
					receivedEvent = &cand
					receivedOffset = msg.Offset
				}
			}
			return nil
		}

		consumer, err := streaming.NewConsumerGroup(consumerCfg, handler)
		if err != nil {
			t.Fatalf("Failed to construct consumer: %v", err)
		}
		defer consumer.Close()

		// 3. Read messages until target event is found or timeout:
		consumeCtx, consumeCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer consumeCancel()

		for receivedEvent == nil {
			select {
			case <-consumeCtx.Done():
				break
			default:
				_, err := consumer.ReadNext(consumeCtx)
				if err != nil {
					time.Sleep(50 * time.Millisecond)
				}
			}
		}

		// 4. Assertions:
		if receivedEvent == nil {
			t.Fatalf("Target event %s was not received by consumer", eventID)
		}
		if receivedEvent.TransactionID != txID {
			t.Errorf("TxID mismatch: expected %s, got %s", txID, receivedEvent.TransactionID)
		}
		if receivedEvent.AmountCents != 17500 {
			t.Errorf("AmountCents mismatch: expected 17500, got %d", receivedEvent.AmountCents)
		}
		if receivedOffset < 0 {
			t.Errorf("Expected valid positive offset, got %d", receivedOffset)
		}

		stats := consumer.Stats()
		if stats.TotalCommitted < 1 {
			t.Errorf("Expected at least 1 committed offset, got %d", stats.TotalCommitted)
		}
	})

	// --------------------------------------------------------------------------
	// SUB-TEST 3: Strict FIFO In-Order Reading Invariant
	// --------------------------------------------------------------------------
	// Proves Lab L3 & Outcome: "confirm the consumer reads every published event, in order."
	t.Run("SubTest_3_StrictFIFOInOrderReading", func(t *testing.T) {
		producer, err := streaming.NewProducer(&kafkaCfg)
		if err != nil {
			t.Fatalf("Failed to construct producer: %v", err)
		}
		defer producer.Close()

		batchID := uuid.New().String()[:8]
		payerID := "c0000000-0000-0000-0000-000000000099" // Same key guarantees same partition!

		// Publish 5 sequenced events:
		for i := 1; i <= 5; i++ {
			evt := &streaming.TransactionInitiatedEvent{
				EventID:       uuid.New().String(),
				EventType:     streaming.EventTypeTxInitiated,
				Version:       streaming.EventSchemaVersion1,
				TransactionID: uuid.New().String(),
				PayerID:       payerID,
				PayeeID:       "c0000000-0000-0000-0000-000000000002",
				AmountCents:   int64(i * 1000),
				Currency:      "USD",
				Timestamp:     time.Now().UTC(),
				Metadata: map[string]string{
					"batch_id": batchID,
					"seq":      fmt.Sprintf("%d", i),
				},
			}
			pubCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			_, err := producer.PublishTxInitiated(pubCtx, evt)
			cancel()
			if err != nil {
				t.Fatalf("Failed to publish seq %d: %v", i, err)
			}
		}

		// Consume events in fresh group:
		uniqueGroup := fmt.Sprintf("subtest3-fifo-%s", batchID)
		consumerCfg := streaming.DefaultConsumerConfig(uniqueGroup, streaming.TopicTxInitiated)
		consumerCfg.Brokers = kafkaCfg.Brokers
		consumerCfg.StartOffset = kafka.FirstOffset

		var receivedSeqs []int

		handler := func(ctx context.Context, msg kafka.Message) error {
			var evt streaming.TransactionInitiatedEvent
			if err := json.Unmarshal(msg.Value, &evt); err == nil {
				if evt.Metadata != nil && evt.Metadata["batch_id"] == batchID {
					var s int
					_, _ = fmt.Sscanf(evt.Metadata["seq"], "%d", &s)
					receivedSeqs = append(receivedSeqs, s)
				}
			}
			return nil
		}

		consumer, err := streaming.NewConsumerGroup(consumerCfg, handler)
		if err != nil {
			t.Fatalf("Failed to construct consumer: %v", err)
		}
		defer consumer.Close()

		consumeCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		for len(receivedSeqs) < 5 {
			select {
			case <-consumeCtx.Done():
				break
			default:
				_, err := consumer.ReadNext(consumeCtx)
				if err != nil {
					time.Sleep(50 * time.Millisecond)
				}
			}
		}

		if len(receivedSeqs) != 5 {
			t.Fatalf("Expected 5 consumed events, got %d (received: %v)", len(receivedSeqs), receivedSeqs)
		}

		// Assert strict monotonic order: 1, 2, 3, 4, 5
		for i := 0; i < 5; i++ {
			expected := i + 1
			if receivedSeqs[i] != expected {
				t.Fatalf("FIFO violation at index %d: expected %d, got %d", i, expected, receivedSeqs[i])
			}
		}
	})

	// --------------------------------------------------------------------------
	// SUB-TEST 4: Multi-Consumer Group Fan-Out
	// --------------------------------------------------------------------------
	// Proves that two distinct consumer groups (Settlement & Fraud ML) independently
	// receive the exact same published event without stealing messages from each other.
	t.Run("SubTest_4_MultiConsumerGroupFanOut", func(t *testing.T) {
		producer, err := streaming.NewProducer(&kafkaCfg)
		if err != nil {
			t.Fatalf("Failed to construct producer: %v", err)
		}
		defer producer.Close()

		uniqueTxID := uuid.New().String()
		event := &streaming.TransactionInitiatedEvent{
			EventID:       uuid.New().String(),
			EventType:     streaming.EventTypeTxInitiated,
			TransactionID: uniqueTxID,
			PayerID:       "acc_fanout_test",
			PayeeID:       "acc_payee",
			AmountCents:   4500,
			Currency:      "USD",
			Timestamp:     time.Now().UTC(),
		}

		pubCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_, err = producer.PublishTxInitiated(pubCtx, event)
		cancel()
		if err != nil {
			t.Fatalf("Failed to publish fanout event: %v", err)
		}

		// Setup Group A (Settlement) and Group B (Fraud ML):
		groupA := fmt.Sprintf("fanout-settlement-%s", uuid.New().String()[:8])
		groupB := fmt.Sprintf("fanout-fraud-%s", uuid.New().String()[:8])

		var groupAFound, groupBFound bool

		handlerA := func(ctx context.Context, msg kafka.Message) error {
			var evt streaming.TransactionInitiatedEvent
			if err := json.Unmarshal(msg.Value, &evt); err == nil && evt.TransactionID == uniqueTxID {
				groupAFound = true
			}
			return nil
		}
		handlerB := func(ctx context.Context, msg kafka.Message) error {
			var evt streaming.TransactionInitiatedEvent
			if err := json.Unmarshal(msg.Value, &evt); err == nil && evt.TransactionID == uniqueTxID {
				groupBFound = true
			}
			return nil
		}

		cfgA := streaming.DefaultConsumerConfig(groupA, streaming.TopicTxInitiated)
		cfgA.Brokers = kafkaCfg.Brokers
		cfgA.StartOffset = kafka.FirstOffset

		cfgB := streaming.DefaultConsumerConfig(groupB, streaming.TopicTxInitiated)
		cfgB.Brokers = kafkaCfg.Brokers
		cfgB.StartOffset = kafka.FirstOffset

		consA, err := streaming.NewConsumerGroup(cfgA, handlerA)
		if err != nil {
			t.Fatalf("Failed to construct consumer A: %v", err)
		}
		defer consA.Close()

		consB, err := streaming.NewConsumerGroup(cfgB, handlerB)
		if err != nil {
			t.Fatalf("Failed to construct consumer B: %v", err)
		}
		defer consB.Close()

		// Read concurrently in both groups:
		cCtx, cCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cCancel()

		for (!groupAFound || !groupBFound) && cCtx.Err() == nil {
			if !groupAFound {
				_, _ = consA.ReadNext(cCtx)
			}
			if !groupBFound {
				_, _ = consB.ReadNext(cCtx)
			}
		}

		if !groupAFound {
			t.Errorf("Consumer Group A (Settlement) failed to receive fanout event")
		}
		if !groupBFound {
			t.Errorf("Consumer Group B (Fraud ML) failed to receive fanout event")
		}
	})

	// --------------------------------------------------------------------------
	// SUB-TEST 5: Full Gateway Ingress -> Kafka Producer -> Settlement Consumer Pipeline
	// --------------------------------------------------------------------------
	t.Run("SubTest_5_GatewayIngressToSettlementConsumerPipeline", func(t *testing.T) {
		producer, err := streaming.NewProducer(&kafkaCfg)
		if err != nil {
			t.Fatalf("Failed to construct producer: %v", err)
		}
		defer producer.Close()
		SetEventProducer(producer)

		// Boot in-memory test server with complete gateway middleware:
		cfg := Config{
			Port:           "8080",
			ServiceName:    "novaris-gateway-test",
			ServiceVersion: "0.17.0",
			Environment:    "test",
		}
		bucketStore := ratelimit.NewBucketStore(10, 2)
		idempStore := NewInMemoryIdempotencyStore()
		handler := SetupServerMux(cfg, nil, bucketStore, idempStore)
		ts := httptest.NewServer(handler)
		defer ts.Close()

		// Mint authentic JWT and compute HMAC signature:
		merchantID := "c0000000-0000-0000-0000-000000000001"
		token, err := auth.GenerateToken(merchantID, "merchant", 60*time.Minute)
		if err != nil {
			t.Fatalf("Failed to generate test token: %v", err)
		}

		reqBody := `{"amount_cents":7800,"recipient":"c0000000-0000-0000-0000-000000000002"}`
		timestamp := time.Now().Unix()
		sig := auth.ComputeHMACSignature([]byte(reqBody), timestamp, auth.GetHMACSecret())
		idempKey := uuid.New().String()

		httpReq, err := http.NewRequest("POST", ts.URL+"/v1/payments", bytes.NewBufferString(reqBody))
		if err != nil {
			t.Fatalf("Failed to construct HTTP request: %v", err)
		}

		httpReq.Header.Set("Authorization", "Bearer "+token)
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("X-Novaris-Timestamp", strconv.FormatInt(timestamp, 10))
		httpReq.Header.Set("X-Novaris-Signature", sig)
		httpReq.Header.Set("Idempotency-Key", idempKey)

		resp, err := http.DefaultClient.Do(httpReq)
		if err != nil {
			t.Fatalf("HTTP request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected HTTP 200 OK, got %d", resp.StatusCode)
		}

		var payResp PaymentResponse
		if err := json.NewDecoder(resp.Body).Decode(&payResp); err != nil {
			t.Fatalf("Failed to decode PaymentResponse: %v", err)
		}

		if !payResp.EventPublished {
			t.Fatalf("Expected EventPublished=true on accepted payment")
		}

		// Now consume from downstream consumer:
		consumerGroup := fmt.Sprintf("e2e-settlement-%s", uuid.New().String()[:8])
		consumerCfg := streaming.DefaultConsumerConfig(consumerGroup, streaming.TopicTxInitiated)
		consumerCfg.Brokers = kafkaCfg.Brokers
		consumerCfg.StartOffset = kafka.FirstOffset

		var consumedEvent *streaming.TransactionInitiatedEvent

		consHandler := func(ctx context.Context, msg kafka.Message) error {
			var evt streaming.TransactionInitiatedEvent
			if err := json.Unmarshal(msg.Value, &evt); err == nil {
				if evt.EventID == payResp.EventID {
					consumedEvent = &evt
				}
			}
			return nil
		}

		consumer, err := streaming.NewConsumerGroup(consumerCfg, consHandler)
		if err != nil {
			t.Fatalf("Failed to construct consumer: %v", err)
		}
		defer consumer.Close()

		cCtx, cCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cCancel()

		for consumedEvent == nil && cCtx.Err() == nil {
			_, _ = consumer.ReadNext(cCtx)
		}

		if consumedEvent == nil {
			t.Fatalf("E2E Pipeline Failure: Settlement consumer did not receive event emitted by Gateway HTTP request!")
		}
		if consumedEvent.AmountCents != 7800 {
			t.Errorf("Amount mismatch: expected 7800, got %d", consumedEvent.AmountCents)
		}
		if consumedEvent.TransactionID != payResp.TransactionID {
			t.Errorf("TransactionID mismatch: expected %s, got %s", payResp.TransactionID, consumedEvent.TransactionID)
		}
	})
}
