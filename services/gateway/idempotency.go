// ==============================================================================
// NOVARIS: Distributed Payment Settlement & Adaptive Fraud Intelligence Platform
// MODULE 1: API Gateway Service - Idempotent Request Engine (Day 10)
// FILE: services/gateway/idempotency.go
// ==============================================================================
//
// Architectural Principles & Invariants:
// 1. Exact-Once Execution Semantics:
//    In distributed financial systems, network timeouts and transient client disconnects
//    are inevitable. An operation must produce the EXACT SAME outcome regardless of how
//    many times it is retried across lossy networks.
//
// 2. Stripe-Compliant Idempotency Standard (draft-ietf-httpapi-idempotency-key-header):
//    - Mutating requests (POST /v1/payments) provide an `Idempotency-Key` header.
//    - Same Idempotency-Key in -> Same cached response out, every time.
//    - The financial ledger is charged EXACTLY ONCE; retries replay the original result.
//
// 3. Payload Fingerprint Integrity:
//    If a client reuses an idempotency key with a DIFFERENT payload (e.g. changing
//    amount from $10 to $500), the gateway detects the fingerprint mismatch (SHA-256),
//    rejects the request with HTTP 422 Unprocessable Entity, and prevents key spoofing.
//
// 4. In-Flight Concurrency Protection:
//    If a duplicate request arrives while the original is still actively being processed
//    in the database, the gateway returns HTTP 409 Conflict, instructing the caller
//    to back off rather than spawning duplicate concurrent settlement workers.
// ==============================================================================

package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/charan21042005/novaris/services/gateway/internal/auth"
)

// Standard Idempotency Lifecycle States
const (
	StatusInProgress = "IN_PROGRESS"
	StatusCompleted  = "COMPLETED"
	StatusFailed     = "FAILED"

	// HeaderIdempotencyKey is the standard IETF header for client idempotency tokens
	HeaderIdempotencyKey = "Idempotency-Key"

	// HeaderIdempotentReplayed indicates whether the response was served from cache
	HeaderIdempotentReplayed = "Idempotent-Replayed"
)

// Sentinel Errors for deterministic error handling across middleware and handlers
var (
	ErrIdempotencyKeyMissing        = errors.New("missing required Idempotency-Key header")
	ErrIdempotencyPayloadMismatch   = errors.New("idempotency key reused with mismatched request payload")
	ErrIdempotencyInProgress        = errors.New("a request with this idempotency key is currently in progress")
	ErrIdempotencyRecordNotFound    = errors.New("idempotency record not found")
)

// IdempotencyRecord represents a stored or cached idempotency state
type IdempotencyRecord struct {
	Key          string          `json:"key"`
	AccountID    string          `json:"account_id"`
	RequestPath  string          `json:"request_path"`
	RequestHash  string          `json:"request_hash"`
	Status       string          `json:"status"` // IN_PROGRESS, COMPLETED, FAILED
	ResponseCode int             `json:"response_code"`
	ResponseBody json.RawMessage `json:"response_body"`
	CreatedAt    time.Time       `json:"created_at"`
	ExpiresAt    time.Time       `json:"expires_at"`
}

// ComputePayloadHash calculates a deterministic SHA-256 hex digest of the raw request body.
func ComputePayloadHash(body []byte) string {
	hash := sha256.Sum256(body)
	return hex.EncodeToString(hash[:])
}

// IdempotencyStore defines the contract for persisting and checking idempotency records.
type IdempotencyStore interface {
	// LockOrGet attempts to atomically claim an idempotency key.
	// Returns:
	//   - record: Existing record if found, or newly initialized record.
	//   - isNew: True if this call claimed the key (proceed with execution). False if already exists.
	//   - err: ErrIdempotencyPayloadMismatch, ErrIdempotencyInProgress, or store error.
	LockOrGet(ctx context.Context, key, accountID, path, payloadHash string) (*IdempotencyRecord, bool, error)

	// SaveResponse marks the idempotency record as COMPLETED and caches the response.
	SaveResponse(ctx context.Context, key string, statusCode int, responseBody []byte) error

	// MarkFailed marks the idempotency record as FAILED (e.g. on unrecoverable internal errors).
	MarkFailed(ctx context.Context, key string) error
}

// ------------------------------------------------------------------------------
// 1. PostgreSQL Idempotency Store (Production Database Implementation)
// ------------------------------------------------------------------------------

// PostgresIdempotencyStore implements IdempotencyStore backed by PostgreSQL.
type PostgresIdempotencyStore struct {
	db *sql.DB
}

// NewPostgresIdempotencyStore initializes a PostgreSQL-backed idempotency store.
func NewPostgresIdempotencyStore(db *sql.DB) *PostgresIdempotencyStore {
	return &PostgresIdempotencyStore{db: db}
}

func (s *PostgresIdempotencyStore) LockOrGet(ctx context.Context, key, accountID, path, payloadHash string) (*IdempotencyRecord, bool, error) {
	if s.db == nil {
		return nil, false, fmt.Errorf("database connection is nil")
	}

	trimmedKey := strings.TrimSpace(key)
	if trimmedKey == "" {
		return nil, false, ErrIdempotencyKeyMissing
	}

	// Step 1: Atomic INSERT ... ON CONFLICT DO NOTHING to claim the key
	insertQuery := `
		INSERT INTO idempotency_keys (key, account_id, request_path, request_hash, status)
		VALUES ($1, $2, $3, $4, 'IN_PROGRESS')
		ON CONFLICT (key) DO NOTHING;
	`
	res, err := s.db.ExecContext(ctx, insertQuery, trimmedKey, accountID, path, payloadHash)
	if err != nil {
		return nil, false, fmt.Errorf("claiming idempotency key: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return nil, false, fmt.Errorf("reading rows affected: %w", err)
	}

	// If rowsAffected == 1, we successfully claimed the key!
	if rowsAffected == 1 {
		return &IdempotencyRecord{
			Key:         trimmedKey,
			AccountID:   accountID,
			RequestPath: path,
			RequestHash: payloadHash,
			Status:      StatusInProgress,
			CreatedAt:   time.Now().UTC(),
		}, true, nil
	}

	// Step 2: Key already exists! Fetch existing record to evaluate state
	selectQuery := `
		SELECT key, account_id, request_path, request_hash, status,
		       COALESCE(response_code, 0), COALESCE(response_body, '{}'::jsonb),
		       created_at, expires_at
		FROM idempotency_keys
		WHERE key = $1;
	`
	var rec IdempotencyRecord
	var rawBody []byte
	err = s.db.QueryRowContext(ctx, selectQuery, trimmedKey).Scan(
		&rec.Key, &rec.AccountID, &rec.RequestPath, &rec.RequestHash, &rec.Status,
		&rec.ResponseCode, &rawBody, &rec.CreatedAt, &rec.ExpiresAt,
	)
	if err != nil {
		return nil, false, fmt.Errorf("fetching existing idempotency record: %w", err)
	}
	rec.ResponseBody = json.RawMessage(rawBody)

	// Step 3: Payload fingerprint verification (Stripe anti-tampering rule)
	if rec.RequestHash != payloadHash {
		return &rec, false, ErrIdempotencyPayloadMismatch
	}

	// Step 4: Check execution status
	if rec.Status == StatusInProgress {
		return &rec, false, ErrIdempotencyInProgress
	}

	return &rec, false, nil
}

func (s *PostgresIdempotencyStore) SaveResponse(ctx context.Context, key string, statusCode int, responseBody []byte) error {
	if s.db == nil {
		return fmt.Errorf("database connection is nil")
	}

	updateQuery := `
		UPDATE idempotency_keys
		SET status = 'COMPLETED',
		    response_code = $1,
		    response_body = $2::jsonb
		WHERE key = $3;
	`
	_, err := s.db.ExecContext(ctx, updateQuery, statusCode, string(responseBody), key)
	if err != nil {
		return fmt.Errorf("saving idempotency response: %w", err)
	}
	return nil
}

func (s *PostgresIdempotencyStore) MarkFailed(ctx context.Context, key string) error {
	if s.db == nil {
		return fmt.Errorf("database connection is nil")
	}

	_, err := s.db.ExecContext(ctx, "UPDATE idempotency_keys SET status = 'FAILED' WHERE key = $1;", key)
	return err
}

// ------------------------------------------------------------------------------
// 2. In-Memory Idempotency Store (Zero-Dependency High-Speed Engine)
// ------------------------------------------------------------------------------

// InMemoryIdempotencyStore provides thread-safe in-memory idempotency tracking for tests & decoupled mode.
type InMemoryIdempotencyStore struct {
	mu      sync.RWMutex
	records map[string]*IdempotencyRecord
}

// NewInMemoryIdempotencyStore initializes an in-memory idempotency store.
func NewInMemoryIdempotencyStore() *InMemoryIdempotencyStore {
	return &InMemoryIdempotencyStore{
		records: make(map[string]*IdempotencyRecord),
	}
}

func (s *InMemoryIdempotencyStore) LockOrGet(ctx context.Context, key, accountID, path, payloadHash string) (*IdempotencyRecord, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	trimmedKey := strings.TrimSpace(key)
	if trimmedKey == "" {
		return nil, false, ErrIdempotencyKeyMissing
	}

	existing, exists := s.records[trimmedKey]
	if !exists {
		rec := &IdempotencyRecord{
			Key:         trimmedKey,
			AccountID:   accountID,
			RequestPath: path,
			RequestHash: payloadHash,
			Status:      StatusInProgress,
			CreatedAt:   time.Now().UTC(),
			ExpiresAt:   time.Now().UTC().Add(24 * time.Hour),
		}
		s.records[trimmedKey] = rec
		return rec, true, nil
	}

	// Payload integrity check
	if existing.RequestHash != payloadHash {
		return existing, false, ErrIdempotencyPayloadMismatch
	}

	// In-progress check
	if existing.Status == StatusInProgress {
		return existing, false, ErrIdempotencyInProgress
	}

	return existing, false, nil
}

func (s *InMemoryIdempotencyStore) SaveResponse(ctx context.Context, key string, statusCode int, responseBody []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	rec, exists := s.records[key]
	if !exists {
		return ErrIdempotencyRecordNotFound
	}

	rec.Status = StatusCompleted
	rec.ResponseCode = statusCode
	rec.ResponseBody = json.RawMessage(responseBody)
	return nil
}

func (s *InMemoryIdempotencyStore) MarkFailed(ctx context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	rec, exists := s.records[key]
	if !exists {
		return ErrIdempotencyRecordNotFound
	}
	rec.Status = StatusFailed
	return nil
}

// ------------------------------------------------------------------------------
// 3. Idempotency HTTP Middleware
// ------------------------------------------------------------------------------

// responseCapture wraps http.ResponseWriter to buffer status code and response body
type responseCapture struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
}

func newResponseCapture(w http.ResponseWriter) *responseCapture {
	return &responseCapture{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
		body:           &bytes.Buffer{},
	}
}

func (rc *responseCapture) WriteHeader(code int) {
	rc.statusCode = code
	rc.ResponseWriter.WriteHeader(code)
}

func (rc *responseCapture) Write(b []byte) (int, error) {
	rc.body.Write(b)
	return rc.ResponseWriter.Write(b)
}

// IdempotencyMiddleware enforces exact-once processing semantics on mutating payment requests.
func IdempotencyMiddleware(store IdempotencyStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only intercept mutating state-altering HTTP methods
			if r.Method != http.MethodPost && r.Method != http.MethodPut && r.Method != http.MethodPatch {
				next.ServeHTTP(w, r)
				return
			}

			// Extract Idempotency-Key header
			idempKey := strings.TrimSpace(r.Header.Get(HeaderIdempotencyKey))
			if idempKey == "" {
				// Reject mutating requests missing the Idempotency-Key header
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string]string{
					"error":   "missing_idempotency_key",
					"message": "Mutating payment operations require a unique 'Idempotency-Key' header.",
				})
				return
			}

			// Read and restore body bytes to compute canonical payload hash
			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			payloadHash := ComputePayloadHash(bodyBytes)
			accountID := "unknown" // resolved from context claims if present
			if claims, ok := r.Context().Value(auth.ClaimsContextKey).(*auth.Claims); ok && claims != nil {
				accountID = claims.MerchantID
			}

			// Atomically check or claim the idempotency key
			record, isNew, err := store.LockOrGet(r.Context(), idempKey, accountID, r.URL.Path, payloadHash)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				if errors.Is(err, ErrIdempotencyPayloadMismatch) {
					w.WriteHeader(http.StatusUnprocessableEntity) // 422
					_ = json.NewEncoder(w).Encode(map[string]interface{}{
						"error":             "idempotency_conflict",
						"message":           "Idempotency key was previously used with a different request payload.",
						"idempotency_key":   idempKey,
						"original_path":     record.RequestPath,
						"original_created":  record.CreatedAt,
					})
					return
				}

				if errors.Is(err, ErrIdempotencyInProgress) {
					w.WriteHeader(http.StatusConflict) // 409
					_ = json.NewEncoder(w).Encode(map[string]string{
						"error":           "idempotency_in_progress",
						"message":         "A request with this idempotency key is currently in progress. Please retry shortly.",
						"idempotency_key": idempKey,
					})
					return
				}

				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			// CASE 1: DUPLICATE COMPLETED REQUEST -> SERVE FROM CACHE!
			if !isNew && record.Status == StatusCompleted {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set(HeaderIdempotentReplayed, "true")
				w.WriteHeader(record.ResponseCode)
				_, _ = w.Write(record.ResponseBody)
				return
			}

			// CASE 2: BRAND NEW REQUEST -> EXECUTE SETTLEMENT PIPELINE!
			w.Header().Set(HeaderIdempotentReplayed, "false")
			capture := newResponseCapture(w)

			// Execute downstream handlers
			next.ServeHTTP(capture, r)

			// Cache response upon completion if successful
			if capture.statusCode >= 200 && capture.statusCode < 500 {
				_ = store.SaveResponse(r.Context(), idempKey, capture.statusCode, capture.body.Bytes())
			} else {
				_ = store.MarkFailed(r.Context(), idempKey)
			}
		})
	}
}
