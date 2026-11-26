// ==============================================================================
// NOVARIS: Distributed Payment Settlement & Adaptive Fraud Intelligence Platform
// FILE: services/gateway/internal/streaming/config.go
// PURPOSE: Apache Kafka Cluster Configuration, Canonical Topics Catalog & TCP Probes
// ==============================================================================
//
// 🎓 COMPUTER SCIENCE & DISTRIBUTED SYSTEMS THEORY FOR BEGINNERS:
//
// 1. The Synchronous Request-Response Anti-Pattern:
//    In traditional monolithic or naive microservice systems, a payment request
//    flows synchronously:
//      Client -> API Gateway -> Auth -> Fraud ML -> Settlement DB -> Notification
//    If any single service in this chain slows down (e.g., Garbage Collection pause,
//    disk contention, or network packet loss), ALL upstream services block waiting
//    for a response. Under heavy traffic, gateway worker pools exhaust their sockets,
//    latency balloons past acceptable SLAs (>200ms), and the entire platform suffers
//    cascading failure.
//
// 2. The Asynchronous Append-Only Commit Log Solution:
//    Apache Kafka decouples the payment ingestion path from downstream processing.
//    The Go API Gateway performs fast authentication and perimeter validation in <15ms,
//    writes an event to Kafka (or via the Transactional Outbox pattern), and responds
//    immediately to the merchant with HTTP 202 Accepted.
//
// 3. Why Configuration Discipline Matters:
//    Distributed streaming brokers require deterministic network timeouts, explicit
//    client identifiers for distributed tracing, and well-defined topic partition counts.
//    This package encapsulates all cluster settings with cloud-native environment
//    variable overrides and robust fallbacks for local Docker Compose development.
// ==============================================================================

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

// ==============================================================================
// Environment Variable Keys (Cloud-Native Configuration Overrides)
// ==============================================================================
// In containerized Kubernetes or Docker deployments, hardcoding IP addresses
// is an anti-pattern. These constants define standard environment variable keys
// that allow 12-factor application configuration without modifying code.
const (
	// EnvKafkaBrokers defines a comma-separated list of Kafka broker addresses (e.g. "kafka1:9092,kafka2:9092").
	EnvKafkaBrokers = "KAFKA_BROKERS"

	// EnvKafkaClientID defines the logical client identifier sent in Kafka request headers for tracing.
	EnvKafkaClientID = "KAFKA_CLIENT_ID"

	// EnvKafkaDialTimeoutMS defines the raw TCP dial timeout in milliseconds.
	EnvKafkaDialTimeoutMS = "KAFKA_DIAL_TIMEOUT_MS"

	// EnvKafkaDefaultPartitions defines the partition count when auto-provisioning topics.
	EnvKafkaDefaultPartitions = "KAFKA_DEFAULT_PARTITIONS"

	// EnvKafkaDefaultReplication defines the replication factor across the broker cluster.
	EnvKafkaDefaultReplication = "KAFKA_DEFAULT_REPLICATION"
)

// ==============================================================================
// Default Configuration Parameters (Local Development & Docker Compose Mesh)
// ==============================================================================
// These defaults ensure that a developer can clone the repository, run `docker compose up`,
// and execute tests immediately without manually configuring environment variables.
const (
	// DefaultBrokerAddress points to the local Docker Compose Kafka port exposed on the host.
	DefaultBrokerAddress = "localhost:9092"

	// DefaultClientID uniquely identifies the API Gateway service across broker audit logs.
	DefaultClientID = "novaris-gateway-service"

	// DefaultDialTimeout specifies the maximum time to wait for a broker TCP connection.
	DefaultDialTimeout = 10 * time.Second

	// DefaultPartitionCount: 3 partitions allows local demonstration of horizontal scaling,
	// partition key hashing, and consumer group rebalancing across parallel workers.
	DefaultPartitionCount = 3

	// DefaultReplicationFactor: In a 1-node local development cluster, the replication factor
	// must be 1. In a multi-broker production cluster, this is set to 3.
	DefaultReplicationFactor = 1

	// DefaultSegmentBytes: 1 GiB commit log segment limit before Kafka rolls to a new segment file.
	DefaultSegmentBytes int64 = 1024 * 1024 * 1024

	// DefaultRetentionMS: 7 days (604,800,000 ms) retention window before old log segments are reclaimed.
	DefaultRetentionMS int64 = 7 * 24 * 60 * 60 * 1000
)

// ==============================================================================
// Canonical NOVARIS Event Topic Catalog
// ==============================================================================
// Each topic represents a distinct semantic stream within the financial transaction lifecycle.
// Decoupling event types into dedicated topics ensures strong domain isolation and allows
// downstream consumer services to subscribe only to the events they care about.
const (
	// TopicTxInitiated carries raw transactions validated by the API Gateway.
	// Architectural Guarantee: Partitioned strictly by AccountID to guarantee per-account FIFO ordering.
	TopicTxInitiated = "novaris.tx.initiated"

	// TopicTxSettled carries financial transactions that have successfully passed
	// double-entry balance verification and written to the immutable Postgres ledger.
	// Subscribed by: Notification Webhook Relay, Merchant Portal Websocket, Analytics Pipeline.
	TopicTxSettled = "novaris.tx.settled"

	// TopicTxFailed carries transactions rejected due to fraud scoring (>0.85 threshold),
	// insufficient funds, or unrecoverable optimistic concurrency conflicts.
	// Subscribed by: Merchant Rejection Webhooks, Security Audit Logs, Risk Dashboard.
	TopicTxFailed = "novaris.tx.failed"

	// TopicTxDLQ is the Dead Letter Queue isolating poisoned, unparsable, or schema-corrupted
	// payloads to prevent head-of-line blocking across worker consumer groups.
	// Retention: 14 days for forensic post-mortem investigation.
	TopicTxDLQ = "novaris.tx.dlq"

	// TopicTxRetry1 is the Level-1 staged retry topic with short exponential backoff (e.g. 5s delay).
	// Used for transient network blips and short database connection timeouts.
	TopicTxRetry1 = "novaris.tx.retry.1"

	// TopicTxRetry2 is the Level-2 staged retry topic with longer exponential backoff (e.g. 30s delay).
	// Used when downstream services require prolonged recovery before giving up.
	TopicTxRetry2 = "novaris.tx.retry.2"
)

// TopicSpec defines the architectural metadata and durability parameters of a Kafka topic.
type TopicSpec struct {
	// Name is the canonical Kafka topic string identifier.
	Name string

	// NumPartitions is the number of parallel commit log partitions on disk.
	// In Kafka, partitions are the fundamental unit of horizontal scalability.
	NumPartitions int

	// ReplicationFactor is the number of broker copies maintained across the cluster.
	// Must be <= the number of active brokers in the cluster.
	ReplicationFactor int

	// Description is a human-readable explanation of the topic's domain role.
	Description string

	// RetentionMS defines data lifetime in milliseconds before log cleanup triggers.
	RetentionMS int64

	// ConfigEntries allows fine-grained broker configuration overrides (e.g. cleanup.policy).
	ConfigEntries map[string]string
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
//
// Pedagogical Note for Beginners:
// Notice how this function handles defensive parsing:
//  1. Trims whitespace from broker lists so "localhost:9092, localhost:9093" doesn't produce leading-space errors.
//  2. Validates positive integers before overriding partition and timeout settings.
//  3. Preserves immutable defaults if any environment variable is empty or malformed.
func LoadConfigFromEnv() Config {
	cfg := Config{
		Brokers:                  []string{DefaultBrokerAddress},
		ClientID:                 DefaultClientID,
		DialTimeout:              DefaultDialTimeout,
		DefaultPartitions:        DefaultPartitionCount,
		DefaultReplicationFactor: DefaultReplicationFactor,
	}

	// 1. Parse Broker List:
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

	// 2. Parse Client ID:
	if clientIDEnv := os.Getenv(EnvKafkaClientID); strings.TrimSpace(clientIDEnv) != "" {
		cfg.ClientID = strings.TrimSpace(clientIDEnv)
	}

	// 3. Parse Dial Timeout (in Milliseconds):
	if timeoutEnv := os.Getenv(EnvKafkaDialTimeoutMS); strings.TrimSpace(timeoutEnv) != "" {
		if ms, err := strconv.Atoi(timeoutEnv); err == nil && ms > 0 {
			cfg.DialTimeout = time.Duration(ms) * time.Millisecond
		}
	}

	// 4. Parse Default Partitions:
	if partEnv := os.Getenv(EnvKafkaDefaultPartitions); strings.TrimSpace(partEnv) != "" {
		if parts, err := strconv.Atoi(partEnv); err == nil && parts > 0 {
			cfg.DefaultPartitions = parts
		}
	}

	// 5. Parse Default Replication Factor:
	if repEnv := os.Getenv(EnvKafkaDefaultReplication); strings.TrimSpace(repEnv) != "" {
		if rep, err := strconv.Atoi(repEnv); err == nil && rep > 0 {
			cfg.DefaultReplicationFactor = rep
		}
	}

	return cfg
}

// GetStandardTopics returns the authoritative catalog of topics required for NOVARIS Phase 2.
//
// Architectural Rationale:
// By defining topics declaratively in Go code rather than running manual CLI scripts,
// our deployment pipeline and automated test harness can verify and auto-provision
// topics idempotently on application boot!
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
			RetentionMS:       DefaultRetentionMS * 2, // 14 days retention for post-mortem forensic investigation
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
//
// Why is this critical in production?
// When a container is starting or a host port is misconfigured, full Kafka protocol clients
// often hang or output cryptic protocol framing errors after long timeout delays.
// A low-level TCP SYN-ACK dial fails immediately (<50ms), allowing our diagnostic CLI
// to inform the developer precisely whether the Docker container is stopped or blocked!
func QuickTCPProbe(broker string, timeout time.Duration) (time.Duration, error) {
	start := time.Now()
	conn, err := net.DialTimeout("tcp", broker, timeout)
	if err != nil {
		return 0, fmt.Errorf("TCP connection to Kafka broker %s failed: %w", broker, err)
	}
	defer conn.Close()
	return time.Since(start), nil
}

// FindFirstReachableBroker iterates through configured bootstrap brokers and returns the first broker
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
