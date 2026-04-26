package services

import (
	"calorie-tracker/db"
	"testing"
)

// TestNormalizeNameComprehensive tests the normalizeName function comprehensively
func TestNormalizeNameComprehensive(t *testing.T) {
	mockDB := db.NewMockDB()
	matcher := NewFoodMatcher(mockDB)

	// Test the boundary condition and complex logic in normalizeName
	// This function has loops and conditions that could be mutated
	
	// Test with exactly 10 characters (boundary condition)
	result := matcher.normalizeName("2023-01-01")
	expected := "2023-01-01"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
	
	// Test with less than 10 characters (boundary condition)
	result = matcher.normalizeName("2023-01")
	expected = "2023-01"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
	
	// Test with exactly 9 characters (boundary condition)
	result = matcher.normalizeName("2023-01-1")
	expected = "2023-01-1"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
	
	// Test with filler words that should be removed (complex logic)
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