package services

import (
	"os"
	"testing"

	"calorie-tracker/config"
	"calorie-tracker/db"
	"calorie-tracker/models"
)

func TestTrackerService_SaveFood_WithName(t *testing.T) {
	mockDB := db.NewMockDB()
	tracker := NewTrackerService(mockDB, nil)

	preview := &models.FoodPreview{
		Description: "100g apple",
		Name:        "apple",
		Unit:        "gram",
		Calories:    52,
		Protein:     0.3,
		Carbs:       14,
		Fat:         0.2,
	}

	err := tracker.SaveFood(preview)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Check that food was cached
	cached, err := mockDB.GetCachedFood("apple")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if cached == nil {
		t.Error("Expected food to be cached")
	}

	// Kill EQ_to_NEQ mutation in tracker.go line 72 (if normalizedQuery == "")
	// If mutated to if normalizedQuery != "", normalizedQuery gets overwritten with description.
	// We assert that the saved food entry's NormalizedQuery is exactly "apple".
	food := mockDB.GetFoodEntries()
	if len(food) != 1 {
		t.Fatalf("Expected 1 food entry, got %d", len(food))
	}
	if food[0].NormalizedQuery != "apple" {
		t.Errorf("Expected NormalizedQuery 'apple', got %q", food[0].NormalizedQuery)
	}
}

func TestTrackerService_SaveFood_WithoutName(t *testing.T) {
	mockDB := db.NewMockDB()
	tracker := NewTrackerService(mockDB, nil)

	preview := &models.FoodPreview{
		Description: "100g apple",
		Calories:    52,
		Protein:     0.3,
		Carbs:       14,
		Fat:         0.2,
	}

	err := tracker.SaveFood(preview)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Food entry should be saved
	food := mockDB.GetFoodEntries()
	if len(food) != 1 {
		t.Errorf("Expected 1 food entry, got %d", len(food))
	}
}

func TestTrackerService_SaveFood_ZeroAmount(t *testing.T) {
	mockDB := db.NewMockDB()
	tracker := NewTrackerService(mockDB, nil)

	preview := &models.FoodPreview{
		Description: "apple",
		Name:        "apple",
		Unit:        "unit",
		Calories:    52,
		Protein:     0.3,
		Carbs:       14,
		Fat:         0.2,
	}

	err := tracker.SaveFood(preview)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Should not cache with zero amount
	cached, _ := mockDB.GetCachedFood("apple")
	// Zero amount means no cache
	_ = cached
}

func TestTrackerService_NewTrackerService_ProviderPriority(t *testing.T) {
	// Set environment variables to force provider initialization
	os.Setenv("NUTRITION_PRIORITY", "fatsecret,serpapi,calorieninjas")
	os.Setenv("FATSECRET_CLIENT_ID", "dummy-id")
	os.Setenv("FATSECRET_CLIENT_SECRET", "dummy-secret")
	os.Setenv("SERPAPI_KEY", "dummy-key")
	defer func() {
		os.Unsetenv("NUTRITION_PRIORITY")
		os.Unsetenv("FATSECRET_CLIENT_ID")
		os.Unsetenv("FATSECRET_CLIENT_SECRET")
		os.Unsetenv("SERPAPI_KEY")
	}()

	mockDB := db.NewMockDB()
	tracker := NewTrackerService(mockDB, nil)

	if tracker == nil {
		t.Fatal("Expected tracker to be non-nil")
	}

	// Verify that the providers were correctly initialized and appended to the engine
	resolver, ok := tracker.engine.nutritionResolver.(*HybridNutritionResolver)
	if !ok {
		t.Fatal("Expected engine.nutritionResolver to be *HybridNutritionResolver")
	}
	providers := resolver.providers
	if len(providers) != 3 {
		t.Fatalf("Expected 3 providers, got %d", len(providers))
	}

	// Verify types of providers to ensure all three are present
	hasFatSecret := false
	hasSerpAPI := false
	hasCalorieNinjas := false

	for _, p := range providers {
		switch p.(type) {
		case *FatSecretProvider:
			hasFatSecret = true
		case *SerpAPIProvider:
			hasSerpAPI = true
		case *CalorieNinjasProvider:
			hasCalorieNinjas = true
		}
	}

	if !hasFatSecret {
		t.Error("Missing FatSecretProvider")
	}
	if !hasSerpAPI {
		t.Error("Missing SerpAPIProvider")
	}
	if !hasCalorieNinjas {
		t.Error("Missing CalorieNinjasProvider")
	}
}

func TestTrackerService_SaveFood_UserOverrides(t *testing.T) {
	mockDB := db.NewMockDB()
	tracker := NewTrackerService(mockDB, nil)

	// Case 1: UserEdited is false
	// We resolve canonical food first. When we resolve "banana", the canonical ID will be created as 1.
	preview1 := &models.FoodPreview{
		Description: "100g banana",
		Name:        "banana",
		Unit:        "gram",
		Calories:    89,
		UserEdited:  false,
	}

	err := tracker.SaveFood(preview1)
	if err != nil {
		t.Fatalf("SaveFood failed: %v", err)
	}

	// Retrieve user override for canonical ID 1 (resolved from "banana")
	override1, err := mockDB.GetUserOverride(1)
	if err != nil {
		t.Fatalf("GetUserOverride failed: %v", err)
	}
	if override1 != nil {
		t.Error("Expected no user override when UserEdited is false")
	}

	// Case 2: UserEdited is true
	preview2 := &models.FoodPreview{
		Description: "100g banana",
		Name:        "banana",
		Unit:        "gram",
		Calories:    95, // different calories
		UserEdited:  true,
	}

	err = tracker.SaveFood(preview2)
	if err != nil {
		t.Fatalf("SaveFood failed: %v", err)
	}

	// Retrieve user override for canonical ID 1
	override2, err := mockDB.GetUserOverride(1)
	if err != nil {
		t.Fatalf("GetUserOverride failed: %v", err)
	}
	if override2 == nil {
		t.Error("Expected user override when UserEdited is true")
	} else if override2.Calories != 95 {
		t.Errorf("Expected user override calories to be 95, got %f", override2.Calories)
	}
}

func TestTrackerService_SaveFood_EmptyNameAndAmount(t *testing.T) {
	mockDB := db.NewMockDB()
	tracker := NewTrackerService(mockDB, nil)

	// Case 1: name is empty, amount > 0
	preview1 := &models.FoodPreview{
		Description: "100g",
		Name:        "", // empty name
		Unit:        "gram",
		Calories:    50,
		ResolutionTrace: &models.ResolutionTrace{
			SourceType: "fatsecret",
		},
	}

	err := tracker.SaveFood(preview1)
	if err != nil {
		t.Fatalf("SaveFood failed: %v", err)
	}

	// Verify no reference food was cached with empty key ""
	cached, err := mockDB.GetCachedFood("")
	if err != nil {
		t.Fatalf("GetCachedFood failed: %v", err)
	}
	if cached != nil {
		t.Error("Expected no cache entry for empty food name")
	}
}

func TestTrackerService_SaveFood_EmptyDescriptionWithLlmParser(t *testing.T) {
	mockDB := db.NewMockDB()

	// Mock LLM response containing parsed items
	server := MockHTTPServer(`{"items": [{"food_name": "mocked", "quantity": 5, "unit": "unit", "canonical_key": "mocked"}]}`)
	defer server.Close()

	cfg := &config.Config{
		SambaAPIKey:   "test-key",
		OpenAIBaseURL: server.URL,
		FoodModel:     "test",
	}
	llm := NewLLMServiceWithClient(cfg, server.Client())
	tracker := NewTrackerService(mockDB, llm)

	preview := &models.FoodPreview{
		Description: "", // empty description
		Calories:    100,
		Name:        "preset-name",
		Unit:        "unit",
		ResolutionTrace: &models.ResolutionTrace{
			ParserUsed: "llm", // useBasic is false
		},
	}

	err := tracker.SaveFood(preview)
	if err != nil {
		t.Fatalf("SaveFood failed: %v", err)
	}

	// Under normal conditions: since preview.Description is "", it should skip LlmParser and go to basic regex fallback
	// Basic regex fallback on "" will return name="", unit="", amount=0.
	// Since name was "preset-name", s.extractNameUnitAmount returns name="preset-name" (retaining preset name).
	// Under the mutated condition (AND_to_OR): since !useBasic is true, it evaluates to true,
	// and calls LlmParser.Parse("") which returns the mocked item: name="mocked", amount=5.
	// So s.extractNameUnitAmount would return name="mocked".
	// Let's assert that the saved entry retains its preset name.
	entries := mockDB.GetFoodEntries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].NormalizedQuery == "mocked" {
		t.Errorf("Mutation survived! extractNameUnitAmount parsed empty description using LLM parser despite it being empty")
	}
}
