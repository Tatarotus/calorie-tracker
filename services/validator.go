package services

import (
	"strings"

	"calorie-tracker/models"
)

// SemanticValidator implements linguistic and ingredient safety protection logic.
type SemanticValidator struct {
	llm           *LLMService
	synonymMapper *SynonymMapper
}

// NewSemanticValidator creates a new SemanticValidator.
func NewSemanticValidator(llm *LLMService, mapper *SynonymMapper) *SemanticValidator {
	if mapper == nil {
		mapper = NewSynonymMapper()
	}
	return &SemanticValidator{
		llm:           llm,
		synonymMapper: mapper,
	}
}

// CategoryCompatibility defines allowed taxonomy transitions.
var CategoryCompatibility = map[models.FoodType][]models.FoodType{
	models.FoodTypeBeverage:     {models.FoodTypeBeverage, models.FoodTypeDairy, models.FoodTypeComposite},
	models.FoodTypeDairy:        {models.FoodTypeDairy, models.FoodTypeBeverage, models.FoodTypeComposite},
	models.FoodTypeProtein:      {models.FoodTypeProtein, models.FoodTypeComposite, models.FoodTypePreparedMeal},
	models.FoodTypeGrain:        {models.FoodTypeGrain, models.FoodTypeComposite, models.FoodTypePreparedMeal},
	models.FoodTypeComposite:    {models.FoodTypeComposite, models.FoodTypePreparedMeal, models.FoodTypeBeverage, models.FoodTypeDairy, models.FoodTypeProtein, models.FoodTypeGrain, models.FoodTypeVegetable},
	models.FoodTypeFruit:        {models.FoodTypeFruit, models.FoodTypeComposite},
	models.FoodTypeVegetable:    {models.FoodTypeVegetable, models.FoodTypeComposite, models.FoodTypePreparedMeal},
	models.FoodTypePreparedMeal: {models.FoodTypePreparedMeal, models.FoodTypeComposite, models.FoodTypeProtein, models.FoodTypeGrain, models.FoodTypeVegetable},
}

// Validate executes strict match category validation and unrealistic macro density checks.
func (v *SemanticValidator) Validate(queryType, matchedType models.FoodType, resolved *models.ReferenceFood) (bool, []string) {
	if resolved == nil {
		return false, []string{"nil resolved reference food"}
	}

	if ok, reasons := v.validateMacros(resolved); !ok {
		return false, reasons
	}

	return v.validateTaxonomy(queryType, matchedType)
}

func (v *SemanticValidator) validateMacros(resolved *models.ReferenceFood) (bool, []string) {
	if resolved.Macros.Calories < 0 || resolved.Macros.Protein < 0 || resolved.Macros.Carbs < 0 || resolved.Macros.Fat < 0 {
		return false, []string{"unrealistic negative macronutrient values detected"}
	}

	unit := strings.ToLower(strings.TrimSpace(resolved.Unit))
	isMassOrVolume := unit == "g" || unit == "gram" || unit == "grams" || unit == "gr" || unit == "ml" || unit == "milliliter" || unit == "milliliters" || unit == "l" || unit == "liter" || unit == "liters" || unit == "oz" || unit == "ounce" || unit == "ounces"
	if resolved.BaseQuantity > 0 && isMassOrVolume {
		caloriesPer100g := (resolved.Macros.Calories / resolved.BaseQuantity) * 100.0
		if caloriesPer100g > 950.0 {
			return false, []string{"unrealistic calorie density detected (> 950 kcal/100g is physically impossible)"}
		}
	}
	return true, nil
}

func (v *SemanticValidator) validateTaxonomy(queryType, matchedType models.FoodType) (bool, []string) {
	if queryType == "" || matchedType == "" || queryType == "auto" || matchedType == "auto" || queryType == "legacy" || matchedType == "legacy" {
		return true, []string{"uncertain taxonomy due to unclassified food type"}
	}

	allowedTypes, exists := CategoryCompatibility[queryType]
	if !exists {
		return true, []string{"uncertain taxonomy: unknown query food type"}
	}

	for _, t := range allowedTypes {
		if t == matchedType {
			return true, nil
		}
	}

	return false, []string{"food type mismatch: query category '" + string(queryType) + "' is incompatible with matched category '" + string(matchedType) + "'"}
}

// SemanticTokenCheck checks if there's a strong name compatibility.
func (v *SemanticValidator) SemanticTokenCheck(parsedName, resolvedName string) bool {
	norm := NewNormalizer()
	pNorm := norm.Normalize(parsedName)
	rNorm := norm.Normalize(resolvedName)

	if pNorm == rNorm {
		return true
	}

	if strings.Contains(rNorm, pNorm) || strings.Contains(pNorm, rNorm) {
		return true
	}

	pWords := strings.Fields(pNorm)
	rWords := strings.Fields(rNorm)

	isStopWord := func(w string) bool {
		stops := map[string]bool{
			"com": true, "de": true, "e": true, "with": true, "and": true, "or": true, "in": true, "on": true, "at": true, "a": true, "o": true, "da": true, "do": true, "para": true,
		}
		return stops[w] || len(w) <= 1
	}

	for _, pw := range pWords {
		if isStopWord(pw) {
			continue
		}
		for _, rw := range rWords {
			if isStopWord(rw) {
				continue
			}
			if pw == rw {
				return true
			}
		}
	}

	return false
}
