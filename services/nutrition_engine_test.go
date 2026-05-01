package services

import (
	"testing"
)

func TestNutritionEngine_EstimateGrams(t *testing.T) {
	// e := NewNutritionEngine(nil, nil)
	e := &NutritionEngine{}

	tests := []struct {
		name     string
		parsed   ParsedFood
		expected float64
	}{
		{
			name: "default cup",
			parsed: ParsedFood{Amount: 1, Unit: "cup", Name: "water"},
			expected: 240,
		},
		{
			name: "cup of rice (override)",
			parsed: ParsedFood{Amount: 1, Unit: "cup", Name: "arroz branco"},
			expected: 158,
		},
		{
			name: "tablespoon default",
			parsed: ParsedFood{Amount: 2, Unit: "tablespoon", Name: "sugar"},
			expected: 30, // 2 * 15
		},
		{
			name: "tablespoon oil (override)",
			parsed: ParsedFood{Amount: 2, Unit: "tablespoon", Name: "azeite"},
			expected: 27, // 2 * 13.5
		},
		{
			name: "unit of egg (override)",
			parsed: ParsedFood{Amount: 3, Unit: "unit", Name: "ovo"},
			expected: 150, // 3 * 50
		},
		{
			name: "unknown unit defaults to 0",
			parsed: ParsedFood{Amount: 1, Unit: "unknown_unit", Name: "something"},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := e.estimateGrams(tt.parsed)
			if result != tt.expected {
				t.Errorf("estimateGrams() = %v, want %v", result, tt.expected)
			}
		})
	}
}
