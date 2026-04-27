package services

import (
	"calorie-tracker/db"
	"testing"
)

func TestFoodMatcher_normalizeName(t *testing.T) {
	mockDB := db.NewMockDB()
	matcher := NewFoodMatcher(mockDB)

	testCases := []struct {
		input    string
		expected string
	}{
		{"of the apple", "apple"},
		{"a cup of tea", "cup tea"},
		{"an orange and a banana", "orange banana"},
		{"butter with salt", "butter with salt"},
		{"CAFÉ com leite", "cafe com leite"}, // accent removal + lowercase
	}

	for _, tc := range testCases {
		result := matcher.normalizeName(tc.input)
		if result != tc.expected {
			t.Errorf("normalizeName(%s) = %s, want %s", tc.input, result, tc.expected)
		}
	}
}

func TestFoodMatcher_normalizeUnit(t *testing.T) {
	mockDB := db.NewMockDB()
	matcher := NewFoodMatcher(mockDB)

	testCases := []struct {
		input    string
		expected string
	}{
		{"Grams", "gram"},
		{"CUPS", "cup"},
		{"Ounces", "ounce"},
		{"ml", "ml"},
		{"liters", "liter"},
		{"unknown", "unknown"},
	}

	for _, tc := range testCases {
		result := matcher.normalizeUnit(tc.input)
		if result != tc.expected {
			t.Errorf("normalizeUnit(%s) = %s, want %s", tc.input, result, tc.expected)
		}
	}
}

func TestFoodMatcher_Parse(t *testing.T) {
	mockDB := db.NewMockDB()
	matcher := NewFoodMatcher(mockDB)

	testCases := []struct {
		input    string
		expected ParsedFood
	}{
		{"2 cups of milk", ParsedFood{Amount: 2, Unit: "cup", Name: "milk"}},
		{"100g chicken breast", ParsedFood{Amount: 100, Unit: "gram", Name: "chicken breast"}},
		{"1.5 liters of water", ParsedFood{Amount: 1.5, Unit: "liter", Name: "water"}},
		{"apple", ParsedFood{Amount: 0, Unit: "", Name: "apple"}},
		{"cerca de 200g de arroz", ParsedFood{Amount: 200, Unit: "gram", Name: "arroz"}},
	}

	for _, tc := range testCases {
		result := matcher.Parse(tc.input)
		if result.Amount != tc.expected.Amount {
			t.Errorf("Parse(%s) amount = %f, want %f", tc.input, result.Amount, tc.expected.Amount)
		}
		if result.Unit != tc.expected.Unit {
			t.Errorf("Parse(%s) unit = %s, want %s", tc.input, result.Unit, tc.expected.Unit)
		}
		if result.Name != tc.expected.Name {
			t.Errorf("Parse(%s) name = %s, want %s", tc.input, result.Name, tc.expected.Name)
		}
	}
}
