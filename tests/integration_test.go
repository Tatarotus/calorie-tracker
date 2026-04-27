package main_test

import (
	"calorie-tracker/config"
	"calorie-tracker/db"
	"calorie-tracker/models"
	"calorie-tracker/services"
	"testing"
	"time"
)

// TestFullFoodTrackingFlow tests the complete flow from food parsing to persistence
func TestFullFoodTrackingFlow(t *testing.T) {
	// Setup real DB
	testDB, err := db.NewTestDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test DB: %v", err)
	}
	defer testDB.Close()

	// Setup LLM service (will use mock responses in real scenario)
	cfg := &config.Config{
		SambaAPIKey:   "test-key",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := services.NewLLMService(cfg)

	// Create tracker service with real DB
	tracker := services.NewTrackerService(testDB, llm)

	// Step 1: Parse food (this would normally call LLM, but we'll test the flow)
	// For this test, we'll directly save a food preview to simulate LLM parsing
	preview := &models.FoodPreview{
		Description: "Grilled Chicken Salad",
		Calories:    350,
		Protein:     30,
		Carbs:       15,
		Fat:         18,
	}

	// Step 2: Save food to DB
	err = tracker.SaveFood(preview)
	if err != nil {
		t.Fatalf("Failed to save food: %v", err)
	}

	// Step 3: Verify food was saved
	entries, err := testDB.GetDailyFoodEntries(time.Now())
	if err != nil {
		t.Fatalf("Failed to get entries: %v", err)
	}

	if len(entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(entries))
	}

	entry := entries[0]
	if entry.Description != "Grilled Chicken Salad" {
		t.Errorf("Expected 'Grilled Chicken Salad', got '%s'", entry.Description)
	}
	if entry.Calories != 350 {
		t.Errorf("Expected 350 calories, got %f", entry.Calories)
	}

	// Step 4: Verify cache was updated
	cached, err := testDB.GetCachedFood("Grilled Chicken Salad")
	if err != nil {
		t.Fatalf("Failed to get cached food: %v", err)
	}

	if cached == nil {
		t.Fatal("Expected cached entry")
	}

	if cached.Calories != 350 {
		t.Errorf("Expected cached calories 350, got %f", cached.Calories)
	}
}

// TestWaterTrackingFlow tests water entry flow
func TestWaterTrackingFlow(t *testing.T) {
	testDB, err := db.NewTestDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test DB: %v", err)
	}
	defer testDB.Close()

	cfg := &config.Config{
		SambaAPIKey:   "test-key",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := services.NewLLMService(cfg)
	tracker := services.NewTrackerService(testDB, llm)

	// Add water
	err = tracker.AddWater(500)
	if err != nil {
		t.Fatalf("Failed to add water: %v", err)
	}

	// Verify water was saved
	entries, err := testDB.GetDailyWaterEntries(time.Now())
	if err != nil {
		t.Fatalf("Failed to get water entries: %v", err)
	}

	if len(entries) != 1 {
		t.Errorf("Expected 1 water entry, got %d", len(entries))
	}

	if entries[0].AmountML != 500 {
		t.Errorf("Expected 500ml, got %f", entries[0].AmountML)
	}
}

// TestGoalSettingFlow tests goal setting and retrieval
func TestGoalSettingFlow(t *testing.T) {
	testDB, err := db.NewTestDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test DB: %v", err)
	}
	defer testDB.Close()

	cfg := &config.Config{
		SambaAPIKey:   "test-key",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := services.NewLLMService(cfg)
	tracker := services.NewTrackerService(testDB, llm)

	// Set goal
	err = tracker.SetGoal("Lose 10kg in 6 months")
	if err != nil {
		t.Fatalf("Failed to set goal: %v", err)
	}

	// Retrieve goal
	goal, err := tracker.GetGoal()
	if err != nil {
		t.Fatalf("Failed to get goal: %v", err)
	}

	if goal != "Lose 10kg in 6 months" {
		t.Errorf("Expected 'Lose 10kg in 6 months', got '%s'", goal)
	}
}

// TestDailyStatsAggregationFlow tests stats calculation across multiple entries
func TestDailyStatsAggregationFlow(t *testing.T) {
	testDB, err := db.NewTestDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test DB: %v", err)
	}
	defer testDB.Close()

	cfg := &config.Config{
		SambaAPIKey:   "test-key",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := services.NewLLMService(cfg)
	tracker := services.NewTrackerService(testDB, llm)

	now := time.Now()

	// Add multiple food entries
	foods := []struct {
		desc     string
		calories float64
		protein  float64
		carbs    float64
		fat      float64
	}{
		{"Breakfast", 300, 15, 40, 10},
		{"Lunch", 500, 25, 60, 20},
		{"Dinner", 600, 30, 50, 25},
	}

	for _, f := range foods {
		preview := &models.FoodPreview{
			Description: f.desc,
			Calories:    f.calories,
			Protein:     f.protein,
			Carbs:       f.carbs,
			Fat:         f.fat,
		}
		if err := tracker.SaveFood(preview); err != nil {
			t.Fatalf("Failed to save food %s: %v", f.desc, err)
		}
	}

	// Add water entries
	if err := tracker.AddWater(250); err != nil {
		t.Fatalf("Failed to add water: %v", err)
	}
	if err := tracker.AddWater(500); err != nil {
		t.Fatalf("Failed to add water: %v", err)
	}

	// Get daily stats
	stats, err := tracker.GetDailyStats(now)
	if err != nil {
		t.Fatalf("Failed to get daily stats: %v", err)
	}

	// Verify aggregation
	expectedCalories := 1400.0 // 300 + 500 + 600
	expectedProtein := 70.0    // 15 + 25 + 30
	expectedCarbs := 150.0     // 40 + 60 + 50
	expectedFat := 55.0        // 10 + 20 + 25
	expectedWater := 750.0     // 250 + 500

	if stats.Calories != expectedCalories {
		t.Errorf("Expected %.0f calories, got %.0f", expectedCalories, stats.Calories)
	}
	if stats.Protein != expectedProtein {
		t.Errorf("Expected %.0f protein, got %.0f", expectedProtein, stats.Protein)
	}
	if stats.Carbs != expectedCarbs {
		t.Errorf("Expected %.0f carbs, got %.0f", expectedCarbs, stats.Carbs)
	}
	if stats.Fat != expectedFat {
		t.Errorf("Expected %.0f fat, got %.0f", expectedFat, stats.Fat)
	}
	if stats.WaterML != expectedWater {
		t.Errorf("Expected %.0f water, got %.0f", expectedWater, stats.WaterML)
	}
}

// TestUndoLastEntryFlow tests removing the last entry
func TestUndoLastEntryFlow(t *testing.T) {
	testDB, err := db.NewTestDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test DB: %v", err)
	}
	defer testDB.Close()

	cfg := &config.Config{
		SambaAPIKey:   "test-key",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := services.NewLLMService(cfg)
	tracker := services.NewTrackerService(testDB, llm)

	// Add two food entries
	preview1 := &models.FoodPreview{Description: "Entry 1", Calories: 100}
	preview2 := &models.FoodPreview{Description: "Entry 2", Calories: 200}

	tracker.SaveFood(preview1)
	tracker.SaveFood(preview2)

	// Verify both entries exist
	entries, _ := testDB.GetDailyFoodEntries(time.Now())
	if len(entries) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(entries))
	}

	// Remove last entry
	err = tracker.RemoveLastEntry()
	if err != nil {
		t.Fatalf("Failed to remove last entry: %v", err)
	}

	// Verify only one entry remains
	entries, _ = testDB.GetDailyFoodEntries(time.Now())
	if len(entries) != 1 {
		t.Errorf("Expected 1 entry after removal, got %d", len(entries))
	}
}

// TestMultipleDaysStatsFlow tests stats across multiple days
func TestMultipleDaysStatsFlow(t *testing.T) {
	testDB, err := db.NewTestDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test DB: %v", err)
	}
	defer testDB.Close()

	cfg := &config.Config{
		SambaAPIKey:   "test-key",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := services.NewLLMService(cfg)
	tracker := services.NewTrackerService(testDB, llm)

	// Add entries for different days
	for i := 0; i < 3; i++ {
		date := time.Now().AddDate(0, 0, -i)
		preview := &models.FoodPreview{
			Description: "Day " + string(rune('0'+i)),
			Calories:    500 * float64(i+1),
		}
		// Manually set timestamp by directly adding to DB
		entry := models.FoodEntry{
			Timestamp:   date,
			Description: preview.Description,
			Calories:    preview.Calories,
		}
		testDB.AddFoodEntry(entry)
	}

	// Get food entries range
	entries, err := tracker.GetFoodEntriesRange(7)
	if err != nil {
		t.Fatalf("Failed to get entries range: %v", err)
	}

	if len(entries) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(entries))
	}
}
