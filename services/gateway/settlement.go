// ==============================================================================
// NOVARIS: Distributed Payment Settlement & Adaptive Fraud Intelligence Platform
// MODULE 1: API Gateway Service - Core Settlement Engine (Day 6)
// FILE: services/gateway/settlement.go
// ==============================================================================
//
// Package main contains the core double-entry financial settlement logic.
// This is the function that actually moves money, safely, atomically, or not at all.
//
// Architectural Invariants:
// 1. Atomicity (ACID "A"): Either all three rows (1 transaction + 2 journal entries)
//    are permanently committed to the ledger, or NONE of them are.
// 2. Double-Entry Balance Invariant: Every transaction produces balanced journal rows:
//    Sum(Debits) == Sum(Credits) == amountCents. Money is never created or destroyed.
// 3. Defer Rollback Safety Net: defer tx.Rollback() is registered immediately after Begin().
//    If any error occurs or a panic unwinds the stack, all partial writes are discarded.
//    If tx.Commit() succeeds first, tx.Rollback() safely returns sql.ErrTxDone (a no-op).
// 4. Parameterized Statements: Every query uses positional parameter placeholders ($1, $2, ...)
//    preventing SQL injection vulnerabilities at the compiler/driver level.
// ==============================================================================

package main

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/charan21042005/novaris/services/gateway/internal/outbox"
)

// settleTransaction inserts one transaction row and two balanced journal entries
// (one DEBIT on sender, one CREDIT on recipient) inside a single ACID database transaction.
//
// Parameters:
//   - db: The shared *sql.DB connection pool.
//   - senderID: The UUID of the account whose funds are being debited.
//   - recipientID: The UUID of the account receiving the funds.
//   - amountCents: The transfer magnitude in integer cents (strictly positive).
//
// Returns:
//   - transactionID: The cryptographically random UUID assigned to the new transaction.
//   - err: Any error encountered during transaction execution, wrapped with %w context.
func settleTransaction(db *sql.DB, senderID, recipientID string, amountCents int64) (string, error) {
	// --------------------------------------------------------------------------
	// STEP 0: Defensive Pre-Flight Validation (Fail-Fast before opening DB tx)
	// --------------------------------------------------------------------------
	if db == nil {
		return "", fmt.Errorf("settleTransaction: database connection pool is nil")
	}

	if amountCents <= 0 {
		return "", fmt.Errorf("settleTransaction: invalid amount %d: amount must be strictly positive", amountCents)
	}

	trimmedSender := strings.TrimSpace(senderID)
	if trimmedSender == "" {
		return "", fmt.Errorf("settleTransaction: sender account ID cannot be empty")
	}

	trimmedRecipient := strings.TrimSpace(recipientID)
	if trimmedRecipient == "" {
		return "", fmt.Errorf("settleTransaction: recipient account ID cannot be empty")
	}

	if trimmedSender == trimmedRecipient {
		return "", fmt.Errorf("settleTransaction: self-transfers disallowed: sender and recipient account IDs (%s) must be distinct", trimmedSender)
	}

	// --------------------------------------------------------------------------
	// STEP 1: Begin the ACID database transaction & register the safety net.
	// --------------------------------------------------------------------------
	// db.Begin() allocates a dedicated connection from the pool and issues "BEGIN"
	// to PostgreSQL. All subsequent queries must use the returned *sql.Tx instance.
	tx, err := db.Begin()
	if err != nil {
		return "", fmt.Errorf("beginning database transaction: %w", err)
	}

	// SAFETY NET IDIOM: Register defer tx.Rollback() immediately after a successful Begin().
	// Why this is safe and mandatory in production Go:
	// 1. If any subsequent insert returns an error, the function exits early, and
	//    this deferred call automatically issues "ROLLBACK" to PostgreSQL, aborting
	//    all prior uncommitted inserts in this transaction.
	// 2. If a runtime panic occurs, Go executes deferred calls during stack unwinding,
	//    guaranteeing that open database locks and incomplete mutations are freed.
	// 3. If tx.Commit() succeeds at the end, the transaction is marked 'done'.
	//    When defer tx.Rollback() executes on function return, it encounters
	//    sql.ErrTxDone and silently returns without doing anything. It is a 100% safe no-op.
	defer tx.Rollback()

	// Generate a version-4 UUID for the new transaction record
	transactionID := uuid.New().String()

	// Generate a unique idempotency key for Day 6 (ensures uniqueness in transactions table)
	idempotencyKey := fmt.Sprintf("idemp_%s", transactionID)

	// --------------------------------------------------------------------------
	// STEP 2: Insert the parent transaction record.
	// --------------------------------------------------------------------------
	// CRITICAL RULE: Use tx.Exec(), NEVER db.Exec()! Calling db.Exec() executes
	// outside this transaction on an arbitrary pooled connection, bypassing atomicity!
	_, err = tx.Exec(
		`INSERT INTO transactions
		 (id, idempotency_key, sender_account_id, recipient_account_id, amount_cents, status)
		 VALUES ($1, $2, $3, $4, $5, 'SETTLED')`,
		transactionID, idempotencyKey, senderID, recipientID, amountCents,
	)
	if err != nil {
		// Return error wrapped with %w to preserve the underlying driver error (e.g. FK violation)
		return "", fmt.Errorf("inserting transaction: %w", err)
	}

	// --------------------------------------------------------------------------
	// STEP 3: Insert the DEBIT journal entry (money LEAVES the sender).
	// --------------------------------------------------------------------------
	_, err = tx.Exec(
		`INSERT INTO journal_entries
		 (transaction_id, account_id, entry_type, amount_cents)
		 VALUES ($1, $2, 'DEBIT', $3)`,
		transactionID, senderID, amountCents,
	)
	if err != nil {
		return "", fmt.Errorf("inserting debit journal entry: %w", err)
	}

	// --------------------------------------------------------------------------
	// STEP 4: Insert the CREDIT journal entry (money ENTERS the recipient).
	// --------------------------------------------------------------------------
	_, err = tx.Exec(
		`INSERT INTO journal_entries
		 (transaction_id, account_id, entry_type, amount_cents)
		 VALUES ($1, $2, 'CREDIT', $3)`,
		transactionID, recipientID, amountCents,
	)
	if err != nil {
		return "", fmt.Errorf("inserting credit journal entry: %w", err)
	}

	// --------------------------------------------------------------------------
	// STEP 5: Commit the transaction.
	// --------------------------------------------------------------------------
	// Sends "COMMIT" to PostgreSQL. At this exact microsecond, all three rows
	// become permanently durable and atomically visible to all other transactions.
	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("committing transaction: %w", err)
	}

	return transactionID, nil
}

// SettleTransactionWithOutbox executes atomic financial settlement and enlists
// an outbox event in the EXACT SAME database transaction (Day 18 Transactional Outbox).
//
// Dual-Write Invariant:
// Either all financial entries AND the outbox event commit together, or NONE do.
// If the database transaction aborts, no phantom outbox event is ever persisted.
// If it commits, the OutboxRelay is guaranteed to find it and publish to Kafka.
func SettleTransactionWithOutbox(
	ctx context.Context,
	db *sql.DB,
	outboxRepo outbox.Repository,
	senderID, recipientID string,
	amountCents int64,
	eventData interface{},
) (string, *outbox.OutboxRecord, error) {
	if db == nil {
		return "", nil, fmt.Errorf("SettleTransactionWithOutbox: database connection pool is nil")
	}
	if outboxRepo == nil {
		return "", nil, fmt.Errorf("SettleTransactionWithOutbox: outbox repository is nil")
	}

	if amountCents <= 0 {
		return "", nil, fmt.Errorf("SettleTransactionWithOutbox: invalid amount %d: must be strictly positive", amountCents)
	}

	trimmedSender := strings.TrimSpace(senderID)
	trimmedRecipient := strings.TrimSpace(recipientID)
	if trimmedSender == "" || trimmedRecipient == "" {
		return "", nil, fmt.Errorf("SettleTransactionWithOutbox: account IDs cannot be empty")
	}
	if trimmedSender == trimmedRecipient {
		return "", nil, fmt.Errorf("SettleTransactionWithOutbox: self-transfers disallowed (%s)", trimmedSender)
	}

	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return "", nil, fmt.Errorf("beginning outbox settlement transaction: %w", err)
	}
	defer tx.Rollback()

	transactionID := uuid.New().String()
	idempotencyKey := fmt.Sprintf("idemp_outbox_%s", transactionID)

	// 1. Insert parent transaction record
	_, err = tx.ExecContext(ctx,
		`INSERT INTO transactions
		 (id, idempotency_key, sender_account_id, recipient_account_id, amount_cents, status)
		 VALUES ($1, $2, $3, $4, $5, 'SETTLED')`,
		transactionID, idempotencyKey, trimmedSender, trimmedRecipient, amountCents,
	)
	if err != nil {
		return "", nil, fmt.Errorf("inserting transaction: %w", err)
	}

	// 2. Insert debit journal entry
	_, err = tx.ExecContext(ctx,
		`INSERT INTO journal_entries
		 (transaction_id, account_id, entry_type, amount_cents)
		 VALUES ($1, $2, 'DEBIT', $3)`,
		transactionID, trimmedSender, amountCents,
	)
	if err != nil {
		return "", nil, fmt.Errorf("inserting debit journal entry: %w", err)
	}

	// 3. Insert credit journal entry
	_, err = tx.ExecContext(ctx,
		`INSERT INTO journal_entries
		 (transaction_id, account_id, entry_type, amount_cents)
		 VALUES ($1, $2, 'CREDIT', $3)`,
		transactionID, trimmedRecipient, amountCents,
	)
	if err != nil {
		return "", nil, fmt.Errorf("inserting credit journal entry: %w", err)
	}

	// 4. Enlist Outbox Event inside the SAME database transaction
	outboxRec, err := outbox.NewOutboxRecord(
		outbox.AggregateTransaction,
		trimmedSender, // Partition Key: sender UUID guarantees strict per-account FIFO!
		"novaris.tx.initiated",
		eventData,
	)
	if err != nil {
		return "", nil, fmt.Errorf("constructing outbox record: %w", err)
	}

	if err := outboxRepo.InsertTx(ctx, tx, outboxRec); err != nil {
		return "", nil, fmt.Errorf("inserting outbox record in transaction: %w", err)
	}

	// 5. Commit atomic transaction (ACID guarantees all 4 mutations commit simultaneously)
	if err := tx.Commit(); err != nil {
		return "", nil, fmt.Errorf("committing outbox settlement transaction: %w", err)
	}

	return transactionID, outboxRec, nil
}
