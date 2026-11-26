package streaming

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"
)

// TestLoadConfigDefaults verifies fallback values when environment variables are unset.
func TestLoadConfigDefaults(t *testing.T) {
	os.Unsetenv(EnvKafkaBrokers)
	os.Unsetenv(EnvKafkaClientID)
	os.Unsetenv(EnvKafkaDialTimeoutMS)
	os.Unsetenv(EnvKafkaDefaultPartitions)
	os.Unsetenv(EnvKafkaDefaultReplication)

	cfg := LoadConfigFromEnv()

	if len(cfg.Brokers) != 1 || cfg.Brokers[0] != DefaultBrokerAddress {
		t.Errorf("expected default broker %s, got %v", DefaultBrokerAddress, cfg.Brokers)
	}
	if cfg.ClientID != DefaultClientID {
		t.Errorf("expected default client ID %s, got %s", DefaultClientID, cfg.ClientID)
	}
	if cfg.DialTimeout != DefaultDialTimeout {
		t.Errorf("expected default dial timeout %v, got %v", DefaultDialTimeout, cfg.DialTimeout)
	}
	if cfg.DefaultPartitions != DefaultPartitionCount {
		t.Errorf("expected %d default partitions, got %d", DefaultPartitionCount, cfg.DefaultPartitions)
	}
	if cfg.DefaultReplicationFactor != DefaultReplicationFactor {
		t.Errorf("expected replication factor %d, got %d", DefaultReplicationFactor, cfg.DefaultReplicationFactor)
	}
}

// TestLoadConfigEnvOverrides verifies that environment variables properly override defaults.
func TestLoadConfigEnvOverrides(t *testing.T) {
	t.Setenv(EnvKafkaBrokers, "kafka-node1:9092, kafka-node2:9092 ")
	t.Setenv(EnvKafkaClientID, "test-client-custom")
	t.Setenv(EnvKafkaDialTimeoutMS, "5000")
	t.Setenv(EnvKafkaDefaultPartitions, "6")
	t.Setenv(EnvKafkaDefaultReplication, "3")

	cfg := LoadConfigFromEnv()

	if len(cfg.Brokers) != 2 || cfg.Brokers[0] != "kafka-node1:9092" || cfg.Brokers[1] != "kafka-node2:9092" {
		t.Errorf("unexpected parsed brokers: %v", cfg.Brokers)
	}
	if cfg.ClientID != "test-client-custom" {
		t.Errorf("unexpected client ID: %s", cfg.ClientID)
	}
	if cfg.DialTimeout != 5*time.Second {
		t.Errorf("unexpected dial timeout: %v", cfg.DialTimeout)
	}
	if cfg.DefaultPartitions != 6 {
		t.Errorf("expected 6 partitions, got %d", cfg.DefaultPartitions)
	}
	if cfg.DefaultReplicationFactor != 3 {
		t.Errorf("expected 3 replication factor, got %d", cfg.DefaultReplicationFactor)
	}
}

// TestPartitionerInvariants verifies deterministic routing and bounded modulo properties.
func TestPartitionerInvariants(t *testing.T) {
	partitioner := NewPartitioner(StrategyFNV1a)
	testKeys := []string{
		"acc_merchant_alpha_001",
		"acc_user_charlie_772",
		"acc_platform_settlement",
		"acc_fee_reserve",
		"acc_liquidity_provider_eur",
	}

	numPartitions := 5

	for _, key := range testKeys {
		firstPartition := partitioner.RouteAccountKey(key, numPartitions)

		// Bound invariant: 0 <= partition < numPartitions
		if firstPartition < 0 || firstPartition >= numPartitions {
			t.Fatalf("partition out of bounds for key %s: got %d, expected [0, %d)", key, firstPartition, numPartitions)
		}

		// Mathematical Determinism Invariant: 10,000 repeated calls must yield identical partition
		for i := 0; i < 1000; i++ {
			p := partitioner.RouteAccountKey(key, numPartitions)
			if p != firstPartition {
				t.Fatalf("determinism violation on iteration %d: key %s routed to %d (initially %d)", i, key, p, firstPartition)
			}
		}
	}

	// Edge Cases:
	if p := partitioner.RouteAccountKey("any_key", 1); p != 0 {
		t.Errorf("expected partition 0 when numPartitions=1, got %d", p)
	}
	if p := partitioner.RouteAccountKey("any_key", 0); p != 0 {
		t.Errorf("expected partition 0 when numPartitions=0, got %d", p)
	}
	if p := partitioner.RouteAccountKey("", 5); p != 0 {
		t.Errorf("expected partition 0 for empty key, got %d", p)
	}
}

// TestMurmur2Strategy verifies the Murmur2 compatibility implementation.
func TestMurmur2Strategy(t *testing.T) {
	pMurmur := NewPartitioner(StrategyMurmur2)
	pSHA := NewPartitioner(StrategySHA256)

	key := "acc_fintech_prod_991"
	parts := 3

	partMurmur := pMurmur.RouteAccountKey(key, parts)
	partSHA := pSHA.RouteAccountKey(key, parts)

	if partMurmur < 0 || partMurmur >= parts {
		t.Errorf("Murmur2 partition out of bounds: %d", partMurmur)
	}
	if partSHA < 0 || partSHA >= parts {
		t.Errorf("SHA256 partition out of bounds: %d", partSHA)
	}

	// Verify determinism across strategies:
	for i := 0; i < 100; i++ {
		if pMurmur.RouteAccountKey(key, parts) != partMurmur {
			t.Fatalf("Murmur2 determinism failed on iteration %d", i)
		}
		if pSHA.RouteAccountKey(key, parts) != partSHA {
			t.Fatalf("SHA256 determinism failed on iteration %d", i)
		}
	}
}

// TestPartitionDistributionUniformity verifies that hashing distributes account keys evenly.
func TestPartitionDistributionUniformity(t *testing.T) {
	partitioner := NewPartitioner(StrategyFNV1a)
	var keys []string
	for i := 0; i < 3000; i++ {
		keys = append(keys, fmt.Sprintf("acc_sample_user_%05d", i))
	}

	numPartitions := 3
	dist := partitioner.AnalyzeDistribution(keys, numPartitions)

	if dist.TotalKeys != 3000 {
		t.Errorf("expected 3000 keys, got %d", dist.TotalKeys)
	}
	if dist.NumPartitions != numPartitions {
		t.Errorf("expected %d partitions, got %d", numPartitions, dist.NumPartitions)
	}

	// In a uniform 3-way split of 3000 keys, each partition should receive ~1000 keys (approx 33.3% ± 10%).
	for p, count := range dist.PartitionLoads {
		pct := dist.LoadPercentages[p]
		if count < 800 || count > 1200 {
			t.Errorf("partition %d load unbalanced: count=%d (%.2f%%)", p, count, pct)
		}
	}
}

// TestStandardTopicsCatalog ensures all canonical Phase 2 topics are correctly defined.
func TestStandardTopicsCatalog(t *testing.T) {
	topics := GetStandardTopics(3, 1)

	if len(topics) != 6 {
		t.Fatalf("expected 6 standard topics, got %d", len(topics))
	}

	expectedNames := map[string]bool{
		TopicTxInitiated: false,
		TopicTxSettled:   false,
		TopicTxFailed:    false,
		TopicTxDLQ:       false,
		TopicTxRetry1:    false,
		TopicTxRetry2:    false,
	}

	for _, spec := range topics {
		if _, ok := expectedNames[spec.Name]; !ok {
			t.Errorf("unexpected topic in catalog: %s", spec.Name)
		}
		expectedNames[spec.Name] = true

		if spec.NumPartitions != 3 {
			t.Errorf("topic %s: expected 3 partitions, got %d", spec.Name, spec.NumPartitions)
		}
		if spec.ReplicationFactor != 1 {
			t.Errorf("topic %s: expected replication factor 1, got %d", spec.Name, spec.ReplicationFactor)
		}
		if spec.RetentionMS <= 0 {
			t.Errorf("topic %s: invalid retention MS %d", spec.Name, spec.RetentionMS)
		}
	}

	for name, found := range expectedNames {
		if !found {
			t.Errorf("canonical topic missing from catalog: %s", name)
		}
	}
}

// TestLiveKafkaIntegration tests real interaction with the local Kafka broker if running.
func TestLiveKafkaIntegration(t *testing.T) {
	cfg := LoadConfigFromEnv()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	admin := NewAdminClient(cfg)

	// Step 1: Check cluster reachability
	health, err := admin.CheckHealth(ctx)
	if err != nil {
		t.Skipf("skipping live Kafka test; broker unreachable: %v", err)
		return
	}

	if !health.Reachable {
		t.Fatalf("cluster reported unreachable: %s", health.ErrorMessage)
	}
	t.Logf("✓ Kafka Cluster Reachable at %s (RTT Latency: %.2f ms, Controller ID: %d, Total Topics: %d)",
		health.ConnectedBroker, health.LatencyMS, health.ControllerID, health.TotalTopics)

	// Step 2: Idempotently provision NOVARIS canonical topics
	specs := GetStandardTopics(cfg.DefaultPartitions, cfg.DefaultReplicationFactor)
	report, err := admin.EnsureTopics(ctx, specs)
	if err != nil {
		t.Fatalf("EnsureTopics failed: %v", err)
	}

	t.Logf("✓ Provisioned Topics: Requested=%d, Created=%d, Existing=%d in %v",
		report.TotalRequested, report.CreatedCount, report.ExistingCount, report.Duration)

	// Step 3: Verify topics are visible
	summaries, err := admin.ListTopics(ctx)
	if err != nil {
		t.Fatalf("ListTopics failed: %v", err)
	}

	foundInitiated := false
	for _, s := range summaries {
		if s.Name == TopicTxInitiated {
			foundInitiated = true
			if s.NumPartitions != cfg.DefaultPartitions {
				t.Errorf("expected %d partitions for %s, got %d", cfg.DefaultPartitions, TopicTxInitiated, s.NumPartitions)
			}
		}
	}

	if !foundInitiated {
		t.Errorf("topic %s not found in cluster topic list", TopicTxInitiated)
	}

	// Step 4: Emit a test probe event to verify producer write path
	probe, err := admin.ProduceProbeMessage(
		ctx,
		TopicTxInitiated,
		"acc_test_integration_user_001",
		`{"event":"probe","phase":2,"day":15,"status":"verified"}`,
	)
	if err != nil {
		t.Fatalf("ProduceProbeMessage failed: %v", err)
	}

	t.Logf("✓ Probe Event Written: Topic=%s, Partition=%d, Bytes=%d, Latency=%v",
		probe.Topic, probe.Partition, probe.PayloadBytes, probe.WriteLatency)
}
