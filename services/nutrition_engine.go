package services

import (
	"calorie-tracker/data"
	"calorie-tracker/db"
	"calorie-tracker/models"
	"encoding/json"
	"fmt"
	"strings"
	"time"
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
	db            db.DBProvider
	parser        *FoodParser
	calculator    *MacroCalculator
	llm           *LLMService
	providers     []NutritionProvider
	fuzzyMatcher  *FuzzyMatcher
	synonymMapper *SynonymMapper
	llmCache      *LLMCache
}

func NewNutritionEngine(db db.DBProvider, llm *LLMService) *NutritionEngine {
	return NewNutritionEngineWithProviders(db, llm, nil)
}

func NewNutritionEngineWithProvider(db db.DBProvider, llm *LLMService, provider NutritionProvider) *NutritionEngine {
	var providers []NutritionProvider
	if provider != nil {
		providers = append(providers, provider)
	}
	return NewNutritionEngineWithProviders(db, llm, providers)
}

func NewNutritionEngineWithProviders(db db.DBProvider, llm *LLMService, providers []NutritionProvider) *NutritionEngine {
	return &NutritionEngine{
		db:            db,
		parser:        NewFoodParser(),
		calculator:    NewMacroCalculator(),
		llm:           llm,
		providers:     providers,
		fuzzyMatcher:  NewFuzzyMatcher(0.8),
		synonymMapper: NewSynonymMapper(),
		llmCache:      NewLLMCache(24*time.Hour, 1000),
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

	// If split into multiple items but couldn't resolve all of them,
	// try resolving the whole description as a single item before falling back to LLM.
	if len(items) > 1 {
		singleItem := e.parser.Parse(description)
		if singleItem.Name != "" {
			preview, ok, err := e.resolveSingle(singleItem)
			if err == nil && ok {
				return preview, nil
			}
		}
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
	if preview, ok := e.tryExactCache(parsed); ok {
		return preview, true, nil
	}

	// 2. Synonym-based Cache Lookup
	if preview, ok := e.trySynonymCache(parsed); ok {
		return preview, true, nil
	}

	// 3. Fuzzy Cache Lookup
	if preview, ok := e.tryFuzzyCache(parsed); ok {
		return preview, true, nil
	}

	// 4. Exact Match Reference
	if preview, ok := e.tryExactReference(parsed); ok {
		return preview, true, nil
	}

	// 5. Synonym-based Reference Lookup
	if preview, ok := e.trySynonymReference(parsed); ok {
		return preview, true, nil
	}

	// 6. Fuzzy Match Reference
	if preview, ok := e.tryFuzzyReference(parsed); ok {
		return preview, true, nil
	}

	// 7. External Nutrition Providers
	return e.tryExternalProviders(parsed)
}

// tryExactCache attempts an exact match from the food cache
func (e *NutritionEngine) tryExactCache(parsed ParsedFood) (*models.FoodPreview, bool) {
	cached, err := e.db.GetCachedFood(parsed.Name)
	if err != nil || cached == nil {
		return nil, false
	}
	if amount, ok := e.scaledAmount(*cached, parsed, false); ok {
		preview := e.calculator.Scale(*cached, amount)
		preview.Name = parsed.Name
		preview.Unit = parsed.Unit
		preview.Description = e.formatDescription(parsed.Amount, parsed.Unit, parsed.Name)
		return &preview, true
	}
	return nil, false
}

// trySynonymCache attempts a synonym-based match from the food cache
func (e *NutritionEngine) trySynonymCache(parsed ParsedFood) (*models.FoodPreview, bool) {
	canonicalName := e.synonymMapper.GetCanonical(parsed.Name)
	if canonicalName == parsed.Name {
		return nil, false
	}
	cached, err := e.db.GetCachedFood(canonicalName)
	if err != nil || cached == nil {
		return nil, false
	}
	if amount, ok := e.scaledAmount(*cached, parsed, false); ok {
		preview := e.calculator.Scale(*cached, amount)
		preview.Name = parsed.Name
		preview.Unit = parsed.Unit
		preview.Description = e.formatDescription(parsed.Amount, parsed.Unit, parsed.Name)
		return &preview, true
	}
	return nil, false
}

// tryFuzzyCache attempts a fuzzy match from the food cache
func (e *NutritionEngine) tryFuzzyCache(parsed ParsedFood) (*models.FoodPreview, bool) {
	fuzzyCached, err := e.fuzzyFindInCache(parsed.Name)
	if err != nil || fuzzyCached == nil {
		return nil, false
	}
	if amount, ok := e.scaledAmount(*fuzzyCached, parsed, false); ok {
		preview := e.calculator.Scale(*fuzzyCached, amount)
		preview.Name = parsed.Name
		preview.Unit = parsed.Unit
		preview.Description = e.formatDescription(parsed.Amount, parsed.Unit, parsed.Name)
		return &preview, true
	}
	return nil, false
}

// tryExactReference attempts an exact match from the reference database
func (e *NutritionEngine) tryExactReference(parsed ParsedFood) (*models.FoodPreview, bool) {
	ref, err := e.db.GetReferenceFood(parsed.Name)
	if err != nil || ref == nil || !strings.EqualFold(ref.Name, parsed.Name) {
		return nil, false
	}
	if amount, ok := e.scaledAmount(*ref, parsed, true); ok {
		preview := e.calculator.Scale(*ref, amount)
		preview.Name = parsed.Name
		preview.Unit = parsed.Unit
		preview.Description = e.formatDescription(parsed.Amount, parsed.Unit, parsed.Name)
		return &preview, true
	}
	return nil, false
}

// trySynonymReference attempts a synonym-based match from the reference database
func (e *NutritionEngine) trySynonymReference(parsed ParsedFood) (*models.FoodPreview, bool) {
	canonicalName := e.synonymMapper.GetCanonical(parsed.Name)
	if canonicalName == parsed.Name {
		return nil, false
	}
	ref, err := e.db.GetReferenceFood(canonicalName)
	if err != nil || ref == nil {
		return nil, false
	}
	if amount, ok := e.scaledAmount(*ref, parsed, true); ok {
		preview := e.calculator.Scale(*ref, amount)
		preview.Name = parsed.Name
		preview.Unit = parsed.Unit
		preview.Description = e.formatDescription(parsed.Amount, parsed.Unit, parsed.Name)
		return &preview, true
	}
	return nil, false
}

// tryFuzzyReference attempts a fuzzy match from the reference database
func (e *NutritionEngine) tryFuzzyReference(parsed ParsedFood) (*models.FoodPreview, bool) {
	ref, err := e.db.GetReferenceFood(parsed.Name)
	if err != nil || ref == nil {
		return nil, false
	}
	if amount, ok := e.scaledAmount(*ref, parsed, true); ok {
		preview := e.calculator.Scale(*ref, amount)
		preview.Name = parsed.Name
		preview.Unit = parsed.Unit
		preview.Description = e.formatDescription(amount, e.parser.normalizeUnit(ref.Unit), ref.Name)
		return &preview, true
	}
	return nil, false
}

// tryExternalProviders attempts to resolve food via external nutrition providers
func (e *NutritionEngine) tryExternalProviders(parsed ParsedFood) (*models.FoodPreview, bool, error) {
	if len(e.providers) == 0 {
		return nil, false, nil
	}

	for _, provider := range e.providers {
		ref, err := provider.ResolveFood(parsed)
		if err != nil {
			fmt.Printf("Warning: nutrition provider lookup error: %v\n", err)
			continue
		}
		if ref == nil {
			continue
		}
		ref.Name = parsed.Name
		if err := e.validateMacros(ref.Macros); err != nil {
			fmt.Printf("Warning: invalid macros from provider: %v\n", err)
			continue
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

	return nil, false, nil
}

// fuzzyFindInCache searches the cache for a fuzzy match
func (e *NutritionEngine) fuzzyFindInCache(name string) (*models.ReferenceFood, error) {
	allCache, err := e.db.GetAllCacheEntries()
	if err != nil || len(allCache) == 0 {
		return nil, err
	}

	candidates := make([]string, 0, len(allCache))
	cacheMap := make(map[string]models.ReferenceFood)
	for _, entry := range allCache {
		candidates = append(candidates, entry.Name)
		cacheMap[entry.Name] = entry
	}

	bestMatch, score := e.fuzzyMatcher.FindBestMatch(name, candidates)
	if score >= e.fuzzyMatcher.threshold && bestMatch != "" {
		if entry, ok := cacheMap[bestMatch]; ok {
			return &entry, nil
		}
	}

	return nil, nil
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
	if e.llm == nil {
		return nil, fmt.Errorf("no compatible nutrition data found for %q with unit %q", parsed.Name, parsed.Unit)
	}

	if cached, found := e.llmCache.Get(description); found {
		cached.Name = parsed.Name
		if err := e.db.CacheFood(*cached); err != nil {
			fmt.Printf("Warning: failed to cache LLM result: %v\n", err)
		}
		preview := e.calculator.Scale(*cached, parsed.Amount)
		preview.Name = parsed.Name
		preview.Unit = parsed.Unit
		return &preview, nil
	}

	llmRef, err := e.llm.ParseFood(description)
	if err != nil {
		return nil, err
	}

	if err := e.validateMacros(llmRef.Macros); err != nil {
		return nil, err
	}

	e.llmCache.Set(description, *llmRef)

	llmRef.Name = parsed.Name
	if err := e.db.CacheFood(*llmRef); err != nil {
		fmt.Printf("Warning: failed to cache LLM result: %v\n", err)
	}

	preview := e.calculator.Scale(*llmRef, parsed.Amount)
	preview.Name = parsed.Name
	preview.Unit = parsed.Unit
	return &preview, nil
}

func (e *NutritionEngine) analyzeWholeMealWithLLMFallback(description string) (*models.FoodPreview, error) {
	if e.llm == nil {
		return nil, fmt.Errorf("could not resolve every food item without LLM: %s", description)
	}

	if cached, found := e.llmCache.Get(description); found {
		cached.Name = e.parser.normalizeName(description)
		cached.BaseQuantity = 1
		cached.Unit = "unit"
		if err := e.db.CacheFood(*cached); err != nil {
			fmt.Printf("Warning: failed to cache LLM result: %v\n", err)
		}
		preview := e.calculator.Scale(*cached, 1)
		preview.Description = cached.Name
		return &preview, nil
	}

	llmRef, err := e.llm.ParseFood(description)
	if err != nil {
		return nil, err
	}
	if err := e.validateMacros(llmRef.Macros); err != nil {
		return nil, err
	}

	e.llmCache.Set(description, *llmRef)

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
		if normRefUnit == "" || normRefUnit == "unit" || normRefUnit == "gram" {
			return parsed.Amount, true
		}
		return parsed.Amount, true
	}

	if normRefUnit == parsed.Unit {
		return parsed.Amount, true
	}

	if normRefUnit == "" {
		if parsed.Unit == "unit" || parsed.Unit == "gram" {
			return parsed.Amount, true
		}
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
		for _, term := range override.Terms {
			if strings.Contains(parsed.Name, term) {
				return amount * override.Value
			}
		}
	}
	return amount * rule.Default
}

func (e *NutritionEngine) validateMacros(m models.Macros) error {
	if m.Calories < 0 || m.Protein < 0 || m.Carbs < 0 || m.Fat < 0 {
		return fmt.Errorf("LLM returned unrealistic negative values")
	}
	if m.Calories > 5000 {
		return fmt.Errorf("LLM returned unrealistic calorie value: %.0f", m.Calories)
	}
	return nil
}
