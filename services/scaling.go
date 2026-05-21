package services

import (
	"fmt"
	"strings"

	"calorie-tracker/models"
)

// MacroScaler manages portion conversions, density scaling (ml ↔ g), and description formatting.
type MacroScaler struct{}

// NewMacroScaler creates a new MacroScaler.
func NewMacroScaler() *MacroScaler {
	return &MacroScaler{}
}

// Scale computes the scaled macro nutrients and quantity for a parsed item based on its reference data.
func (s *MacroScaler) Scale(canonical *models.CanonicalFood, ref *models.ReferenceFood, quantity float64, unit string) (models.Macros, float64) {
	if ref == nil {
		return models.Macros{}, 0.0
	}

	qty := quantity
	if qty <= 0 {
		qty = ref.BaseQuantity
	}

	parsedUnit := normalizeUnit(unit)
	refUnit := normalizeUnit(ref.Unit)

	scale := qty / ref.BaseQuantity

	if parsedUnit != refUnit {
		switch refUnit {
		case "gram":
			if parsedUnit == "unit" && canonical.DefaultServingUnit == "gram" && canonical.DefaultServingAmount > 0 {
				scale = (qty * canonical.DefaultServingAmount) / ref.BaseQuantity
			} else {
				estimatedGrams := s.EstimateGrams(canonical.NormalizedName, qty, parsedUnit)
				if estimatedGrams > 0 {
					scale = estimatedGrams / ref.BaseQuantity
				} else if parsedUnit == "ml" && canonical.GramsPerML > 0 {
					scale = (qty * canonical.GramsPerML) / ref.BaseQuantity
				}
			}
		case "ml":
			if parsedUnit == "gram" && canonical.GramsPerML > 0 {
				scale = (qty / canonical.GramsPerML) / ref.BaseQuantity
			} else {
				estimatedGrams := s.EstimateGrams(canonical.NormalizedName, qty, parsedUnit)
				if estimatedGrams > 0 {
					mlVal := estimatedGrams
					if canonical.GramsPerML > 0 {
						mlVal = estimatedGrams / canonical.GramsPerML
					}
					scale = mlVal / ref.BaseQuantity
				}
			}
		}
	}

	return models.Macros{
		Calories: ref.Macros.Calories * scale,
		Protein:  ref.Macros.Protein * scale,
		Carbs:    ref.Macros.Carbs * scale,
		Fat:      ref.Macros.Fat * scale,
	}, qty
}

// FormatDescription standardizes portion text descriptions.
func (s *MacroScaler) FormatDescription(amount float64, unit string, name string) string {
	if unit == "unit" || unit == "" {
		if amount == 1 {
			return fmt.Sprintf("%g %s", amount, name)
		}
		return fmt.Sprintf("%.1f %s", amount, name)
	}
	if unit == "gram" {
		return fmt.Sprintf("%.1fg %s", amount, name)
	}
	if unit == "slice" {
		if amount == 1 {
			return fmt.Sprintf("%g fatia de %s", amount, name)
		}
		return fmt.Sprintf("%g fatias de %s", amount, name)
	}
	return fmt.Sprintf("%.1f%s %s", amount, unit, name)
}

// EstimateGrams approximates the weight in grams for portion unit types.
func (s *MacroScaler) EstimateGrams(name string, amount float64, unit string) float64 {
	if amount <= 0 {
		amount = 1
	}

	rule, ok := globalRules.Units[unit]
	if !ok {
		return 0
	}

	for _, override := range rule.Overrides {
		for _, term := range override.Terms {
			if strings.Contains(name, term) {
				return amount * override.Value
			}
		}
	}
	return amount * rule.Default
}
