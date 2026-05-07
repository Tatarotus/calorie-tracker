package services

import (
	"testing"
)

func TestNewFuzzyMatcher_DefaultThreshold(t *testing.T) {
	// Test with invalid threshold (should default to 0.8)
	m := NewFuzzyMatcher(0)
	if m.threshold != 0.8 {
		t.Errorf("Expected default threshold 0.8, got %f", m.threshold)
	}

	m = NewFuzzyMatcher(-1)
	if m.threshold != 0.8 {
		t.Errorf("Expected default threshold 0.8 for negative, got %f", m.threshold)
	}

	m = NewFuzzyMatcher(1.5)
	if m.threshold != 0.8 {
		t.Errorf("Expected default threshold 0.8 for >1, got %f", m.threshold)
	}

	// Test with valid threshold
	m = NewFuzzyMatcher(0.9)
	if m.threshold != 0.9 {
		t.Errorf("Expected threshold 0.9, got %f", m.threshold)
	}
}

func TestLevenshteinDistance_LongStrings(t *testing.T) {
	tests := []struct {
		s1       string
		s2       string
		expected int
	}{
		{"a", "b", 1},
		{"ab", "ba", 2},
		{"abc", "def", 3},
		{"hello world", "hello world", 0},
		{"hello world", "hello worl", 1},
	}

	for _, tt := range tests {
		t.Run(tt.s1+"_"+tt.s2, func(t *testing.T) {
			got := levenshteinDistance(tt.s1, tt.s2)
			if got != tt.expected {
				t.Errorf("levenshteinDistance(%q, %q) = %d, want %d", tt.s1, tt.s2, got, tt.expected)
			}
		})
	}
}

func TestFuzzyMatcherSimilarityScore_Whitespace(t *testing.T) {
	m := NewFuzzyMatcher(0.8)

	// Should normalize whitespace and case
	score1 := m.similarityScore("Apple", "apple")
	score2 := m.similarityScore("  Apple  ", "apple")

	if score1 != 1.0 {
		t.Errorf("Expected perfect match for 'Apple' vs 'apple', got %f", score1)
	}
	if score2 != 1.0 {
		t.Errorf("Expected perfect match for '  Apple  ' vs 'apple', got %f", score2)
	}
}

func TestFuzzyMatcherSimilarityScore_EmptyStrings(t *testing.T) {
	m := NewFuzzyMatcher(0.8)

	// Empty vs non-empty
	score := m.similarityScore("", "abc")
	if score != 0.0 {
		t.Errorf("Expected 0.0 for empty vs non-empty, got %f", score)
	}

	// Both empty
	score = m.similarityScore("", "")
	if score != 1.0 {
		t.Errorf("Expected 1.0 for both empty, got %f", score)
	}
}

func TestMin(t *testing.T) {
	tests := []struct {
		input    []int
		expected int
	}{
		{[]int{1, 2, 3}, 1},
		{[]int{3, 2, 1}, 1},
		{[]int{5, 5, 5}, 5},
		{[]int{-1, 0, 1}, -1},
		{[]int{10}, 10},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := min(tt.input...)
			if got != tt.expected {
				t.Errorf("min(%v) = %d, want %d", tt.input, got, tt.expected)
			}
		})
	}
}
