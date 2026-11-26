package main

import (
	"context"
	"testing"
	"time"

	"github.com/charan21042005/novaris/services/gateway/internal/streaming"
)

// TestDay15KafkaStreamingCore tests the Day 15 deliverables of NOVARIS Phase 2:
//  1. Local Kafka cluster reachability over TCP (localhost:9092).
//  2. Retrieval of KRaft metadata (controller ID, broker count, round-trip latency).
//  3. Idempotent creation and validation of canonical Phase 2 topics (novaris.tx.initiated,
//     novaris.tx.settled, novaris.tx.failed, novaris.tx.dlq, novaris.tx.retry.1, novaris.tx.retry.2).
//  4. Mathematical proof of deterministic account-based partition routing (per-account FIFO).
//  5. Successful production of a test probe event to novaris.tx.initiated with leader ACK.
func TestDay15KafkaStreamingCore(t *testing.T) {
	cfg := streaming.LoadConfigFromEnv()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	admin := streaming.NewAdminClient(cfg)

	// Step 1: Pre-flight check & cluster health
	t.Run("01_ClusterReachabilityAndKRaftMetadata", func(t *testing.T) {
		health, err := admin.CheckHealth(ctx)
		if err != nil {
			t.Fatalf("Kafka cluster unreachable at %v: %v", cfg.Brokers, err)
		}
		if !health.Reachable {
			t.Fatalf("Kafka cluster reported unreachable: %s", health.ErrorMessage)
		}

		t.Logf("✓ Connected to Kafka Broker: %s (RTT: %.2f ms)", health.ConnectedBroker, health.LatencyMS)
		t.Logf("✓ KRaft Controller Node ID: %d (%s)", health.ControllerID, health.ControllerHost)
		t.Logf("✓ Registered Brokers: %d node(s)", len(health.Brokers))

		if health.ControllerID < 0 {
			t.Errorf("expected valid KRaft controller ID, got %d", health.ControllerID)
		}
		if len(health.Brokers) == 0 {
			t.Errorf("expected at least 1 broker in cluster, got 0")
		}
	})

	// Step 2: Idempotent Topic Provisioning
	t.Run("02_EnsureCanonicalPhase2Topics", func(t *testing.T) {
		specs := streaming.GetStandardTopics(cfg.DefaultPartitions, cfg.DefaultReplicationFactor)
		report, err := admin.EnsureTopics(ctx, specs)
		if err != nil {
			t.Fatalf("EnsureTopics failed: %v", err)
		}

		t.Logf("✓ Topic Provisioning Complete in %v", report.Duration)
		t.Logf("  Total Requested: %d | Newly Created: %d | Pre-Existing: %d",
			report.TotalRequested, report.CreatedCount, report.ExistingCount)

		if report.TotalRequested != 6 {
			t.Errorf("expected 6 canonical topics requested, got %d", report.TotalRequested)
		}
	})

	// Step 3: Verify Topic Specifications & Partition Counts
	t.Run("03_VerifyTopicCatalogAndPartitions", func(t *testing.T) {
		summaries, err := admin.ListTopics(ctx)
		if err != nil {
			t.Fatalf("ListTopics failed: %v", err)
		}

		topicMap := make(map[string]streaming.TopicSummary)
		for _, s := range summaries {
			topicMap[s.Name] = s
		}

		canonicalTopics := []string{
			streaming.TopicTxInitiated,
			streaming.TopicTxSettled,
			streaming.TopicTxFailed,
			streaming.TopicTxDLQ,
			streaming.TopicTxRetry1,
			streaming.TopicTxRetry2,
		}

		for _, name := range canonicalTopics {
			summary, found := topicMap[name]
			if !found {
				t.Errorf("canonical topic %s not found in cluster topic list", name)
				continue
			}

			if summary.NumPartitions != cfg.DefaultPartitions {
				t.Errorf("topic %s: expected %d partitions, got %d", name, cfg.DefaultPartitions, summary.NumPartitions)
			}
			t.Logf("✓ Verified Topic: %-26s | %d partitions | Replicas: %d",
				name, summary.NumPartitions, summary.ReplicationFactor)
		}
	})

	// Step 4: Mathematical Partition Routing Invariant
	t.Run("04_DeterministicAccountKeyRouting", func(t *testing.T) {
		partitioner := streaming.NewPartitioner(streaming.StrategyFNV1a)
		accounts := []string{
			"acc_merchant_alpha_001",
			"acc_user_bob_4920",
			"acc_merchant_strata_88",
			"acc_treasury_escrow",
			"acc_fee_collection",
		}

		numPartitions := cfg.DefaultPartitions

		for _, acc := range accounts {
			initialPartition := partitioner.RouteAccountKey(acc, numPartitions)
			if initialPartition < 0 || initialPartition >= numPartitions {
				t.Fatalf("partition out of bounds for key %s: got %d, expected [0, %d)", acc, initialPartition, numPartitions)
			}

			// Invariant: Hash determinism across 5,000 evaluations
			for i := 0; i < 5000; i++ {
				p := partitioner.RouteAccountKey(acc, numPartitions)
				if p != initialPartition {
					t.Fatalf("determinism violation on iteration %d: key %s routed to %d (expected %d)", i, acc, p, initialPartition)
				}
			}

			t.Logf("✓ Invariant Verified: Account '%s' strictly routes to Partition %d", acc, initialPartition)
		}
	})

	// Step 5: Diagnostic Probe Event Write
	t.Run("05_DiagnosticProbeWritePath", func(t *testing.T) {
		probeTarget := streaming.TopicTxInitiated
		probeKey := "acc_merchant_alpha_001"
		probePayload := `{"event_type":"PROBE_INGEST","account_id":"acc_merchant_alpha_001","amount_cents":1000,"currency":"USD","phase":2,"day":15,"status":"VERIFIED"}`

		probe, err := admin.ProduceProbeMessage(ctx, probeTarget, probeKey, probePayload)
		if err != nil {
			t.Fatalf("ProduceProbeMessage failed: %v", err)
		}

		t.Logf("✓ Probe Event Appended to Commit Log:")
		t.Logf("  Topic:        %s", probe.Topic)
		t.Logf("  Partition:    %d", probe.Partition)
		t.Logf("  Key:          %s", probe.Key)
		t.Logf("  Payload Size: %d bytes", probe.PayloadBytes)
		t.Logf("  Write Latency: %v", probe.WriteLatency)

		if probe.PayloadBytes <= 0 {
			t.Errorf("invalid payload bytes %d", probe.PayloadBytes)
		}
	})
}
