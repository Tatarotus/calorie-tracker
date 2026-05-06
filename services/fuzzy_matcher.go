package services

import (
	"math"
	"strings"
)

// FuzzyMatcher provides fuzzy string matching for food names
type FuzzyMatcher struct {
	threshold float64
}

// NewFuzzyMatcher creates a new fuzzy matcher with the given similarity threshold (0.0 to 1.0)
func NewFuzzyMatcher(threshold float64) *FuzzyMatcher {
	if threshold <= 0 || threshold > 1 {
		threshold = 0.8
	}
	return &FuzzyMatcher{threshold: threshold}
}

// levenshteinDistance calculates the edit distance between two strings
func levenshteinDistance(s1, s2 string) int {
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	// Create a 2D slice for dynamic programming
	prev := make([]int, len(s2)+1)
	curr := make([]int, len(s2)+1)

	for j := 0; j <= len(s2); j++ {
		prev[j] = j
	}

	for i := 1; i <= len(s1); i++ {
		curr[0] = i
		for j := 1; j <= len(s2); j++ {
			cost := 0
			if s1[i-1] != s2[j-1] {
				cost = 1
			}
			curr[j] = min(prev[j]+1, curr[j-1]+1, prev[j-1]+cost)
		}
		// Swap slices for next iteration
		prev, curr = curr, prev
	}

	return prev[len(s2)]
}

// similarityScore returns a similarity score between 0.0 and 1.0
func (m *FuzzyMatcher) similarityScore(s1, s2 string) float64 {
	s1 = strings.ToLower(strings.TrimSpace(s1))
	s2 = strings.ToLower(strings.TrimSpace(s2))

	if s1 == s2 {
		return 1.0
	}

	maxLen := math.Max(float64(len(s1)), float64(len(s2)))
	if maxLen == 0 {
		return 0.0
	}

	distance := levenshteinDistance(s1, s2)
	return 1.0 - float64(distance)/maxLen
}

// IsMatch returns true if the two strings are similar enough
func (m *FuzzyMatcher) IsMatch(s1, s2 string) bool {
	return m.similarityScore(s1, s2) >= m.threshold
}

// FindBestMatch finds the best matching string from candidates and returns it with its score
func (m *FuzzyMatcher) FindBestMatch(query string, candidates []string) (string, float64) {
	var bestMatch string
	var bestScore float64

	for _, candidate := range candidates {
		score := m.similarityScore(query, candidate)
		if score > bestScore {
			bestScore = score
			bestMatch = candidate
		}
	}

	return bestMatch, bestScore
}

// min returns the minimum of multiple integers
func min(a ...int) int {
	m := a[0]
	for _, v := range a[1:] {
		if v < m {
			m = v
		}
	}
	return m
}
