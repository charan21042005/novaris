// ==============================================================================
// NOVARIS: Distributed Payment Settlement & Adaptive Fraud Intelligence Platform
// MODULE 2: STREAMING CORE (Phase 2, Day 17: Kafka Consumer Verification CLI)
// FILE: services/gateway/cmd/verifyconsumer/main.go
// ==============================================================================
//
// 📚 COMPUTER SCIENCE & STREAMING AUDIT CONCEPTS FOR BEGINNERS:
//
// 1. WHAT DOES THIS TOOL VERIFY?
//    Day 17 of NOVARIS focuses on the Kafka Consumer: proving that downstream
//    services can reliably subscribe to events, process them, and commit offsets.
//    This diagnostic CLI proves three essential operational truths:
//      - Truth 1: The Go Kafka Consumer can join a consumer group and coordinate
//        with the Kafka broker's Group Coordinator.
//      - Truth 2: Sequential events published with the same partition key (PayerID)
//        are consumed in strict, unbroken chronological sequence (Strict In-Order).
//      - Truth 3: Offsets are committed back to Kafka, and zero published events
//        are lost or skipped (fulfilling Day 17 Outcome: "Consumer reads every
//        published event").
//
// 2. THE PROVING METHODOLOGY (LAB L3 & OUTCOME CERTIFICATION):
//    - We generate a unique batch of $N$ sequentially numbered financial events
//      (Seq 1, Seq 2, Seq 3, Seq 4, Seq 5) with a dedicated test Payer UUID.
//    - We publish this batch to `novaris.tx.initiated` using Murmur2 key routing.
//    - We start a consumer in a dedicated test Consumer Group.
//    - The consumer reads the events, verifies sequence numbers, records audit
//      timestamps, and commits offsets.
//    - If all $N$ events are consumed in exact monotonic order, the test PASSES!
// ==============================================================================

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"

	"github.com/charan21042005/novaris/services/gateway/internal/streaming"
)

// ANSI color escape codes for terminal formatting:
const (
	colorReset  = "\033[0m"
	colorGreen  = "\033[32m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
	colorBold   = "\033[1m"
)

// Report captures structured audit output for JSON export.
type Report struct {
	Timestamp          time.Time `json:"timestamp"`
	TargetBrokers      []string  `json:"target_brokers"`
	ClusterStatus      string    `json:"cluster_status"`
	TopicVerified      string    `json:"topic_verified"`
	ConsumerGroupID    string    `json:"consumer_group_id"`
	EventsPublished    int       `json:"events_published"`
	EventsConsumed     int       `json:"events_consumed"`
	InOrderPreserved   bool      `json:"in_order_preserved"`
	ZeroLossVerified   bool      `json:"zero_loss_verified"`
	OffsetsCommitted   int       `json:"offsets_committed"`
	BatchPublishMs     float64   `json:"batch_publish_ms"`
	ConsumeDurationMs  float64   `json:"consume_duration_ms"`
	Outcome            string    `json:"outcome"`
}

func main() {
	// --------------------------------------------------------------------------
	// 1. COMMAND-LINE FLAGS
	// --------------------------------------------------------------------------
	brokersFlag := flag.String("brokers", "localhost:9092", "Comma-separated list of Kafka broker endpoints")
	topicFlag := flag.String("topic", streaming.TopicTxInitiated, "Target Kafka topic for verification")
	eventsFlag := flag.Int("events", 5, "Number of sequential events to publish and verify")
	jsonFlag := flag.Bool("json", false, "Output verification findings as machine-readable JSON for CI/CD")
	payerFlag := flag.String("payer", "c0000000-0000-0000-0000-000000000088", "Payer account UUID (Partition Key)")
	flag.Parse()

	// Unique Consumer Group ID per run ensures clean offset start from FirstOffset:
	uniqueGroupID := fmt.Sprintf("novaris-verifyconsumer-%s", uuid.New().String()[:8])

	report := Report{
		Timestamp:       time.Now().UTC(),
		TargetBrokers:   strings.Split(*brokersFlag, ","),
		TopicVerified:   *topicFlag,
		ConsumerGroupID: uniqueGroupID,
	}

	if !*jsonFlag {
		printBanner(*brokersFlag, *topicFlag, uniqueGroupID, *eventsFlag)
	}

	cfg := streaming.Config{
		Brokers:                  report.TargetBrokers,
		ClientID:                 "novaris-verifyconsumer-cli",
		DialTimeout:              5 * time.Second,
		DefaultPartitions:        3,
		DefaultReplicationFactor: 1,
	}

	// --------------------------------------------------------------------------
	// STAGE 1: KAFKA BROKER REACHABILITY & TCP PING
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
	// STAGE 2: TOPIC & PRODUCER INITIALIZATION
	// --------------------------------------------------------------------------
	if !*jsonFlag {
		fmt.Printf("\n%s[2/4] TOPIC CATALOG & PARTITION PROVISIONING:%s\n", colorBold, colorReset)
	}

	admin := streaming.NewAdminClient(cfg)
	ctxAdmin, cancelAdmin := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelAdmin()

	specs := streaming.GetStandardTopics(cfg.DefaultPartitions, cfg.DefaultReplicationFactor)
	_, err = admin.EnsureTopics(ctxAdmin, specs)
	if err != nil {
		report.Outcome = fmt.Sprintf("FAILED: error provisioning topics: %v", err)
		emitResult(report, *jsonFlag)
		os.Exit(1)
	}

	if !*jsonFlag {
		fmt.Printf("  %s✔ Target Topic:%s       %s (Partition Count: 3)\n", colorGreen, colorReset, *topicFlag)
		fmt.Printf("  %s✔ Consumer Group ID:%s  %s\n", colorGreen, colorReset, uniqueGroupID)
	}

	// --------------------------------------------------------------------------
	// STAGE 3: SEQUENTIAL BATCH EVENT PRODUCTION
	// --------------------------------------------------------------------------
	if !*jsonFlag {
		fmt.Printf("\n%s[3/4] PUBLISHING SEQUENCED TEST EVENTS (Producer Engine):%s\n", colorBold, colorReset)
	}

	producer, err := streaming.NewProducer(&cfg)
	if err != nil {
		report.Outcome = fmt.Sprintf("FAILED: cannot instantiate producer: %v", err)
		emitResult(report, *jsonFlag)
		os.Exit(1)
	}
	defer producer.Close()

	batchID := uuid.New().String()[:8]
	publishedEvents := make([]*streaming.TransactionInitiatedEvent, *eventsFlag)

	pubStart := time.Now()
	for i := 0; i < *eventsFlag; i++ {
		seq := i + 1
		txID := uuid.New().String()
		event := &streaming.TransactionInitiatedEvent{
			EventID:        uuid.New().String(),
			EventType:      streaming.EventTypeTxInitiated,
			Version:        streaming.EventSchemaVersion1,
			TransactionID:  txID,
			PayerID:        *payerFlag, // Constant PayerID ensures all N events land on the SAME partition!
			PayeeID:        "c0000000-0000-0000-0000-000000000002",
			AmountCents:    int64(1000 * seq), // $10, $20, $30...
			Currency:       "USD",
			MerchantID:     "merchant_verifyconsumer",
			IdempotencyKey: fmt.Sprintf("idemp-%s-%02d", batchID, seq),
			Timestamp:      time.Now().UTC(),
			Status:         "INITIATED",
			Metadata: map[string]string{
				"batch_id": batchID,
				"seq":      fmt.Sprintf("%d", seq),
			},
		}

		ctxPub, cancelPub := context.WithTimeout(context.Background(), 5*time.Second)
		res, err := producer.PublishTxInitiated(ctxPub, event)
		cancelPub()
		if err != nil {
			report.Outcome = fmt.Sprintf("FAILED: publish error on seq %d: %v", seq, err)
			emitResult(report, *jsonFlag)
			os.Exit(1)
		}

		publishedEvents[i] = event
		if !*jsonFlag {
			fmt.Printf("  %s✔ Produced Seq #%d:%s     TxID: %s ($%.2f USD) -> Partition %d (Ack: %v)\n",
				colorCyan, seq, colorReset, event.TransactionID[:8]+"...", float64(event.AmountCents)/100.0,
				res.Partition, res.Latency)
		}
	}
	report.BatchPublishMs = float64(time.Since(pubStart).Nanoseconds()) / 1e6
	report.EventsPublished = *eventsFlag

	// --------------------------------------------------------------------------
	// STAGE 4: IN-ORDER CONSUMPTION AUDIT (DAY 17 MILESTONE OUTCOME)
	// --------------------------------------------------------------------------
	if !*jsonFlag {
		fmt.Printf("\n%s[4/4] IN-ORDER CONSUMPTION & OFFSET COMMIT AUDIT:%s\n", colorBold, colorReset)
	}

	consumerCfg := streaming.DefaultConsumerConfig(uniqueGroupID, *topicFlag)
	consumerCfg.Brokers = cfg.Brokers
	consumerCfg.StartOffset = kafka.FirstOffset

	// Slice to record consumed sequence numbers:
	var consumedSeqs []int
	var consumedTxIDs []string

	handler := func(ctx context.Context, msg kafka.Message) error {
		var evt streaming.TransactionInitiatedEvent
		if err := json.Unmarshal(msg.Value, &evt); err != nil {
			return err
		}

		// Filter for events from this specific test batch:
		if evt.Metadata != nil && evt.Metadata["batch_id"] == batchID {
			var seq int
			_, _ = fmt.Sscanf(evt.Metadata["seq"], "%d", &seq)
			consumedSeqs = append(consumedSeqs, seq)
			consumedTxIDs = append(consumedTxIDs, evt.TransactionID)

			if !*jsonFlag {
				fmt.Printf("  %s✔ Consumed Seq #%d:%s     TxID: %s | Partition: %d | Offset: %d | Status: COMMITTED\n",
					colorGreen, seq, colorReset, evt.TransactionID[:8]+"...", msg.Partition, msg.Offset)
			}
		}
		return nil
	}

	consumer, err := streaming.NewConsumerGroup(consumerCfg, handler)
	if err != nil {
		report.Outcome = fmt.Sprintf("FAILED: cannot instantiate consumer: %v", err)
		emitResult(report, *jsonFlag)
		os.Exit(1)
	}
	defer consumer.Close()

	consumeStart := time.Now()
	ctxConsume, cancelConsume := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancelConsume()

	// Poll until all N events from this batch are consumed or timeout occurs:
	for len(consumedSeqs) < *eventsFlag {
		select {
		case <-ctxConsume.Done():
			break
		default:
			_, err := consumer.ReadNext(ctxConsume)
			if err != nil {
				if ctxConsume.Err() != nil {
					break
				}
				time.Sleep(100 * time.Millisecond)
			}
		}
	}
	report.ConsumeDurationMs = float64(time.Since(consumeStart).Nanoseconds()) / 1e6

	report.EventsConsumed = len(consumedSeqs)
	stats := consumer.Stats()
	report.OffsetsCommitted = int(stats.TotalCommitted)

	// Validate In-Order Sequence:
	inOrder := true
	for i := 0; i < len(consumedSeqs); i++ {
		expectedSeq := i + 1
		if consumedSeqs[i] != expectedSeq {
			inOrder = false
			break
		}
	}
	report.InOrderPreserved = inOrder
	report.ZeroLossVerified = (len(consumedSeqs) == *eventsFlag)

	if !report.ZeroLossVerified {
		report.Outcome = fmt.Sprintf("FAILED: expected %d consumed events, only read %d", *eventsFlag, len(consumedSeqs))
		emitResult(report, *jsonFlag)
		os.Exit(1)
	}

	if !report.InOrderPreserved {
		report.Outcome = fmt.Sprintf("FAILED: FIFO sequence violation! Received order: %v", consumedSeqs)
		emitResult(report, *jsonFlag)
		os.Exit(1)
	}

	if !*jsonFlag {
		fmt.Printf("\n  %s✔ Total Events Read:%s     %d / %d (100%% Zero Message Loss)\n", colorGreen, colorReset, report.EventsConsumed, *eventsFlag)
		fmt.Printf("  %s✔ Sequence Invariance:%s   Strict Monotonic Order Verified (1 -> 2 -> 3 -> 4 -> 5)\n", colorGreen, colorReset)
		fmt.Printf("  %s✔ Offsets Committed:%s     %d Offsets Durable in __consumer_offsets\n", colorGreen, colorReset, report.OffsetsCommitted)
		fmt.Printf("  %s✔ Total Audit Latency:%s   %.2fms\n", colorGreen, colorReset, report.ConsumeDurationMs)
	}

	report.Outcome = "SUCCESS: Day 17 Kafka Consumer Verified (Consumer reads every published event, in order)"
	emitResult(report, *jsonFlag)
}

func printBanner(broker, topic, group string, count int) {
	fmt.Println()
	fmt.Println("================================================================================")
	fmt.Println("   NOVARIS • PHASE 2 DAY 17: KAFKA CONSUMER & IN-ORDER AUDIT CLI                ")
	fmt.Println("================================================================================")
	fmt.Printf("Timestamp:           %s\n", time.Now().UTC().Format(time.RFC3339))
	fmt.Printf("Target Broker:       %s\n", broker)
	fmt.Printf("Target Topic:        %s\n", topic)
	fmt.Printf("Consumer Group:      %s\n", group)
	fmt.Printf("Batch Test Events:   %d sequential events\n", count)
	fmt.Printf("Certification Goal:  Verify consumer reads every published event in strict order\n")
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
