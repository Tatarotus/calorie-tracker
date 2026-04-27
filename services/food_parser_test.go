package services

import (
	"testing"
)

func TestFoodParserRemoveAccents(t *testing.T) {
	p := NewFoodParser()

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
			input:    "mãçã",
			expected: "maca",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := p.removeAccents(tt.input); got != tt.expected {
				t.Errorf("removeAccents() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestFoodParserNormalizeUnit(t *testing.T) {
	p := NewFoodParser()

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
			input:    "slice",
			expected: "slice",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := p.normalizeUnit(tt.input); got != tt.expected {
				t.Errorf("normalizeUnit() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestFoodParserParse(t *testing.T) {
	p := NewFoodParser()

	testCases := []struct {
		input    string
		expected ParsedFood
	}{
		{"1 apple", ParsedFood{Amount: 1, Unit: "", Name: "apple"}},
		{"1.5 cups milk", ParsedFood{Amount: 1.5, Unit: "cup", Name: "milk"}},
		{"apple", ParsedFood{Amount: 0, Unit: "", Name: "apple"}},
		{"100g arroz", ParsedFood{Amount: 100, Unit: "gram", Name: "arroz"}},
		{"2 unidades de ovo", ParsedFood{Amount: 2, Unit: "unit", Name: "ovo"}},
	}

	for _, tc := range testCases {
		result := p.Parse(tc.input)
		if result.Amount != tc.expected.Amount {
			t.Errorf("Parse(%q) amount = %f, want %f", tc.input, result.Amount, tc.expected.Amount)
		}
		if result.Unit != tc.expected.Unit {
			t.Errorf("Parse(%q) unit = %q, want %q", tc.input, result.Unit, tc.expected.Unit)
		}
		if result.Name != tc.expected.Name {
			t.Errorf("Parse(%q) name = %q, want %q", tc.input, result.Name, tc.expected.Name)
		}
	}
}
