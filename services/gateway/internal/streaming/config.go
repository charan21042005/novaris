// Package streaming provides high-performance, fault-tolerant Apache Kafka event streaming
// primitives for the NOVARIS payment settlement and adaptive fraud intelligence platform.
//
// Computer Science & Distributed Systems Architecture (First Principles):
//
// In an industrial FinTech architecture, synchronous request-response chains (e.g., HTTP gateway
// calling fraud ML, calling anti-money laundering, calling ledger settlement) suffer from cascading
// latency degradation and total systemic collapse if any single downstream service experiences a hiccup.
//
// Apache Kafka serves as the decoupled, durable, append-only commit log backbone of NOVARIS:
//  1. Ingestion Decoupling: The Go API Gateway ingests payments, verifies HMAC and tokens in <10ms,
//     persists the event to the database outbox, and hands it off to Kafka.
//  2. Scalable Consumer Isolation: Downstream worker pools (Fraud ML Scoring, Settlement Worker,
//     Reconciliation, Merchant Webhooks) consume events asynchronously at their own native velocities.
//  3. Fault Tolerance & Zero Loss: Even if all downstream workers crash simultaneously, unread events
//     remain safely stored on disk in Kafka partition segments with zero data loss.
package streaming

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

// Environment variable keys for cloud-native configuration overrides:
const (
	EnvKafkaBrokers            = "KAFKA_BROKERS"
	EnvKafkaClientID           = "KAFKA_CLIENT_ID"
	EnvKafkaDialTimeoutMS      = "KAFKA_DIAL_TIMEOUT_MS"
	EnvKafkaDefaultPartitions  = "KAFKA_DEFAULT_PARTITIONS"
	EnvKafkaDefaultReplication = "KAFKA_DEFAULT_REPLICATION"
)

// Default configuration parameters for local development & Docker Compose mesh:
const (
	DefaultBrokerAddress          = "localhost:9092"
	DefaultClientID               = "novaris-gateway-service"
	DefaultDialTimeout            = 10 * time.Second
	DefaultPartitionCount         = 3
	DefaultReplicationFactor      = 1
	DefaultSegmentBytes     int64 = 1024 * 1024 * 1024 // 1 GiB commit log segment rolling limit
	DefaultRetentionMS      int64 = 7 * 24 * 60 * 60 * 1000 // 7 days retention policy in milliseconds
)

// Canonical NOVARIS Topic Names:
// Every topic represents a distinct semantic event stream within the transaction lifecycle.
const (
	// TopicTxInitiated carries raw transactions validated by the API Gateway.
	// Partitioned by AccountID to guarantee strict per-account FIFO ordering.
	TopicTxInitiated = "novaris.tx.initiated"

	// TopicTxSettled carries financial transactions that have successfully passed
	// double-entry balance verification and written to the immutable Postgres ledger.
	TopicTxSettled = "novaris.tx.settled"

	// TopicTxFailed carries transactions rejected due to fraud scoring (>0.85 threshold),
	// insufficient funds, or unrecoverable optimistic concurrency conflicts.
	TopicTxFailed = "novaris.tx.failed"

	// TopicTxDLQ is the Dead Letter Queue isolating poisoned, unparsable, or schema-corrupted
	// payloads to prevent head-of-line blocking across worker consumer groups.
	TopicTxDLQ = "novaris.tx.dlq"

	// TopicTxRetry1 is the Level-1 staged retry topic with short exponential backoff (e.g. 5s).
	TopicTxRetry1 = "novaris.tx.retry.1"

	// TopicTxRetry2 is the Level-2 staged retry topic with longer exponential backoff (e.g. 30s).
	TopicTxRetry2 = "novaris.tx.retry.2"
)

// TopicSpec defines the architectural metadata and durability parameters of a Kafka topic.
type TopicSpec struct {
	Name              string            // Logical topic name (e.g., "novaris.tx.initiated")
	NumPartitions     int               // Number of parallel log partitions (unit of horizontal scale)
	ReplicationFactor int               // Number of broker copies across cluster (durability factor)
	Description       string            // Human-readable architectural purpose
	RetentionMS       int64             // Data retention window in milliseconds
	ConfigEntries     map[string]string // Low-level topic configuration overrides
}

// Config represents the Kafka cluster connection and operational settings.
type Config struct {
	// Brokers is a list of host:port strings for Kafka bootstrap brokers.
	Brokers []string

	// ClientID is a logical identifier sent in every Kafka request header for audit logging and tracing.
	ClientID string

	// DialTimeout is the maximum time allowed to establish a raw TCP connection to a broker.
	DialTimeout time.Duration

	// DefaultPartitions is the default partition count when creating new topics.
	DefaultPartitions int

	// DefaultReplicationFactor is the default replication factor (1 for local, 3 for multi-broker prod).
	DefaultReplicationFactor int
}

// LoadConfigFromEnv builds a Config struct, honoring environment variables with production-safe fallbacks.
func LoadConfigFromEnv() Config {
	cfg := Config{
		Brokers:                  []string{DefaultBrokerAddress},
		ClientID:                 DefaultClientID,
		DialTimeout:              DefaultDialTimeout,
		DefaultPartitions:        DefaultPartitionCount,
		DefaultReplicationFactor: DefaultReplicationFactor,
	}

	if brokersEnv := os.Getenv(EnvKafkaBrokers); strings.TrimSpace(brokersEnv) != "" {
		rawBrokers := strings.Split(brokersEnv, ",")
		var cleaned []string
		for _, b := range rawBrokers {
			if trimmed := strings.TrimSpace(b); trimmed != "" {
				cleaned = append(cleaned, trimmed)
			}
		}
		if len(cleaned) > 0 {
			cfg.Brokers = cleaned
		}
	}

	if clientIDEnv := os.Getenv(EnvKafkaClientID); strings.TrimSpace(clientIDEnv) != "" {
		cfg.ClientID = strings.TrimSpace(clientIDEnv)
	}

	if timeoutEnv := os.Getenv(EnvKafkaDialTimeoutMS); strings.TrimSpace(timeoutEnv) != "" {
		if ms, err := strconv.Atoi(timeoutEnv); err == nil && ms > 0 {
			cfg.DialTimeout = time.Duration(ms) * time.Millisecond
		}
	}

	if partEnv := os.Getenv(EnvKafkaDefaultPartitions); strings.TrimSpace(partEnv) != "" {
		if parts, err := strconv.Atoi(partEnv); err == nil && parts > 0 {
			cfg.DefaultPartitions = parts
		}
	}

	if repEnv := os.Getenv(EnvKafkaDefaultReplication); strings.TrimSpace(repEnv) != "" {
		if rep, err := strconv.Atoi(repEnv); err == nil && rep > 0 {
			cfg.DefaultReplicationFactor = rep
		}
	}

	return cfg
}

// GetStandardTopics returns the authoritative catalog of topics required for NOVARIS Phase 2.
func GetStandardTopics(defaultPartitions, defaultReplication int) []TopicSpec {
	if defaultPartitions <= 0 {
		defaultPartitions = DefaultPartitionCount
	}
	if defaultReplication <= 0 {
		defaultReplication = DefaultReplicationFactor
	}

	return []TopicSpec{
		{
			Name:              TopicTxInitiated,
			NumPartitions:     defaultPartitions,
			ReplicationFactor: defaultReplication,
			Description:       "Ingested payment events emitted by Gateway & Outbox Relay; partitioned by AccountID for per-account FIFO",
			RetentionMS:       DefaultRetentionMS,
			ConfigEntries: map[string]string{
				"cleanup.policy": "delete",
				"retention.ms":   strconv.FormatInt(DefaultRetentionMS, 10),
			},
		},
		{
			Name:              TopicTxSettled,
			NumPartitions:     defaultPartitions,
			ReplicationFactor: defaultReplication,
			Description:       "Committed settlement events emitted by Settlement Worker upon double-entry ledger finalization",
			RetentionMS:       DefaultRetentionMS,
			ConfigEntries: map[string]string{
				"cleanup.policy": "delete",
				"retention.ms":   strconv.FormatInt(DefaultRetentionMS, 10),
			},
		},
		{
			Name:              TopicTxFailed,
			NumPartitions:     defaultPartitions,
			ReplicationFactor: defaultReplication,
			Description:       "Aborted transaction events (fraud block, negative balance, OCC collision exhaust)",
			RetentionMS:       DefaultRetentionMS,
			ConfigEntries: map[string]string{
				"cleanup.policy": "delete",
				"retention.ms":   strconv.FormatInt(DefaultRetentionMS, 10),
			},
		},
		{
			Name:              TopicTxDLQ,
			NumPartitions:     defaultPartitions,
			ReplicationFactor: defaultReplication,
			Description:       "Dead Letter Queue isolating corrupted/poison payloads to prevent consumer crash loops",
			RetentionMS:       DefaultRetentionMS * 2, // 14 days retention for post-mortem forensics
			ConfigEntries: map[string]string{
				"cleanup.policy": "delete",
				"retention.ms":   strconv.FormatInt(DefaultRetentionMS*2, 10),
			},
		},
		{
			Name:              TopicTxRetry1,
			NumPartitions:     defaultPartitions,
			ReplicationFactor: defaultReplication,
			Description:       "Staged Level-1 retry topic with short exponential backoff delay (5s)",
			RetentionMS:       DefaultRetentionMS,
			ConfigEntries: map[string]string{
				"cleanup.policy": "delete",
				"retention.ms":   strconv.FormatInt(DefaultRetentionMS, 10),
			},
		},
		{
			Name:              TopicTxRetry2,
			NumPartitions:     defaultPartitions,
			ReplicationFactor: defaultReplication,
			Description:       "Staged Level-2 retry topic with long exponential backoff delay (30s)",
			RetentionMS:       DefaultRetentionMS,
			ConfigEntries: map[string]string{
				"cleanup.policy": "delete",
				"retention.ms":   strconv.FormatInt(DefaultRetentionMS, 10),
			},
		},
	}
}

// QuickTCPProbe tests low-level TCP port reachability to a broker before invoking full Kafka protocol handshakes.
// This provides immediate and actionable error messages if Docker containers are stopped or host ports are blocked.
func QuickTCPProbe(broker string, timeout time.Duration) (time.Duration, error) {
	start := time.Now()
	conn, err := net.DialTimeout("tcp", broker, timeout)
	if err != nil {
		return 0, fmt.Errorf("TCP connection to Kafka broker %s failed: %w", broker, err)
	}
	defer conn.Close()
	return time.Since(start), nil
}

// FindFirstReachableBroker iterates through configured brokers and returns the first broker
// that accepts a TCP handshake, along with measured round-trip ping latency.
func FindFirstReachableBroker(ctx context.Context, brokers []string, timeout time.Duration) (string, time.Duration, error) {
	if len(brokers) == 0 {
		return "", 0, fmt.Errorf("no broker addresses configured")
	}

	var errs []string
	for _, b := range brokers {
		select {
		case <-ctx.Done():
			return "", 0, ctx.Err()
		default:
		}

		latency, err := QuickTCPProbe(b, timeout)
		if err == nil {
			return b, latency, nil
		}
		errs = append(errs, fmt.Sprintf("[%s: %v]", b, err))
	}

	return "", 0, fmt.Errorf("all Kafka brokers unreachable: %s", strings.Join(errs, ", "))
}
