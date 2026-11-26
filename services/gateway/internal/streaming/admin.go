// Package streaming provides high-performance, fault-tolerant Apache Kafka event streaming
// primitives for the NOVARIS payment settlement and adaptive fraud intelligence platform.
package streaming

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
)

// BrokerMetadata represents the connection and role details of a single Kafka node.
type BrokerMetadata struct {
	ID           int    `json:"id"`
	Host         string `json:"host"`
	Port         int    `json:"port"`
	Address      string `json:"address"`
	IsController bool   `json:"is_controller"`
	Rack         string `json:"rack,omitempty"`
}

// PartitionMetadata represents the log partition details, leader broker, and replica set.
type PartitionMetadata struct {
	Topic     string `json:"topic"`
	ID        int    `json:"id"`
	LeaderID  int    `json:"leader_id"`
	Replicas  []int  `json:"replicas"`
	Isr       []int  `json:"isr"` // In-Sync Replicas
	ErrorDesc string `json:"error,omitempty"`
}

// TopicSummary represents a topic's active operational state across the cluster.
type TopicSummary struct {
	Name              string              `json:"name"`
	NumPartitions     int                 `json:"num_partitions"`
	ReplicationFactor int                 `json:"replication_factor"`
	Partitions        []PartitionMetadata `json:"partitions"`
	IsInternal        bool                `json:"is_internal"`
}

// ClusterHealth encapsulates the overall operational status and latency of the Kafka cluster.
type ClusterHealth struct {
	Reachable        bool             `json:"reachable"`
	BootstrapBrokers []string         `json:"bootstrap_brokers"`
	ConnectedBroker  string           `json:"connected_broker"`
	Latency          time.Duration    `json:"latency"`
	LatencyMS        float64          `json:"latency_ms"`
	ControllerID     int              `json:"controller_id"`
	ControllerHost   string           `json:"controller_host"`
	Brokers          []BrokerMetadata `json:"brokers"`
	TotalTopics      int              `json:"total_topics"`
	Timestamp        time.Time        `json:"timestamp"`
	ErrorMessage     string           `json:"error_message,omitempty"`
}

// TopicProvisionResult records the outcome of provisioning a single Kafka topic.
type TopicProvisionResult struct {
	Topic             string `json:"topic"`
	Partitions        int    `json:"partitions"`
	ReplicationFactor int    `json:"replication_factor"`
	Status            string `json:"status"` // "CREATED", "ALREADY_EXISTS", "FAILED"
	Details           string `json:"details"`
}

// TopicProvisionReport aggregates the results of an idempotent topic provisioning run.
type TopicProvisionReport struct {
	TotalRequested int                    `json:"total_requested"`
	CreatedCount   int                    `json:"created_count"`
	ExistingCount  int                    `json:"existing_count"`
	FailedCount    int                    `json:"failed_count"`
	Results        []TopicProvisionResult `json:"results"`
	Duration       time.Duration          `json:"duration"`
}

// ProbeResult captures the round-trip diagnostic metrics of a test event write.
type ProbeResult struct {
	Topic        string        `json:"topic"`
	Partition    int           `json:"partition"`
	Offset       int64         `json:"offset"`
	Key          string        `json:"key"`
	PayloadBytes int           `json:"payload_bytes"`
	WriteLatency time.Duration `json:"write_latency"`
	Timestamp    time.Time     `json:"timestamp"`
}

// AdminClient provides programmatic management, health probing, and topic provisioning for Apache Kafka.
type AdminClient struct {
	cfg Config
}

// NewAdminClient creates a new Kafka AdminClient with the provided configuration.
func NewAdminClient(cfg Config) *AdminClient {
	return &AdminClient{cfg: cfg}
}

// CheckHealth probes the Kafka cluster over TCP, queries the KRaft controller metadata,
// and returns a detailed cluster health report.
func (a *AdminClient) CheckHealth(ctx context.Context) (*ClusterHealth, error) {
	start := time.Now()

	reachableBroker, _, err := FindFirstReachableBroker(ctx, a.cfg.Brokers, a.cfg.DialTimeout)
	if err != nil {
		return &ClusterHealth{
			Reachable:        false,
			BootstrapBrokers: a.cfg.Brokers,
			Timestamp:        time.Now().UTC(),
			ErrorMessage:     err.Error(),
		}, err
	}

	dialer := &kafka.Dialer{
		Timeout:   a.cfg.DialTimeout,
		DualStack: true,
		ClientID:  a.cfg.ClientID,
	}

	conn, err := dialer.DialContext(ctx, "tcp", reachableBroker)
	if err != nil {
		return &ClusterHealth{
			Reachable:        false,
			BootstrapBrokers: a.cfg.Brokers,
			ConnectedBroker:  reachableBroker,
			Timestamp:        time.Now().UTC(),
			ErrorMessage:     fmt.Sprintf("Kafka protocol dial failed: %v", err),
		}, err
	}
	defer conn.Close()

	// Query cluster brokers:
	rawBrokers, err := conn.Brokers()
	if err != nil {
		return &ClusterHealth{
			Reachable:        false,
			BootstrapBrokers: a.cfg.Brokers,
			ConnectedBroker:  reachableBroker,
			Timestamp:        time.Now().UTC(),
			ErrorMessage:     fmt.Sprintf("failed to query broker metadata: %v", err),
		}, err
	}

	// Query controller node:
	controllerBroker, err := conn.Controller()
	controllerID := -1
	controllerHost := ""
	if err == nil {
		controllerID = controllerBroker.ID
		controllerHost = fmt.Sprintf("%s:%d", controllerBroker.Host, controllerBroker.Port)
	}

	var brokerMetas []BrokerMetadata
	for _, b := range rawBrokers {
		brokerMetas = append(brokerMetas, BrokerMetadata{
			ID:           b.ID,
			Host:         b.Host,
			Port:         b.Port,
			Address:      fmt.Sprintf("%s:%d", b.Host, b.Port),
			IsController: b.ID == controllerID,
			Rack:         b.Rack,
		})
	}

	// Read partitions/topics metadata:
	partitions, err := conn.ReadPartitions()
	uniqueTopics := make(map[string]struct{})
	if err == nil {
		for _, p := range partitions {
			uniqueTopics[p.Topic] = struct{}{}
		}
	}

	latency := time.Since(start)

	return &ClusterHealth{
		Reachable:        true,
		BootstrapBrokers: a.cfg.Brokers,
		ConnectedBroker:  reachableBroker,
		Latency:          latency,
		LatencyMS:        float64(latency.Microseconds()) / 1000.0,
		ControllerID:     controllerID,
		ControllerHost:   controllerHost,
		Brokers:          brokerMetas,
		TotalTopics:      len(uniqueTopics),
		Timestamp:        time.Now().UTC(),
	}, nil
}

// ListTopics inspects the Kafka cluster and returns details on all existing topics and their partitions.
func (a *AdminClient) ListTopics(ctx context.Context) ([]TopicSummary, error) {
	reachableBroker, _, err := FindFirstReachableBroker(ctx, a.cfg.Brokers, a.cfg.DialTimeout)
	if err != nil {
		return nil, err
	}

	dialer := &kafka.Dialer{
		Timeout:   a.cfg.DialTimeout,
		DualStack: true,
		ClientID:  a.cfg.ClientID,
	}

	conn, err := dialer.DialContext(ctx, "tcp", reachableBroker)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to broker %s: %w", reachableBroker, err)
	}
	defer conn.Close()

	rawPartitions, err := conn.ReadPartitions()
	if err != nil {
		return nil, fmt.Errorf("failed to read cluster partitions: %w", err)
	}

	topicMap := make(map[string][]PartitionMetadata)
	for _, p := range rawPartitions {
		var isrList []int
		for _, node := range p.Isr {
			isrList = append(isrList, node.ID)
		}
		var repList []int
		for _, node := range p.Replicas {
			repList = append(repList, node.ID)
		}

		topicMap[p.Topic] = append(topicMap[p.Topic], PartitionMetadata{
			Topic:    p.Topic,
			ID:       p.ID,
			LeaderID: p.Leader.ID,
			Replicas: repList,
			Isr:      isrList,
		})
	}

	var summaries []TopicSummary
	for name, parts := range topicMap {
		repFactor := 0
		if len(parts) > 0 {
			repFactor = len(parts[0].Replicas)
		}
		summaries = append(summaries, TopicSummary{
			Name:              name,
			NumPartitions:     len(parts),
			ReplicationFactor: repFactor,
			Partitions:        parts,
			IsInternal:        strings.HasPrefix(name, "__"),
		})
	}

	return summaries, nil
}

// EnsureTopics idempotently verifies that all required NOVARIS topics exist.
// If a topic is missing, it is created with the specified partitions and replication factor.
// If it already exists, its status is verified and recorded without raising an error.
func (a *AdminClient) EnsureTopics(ctx context.Context, specs []TopicSpec) (*TopicProvisionReport, error) {
	start := time.Now()
	report := &TopicProvisionReport{
		TotalRequested: len(specs),
	}

	reachableBroker, _, err := FindFirstReachableBroker(ctx, a.cfg.Brokers, a.cfg.DialTimeout)
	if err != nil {
		return nil, err
	}

	dialer := &kafka.Dialer{
		Timeout:   a.cfg.DialTimeout,
		DualStack: true,
		ClientID:  a.cfg.ClientID,
	}

	// Dial the controller broker if possible for authoritative metadata modification:
	conn, err := dialer.DialContext(ctx, "tcp", reachableBroker)
	if err != nil {
		return nil, fmt.Errorf("failed to dial broker %s: %w", reachableBroker, err)
	}
	defer conn.Close()

	controller, err := conn.Controller()
	var targetConn *kafka.Conn
	if err == nil {
		controllerAddr := net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port))
		cConn, cErr := dialer.DialContext(ctx, "tcp", controllerAddr)
		if cErr == nil {
			targetConn = cConn
			defer targetConn.Close()
		}
	}
	if targetConn == nil {
		targetConn = conn
	}

	// Read existing topics:
	existingPartitions, err := targetConn.ReadPartitions()
	if err != nil {
		return nil, fmt.Errorf("failed to read existing partitions: %w", err)
	}

	existingTopicParts := make(map[string]int)
	for _, p := range existingPartitions {
		existingTopicParts[p.Topic]++
	}

	var toCreate []kafka.TopicConfig
	for _, spec := range specs {
		if parts, exists := existingTopicParts[spec.Name]; exists {
			report.ExistingCount++
			report.Results = append(report.Results, TopicProvisionResult{
				Topic:             spec.Name,
				Partitions:        parts,
				ReplicationFactor: spec.ReplicationFactor,
				Status:            "ALREADY_EXISTS",
				Details:           fmt.Sprintf("Topic active with %d partition(s)", parts),
			})
			continue
		}

		var configs []kafka.ConfigEntry
		for k, v := range spec.ConfigEntries {
			configs = append(configs, kafka.ConfigEntry{
				ConfigName:  k,
				ConfigValue: v,
			})
		}

		toCreate = append(toCreate, kafka.TopicConfig{
			Topic:             spec.Name,
			NumPartitions:     spec.NumPartitions,
			ReplicationFactor: spec.ReplicationFactor,
			ConfigEntries:     configs,
		})
	}

	if len(toCreate) > 0 {
		err = targetConn.CreateTopics(toCreate...)
		if err != nil {
			// Some versions return TopicAlreadyExists error if another client created it concurrently
			if !isTopicAlreadyExistsErr(err) {
				return nil, fmt.Errorf("failed to create Kafka topics: %w", err)
			}
		}

		for _, tc := range toCreate {
			report.CreatedCount++
			report.Results = append(report.Results, TopicProvisionResult{
				Topic:             tc.Topic,
				Partitions:        tc.NumPartitions,
				ReplicationFactor: tc.ReplicationFactor,
				Status:            "CREATED",
				Details:           fmt.Sprintf("Successfully provisioned with %d partitions and replication %d", tc.NumPartitions, tc.ReplicationFactor),
			})
		}
	}

	report.Duration = time.Since(start)
	return report, nil
}

// ProduceProbeMessage sends a lightweight diagnostic event to the specified topic
// using the provided routing key, measuring write latency and validating the producer path.
func (a *AdminClient) ProduceProbeMessage(ctx context.Context, topic, key, payload string) (*ProbeResult, error) {
	start := time.Now()

	writer := &kafka.Writer{
		Addr:         kafka.TCP(a.cfg.Brokers...),
		Topic:        topic,
		Balancer:     &kafka.Hash{}, // Guarantees deterministic partition hashing by key
		WriteTimeout: a.cfg.DialTimeout,
		ReadTimeout:  a.cfg.DialTimeout,
		RequiredAcks: kafka.RequireOne, // Leader ACK confirmation
	}
	defer writer.Close()

	msg := kafka.Message{
		Key:   []byte(key),
		Value: []byte(payload),
		Time:  time.Now().UTC(),
		Headers: []kafka.Header{
			{Key: "novaris-probe-version", Value: []byte("day15")},
			{Key: "novaris-producer", Value: []byte(a.cfg.ClientID)},
		},
	}

	err := writer.WriteMessages(ctx, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to write probe message to topic %s: %w", topic, err)
	}

	latency := time.Since(start)

	// Determine which partition it was routed to via Partitioner:
	partitioner := NewPartitioner(StrategyFNV1a)
	estimatedPartition := partitioner.RouteAccountKey(key, a.cfg.DefaultPartitions)

	return &ProbeResult{
		Topic:        topic,
		Partition:    estimatedPartition,
		Key:          key,
		PayloadBytes: len(payload),
		WriteLatency: latency,
		Timestamp:    time.Now().UTC(),
	}, nil
}

func isTopicAlreadyExistsErr(err error) bool {
	if err == nil {
		return false
	}
	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "already exists") || errors.Is(err, kafka.TopicAlreadyExists)
}
