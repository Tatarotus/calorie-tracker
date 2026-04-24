package utils

import (
	"testing"
	"time"
)

func TestFormatDate(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected string
	}{
		{
			name:     "standard date",
			input:    time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			expected: "2024-01-15",
		},
		{
			name:     "another date",
			input:    time.Date(2023, 12, 31, 23, 59, 59, 0, time.UTC),
			expected: "2023-12-31",
		},
		{
			name:     "zero time",
			input:    time.Time{},
			expected: "0001-01-01",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatDate(tt.input)
			if result != tt.expected {
				t.Errorf("FormatDate(%v) = %s, expected %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestBeginningOfDay(t *testing.T) {
	// Test with a specific time
	input := time.Date(2024, 1, 15, 14, 30, 45, 123456789, time.UTC)
	result := BeginningOfDay(input)

	expected := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	if !result.Equal(expected) {
		t.Errorf("BeginningOfDay(%v) = %v, expected %v", input, result, expected)
	}

	// Verify time components are zeroed
	if result.Hour() != 0 || result.Minute() != 0 || result.Second() != 0 || result.Nanosecond() != 0 {
		t.Errorf("BeginningOfDay did not zero time components: %v", result)
	}

	// Verify date components are preserved
	if result.Year() != input.Year() || result.Month() != input.Month() || result.Day() != input.Day() {
		t.Errorf("BeginningOfDay did not preserve date components")
	}
}
