// ==============================================================================
// NOVARIS: Distributed Payment Settlement & Adaptive Fraud Intelligence Platform
// UNIT TEST SUITE: Atomic Settlement Engine & Rollback Proofs (Day 6)
// FILE: services/gateway/settlement_test.go
// ==============================================================================

package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
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
