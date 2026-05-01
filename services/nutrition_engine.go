package services

import (
	"calorie-tracker/data"
	"calorie-tracker/db"
	"calorie-tracker/models"
	"encoding/json"
	"fmt"
	"strings"
)

type ConversionOverride struct {
	Terms []string `json:"terms"`
	Value float64  `json:"value"`
}

type UnitRule struct {
	Default   float64              `json:"default"`
	Overrides []ConversionOverride `json:"overrides"`
}

type RulesData struct {
	Units map[string]UnitRule `json:"units"`
}

var globalRules RulesData

func init() {
	if err := json.Unmarshal(data.RulesJSON, &globalRules); err != nil {
		panic(fmt.Sprintf("failed to parse rules.json: %v", err))
	}
}

// NutritionEngine orchestrates the hybrid nutrition lookup flow
type NutritionEngine struct {
	db         db.DBProvider
	parser     *FoodParser
	calculator *MacroCalculator
	llm        *LLMService
	provider   NutritionProvider
}

func NewNutritionEngine(db db.DBProvider, llm *LLMService) *NutritionEngine {
	return NewNutritionEngineWithProvider(db, llm, nil)
}

func NewNutritionEngineWithProvider(db db.DBProvider, llm *LLMService, provider NutritionProvider) *NutritionEngine {
	return &NutritionEngine{
		db:         db,
		parser:     NewFoodParser(),
		calculator: NewMacroCalculator(),
		llm:        llm,
		provider:   provider,
	}
}

func (e *NutritionEngine) Analyze(description string) (*models.FoodPreview, error) {
	// Step 1: NLP-lite parsing into one or more structured food items.
	items := e.parser.ParseMeal(description)
	if len(items) == 0 {
		return nil, fmt.Errorf("could not parse food name from: %s", description)
	}

	preview, ok, err := e.resolveDeterministically(items)
	if err != nil {
		return nil, err
	}
	if ok {
		return preview, nil
	}

	if e.llm != nil {
		llmItems, err := e.llm.ParseFoodItems(description)
		if err != nil {
			return nil, err
		}
		if len(llmItems) > 0 {
			preview, ok, err := e.resolveDeterministically(llmItems)
			if err != nil {
				return nil, err
			}
			if ok {
				return preview, nil
			}
		}
	}

	if len(items) > 1 {
		return e.analyzeWholeMealWithLLMFallback(description)
	}

	return e.analyzeWithLLMFallback(description, items[0])
}

func (e *NutritionEngine) resolveDeterministically(items []ParsedFood) (*models.FoodPreview, bool, error) {
	total := models.FoodPreview{}
	descriptions := make([]string, 0, len(items))

	for _, item := range items {
		preview, ok, err := e.resolveSingle(item)
		if err != nil || !ok {
			return nil, ok, err
		}

		descriptions = append(descriptions, preview.Description)
		total.Calories += preview.Calories
		total.Protein += preview.Protein
		total.Carbs += preview.Carbs
		total.Fat += preview.Fat
	}

	total.Description = strings.Join(descriptions, " + ")
	return &total, true, nil
}

func (e *NutritionEngine) resolveSingle(parsed ParsedFood) (*models.FoodPreview, bool, error) {
	if parsed.Name == "" {
		return nil, false, nil
	}

	// 1. Exact Match Cache (Highest Priority)
	cached, err := e.db.GetCachedFood(parsed.Name)
	if err == nil && cached != nil {
		// For Cache, we only allow scaling if the units match exactly or ref has no unit.
		// We DON'T want to estimate grams for a cache hit if they specifically asked for 'unit'.
		if amount, ok := e.scaledAmount(*cached, parsed, false); ok {
			preview := e.calculator.Scale(*cached, amount)
			preview.Name = parsed.Name
			preview.Unit = parsed.Unit
			preview.Description = e.formatDescription(parsed.Amount, parsed.Unit, parsed.Name)
			return &preview, true, nil
		}
	}

	// 2. Exact Match Reference
	ref, err := e.db.GetReferenceFood(parsed.Name)
	if err == nil && ref != nil && strings.EqualFold(ref.Name, parsed.Name) {
		if amount, ok := e.scaledAmount(*ref, parsed, true); ok {
			preview := e.calculator.Scale(*ref, amount)
			preview.Name = parsed.Name
			preview.Unit = parsed.Unit
			preview.Description = e.formatDescription(parsed.Amount, parsed.Unit, parsed.Name)
			return &preview, true, nil
		}
	}

	// 3. Fuzzy Match Reference
	if ref != nil {
		if amount, ok := e.scaledAmount(*ref, parsed, true); ok {
			preview := e.calculator.Scale(*ref, amount)
			preview.Name = parsed.Name
			preview.Unit = parsed.Unit
			preview.Description = e.formatDescription(amount, e.parser.normalizeUnit(ref.Unit), ref.Name)
			return &preview, true, nil
		}
	}

	if e.provider != nil {
		ref, err := e.provider.ResolveFood(parsed)
		if err != nil {
			return nil, false, fmt.Errorf("nutrition provider lookup error: %w", err)
		}
		if ref != nil {
			ref.Name = parsed.Name
			if err := e.validateMacros(ref.Macros); err != nil {
				return nil, false, err
			}
			if err := e.db.CacheFood(*ref); err != nil {
				fmt.Printf("Warning: failed to cache nutrition provider result: %v\n", err)
			}
			if amount, ok := e.scaledAmount(*ref, parsed, true); ok {
				preview := e.calculator.Scale(*ref, amount)
				preview.Name = parsed.Name
				preview.Unit = parsed.Unit
				return &preview, true, nil
			}
		}
	}

	return nil, false, nil
}

func (e *NutritionEngine) formatDescription(amount float64, unit string, name string) string {
	if unit == "unit" || unit == "" {
		if amount == 1 {
			return fmt.Sprintf("%g %s", amount, name)
		}
		return fmt.Sprintf("%.1f %s", amount, name)
	}
	if unit == "gram" {
		return fmt.Sprintf("%.1fg %s", amount, name)
	}
	return fmt.Sprintf("%.1f%s %s", amount, unit, name)
}

func (e *NutritionEngine) analyzeWithLLMFallback(description string, parsed ParsedFood) (*models.FoodPreview, error) {
	// Step 5: LLM Fallback (Only if missing everywhere)
	if e.llm == nil {
		return nil, fmt.Errorf("no compatible nutrition data found for %q with unit %q", parsed.Name, parsed.Unit)
	}

	llmRef, err := e.llm.ParseFood(description)
	if err != nil {
		return nil, err
	}

	// Validate realistic values
	if err := e.validateMacros(llmRef.Macros); err != nil {
		return nil, err
	}

	// Store result in cache for future use
	// Ensure the cache entry name matches the normalized parsed name
	llmRef.Name = parsed.Name
	if err := e.db.CacheFood(*llmRef); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Warning: failed to cache LLM result: %v\n", err)
	}

	// Scale to requested amount
	preview := e.calculator.Scale(*llmRef, parsed.Amount)
	preview.Name = parsed.Name
	preview.Unit = parsed.Unit
	return &preview, nil
}

func (e *NutritionEngine) analyzeWholeMealWithLLMFallback(description string) (*models.FoodPreview, error) {
	if e.llm == nil {
		return nil, fmt.Errorf("could not resolve every food item without LLM: %s", description)
	}

	llmRef, err := e.llm.ParseFood(description)
	if err != nil {
		return nil, err
	}
	if err := e.validateMacros(llmRef.Macros); err != nil {
		return nil, err
	}

	llmRef.Name = e.parser.normalizeName(description)
	llmRef.BaseQuantity = 1
	llmRef.Unit = "unit"
	if err := e.db.CacheFood(*llmRef); err != nil {
		fmt.Printf("Warning: failed to cache LLM result: %v\n", err)
	}

	preview := e.calculator.Scale(*llmRef, 1)
	preview.Description = llmRef.Name
	return &preview, nil
}

func (e *NutritionEngine) scaledAmount(ref models.ReferenceFood, parsed ParsedFood, allowEstimation bool) (float64, bool) {
	if parsed.Amount <= 0 {
		return ref.BaseQuantity, true
	}

	normRefUnit := e.parser.normalizeUnit(ref.Unit)
	if parsed.Unit == "" {
		if normRefUnit == "" || normRefUnit == "unit" {
			return parsed.Amount, true
		}
		// If ref is gram, we can assume unitless amount is grams
		if normRefUnit == "gram" {
			return parsed.Amount, true
		}
		return parsed.Amount, true
	}

	if normRefUnit == parsed.Unit {
		return parsed.Amount, true
	}

	if allowEstimation && normRefUnit == "gram" {
		grams := e.estimateGrams(parsed)
		return grams, grams > 0
	}

	return 0, false
}

func (e *NutritionEngine) estimateGrams(parsed ParsedFood) float64 {
	amount := parsed.Amount
	if amount <= 0 {
		amount = 1
	}

	rule, ok := globalRules.Units[parsed.Unit]
	if !ok {
		return 0
	}

	for _, override := range rule.Overrides {
		if containsAny(parsed.Name, override.Terms...) {
			return amount * override.Value
		}
	}
	return amount * rule.Default
}

func containsAny(value string, terms ...string) bool {
	for _, term := range terms {
		if strings.Contains(value, term) {
			return true
		}
	}
	return false
}

func (e *NutritionEngine) validateMacros(m models.Macros) error {
	if m.Calories < 0 || m.Protein < 0 || m.Carbs < 0 || m.Fat < 0 {
		return fmt.Errorf("LLM returned unrealistic negative values")
	}
	// Simple rule: max 900kcal per 100g (pure fat)
	if m.Calories > 5000 {
		return fmt.Errorf("LLM returned unrealistic calorie value: %.0f", m.Calories)
	}
	return nil
}
