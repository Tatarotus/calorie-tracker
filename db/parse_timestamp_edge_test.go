package db

import (
	"testing"
)

// TestParseTimestampEdgeCases tests edge cases for parseTimestamp function
func TestParseTimestampEdgeCases(t *testing.T) {
	// Test the boundary condition: len(ts) >= 10

	testCases := []struct {
		input    string
		expected bool // true if we expect a non-zero time
	}{
		{"2023-01-01", true},
		{"2023-01-1", false},
		{"2023-01-01 extra", true},
		{"abc", false},
		{"", false},
	}

	for _, tc := range testCases {
		result := parseTimestamp(tc.input)
		if tc.expected && result.IsZero() {
			t.Errorf("parseTimestamp(%s) returned zero time, expected non-zero", tc.input)
		}
		if !tc.expected && !result.IsZero() {
			t.Errorf("parseTimestamp(%s) returned non-zero time, expected zero", tc.input)
		}
	}
}

func TestParseTimestampFormats(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"2023-01-01T15:04:05Z", "2023-01-01 15:04:05"},
		{"2023-01-01 15:04:05", "2023-01-01 15:04:05"},
		{"2023-01-01", "2023-01-01 00:00:00"},
	}

	for _, tc := range testCases {
		result := parseTimestamp(tc.input)
		if result.IsZero() {
			t.Errorf("parseTimestamp(%s) failed to parse", tc.input)
			continue
		}
		formatted := result.UTC().Format("2006-01-02 15:04:05")
		if formatted != tc.expected {
			t.Errorf("parseTimestamp(%s) = %s, want %s", tc.input, formatted, tc.expected)
		}
	}
}
