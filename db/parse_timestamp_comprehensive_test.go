package db

import (
	"testing"
)

// TestParseTimestampComprehensive tests the parseTimestamp function comprehensively
func TestParseTimestampComprehensive(t *testing.T) {
	// Test with exactly 10 characters (boundary condition)
	ts := "2023-01-01"
	result := parseTimestamp(ts)
	if result.IsZero() {
		t.Errorf("Expected non-zero time for exactly 10 characters, got zero time")
	}

	// Test with less than 10 characters (should return zero time)
	ts = "2023-01"
	result = parseTimestamp(ts)
	if !result.IsZero() {
		t.Errorf("Expected zero time for less than 10 characters, got %v", result)
	}

	// Test with exactly 9 characters
	ts = "2023-01-1"
	result = parseTimestamp(ts)
	if !result.IsZero() {
		t.Errorf("Expected zero time for 9 characters, got %v", result)
	}
}
