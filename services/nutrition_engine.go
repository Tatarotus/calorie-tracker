package services

import (
	"calorie-tracker/db"
	"calorie-tracker/models"
	"fmt"
	"strings"
)

// NutritionEngine orchestrates the ingestion pipeline
type NutritionEngine struct {
	db                db.DBProvider
	parser            Parser
	normalizer        *Normalizer
	canonicalResolver *CanonicalResolverService
	cacheResolver     CacheResolver
	nutritionResolver NutritionResolver
	validator         *SemanticValidator
	scaler            *MacroScaler
	confidenceCalc    *ConfidenceCalculator
	fuzzyResolver     *FuzzyResolver
	ttlPolicy         *TTLPolicy
	calculator        *MacroCalculator
	fuzzyMatcher      *FuzzyMatcher
	synonymMapper     *SynonymMapper
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
	synonymMapper := NewSynonymMapper()
	_ = synonymMapper.LoadFromDatabase(db)

	cacheResolver := &DbCacheResolver{db: db}
	fuzzyMatcher := NewFuzzyMatcher(0.8)
	normalizer := NewNormalizer()
	canonicalService := NewCanonicalResolverService(db, synonymMapper, normalizer)
	fuzzyResolver := NewFuzzyResolver(db, fuzzyMatcher, synonymMapper)
	ttlPolicy := NewTTLPolicy()
	scaler := NewMacroScaler()
	confidenceCalc := NewConfidenceCalculator()
	validator := NewSemanticValidator(llm, synonymMapper)

	return &NutritionEngine{
		db:                db,
		parser:            &LlmParser{llm: llm},
		normalizer:        normalizer,
		canonicalResolver: canonicalService,
		cacheResolver:     cacheResolver,
		nutritionResolver: &HybridNutritionResolver{
			db:                db,
			llm:               llm,
			providers:         providers,
			fuzzyResolver:     fuzzyResolver,
			ttlPolicy:         ttlPolicy,
			cacheResolver:     cacheResolver,
			validator:         validator,
			canonicalResolver: canonicalService,
		},
		validator:      validator,
		scaler:         scaler,
		confidenceCalc: confidenceCalc,
		fuzzyResolver:  fuzzyResolver,
		ttlPolicy:      ttlPolicy,
		calculator:     NewMacroCalculator(),
		fuzzyMatcher:   fuzzyMatcher,
		synonymMapper:  synonymMapper,
	}
}

// Analyze parses raw user inputs and resolves them in the 11-stage pipeline.
func (e *NutritionEngine) Analyze(description string) (*models.FoodPreview, error) {
	// Normalization Layer executes first
	normalizedDesc := e.normalizer.Normalize(description)

	// Stage 1: Linguistic parsing (using already-normalized input)
	var items []ParsedFoodItem
	var err error

	if e.parser != nil {
		items, err = e.parser.Parse(normalizedDesc)
	}

	basicItems := parsedItemsFromBasicParser(e, normalizedDesc)
	if shouldPreferBasicParsedItems(basicItems, items) {
		items = basicItems
		err = nil
	}

	// Fallback to basic regex parsing if LLM is unavailable or failed
	if err != nil || len(items) == 0 {
		items = basicItems
	}
	e.applyDisplayNames(description, items)

	if len(items) == 0 {
		return nil, fmt.Errorf("could not parse food items from: %s", description)
	}

	totalPreview := &models.FoodPreview{}
	descriptions := make([]string, 0, len(items))

	for _, item := range items {
		// Stage 2: Canonical identity resolution
		canonical, err := e.canonicalResolver.Resolve(item.FoodName)
		if err != nil {
			return nil, fmt.Errorf("canonical resolution error: %w", err)
		}

		itemPreview, err := e.analyzeItem(canonical, item)
		if err != nil {
			return nil, err
		}

		descriptions = append(descriptions, itemPreview.Description)
		totalPreview.Calories += itemPreview.Calories
		totalPreview.Protein += itemPreview.Protein
		totalPreview.Carbs += itemPreview.Carbs
		totalPreview.Fat += itemPreview.Fat
		totalPreview.Confidence = itemPreview.Confidence
		totalPreview.ResolutionTrace = itemPreview.ResolutionTrace
	}

	totalPreview.Description = strings.Join(descriptions, " + ")
	return totalPreview, nil
}

func parsedItemsFromBasicParser(e *NutritionEngine, description string) []ParsedFoodItem {
	basicParser := NewFoodParser()
	basicItems := basicParser.ParseMeal(description)
	items := make([]ParsedFoodItem, 0, len(basicItems))
	for _, b := range basicItems {
		items = append(items, ParsedFoodItem{
			Quantity:     b.Amount,
			Unit:         b.Unit,
			FoodName:     b.Name,
			CanonicalKey: e.synonymMapper.GetCanonical(b.Name),
		})
	}
	return items
}

func shouldPreferBasicParsedItems(basicItems, llmItems []ParsedFoodItem) bool {
	if len(basicItems) == 0 || len(basicItems) != len(llmItems) {
		return false
	}
	for i := range basicItems {
		if len(preparationTerms(basicItems[i].FoodName)) > 0 && hasPreparationMismatch(basicItems[i].FoodName, llmItems[i].FoodName) {
			return true
		}
	}
	return false
}

func (e *NutritionEngine) applyDisplayNames(description string, items []ParsedFoodItem) {
	if len(items) == 0 {
		return
	}
	parser := NewFoodParser()
	displayItems := parser.ParseMealForDisplay(description)
	if len(displayItems) != len(items) {
		return
	}
	for i := range items {
		if displayItems[i].Name != "" {
			items[i].DisplayName = displayItems[i].Name
		}
		if displayItems[i].Unit != "" {
			items[i].Unit = displayItems[i].Unit
		}
		if displayItems[i].Amount > 0 {
			items[i].Quantity = displayItems[i].Amount
		}
	}
}

func (e *NutritionEngine) analyzeItem(canonical *models.CanonicalFood, item ParsedFoodItem) (*models.FoodPreview, error) {
	// Stage 3 to 9: Resolve base reference nutrition
	resolvedRef, trace, err := e.nutritionResolver.Resolve(canonical, item)
	if err != nil {
		return nil, fmt.Errorf("nutrition resolution error: %w", err)
	}

	// Stage 10: Conditional Semantic Validation
	resolvedRef, err = e.performSemanticValidation(canonical, resolvedRef, trace)
	if err != nil {
		return nil, err
	}

	// Stage 11: Final macro scaling, confidence scoring & auditing
	var scaledMacros models.Macros
	var scaledQty float64

	if resolvedRef != nil {
		scaledMacros, scaledQty = e.scaler.Scale(canonical, resolvedRef, item.Quantity, item.Unit)
		if trace.SourceConfidence == 0 {
			trace.SourceConfidence = 1.0 // fallback default
		}
	} else {
		// Principle of Correctness > Coverage: fallback to 0 macros rather than hallucinating
		scaledMacros = models.Macros{}
		scaledQty = item.Quantity
		trace.ResolutionMethod = "unresolved_fallback"
		trace.SourceConfidence = 0.0
	}

	// Final confidence score aggregation
	finalConfidence := e.confidenceCalc.Calculate(trace)

	itemPreview := &models.FoodPreview{
		Description:     e.scaler.FormatDescription(scaledQty, item.Unit, displayNameForItem(item)),
		Calories:        scaledMacros.Calories,
		Protein:         scaledMacros.Protein,
		Carbs:           scaledMacros.Carbs,
		Fat:             scaledMacros.Fat,
		Confidence:      finalConfidence,
		ResolutionTrace: trace,
	}

	return itemPreview, nil
}

func displayNameForItem(item ParsedFoodItem) string {
	if strings.TrimSpace(item.DisplayName) != "" {
		return item.DisplayName
	}
	return item.FoodName
}

func (e *NutritionEngine) performSemanticValidation(canonical *models.CanonicalFood, resolvedRef *models.ReferenceFood, trace *models.ResolutionTrace) (*models.ReferenceFood, error) {
	if resolvedRef == nil {
		return nil, nil
	}
	if !e.ShouldValidate(trace) || e.validator == nil {
		return resolvedRef, nil
	}

	trace.ValidationTriggered = true
	queryType := models.FoodType(canonical.FoodType)
	matchedType := models.FoodType("")
	if matchedCanonical, err := e.canonicalResolver.Resolve(resolvedRef.Name); err == nil && matchedCanonical != nil {
		matchedType = models.FoodType(matchedCanonical.FoodType)
	}
	valid, warnings := e.validator.Validate(queryType, matchedType, resolvedRef)
	trace.ValidationWarnings = append(trace.ValidationWarnings, warnings...)
	if !valid {
		// Check for unrealistic calorie density (> 950 kcal/100g is physically impossible)
		if resolvedRef.BaseQuantity > 0 {
			unit := strings.ToLower(strings.TrimSpace(resolvedRef.Unit))
			if isMassOrVolumeUnit(unit) && (resolvedRef.Macros.Calories/resolvedRef.BaseQuantity)*100.0 > 950.0 {
				return nil, fmt.Errorf("semantic validation failed: unrealistic calorie density detected (> 950 kcal/100g is physically impossible)")
			}
		}
		// Semantic mismatch risk: drop reference to preserve correctness!
		trace.ResolutionMethod = "unresolved_rejected"
		trace.SourceConfidence = 0.0
		trace.ValidationResult = "INVALID"
		return nil, nil
	}
	trace.ValidationResult = "VALID"
	return resolvedRef, nil
}

// ShouldValidate determines if the semantic validator should trigger
func (e *NutritionEngine) ShouldValidate(trace *models.ResolutionTrace) bool {
	if trace == nil {
		return false
	}
	if trace.SourceType == "user_override" || trace.CacheHit || (trace.SourceType == "fatsecret" && trace.SourceConfidence >= 1.0) {
		return false // Bypassed
	}
	if trace.SourceConfidence < 0.9 || trace.SourceType == "serpapi_fallback" || trace.SourceType == "fuzzy_fallback" || trace.SourceType == "reference_db" || trace.SourceType == "llm" {
		return true
	}
	return false
}

// Compatibility wrappers for existing tests

func (e *NutritionEngine) estimateGrams(name string, amount float64, unit string) float64 {
	return e.scaler.EstimateGrams(name, amount, unit)
}

func (e *NutritionEngine) formatDescription(amount float64, unit string, name string) string {
	return e.scaler.FormatDescription(amount, unit, name)
}

func semanticTokenCheck(parsedName, resolvedName string) bool {
	v := SemanticValidator{}
	return v.SemanticTokenCheck(parsedName, resolvedName)
}
