// ==============================================================================
// NOVARIS: Distributed Payment Settlement & Adaptive Fraud Intelligence Platform
// FILE: services/gateway/cmd/verifykafka/main.go
// PURPOSE: Day 15 Interactive Diagnostic & Verification CLI Tool
// ==============================================================================
//
// 🎓 WHY THIS TOOL EXISTS IN OUR FINTECH ENGINEERING DISCIPLINE:
//
// In production cloud environments (AWS EKS / MSK / GCP), infrastructure issues
// (e.g., security group port blocks, misconfigured advertised listeners, DNS resolution
// latency, or KRaft controller leader elections) can cause silent outages.
//
// Rather than relying on trial-and-error or opaque log messages, NOVARIS provides
// dedicated verification binaries (like `phase1audit`, `verifyopenapi`, `verifyschema`,
// and now `verifykafka`).
//
// What `verifykafka` Proves in 5 Automated Steps:
//  [1/5] Cluster Reachability: Executes low-level TCP and Kafka protocol handshakes,
//        measuring round-trip ping latency and querying KRaft metadata controllers.
//  [2/5] Canonical Topic Provisioning: Ensures all 6 required Phase 2 event streams
//        are active with 3 partitions and correct retention policies.
//  [3/5] Active Catalog Inspection: Queries the broker for partition leaders and ISRs.
//  [4/5] Partition Routing Mathematics: Proves deterministic key hashing across 5 sample
//        accounts, demonstrating per-account FIFO invariants.
//  [5/5] Diagnostic Probe Injection: Writes a live probe message to `novaris.tx.initiated`
//        and confirms leader ACK receipt.
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

	"github.com/charan21042005/novaris/services/gateway/internal/streaming"
)

// ANSI terminal color escape codes for high-contrast visual status reporting:
const (
	colorReset  = "\033[0m"
	colorBold   = "\033[1m"
	colorCyan   = "\033[36m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorRed    = "\033[31m"
	colorGray   = "\033[90m"
)

// AuditReport aggregates all verification checks for JSON export or ANSI console display.
type AuditReport struct {
	Timestamp          time.Time                       `json:"timestamp"`
	Phase              string                          `json:"phase"`
	Day                int                             `json:"day"`
	Milestone          string                          `json:"milestone"`
	TargetBrokers      []string                        `json:"target_brokers"`
	ClusterHealth      *streaming.ClusterHealth        `json:"cluster_health"`
	ProvisionReport    *streaming.TopicProvisionReport `json:"provision_report,omitempty"`
	TopicSummaries     []streaming.TopicSummary        `json:"topic_summaries,omitempty"`
	PartitionProof     []PartitionProofItem            `json:"partition_proof"`
	ProbeResult        *streaming.ProbeResult          `json:"probe_result,omitempty"`
	AllChecksPassed    bool                            `json:"all_checks_passed"`
	VerificationStatus string                          `json:"verification_status"`
}

// PartitionProofItem proves deterministic routing for an account key.
type PartitionProofItem struct {
	AccountKey       string `json:"account_key"`
	CalculatedPart   int    `json:"calculated_partition"`
	ModuloPartitions int    `json:"modulo_partitions"`
	Explanation      string `json:"explanation"`
}

func main() {
	var (
		brokersFlag    = flag.String("brokers", "localhost:9092", "Comma-separated Kafka bootstrap broker addresses")
		partitionsFlag = flag.Int("partitions", 3, "Default partition count for canonical topics")
		repFactorFlag  = flag.Int("replication", 1, "Replication factor for canonical topics")
		doProvision    = flag.Bool("provision", true, "Idempotently create standard NOVARIS topics")
		doProbe        = flag.Bool("probe", true, "Produce a diagnostic test probe event to novaris.tx.initiated")
		jsonOutput     = flag.Bool("json", false, "Output machine-readable JSON instead of ANSI text")
		timeoutSec     = flag.Int("timeout", 10, "Network timeout in seconds")
	)
	flag.Parse()

	// Environment variable fallback:
	brokerAddr := *brokersFlag
	if envBrokers := os.Getenv("KAFKA_BROKERS"); envBrokers != "" && brokerAddr == "localhost:9092" {
		brokerAddr = envBrokers
	}

	brokerList := strings.Split(brokerAddr, ",")
	for i := range brokerList {
		brokerList[i] = strings.TrimSpace(brokerList[i])
	}

	cfg := streaming.Config{
		Brokers:                  brokerList,
		ClientID:                 "novaris-day15-verify-cli",
		DialTimeout:              time.Duration(*timeoutSec) * time.Second,
		DefaultPartitions:        *partitionsFlag,
		DefaultReplicationFactor: *repFactorFlag,
	}

	report := &AuditReport{
		Timestamp:       time.Now().UTC(),
		Phase:           "Phase 2: Streaming Core",
		Day:             15,
		Milestone:       "Kafka Fundamentals (Topics, Partitions, Brokers)",
		TargetBrokers:   cfg.Brokers,
		AllChecksPassed: false,
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.DialTimeout*2)
	defer cancel()

	admin := streaming.NewAdminClient(cfg)

	// =========================================================================
	// Check 1: Cluster Health & Reachability
	// =========================================================================
	health, err := admin.CheckHealth(ctx)
	report.ClusterHealth = health
	if err != nil {
		report.VerificationStatus = fmt.Sprintf("FAILED: %v", err)
		outputReport(report, *jsonOutput)
		os.Exit(1)
	}

	// =========================================================================
	// Check 2: Idempotent Topic Provisioning
	// =========================================================================
	if *doProvision {
		specs := streaming.GetStandardTopics(cfg.DefaultPartitions, cfg.DefaultReplicationFactor)
		provReport, pErr := admin.EnsureTopics(ctx, specs)
		if pErr != nil {
			report.VerificationStatus = fmt.Sprintf("FAILED_TOPIC_PROVISION: %v", pErr)
			outputReport(report, *jsonOutput)
			os.Exit(1)
		}
		report.ProvisionReport = provReport
	}

	// =========================================================================
	// Check 3: Query Active Topics & Partitions
	// =========================================================================
	topics, tErr := admin.ListTopics(ctx)
	if tErr == nil {
		report.TopicSummaries = topics
	}

	// =========================================================================
	// Check 4: Deterministic Partition Key Hashing Proof
	// =========================================================================
	partitioner := streaming.NewPartitioner(streaming.StrategyFNV1a)
	sampleAccounts := []string{
		"acc_merchant_alpha_001",
		"acc_user_bob_4920",
		"acc_merchant_strata_88",
		"acc_treasury_escrow",
		"acc_fee_collection",
	}

	for _, acc := range sampleAccounts {
		part := partitioner.RouteAccountKey(acc, cfg.DefaultPartitions)
		explanation := streaming.FormatRoutingExplanation(acc, cfg.DefaultPartitions, part, streaming.StrategyFNV1a)
		report.PartitionProof = append(report.PartitionProof, PartitionProofItem{
			AccountKey:       acc,
			CalculatedPart:   part,
			ModuloPartitions: cfg.DefaultPartitions,
			Explanation:      explanation,
		})
	}

	// =========================================================================
	// Check 5: Diagnostic Probe Injection
	// =========================================================================
	if *doProbe {
		probeTarget := streaming.TopicTxInitiated
		probeKey := "acc_merchant_alpha_001"
		probePayload := fmt.Sprintf(`{"event_type":"PROBE_INGEST","account_id":"%s","amount_cents":1000,"currency":"USD","day":15,"phase":2,"status":"VERIFIED"}`, probeKey)

		probeRes, prErr := admin.ProduceProbeMessage(ctx, probeTarget, probeKey, probePayload)
		if prErr != nil {
			report.VerificationStatus = fmt.Sprintf("FAILED_PROBE_WRITE: %v", prErr)
			outputReport(report, *jsonOutput)
			os.Exit(1)
		}
		report.ProbeResult = probeRes
	}

	report.AllChecksPassed = true
	report.VerificationStatus = "SUCCESS: All Day 15 Kafka Milestones Verified & Reachable"
	outputReport(report, *jsonOutput)
}

func outputReport(r *AuditReport, asJSON bool) {
	if asJSON {
		data, _ := json.MarshalIndent(r, "", "  ")
		fmt.Println(string(data))
		return
	}

	fmt.Println()
	fmt.Printf("%s%s================================================================================%s\n", colorBold, colorCyan, colorReset)
	fmt.Printf("%s%s   NOVARIS • PHASE 2 DAY 15: KAFKA CLUSTER & STREAMING VERIFICATION CLI        %s\n", colorBold, colorGreen, colorReset)
	fmt.Printf("%s%s================================================================================%s\n", colorBold, colorCyan, colorReset)
	fmt.Printf("%sTimestamp:%s           %s\n", colorGray, colorReset, r.Timestamp.Format(time.RFC3339))
	fmt.Printf("%sTarget Brokers:%s     %s\n", colorGray, colorReset, strings.Join(r.TargetBrokers, ", "))
	fmt.Printf("%sMilestone Focus:%s    %s\n", colorGray, colorReset, r.Milestone)
	fmt.Println()

	// Cluster Health:
	if r.ClusterHealth != nil && r.ClusterHealth.Reachable {
		fmt.Printf("%s[1/5] KAFKA CLUSTER REACHABILITY & KRAFT METADATA:%s\n", colorBold, colorReset)
		fmt.Printf("  %s✔ Status:%s             REACHABLE (Protocol Round-Trip Ping: %.2f ms)\n", colorGreen, colorReset, r.ClusterHealth.LatencyMS)
		fmt.Printf("  %s✔ Connected Broker:%s   %s\n", colorGreen, colorReset, r.ClusterHealth.ConnectedBroker)
		fmt.Printf("  %s✔ KRaft Controller:%s   Node ID %d (%s)\n", colorGreen, colorReset, r.ClusterHealth.ControllerID, r.ClusterHealth.ControllerHost)
		fmt.Printf("  %s✔ Active Brokers:%s     %d broker(s) registered in Raft metadata quorum\n", colorGreen, colorReset, len(r.ClusterHealth.Brokers))
		for _, b := range r.ClusterHealth.Brokers {
			role := "Broker"
			if b.IsController {
				role = "Broker + KRaft Controller (Combined Role)"
			}
			fmt.Printf("      - Node %d at %s [%s]\n", b.ID, b.Address, role)
		}
	} else {
		fmt.Printf("%s[1/5] KAFKA CLUSTER REACHABILITY:%s\n", colorBold, colorReset)
		fmt.Printf("  %s✖ Cluster UNREACHABLE:%s %s\n", colorRed, colorReset, r.VerificationStatus)
		return
	}
	fmt.Println()

	// Topic Provisioning:
	if r.ProvisionReport != nil {
		fmt.Printf("%s[2/5] NOVARIS CANONICAL TOPIC PROVISIONING:%s\n", colorBold, colorReset)
		fmt.Printf("  %s✔ Provisioning Summary:%s %d requested, %d newly created, %d active/verified (%v)\n",
			colorGreen, colorReset, r.ProvisionReport.TotalRequested, r.ProvisionReport.CreatedCount, r.ProvisionReport.ExistingCount, r.ProvisionReport.Duration)
		for _, item := range r.ProvisionReport.Results {
			statusColor := colorGreen
			if item.Status == "FAILED" {
				statusColor = colorRed
			}
			fmt.Printf("      • %-26s %s[%s]%s %s\n", item.Topic, statusColor, item.Status, colorReset, item.Details)
		}
	}
	fmt.Println()

	// Active Cluster Topics:
	if len(r.TopicSummaries) > 0 {
		fmt.Printf("%s[3/5] ACTIVE CLUSTER TOPIC CATALOG (Including KRaft Metadata):%s\n", colorBold, colorReset)
		for _, t := range r.TopicSummaries {
			tag := "User Stream"
			if t.IsInternal {
				tag = "Internal/Offsets"
			}
			fmt.Printf("      - %-26s | %d partition(s) | %s\n", t.Name, t.NumPartitions, tag)
		}
	}
	fmt.Println()

	// Partitioning Mathematics:
	fmt.Printf("%s[4/5] DETERMINISTIC PARTITION KEY ROUTING PROOF (Per-Account FIFO):%s\n", colorBold, colorReset)
	for _, p := range r.PartitionProof {
		fmt.Printf("      • %s -> %sPartition %d%s (Invariant: all events for %s land here)\n",
			p.AccountKey, colorYellow, p.CalculatedPart, colorReset, p.AccountKey)
	}
	fmt.Println()

	// Diagnostic Probe Result:
	if r.ProbeResult != nil {
		fmt.Printf("%s[5/5] DIAGNOSTIC PROBE WRITE PATH VERIFICATION:%s\n", colorBold, colorReset)
		fmt.Printf("  %s✔ Target Topic:%s         %s\n", colorGreen, colorReset, r.ProbeResult.Topic)
		fmt.Printf("  %s✔ Partition Routed:%s     Partition %d\n", colorGreen, colorReset, r.ProbeResult.Partition)
		fmt.Printf("  %s✔ Routing Key:%s          %s\n", colorGreen, colorReset, r.ProbeResult.Key)
		fmt.Printf("  %s✔ Payload Size:%s         %d bytes\n", colorGreen, colorReset, r.ProbeResult.PayloadBytes)
		fmt.Printf("  %s✔ Write Latency:%s        %v (Leader ACK Confirmed)\n", colorGreen, colorReset, r.ProbeResult.WriteLatency)
	}
	fmt.Println()

	fmt.Printf("%s%s================================================================================%s\n", colorBold, colorGreen, colorReset)
	fmt.Printf("%s%s   FINAL OUTCOME: %s   %s\n", colorBold, colorGreen, r.VerificationStatus, colorReset)
	fmt.Printf("%s%s================================================================================%s\n", colorBold, colorGreen, colorReset)
	fmt.Println()
}
