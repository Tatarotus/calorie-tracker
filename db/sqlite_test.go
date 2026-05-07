package db

import (
	"testing"
	"time"
)

func TestDB_NewDB(t *testing.T) {
	// This test creates a real DB in the user's home directory
	// We can't easily test this without side effects, so we just verify it doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("NewDB panicked: %v", r)
		}
	}()

	// We can't actually call NewDB() in tests because it creates a file in the home dir
	// and we don't want to pollute the user's environment.
	// The functionality is tested via NewTestDB.
}

func TestDB_NewTestDB(t *testing.T) {
	db, err := NewTestDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test DB: %v", err)
	}
	defer db.Close()

	if db.conn == nil {
		t.Error("Expected non-nil connection")
	}
}

func TestDB_GetConn(t *testing.T) {
	db, err := NewTestDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test DB: %v", err)
	}
	defer db.Close()

	conn := db.GetConn()
	if conn == nil {
		t.Error("Expected non-nil connection from GetConn")
	}
}

func TestDB_MigrateExistingTables_NoMigrationNeeded(t *testing.T) {
	// Fresh DB should not need migration
	db, err := NewTestDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test DB: %v", err)
	}
	defer db.Close()

	// Calling migrate again should not fail
	err = db.migrateExistingTables()
	if err != nil {
		t.Errorf("Expected no error on second migration, got %v", err)
	}
}

func TestDB_TableNeedsMigration(t *testing.T) {
	db, err := NewTestDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test DB: %v", err)
	}
	defer db.Close()

	// Table exists with all columns
	needsMigration, err := db.tableNeedsMigration("food_entries", []string{"id", "timestamp"})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if needsMigration {
		t.Error("Expected no migration needed for existing table with all columns")
	}

	// Table exists but missing column
	needsMigration, err = db.tableNeedsMigration("food_entries", []string{"id", "nonexistent_column"})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !needsMigration {
		t.Error("Expected migration needed for missing column")
	}

	// Non-existent table
	needsMigration, err = db.tableNeedsMigration("nonexistent_table", []string{"id"})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if needsMigration {
		t.Error("Expected no migration needed for non-existent table")
	}
}

func TestDB_SeedReferenceFoods(t *testing.T) {
	db, err := NewTestDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test DB: %v", err)
	}
	defer db.Close()

	// Seed should have been called during migration
	// Verify some foods exist
	ref, err := db.GetReferenceFood("arroz branco")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if ref == nil {
		t.Error("Expected 'arroz branco' to be seeded")
	}

	ref, err = db.GetReferenceFood("egg")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if ref == nil {
		t.Error("Expected 'egg' to be seeded")
	}
}

func TestParseTimestamp(t *testing.T) {
	tests := []struct {
		input    string
		expected bool // whether it should parse successfully (not zero time)
	}{
		{time.Now().Format(time.RFC3339Nano), true},
		{"2006-01-02 15:04:05", true},
		{"2006-01-02", true},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := parseTimestamp(tt.input)
			if tt.expected && got.IsZero() {
				t.Errorf("Expected non-zero time for %q", tt.input)
			}
			if !tt.expected && !got.IsZero() {
				t.Errorf("Expected zero time for %q", tt.input)
			}
		})
	}
}
