package services

import (
	"calorie-tracker/models"
	"fmt"
	"math"
	"strings"
)

func parsedFoodFromItem(item ParsedFoodItem) ParsedFood {
	return ParsedFood{
		Amount: item.Quantity,
		Unit:   item.Unit,
		Name:   item.FoodName,
	}
}

func isSerpAPIProvider(provider NutritionProvider) bool {
	_, ok := provider.(*SerpAPIProvider)
	return ok
}

func macrosConsistent(a, b *models.ReferenceFood) bool {
	if a == nil || b == nil {
		return false
	}
	aPer100, okA := macrosPer100(a)
	bPer100, okB := macrosPer100(b)
	if !okA || !okB {
		return false
	}
	return closeEnough(aPer100.Calories, bPer100.Calories, 0.25) &&
		closeEnough(aPer100.Protein, bPer100.Protein, 0.35) &&
		closeEnough(aPer100.Carbs, bPer100.Carbs, 0.35) &&
		closeEnough(aPer100.Fat, bPer100.Fat, 0.35)
}

func hasUsefulMacroProfile(ref *models.ReferenceFood) bool {
	if ref == nil || ref.Macros.Calories <= 0 {
		return false
	}
	nonZeroMacros := 0
	if ref.Macros.Protein > 0 {
		nonZeroMacros++
	}
	if ref.Macros.Carbs > 0 {
		nonZeroMacros++
	}
	if ref.Macros.Fat > 0 {
		nonZeroMacros++
	}
	return nonZeroMacros >= 2
}

func macrosPer100(ref *models.ReferenceFood) (models.Macros, bool) {
	if ref.BaseQuantity <= 0 {
		return models.Macros{}, false
	}
	factor := 100.0 / ref.BaseQuantity
	return models.Macros{
		Calories: ref.Macros.Calories * factor,
		Protein:  ref.Macros.Protein * factor,
		Carbs:    ref.Macros.Carbs * factor,
		Fat:      ref.Macros.Fat * factor,
	}, true
}

func closeEnough(a, b, tolerance float64) bool {
	maxVal := math.Max(math.Abs(a), math.Abs(b))
	if maxVal == 0 {
		return true
	}
	return math.Abs(a-b)/maxVal <= tolerance
}

func hasPreparationMismatch(query, resolved string) bool {
	queryTerms := preparationTerms(query)
	if len(queryTerms) == 0 {
		return false
	}
	resolvedTerms := preparationTerms(resolved)
	for term := range queryTerms {
		if !resolvedTerms[term] {
			return true
		}
	}
	return false
}

func preparationTerms(value string) map[string]bool {
	norm := NewNormalizer().Normalize(value)
	terms := map[string][]string{
		"fried":   {"frito", "frita", "fritos", "fritas", "fried"},
		"boiled":  {"cozido", "cozida", "cozidos", "cozidas", "boiled", "cooked"},
		"grilled": {"grelhado", "grelhada", "grelhados", "grelhadas", "grilled"},
		"roasted": {"assado", "assada", "assados", "assadas", "roasted", "baked"},
	}
	found := make(map[string]bool)
	for canonical, aliases := range terms {
		for _, alias := range aliases {
			if strings.Contains(norm, alias) {
				found[canonical] = true
				break
			}
		}
	}
	return found
}

func checkUnrealisticData(ref *models.ReferenceFood) error {
	if ref == nil {
		return nil
	}
	if ref.Macros.Calories < 0 || ref.Macros.Protein < 0 || ref.Macros.Carbs < 0 || ref.Macros.Fat < 0 {
		return fmt.Errorf("unrealistic negative macronutrient values detected")
	}
	if ref.BaseQuantity > 0 {
		unit := strings.ToLower(strings.TrimSpace(ref.Unit))
		if isMassOrVolumeUnit(unit) {
			caloriesPer100g := (ref.Macros.Calories / ref.BaseQuantity) * 100.0
			if caloriesPer100g > 950.0 {
				return fmt.Errorf("semantic validation failed: unrealistic calorie density detected (> 950 kcal/100g is physically impossible)")
			}
		}
	}
	return nil
}

func isMassOrVolumeUnit(unit string) bool {
	switch unit {
	case "g", "gram", "grams", "gr", "ml", "milliliter", "milliliters", "l", "liter", "liters", "oz", "ounce", "ounces":
		return true
	default:
		return false
	}
}
