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

	unitDisplay := ref.Unit
	switch ref.Unit {
	case "gram":
		unitDisplay = "g"
	case "cup":
		unitDisplay = "cup"
	case "tablespoon":
		unitDisplay = "tbsp"
	case "teaspoon":
		unitDisplay = "tsp"
	case "ounce":
		unitDisplay = "oz"
	case "pound":
		unitDisplay = "lb"
	case "liter":
		unitDisplay = "L"
	case "unit", "unidade":
		unitDisplay = ""
	}

	var desc string
	if unitDisplay == "" {
		desc = fmt.Sprintf("%.0f %s", amount, ref.Name)
	} else {
		desc = fmt.Sprintf("%.1f%s %s", amount, unitDisplay, ref.Name)
	}

	return models.FoodPreview{
		Description: desc,
		Calories:    ref.Macros.Calories * scale,
		Protein:     ref.Macros.Protein * scale,
		Carbs:       ref.Macros.Carbs * scale,
		Fat:         ref.Macros.Fat * scale,
	}
}
