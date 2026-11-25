// ==============================================================================
// NOVARIS: Distributed Payment Settlement & Adaptive Fraud Intelligence Platform
// UNIT TEST SUITE: Schema Static Analysis & Blueprint Conformance (Day 5)
// FILE: services/gateway/cmd/verifyschema/main_test.go
// ==============================================================================

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func findMigrationsDir(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	curr := wd
	for i := 0; i < 5; i++ {
		candidate := filepath.Join(curr, "migrations")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
		parent := filepath.Dir(curr)
		if parent == curr {
			break
		}
		curr = parent
	}
	t.Fatalf("could not locate 'migrations' directory starting from %s", wd)
	return ""
}

func TestMigrationFiles_ExistenceAndNaming(t *testing.T) {
	migDir := findMigrationsDir(t)

	// 1. Verify 001_initial_ledger_schema.sql exists
	file001 := filepath.Join(migDir, "001_initial_ledger_schema.sql")
	if _, err := os.Stat(file001); os.IsNotExist(err) {
		t.Fatalf("expected migration file %s to exist, but was missing", file001)
	}

	// 2. Verify 002_fraud_and_features.sql exists
	file002 := filepath.Join(migDir, "002_fraud_and_features.sql")
	if _, err := os.Stat(file002); os.IsNotExist(err) {
		t.Fatalf("expected migration file %s to exist, but was missing", file002)
	}

	// 3. Verify 003_idempotency_keys.sql exists
	file003 := filepath.Join(migDir, "003_idempotency_keys.sql")
	if _, err := os.Stat(file003); os.IsNotExist(err) {
		t.Fatalf("expected migration file %s to exist, but was missing", file003)
	}
}

func TestMigration003_IdempotencySchemaCompleteness(t *testing.T) {
	migDir := findMigrationsDir(t)
	target := filepath.Join(migDir, "003_idempotency_keys.sql")

	content, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("could not read %s: %v", target, err)
	}

	sqlUpper := strings.ToUpper(string(content))

	// Verify idempotency_keys table
	if !strings.Contains(sqlUpper, "TABLE IF NOT EXISTS IDEMPOTENCY_KEYS") && !strings.Contains(sqlUpper, "TABLE IDEMPOTENCY_KEYS") {
		t.Error("migration 003 missing CREATE TABLE statement for idempotency_keys")
	}

	// Verify CHECK constraint on status
	if !strings.Contains(sqlUpper, "CHECK (STATUS IN ('IN_PROGRESS', 'COMPLETED', 'FAILED'))") {
		t.Error("missing strict CHECK constraint on idempotency_keys.status")
	}

	// Verify SHA-256 request_hash column
	if !strings.Contains(sqlUpper, "REQUEST_HASH VARCHAR(64) NOT NULL") {
		t.Error("idempotency_keys.request_hash must be defined as VARCHAR(64) NOT NULL")
	}

	// Verify response_body JSONB column
	if !strings.Contains(sqlUpper, "RESPONSE_BODY JSONB") {
		t.Error("idempotency_keys.response_body must support JSONB storage format")
	}

	// Verify TTL index
	if !strings.Contains(sqlUpper, "INDEX IF NOT EXISTS IDX_IDEMP_EXPIRES_AT") {
		t.Error("missing index idx_idemp_expires_at for automated TTL eviction")
	}
}

func TestMigration002_SchemaCompleteness(t *testing.T) {
	migDir := findMigrationsDir(t)
	target := filepath.Join(migDir, "002_fraud_and_features.sql")

	content, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("could not read %s: %v", target, err)
	}

	sqlUpper := strings.ToUpper(string(content))

	// Verify required tables
	expectedTables := []string{"FRAUD_REVIEWS", "FEATURE_SNAPSHOTS"}
	for _, table := range expectedTables {
		if !strings.Contains(sqlUpper, "TABLE IF NOT EXISTS "+table) && !strings.Contains(sqlUpper, "TABLE "+table) {
			t.Errorf("migration 002 missing CREATE TABLE statement for %s", table)
		}
	}

	// Verify CHECK constraint on verdict
	if !strings.Contains(sqlUpper, "CHECK (VERDICT IN ('CONFIRMED_FRAUD', 'FALSE_POSITIVE'))") {
		t.Error("missing strict CHECK constraint on fraud_reviews.verdict")
	}

	// Verify JSONB data type for feature_data
	if !strings.Contains(sqlUpper, "FEATURE_DATA JSONB NOT NULL") {
		t.Error("feature_snapshots.feature_data must be defined as JSONB NOT NULL")
	}

	// Verify Composite Index with correct column ordering: account_id first, then computed_at DESC
	if !strings.Contains(sqlUpper, "INDEX IF NOT EXISTS IDX_FEATURE_SNAPSHOTS_ACCOUNT_TIME") {
		t.Error("missing composite index idx_feature_snapshots_account_time")
	}
	if !strings.Contains(sqlUpper, "ON FEATURE_SNAPSHOTS(ACCOUNT_ID, COMPUTED_AT DESC)") {
		t.Error("composite index must place account_id before computed_at DESC")
	}

	// Verify Foreign Key references
	if !strings.Contains(sqlUpper, "REFERENCES TRANSACTIONS(ID)") {
		t.Error("missing foreign key reference to transactions(id)")
	}
	if !strings.Contains(sqlUpper, "REFERENCES ACCOUNTS(ID)") {
		t.Error("missing foreign key reference to accounts(id)")
	}
}

func TestMaskConnectionURL(t *testing.T) {
	raw := "postgres://novaris_admin:super_secret_pwd@localhost:5432/novaris_db?sslmode=disable"
	masked := maskConnectionURL(raw)

	if strings.Contains(masked, "super_secret_pwd") {
		t.Errorf("maskConnectionURL leaked plaintext password in: %s", masked)
	}
	if !strings.Contains(masked, "*****@localhost:5432") {
		t.Errorf("unexpected masked format: %s", masked)
	}
}
