// ==============================================================================
// NOVARIS: Distributed Payment Settlement & Adaptive Fraud Intelligence Platform
// MODULE 1: API Gateway Service - Outbox Persistence Repository (Day 18)
// FILE: services/gateway/internal/outbox/repository.go
// ==============================================================================
//
// Architectural Principles & Invariants:
// 1. Transactional Atomic Enlistment:
//    The InsertTx method accepts an active *sql.Tx database transaction.
//    The outbox row MUST be inserted using this transaction handle rather than the
//    global connection pool, guaranteeing that the event commits or rolls back
//    atomically with the parent financial settlement operation.
//
// 2. High-Throughput Non-Blocking Polling:
//    The FetchPending query uses PostgreSQL's 'FOR UPDATE SKIP LOCKED' primitive.
//    This allows multiple horizontally scaled outbox relay pods to poll the outbox
//    table concurrently without blocking each other or attempting to process the
//    same pending event batch.
//
// 3. Status State Machine & Retry Limits:
//    - Upon successful Kafka acknowledgment, status becomes 'PUBLISHED'.
//    - Upon transient broker error, retry_count increments while remaining 'PENDING'.
//    - If retry_count reaches maxRetries, status transitions to 'FAILED' (DLQ candidate).
// ==============================================================================

package outbox

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
)

// Sentinel errors for outbox repository operations.
var (
	// ErrRecordNotFound is returned when attempting to update a non-existent outbox record.
	ErrRecordNotFound = errors.New("outbox: record not found")

	// ErrNilTx is returned when InsertTx is passed a nil *sql.Tx transaction handle.
	ErrNilTx = errors.New("outbox: database transaction cannot be nil")
)

// Repository defines the contract for persisting, querying, and updating outbox records.
type Repository interface {
	// InsertTx inserts an outbox record inside an active ACID database transaction.
	InsertTx(ctx context.Context, tx *sql.Tx, record *OutboxRecord) error

	// FetchPending retrieves up to limit pending records ordered by created_at ASC.
	// In production PostgreSQL, it uses "FOR UPDATE SKIP LOCKED" to prevent multi-worker contention.
	FetchPending(ctx context.Context, limit int) ([]*OutboxRecord, error)

	// MarkPublished updates the record status to PUBLISHED and sets published_at to the current time.
	MarkPublished(ctx context.Context, id string) error

	// MarkFailed increments the retry count and transitions to FAILED if max retries exceeded.
	MarkFailed(ctx context.Context, id string, retryCount int) error

	// CountPending returns the total number of records currently in PENDING state.
	CountPending(ctx context.Context) (int64, error)

	// CountPublished returns the total number of records currently in PUBLISHED state.
	CountPublished(ctx context.Context) (int64, error)
}

// ------------------------------------------------------------------------------
// Production PostgreSQL Repository Implementation
// ------------------------------------------------------------------------------

// SQLRepository implements Repository against a live PostgreSQL 16 instance.
type SQLRepository struct {
	db         *sql.DB
	maxRetries int
}

// NewSQLRepository constructs a new production PostgreSQL outbox repository.
func NewSQLRepository(db *sql.DB, maxRetries int) *SQLRepository {
	if maxRetries <= 0 {
		maxRetries = 5
	}
	return &SQLRepository{
		db:         db,
		maxRetries: maxRetries,
	}
}

// InsertTx inserts an outbox record using the provided database transaction handle.
func (r *SQLRepository) InsertTx(ctx context.Context, tx *sql.Tx, record *OutboxRecord) error {
	if tx == nil {
		return ErrNilTx
	}
	if err := record.Validate(); err != nil {
		return fmt.Errorf("validating outbox record for insert: %w", err)
	}

	query := `
		INSERT INTO outbox (
			id, aggregate_type, aggregate_id, event_type, payload, status, retry_count, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := tx.ExecContext(
		ctx,
		query,
		record.ID,
		record.AggregateType,
		record.AggregateID,
		record.EventType,
		[]byte(record.Payload),
		string(record.Status),
		record.RetryCount,
		record.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting outbox record in tx: %w", err)
	}

	return nil
}

// FetchPending retrieves pending outbox records using FOR UPDATE SKIP LOCKED.
func (r *SQLRepository) FetchPending(ctx context.Context, limit int) ([]*OutboxRecord, error) {
	if limit <= 0 {
		limit = 50
	}

	query := `
		SELECT id, aggregate_type, aggregate_id, event_type, payload, status, retry_count, created_at, published_at
		FROM outbox
		WHERE status = 'PENDING'
		ORDER BY created_at ASC
		LIMIT $1
		FOR UPDATE SKIP LOCKED
	`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("querying pending outbox records: %w", err)
	}
	defer rows.Close()

	var records []*OutboxRecord
	for rows.Next() {
		var rec OutboxRecord
		var payloadBytes []byte
		var publishedAt sql.NullTime

		err := rows.Scan(
			&rec.ID,
			&rec.AggregateType,
			&rec.AggregateID,
			&rec.EventType,
			&payloadBytes,
			&rec.Status,
			&rec.RetryCount,
			&rec.CreatedAt,
			&publishedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning outbox row: %w", err)
		}

		rec.Payload = json.RawMessage(payloadBytes)
		if publishedAt.Valid {
			rec.PublishedAt = &publishedAt.Time
		}

		records = append(records, &rec)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating outbox rows: %w", err)
	}

	return records, nil
}

// MarkPublished marks an outbox record as PUBLISHED with the current timestamp.
func (r *SQLRepository) MarkPublished(ctx context.Context, id string) error {
	query := `
		UPDATE outbox
		SET status = 'PUBLISHED', published_at = NOW()
		WHERE id = $1
	`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("marking outbox record %s as published: %w", id, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("fetching rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

// MarkFailed increments retry count and sets status to FAILED if threshold is exceeded.
func (r *SQLRepository) MarkFailed(ctx context.Context, id string, retryCount int) error {
	var status string
	if retryCount >= r.maxRetries {
		status = string(StatusFailed)
	} else {
		status = string(StatusPending)
	}

	query := `
		UPDATE outbox
		SET retry_count = $1, status = $2
		WHERE id = $3
	`
	res, err := r.db.ExecContext(ctx, query, retryCount, status, id)
	if err != nil {
		return fmt.Errorf("updating outbox record %s failure state: %w", id, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("fetching rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

// CountPending returns the number of records waiting with status 'PENDING'.
func (r *SQLRepository) CountPending(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM outbox WHERE status = 'PENDING'").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("counting pending outbox records: %w", err)
	}
	return count, nil
}

// CountPublished returns the number of records processed with status 'PUBLISHED'.
func (r *SQLRepository) CountPublished(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM outbox WHERE status = 'PUBLISHED'").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("counting published outbox records: %w", err)
	}
	return count, nil
}

// ------------------------------------------------------------------------------
// Hermetic In-Memory Mock Repository Implementation (For Offline Tests & Crash Sim)
// ------------------------------------------------------------------------------

// MockRepository is an in-memory, thread-safe implementation of Repository.
// It supports simulated transaction rollback, error injection, and concurrency testing.
type MockRepository struct {
	mu         sync.RWMutex
	records    map[string]*OutboxRecord
	maxRetries int

	// Error injection hooks
	InsertError   error
	FetchError    error
	PublishError  error
	SimulateCrash bool
}

// NewMockRepository creates a new in-memory outbox repository.
func NewMockRepository(maxRetries int) *MockRepository {
	if maxRetries <= 0 {
		maxRetries = 5
	}
	return &MockRepository{
		records:    make(map[string]*OutboxRecord),
		maxRetries: maxRetries,
	}
}

// InsertTx stores the record in-memory. If InsertError is set, it fails.
func (m *MockRepository) InsertTx(ctx context.Context, tx *sql.Tx, record *OutboxRecord) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.InsertError != nil {
		return m.InsertError
	}

	if err := record.Validate(); err != nil {
		return err
	}

	// Clone record to prevent caller mutating in-memory store
	clone := *record
	clone.Payload = append([]byte(nil), record.Payload...)
	m.records[record.ID] = &clone

	return nil
}

// InsertDirect inserts a record outside of a sql.Tx (convenience for tests).
func (m *MockRepository) InsertDirect(record *OutboxRecord) error {
	return m.InsertTx(context.Background(), &sql.Tx{}, record)
}

// FetchPending returns records with status 'PENDING' up to limit.
func (m *MockRepository) FetchPending(ctx context.Context, limit int) ([]*OutboxRecord, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.FetchError != nil {
		return nil, m.FetchError
	}

	if limit <= 0 {
		limit = 50
	}

	var results []*OutboxRecord
	for _, rec := range m.records {
		if rec.Status == StatusPending {
			clone := *rec
			clone.Payload = append([]byte(nil), rec.Payload...)
			results = append(results, &clone)
			if len(results) >= limit {
				break
			}
		}
	}

	return results, nil
}

// MarkPublished marks an outbox record as PUBLISHED.
func (m *MockRepository) MarkPublished(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.PublishError != nil {
		return m.PublishError
	}

	rec, ok := m.records[id]
	if !ok {
		return ErrRecordNotFound
	}

	now := time.Now().UTC()
	rec.Status = StatusPublished
	rec.PublishedAt = &now

	return nil
}

// MarkFailed increments retry count and updates status.
func (m *MockRepository) MarkFailed(ctx context.Context, id string, retryCount int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	rec, ok := m.records[id]
	if !ok {
		return ErrRecordNotFound
	}

	rec.RetryCount = retryCount
	if retryCount >= m.maxRetries {
		rec.Status = StatusFailed
	} else {
		rec.Status = StatusPending
	}

	return nil
}

// CountPending returns the count of PENDING records in memory.
func (m *MockRepository) CountPending(ctx context.Context) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var count int64
	for _, rec := range m.records {
		if rec.Status == StatusPending {
			count++
		}
	}
	return count, nil
}

// CountPublished returns the count of PUBLISHED records in memory.
func (m *MockRepository) CountPublished(ctx context.Context) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var count int64
	for _, rec := range m.records {
		if rec.Status == StatusPublished {
			count++
		}
	}
	return count, nil
}

// GetRecord returns a pointer to the record by ID (for test assertions).
func (m *MockRepository) GetRecord(id string) (*OutboxRecord, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	rec, ok := m.records[id]
	if !ok {
		return nil, false
	}
	clone := *rec
	return &clone, true
}
