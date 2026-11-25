// ==============================================================================
// NOVARIS: Distributed Payment Settlement & Adaptive Fraud Intelligence Platform
// MODULE 1: API Gateway Service - Optimistic Concurrency Control (OCC) Engine (Day 9)
// FILE: services/gateway/occ.go
// ==============================================================================
//
// Architectural Principles & Invariants:
// 1. Optimistic Concurrency Control (OCC): Instead of acquiring long-lived pessimistic
//    row locks (SELECT ... FOR UPDATE) which serialize throughput and induce lock queues,
//    OCC assumes transaction collisions are rare. We read the row's version number,
//    perform business calculations in application memory, and execute an atomic
//    conditional write:
//        UPDATE accounts
//        SET balance_cents = balance_cents + $1,
//            version = version + 1,
//            updated_at = NOW()
//        WHERE id = $2 AND version = $3;
//
// 2. Atomic Collision Detection:
//    If another concurrent transaction modified the account row between our read and
//    our update, the database row's version has already incremented.
//    PostgreSQL's MVCC evaluates the WHERE clause and matches ZERO rows.
//    When RowsAffected() == 0, we detect the collision with 100% mathematical certainty,
//    abort the transaction immediately, and safely return ErrOptimisticLockConflict.
//
// 3. Lost-Update & Double-Spending Neutralization:
//    In a distributed microservice mesh with multiple gateway pods, in-memory mutexes
//    fail across container boundaries. OCC enforces database-level CAS (Compare-And-Swap),
//    guaranteeing that NO transaction can overwrite state without acknowledging
//    intervening mutations.
//
// 4. Safe Automated Retry with Exponential Backoff:
//    When a collision occurs, the operation is not permanently dead; it can be retried
//    by re-reading the newest state and version, recalculating the delta, and attempting
//    the conditional write again up to a maximum retry threshold.
// ==============================================================================

package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Sentinel errors for deterministic error triage across the API gateway and client tiers.
var (
	// ErrOptimisticLockConflict is returned when an UPDATE modifies 0 rows due to a version mismatch.
	ErrOptimisticLockConflict = errors.New("optimistic lock conflict: account version has changed concurrently")

	// ErrAccountNotFound is returned when an account UUID does not exist in the ledger.
	ErrAccountNotFound = errors.New("account not found")

	// ErrInsufficientFunds is returned when an account's balance is lower than the transfer amount.
	ErrInsufficientFunds = errors.New("insufficient funds for settlement")

	// ErrMaxRetriesExceeded is returned when an OCC retry loop fails after all backoff attempts.
	ErrMaxRetriesExceeded = errors.New("maximum OCC retry attempts exceeded")
)

// Account represents a financial wallet, settlement pool, or merchant balance in the ledger.
type Account struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	AccountType  string    `json:"account_type"`
	Currency     string    `json:"currency"`
	BalanceCents int64     `json:"balance_cents"`
	Version      int       `json:"version"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// OCCDebitParams encapsulates the parameters required to conditionally debit an account.
type OCCDebitParams struct {
	AccountID       string
	ExpectedVersion int
	AmountCents     int64
}

// OCCCreditParams encapsulates the parameters required to conditionally credit an account.
type OCCCreditParams struct {
	AccountID       string
	ExpectedVersion int
	AmountCents     int64
}

// OCCSettlementResult contains the outcome of an atomic OCC-protected double-entry transfer.
type OCCSettlementResult struct {
	TransactionID       string    `json:"transaction_id"`
	SenderID            string    `json:"sender_id"`
	RecipientID         string    `json:"recipient_id"`
	AmountCents         int64     `json:"amount_cents"`
	SenderNewVersion    int       `json:"sender_new_version"`
	RecipientNewVersion int       `json:"recipient_new_version"`
	SettledAt           time.Time `json:"settled_at"`
}

// UpdateAccountBalanceOCC executes an atomic conditional balance update on a single account
// using Optimistic Concurrency Control (OCC).
//
// Parameters:
//   - ctx: Context for query timeout and deadline cancellation.
//   - db: Database connection pool or handle.
//   - accountID: UUID string of the account to mutate.
//   - expectedVersion: The version number previously read by the caller.
//   - deltaCents: Monetary change in integer cents (+ for credit, - for debit).
//
// Returns:
//   - newVersion: The incremented version (expectedVersion + 1) upon success.
//   - err: ErrOptimisticLockConflict if rows affected == 0, or driver error.
func UpdateAccountBalanceOCC(ctx context.Context, db *sql.DB, accountID string, expectedVersion int, deltaCents int64) (int, error) {
	if db == nil {
		return 0, fmt.Errorf("UpdateAccountBalanceOCC: database connection pool is nil")
	}

	trimmedID := strings.TrimSpace(accountID)
	if trimmedID == "" {
		return 0, fmt.Errorf("UpdateAccountBalanceOCC: account ID cannot be empty")
	}

	query := `
		UPDATE accounts
		SET balance_cents = balance_cents + $1,
		    version = version + 1,
		    updated_at = NOW()
		WHERE id = $2
		  AND version = $3;
	`

	res, err := db.ExecContext(ctx, query, deltaCents, trimmedID, expectedVersion)
	if err != nil {
		return 0, fmt.Errorf("executing OCC balance update for account %s: %w", trimmedID, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("retrieving rows affected for account %s: %w", trimmedID, err)
	}

	// The Golden OCC Invariant: If 0 rows were updated, someone else modified the row first!
	if rowsAffected == 0 {
		return 0, fmt.Errorf("%w: account %s (expected version %d)", ErrOptimisticLockConflict, trimmedID, expectedVersion)
	}

	return expectedVersion + 1, nil
}

// DebitAccountOCCInTx conditionally debits an account within an existing SQL transaction.
// Enforces BOTH version matching AND non-negative balance in a single atomic SQL statement.
func DebitAccountOCCInTx(ctx context.Context, tx *sql.Tx, accountID string, expectedVersion int, amountCents int64) (int, error) {
	if tx == nil {
		return 0, fmt.Errorf("DebitAccountOCCInTx: transaction handle is nil")
	}
	if amountCents <= 0 {
		return 0, fmt.Errorf("DebitAccountOCCInTx: debit amount must be strictly positive, got %d", amountCents)
	}

	trimmedID := strings.TrimSpace(accountID)
	if trimmedID == "" {
		return 0, fmt.Errorf("DebitAccountOCCInTx: account ID cannot be empty")
	}

	query := `
		UPDATE accounts
		SET balance_cents = balance_cents - $1,
		    version = version + 1,
		    updated_at = NOW()
		WHERE id = $2
		  AND version = $3
		  AND balance_cents >= $1;
	`

	res, err := tx.ExecContext(ctx, query, amountCents, trimmedID, expectedVersion)
	if err != nil {
		return 0, fmt.Errorf("executing OCC debit for account %s: %w", trimmedID, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("retrieving rows affected for debit on account %s: %w", trimmedID, err)
	}

	if rowsAffected == 0 {
		// Zero rows matched: either the version changed OR balance was insufficient.
		return 0, fmt.Errorf("%w: account %s (expected version %d, debit %d cents)", ErrOptimisticLockConflict, trimmedID, expectedVersion, amountCents)
	}

	return expectedVersion + 1, nil
}

// CreditAccountOCCInTx conditionally credits an account within an existing SQL transaction.
func CreditAccountOCCInTx(ctx context.Context, tx *sql.Tx, accountID string, expectedVersion int, amountCents int64) (int, error) {
	if tx == nil {
		return 0, fmt.Errorf("CreditAccountOCCInTx: transaction handle is nil")
	}
	if amountCents <= 0 {
		return 0, fmt.Errorf("CreditAccountOCCInTx: credit amount must be strictly positive, got %d", amountCents)
	}

	trimmedID := strings.TrimSpace(accountID)
	if trimmedID == "" {
		return 0, fmt.Errorf("CreditAccountOCCInTx: account ID cannot be empty")
	}

	query := `
		UPDATE accounts
		SET balance_cents = balance_cents + $1,
		    version = version + 1,
		    updated_at = NOW()
		WHERE id = $2
		  AND version = $3;
	`

	res, err := tx.ExecContext(ctx, query, amountCents, trimmedID, expectedVersion)
	if err != nil {
		return 0, fmt.Errorf("executing OCC credit for account %s: %w", trimmedID, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("retrieving rows affected for credit on account %s: %w", trimmedID, err)
	}

	if rowsAffected == 0 {
		return 0, fmt.Errorf("%w: account %s (expected version %d, credit %d cents)", ErrOptimisticLockConflict, trimmedID, expectedVersion, amountCents)
	}

	return expectedVersion + 1, nil
}

// SettleTransactionOCC executes a complete double-entry settlement with Optimistic Concurrency Control.
// It atomically:
// 1. Validates inputs
// 2. Debits the sender with OCC version check & balance floor
// 3. Credits the recipient with OCC version check
// 4. Inserts the parent transactions record
// 5. Inserts 2 balanced journal entries (DEBIT and CREDIT)
// 6. Commits all changes atomically
func SettleTransactionOCC(ctx context.Context, db *sql.DB, sender OCCDebitParams, recipient OCCCreditParams) (*OCCSettlementResult, error) {
	if db == nil {
		return nil, fmt.Errorf("SettleTransactionOCC: database connection pool is nil")
	}

	if sender.AmountCents <= 0 || recipient.AmountCents <= 0 {
		return nil, fmt.Errorf("SettleTransactionOCC: transfer amounts must be strictly positive")
	}

	if sender.AmountCents != recipient.AmountCents {
		return nil, fmt.Errorf("SettleTransactionOCC: double-entry invariant violated: sender amount (%d) != recipient amount (%d)",
			sender.AmountCents, recipient.AmountCents)
	}

	senderID := strings.TrimSpace(sender.AccountID)
	recipientID := strings.TrimSpace(recipient.AccountID)
	if senderID == "" || recipientID == "" {
		return nil, fmt.Errorf("SettleTransactionOCC: account IDs cannot be empty")
	}
	if senderID == recipientID {
		return nil, fmt.Errorf("SettleTransactionOCC: self-transfers disallowed between identical account %s", senderID)
	}

	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, fmt.Errorf("beginning OCC settlement transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. Execute conditional OCC debit on sender
	newSenderVer, err := DebitAccountOCCInTx(ctx, tx, senderID, sender.ExpectedVersion, sender.AmountCents)
	if err != nil {
		return nil, fmt.Errorf("debit leg: %w", err)
	}

	// 2. Execute conditional OCC credit on recipient
	newRecipientVer, err := CreditAccountOCCInTx(ctx, tx, recipientID, recipient.ExpectedVersion, recipient.AmountCents)
	if err != nil {
		return nil, fmt.Errorf("credit leg: %w", err)
	}

	// 3. Insert parent transaction record
	transactionID := uuid.New().String()
	idempotencyKey := fmt.Sprintf("idemp_occ_%s", transactionID)

	_, err = tx.ExecContext(ctx, `
		INSERT INTO transactions
		 (id, idempotency_key, sender_account_id, recipient_account_id, amount_cents, status)
		 VALUES ($1, $2, $3, $4, $5, 'SETTLED');
	`, transactionID, idempotencyKey, senderID, recipientID, sender.AmountCents)
	if err != nil {
		return nil, fmt.Errorf("inserting transaction audit record: %w", err)
	}

	// 4. Insert balanced journal entries
	_, err = tx.ExecContext(ctx, `
		INSERT INTO journal_entries
		 (transaction_id, account_id, entry_type, amount_cents)
		 VALUES ($1, $2, 'DEBIT', $3);
	`, transactionID, senderID, sender.AmountCents)
	if err != nil {
		return nil, fmt.Errorf("inserting debit journal entry: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO journal_entries
		 (transaction_id, account_id, entry_type, amount_cents)
		 VALUES ($1, $2, 'CREDIT', $3);
	`, transactionID, recipientID, recipient.AmountCents)
	if err != nil {
		return nil, fmt.Errorf("inserting credit journal entry: %w", err)
	}

	// 5. Commit atomic transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("committing OCC settlement transaction: %w", err)
	}

	return &OCCSettlementResult{
		TransactionID:       transactionID,
		SenderID:            senderID,
		RecipientID:         recipientID,
		AmountCents:         sender.AmountCents,
		SenderNewVersion:    newSenderVer,
		RecipientNewVersion: newRecipientVer,
		SettledAt:           time.Now().UTC(),
	}, nil
}

// VersionFetcher is a function signature that retrieves current account versions.
type VersionFetcher func(ctx context.Context, senderID, recipientID string) (senderVer, recipientVer int, err error)

// SettleWithOCCRetry executes SettleTransactionOCC with an automated retry loop and exponential backoff.
// When an OCC collision occurs, it catches ErrOptimisticLockConflict, re-fetches versions, and retries.
func SettleWithOCCRetry(
	ctx context.Context,
	db *sql.DB,
	fetchVersions VersionFetcher,
	senderID, recipientID string,
	amountCents int64,
	maxRetries int,
	baseBackoff time.Duration,
) (*OCCSettlementResult, int, error) {
	if maxRetries <= 0 {
		maxRetries = 3
	}
	if baseBackoff <= 0 {
		baseBackoff = 5 * time.Millisecond
	}

	attempts := 0
	for attempts = 1; attempts <= maxRetries; attempts++ {
		// Fetch fresh expected versions
		sVer, rVer, err := fetchVersions(ctx, senderID, recipientID)
		if err != nil {
			return nil, attempts, fmt.Errorf("retry attempt %d: fetching account versions: %w", attempts, err)
		}

		senderParams := OCCDebitParams{
			AccountID:       senderID,
			ExpectedVersion: sVer,
			AmountCents:     amountCents,
		}
		recipientParams := OCCCreditParams{
			AccountID:       recipientID,
			ExpectedVersion: rVer,
			AmountCents:     amountCents,
		}

		result, err := SettleTransactionOCC(ctx, db, senderParams, recipientParams)
		if err == nil {
			return result, attempts, nil
		}

		// Check if error is an OCC conflict (stale version)
		if errors.Is(err, ErrOptimisticLockConflict) {
			if attempts == maxRetries {
				return nil, attempts, fmt.Errorf("%w: exhausted %d attempts due to contention: %v",
					ErrMaxRetriesExceeded, maxRetries, err)
			}

			// Exponential backoff with full jitter to avoid thundering herd
			jitter := time.Duration(rand.Int63n(int64(baseBackoff)))
			sleepDuration := (baseBackoff * time.Duration(1<<uint(attempts-1))) + jitter

			select {
			case <-ctx.Done():
				return nil, attempts, ctx.Err()
			case <-time.After(sleepDuration):
				// Continue to next retry
			}
			continue
		}

		// Non-retryable error (e.g. database disconnect, foreign key violation)
		return nil, attempts, err
	}

	return nil, attempts, ErrMaxRetriesExceeded
}

// ------------------------------------------------------------------------------
// In-Memory OCC Engine (High-Performance Zero-Dependency Simulator)
// ------------------------------------------------------------------------------

// InMemoryAccount represents a thread-safe in-memory representation of an account row.
type InMemoryAccount struct {
	ID           string
	BalanceCents int64
	Version      int
}

// InMemoryOCCStore simulates database-level Optimistic Concurrency Control in memory.
// Useful for unit testing concurrency logic at microsecond latency without external dependencies.
type InMemoryOCCStore struct {
	mu       sync.RWMutex
	accounts map[string]*InMemoryAccount
}

// NewInMemoryOCCStore initializes an empty in-memory OCC store.
func NewInMemoryOCCStore() *InMemoryOCCStore {
	return &InMemoryOCCStore{
		accounts: make(map[string]*InMemoryAccount),
	}
}

// CreateAccount registers an account with an initial balance and version 0.
func (s *InMemoryOCCStore) CreateAccount(id string, initialBalance int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.accounts[id] = &InMemoryAccount{
		ID:           id,
		BalanceCents: initialBalance,
		Version:      0,
	}
}

// GetAccount returns a copy of the account's state.
func (s *InMemoryOCCStore) GetAccount(id string) (*InMemoryAccount, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	acc, exists := s.accounts[id]
	if !exists {
		return nil, ErrAccountNotFound
	}
	return &InMemoryAccount{
		ID:           acc.ID,
		BalanceCents: acc.BalanceCents,
		Version:      acc.Version,
	}, nil
}

// UpdateBalanceOCC performs an atomic Compare-And-Swap (CAS) update on the account.
// If the stored version does not match expectedVersion, it rejects the write safely.
func (s *InMemoryOCCStore) UpdateBalanceOCC(id string, expectedVersion int, deltaCents int64) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	acc, exists := s.accounts[id]
	if !exists {
		return 0, ErrAccountNotFound
	}

	// The In-Memory OCC Check:
	if acc.Version != expectedVersion {
		return 0, fmt.Errorf("%w: account %s has version %d, expected %d",
			ErrOptimisticLockConflict, id, acc.Version, expectedVersion)
	}

	if deltaCents < 0 && acc.BalanceCents+deltaCents < 0 {
		return 0, ErrInsufficientFunds
	}

	acc.BalanceCents += deltaCents
	acc.Version++
	return acc.Version, nil
}
