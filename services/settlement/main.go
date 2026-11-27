// ==============================================================================
// NOVARIS: Distributed Payment Settlement & Adaptive Fraud Intelligence Platform
// MODULE 2: STREAMING CORE (Phase 2, Day 17: Settlement Consumer Daemon)
// FILE: services/settlement/main.go
// ==============================================================================
//
// 📚 COMPUTER SCIENCE & PRODUCTION DAEMON ARCHITECTURE FOR BEGINNERS:
//
// 1. WHAT IS A DAEMON SERVICE?
//    In computing, a "daemon" (or background service) is a long-running process
//    that does not have a user interface and runs continuously in the background,
//    waiting for work to arrive.
//    This Settlement Worker Daemon starts up, registers itself with Apache Kafka
//    as a member of the "novaris-settlement-workers" consumer group, and sleeps
//    efficiently on the network socket until payment events arrive on `novaris.tx.initiated`.
//
// 2. THE THREE LIFECYCLE PHASES OF THIS SERVICE:
//    - Phase 1 (Bootstrap & Handshake):
//      Parses CLI flags/env vars, verifies broker reachability via TCP ping,
//      and constructs the Kafka Reader consumer.
//    - Phase 2 (Continuous Processing Loop):
//      Consumes messages from Kafka, calls `SettlementProcessor.HandleMessage`,
//      logs structured transaction audit entries, and commits offsets.
//    - Phase 3 (Graceful Signal Shutdown):
//      Intercepts OS interrupt signals (Ctrl+C / SIGINT, or SIGTERM from Kubernetes),
//      stops polling, allows in-flight database writes to finish, commits final offsets,
//      and terminates cleanly with zero dropped transactions!
// ==============================================================================

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/segmentio/kafka-go"
)

// ANSI color escape codes for clean terminal output:
const (
	colorReset  = "\033[0m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
	colorBold   = "\033[1m"
)

func main() {
	// --------------------------------------------------------------------------
	// 1. CONFIGURATION & COMMAND-LINE FLAGS
	// --------------------------------------------------------------------------
	brokersFlag := flag.String("brokers", "localhost:9092", "Comma-separated list of Kafka broker endpoints")
	groupFlag := flag.String("group", "novaris-settlement-workers", "Kafka Consumer Group ID")
	topicFlag := flag.String("topic", TopicTxInitiated, "Kafka topic to subscribe to")
	flag.Parse()

	printBanner(*brokersFlag, *groupFlag, *topicFlag)

	// Check environment variable overrides:
	if envBrokers := os.Getenv("KAFKA_BROKERS"); envBrokers != "" {
		*brokersFlag = envBrokers
	}
	if envGroup := os.Getenv("KAFKA_CONSUMER_GROUP"); envGroup != "" {
		*groupFlag = envGroup
	}

	brokerList := strings.Split(*brokersFlag, ",")

	// --------------------------------------------------------------------------
	// 2. BROKER TCP REACHABILITY PING
	// --------------------------------------------------------------------------
	log.Printf("[INIT] Probing Kafka broker connection at %s...", brokerList[0])
	startDial := time.Now()
	conn, err := net.DialTimeout("tcp", brokerList[0], 5*time.Second)
	if err != nil {
		log.Fatalf("[FATAL] Cannot connect to Kafka broker %s: %v", brokerList[0], err)
	}
	_ = conn.Close()
	log.Printf("[INIT] ✓ Kafka broker reachable (TCP Connect: %v)", time.Since(startDial))

	// --------------------------------------------------------------------------
	// 3. INITIALIZE BUSINESS PROCESSOR & KAFKA READER
	// --------------------------------------------------------------------------
	processor := NewSettlementProcessor()

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokerList,
		GroupID:        *groupFlag, // Non-empty GroupID enables Kafka Consumer Group protocol!
		Topic:          *topicFlag,
		MinBytes:       1,                      // Fetch immediately when at least 1 byte is available
		MaxBytes:       10e6,                   // 10MB batch ceiling
		MaxWait:        500 * time.Millisecond, // 500ms max broker poll latency
		StartOffset:    kafka.FirstOffset,      // Scan from beginning if new consumer group
		CommitInterval: 0,                      // 0 = Manual At-Least-Once commits via CommitMessages
	})
	defer reader.Close()

	log.Printf("[INIT] ✓ Consumer Group %q successfully initialized on topic %q", *groupFlag, *topicFlag)

	// --------------------------------------------------------------------------
	// 4. GRACEFUL SHUTDOWN SIGNAL LISTENER (SIGINT / SIGTERM)
	// --------------------------------------------------------------------------
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Printf("\n[SHUTDOWN] Received termination signal (%v). Initiating graceful shutdown...", sig)
		cancel() // Triggers consumer polling loop to exit cleanly
	}()

	// --------------------------------------------------------------------------
	// 5. START CONSUMER PROCESSING LOOP (AT-LEAST-ONCE SEMANTICS)
	// --------------------------------------------------------------------------
	log.Println("[RUNNING] Settlement worker active. Listening for tx.initiated events...")

	var totalConsumed int64
	var totalCommitted int64
	var totalErrors int64

	for {
		select {
		case <-ctx.Done():
			goto shutdown
		default:
		}

		// 1. Fetch message WITHOUT auto-committing:
		msg, err := reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				goto shutdown
			}
			log.Printf("[ERROR] Fetch error: %v", err)
			totalErrors++
			time.Sleep(500 * time.Millisecond)
			continue
		}

		totalConsumed++

		// 2. Execute business logic settlement handler:
		if err := processor.HandleMessage(ctx, msg); err != nil {
			log.Printf("[ERROR] Handler failed for offset %d: %v", msg.Offset, err)
			totalErrors++
			continue // Do NOT commit offset on error (At-Least-Once safety)
		}

		// 3. Commit offset to Kafka only after successful business processing:
		if err := reader.CommitMessages(ctx, msg); err != nil {
			log.Printf("[ERROR] Commit failed for offset %d: %v", msg.Offset, err)
			totalErrors++
			continue
		}

		totalCommitted++
	}

shutdown:
	// --------------------------------------------------------------------------
	// 6. FINAL SHUTDOWN AUDIT TELEMETRY
	// --------------------------------------------------------------------------
	debits, credits := processor.TotalVolumes()
	log.Println("================================================================================")
	log.Printf("[AUDIT] Settlement Worker Stopped. Total Events Processed: %d", processor.GetProcessedCount())
	log.Printf("[AUDIT] Total Consumed: %d | Offsets Committed: %d | Errors: %d", totalConsumed, totalCommitted, totalErrors)
	log.Printf("[AUDIT] Cumulative Volume Settled: $%.2f USD (Debits: %d cents, Credits: %d cents)",
		float64(debits)/100.0, debits, credits)
	log.Println("================================================================================")
}

func printBanner(brokers, group, topic string) {
	fmt.Println()
	fmt.Println("================================================================================")
	fmt.Println("   NOVARIS • PHASE 2 DAY 17: SETTLEMENT CONSUMER WORKER DAEMON                  ")
	fmt.Println("================================================================================")
	fmt.Printf("Status:              %sSTARTING WORKER POD%s\n", colorGreen, colorReset)
	fmt.Printf("Kafka Brokers:       %s\n", brokers)
	fmt.Printf("Consumer Group:      %s\n", group)
	fmt.Printf("Subscribed Topic:    %s\n", topic)
	fmt.Printf("Delivery Standard:   At-Least-Once with Manual Offset Commits (CommitMessages)\n")
	fmt.Printf("Audit Standard:      Structured Real-Time Transaction Logging (Day 17 Lab L2)\n")
	fmt.Println("================================================================================")
	fmt.Println()
}
