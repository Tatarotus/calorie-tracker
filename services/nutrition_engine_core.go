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

// Composable Service Interfaces
type Parser interface {
	Parse(desc string) ([]ParsedFoodItem, error)
}

type CanonicalResolver interface {
	Resolve(name string) (*models.CanonicalFood, error)
}

type CacheResolver interface {
	Get(foodID int64, unit string) (*models.NutritionCacheEntry, error)
	GetOverride(foodID int64) (*models.UserOverrideEntry, error)
}

type NutritionResolver interface {
	Resolve(canonical *models.CanonicalFood, item ParsedFoodItem) (*models.ReferenceFood, *models.ResolutionTrace, error)
}

type Validator interface {
	Validate(query string, resolved *models.ReferenceFood) (bool, []string, error)
}

// Default Parser implementation using LLM
type LlmParser struct {
	llm *LLMService
}

func (p *LlmParser) Parse(desc string) ([]ParsedFoodItem, error) {
	if p.llm == nil {
		return nil, fmt.Errorf("no LLM service configured for parsing")
	}
	return p.llm.ParseFoodItems(desc)
}

// Default CanonicalResolver implementation using DB and SynonymMapper
type DbCanonicalResolver struct {
	db            db.DBProvider
	synonymMapper *SynonymMapper
	service       *CanonicalResolverService
}

func (r *DbCanonicalResolver) Resolve(name string) (*models.CanonicalFood, error) {
	if r.service != nil {
		return r.service.Resolve(name)
	}
	canonicalKey := r.synonymMapper.GetCanonical(name)
	f, err := r.db.GetCanonicalFoodByName(canonicalKey)
	if err != nil {
		return nil, err
	}
	if f != nil {
		return f, nil
	}

	// Create and persist a new CanonicalFood entry if not found
	now := time.Now()
	newFood := &models.CanonicalFood{
		CanonicalName:        canonicalKey,
		NormalizedName:       strings.ToLower(strings.TrimSpace(name)),
		AliasesJSON:          `["` + name + `"]`,
		Language:             "en",
		Category:             "auto",
		DefaultServingAmount: 100,
		DefaultServingUnit:   "gram",
		DensityMultiplier:    1.0,
		GramsPerML:           1.0,
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	if strings.Contains(canonicalKey, "ovo") || strings.Contains(canonicalKey, "egg") || strings.Contains(canonicalKey, "banana") {
		newFood.DefaultServingAmount = 1.0
		newFood.DefaultServingUnit = "unit"
	}

	err = r.db.SaveCanonicalFood(newFood)
	if err != nil {
		return nil, err
	}

	return newFood, nil
}

// Default CacheResolver implementation
type DbCacheResolver struct {
	db db.DBProvider
}

func (r *DbCacheResolver) Get(foodID int64, unit string) (*models.NutritionCacheEntry, error) {
	entry, err := r.db.GetNutritionCache(foodID, unit)
	if err == nil && entry != nil {
		return entry, nil
	}

	// Legacy cache lookup fallback for test compatibility and backward compatibility
	cf, err := r.db.GetCanonicalFood(foodID)
	if err == nil && cf != nil {
		ref, err := r.db.GetCachedFood(cf.CanonicalName)
		if err == nil && ref != nil {
			parsedUnit := normalizeUnit(unit)
			refUnit := normalizeUnit(ref.Unit)
			if parsedUnit == "" || refUnit == "" || parsedUnit == refUnit {
				return &models.NutritionCacheEntry{
					CanonicalFoodID:  foodID,
					ServingAmount:    ref.BaseQuantity,
					ServingUnit:      ref.Unit,
					Calories:         ref.Macros.Calories,
					Protein:          ref.Macros.Protein,
					Carbs:            ref.Macros.Carbs,
					Fat:              ref.Macros.Fat,
					SourceType:       "legacy_cache",
					SourceConfidence: 1.0,
					UpdatedAt:        time.Now(),
				}, nil
			}
		}
		// Also try by normalized name
		ref, err = r.db.GetCachedFood(cf.NormalizedName)
		if err == nil && ref != nil {
			parsedUnit := normalizeUnit(unit)
			refUnit := normalizeUnit(ref.Unit)
			if parsedUnit == "" || refUnit == "" || parsedUnit == refUnit {
				return &models.NutritionCacheEntry{
					CanonicalFoodID:  foodID,
					ServingAmount:    ref.BaseQuantity,
					ServingUnit:      ref.Unit,
					Calories:         ref.Macros.Calories,
					Protein:          ref.Macros.Protein,
					Carbs:            ref.Macros.Carbs,
					Fat:              ref.Macros.Fat,
					SourceType:       "legacy_cache",
					SourceConfidence: 1.0,
					UpdatedAt:        time.Now(),
				}, nil
			}
		}
	}

	return nil, nil
}

func (r *DbCacheResolver) GetOverride(foodID int64) (*models.UserOverrideEntry, error) {
	return r.db.GetUserOverride(foodID)
}

// ResolutionCandidate wraps a reference food and its attributes for rank sorting.
type ResolutionCandidate struct {
	ReferenceFood    *models.ReferenceFood
	Confidence       float64
	SourceType       string
	ResolutionMethod string
	IsStale          bool
}

// Default Validator implementation using LLM
type LlmValidator struct {
	llm *LLMService
}

func (v *LlmValidator) Validate(query string, resolved *models.ReferenceFood) (bool, []string, error) {
	if v.llm == nil {
		return true, nil, nil
	}
	return v.llm.Validate(query, resolved)
}
