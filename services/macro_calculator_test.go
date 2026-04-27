package services

import (
	"calorie-tracker/models"
	"testing"
)

func TestMacroCalculatorScale(t *testing.T) {
	calc := NewMacroCalculator()
	ref := models.ReferenceFood{
		Name:         "arroz",
		BaseQuantity: 100,
		Unit:         "gram",
		Macros: models.Macros{
			Calories: 130,
			Protein:  2.5,
			Carbs:    28,
			Fat:      0.3,
		},
	}

	t.Run("Standard scaling 2x", func(t *testing.T) {
		got := calc.Scale(ref, 200)
		if got.Calories != 260 {
			t.Errorf("Expected 260 calories, got %f", got.Calories)
		}
		if got.Protein != 5.0 {
			t.Errorf("Expected 5.0 protein, got %f", got.Protein)
		}
	})

	t.Run("Small scaling 0.5x", func(t *testing.T) {
		got := calc.Scale(ref, 50)
		if got.Calories != 65 {
			t.Errorf("Expected 65 calories, got %f", got.Calories)
		}
	})

	t.Run("Zero or negative amount uses base", func(t *testing.T) {
		got := calc.Scale(ref, 0)
		if got.Calories != 130 {
			t.Errorf("Expected base 130 calories, got %f", got.Calories)
		}
	})

	t.Run("Unit based scaling", func(t *testing.T) {
		unitRef := models.ReferenceFood{
			Name:         "ovo",
			BaseQuantity: 1,
			Unit:         "unit",
			Macros: models.Macros{
				Calories: 70,
			},
		}
		got := calc.Scale(unitRef, 3)
		if got.Calories != 210 {
			t.Errorf("Expected 210 calories for 3 units, got %f", got.Calories)
		}
	})
}
