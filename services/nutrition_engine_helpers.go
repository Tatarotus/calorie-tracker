package services

import (
	"calorie-tracker/models"
	"fmt"
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

func (r *HybridNutritionResolver) fastBypassCheck(canonical *models.CanonicalFood, item ParsedFoodItem, trace *models.ResolutionTrace) (*models.ReferenceFood, bool, error) {
	override, err := r.cacheResolver.GetOverride(canonical.ID)
	if err == nil && override != nil {
		trace.ResolutionMethod = "user_override"
		trace.SourceType = "user_override"
		trace.SourceConfidence = 1.0
		trace.ValidationTriggered = false

		ref := &models.ReferenceFood{
			Name:         canonical.NormalizedName,
			BaseQuantity: override.ServingAmount,
			Unit:         override.ServingUnit,
			Macros: models.Macros{
				Calories: override.Calories,
				Protein:  override.Protein,
				Carbs:    override.Carbs,
				Fat:      override.Fat,
			},
		}
		return ref, true, nil
	}

	cacheEntry, err := r.cacheResolver.Get(canonical.ID, item.Unit)
	if err == nil && cacheEntry != nil {
		isExpired := r.ttlPolicy.IsExpired(cacheEntry)
		if !isExpired && cacheEntry.SourceConfidence >= 0.95 {
			trace.ResolutionMethod = "local_cache"
			trace.SourceType = cacheEntry.SourceType
			trace.SourceConfidence = cacheEntry.SourceConfidence
			trace.CacheHit = true
			trace.ValidationTriggered = false

			ref := &models.ReferenceFood{
				Name:         canonical.NormalizedName,
				BaseQuantity: cacheEntry.ServingAmount,
				Unit:         cacheEntry.ServingUnit,
				Macros: models.Macros{
					Calories: cacheEntry.Calories,
					Protein:  cacheEntry.Protein,
					Carbs:    cacheEntry.Carbs,
					Fat:      cacheEntry.Fat,
				},
			}
			return ref, true, nil
		}
	}
	return nil, false, nil
}

func (r *HybridNutritionResolver) refreshCacheInBackground(canonical *models.CanonicalFood, item ParsedFoodItem) {
	for _, provider := range r.providers {
		if isSerpAPIProvider(provider) {
			continue
		}
		ref, err := provider.ResolveFood(parsedFoodFromItem(item))
		if err != nil {
			continue
		}
		if ref == nil {
			continue
		}

		sourceType := "fatsecret"
		resolutionMethod := "fatsecret_api"
		confidence := 1.0

		if _, ok := provider.(*FatSecretProvider); ok {
			sourceType = "fatsecret"
			resolutionMethod = "fatsecret_api"
			confidence = 1.0
		} else if _, ok := provider.(*CalorieNinjasProvider); ok {
			sourceType = "calorieninjas"
			resolutionMethod = "calorieninjas_api"
			confidence = 0.95
		}

		candidate := &ResolutionCandidate{
			ReferenceFood:    ref,
			Confidence:       confidence,
			SourceType:       sourceType,
			ResolutionMethod: resolutionMethod,
		}

		ok, evalErr := r.evaluateCandidate(canonical, candidate)
		if evalErr == nil && ok {
			// Update the cache silently
			r.updateTraceAndCache(canonical, candidate, &models.ResolutionTrace{})
			break
		}
	}
}
