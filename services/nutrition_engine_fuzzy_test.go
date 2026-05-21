package services

import (
	"calorie-tracker/db"
	"calorie-tracker/models"
	"testing"
)

func TestNutritionEngine_FuzzyCacheLookup(t *testing.T) {
	mockDB := db.NewMockDB()
	mockDB.SeedReferenceFood(models.ReferenceFood{
		Name:         "grilled chicken",
		BaseQuantity: 100,
		Unit:         "gram",
		Macros:       models.Macros{Calories: 165, Protein: 31, Carbs: 0, Fat: 3.6},
	})

	engine := NewNutritionEngine(mockDB, nil)

	// Test fuzzy matching - "chiken" should match "grilled chicken"
	preview, err := engine.Analyze("100g chiken")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if preview != nil {
		if preview.Calories != 165 {
			t.Errorf("expected 165 calories, got %f", preview.Calories)
		}
	}
}

func TestNutritionEngine_SynonymLookup(t *testing.T) {
	mockDB := db.NewMockDB()
	mockDB.SeedReferenceFood(models.ReferenceFood{
		Name:         "arroz branco",
		BaseQuantity: 100,
		Unit:         "gram",
		Macros:       models.Macros{Calories: 130, Protein: 2.7, Carbs: 28, Fat: 0.3},
	})

	engine := NewNutritionEngine(mockDB, nil)

	// Test synonym matching - "white rice" should match "arroz branco"
	preview, err := engine.Analyze("100g white rice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if preview != nil {
		if preview.Calories != 130 {
			t.Errorf("expected 130 calories, got %f", preview.Calories)
		}
	}
}

func TestNutritionEngine_ExpandedUnits(t *testing.T) {
	engine := &NutritionEngine{}

	tests := []struct {
		name     string
		parsed   ParsedFood
		expected float64
	}{
		{
			name:     "ounce default",
			parsed:   ParsedFood{Amount: 1, Unit: "ounce", Name: "cheese"},
			expected: 28.35,
		},
		{
			name:     "pound default",
			parsed:   ParsedFood{Amount: 1, Unit: "pound", Name: "flour"},
			expected: 453.59,
		},
		{
			name:     "ml default",
			parsed:   ParsedFood{Amount: 100, Unit: "ml", Name: "water"},
			expected: 100,
		},
		{
			name:     "liter default",
			parsed:   ParsedFood{Amount: 1, Unit: "liter", Name: "water"},
			expected: 1000,
		},
		{
			name:     "pinch default",
			parsed:   ParsedFood{Amount: 1, Unit: "pinch", Name: "salt"},
			expected: 0.6,
		},
		{
			name:     "dash default",
			parsed:   ParsedFood{Amount: 1, Unit: "dash", Name: "hot sauce"},
			expected: 0.5,
		},
		{
			name:     "ounce cheese override",
			parsed:   ParsedFood{Amount: 1, Unit: "ounce", Name: "cheese"},
			expected: 28.35,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.estimateGrams(tt.parsed.Name, tt.parsed.Amount, tt.parsed.Unit)
			if result != tt.expected {
				t.Errorf("estimateGrams() = %v, want %v", result, tt.expected)
			}
		})
	}
}
