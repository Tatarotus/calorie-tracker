package tests

import (
	"calorie-tracker/config"
	"calorie-tracker/db"
	"calorie-tracker/models"
	"calorie-tracker/services"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestFullFoodTrackingFlow tests the end-to-end flow of tracking food
func TestFullFoodTrackingFlow(t *testing.T) {
	// Step 1: Setup
	testDB, err := db.NewTestDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test DB: %v", err)
	}
	defer testDB.Close()

	cfg := &config.Config{
		SambaAPIKey:   "test-key",
		OpenAIBaseURL: "https://test.com/v1",
		FoodModel:     "test-model",
	}

	// Step 2: Parse a food item
	// We need a mock server that returns a ReferenceFood
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"choices":[{"message":{"content":"{\"name\":\"apple\",\"base_quantity\":100,\"unit\":\"g\",\"macros\":{\"calories\":350,\"protein\":1,\"carbs\":25,\"fat\":0}}"}}]}`)
	}))
	defer server.Close()

	cfg.OpenAIBaseURL = server.URL
	llm := services.NewLLMServiceWithClient(cfg, server.Client())
	tracker := services.NewTrackerService(testDB, llm)

	preview, err := tracker.ParseFood("100g apple")
	if err != nil {
		t.Fatalf("ParseFood failed: %v", err)
	}

	if preview.Calories != 350 {
		t.Errorf("Expected 350 calories, got %f", preview.Calories)
	}

	// Step 3: Save the food
	err = tracker.SaveFood(preview)
	if err != nil {
		t.Fatalf("SaveFood failed: %v", err)
	}

	// Verify it's in the database
	entries, err := testDB.GetDailyFoodEntries(time.Now())
	if err != nil {
		t.Fatalf("Failed to get entries: %v", err)
	}

	if len(entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(entries))
	}

	entry := entries[0]
	if entry.Description != "100.0g apple" {
		t.Errorf("Expected '100.0g apple', got '%s'", entry.Description)
	}
	if entry.Calories != 350 {
		t.Errorf("Expected 350 calories, got %f", entry.Calories)
	}

	// Step 4: Verify cache was updated
	cached, err := testDB.GetCachedFood("apple")
	if err != nil {
		t.Fatalf("Failed to get cached food: %v", err)
	}

	if cached == nil {
		t.Fatal("Expected cached entry")
	}

	if cached.Macros.Calories != 350 {
		t.Errorf("Expected cached calories 350, got %f", cached.Macros.Calories)
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

	// Verify
	stats, err := tracker.GetDailyStats(time.Now())
	if err != nil {
		t.Fatalf("Failed to get stats: %v", err)
	}
	if stats.WaterML != 500 {
		t.Errorf("Expected 500ml water, got %f", stats.WaterML)
	}
}

// TestGoalSettingFlow tests goal setting
func TestGoalSettingFlow(t *testing.T) {
	testDB, err := db.NewTestDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test DB: %v", err)
	}
	defer testDB.Close()

	tracker := services.NewTrackerService(testDB, nil)

	// Set goal
	err = tracker.SetGoal("Lose 5kg")
	if err != nil {
		t.Fatalf("Failed to set goal: %v", err)
	}

	// Get goal
	goal, err := tracker.GetGoal()
	if err != nil {
		t.Fatalf("Failed to get goal: %v", err)
	}
	if goal != "Lose 5kg" {
		t.Errorf("Expected goal 'Lose 5kg', got %s", goal)
	}
}

// TestDailyStatsAggregationFlow tests stats aggregation over multiple entries
func TestDailyStatsAggregationFlow(t *testing.T) {
	testDB, err := db.NewTestDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test DB: %v", err)
	}
	defer testDB.Close()

	tracker := services.NewTrackerService(testDB, nil)
	now := time.Now()

	// Add entries directly to DB for simplicity
	testDB.AddFoodEntry(models.FoodEntry{
		Timestamp: now,
		Calories:  300,
		Protein:   20,
	})
	testDB.AddFoodEntry(models.FoodEntry{
		Timestamp: now,
		Calories:  500,
		Protein:   30,
	})
	testDB.AddWaterEntry(models.WaterEntry{
		Timestamp: now,
		AmountML:  250,
	})

	stats, err := tracker.GetDailyStats(now)
	if err != nil {
		t.Fatalf("Failed to get stats: %v", err)
	}

	if stats.Calories != 800 {
		t.Errorf("Expected 800 calories, got %f", stats.Calories)
	}
	if stats.Protein != 50 {
		t.Errorf("Expected 50g protein, got %f", stats.Protein)
	}
	if stats.WaterML != 250 {
		t.Errorf("Expected 250ml water, got %f", stats.WaterML)
	}
}

// TestUndoLastEntryFlow tests the undo functionality
func TestUndoLastEntryFlow(t *testing.T) {
	testDB, err := db.NewTestDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test DB: %v", err)
	}
	defer testDB.Close()

	tracker := services.NewTrackerService(testDB, nil)

	// Add two entries
	testDB.AddFoodEntry(models.FoodEntry{Description: "Entry 1", Timestamp: time.Now().Add(-1 * time.Minute)})
	testDB.AddFoodEntry(models.FoodEntry{Description: "Entry 2", Timestamp: time.Now()})

	// Undo
	err = tracker.RemoveLastEntry()
	if err != nil {
		t.Fatalf("RemoveLastEntry failed: %v", err)
	}

	// Verify only Entry 1 remains
	entries, _ := tracker.GetTodayFoodEntries()
	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry after undo, got %d", len(entries))
	}
	if entries[0].Description != "Entry 1" {
		t.Errorf("Expected 'Entry 1' to remain, got '%s'", entries[0].Description)
	}
}

// TestMultipleDaysStatsFlow tests stats across multiple days
func TestMultipleDaysStatsFlow(t *testing.T) {
	testDB, err := db.NewTestDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test DB: %v", err)
	}
	defer testDB.Close()

	tracker := services.NewTrackerService(testDB, nil)
	now := time.Now()

	// Today
	testDB.AddFoodEntry(models.FoodEntry{Timestamp: now, Calories: 2000})
	// Yesterday
	testDB.AddFoodEntry(models.FoodEntry{Timestamp: now.AddDate(0, 0, -1), Calories: 1800})

	// Check today's stats
	statsToday, _ := tracker.GetDailyStats(now)
	if statsToday.Calories != 2000 {
		t.Errorf("Expected 2000 calories today, got %f", statsToday.Calories)
	}

	// Check yesterday's stats
	statsYesterday, _ := tracker.GetDailyStats(now.AddDate(0, 0, -1))
	if statsYesterday.Calories != 1800 {
		t.Errorf("Expected 1800 calories yesterday, got %f", statsYesterday.Calories)
	}
}
