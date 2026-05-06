package services

import (
	"testing"
)

func TestLevenshteinDistance(t *testing.T) {
	tests := []struct {
		s1       string
		s2       string
		expected int
	}{
		{"", "", 0},
		{"", "abc", 3},
		{"abc", "", 3},
		{"abc", "abc", 0},
		{"kitten", "sitting", 3},
		{"saturday", "sunday", 3},
		{"book", "back", 2},
		{"arroz", "arroz branco", 7},
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

func TestFuzzyMatcherSimilarityScore(t *testing.T) {
	m := NewFuzzyMatcher(0.8)

	tests := []struct {
		s1        string
		s2        string
		expected  float64
		tolerance float64
	}{
		{"apple", "apple", 1.0, 0.001},
		{"apple", "aple", 0.8, 0.1},
		{"banana", "bananna", 0.857, 0.01},
		{"rice", "arroz", 0.2, 0.1},
		{"", "", 1.0, 0.001},
	}

	for _, tt := range tests {
		t.Run(tt.s1+"_"+tt.s2, func(t *testing.T) {
			got := m.similarityScore(tt.s1, tt.s2)
			if got < tt.expected-tt.tolerance || got > tt.expected+tt.tolerance {
				t.Errorf("similarityScore(%q, %q) = %f, want %f (±%f)", tt.s1, tt.s2, got, tt.expected, tt.tolerance)
			}
		})
	}
}

func TestFuzzyMatcherIsMatch(t *testing.T) {
	m := NewFuzzyMatcher(0.8)

	tests := []struct {
		s1       string
		s2       string
		expected bool
	}{
		{"apple", "apple", true},
		{"apple", "aple", true},
		{"banana", "bananna", true},
		{"rice", "arroz", false},
		{"chicken breast", "chicken", false}, // Too different
	}

	for _, tt := range tests {
		t.Run(tt.s1+"_"+tt.s2, func(t *testing.T) {
			got := m.IsMatch(tt.s1, tt.s2)
			if got != tt.expected {
				t.Errorf("IsMatch(%q, %q) = %v, want %v", tt.s1, tt.s2, got, tt.expected)
			}
		})
	}
}

func TestFuzzyMatcherFindBestMatch(t *testing.T) {
	m := NewFuzzyMatcher(0.8)
	candidates := []string{"apple", "banana", "orange", "grape"}

	tests := []struct {
		query         string
		expectedMatch string
		minScore      float64
	}{
		{"aple", "apple", 0.8},
		{"bananna", "banana", 0.85},
		{"grapefruit", "grape", 0.5},
		{"xyz", "", 0.0}, // No good match
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			match, score := m.FindBestMatch(tt.query, candidates)
			if match != tt.expectedMatch {
				t.Errorf("FindBestMatch(%q) match = %q, want %q", tt.query, match, tt.expectedMatch)
			}
			if score < tt.minScore {
				t.Errorf("FindBestMatch(%q) score = %f, want >= %f", tt.query, score, tt.minScore)
			}
		})
	}
}
