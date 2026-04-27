package services

import (
	"calorie-tracker/models"
	"fmt"
)

// MacroCalculator handles pure deterministic scaling of nutritional values
type MacroCalculator struct{}

func NewMacroCalculator() *MacroCalculator {
	return &MacroCalculator{}
}

func (c *MacroCalculator) Scale(ref models.ReferenceFood, amount float64) models.FoodPreview {
	if amount <= 0 {
		amount = ref.BaseQuantity
	}

	scale := amount / ref.BaseQuantity

	desc := fmt.Sprintf("%.1f%s %s", amount, ref.Unit, ref.Name)
	if ref.Unit == "unit" || ref.Unit == "unidade" {
		desc = fmt.Sprintf("%.0f %s", amount, ref.Name)
	}

	return models.FoodPreview{
		Description: desc,
		Calories:    ref.Macros.Calories * scale,
		Protein:     ref.Macros.Protein * scale,
		Carbs:       ref.Macros.Carbs * scale,
		Fat:         ref.Macros.Fat * scale,
	}
}
