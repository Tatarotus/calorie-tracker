package services

import (
	"encoding/json"
	"testing"
	"time"

	"calorie-tracker/db"
	"calorie-tracker/models"
)

func TestNutritionEngine_FormatDescription(t *testing.T) {
	mockDB := db.NewMockDB()
	engine := NewNutritionEngine(mockDB, nil)

	tests := []struct {
		name     string
		amount   float64
		unit     string
		foodName string
		expected string
	}{
		{"single unit", 1, "unit", "apple", "1 apple"},
		{"multiple units", 2, "unit", "apple", "2.0 apple"},
		{"grams", 100, "gram", "rice", "100.0g rice"},
		{"cups", 1, "cup", "water", "1.0cup water"},
		{"single slice", 1, "slice", "pão de forma", "1 fatia de pão de forma"},
		{"multiple slices", 2, "slice", "pão de forma", "2 fatias de pão de forma"},
		{"empty unit single", 1, "", "banana", "1 banana"},
		{"empty unit multiple", 2, "", "banana", "2.0 banana"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := engine.formatDescription(tt.amount, tt.unit, tt.foodName)
			if got != tt.expected {
				t.Errorf("formatDescription() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestNutritionEngine_Analyze_EmptyItems(t *testing.T) {
	mockDB := db.NewMockDB()
	engine := NewNutritionEngine(mockDB, nil)

	_, err := engine.Analyze("   ")
	if err == nil {
		t.Error("expected error for empty/whitespace input")
	}
}

func TestNutritionEngine_UserOverridePrecedence(t *testing.T) {
	mockDB := db.NewMockDB()

	// Seed canonical food
	cf := &models.CanonicalFood{
		ID:             123,
		CanonicalName:  "cafe_com_leite",
		NormalizedName: "cafe com leite",
		Language:       "pt",
		Category:       "beverage",
	}
	_ = mockDB.SaveCanonicalFood(cf)

	// Save user override (Calories = 50)
	override := &models.UserOverrideEntry{
		CanonicalFoodID: 123,
		ServingAmount:   100.0,
		ServingUnit:     "ml",
		Calories:        50,
		Protein:         2,
		Carbs:           6,
		Fat:             1,
		OverrideReason:  "my custom recipe",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	_ = mockDB.SaveUserOverride(override)

	// Seed normal cache (Calories = 100) - user override should take precedence!
	cacheEntry := &models.NutritionCacheEntry{
		CanonicalFoodID: 123,
		ServingAmount:   100.0,
		ServingUnit:     "ml",
		Calories:        100,
		Protein:         4,
		Carbs:           12,
		Fat:             2,
		SourceType:      "fatsecret",
		UpdatedAt:       time.Now(),
	}
	_ = mockDB.SaveNutritionCache(cacheEntry)

	engine := NewNutritionEngine(mockDB, nil)

	preview, err := engine.Analyze("100ml cafe com leite")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if preview == nil {
		t.Fatal("expected non-nil preview")
	}

	bz, _ := json.MarshalIndent(preview, "", "  ")
	t.Logf("ACTUAL PREVIEW: %s", string(bz))

	if preview.Calories != 50 {
		t.Errorf("expected 50 calories (user override), got %f (cache/other)", preview.Calories)
	}

	if preview.ResolutionTrace == nil {
		t.Fatal("expected non-nil resolution trace")
	}

	if preview.ResolutionTrace.ResolutionMethod != "user_override" {
		t.Errorf("expected resolution method 'user_override', got %q", preview.ResolutionTrace.ResolutionMethod)
	}
}

func TestNutritionEngine_SemanticTokenCheckProtection(t *testing.T) {
	// cafe com leite (parsed) vs black coffee (resolved) should be rejected
	if semanticTokenCheck("cafe com leite", "black coffee") {
		t.Error("expected rejection of 'black coffee' for 'cafe com leite'")
	}

	// cafe com leite vs cafe com leite integral should be accepted
	if !semanticTokenCheck("cafe com leite", "cafe com leite integral") {
		t.Error("expected acceptance of 'cafe com leite integral' for 'cafe com leite'")
	}

	// bread vs apple should be rejected
	if semanticTokenCheck("bread", "apple") {
		t.Error("expected rejection of 'apple' for 'bread'")
	}
}

func TestNutritionEngine_ConfidenceAggregation(t *testing.T) {
	mockDB := db.NewMockDB()
	engine := NewNutritionEngine(mockDB, nil)

	// Check confidence when unresolved_fallback is triggered (should be 0.0)
	preview, err := engine.Analyze("100g unknownweirdfood")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if preview == nil {
		t.Fatal("expected non-nil preview")
	}

	if preview.Confidence != 0.0 {
		t.Errorf("expected 0.0 confidence for unresolved fallback, got %f", preview.Confidence)
	}

	if preview.Calories != 0.0 || preview.Protein != 0.0 {
		t.Errorf("expected 0 macros for unresolved fallback to enforce correctness > coverage, got calories=%f", preview.Calories)
	}
}

func TestSynonymMapper_DynamicAliasLoading(t *testing.T) {
	mockDB := db.NewMockDB()

	// Seed custom canonical food with aliases in database
	cf := &models.CanonicalFood{
		CanonicalName:  "my_custom_food",
		NormalizedName: "my custom food",
		AliasesJSON:    `["custom alias 1", "custom alias 2"]`,
	}
	_ = mockDB.SaveCanonicalFood(cf)

	// Instantiating synonym mapper and loading from DB
	sm := NewSynonymMapper()
	err := sm.LoadFromDatabase(mockDB)
	if err != nil {
		t.Fatalf("unexpected error loading from DB: %v", err)
	}

	if canonical := sm.GetCanonical("custom alias 1"); canonical != "my_custom_food" {
		t.Errorf("expected 'my_custom_food' for 'custom alias 1', got %q", canonical)
	}

	if canonical := sm.GetCanonical("custom alias 2"); canonical != "my_custom_food" {
		t.Errorf("expected 'my_custom_food' for 'custom alias 2', got %q", canonical)
	}
}

func TestTrackerService_SaveFood_LineageAndOverride(t *testing.T) {
	mockDB := db.NewMockDB()
	tracker := NewTrackerService(mockDB, nil)

	preview := &models.FoodPreview{
		Name:        "banana",
		Unit:        "unit",
		Description: "2 banana",
		Calories:    178,
		Protein:     2.2,
		Carbs:       46,
		Fat:         0.6,
		UserEdited:  true,
		ResolutionTrace: &models.ResolutionTrace{
			ParserUsed:       "llm",
			CanonicalKey:     "banana",
			ResolutionMethod: "local_cache",
			SourceType:       "legacy_cache",
			SourceConfidence: 1.0,
		},
	}

	err := tracker.SaveFood(preview)
	if err != nil {
		t.Fatalf("SaveFood failed: %v", err)
	}

	// Verify food entry lineage was logged in DB
	entries, err := mockDB.GetDailyFoodEntries(time.Now())
	if err != nil {
		t.Fatalf("GetDailyFoodEntries failed: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}

	logged := entries[0]
	if logged.OriginalQuery != "2 banana" {
		t.Errorf("expected OriginalQuery '2 banana', got %q", logged.OriginalQuery)
	}
	if logged.CanonicalKey != "banana" {
		t.Errorf("expected CanonicalKey 'banana', got %q", logged.CanonicalKey)
	}

	var trace models.ResolutionTrace
	err = json.Unmarshal([]byte(logged.ResolutionTrace), &trace)
	if err != nil {
		t.Fatalf("failed to unmarshal trace JSON: %v", err)
	}

	if trace.CanonicalKey != "banana" || trace.ParserUsed != "llm" {
		t.Errorf("unmarshaled trace fields incorrect: %+v", trace)
	}

	// Verify UserOverrideEntry was persisted because UserEdited was true
	override, err := mockDB.GetUserOverride(1) // Banana's canonical ID in mockDB will be 1
	if err != nil {
		t.Fatalf("GetUserOverride failed: %v", err)
	}

	if override == nil {
		t.Fatal("expected user override to be persisted, but got nil")
	}

	// Scaling check: 2 bananas (amount=2) edited to have 178 calories.
	// Base quantity is 1 unit. Factor = 1 / 2 = 0.5.
	// Override base calories = 178 * 0.5 = 89.
	if override.Calories != 89 {
		t.Errorf("expected override calories scaled to base unit to be 89, got %f", override.Calories)
	}
}

type mockSpyParser struct {
	called bool
}

func (m *mockSpyParser) Parse(desc string) ([]ParsedFoodItem, error) {
	m.called = true
	return nil, nil
}

func TestNutritionEngine_ZeroLLM_CacheFastPath(t *testing.T) {
	mockDB := db.NewMockDB()

	// Seed canonical food
	cf := &models.CanonicalFood{
		ID:             1,
		CanonicalName:  "batata_frita",
		NormalizedName: "batata frita",
		Language:       "pt",
		Category:       "food",
	}
	_ = mockDB.SaveCanonicalFood(cf)

	// Seed cache entry
	cacheEntry := &models.NutritionCacheEntry{
		CanonicalFoodID:  1,
		ServingAmount:    100.0,
		ServingUnit:      "g",
		Calories:         312,
		Protein:          3.4,
		Carbs:            41.4,
		Fat:              15.0,
		SourceConfidence: 0.95,
		SourceType:       "fatsecret",
		UpdatedAt:        time.Now(),
	}
	_ = mockDB.SaveNutritionCache(cacheEntry)

	engine := NewNutritionEngine(mockDB, nil)
	spyParser := &mockSpyParser{}
	engine.parser = spyParser

	preview, err := engine.Analyze("100g batata frita")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if preview == nil {
		t.Fatal("expected non-nil preview")
	}

	if spyParser.called {
		t.Error("expected LLM parser to be completely bypassed, but Parse was called")
	}

	if preview.Calories != 312 {
		t.Errorf("expected 312 calories, got %f", preview.Calories)
	}
}

type mockProvider struct {
	delay time.Duration
	ref   *models.ReferenceFood
}

func (m *mockProvider) ResolveFood(item ParsedFood) (*models.ReferenceFood, error) {
	if m.delay > 0 {
		time.Sleep(m.delay)
	}
	return m.ref, nil
}

func TestHybridNutritionResolver_StaleWhileRevalidate(t *testing.T) {
	mockDB := db.NewMockDB()

	// Seed canonical food
	cf := &models.CanonicalFood{
		ID:             1,
		CanonicalName:  "batata_frita",
		NormalizedName: "batata frita",
		Language:       "pt",
		Category:       "food",
	}
	_ = mockDB.SaveCanonicalFood(cf)

	// Seed stale cache entry (Calories = 300)
	cacheEntry := &models.NutritionCacheEntry{
		CanonicalFoodID:  1,
		ServingAmount:    100.0,
		ServingUnit:      "g",
		Calories:         300,
		Protein:          3.0,
		Carbs:            40.0,
		Fat:              10.0,
		SourceConfidence: 0.95,
		SourceType:       "fatsecret",
		UpdatedAt:        time.Now().Add(-20 * 24 * time.Hour), // Expired
	}
	_ = mockDB.SaveNutritionCache(cacheEntry)

	// Setup mock provider with fresh results (Calories = 350)
	freshRef := &models.ReferenceFood{
		Name:         "batata frita",
		BaseQuantity: 100.0,
		Unit:         "g",
		Macros: models.Macros{
			Calories: 350,
			Protein:  4.0,
			Carbs:    45.0,
			Fat:      12.0,
		},
	}
	prov := &mockProvider{ref: freshRef}

	engine := NewNutritionEngineWithProviders(mockDB, nil, []NutritionProvider{prov})

	preview, err := engine.Analyze("100g batata frita")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if preview == nil {
		t.Fatal("expected non-nil preview")
	}

	// Verify that the stale entry macros are returned immediately
	if preview.Calories != 300 {
		t.Errorf("expected immediate stale calories of 300, got %f", preview.Calories)
	} // Wait for background goroutine to execute (up to 1 second)
	var updatedCache *models.NutritionCacheEntry
	for i := 0; i < 50; i++ {
		updatedCache, err = mockDB.GetNutritionCache(1, "g")
		if err == nil && updatedCache != nil && updatedCache.Calories == 350 {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	if updatedCache == nil {
		t.Fatal("expected updated cache to be non-nil")
	}

	if updatedCache.Calories != 350 {
		t.Errorf("expected updated cache calories to be 350, got %f", updatedCache.Calories)
	}
}

func TestHybridNutritionResolver_ParallelFanOut(t *testing.T) {
	mockDB := db.NewMockDB()

	// Seed canonical food
	cf := &models.CanonicalFood{
		ID:             1,
		CanonicalName:  "batata_frita",
		NormalizedName: "batata frita",
		Language:       "pt",
		Category:       "food",
	}
	_ = mockDB.SaveCanonicalFood(cf)

	// Setup two mock providers
	// Fast provider resolves immediately
	fastRef := &models.ReferenceFood{
		Name:         "batata frita",
		BaseQuantity: 100.0,
		Unit:         "g",
		Macros: models.Macros{
			Calories: 320,
			Protein:  3.5,
			Carbs:    42.0,
			Fat:      14.0,
		},
	}
	fastProv := &mockProvider{
		delay: 0,
		ref:   fastRef,
	}

	// Slow provider takes a while to resolve
	slowRef := &models.ReferenceFood{
		Name:         "batata frita",
		BaseQuantity: 100.0,
		Unit:         "g",
		Macros: models.Macros{
			Calories: 360,
			Protein:  4.0,
			Carbs:    46.0,
			Fat:      16.0,
		},
	}
	slowProv := &mockProvider{
		delay: 500 * time.Millisecond,
		ref:   slowRef,
	}

	engine := NewNutritionEngineWithProviders(mockDB, nil, []NutritionProvider{slowProv, fastProv})

	start := time.Now()
	preview, err := engine.Analyze("100g batata frita")
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if preview == nil {
		t.Fatal("expected non-nil preview")
	}

	// Should complete after all providers (including slow ones) complete to evaluate best candidate
	if duration < 450*time.Millisecond {
		t.Errorf("expected parallel resolution to wait for all candidates, took %v", duration)
	}

	// Should select the fast provider's result (Calories = 320)
	if preview.Calories != 320 {
		t.Errorf("expected fast provider calories of 320, got %f", preview.Calories)
	}
}
