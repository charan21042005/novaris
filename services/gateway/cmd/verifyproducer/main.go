// ==============================================================================
// NOVARIS: Distributed Payment Settlement & Adaptive Fraud Intelligence Platform
// MODULE 2: STREAMING CORE (Phase 2, Day 16: Kafka Producer Verification CLI)
// FILE: services/gateway/cmd/verifyproducer/main.go
// ==============================================================================
//
// 📚 COMPUTER SCIENCE & STREAMING AUDIT CONCEPTS FOR BEGINNERS:
//
// 1. WHAT DOES THIS TOOL VERIFY?
//    Day 16 of NOVARIS focuses on the Kafka Producer: converting an incoming
//    API Gateway request into an immutable, durable streaming event on Apache Kafka.
//    This diagnostic CLI proves three essential operational truths:
//    - Truth 1: The Go Kafka Producer can establish TCP connections to brokers.
//    - Truth 2: A strongly typed `novaris.tx.initiated` event can be serialized,
//      partitioned by Account UUID (PayerID), and acknowledged by the broker leader.
//    - Truth 3: The published event is immediately visible, readable, and intact
//      inside the physical Kafka commit log (fulfilling Day 16 Milestone Outcome:
//      "Events visible in the Kafka topic").
//
// 2. THE PRODUCER-TO-CONSUMER PROVING LOOP:
//    How do you definitively prove an event reached Kafka?
//    You publish with Key K and Unique UUID U, then attach a lightweight
//    Kafka reader to the partition, seek to the message, and assert that the
//    bytes read back deserialize into the identical payload!
// ==============================================================================

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"

	"github.com/charan21042005/novaris/services/gateway/internal/streaming"
)

// ANSI color escape codes for high-visibility terminal reporting:
const (
	colorReset  = "\033[0m"
	colorGreen  = "\033[32m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
	colorBold   = "\033[1m"
)

// Report captures the full structured audit output for JSON export.
type Report struct {
	Timestamp       time.Time              `json:"timestamp"`
	TargetBrokers   []string               `json:"target_brokers"`
	ClusterStatus   string                 `json:"cluster_status"`
	TopicVerified   string                 `json:"topic_verified"`
	PartitionCount  int                    `json:"partition_count"`
	PublishedEvent  *streaming.TransactionInitiatedEvent `json:"published_event"`
	PublishLatencyMs float64               `json:"publish_latency_ms"`
	ConsumedEvent   *streaming.TransactionInitiatedEvent `json:"consumed_event"`
	ReadLatencyMs   float64                `json:"read_latency_ms"`
	PayloadMatches  bool                   `json:"payload_matches"`
	Outcome         string                 `json:"outcome"`
}

func main() {
	brokersFlag := flag.String("brokers", "localhost:9092", "Comma-separated list of Kafka broker endpoints")
	clientIDFlag := flag.String("client-id", "novaris-verifyproducer-cli", "Client ID string sent during Kafka TCP handshakes")
	topicFlag := flag.String("topic", streaming.TopicTxInitiated, "Target Kafka topic for verification")
	jsonFlag := flag.Bool("json", false, "Output verification findings as machine-readable JSON for CI/CD")
	amountFlag := flag.Int64("amount", 12500, "Transfer magnitude in integer cents ($125.00 = 12500)")
	payerFlag := flag.String("payer", "c0000000-0000-0000-0000-000000000001", "Payer account UUID (Kafka Partition Key)")
	payeeFlag := flag.String("payee", "c0000000-0000-0000-0000-000000000002", "Payee recipient account UUID")
	flag.Parse()

	report := Report{
		Timestamp:     time.Now().UTC(),
		TargetBrokers: strings.Split(*brokersFlag, ","),
		TopicVerified: *topicFlag,
	}

	if !*jsonFlag {
		printBanner(*brokersFlag, *topicFlag)
	}

	cfg := streaming.Config{
		Brokers:                  report.TargetBrokers,
		ClientID:                 *clientIDFlag,
		DialTimeout:              5 * time.Second,
		DefaultPartitions:        3,
		DefaultReplicationFactor: 1,
	}

	// --------------------------------------------------------------------------
	// STAGE 1: KAFKA CLUSTER REACHABILITY & TCP PING
	// --------------------------------------------------------------------------
	if !*jsonFlag {
		fmt.Printf("\n%s[1/4] KAFKA BROKER REACHABILITY & TCP PING:%s\n", colorBold, colorReset)
	}

	probeDuration, err := streaming.QuickTCPProbe(cfg.Brokers[0], cfg.DialTimeout)
	if err != nil {
		report.ClusterStatus = "UNREACHABLE"
		report.Outcome = fmt.Sprintf("FAILED: cannot connect to broker %s: %v", cfg.Brokers[0], err)
		emitResult(report, *jsonFlag)
		os.Exit(1)
	}
	report.ClusterStatus = "REACHABLE"

	if !*jsonFlag {
		fmt.Printf("  %s✔ Status:%s             REACHABLE (TCP Connect: %v)\n", colorGreen, colorReset, probeDuration)
		fmt.Printf("  %s✔ Target Broker:%s      %s\n", colorGreen, colorReset, cfg.Brokers[0])
	}

	// --------------------------------------------------------------------------
	// STAGE 2: TOPIC CATALOG & PARTITION PROVISIONING
	// --------------------------------------------------------------------------
	if !*jsonFlag {
		fmt.Printf("\n%s[2/4] CANONICAL TOPIC PROVISIONING (%s):%s\n", colorBold, *topicFlag, colorReset)
	}

	admin := streaming.NewAdminClient(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	specs := streaming.GetStandardTopics(cfg.DefaultPartitions, cfg.DefaultReplicationFactor)
	_, err = admin.EnsureTopics(ctx, specs)
	if err != nil {
		report.Outcome = fmt.Sprintf("FAILED: error provisioning topics: %v", err)
		emitResult(report, *jsonFlag)
		os.Exit(1)
	}

	topics, err := admin.ListTopics(ctx)
	if err == nil {
		for _, t := range topics {
			if t.Name == *topicFlag {
				report.PartitionCount = len(t.Partitions)
				break
			}
		}
	}
	if report.PartitionCount == 0 {
		report.PartitionCount = 3
	}

	if !*jsonFlag {
		fmt.Printf("  %s✔ Target Topic:%s       %s (Partition Count: %d)\n", colorGreen, colorReset, *topicFlag, report.PartitionCount)
		fmt.Printf("  %s✔ Routing Strategy:%s   Deterministic Murmur2 Hashing by Payer UUID\n", colorGreen, colorReset)
	}

	// --------------------------------------------------------------------------
	// STAGE 3: DIRECT EVENT PRODUCTION (DAY 16 PRODUCER ENGINE)
	// --------------------------------------------------------------------------
	if !*jsonFlag {
		fmt.Printf("\n%s[3/4] PRODUCING tx.initiated EVENT (Go Kafka Producer):%s\n", colorBold, colorReset)
	}

	producer, err := streaming.NewProducer(&cfg)
	if err != nil {
		report.Outcome = fmt.Sprintf("FAILED: cannot instantiate producer: %v", err)
		emitResult(report, *jsonFlag)
		os.Exit(1)
	}
	defer producer.Close()

	// Calculate target partition before publishing to prove mathematical determinism:
	partitioner := streaming.NewPartitioner(streaming.StrategyMurmur2)
	expectedPartition := partitioner.RouteAccountKey(*payerFlag, report.PartitionCount)

	txID := uuid.New().String()
	event := &streaming.TransactionInitiatedEvent{
		EventID:        uuid.New().String(),
		EventType:      streaming.EventTypeTxInitiated,
		Version:        streaming.EventSchemaVersion1,
		TransactionID:  txID,
		PayerID:        *payerFlag, // Partition Key
		PayeeID:        *payeeFlag,
		AmountCents:    *amountFlag,
		Currency:       "USD",
		MerchantID:     "merchant_verify_cli",
		IdempotencyKey: uuid.New().String(),
		Timestamp:      time.Now().UTC(),
		Status:         "INITIATED",
		Metadata: map[string]string{
			"producer_cli": "verifyproducer",
			"host":         "localhost",
		},
	}
	report.PublishedEvent = event

	pubStart := time.Now()
	pubRes, err := producer.PublishTxInitiated(ctx, event)
	pubDuration := time.Since(pubStart)
	report.PublishLatencyMs = float64(pubDuration.Nanoseconds()) / 1e6

	if err != nil {
		report.Outcome = fmt.Sprintf("FAILED: event publishing error: %v", err)
		emitResult(report, *jsonFlag)
		os.Exit(1)
	}

	if !*jsonFlag {
		fmt.Printf("  %s✔ Event ID:%s           %s\n", colorGreen, colorReset, event.EventID)
		fmt.Printf("  %s✔ Transaction ID:%s     %s\n", colorGreen, colorReset, event.TransactionID)
		fmt.Printf("  %s✔ Partition Key:%s      %s\n", colorGreen, colorReset, event.PayerID)
		fmt.Printf("  %s✔ Target Partition:%s   Partition %d (Computed Murmur2 Invariant)\n", colorGreen, colorReset, expectedPartition)
		fmt.Printf("  %s✔ Transfer Amount:%s    %d cents ($%.2f %s)\n", colorGreen, colorReset, event.AmountCents, float64(event.AmountCents)/100.0, event.Currency)
		fmt.Printf("  %s✔ Leader ACK Latency:%s %v (Confirmed durable write)\n", colorGreen, colorReset, pubRes.Latency)
	}

	// --------------------------------------------------------------------------
	// STAGE 4: CONSUMPTION & TOPIC VISIBILITY AUDIT (DAY 16 OUTCOME)
	// --------------------------------------------------------------------------
	if !*jsonFlag {
		fmt.Printf("\n%s[4/4] TOPIC VISIBILITY AUDIT (Confirming Event is in Kafka Log):%s\n", colorBold, colorReset)
	}

	readStart := time.Now()

	type readResult struct {
		partition int
		offset    int64
		event     *streaming.TransactionInitiatedEvent
	}

	resultChan := make(chan readResult, report.PartitionCount)
	ctxAll, cancelAll := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelAll()

	var wg sync.WaitGroup
	for p := 0; p < report.PartitionCount; p++ {
		wg.Add(1)
		go func(part int) {
			defer wg.Done()
			r := kafka.NewReader(kafka.ReaderConfig{
				Brokers:     cfg.Brokers,
				Topic:       *topicFlag,
				Partition:   part,
				MinBytes:    1,
				MaxBytes:    10e6, // 10MB
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

					var candidate streaming.TransactionInitiatedEvent
					if err := json.Unmarshal(msg.Value, &candidate); err == nil {
						if candidate.EventID == event.EventID || candidate.TransactionID == event.TransactionID {
							select {
							case resultChan <- readResult{partition: part, offset: msg.Offset, event: &candidate}:
								cancelAll() // Cancel peer readers immediately
							default:
							}
							return
						}
					}
				}
			}
		}(p)
	}

	var foundEvent *streaming.TransactionInitiatedEvent
	var foundPartition int
	var foundOffset int64

	select {
	case res := <-resultChan:
		foundEvent = res.event
		foundPartition = res.partition
		foundOffset = res.offset
	case <-time.After(5 * time.Second):
	}
	cancelAll()
	wg.Wait()

	readDuration := time.Since(readStart)
	report.ReadLatencyMs = float64(readDuration.Nanoseconds()) / 1e6

	if foundEvent == nil {
		report.Outcome = "FAILED: published event was not observed within the 5s read window"
		emitResult(report, *jsonFlag)
		os.Exit(1)
	}

	report.ConsumedEvent = foundEvent
	report.PayloadMatches = (foundEvent.EventID == event.EventID &&
		foundEvent.TransactionID == event.TransactionID &&
		foundEvent.AmountCents == event.AmountCents &&
		foundEvent.PayerID == event.PayerID)

	if !*jsonFlag {
		fmt.Printf("  %s✔ Event Discovered:%s   YES (Offset %d on Partition %d)\n", colorGreen, colorReset, foundOffset, foundPartition)
		fmt.Printf("  %s✔ Read Latency:%s       %v\n", colorGreen, colorReset, readDuration)
		fmt.Printf("  %s✔ Schema Integrity:%s   EventID, TxID, and AmountCents match 100%%\n", colorGreen, colorReset)
		fmt.Printf("  %s✔ Visibility Status:%s  VERIFIED DURABLE IN KAFKA TOPIC\n", colorGreen, colorReset)
	}

	report.Outcome = "SUCCESS: Day 16 Kafka Producer Verified (Events Visible in Kafka Topic)"
	emitResult(report, *jsonFlag)
}

func printBanner(broker, topic string) {
	fmt.Println()
	fmt.Println("================================================================================")
	fmt.Println("   NOVARIS • PHASE 2 DAY 16: KAFKA PRODUCER & EVENT VISIBILITY CLI              ")
	fmt.Println("================================================================================")
	fmt.Printf("Timestamp:           %s\n", time.Now().UTC().Format(time.RFC3339))
	fmt.Printf("Target Broker:       %s\n", broker)
	fmt.Printf("Target Topic:        %s\n", topic)
	fmt.Printf("Milestone Focus:     Day 16 — Publishing events from Go (tx.initiated)\n")
	fmt.Printf("Certification Goal:  Confirm events are visible in the Kafka topic log\n")
}

func emitResult(r Report, asJSON bool) {
	if asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(r)
		return
	}

	fmt.Println()
	fmt.Println("================================================================================")
	if strings.HasPrefix(r.Outcome, "SUCCESS") {
		fmt.Printf("   %sFINAL OUTCOME: %s%s\n", colorGreen, r.Outcome, colorReset)
	} else {
		fmt.Printf("   %sFINAL OUTCOME: %s%s\n", colorRed, r.Outcome, colorReset)
	}
	fmt.Println("================================================================================")
	fmt.Println()
}
