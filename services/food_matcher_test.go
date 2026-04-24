package services

import (
	"testing"
)

func TestFoodMatcherRemoveAccents(t *testing.T) {
	matcher := NewFoodMatcher(nil)
	
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no accents",
			input:    "apple",
			expected: "apple",
		},
		{
			name:     "with accents",
			input:    "café",
			expected: "cafe",
		},
		{
			name:     "mixed",
			input:    "naïve résumé",
			expected: "naive resume",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matcher.removeAccents(tt.input)
			if result != tt.expected {
				t.Errorf("removeAccents(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFoodMatcherNormalizeUnit(t *testing.T) {
	matcher := NewFoodMatcher(nil)
	
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "cup",
			input:    "cup",
			expected: "cup",
		},
		{
			name:     "cups",
			input:    "cups",
			expected: "cup",
		},
		{
			name:     "tablespoon",
			input:    "tablespoon",
			expected: "tablespoon",
		},
		{
			name:     "tablespoons",
			input:    "tablespoons",
			expected: "tablespoon",
		},
		{
			name:     "teaspoon",
			input:    "teaspoon",
			expected: "teaspoon",
		},
		{
			name:     "teaspoons",
			input:    "teaspoons",
			expected: "teaspoon",
		},
		{
			name:     "gram",
			input:    "gram",
			expected: "gram",
		},
		{
			name:     "grams",
			input:    "grams",
			expected: "gram",
		},
		{
			name:     "ounce",
			input:    "ounce",
			expected: "ounce",
		},
		{
			name:     "ounces",
			input:    "ounces",
			expected: "ounce",
		},
		{
			name:     "pound",
			input:    "pound",
			expected: "pound",
		},
		{
			name:     "pounds",
			input:    "pounds",
			expected: "pound",
		},
		{
			name:     "unknown",
			input:    "unknown",
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matcher.normalizeUnit(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeUnit(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNewFoodMatcher(t *testing.T) {
	// This just tests that we can create a matcher
	matcher := NewFoodMatcher(nil)
	if matcher == nil {
		t.Fatal("Expected non-nil matcher")
	}
}

func TestFoodMatcherParse(t *testing.T) {
	matcher := NewFoodMatcher(nil)
	
	// Test parsing a simple string
	result := matcher.Parse("1 apple")
	// Just verify it doesn't panic and returns a struct
	if result.Amount != 0 || result.Unit != "" || result.Name != "" {
		t.Logf("Parse returned: quantity=%v, unit=%s, name=%s", result.Amount, result.Unit, result.Name)
	}
}
