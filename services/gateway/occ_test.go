// ==============================================================================
// NOVARIS: Distributed Payment Settlement & Adaptive Fraud Intelligence Platform
// UNIT & CONCURRENCY TEST SUITE: Optimistic Concurrency Control (OCC) Engine (Day 9)
// FILE: services/gateway/occ_test.go
// ==============================================================================

package main

import (
	"context"
	"errors"
	"regexp"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

// TestOCC_SingleThreadedSuccessfulUpdate verifies that standard version check-and-increment
// succeeds sequentially from version 0 to 1 to 2.
func TestOCC_SingleThreadedSuccessfulUpdate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to initialize sqlmock: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	accountID := uuid.New().String()

	// Expectation: version 0 -> 1 (+1000 cents)
	mock.ExpectExec(regexp.QuoteMeta(`
		UPDATE accounts
		SET balance_cents = balance_cents + $1,
		    version = version + 1,
		    updated_at = NOW()
		WHERE id = $2
		  AND version = $3;
	`)).
		WithArgs(int64(1000), accountID, 0).
		WillReturnResult(sqlmock.NewResult(1, 1))

	newVer, err := UpdateAccountBalanceOCC(ctx, db, accountID, 0, 1000)
	if err != nil {
		t.Fatalf("expected successful OCC update, got err: %v", err)
	}
	if newVer != 1 {
		t.Errorf("expected new version 1, got %d", newVer)
	}

	// Expectation: version 1 -> 2 (+500 cents)
	mock.ExpectExec(regexp.QuoteMeta(`
		UPDATE accounts
		SET balance_cents = balance_cents + $1,
		    version = version + 1,
		    updated_at = NOW()
		WHERE id = $2
		  AND version = $3;
	`)).
		WithArgs(int64(500), accountID, 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	newVer, err = UpdateAccountBalanceOCC(ctx, db, accountID, 1, 500)
	if err != nil {
		t.Fatalf("expected second OCC update to succeed, got: %v", err)
	}
	if newVer != 2 {
		t.Errorf("expected new version 2, got %d", newVer)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled mock expectations: %v", err)
	}
}

// TestOCC_StaleVersionConflictRejection verifies that when an update passes a stale expected version,
// the SQL query affects 0 rows, and UpdateAccountBalanceOCC safely rejects the write with ErrOptimisticLockConflict.
func TestOCC_StaleVersionConflictRejection(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to initialize sqlmock: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	accountID := uuid.New().String()

	// Simulate that the row in Postgres already has version 1, but the caller passes expected version 0.
	// Postgres WHERE version = 0 matches 0 rows!
	mock.ExpectExec(regexp.QuoteMeta(`
		UPDATE accounts
		SET balance_cents = balance_cents + $1,
		    version = version + 1,
		    updated_at = NOW()
		WHERE id = $2
		  AND version = $3;
	`)).
		WithArgs(int64(2000), accountID, 0).
		WillReturnResult(sqlmock.NewResult(0, 0)) // 0 rows affected!

	newVer, err := UpdateAccountBalanceOCC(ctx, db, accountID, 0, 2000)
	if err == nil {
		t.Fatal("expected ErrOptimisticLockConflict, but got nil error")
	}
	if !errors.Is(err, ErrOptimisticLockConflict) {
		t.Errorf("expected error to be ErrOptimisticLockConflict, got: %v", err)
	}
	if newVer != 0 {
		t.Errorf("expected return version 0 on failure, got %d", newVer)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled mock expectations: %v", err)
	}
}

// TestOCC_SettleTransactionOCC_HappyPath tests an end-to-end atomic OCC double-entry settlement.
func TestOCC_SettleTransactionOCC_HappyPath(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to initialize sqlmock: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	senderID := uuid.New().String()
	recipientID := uuid.New().String()
	amountCents := int64(3000)

	senderParams := OCCDebitParams{
		AccountID:       senderID,
		ExpectedVersion: 2,
		AmountCents:     amountCents,
	}
	recipientParams := OCCCreditParams{
		AccountID:       recipientID,
		ExpectedVersion: 5,
		AmountCents:     amountCents,
	}

	// Chronological expectation sequence:
	mock.ExpectBegin()

	// 1. Debit leg (sender)
	mock.ExpectExec(regexp.QuoteMeta(`
		UPDATE accounts
		SET balance_cents = balance_cents - $1,
		    version = version + 1,
		    updated_at = NOW()
		WHERE id = $2
		  AND version = $3
		  AND balance_cents >= $1;
	`)).
		WithArgs(amountCents, senderID, 2).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// 2. Credit leg (recipient)
	mock.ExpectExec(regexp.QuoteMeta(`
		UPDATE accounts
		SET balance_cents = balance_cents + $1,
		    version = version + 1,
		    updated_at = NOW()
		WHERE id = $2
		  AND version = $3;
	`)).
		WithArgs(amountCents, recipientID, 5).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// 3. Insert transaction
	mock.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO transactions
		 (id, idempotency_key, sender_account_id, recipient_account_id, amount_cents, status)
		 VALUES ($1, $2, $3, $4, $5, 'SETTLED');
	`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), senderID, recipientID, amountCents).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// 4. Debit journal entry
	mock.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO journal_entries
		 (transaction_id, account_id, entry_type, amount_cents)
		 VALUES ($1, $2, 'DEBIT', $3);
	`)).
		WithArgs(sqlmock.AnyArg(), senderID, amountCents).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// 5. Credit journal entry
	mock.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO journal_entries
		 (transaction_id, account_id, entry_type, amount_cents)
		 VALUES ($1, $2, 'CREDIT', $3);
	`)).
		WithArgs(sqlmock.AnyArg(), recipientID, amountCents).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	result, err := SettleTransactionOCC(ctx, db, senderParams, recipientParams)
	if err != nil {
		t.Fatalf("expected successful OCC settlement, got: %v", err)
	}

	if result.SenderNewVersion != 3 {
		t.Errorf("expected sender new version 3, got %d", result.SenderNewVersion)
	}
	if result.RecipientNewVersion != 6 {
		t.Errorf("expected recipient new version 6, got %d", result.RecipientNewVersion)
	}
	if result.AmountCents != amountCents {
		t.Errorf("expected settled amount %d, got %d", amountCents, result.AmountCents)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled mock expectations: %v", err)
	}
}

// TestOCC_SettleTransactionOCC_SenderConflictRollsBack verifies that if the sender's version
// changed concurrently, the debit affects 0 rows, returns ErrOptimisticLockConflict,
// and rolls back the database transaction immediately without executing credit or journal inserts.
func TestOCC_SettleTransactionOCC_SenderConflictRollsBack(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to initialize sqlmock: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	senderID := uuid.New().String()
	recipientID := uuid.New().String()

	senderParams := OCCDebitParams{
		AccountID:       senderID,
		ExpectedVersion: 1, // Stale version!
		AmountCents:     1000,
	}
	recipientParams := OCCCreditParams{
		AccountID:       recipientID,
		ExpectedVersion: 3,
		AmountCents:     1000,
	}

	mock.ExpectBegin()

	// Debit leg matches 0 rows!
	mock.ExpectExec(regexp.QuoteMeta(`
		UPDATE accounts
		SET balance_cents = balance_cents - $1,
		    version = version + 1,
		    updated_at = NOW()
		WHERE id = $2
		  AND version = $3
		  AND balance_cents >= $1;
	`)).
		WithArgs(int64(1000), senderID, 1).
		WillReturnResult(sqlmock.NewResult(0, 0))

	// The defer tx.Rollback() must be called
	mock.ExpectRollback()

	result, err := SettleTransactionOCC(ctx, db, senderParams, recipientParams)
	if err == nil {
		t.Fatal("expected settlement to fail due to sender OCC conflict, got nil error")
	}
	if !errors.Is(err, ErrOptimisticLockConflict) {
		t.Errorf("expected ErrOptimisticLockConflict, got: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil result on failure, got: %+v", result)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled mock expectations: %v", err)
	}
}

// TestOCC_SettleTransactionOCC_RecipientConflictRollsBack verifies that if the sender debit succeeds
// but the recipient version changed concurrently, the transaction rolls back cleanly, discarding the debit.
func TestOCC_SettleTransactionOCC_RecipientConflictRollsBack(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to initialize sqlmock: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	senderID := uuid.New().String()
	recipientID := uuid.New().String()

	senderParams := OCCDebitParams{
		AccountID:       senderID,
		ExpectedVersion: 1,
		AmountCents:     1000,
	}
	recipientParams := OCCCreditParams{
		AccountID:       recipientID,
		ExpectedVersion: 3, // Stale recipient version!
		AmountCents:     1000,
	}

	mock.ExpectBegin()

	// 1. Debit leg succeeds
	mock.ExpectExec(regexp.QuoteMeta(`
		UPDATE accounts
		SET balance_cents = balance_cents - $1,
		    version = version + 1,
		    updated_at = NOW()
		WHERE id = $2
		  AND version = $3
		  AND balance_cents >= $1;
	`)).
		WithArgs(int64(1000), senderID, 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// 2. Credit leg matches 0 rows (conflict!)
	mock.ExpectExec(regexp.QuoteMeta(`
		UPDATE accounts
		SET balance_cents = balance_cents + $1,
		    version = version + 1,
		    updated_at = NOW()
		WHERE id = $2
		  AND version = $3;
	`)).
		WithArgs(int64(1000), recipientID, 3).
		WillReturnResult(sqlmock.NewResult(0, 0))

	// Rollback must fire
	mock.ExpectRollback()

	result, err := SettleTransactionOCC(ctx, db, senderParams, recipientParams)
	if err == nil {
		t.Fatal("expected settlement to fail due to recipient OCC conflict, got nil error")
	}
	if !errors.Is(err, ErrOptimisticLockConflict) {
		t.Errorf("expected ErrOptimisticLockConflict, got: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil result on failure, got: %+v", result)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled mock expectations: %v", err)
	}
}

// TestOCC_ConcurrentCollision50Workers simulates 50 concurrent goroutines racing to update
// the exact same account at expectedVersion = 0.
// INVARIANT PROOF: Exactly 1 goroutine must win the race (succeed);
// Exactly 49 goroutines must be rejected safely with ErrOptimisticLockConflict!
func TestOCC_ConcurrentCollision50Workers(t *testing.T) {
	const workers = 50
	store := NewInMemoryOCCStore()
	accountID := "acc-collision-test"
	initialBalance := int64(100000) // $1,000.00

	store.CreateAccount(accountID, initialBalance)

	var wg sync.WaitGroup
	startGate := make(chan struct{})

	var successCount atomic.Int64
	var conflictCount atomic.Int64
	var otherErrorCount atomic.Int64

	// Spawn 50 goroutines lined up at the start gate
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			<-startGate // Wait for synchronized nanosecond release

			// All 50 workers believe the version is 0!
			_, err := store.UpdateBalanceOCC(accountID, 0, 100)
			if err == nil {
				successCount.Add(1)
			} else if errors.Is(err, ErrOptimisticLockConflict) {
				conflictCount.Add(1)
			} else {
				otherErrorCount.Add(1)
			}
		}(i)
	}

	// Fire start barrier
	close(startGate)
	wg.Wait()

	// Verify OCC mathematical invariants
	if successCount.Load() != 1 {
		t.Errorf("OCC INVARIANT VIOLATION: expected exactly 1 winner, got %d", successCount.Load())
	}
	if conflictCount.Load() != int64(workers-1) {
		t.Errorf("OCC INVARIANT VIOLATION: expected %d conflicts, got %d", workers-1, conflictCount.Load())
	}
	if otherErrorCount.Load() != 0 {
		t.Errorf("unexpected errors encountered: %d", otherErrorCount.Load())
	}

	// Verify account state
	acc, err := store.GetAccount(accountID)
	if err != nil {
		t.Fatalf("failed to retrieve account: %v", err)
	}
	if acc.Version != 1 {
		t.Errorf("expected final account version 1, got %d", acc.Version)
	}
	if acc.BalanceCents != initialBalance+100 {
		t.Errorf("expected final balance %d, got %d", initialBalance+100, acc.BalanceCents)
	}

	t.Logf("OCC COLLISION TEST PROVEN: 1 winner succeeded, %d stale writes safely rejected!", conflictCount.Load())
}

// TestOCC_RetryLoop_ResolvesContention simulates an automated retry loop where 10 concurrent
// workers hit the account, and all 10 eventually succeed sequentially as versions advance.
func TestOCC_RetryLoop_ResolvesContention(t *testing.T) {
	const workers = 10
	store := NewInMemoryOCCStore()
	accountID := "acc-retry-test"
	initialBalance := int64(50000)

	store.CreateAccount(accountID, initialBalance)

	var wg sync.WaitGroup
	startGate := make(chan struct{})
	var successCount atomic.Int64
	var totalRetries atomic.Int64

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			<-startGate

			// Worker retry loop
			const maxRetries = 20
			for attempt := 1; attempt <= maxRetries; attempt++ {
				acc, err := store.GetAccount(accountID)
				if err != nil {
					return
				}

				_, err = store.UpdateBalanceOCC(accountID, acc.Version, 100)
				if err == nil {
					successCount.Add(1)
					return
				}

				if errors.Is(err, ErrOptimisticLockConflict) {
					totalRetries.Add(1)
					time.Sleep(time.Duration(1+workerID%3) * time.Millisecond)
					continue
				}
				return
			}
		}(i)
	}

	close(startGate)
	wg.Wait()

	if successCount.Load() != workers {
		t.Errorf("expected all %d workers to eventually succeed with retry, got %d", workers, successCount.Load())
	}

	acc, _ := store.GetAccount(accountID)
	expectedFinalBalance := initialBalance + int64(workers*100)
	if acc.BalanceCents != expectedFinalBalance {
		t.Errorf("expected final balance %d, got %d", expectedFinalBalance, acc.BalanceCents)
	}
	if acc.Version != workers {
		t.Errorf("expected final version %d, got %d", workers, acc.Version)
	}

	t.Logf("OCC RETRY LOOP PROVEN: All %d transfers settled, version advanced to %d across %d collisions!",
		workers, acc.Version, totalRetries.Load())
}

// TestOCC_DefensiveValidation tests pre-flight parameter validation.
func TestOCC_DefensiveValidation(t *testing.T) {
	ctx := context.Background()

	// 1. Nil DB
	_, err := UpdateAccountBalanceOCC(ctx, nil, "acc-1", 0, 100)
	if err == nil {
		t.Error("expected error on nil db")
	}

	// 2. Empty account ID
	db, _, _ := sqlmock.New()
	defer db.Close()
	_, err = UpdateAccountBalanceOCC(ctx, db, "   ", 0, 100)
	if err == nil {
		t.Error("expected error on empty account ID")
	}

	// 3. SettleTransactionOCC with mismatched amounts
	sender := OCCDebitParams{AccountID: "a1", ExpectedVersion: 0, AmountCents: 100}
	recipient := OCCCreditParams{AccountID: "a2", ExpectedVersion: 0, AmountCents: 200}
	_, err = SettleTransactionOCC(ctx, db, sender, recipient)
	if err == nil || !errors.Is(err, nil) && err.Error() == "" {
		t.Error("expected error on mismatched transfer amounts")
	}

	// 4. SettleTransactionOCC with self-transfer
	sender.AccountID = "a1"
	recipient.AccountID = "a1"
	recipient.AmountCents = 100
	_, err = SettleTransactionOCC(ctx, db, sender, recipient)
	if err == nil {
		t.Error("expected error on self-transfer")
	}
}
