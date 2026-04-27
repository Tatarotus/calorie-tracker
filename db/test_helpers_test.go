package db

import (
	"testing"

	_ "modernc.org/sqlite"
)

// Test helpers
func setupTestDB(t *testing.T) *DB {
	t.Helper()
	db, err := NewTestDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test DB: %v", err)
	}
	return db
}

func cleanupTestDB(t *testing.T, db *DB) {
	t.Helper()
	if err := db.Close(); err != nil {
		t.Errorf("Failed to close test DB: %v", err)
	}
}
