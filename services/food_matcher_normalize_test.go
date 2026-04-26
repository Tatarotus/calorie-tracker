package services

import (
	"calorie-tracker/db"
	"testing"
)

// TestNormalizeNameBoundaryConditions tests edge cases for normalizeName function
func TestNormalizeNameBoundaryConditions(t *testing.T) {
	mockDB := db.NewMockDB()
	matcher := NewFoodMatcher(mockDB)

	// Test with exactly 10 characters
	result := matcher.normalizeName("2023-01-01")
	expected := "2023-01-01"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}

	// Test with less than 10 characters
	result = matcher.normalizeName("2023-01")
	expected = "2023-01"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}

	// Test with exactly 9 characters
	result = matcher.normalizeName("2023-01-1")
	expected = "2023-01-1"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

// TestNormalizeNameWithFillerWords tests the filler word removal logic
func TestNormalizeNameWithFillerWords(t *testing.T) {
	mockDB := db.NewMockDB()
	matcher := NewFoodMatcher(mockDB)

	// Test with filler words that should be removed
	testCases := []struct {
		input    string
		expected string
	}{
		{"of the apple", "apple"},
		{"a cup of tea", "cup tea"},
		{"an orange and a banana", "orange banana"},
		{"butter with salt", "butter with salt"}, // "with" is not in fillerWords
		{"no filler words here", "no filler words here"}, // none are filler words
	}

	for _, tc := range testCases {
		result := matcher.normalizeName(tc.input)
		if result != tc.expected {
			t.Errorf("Input '%s': expected '%s', got '%s'", tc.input, tc.expected, result)
		}
	}
}