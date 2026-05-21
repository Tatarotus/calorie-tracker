package services

import (
	"calorie-tracker/db"
	"calorie-tracker/models"
	"math"
	"testing"
)

func TestNutritionEngine_DeterministicParserCoverage(t *testing.T) {
	mockDB := db.NewMockDB()

	// Seed reference foods for deterministic parser verification
	mockDB.SeedReferenceFood(models.ReferenceFood{
		Name:         "banana",
		BaseQuantity: 100,
		Unit:         "gram",
		Macros: models.Macros{
			Calories: 89,
			Protein:  1.1,
		},
	})
	mockDB.SeedReferenceFood(models.ReferenceFood{
		Name:         "leite",
		BaseQuantity: 100,
		Unit:         "ml",
		Macros: models.Macros{
			Calories: 60,
			Protein:  3.2,
		},
	})
	mockDB.SeedReferenceFood(models.ReferenceFood{
		Name:         "cafe com leite",
		BaseQuantity: 100,
		Unit:         "ml",
		Macros: models.Macros{
			Calories: 40,
			Protein:  2,
		},
	})
	mockDB.SeedReferenceFood(models.ReferenceFood{
		Name:         "ovo",
		BaseQuantity: 1,
		Unit:         "unit",
		Macros: models.Macros{
			Calories: 70,
			Protein:  6.0,
		},
	})
	mockDB.SeedReferenceFood(models.ReferenceFood{
		Name:         "maca",
		BaseQuantity: 1,
		Unit:         "unit",
		Macros: models.Macros{
			Calories: 52,
			Protein:  0.3,
		},
	})
	mockDB.SeedReferenceFood(models.ReferenceFood{
		Name:         "arroz branco",
		BaseQuantity: 100,
		Unit:         "gram",
		Macros: models.Macros{
			Calories: 130,
			Protein:  2.7,
		},
	})
	mockDB.SeedReferenceFood(models.ReferenceFood{
		Name:         "pao de forma",
		BaseQuantity: 100,
		Unit:         "gram",
		Macros: models.Macros{
			Calories: 265,
			Protein:  9,
			Carbs:    49,
			Fat:      3.2,
		},
	})

	engine := NewNutritionEngine(mockDB, nil)

	testCases := []struct {
		input            string
		expectedCalories float64
		expectedDesc     string
	}{
		{"200g banana", 178, "200.0g banana"},
		{"300ml milk", 180, "300.0ml milk"},
		{"2 eggs", 140, "2.0 egg"},
		{"1 apple", 52, "1 apple"},
		{"150g rice", 195, "150.0g rice"},
		{"2 fatias de pão de forma", 132.5, "2 fatias de pão de forma"},
		{"200ml de café com leite", 80, "200.0ml café com leite"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			preview, err := engine.Analyze(tc.input)
			if err != nil {
				t.Fatalf("Analyze failed for %q: %v", tc.input, err)
			}
			if math.Abs(preview.Calories-tc.expectedCalories) > 0.01 {
				t.Errorf("Expected %.2f calories, got %.2f", tc.expectedCalories, preview.Calories)
			}
			if preview.Description != tc.expectedDesc {
				t.Errorf("Expected description %q, got %q", tc.expectedDesc, preview.Description)
			}
		})
	}
}

func createSeededMockDB() *db.MockDB {
	mockDB := db.NewMockDB()

	// Seed regression test data
	mockDB.SeedReferenceFood(models.ReferenceFood{
		Name:         "cafe com leite",
		BaseQuantity: 100,
		Unit:         "ml",
		Macros: models.Macros{
			Calories: 40,
			Protein:  2.0,
		},
	})
	mockDB.SeedReferenceFood(models.ReferenceFood{
		Name:         "macaxeira",
		BaseQuantity: 100,
		Unit:         "gram",
		Macros: models.Macros{
			Calories: 160,
			Protein:  1.5,
		},
	})
	mockDB.SeedReferenceFood(models.ReferenceFood{
		Name:         "ovo",
		BaseQuantity: 1,
		Unit:         "unit",
		Macros: models.Macros{
			Calories: 70,
			Protein:  6.0,
		},
	})
	return mockDB
}

func TestNutritionEngine_FullE2EPipelineRegression_Basic(t *testing.T) {
	mockDB := createSeededMockDB()
	engine := NewNutritionEngine(mockDB, nil)

	t.Run("200ml de cafe com leite", func(t *testing.T) {
		preview, err := engine.Analyze("200ml de café com leite")
		if err != nil {
			t.Fatalf("Analyze failed: %v", err)
		}
		if preview.Calories != 80 {
			t.Errorf("Expected 80 calories, got %f", preview.Calories)
		}
	})

	t.Run("1 copo de cafe com leite", func(t *testing.T) {
		preview, err := engine.Analyze("1 copo de café com leite")
		if err != nil {
			t.Fatalf("Analyze failed: %v", err)
		}
		expected := 97.6
		if math.Abs(preview.Calories-expected) > 0.01 {
			t.Errorf("Expected %.2f calories, got %.2f", expected, preview.Calories)
		}
	})

	t.Run("150g de macaxeira", func(t *testing.T) {
		preview, err := engine.Analyze("150g de macaxeira")
		if err != nil {
			t.Fatalf("Analyze failed: %v", err)
		}
		if preview.Calories != 240 {
			t.Errorf("Expected 240 calories, got %f", preview.Calories)
		}
	})

	t.Run("2 ovos e 100g de aipim", func(t *testing.T) {
		preview, err := engine.Analyze("2 ovos e 100g de aipim")
		if err != nil {
			t.Fatalf("Analyze failed: %v", err)
		}
		if preview.Calories != 300 {
			t.Errorf("Expected 300 calories, got %f", preview.Calories)
		}
	})
}

func TestNutritionEngine_FullE2EPipelineRegression_Overrides(t *testing.T) {
	mockDB := createSeededMockDB()
	engine := NewNutritionEngine(mockDB, nil)

	t.Run("User override precedence", func(t *testing.T) {
		banana, _ := engine.canonicalResolver.Resolve("banana")
		err := mockDB.SaveUserOverride(&models.UserOverrideEntry{
			CanonicalFoodID: banana.ID,
			ServingAmount:   1,
			ServingUnit:     "unit",
			Calories:        500,
			OverrideReason:  "Special dynamic scale override",
		})
		if err != nil {
			t.Fatalf("Failed to save user override: %v", err)
		}

		preview, err := engine.Analyze("1 banana")
		if err != nil {
			t.Fatalf("Analyze failed: %v", err)
		}
		if preview.Calories != 500 {
			t.Errorf("Expected 500 calories from user override, got %f", preview.Calories)
		}
	})
}

func TestNutritionEngine_FullE2EPipelineRegression_Rejection(t *testing.T) {
	mockDB := createSeededMockDB()

	t.Run("Match rejection threshold", func(t *testing.T) {
		mockDB.SeedReferenceFood(models.ReferenceFood{
			Name:         "black coffee",
			BaseQuantity: 100,
			Unit:         "ml",
			Macros: models.Macros{
				Calories: 2,
			},
		})

		db2 := db.NewMockDB()
		db2.SeedReferenceFood(models.ReferenceFood{
			Name:         "black coffee",
			BaseQuantity: 100,
			Unit:         "ml",
			Macros: models.Macros{
				Calories: 2,
			},
		})
		engine2 := NewNutritionEngine(db2, nil)
		preview2, err2 := engine2.Analyze("200ml cafe com leite")
		if err2 != nil {
			t.Fatalf("Analyze failed: %v", err2)
		}
		if preview2.Calories != 0 {
			t.Errorf("Expected 0 calories (black coffee rejected for missing milk), got %f", preview2.Calories)
		}
	})
}

func TestNutritionEngine_FullE2EPipelineRegression_Fallbacks(t *testing.T) {
	mockDB := createSeededMockDB()

	t.Run("FatSecret miss and SerpAPI fallback", func(t *testing.T) {
		serp := &stubNutritionProvider{
			ref: &models.ReferenceFood{
				Name:         "unique food",
				BaseQuantity: 100,
				Unit:         "gram",
				Macros: models.Macros{
					Calories: 150,
				},
			},
		}

		engineWithSerp := NewNutritionEngineWithProvider(mockDB, nil, serp)
		preview, err := engineWithSerp.Analyze("100g unique food")
		if err != nil {
			t.Fatalf("Analyze failed: %v", err)
		}
		if preview.Calories != 150 {
			t.Errorf("Expected 150 calories from SerpAPI provider fallback, got %f", preview.Calories)
		}
	})

	t.Run("Unresolved preview fallback", func(t *testing.T) {
		engine := NewNutritionEngine(mockDB, nil)
		preview, err := engine.Analyze("100g completely unknown food")
		if err != nil {
			t.Fatalf("Analyze failed: %v", err)
		}
		if preview.Calories != 0 {
			t.Errorf("Expected 0 calories for unresolved food, got %f", preview.Calories)
		}
		if preview.ResolutionTrace.ResolutionMethod != "unresolved_fallback" {
			t.Errorf("Expected resolution method unresolved_fallback, got %q", preview.ResolutionTrace.ResolutionMethod)
		}
	})
}
