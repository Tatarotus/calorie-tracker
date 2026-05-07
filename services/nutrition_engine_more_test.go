package services

import (
	"testing"

	"calorie-tracker/db"
	"calorie-tracker/models"
)

func TestNutritionEngine_ValidateMacros(t *testing.T) {
	mockDB := db.NewMockDB()
	engine := NewNutritionEngine(mockDB, nil)

	tests := []struct {
		name    string
		macros  models.Macros
		wantErr bool
	}{
		{
			name:    "valid macros",
			macros:  models.Macros{Calories: 100, Protein: 5, Carbs: 20, Fat: 3},
			wantErr: false,
		},
		{
			name:    "negative calories",
			macros:  models.Macros{Calories: -1, Protein: 5, Carbs: 20, Fat: 3},
			wantErr: true,
		},
		{
			name:    "negative protein",
			macros:  models.Macros{Calories: 100, Protein: -1, Carbs: 20, Fat: 3},
			wantErr: true,
		},
		{
			name:    "negative carbs",
			macros:  models.Macros{Calories: 100, Protein: 5, Carbs: -1, Fat: 3},
			wantErr: true,
		},
		{
			name:    "negative fat",
			macros:  models.Macros{Calories: 100, Protein: 5, Carbs: 20, Fat: -1},
			wantErr: true,
		},
		{
			name:    "too high calories",
			macros:  models.Macros{Calories: 5001, Protein: 5, Carbs: 20, Fat: 3},
			wantErr: true,
		},
		{
			name:    "zero calories ok",
			macros:  models.Macros{Calories: 0, Protein: 0, Carbs: 0, Fat: 0},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := engine.validateMacros(tt.macros)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateMacros() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNutritionEngine_FormatDescription(t *testing.T) {
	mockDB := db.NewMockDB()
	engine := NewNutritionEngine(mockDB, nil)

	tests := []struct {
		name     string
		amount   float64
		unit     string
		foodName string
		expected string
	}{
		{"single unit", 1, "unit", "apple", "1 apple"},
		{"multiple units", 2, "unit", "apple", "2.0 apple"},
		{"grams", 100, "gram", "rice", "100.0g rice"},
		{"cups", 1, "cup", "water", "1.0cup water"},
		{"empty unit single", 1, "", "banana", "1 banana"},
		{"empty unit multiple", 2, "", "banana", "2.0 banana"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := engine.formatDescription(tt.amount, tt.unit, tt.foodName)
			if got != tt.expected {
				t.Errorf("formatDescription() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestNutritionEngine_ScaledAmount(t *testing.T) {
	mockDB := db.NewMockDB()
	engine := NewNutritionEngine(mockDB, nil)

	tests := []struct {
		name            string
		ref             models.ReferenceFood
		parsed          ParsedFood
		allowEstimation bool
		wantAmount      float64
		wantOK          bool
	}{
		{
			name:            "zero amount returns base quantity",
			ref:             models.ReferenceFood{BaseQuantity: 100, Unit: "gram"},
			parsed:          ParsedFood{Amount: 0, Unit: "gram"},
			allowEstimation: true,
			wantAmount:      100,
			wantOK:          true,
		},
		{
			name:            "matching units",
			ref:             models.ReferenceFood{BaseQuantity: 100, Unit: "gram"},
			parsed:          ParsedFood{Amount: 50, Unit: "gram"},
			allowEstimation: true,
			wantAmount:      50,
			wantOK:          true,
		},
		{
			name:            "empty parsed unit with gram ref",
			ref:             models.ReferenceFood{BaseQuantity: 100, Unit: "gram"},
			parsed:          ParsedFood{Amount: 50, Unit: ""},
			allowEstimation: true,
			wantAmount:      50,
			wantOK:          true,
		},
		{
			name:            "unit to gram estimation",
			ref:             models.ReferenceFood{BaseQuantity: 100, Unit: "gram"},
			parsed:          ParsedFood{Amount: 1, Unit: "unit", Name: "ovo"},
			allowEstimation: true,
			wantAmount:      50, // egg override
			wantOK:          true,
		},
		{
			name:            "no estimation allowed",
			ref:             models.ReferenceFood{BaseQuantity: 100, Unit: "gram"},
			parsed:          ParsedFood{Amount: 1, Unit: "unit", Name: "ovo"},
			allowEstimation: false,
			wantAmount:      0,
			wantOK:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := engine.scaledAmount(tt.ref, tt.parsed, tt.allowEstimation)
			if got != tt.wantAmount || ok != tt.wantOK {
				t.Errorf("scaledAmount() = (%v, %v), want (%v, %v)", got, ok, tt.wantAmount, tt.wantOK)
			}
		})
	}
}

func TestNutritionEngine_ResolveDeterministically(t *testing.T) {
	mockDB := db.NewMockDB()
	// Seed reference foods
	mockDB.SeedReferenceFood(models.ReferenceFood{
		Name:         "arroz branco",
		BaseQuantity: 100,
		Unit:         "gram",
		Macros:       models.Macros{Calories: 130, Protein: 2.7, Carbs: 28, Fat: 0.3},
	})

	engine := NewNutritionEngine(mockDB, nil)

	items := []ParsedFood{
		{Amount: 100, Unit: "gram", Name: "arroz branco"},
	}

	preview, ok, err := engine.resolveDeterministically(items)
	if err != nil {
		t.Fatalf("resolveDeterministically() error = %v", err)
	}
	if !ok {
		t.Fatal("resolveDeterministically() expected ok = true")
	}
	if preview == nil {
		t.Fatal("resolveDeterministically() expected non-nil preview")
	}
	if preview.Calories != 130 {
		t.Errorf("expected 130 calories, got %f", preview.Calories)
	}
}

func TestNutritionEngine_ResolveSingle_EmptyName(t *testing.T) {
	mockDB := db.NewMockDB()
	engine := NewNutritionEngine(mockDB, nil)

	preview, ok, err := engine.resolveSingle(ParsedFood{Name: ""})
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if ok {
		t.Error("expected ok = false for empty name")
	}
	if preview != nil {
		t.Error("expected nil preview for empty name")
	}
}

func TestNutritionEngine_Analyze_EmptyItems(t *testing.T) {
	mockDB := db.NewMockDB()
	engine := NewNutritionEngine(mockDB, nil)

	_, err := engine.Analyze("   ")
	if err == nil {
		t.Error("expected error for empty/whitespace input")
	}
}
