package services

import (
	"calorie-tracker/db"
	"calorie-tracker/models"
	"fmt"
)

// NutritionEngine orchestrates the hybrid nutrition lookup flow
type NutritionEngine struct {
	db         db.DBProvider
	parser     *FoodParser
	calculator *MacroCalculator
	llm        *LLMService
}

func NewNutritionEngine(db db.DBProvider, llm *LLMService) *NutritionEngine {
	return &NutritionEngine{
		db:         db,
		parser:     NewFoodParser(),
		calculator: NewMacroCalculator(),
		llm:        llm,
	}
}

func (e *NutritionEngine) Analyze(description string) (*models.FoodPreview, error) {
	// Step 1: Normalize Input
	parsed := e.parser.Parse(description)
	if parsed.Name == "" {
		return nil, fmt.Errorf("could not parse food name from: %s", description)
	}

	// Step 2: Check Local Reference Database (Source of Truth)
	ref, err := e.db.GetReferenceFood(parsed.Name)
	if err != nil {
		return nil, fmt.Errorf("reference lookup error: %w", err)
	}

	if ref != nil {
		// Step 3: Deterministic Scaling
		preview := e.calculator.Scale(*ref, parsed.Amount)
		return &preview, nil
	}

	// Step 4: Check Cache Layer (Previous LLM results)
	cached, err := e.db.GetCachedFood(parsed.Name)
	if err != nil {
		return nil, fmt.Errorf("cache lookup error: %w", err)
	}

	if cached != nil {
		// Scale cached base values to requested amount
		preview := e.calculator.Scale(*cached, parsed.Amount)
		return &preview, nil
	}

	// Step 5: LLM Fallback (Only if missing everywhere)
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
	return &preview, nil
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
