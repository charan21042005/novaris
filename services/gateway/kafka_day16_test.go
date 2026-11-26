// ==============================================================================
// NOVARIS: Distributed Payment Settlement & Adaptive Fraud Intelligence Platform
// MODULE 2: STREAMING CORE (Phase 2, Day 16: Kafka Producer Integration Tests)
// FILE: services/gateway/kafka_day16_test.go
// ==============================================================================
//
// 📚 COMPUTER SCIENCE & TEST METHODOLOGY FOR BEGINNERS:
//
// 1. WHAT DOES THIS SUITE VALIDATE?
//    This test suite verifies Day 16 deliverables against the live Apache Kafka
//    cluster running via Docker Compose in KRaft mode:
//    - Sub-test 1: Producer Initialization & Config Validation.
//    - Sub-test 2: Synchronous Event Publication (`novaris.tx.initiated`).
//    - Sub-test 3: Direct Event Log Visibility & Deserialization (L3 / Outcome).
//    - Sub-test 4: Deterministic Per-Account Partition Invariant.
//    - Sub-test 5: Full Gateway HTTP Ingress -> Kafka Producer Integration.
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
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"

	"github.com/charan21042005/novaris/services/gateway/internal/auth"
	"github.com/charan21042005/novaris/services/gateway/internal/ratelimit"
	"github.com/charan21042005/novaris/services/gateway/internal/streaming"
)

func TestDay16_KafkaProducerMasterSuite(t *testing.T) {
	kafkaCfg := streaming.LoadConfigFromEnv()

	// Verify broker reachability before executing live cluster tests:
	if _, err := streaming.QuickTCPProbe(kafkaCfg.Brokers[0], 2*time.Second); err != nil {
		t.Skipf("Skipping Day 16 live Kafka tests: broker %s unreachable (%v)", kafkaCfg.Brokers[0], err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// Ensure canonical topics exist on the broker:
	admin := streaming.NewAdminClient(kafkaCfg)
	specs := streaming.GetStandardTopics(kafkaCfg.DefaultPartitions, kafkaCfg.DefaultReplicationFactor)
	_, err := admin.EnsureTopics(ctx, specs)
	if err != nil {
		t.Fatalf("Failed to provision canonical topics: %v", err)
	}

	// --------------------------------------------------------------------------
	// SUB-TEST 1: Producer Constructor & Clean Lifecycle
	// --------------------------------------------------------------------------
	t.Run("SubTest_1_ProducerLifecycle", func(t *testing.T) {
		producer, err := streaming.NewProducer(&kafkaCfg)
		if err != nil {
			t.Fatalf("Failed to construct KafkaProducer: %v", err)
		}
		if producer == nil {
			t.Fatal("Expected non-nil producer instance")
		}

		// Initial metrics check:
		stats := producer.Stats()
		if stats.TotalPublished != 0 {
			t.Errorf("Expected initial TotalPublished=0, got %d", stats.TotalPublished)
		}

		if err := producer.Close(); err != nil {
			t.Fatalf("Failed to close producer cleanly: %v", err)
		}
	})

	// --------------------------------------------------------------------------
	// SUB-TEST 2: Synchronous Event Publishing with Leader ACK
	// --------------------------------------------------------------------------
	t.Run("SubTest_2_PublishTxInitiated", func(t *testing.T) {
		producer, err := streaming.NewProducer(&kafkaCfg)
		if err != nil {
			t.Fatalf("Failed to construct producer: %v", err)
		}
		defer producer.Close()

		txID := uuid.New().String()
		payerID := "c0000000-0000-0000-0000-000000000001"
		payeeID := "c0000000-0000-0000-0000-000000000002"

		event := &streaming.TransactionInitiatedEvent{
			EventID:        uuid.New().String(),
			EventType:      streaming.EventTypeTxInitiated,
			Version:        streaming.EventSchemaVersion1,
			TransactionID:  txID,
			PayerID:        payerID,
			PayeeID:        payeeID,
			AmountCents:    8500,
			Currency:       "USD",
			MerchantID:     "merchant_test_suite",
			IdempotencyKey: uuid.New().String(),
			Timestamp:      time.Now().UTC(),
			Status:         "INITIATED",
		}

		pubCtx, pubCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer pubCancel()

		res, err := producer.PublishTxInitiated(pubCtx, event)
		if err != nil {
			t.Fatalf("Failed to publish tx.initiated event: %v", err)
		}

		if res.Topic != streaming.TopicTxInitiated {
			t.Errorf("Expected topic %s, got %s", streaming.TopicTxInitiated, res.Topic)
		}
		if res.Key != payerID {
			t.Errorf("Expected partition key %s, got %s", payerID, res.Key)
		}
		if res.Latency <= 0 {
			t.Errorf("Expected positive write latency, got %v", res.Latency)
		}

		stats := producer.Stats()
		if stats.TotalPublished != 1 {
			t.Errorf("Expected TotalPublished=1, got %d", stats.TotalPublished)
		}
	})

	// --------------------------------------------------------------------------
	// SUB-TEST 3: Direct Event Log Visibility (Confirming Event is in Kafka Log)
	// --------------------------------------------------------------------------
	t.Run("SubTest_3_TopicVisibilityAudit", func(t *testing.T) {
		producer, err := streaming.NewProducer(&kafkaCfg)
		if err != nil {
			t.Fatalf("Failed to construct producer: %v", err)
		}
		defer producer.Close()

		txID := uuid.New().String()
		uniqueEventID := uuid.New().String()
		payerID := "c0000000-0000-0000-0000-000000000088"

		event := &streaming.TransactionInitiatedEvent{
			EventID:        uniqueEventID,
			EventType:      streaming.EventTypeTxInitiated,
			Version:        streaming.EventSchemaVersion1,
			TransactionID:  txID,
			PayerID:        payerID,
			PayeeID:        "c0000000-0000-0000-0000-000000000099",
			AmountCents:    9999,
			Currency:       "USD",
			MerchantID:     "merchant_visibility_tester",
			IdempotencyKey: uuid.New().String(),
			Timestamp:      time.Now().UTC(),
			Status:         "INITIATED",
		}

		pubCtx, pubCancel := context.WithTimeout(context.Background(), 5*time.Second)
		_, err = producer.PublishTxInitiated(pubCtx, event)
		pubCancel()
		if err != nil {
			t.Fatalf("Failed to publish event for visibility test: %v", err)
		}

		// Read back concurrently across all partitions to verify durability and visibility:
		type foundResult struct {
			event  *streaming.TransactionInitiatedEvent
			part   int
			offset int64
		}

		resultChan := make(chan foundResult, 3)
		ctxAll, cancelAll := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelAll()

		var wg sync.WaitGroup
		for p := 0; p < 3; p++ {
			wg.Add(1)
			go func(part int) {
				defer wg.Done()
				r := kafka.NewReader(kafka.ReaderConfig{
					Brokers:     kafkaCfg.Brokers,
					Topic:       streaming.TopicTxInitiated,
					Partition:   part,
					MinBytes:    1,
					MaxBytes:    10e6,
					StartOffset: kafka.FirstOffset,
				})
				defer r.Close()

				for {
					select {
					case <-ctxAll.Done():
						return
					default:
						msg, err := r.ReadMessage(ctxAll)
						if err != nil {
							return
						}
						var cand streaming.TransactionInitiatedEvent
						if err := json.Unmarshal(msg.Value, &cand); err == nil {
							if cand.EventID == uniqueEventID {
								select {
								case resultChan <- foundResult{event: &cand, part: part, offset: msg.Offset}:
									cancelAll() // Found target event; cancel peer readers
								default:
								}
								return
							}
						}
					}
				}
			}(p)
		}

		var discoveredEvent *streaming.TransactionInitiatedEvent

		select {
		case res := <-resultChan:
			discoveredEvent = res.event
		case <-time.After(5 * time.Second):
		}
		cancelAll()
		wg.Wait()

		if discoveredEvent == nil {
			t.Fatalf("Milestone Outcome Failure: Event %s was not discovered in Kafka topic %s",
				uniqueEventID, streaming.TopicTxInitiated)
		}

		if discoveredEvent.TransactionID != txID {
			t.Errorf("TransactionID mismatch: expected %s, got %s", txID, discoveredEvent.TransactionID)
		}
		if discoveredEvent.AmountCents != 9999 {
			t.Errorf("AmountCents mismatch: expected 9999, got %d", discoveredEvent.AmountCents)
		}
	})

	// --------------------------------------------------------------------------
	// SUB-TEST 4: Deterministic Per-Account Partition Routing
	// --------------------------------------------------------------------------
	t.Run("SubTest_4_PartitionDeterminism", func(t *testing.T) {
		partitioner := streaming.NewPartitioner(streaming.StrategyMurmur2)
		testAccount := "acc_merchant_alpha_001"

		initialPartition := partitioner.RouteAccountKey(testAccount, 3)

		// Test 100 iterations: account key must consistently route to the identical partition:
		for i := 0; i < 100; i++ {
			part := partitioner.RouteAccountKey(testAccount, 3)
			if part != initialPartition {
				t.Fatalf("Partition determinism violated at iteration %d: expected %d, got %d",
					i, initialPartition, part)
			}
		}
	})

	// --------------------------------------------------------------------------
	// SUB-TEST 5: Full Gateway HTTP Ingress -> Kafka Producer Integration
	// --------------------------------------------------------------------------
	t.Run("SubTest_5_GatewayIngressToKafkaPipeline", func(t *testing.T) {
		// Set live Kafka Producer on Gateway:
		producer, err := streaming.NewProducer(&kafkaCfg)
		if err != nil {
			t.Fatalf("Failed to construct gateway producer: %v", err)
		}
		defer producer.Close()
		SetEventProducer(producer)

		// Boot in-memory test server with complete gateway middleware pipeline:
		cfg := Config{
			Port:           "8080",
			ServiceName:    "novaris-gateway-test",
			ServiceVersion: "0.16.0",
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

		reqBody := `{"amount_cents":5400,"recipient":"c0000000-0000-0000-0000-000000000002"}`
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

		if payResp.Status != "settled" {
			t.Errorf("Expected status 'settled', got %s", payResp.Status)
		}
		if !payResp.EventPublished {
			t.Errorf("Expected EventPublished=true on accepted payment")
		}
		if payResp.EventID == "" {
			t.Errorf("Expected non-empty EventID in response")
		}
	})
}
