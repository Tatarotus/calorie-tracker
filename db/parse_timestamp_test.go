package db

import (
	"testing"
)

// TestParseTimestampBoundaryConditions tests edge cases for parseTimestamp function
func TestParseTimestampBoundaryConditions(t *testing.T) {
	// Test with exactly 10 characters (boundary condition)
	ts := "2023-01-01"
	result := parseTimestamp(ts)
	if result.IsZero() {
		t.Error("Expected non-zero time for exactly 10 characters")
	}

	// Test with less than 10 characters (should fail)
	// Test with exactly 9 characters (boundary condition)
	ts = "2023-01-1"
	result = parseTimestamp(ts)
	if !result.IsZero() {
		t.Error("Expected zero time for less than 10 characters")
	}
}
