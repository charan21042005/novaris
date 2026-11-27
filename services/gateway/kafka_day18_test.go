// ==============================================================================
// NOVARIS: Distributed Payment Settlement & Adaptive Fraud Intelligence Platform
// INTEGRATION TEST SUITE: Day 18 Transactional Outbox & Crash Recovery Master Suite
// FILE: services/gateway/kafka_day18_test.go
// ==============================================================================
//
// Verification Scope (Day 18 Curriculum Standards):
//   1. Subtest 1: Dual-Write Atomicity (Rollback vs. Commit Isolation)
//   2. Subtest 2: Outbox Relay Batch Polling & Kafka Partition Key Fidelity
//   3. Subtest 3: Worker Crash Simulation & Zero-Loss Recovery (Lab L3)
//   4. Subtest 4: Full End-to-End Pipeline (Gateway Outbox -> Relay -> Settlement Consumer)
//   5. Subtest 5: Concurrency Safety & SKIP LOCKED Multi-Worker Emulation
// ==============================================================================

package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"

	"github.com/charan21042005/novaris/services/gateway/internal/outbox"
	"github.com/charan21042005/novaris/services/gateway/internal/streaming"
)

// TestDay18_OutboxMasterSuite runs the complete Day 18 Transactional Outbox verification.
func TestDay18_OutboxMasterSuite(t *testing.T) {
	t.Run("Subtest1_DualWriteAtomicity_RollbackIsolation", func(t *testing.T) {
		repo := outbox.NewMockRepository(5)
		ctx := context.Background()

		// Simulate transaction rollback: If the business transaction fails,
		// NO outbox event should ever be committed to the database.
		outboxRec, err := outbox.NewOutboxRecord(
			outbox.AggregateTransaction,
			"payer_rollback_test",
			"novaris.tx.initiated",
			map[string]interface{}{"amount": 5000},
		)
		if err != nil {
			t.Fatalf("failed creating record: %v", err)
		}

		// Inject simulated SQL error during settlement:
		repo.InsertError = errors.New("simulated foreign key violation")
		err = repo.InsertTx(ctx, &sql.Tx{}, outboxRec)
		if err == nil {
			t.Fatalf("expected error on failed insert")
		}

		// Assert zero pending records
		pending, _ := repo.CountPending(ctx)
		if pending != 0 {
			t.Fatalf("expected 0 pending records after rollback, got %d", pending)
		}

		// Now simulate successful commit
		repo.InsertError = nil
		err = repo.InsertTx(ctx, &sql.Tx{}, outboxRec)
		if err != nil {
			t.Fatalf("expected successful insert on commit: %v", err)
		}

		pending, _ = repo.CountPending(ctx)
		if pending != 1 {
			t.Fatalf("expected 1 pending record after commit, got %d", pending)
		}
	})

	t.Run("Subtest2_OutboxRelay_BatchSweepAndPartitionKeyFidelity", func(t *testing.T) {
		repo := outbox.NewMockRepository(5)
		publisher := outbox.NewMockPublisher()
		cfg := outbox.DefaultRelayConfig()
		cfg.BatchSize = 10

		relay, err := outbox.NewOutboxRelay(repo, publisher, cfg)
		if err != nil {
			t.Fatalf("failed creating relay: %v", err)
		}

		// Seed 5 records with specific payer UUIDs
		payerUUID := "c0000000-0000-0000-0000-000000000001"
		for i := 1; i <= 5; i++ {
			rec, _ := outbox.NewOutboxRecord(outbox.AggregateTransaction, payerUUID, "novaris.tx.initiated", map[string]interface{}{
				"seq":          i,
				"amount_cents": i * 1000,
			})
			_ = repo.InsertDirect(rec)
		}

		ctx := context.Background()
		processed, err := relay.ProcessBatch(ctx)
		if err != nil {
			t.Fatalf("unexpected error processing batch: %v", err)
		}
		if processed != 5 {
			t.Fatalf("expected 5 records processed, got %d", processed)
		}

		// Verify partition key fidelity in Kafka publisher
		messages := publisher.GetMessages()
		if len(messages) != 5 {
			t.Fatalf("expected 5 messages published, got %d", len(messages))
		}

		for idx, msg := range messages {
			if msg.Key != payerUUID {
				t.Errorf("msg[%d]: expected partition key %s, got %s", idx, payerUUID, msg.Key)
			}
			if msg.Topic != "novaris.tx.initiated" {
				t.Errorf("msg[%d]: expected topic novaris.tx.initiated, got %s", idx, msg.Topic)
			}
		}

		// Assert all records in repo are now PUBLISHED
		pending, _ := repo.CountPending(ctx)
		if pending != 0 {
			t.Fatalf("expected 0 pending records after relay sweep, got %d", pending)
		}
		published, _ := repo.CountPublished(ctx)
		if published != 5 {
			t.Fatalf("expected 5 published records, got %d", published)
		}
	})

	t.Run("Subtest3_WorkerCrashSimulation_ZeroLossRecovery", func(t *testing.T) {
		// Lab L3 Verification:
		// Proves that when a relay worker crashes immediately post-publish,
		// the unacknowledged record is safely recovered and re-emitted by the replacement worker.
		repo := outbox.NewMockRepository(5)
		publisher := outbox.NewMockPublisher()
		cfg := outbox.DefaultRelayConfig()

		var recordIDs []string
		for i := 1; i <= 3; i++ {
			rec, _ := outbox.NewOutboxRecord(outbox.AggregateTransaction, "payer_crash_suite", "novaris.tx.initiated", map[string]interface{}{
				"tx_id":        uuid.New().String(),
				"amount_cents": i * 5000,
			})
			_ = repo.InsertDirect(rec)
			recordIDs = append(recordIDs, rec.ID)
		}

		ctx := context.Background()

		// Worker 1 crashes on Record 2
		worker1, _ := outbox.NewOutboxRelay(repo, publisher, cfg)
		crashErr := errors.New("FAULT: worker process killed mid-batch")
		worker1.SetPostPublishCrashHook(func(rec *outbox.OutboxRecord) error {
			if rec.ID == recordIDs[1] {
				return crashErr
			}
			return nil
		})

		_, _ = worker1.ProcessBatch(ctx)

		// Assert post-crash state: Record 2 is published to Kafka, but STILL PENDING in DB
		rec2, _ := repo.GetRecord(recordIDs[1])
		if rec2.Status != outbox.StatusPending {
			t.Fatalf("expected record 2 to be PENDING in DB after crash, got %s", rec2.Status)
		}

		// Worker 2 reboots and sweeps
		worker2, _ := outbox.NewOutboxRelay(repo, publisher, cfg)
		recovered, err := worker2.ProcessBatch(ctx)
		if err != nil {
			t.Fatalf("error in worker 2 recovery: %v", err)
		}
		if recovered != 2 {
			t.Fatalf("expected worker 2 to recover 2 records, got %d", recovered)
		}

		// Proves zero data loss: All 3 records are now PUBLISHED
		for idx, id := range recordIDs {
			r, _ := repo.GetRecord(id)
			if r.Status != outbox.StatusPublished {
				t.Fatalf("record[%d] expected PUBLISHED, got %s", idx, r.Status)
			}
		}

		// Total messages published across crash: 4 (Record 1 once, Record 2 twice, Record 3 once)
		if publisher.MessageCount() != 4 {
			t.Fatalf("expected 4 total published events, got %d", publisher.MessageCount())
		}
	})

	t.Run("Subtest4_FullPipeline_GatewayOutboxToRelayToConsumer", func(t *testing.T) {
		// Tests the full chain:
		// SettleTransactionWithOutbox -> OutboxRelay.ProcessBatch -> Consumer Deduplication
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("failed creating sqlmock: %v", err)
		}
		defer db.Close()

		senderID := "c0000000-0000-0000-0000-000000000001"
		recipientID := "c0000000-0000-0000-0000-000000000002"
		amountCents := int64(12500)

		setupMockHappyPath(mock, senderID, recipientID, amountCents)

		outboxRepo := outbox.NewMockRepository(5)
		publisher := outbox.NewMockPublisher()

		eventPayload := &streaming.TransactionInitiatedEvent{
			EventID:        uuid.New().String(),
			EventType:      "novaris.tx.initiated",
			TransactionID:  "",
			PayerID:        senderID,
			PayeeID:        recipientID,
			AmountCents:    amountCents,
			Currency:       "USD",
			Timestamp:      time.Now().UTC(),
		}

		// 1. Gateway enlists event in DB transaction
		txID, outboxRec, err := SettleTransactionWithOutbox(
			context.Background(),
			db,
			outboxRepo,
			senderID,
			recipientID,
			amountCents,
			eventPayload,
		)
		if err != nil {
			t.Fatalf("failed settling with outbox: %v", err)
		}
		if txID == "" || outboxRec == nil {
			t.Fatalf("invalid settlement result")
		}

		// 2. Outbox Relay sweeps and emits to Kafka
		relay, _ := outbox.NewOutboxRelay(outboxRepo, publisher, outbox.DefaultRelayConfig())
		processed, err := relay.ProcessBatch(context.Background())
		if err != nil {
			t.Fatalf("relay failed: %v", err)
		}
		if processed != 1 {
			t.Fatalf("expected 1 record relayed, got %d", processed)
		}

		// 3. Verify downstream consumer receives payload
		messages := publisher.GetMessages()
		if len(messages) != 1 {
			t.Fatalf("expected 1 message in publisher, got %d", len(messages))
		}

		var consumedEvent streaming.TransactionInitiatedEvent
		if err := json.Unmarshal(messages[0].Payload, &consumedEvent); err != nil {
			t.Fatalf("failed decoding relayed payload: %v", err)
		}

		if consumedEvent.PayerID != senderID {
			t.Errorf("expected payer ID %s, got %s", senderID, consumedEvent.PayerID)
		}
		if consumedEvent.AmountCents != amountCents {
			t.Errorf("expected amount %d, got %d", amountCents, consumedEvent.AmountCents)
		}
	})

	t.Run("Subtest5_ConcurrencySafety_MultiWorkerNoDuplicateClaim", func(t *testing.T) {
		repo := outbox.NewMockRepository(5)
		publisher := outbox.NewMockPublisher()
		cfg := outbox.DefaultRelayConfig()
		cfg.BatchSize = 1

		// Seed 10 records
		for i := 1; i <= 10; i++ {
			rec, _ := outbox.NewOutboxRecord(outbox.AggregateTransaction, "payer_concurrent", "novaris.tx.initiated", map[string]interface{}{
				"id": i,
			})
			_ = repo.InsertDirect(rec)
		}

		// Spawn 3 concurrent relay workers
		var wg sync.WaitGroup
		ctx := context.Background()

		for w := 0; w < 3; w++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				relay, _ := outbox.NewOutboxRelay(repo, publisher, cfg)
				for {
					count, _ := relay.ProcessBatch(ctx)
					if count == 0 {
						break
					}
				}
			}()
		}

		wg.Wait()

		// All 10 records must be published
		published, _ := repo.CountPublished(ctx)
		if published != 10 {
			t.Fatalf("expected 10 published records, got %d", published)
		}
		pending, _ := repo.CountPending(ctx)
		if pending != 0 {
			t.Fatalf("expected 0 pending records, got %d", pending)
		}
	})
}

// TestDay18_LiveClusterIntegration runs when live PostgreSQL and Kafka are available.
func TestDay18_LiveClusterIntegration(t *testing.T) {
	conn, err := net.DialTimeout("tcp", "localhost:9092", 1*time.Second)
	if err != nil {
		t.Skipf("skipping live cluster test: Kafka broker unreachable at localhost:9092 (%v)", err)
		return
	}
	_ = conn.Close()

	db, err := sql.Open("postgres", "postgres://novaris_admin:novaris_secure_password_2026@localhost:5432/novaris_db?sslmode=disable")
	if err != nil {
		t.Skipf("skipping live cluster test: PostgreSQL unreachable: %v", err)
		return
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		t.Skipf("skipping live cluster test: PostgreSQL ping failed: %v", err)
		return
	}

	sqlRepo := outbox.NewSQLRepository(db, 5)
	kafkaCfg := streaming.LoadConfigFromEnv()
	liveProd, err := streaming.NewProducer(&kafkaCfg)
	if err != nil {
		t.Skipf("skipping live cluster test: Kafka producer failed: %v", err)
		return
	}
	defer liveProd.Close()

	adapter, _ := outbox.NewKafkaAdapter(liveProd)
	relay, err := outbox.NewOutboxRelay(sqlRepo, adapter, outbox.DefaultRelayConfig())
	if err != nil {
		t.Fatalf("failed creating outbox relay: %v", err)
	}

	// Process batch against live cluster
	processed, err := relay.ProcessBatch(ctx)
	if err != nil {
		t.Fatalf("live relay batch failed: %v", err)
	}
	t.Logf("Live cluster outbox relay processed %d pending records", processed)
}
