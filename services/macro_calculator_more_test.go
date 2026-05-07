package services

import (
	"testing"

	"calorie-tracker/models"
)

func TestMacroCalculatorScale_UnitDisplays(t *testing.T) {
	calc := NewMacroCalculator()

	tests := []struct {
		name     string
		ref      models.ReferenceFood
		amount   float64
		wantDesc string
	}{
		{
			name:     "gram unit",
			ref:      models.ReferenceFood{Name: "rice", BaseQuantity: 100, Unit: "gram", Macros: models.Macros{Calories: 130}},
			amount:   200,
			wantDesc: "200.0g rice",
		},
		{
			name:     "cup unit",
			ref:      models.ReferenceFood{Name: "water", BaseQuantity: 1, Unit: "cup", Macros: models.Macros{Calories: 0}},
			amount:   2,
			wantDesc: "2.0cup water",
		},
		{
			name:     "tablespoon unit",
			ref:      models.ReferenceFood{Name: "oil", BaseQuantity: 1, Unit: "tablespoon", Macros: models.Macros{Calories: 120}},
			amount:   3,
			wantDesc: "3.0tbsp oil",
		},
		{
			name:     "teaspoon unit",
			ref:      models.ReferenceFood{Name: "sugar", BaseQuantity: 1, Unit: "teaspoon", Macros: models.Macros{Calories: 16}},
			amount:   2,
			wantDesc: "2.0tsp sugar",
		},
		{
			name:     "ounce unit",
			ref:      models.ReferenceFood{Name: "cheese", BaseQuantity: 1, Unit: "ounce", Macros: models.Macros{Calories: 110}},
			amount:   2,
			wantDesc: "2.0oz cheese",
		},
		{
			name:     "pound unit",
			ref:      models.ReferenceFood{Name: "beef", BaseQuantity: 1, Unit: "pound", Macros: models.Macros{Calories: 1100}},
			amount:   0.5,
			wantDesc: "0.5lb beef",
		},
		{
			name:     "liter unit",
			ref:      models.ReferenceFood{Name: "water", BaseQuantity: 1, Unit: "liter", Macros: models.Macros{Calories: 0}},
			amount:   2,
			wantDesc: "2.0L water",
		},
		{
			name:     "unit display empty",
			ref:      models.ReferenceFood{Name: "egg", BaseQuantity: 1, Unit: "unit", Macros: models.Macros{Calories: 70}},
			amount:   2,
			wantDesc: "2 egg",
		},
		{
			name:     "unidade display empty",
			ref:      models.ReferenceFood{Name: "ovo", BaseQuantity: 1, Unit: "unidade", Macros: models.Macros{Calories: 70}},
			amount:   1,
			wantDesc: "1 ovo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calc.Scale(tt.ref, tt.amount)
			if got.Description != tt.wantDesc {
				t.Errorf("Scale() description = %q, want %q", got.Description, tt.wantDesc)
			}
		})
	}
}

func TestMacroCalculatorScale_NegativeAmount(t *testing.T) {
	calc := NewMacroCalculator()
	ref := models.ReferenceFood{
		Name:         "rice",
		BaseQuantity: 100,
		Unit:         "gram",
		Macros:       models.Macros{Calories: 130, Protein: 2.5, Carbs: 28, Fat: 0.3},
	}

	got := calc.Scale(ref, -10)
	if got.Calories != 130 {
		t.Errorf("Expected base 130 calories for negative amount, got %f", got.Calories)
	}
}
