package services

import (
	"testing"
	"time"

	"calorie-tracker/config"
	"calorie-tracker/db"
	"calorie-tracker/models"
)

func TestTrackerService_NewTrackerService(t *testing.T) {
	mockDB := db.NewMockDB()
	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
		FoodModel:     "test",
		ReviewModel:   "test",
	}
	llm := NewLLMService(cfg)

	svc := NewTrackerService(mockDB, llm)

	if svc == nil {
		t.Fatal("Expected non-nil TrackerService")
	}
	if svc.db == nil {
		t.Error("Expected non-nil db")
	}
	if svc.llm == nil {
		t.Error("Expected non-nil llm")
	}
	if svc.matcher == nil {
		t.Error("Expected non-nil matcher")
	}
}

func TestTrackerService_AddWater(t *testing.T) {
	mockDB := db.NewMockDB()
	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := NewLLMService(cfg)
	svc := NewTrackerService(mockDB, llm)

	err := svc.AddWater(250)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	entries := mockDB.GetWaterEntries()
	if len(entries) != 1 {
		t.Errorf("Expected 1 water entry, got %d", len(entries))
	}

	if entries[0].AmountML != 250 {
		t.Errorf("Expected AmountML 250, got %f", entries[0].AmountML)
	}
}

func TestTrackerService_SetGoal(t *testing.T) {
	mockDB := db.NewMockDB()
	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := NewLLMService(cfg)
	svc := NewTrackerService(mockDB, llm)

	err := svc.SetGoal("Lose 5kg")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	goal, _ := mockDB.GetLatestGoal()
	if goal == nil {
		t.Fatal("Expected goal to be set")
	}

	if goal.Description != "Lose 5kg" {
		t.Errorf("Expected 'Lose 5kg', got %s", goal.Description)
	}
}

func TestTrackerService_GetGoal_NoGoal(t *testing.T) {
	mockDB := db.NewMockDB()
	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := NewLLMService(cfg)
	svc := NewTrackerService(mockDB, llm)

	goal, err := svc.GetGoal()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if goal != "No goal set" {
		t.Errorf("Expected 'No goal set', got %s", goal)
	}
}

func TestTrackerService_GetGoal_WithGoal(t *testing.T) {
	mockDB := db.NewMockDB()
	mockDB.SetGoal(models.Goal{
		Description: "Gain muscle",
	})
	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := NewLLMService(cfg)
	svc := NewTrackerService(mockDB, llm)

	goal, err := svc.GetGoal()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if goal != "Gain muscle" {
		t.Errorf("Expected 'Gain muscle', got %s", goal)
	}
}

func TestTrackerService_RemoveLastEntry(t *testing.T) {
	mockDB := db.NewMockDB()
	mockDB.AddFoodEntry(models.FoodEntry{Description: "Entry 1"})
	mockDB.AddFoodEntry(models.FoodEntry{Description: "Entry 2"})

	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := NewLLMService(cfg)
	svc := NewTrackerService(mockDB, llm)

	err := svc.RemoveLastEntry()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	entries := mockDB.GetFoodEntries()
	if len(entries) != 1 {
		t.Errorf("Expected 1 food entry, got %d", len(entries))
	}

	if entries[0].Description != "Entry 1" {
		t.Errorf("Expected 'Entry 1', got %s", entries[0].Description)
	}
}

func TestTrackerService_GetTodayFoodEntries(t *testing.T) {
	mockDB := db.NewMockDB()
	mockDB.AddFoodEntry(models.FoodEntry{
		Description: "Breakfast",
		Calories:    300,
		Timestamp:   time.Now(),
	})
	mockDB.AddFoodEntry(models.FoodEntry{
		Description: "Lunch",
		Calories:    500,
		Timestamp:   time.Now(),
	})

	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := NewLLMService(cfg)
	svc := NewTrackerService(mockDB, llm)

	entries, err := svc.GetTodayFoodEntries()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(entries) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(entries))
	}
}

func TestTrackerService_GetFoodEntriesRange(t *testing.T) {
	mockDB := db.NewMockDB()
	mockDB.AddFoodEntry(models.FoodEntry{Description: "Day 1"})
	mockDB.AddFoodEntry(models.FoodEntry{Description: "Day 2"})

	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := NewLLMService(cfg)
	svc := NewTrackerService(mockDB, llm)

	entries, err := svc.GetFoodEntriesRange(7)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(entries) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(entries))
	}
}

func TestTrackerService_GetDailyStats(t *testing.T) {
	mockDB := db.NewMockDB()
	now := time.Now()
	mockDB.AddFoodEntry(models.FoodEntry{
		Calories: 100,
		Protein:  5,
		Carbs:    20,
		Fat:      2,
		Timestamp: now,
	})
	mockDB.AddFoodEntry(models.FoodEntry{
		Calories: 200,
		Protein:  10,
		Carbs:    30,
		Fat:      5,
		Timestamp: now,
	})
	mockDB.AddWaterEntry(models.WaterEntry{
		AmountML: 250,
		Timestamp: now,
	})
	mockDB.AddWaterEntry(models.WaterEntry{
		AmountML: 500,
		Timestamp: now,
	})

	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := NewLLMService(cfg)
	svc := NewTrackerService(mockDB, llm)

	stats, err := svc.GetDailyStats(now)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	expectedCalories := 300.0
	expectedProtein := 15.0
	expectedCarbs := 50.0
	expectedFat := 7.0
	expectedWater := 750.0

	if stats.Calories != expectedCalories {
		t.Errorf("Expected Calories %f, got %f", expectedCalories, stats.Calories)
	}
	if stats.Protein != expectedProtein {
		t.Errorf("Expected Protein %f, got %f", expectedProtein, stats.Protein)
	}
	if stats.Carbs != expectedCarbs {
		t.Errorf("Expected Carbs %f, got %f", expectedCarbs, stats.Carbs)
	}
	if stats.Fat != expectedFat {
		t.Errorf("Expected Fat %f, got %f", expectedFat, stats.Fat)
	}
	if stats.WaterML != expectedWater {
		t.Errorf("Expected WaterML %f, got %f", expectedWater, stats.WaterML)
	}
}

func TestTrackerService_SaveFood(t *testing.T) {
	mockDB := db.NewMockDB()
	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := NewLLMService(cfg)
	svc := NewTrackerService(mockDB, llm)

	preview := &models.FoodPreview{
		Description: "Test Food",
		Calories:    100,
		Protein:     5,
		Carbs:       20,
		Fat:         2,
	}

	err := svc.SaveFood(preview)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	entries := mockDB.GetFoodEntries()
	if len(entries) != 1 {
		t.Errorf("Expected 1 food entry, got %d", len(entries))
	}

	entry := entries[0]
	if entry.Description != "Test Food" {
		t.Errorf("Expected 'Test Food', got %s", entry.Description)
	}
	if entry.Calories != 100 {
		t.Errorf("Expected Calories 100, got %f", entry.Calories)
	}
}

func TestTrackerService_ParseFood_Matched(t *testing.T) {
	mockDB := db.NewMockDB()
	// Add a cached entry
	mockDB.AddFoodEntry(models.FoodEntry{
		Description: "apple",
		Calories:    95,
		Protein:     0.5,
		Carbs:       25,
		Fat:         0.3,
	})
	mockDB.CacheFood(models.FoodEntry{
		Description: "apple",
		Calories:    95,
		Protein:     0.5,
		Carbs:       25,
		Fat:         0.3,
	})

	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := NewLLMService(cfg)
	svc := NewTrackerService(mockDB, llm)

	preview, err := svc.ParseFood("apple")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if preview == nil {
		t.Fatal("Expected non-nil preview")
	}

	if preview.Calories != 95 {
		t.Errorf("Expected Calories 95, got %f", preview.Calories)
	}
}

func TestTrackerService_DailyStatsDateFormatting(t *testing.T) {
	mockDB := db.NewMockDB()
	testDate := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)

	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := NewLLMService(cfg)
	svc := NewTrackerService(mockDB, llm)

	stats, err := svc.GetDailyStats(testDate)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if stats.Date != "2024-01-15" {
		t.Errorf("Expected date '2024-01-15', got %s", stats.Date)
	}
}

func TestTrackerService_ErrorHandling_AddFoodEntry(t *testing.T) {
	mockDB := db.NewMockDB()
	mockDB.SetError("AddFoodEntry", testError("database error"))

	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := NewLLMService(cfg)
	svc := NewTrackerService(mockDB, llm)

	preview := &models.FoodPreview{
		Description: "Test",
		Calories:    100,
	}

	err := svc.SaveFood(preview)
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestTrackerService_ErrorHandling_GetLatestGoal(t *testing.T) {
	mockDB := db.NewMockDB()
	mockDB.SetError("GetLatestGoal", testError("database error"))

	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := NewLLMService(cfg)
	svc := NewTrackerService(mockDB, llm)

	_, err := svc.GetGoal()
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

// testError is a simple error type for testing
type testError string

func (e testError) Error() string {
	return string(e)
}
