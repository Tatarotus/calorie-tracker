package services

import (
	"calorie-tracker/config"
	"calorie-tracker/db"
	"calorie-tracker/models"
	"fmt"
	"testing"
	"time"
)

func TestTrackerService_NewTrackerService(t *testing.T) {
	mockDB := db.NewMockDB()
	cfg := &config.Config{}
	llm := NewLLMService(cfg)
	tracker := NewTrackerService(mockDB, llm)

	if tracker == nil {
		t.Fatal("Expected tracker to be non-nil")
	}
	if tracker.db != mockDB {
		t.Error("Tracker DB not set correctly")
	}
}

func TestTrackerService_AddWater(t *testing.T) {
	mockDB := db.NewMockDB()
	tracker := NewTrackerService(mockDB, nil)

	err := tracker.AddWater(500)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	water := mockDB.GetWaterEntries()
	if len(water) != 1 {
		t.Errorf("Expected 1 water entry, got %d", len(water))
	}
	if water[0].AmountML != 500 {
		t.Errorf("Expected 500ml, got %f", water[0].AmountML)
	}
}

func TestTrackerService_SetGoal(t *testing.T) {
	mockDB := db.NewMockDB()
	tracker := NewTrackerService(mockDB, nil)

	err := tracker.SetGoal("Lose weight")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	goal, _ := tracker.GetGoal()
	if goal != "Lose weight" {
		t.Errorf("Expected goal 'Lose weight', got %s", goal)
	}
}

func TestTrackerService_GetGoal_NoGoal(t *testing.T) {
	mockDB := db.NewMockDB()
	tracker := NewTrackerService(mockDB, nil)

	goal, err := tracker.GetGoal()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if goal != "No goal set" {
		t.Errorf("Expected 'No goal set', got %s", goal)
	}
}

func TestTrackerService_GetGoal_WithGoal(t *testing.T) {
	mockDB := db.NewMockDB()
	mockDB.SetGoal(models.Goal{Description: "Gain muscle"})
	tracker := NewTrackerService(mockDB, nil)

	goal, err := tracker.GetGoal()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if goal != "Gain muscle" {
		t.Errorf("Expected 'Gain muscle', got %s", goal)
	}
}

func TestTrackerService_RemoveLastEntry(t *testing.T) {
	mockDB := db.NewMockDB()
	mockDB.AddFoodEntry(models.FoodEntry{Description: "Apple", Timestamp: time.Now()})
	tracker := NewTrackerService(mockDB, nil)

	err := tracker.RemoveLastEntry()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	food := mockDB.GetFoodEntries()
	if len(food) != 0 {
		t.Errorf("Expected 0 food entries, got %d", len(food))
	}
}

func TestTrackerService_GetTodayFoodEntries(t *testing.T) {
	mockDB := db.NewMockDB()
	mockDB.AddFoodEntry(models.FoodEntry{Description: "Apple", Timestamp: time.Now()})
	tracker := NewTrackerService(mockDB, nil)

	entries, err := tracker.GetTodayFoodEntries()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(entries))
	}
}

func TestTrackerService_GetFoodEntriesRange(t *testing.T) {
	mockDB := db.NewMockDB()
	mockDB.AddFoodEntry(models.FoodEntry{Description: "Apple", Timestamp: time.Now()})
	tracker := NewTrackerService(mockDB, nil)

	entries, err := tracker.GetFoodEntriesRange(7)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(entries))
	}
}

func TestTrackerService_GetDailyStats(t *testing.T) {
	mockDB := db.NewMockDB()
	now := time.Now()
	mockDB.AddFoodEntry(models.FoodEntry{
		Description: "Apple",
		Timestamp:   now,
		Calories:    95,
	})
	mockDB.AddWaterEntry(models.WaterEntry{
		Timestamp: now,
		AmountML:  500,
	})
	tracker := NewTrackerService(mockDB, nil)

	stats, err := tracker.GetDailyStats(now)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if stats.Calories != 95 {
		t.Errorf("Expected 95 calories, got %f", stats.Calories)
	}
	if stats.WaterML != 500 {
		t.Errorf("Expected 500ml water, got %f", stats.WaterML)
	}
}

func TestTrackerService_SaveFood(t *testing.T) {
	mockDB := db.NewMockDB()
	tracker := NewTrackerService(mockDB, nil)

	preview := &models.FoodPreview{
		Description: "Super Food",
		Calories:    500,
		Protein:     20,
		Carbs:       50,
		Fat:         10,
	}

	err := tracker.SaveFood(preview)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	food := mockDB.GetFoodEntries()
	found := false
	for _, e := range food {
		if e.Description == "Super Food" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Food entry not found in database")
	}
}

func TestTrackerService_ParseFood_CacheFirst(t *testing.T) {
	mockDB := db.NewMockDB()
	// Add to cache
	mockDB.CacheFood(models.ReferenceFood{
		Name:         "apple",
		BaseQuantity: 100,
		Unit:         "gram",
		Macros: models.Macros{
			Calories: 95,
		},
	})

	// Setup LLM that would fail if called
	server := MockHTTPServerError(500)
	defer server.Close()

	cfg := &config.Config{
		SambaAPIKey:   "test-key",
		OpenAIBaseURL: server.URL,
	}
	llm := NewLLMServiceWithClient(cfg, server.Client())
	tracker := NewTrackerService(mockDB, llm)

	// Should match "apple" in cache and NOT call LLM
	preview, err := tracker.ParseFood("apple")
	if err != nil {
		t.Fatalf("ParseFood failed: %v", err)
	}
	if preview.Calories != 95 {
		t.Errorf("Expected 95 calories from cache, got %f", preview.Calories)
	}
}

func TestTrackerService_ParseFood_LLMFallback(t *testing.T) {
	mockDB := db.NewMockDB()

	// Mock LLM success using ReferenceFood structure
	server := MockHTTPServer(`{"name": "banana", "base_quantity": 100, "unit": "g", "macros": {"calories": 100, "protein": 1, "carbs": 25, "fat": 0}}`)
	defer server.Close()

	cfg := &config.Config{
		SambaAPIKey:   "test-key",
		OpenAIBaseURL: server.URL,
		FoodModel:     "test",
	}
	llm := NewLLMServiceWithClient(cfg, server.Client())
	tracker := NewTrackerService(mockDB, llm)

	preview, err := tracker.ParseFood("banana")
	if err != nil {
		t.Fatalf("ParseFood failed: %v", err)
	}
	if preview.Calories != 100 {
		t.Errorf("Expected 100 calories from LLM, got %f", preview.Calories)
	}
}

func TestTrackerService_RunReview_Success(t *testing.T) {
	mockDB := db.NewMockDB()
	mockDB.SetGoal(models.Goal{Description: "Lose weight"})

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, now.Location())
	yesterday := today.AddDate(0, 0, -1)

	mockDB.AddFoodEntry(models.FoodEntry{
		Description: "Apple",
		Timestamp:   today,
		Calories:    95,
	})
	mockDB.AddFoodEntry(models.FoodEntry{
		Description: "Banana",
		Timestamp:   yesterday,
		Calories:    105,
	})
	mockDB.AddWaterEntry(models.WaterEntry{
		Timestamp: today,
		AmountML:  500,
	})

	server := MockHTTPServer(`{"summary": "Good progress", "goal_progress": "On track", "progress": "improving", "score": 70, "issues": ["Evening snacks"], "suggestions": ["Eat more veggies"], "patterns": ["Higher calories on weekends"]}`)
	defer server.Close()

	cfg := &config.Config{
		SambaAPIKey:   "test-key",
		OpenAIBaseURL: server.URL,
		ReviewModel:   "test",
	}
	llm := NewLLMServiceWithClient(cfg, server.Client())
	tracker := NewTrackerService(mockDB, llm)

	result, err := tracker.RunReview()
	if err != nil {
		t.Fatalf("RunReview failed: %v", err)
	}
	if result == nil {
		t.Fatal("Expected non-nil review result")
	}
}

func TestTrackerService_RunReview_NoGoal(t *testing.T) {
	mockDB := db.NewMockDB()

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, now.Location())

	mockDB.AddFoodEntry(models.FoodEntry{
		Description: "Apple",
		Timestamp:   today,
		Calories:    95,
	})

	server := MockHTTPServer(`{"summary": "Okay", "goal_progress": "No goal set", "progress": "stable", "score": 50, "issues": [], "suggestions": [], "patterns": []}`)
	defer server.Close()

	cfg := &config.Config{
		SambaAPIKey:   "test-key",
		OpenAIBaseURL: server.URL,
		ReviewModel:   "test",
	}
	llm := NewLLMServiceWithClient(cfg, server.Client())
	tracker := NewTrackerService(mockDB, llm)

	result, err := tracker.RunReview()
	if err != nil {
		t.Fatalf("RunReview failed: %v", err)
	}
	if result == nil {
		t.Fatal("Expected non-nil review result even without goal")
	}
}

func TestTrackerService_DailyStatsDateFormatting(t *testing.T) {
	mockDB := db.NewMockDB()

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, now.Location())
	yesterday := today.AddDate(0, 0, -1)

	mockDB.AddFoodEntry(models.FoodEntry{
		Description: "Apple",
		Timestamp:   today,
		Calories:    95,
	})
	mockDB.AddFoodEntry(models.FoodEntry{
		Description: "Banana",
		Timestamp:   yesterday,
		Calories:    105,
	})

	tracker := NewTrackerService(mockDB, nil)

	todayStats, err := tracker.GetDailyStats(today)
	if err != nil {
		t.Fatalf("GetDailyStats failed: %v", err)
	}
	if todayStats.Calories != 95 {
		t.Errorf("Expected 95 calories for today, got %f", todayStats.Calories)
	}

	yesterdayStats, err := tracker.GetDailyStats(yesterday)
	if err != nil {
		t.Fatalf("GetDailyStats failed: %v", err)
	}
	if yesterdayStats.Calories != 105 {
		t.Errorf("Expected 105 calories for yesterday, got %f", yesterdayStats.Calories)
	}
}

func TestTrackerService_ErrorHandling_AddFoodEntry(t *testing.T) {
	mockDB := db.NewMockDB()
	mockDB.SetError("AddFoodEntry", fmt.Errorf("db error"))
	tracker := NewTrackerService(mockDB, nil)

	preview := &models.FoodPreview{
		Description: "Test Food",
		Calories:    100,
	}

	err := tracker.SaveFood(preview)
	if err == nil {
		t.Error("Expected error when AddFoodEntry fails")
	}
}

func TestTrackerService_ErrorHandling_GetLatestGoal(t *testing.T) {
	mockDB := db.NewMockDB()
	mockDB.SetError("GetLatestGoal", fmt.Errorf("db error"))
	tracker := NewTrackerService(mockDB, nil)

	_, err := tracker.GetGoal()
	if err == nil {
		t.Error("Expected error when GetLatestGoal fails")
	}
}
