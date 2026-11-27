// ==============================================================================
// NOVARIS: Distributed Payment Settlement & Adaptive Fraud Intelligence Platform
// UNIT TEST SUITE: Atomic Settlement Engine & Rollback Proofs (Day 6)
// FILE: services/gateway/settlement_test.go
// ==============================================================================

package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/charan21042005/novaris/services/gateway/internal/outbox"
)

// TestSettleTransaction_Success verifies the happy path of settleTransaction:
// 1. Begins database transaction
// 2. Inserts 1 transaction row
// 3. Inserts 1 DEBIT journal entry
// 4. Inserts 1 CREDIT journal entry
// 5. Commits the transaction
// 6. Returns a non-empty, valid UUID
func TestSettleTransaction_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock instance: %v", err)
	}
	defer db.Close()

	senderID := "c0000000-0000-0000-0000-000000000001"
	recipientID := "c0000000-0000-0000-0000-000000000002"
	amountCents := int64(5000)

	// Set up expectations in strict chronological sequence:
	mock.ExpectBegin()

	// 1. Transaction insert
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO transactions
		 (id, idempotency_key, sender_account_id, recipient_account_id, amount_cents, status)
		 VALUES ($1, $2, $3, $4, $5, 'SETTLED')`)).
		WithArgs(
			sqlmock.AnyArg(), // transactionID
			sqlmock.AnyArg(), // idempotencyKey
			senderID,
			recipientID,
			amountCents,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// 2. Debit journal entry
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO journal_entries
		 (transaction_id, account_id, entry_type, amount_cents)
		 VALUES ($1, $2, 'DEBIT', $3)`)).
		WithArgs(
			sqlmock.AnyArg(), // transactionID
			senderID,
			amountCents,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// 3. Credit journal entry
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO journal_entries
		 (transaction_id, account_id, entry_type, amount_cents)
		 VALUES ($1, $2, 'CREDIT', $3)`)).
		WithArgs(
			sqlmock.AnyArg(), // transactionID
			recipientID,
			amountCents,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// 4. Commit transaction
	mock.ExpectCommit()

	// Execute settlement
	txID, err := settleTransaction(db, senderID, recipientID, amountCents)
	if err != nil {
		t.Fatalf("expected successful settlement, got error: %v", err)
	}

	// Verify returned transactionID is a valid UUID
	if _, err := uuid.Parse(txID); err != nil {
		t.Errorf("expected valid UUID for transactionID, got '%s'", txID)
	}

	// Verify all SQL expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled sqlmock expectations: %v", err)
	}
}

// TestSettleTransaction_RollbackOnFailure_CreditError proves that when an insert
// partway through fails (e.g. credit insert fails due to foreign key violation):
// 1. The function returns an error wrapped with context.
// 2. tx.Rollback() is automatically triggered by the deferred safety net.
// 3. tx.Commit() is NEVER called.
func TestSettleTransaction_RollbackOnFailure_CreditError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock instance: %v", err)
	}
	defer db.Close()

	senderID := "c0000000-0000-0000-0000-000000000001"
	invalidRecipientID := "00000000-0000-0000-0000-000000000000" // non-existent account
	amountCents := int64(10000)

	mock.ExpectBegin()

	// 1. Transaction insert succeeds
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO transactions
		 (id, idempotency_key, sender_account_id, recipient_account_id, amount_cents, status)
		 VALUES ($1, $2, $3, $4, $5, 'SETTLED')`)).
		WithArgs(
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			senderID,
			invalidRecipientID,
			amountCents,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// 2. Debit insert succeeds
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO journal_entries
		 (transaction_id, account_id, entry_type, amount_cents)
		 VALUES ($1, $2, 'DEBIT', $3)`)).
		WithArgs(
			sqlmock.AnyArg(),
			senderID,
			amountCents,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// 3. Credit insert FAILS (simulating PostgreSQL Foreign Key violation 23503)
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO journal_entries
		 (transaction_id, account_id, entry_type, amount_cents)
		 VALUES ($1, $2, 'CREDIT', $3)`)).
		WithArgs(
			sqlmock.AnyArg(),
			invalidRecipientID,
			amountCents,
		).
		WillReturnError(fmt.Errorf("pq: insert or update on table \"journal_entries\" violates foreign key constraint \"journal_entries_account_id_fkey\""))

	// 4. PROOF OF ROLLBACK: defer tx.Rollback() must fire!
	mock.ExpectRollback()

	// Execute settlement
	txID, err := settleTransaction(db, senderID, invalidRecipientID, amountCents)
	if err == nil {
		t.Fatalf("expected settlement to fail for invalid recipient, got nil error and txID=%s", txID)
	}

	// Verify error wrapping
	expectedErrMsg := "inserting credit journal entry"
	if !strings.Contains(err.Error(), expectedErrMsg) {
		t.Errorf("expected error containing %q, got: %v", expectedErrMsg, err)
	} else {
		t.Logf("Successfully captured wrapped rollback error: %v", err)
	}

	// Verify that tx.Rollback was executed and all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("rollback verification failed: %v", err)
	}
}

// TestSettleTransaction_NilDB verifies fail-safe behavior when the DB pool is uninitialized.
func TestSettleTransaction_NilDB(t *testing.T) {
	_, err := settleTransaction(nil, "acc1", "acc2", 100)
	if err == nil {
		t.Fatal("expected error when passing nil DB, got nil")
	}
}

// TestSettleTransaction_LiveDatabaseIntegration connects to a live PostgreSQL
// database if reachable, verifying true ACID rollback in a real engine.
func TestSettleTransaction_LiveDatabaseIntegration(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://novaris_admin:novaris_secure_password_2026@localhost:5432/novaris_db?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Skip("skipping live database test: unable to open postgres driver")
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		t.Skipf("skipping live database test: database not reachable at %s (%v)", dbURL, err)
	}

	// 1. Record count before failed operation
	var countBefore int
	if err := db.QueryRow("SELECT COUNT(*) FROM transactions").Scan(&countBefore); err != nil {
		t.Fatalf("failed to query transactions count before test: %v", err)
	}

	// 2. Trigger deliberate failure with non-existent recipient UUID
	nonExistentRecipient := "00000000-0000-0000-0000-000000000099"
	validSender := "c0000000-0000-0000-0000-000000000001"
	_, err = settleTransaction(db, validSender, nonExistentRecipient, 5000)
	if err == nil {
		t.Fatal("expected settlement error for non-existent recipient account, got nil")
	}

	// 3. Record count after failed operation
	var countAfter int
	if err := db.QueryRow("SELECT COUNT(*) FROM transactions").Scan(&countAfter); err != nil {
		t.Fatalf("failed to query transactions count after test: %v", err)
	}

	// 4. Verify ZERO rows left behind (true atomicity proof)
	if countBefore != countAfter {
		t.Errorf("ATOMICITY VIOLATION! Expected transactions count to remain %d, but became %d after failed call",
			countBefore, countAfter)
	} else {
		t.Logf("LIVE ACID PROOF: Transactions count strictly unchanged (%d == %d). Rollback verified!",
			countBefore, countAfter)
	}
}

// ==============================================================================
// DAY 7: TABLE-DRIVEN UNIT TESTING SUITE (COMPREHENSIVE SAFETY NET)
// ==============================================================================

// setupMockHappyPath registers expectations for a complete 3-insert settlement sequence.
func setupMockHappyPath(mock sqlmock.Sqlmock, senderID, recipientID string, amountCents int64) {
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO transactions
		 (id, idempotency_key, sender_account_id, recipient_account_id, amount_cents, status)
		 VALUES ($1, $2, $3, $4, $5, 'SETTLED')`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), senderID, recipientID, amountCents).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO journal_entries
		 (transaction_id, account_id, entry_type, amount_cents)
		 VALUES ($1, $2, 'DEBIT', $3)`)).
		WithArgs(sqlmock.AnyArg(), senderID, amountCents).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO journal_entries
		 (transaction_id, account_id, entry_type, amount_cents)
		 VALUES ($1, $2, 'CREDIT', $3)`)).
		WithArgs(sqlmock.AnyArg(), recipientID, amountCents).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()
}

// TestSettleTransaction_TableDriven implements an exhaustive, table-driven unit test suite
// covering equivalence classes, boundary value analysis, defensive pre-flight validation,
// and database-layer fault injection with automatic rollback verification.
func TestSettleTransaction_TableDriven(t *testing.T) {
	type testCase struct {
		name           string
		senderID       string
		recipientID    string
		amountCents    int64
		passNilDB      bool
		setupMock      func(mock sqlmock.Sqlmock, senderID, recipientID string, amountCents int64)
		expectSuccess  bool
		expectedErrSub string
	}

	tests := []testCase{
		// ----------------------------------------------------------------------
		// CLASS 1: Happy Paths (Equivalence Partitioning & Boundary Value Analysis)
		// ----------------------------------------------------------------------
		{
			name:          "HappyPath_StandardTransfer_FiftyDollars",
			senderID:      "c0000000-0000-0000-0000-000000000001",
			recipientID:   "c0000000-0000-0000-0000-000000000002",
			amountCents:   5000, // $50.00
			setupMock:     setupMockHappyPath,
			expectSuccess: true,
		},
		{
			name:          "HappyPath_BoundaryMinimum_OneCent",
			senderID:      "c0000000-0000-0000-0000-000000000001",
			recipientID:   "c0000000-0000-0000-0000-000000000002",
			amountCents:   1, // $0.01 lowest indivisible currency unit
			setupMock:     setupMockHappyPath,
			expectSuccess: true,
		},
		{
			name:          "HappyPath_LargeInstitutionalTransfer_TenMillionDollars",
			senderID:      "c0000000-0000-0000-0000-000000000001",
			recipientID:   "c0000000-0000-0000-0000-000000000002",
			amountCents:   1_000_000_000, // $10,000,000.00 (within int64 range)
			setupMock:     setupMockHappyPath,
			expectSuccess: true,
		},

		// ----------------------------------------------------------------------
		// CLASS 2: Defensive Pre-Flight Input Validation (Fail-Fast before DB)
		// ----------------------------------------------------------------------
		{
			name:           "ValidationError_ZeroAmount",
			senderID:       "c0000000-0000-0000-0000-000000000001",
			recipientID:    "c0000000-0000-0000-0000-000000000002",
			amountCents:    0,
			expectSuccess:  false,
			expectedErrSub: "amount must be strictly positive",
		},
		{
			name:           "ValidationError_NegativeAmount",
			senderID:       "c0000000-0000-0000-0000-000000000001",
			recipientID:    "c0000000-0000-0000-0000-000000000002",
			amountCents:    -5000,
			expectSuccess:  false,
			expectedErrSub: "amount must be strictly positive",
		},
		{
			name:           "ValidationError_EmptySenderID",
			senderID:       "",
			recipientID:    "c0000000-0000-0000-0000-000000000002",
			amountCents:    5000,
			expectSuccess:  false,
			expectedErrSub: "sender account ID cannot be empty",
		},
		{
			name:           "ValidationError_WhitespaceSenderID",
			senderID:       "    ",
			recipientID:    "c0000000-0000-0000-0000-000000000002",
			amountCents:    5000,
			expectSuccess:  false,
			expectedErrSub: "sender account ID cannot be empty",
		},
		{
			name:           "ValidationError_EmptyRecipientID",
			senderID:       "c0000000-0000-0000-0000-000000000001",
			recipientID:    "",
			amountCents:    5000,
			expectSuccess:  false,
			expectedErrSub: "recipient account ID cannot be empty",
		},
		{
			name:           "ValidationError_SelfTransferDisallowed",
			senderID:       "c0000000-0000-0000-0000-000000000001",
			recipientID:    "c0000000-0000-0000-0000-000000000001", // identical account
			amountCents:    5000,
			expectSuccess:  false,
			expectedErrSub: "self-transfers disallowed",
		},
		{
			name:           "ValidationError_NilDatabasePool",
			senderID:       "c0000000-0000-0000-0000-000000000001",
			recipientID:    "c0000000-0000-0000-0000-000000000002",
			amountCents:    5000,
			passNilDB:      true,
			expectSuccess:  false,
			expectedErrSub: "database connection pool is nil",
		},

		// ----------------------------------------------------------------------
		// CLASS 3: Database Fault Injection & Atomicity Rollback Proofs
		// ----------------------------------------------------------------------
		{
			name:        "DBFailure_BeginTransactionError",
			senderID:    "c0000000-0000-0000-0000-000000000001",
			recipientID: "c0000000-0000-0000-0000-000000000002",
			amountCents: 5000,
			setupMock: func(mock sqlmock.Sqlmock, s, r string, amt int64) {
				mock.ExpectBegin().WillReturnError(fmt.Errorf("connection pool exhausted: dial tcp timeout"))
			},
			expectSuccess:  false,
			expectedErrSub: "beginning database transaction",
		},
		{
			name:        "DBFailure_TransactionInsertFailed_RollbackFired",
			senderID:    "c0000000-0000-0000-0000-000000000001",
			recipientID: "c0000000-0000-0000-0000-000000000002",
			amountCents: 5000,
			setupMock: func(mock sqlmock.Sqlmock, s, r string, amt int64) {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO transactions`)).
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), s, r, amt).
					WillReturnError(fmt.Errorf("pq: unique constraint violation on idempotency_key"))
				mock.ExpectRollback()
			},
			expectSuccess:  false,
			expectedErrSub: "inserting transaction",
		},
		{
			name:        "DBFailure_DebitLegFailed_NonExistentSender_RollbackFired",
			senderID:    "c0000000-0000-0000-0000-000000000099",
			recipientID: "c0000000-0000-0000-0000-000000000002",
			amountCents: 5000,
			setupMock: func(mock sqlmock.Sqlmock, s, r string, amt int64) {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO transactions`)).
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), s, r, amt).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO journal_entries`)).
					WithArgs(sqlmock.AnyArg(), s, amt).
					WillReturnError(fmt.Errorf("pq: insert or update on table \"journal_entries\" violates foreign key constraint \"journal_entries_account_id_fkey\""))
				mock.ExpectRollback()
			},
			expectSuccess:  false,
			expectedErrSub: "inserting debit journal entry",
		},
		{
			name:        "DBFailure_CreditLegFailed_NonExistentRecipient_RollbackFired",
			senderID:    "c0000000-0000-0000-0000-000000000001",
			recipientID: "c0000000-0000-0000-0000-000000000099",
			amountCents: 5000,
			setupMock: func(mock sqlmock.Sqlmock, s, r string, amt int64) {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO transactions`)).
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), s, r, amt).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO journal_entries`)).
					WithArgs(sqlmock.AnyArg(), s, amt).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO journal_entries`)).
					WithArgs(sqlmock.AnyArg(), r, amt).
					WillReturnError(fmt.Errorf("pq: insert or update on table \"journal_entries\" violates foreign key constraint \"journal_entries_account_id_fkey\""))
				mock.ExpectRollback()
			},
			expectSuccess:  false,
			expectedErrSub: "inserting credit journal entry",
		},
		{
			name:        "DBFailure_CommitFailed_DiskFull",
			senderID:    "c0000000-0000-0000-0000-000000000001",
			recipientID: "c0000000-0000-0000-0000-000000000002",
			amountCents: 5000,
			setupMock: func(mock sqlmock.Sqlmock, s, r string, amt int64) {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO transactions`)).
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), s, r, amt).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO journal_entries`)).
					WithArgs(sqlmock.AnyArg(), s, amt).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO journal_entries`)).
					WithArgs(sqlmock.AnyArg(), r, amt).
					WillReturnResult(sqlmock.NewResult(1, 1))
				// Once Commit() is called, the driver marks the transaction done;
				// the deferred Rollback() safely returns sql.ErrTxDone without calling driver Rollback.
				mock.ExpectCommit().WillReturnError(fmt.Errorf("pq: disk full: cannot write WAL record"))
			},
			expectSuccess:  false,
			expectedErrSub: "committing transaction",
		},
	}

	for _, tc := range tests {
		tc := tc // Capture range variable for parallel execution safety
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var targetDB *sql.DB
			var mock sqlmock.Sqlmock
			var err error

			if !tc.passNilDB {
				targetDB, mock, err = sqlmock.New()
				if err != nil {
					t.Fatalf("failed to create sqlmock instance: %v", err)
				}
				defer targetDB.Close()

				if tc.setupMock != nil {
					tc.setupMock(mock, tc.senderID, tc.recipientID, tc.amountCents)
				}
			}

			// Execute settlement logic under test
			txID, err := settleTransaction(targetDB, tc.senderID, tc.recipientID, tc.amountCents)

			if tc.expectSuccess {
				if err != nil {
					t.Fatalf("expected settlement success, but encountered error: %v", err)
				}
				if txID == "" {
					t.Fatal("expected non-empty transactionID, got empty string")
				}
				if _, err := uuid.Parse(txID); err != nil {
					t.Fatalf("expected valid UUID for transactionID, got: %s", txID)
				}
			} else {
				if err == nil {
					t.Fatalf("expected settlement failure, but got success with txID: %s", txID)
				}
				if tc.expectedErrSub != "" && !strings.Contains(err.Error(), tc.expectedErrSub) {
					t.Errorf("expected error containing %q, got: %v", tc.expectedErrSub, err)
				}
			}

			// Verify that all mock expectations were met (including Rollback assertions)
			if mock != nil {
				if err := mock.ExpectationsWereMet(); err != nil {
					t.Errorf("unfulfilled sqlmock expectations: %v", err)
				}
			}
		})
	}
}

// BenchmarkSettleTransaction_Mock measures in-memory execution speed and allocations.
func BenchmarkSettleTransaction_Mock(b *testing.B) {
	senderID := "c0000000-0000-0000-0000-000000000001"
	recipientID := "c0000000-0000-0000-0000-000000000002"
	amountCents := int64(5000)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		db, mock, err := sqlmock.New()
		if err != nil {
			b.Fatalf("failed to initialize sqlmock for benchmark: %v", err)
		}
		setupMockHappyPath(mock, senderID, recipientID, amountCents)
		b.StartTimer()

		_, err = settleTransaction(db, senderID, recipientID, amountCents)
		if err != nil {
			b.Fatalf("benchmark iteration failed: %v", err)
		}

		b.StopTimer()
		db.Close()
		b.StartTimer()
	}
}

// TestSettleTransactionWithOutbox_Success verifies atomic dual-write enlistment of OutboxRecord (Day 18).
func TestSettleTransactionWithOutbox_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to initialize sqlmock: %v", err)
	}
	defer db.Close()

	senderID := "c0000000-0000-0000-0000-000000000001"
	recipientID := "c0000000-0000-0000-0000-000000000002"
	amountCents := int64(7500)

	setupMockHappyPath(mock, senderID, recipientID, amountCents)
	outboxRepo := outbox.NewMockRepository(5)

	eventData := map[string]interface{}{
		"amount_cents": amountCents,
		"currency":     "USD",
		"payer_id":     senderID,
	}

	txID, rec, err := SettleTransactionWithOutbox(
		context.Background(),
		db,
		outboxRepo,
		senderID,
		recipientID,
		amountCents,
		eventData,
	)
	if err != nil {
		t.Fatalf("expected successful outbox settlement, got: %v", err)
	}

	if txID == "" {
		t.Errorf("expected non-empty transaction ID")
	}
	if rec == nil {
		t.Fatalf("expected non-nil OutboxRecord")
	}
	if rec.AggregateType != outbox.AggregateTransaction {
		t.Errorf("expected aggregate type TRANSACTION, got %s", rec.AggregateType)
	}
	if rec.AggregateID != senderID {
		t.Errorf("expected partition key %s, got %s", senderID, rec.AggregateID)
	}
	if rec.Status != outbox.StatusPending {
		t.Errorf("expected status PENDING, got %s", rec.Status)
	}

	// Verify outbox repo has 1 pending record
	pending, _ := outboxRepo.CountPending(context.Background())
	if pending != 1 {
		t.Errorf("expected 1 pending record in outbox repo, got %d", pending)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled sql expectations: %v", err)
	}
}

// TestSettleTransactionWithOutbox_RollbackOnOutboxFailure asserts that if outbox insertion
// fails, the entire database transaction rolls back, guaranteeing zero partial/phantom state!
func TestSettleTransactionWithOutbox_RollbackOnOutboxFailure(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to initialize sqlmock: %v", err)
	}
	defer db.Close()

	senderID := "c0000000-0000-0000-0000-000000000001"
	recipientID := "c0000000-0000-0000-0000-000000000002"
	amountCents := int64(5000)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO transactions")).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), senderID, recipientID, amountCents).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO journal_entries")).
		WithArgs(sqlmock.AnyArg(), senderID, amountCents).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO journal_entries")).
		WithArgs(sqlmock.AnyArg(), recipientID, amountCents).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectRollback()

	outboxRepo := outbox.NewMockRepository(5)
	outboxRepo.InsertError = errors.New("simulated outbox disk error")

	txID, rec, err := SettleTransactionWithOutbox(
		context.Background(),
		db,
		outboxRepo,
		senderID,
		recipientID,
		amountCents,
		map[string]interface{}{"amount": 5000},
	)
	if err == nil {
		t.Fatalf("expected error when outbox insert fails, got nil")
	}
	if txID != "" || rec != nil {
		t.Errorf("expected empty txID and nil rec on failure")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled sql expectations: %v", err)
	}
}
