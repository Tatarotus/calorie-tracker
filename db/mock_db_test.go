package db

import (
	"testing"
	"time"

	"calorie-tracker/models"
)

func TestNewMockDB(t *testing.T) {
	db := NewMockDB()
	if db == nil {
		t.Fatal("Expected non-nil MockDB")
	}
	if db.cache == nil {
		t.Error("Expected cache to be initialized")
	}
	if db.errorOnCall == nil {
		t.Error("Expected errorOnCall to be initialized")
	}
}

func TestMockDB_SetError(t *testing.T) {
	db := NewMockDB()
	testErr := testError("test error")

	db.SetError("AddFoodEntry", testErr)

	err, ok := db.errorOnCall["AddFoodEntry"]
	if !ok {
		t.Error("Expected error to be set")
	}
	if err.Error() != "test error" {
		t.Errorf("Expected 'test error', got %s", err.Error())
	}
}

func TestMockDB_ClearError(t *testing.T) {
	db := NewMockDB()
	testErr := testError("test error")

	db.SetError("AddFoodEntry", testErr)
	db.ClearError("AddFoodEntry")

	_, ok := db.errorOnCall["AddFoodEntry"]
	if ok {
		t.Error("Expected error to be cleared")
	}
}

func TestMockDB_AddFoodEntry(t *testing.T) {
	db := NewMockDB()
	entry := models.FoodEntry{Description: "Test", Calories: 100}

	err := db.AddFoodEntry(entry)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	entries := db.GetFoodEntries()
	if len(entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(entries))
	}
}

func TestMockDB_AddFoodEntry_WithError(t *testing.T) {
	db := NewMockDB()
	db.SetError("AddFoodEntry", testError("db error"))

	entry := models.FoodEntry{Description: "Test"}
	err := db.AddFoodEntry(entry)

	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestMockDB_GetDailyFoodEntries(t *testing.T) {
	db := NewMockDB()
	now := time.Now()
	entry := models.FoodEntry{
		Description: "Test",
		Timestamp:   now,
	}
	db.AddFoodEntry(entry)

	entries, err := db.GetDailyFoodEntries(now)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(entries))
	}
}

func TestMockDB_GetDailyFoodEntries_DifferentDate(t *testing.T) {
	db := NewMockDB()
	now := time.Now()
	entry := models.FoodEntry{
		Description: "Test",
		Timestamp:   now,
	}
	db.AddFoodEntry(entry)

	yesterday := now.AddDate(0, 0, -1)
	entries, err := db.GetDailyFoodEntries(yesterday)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("Expected 0 entries, got %d", len(entries))
	}
}

func TestMockDB_GetFoodEntriesRange(t *testing.T) {
	db := NewMockDB()
	db.AddFoodEntry(models.FoodEntry{Description: "Entry 1"})
	db.AddFoodEntry(models.FoodEntry{Description: "Entry 2"})

	entries, err := db.GetFoodEntriesRange(7)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(entries))
	}
}

func TestMockDB_CacheFood(t *testing.T) {
	db := NewMockDB()
	ref := models.ReferenceFood{
		Name:         "Apple",
		BaseQuantity: 100,
		Unit:         "gram",
		Macros: models.Macros{
			Calories: 95,
		},
	}

	err := db.CacheFood(ref)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	cached, _ := db.GetCachedFood("Apple")
	if cached == nil {
		t.Fatal("Expected cached entry")
	}
	if cached.Macros.Calories != 95 {
		t.Errorf("Expected Calories 95, got %f", cached.Macros.Calories)
	}
}

func TestMockDB_GetCachedFood_NotFound(t *testing.T) {
	db := NewMockDB()

	cached, err := db.GetCachedFood("NonExistent")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if cached != nil {
		t.Error("Expected nil for non-existent entry")
	}
}

func TestMockDB_AddWaterEntry(t *testing.T) {
	db := NewMockDB()
	entry := models.WaterEntry{AmountML: 250}

	err := db.AddWaterEntry(entry)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	entries := db.GetWaterEntries()
	if len(entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(entries))
	}
}

func TestMockDB_GetDailyWaterEntries(t *testing.T) {
	db := NewMockDB()
	now := time.Now()
	entry := models.WaterEntry{
		AmountML:  250,
		Timestamp: now,
	}
	db.AddWaterEntry(entry)

	entries, err := db.GetDailyWaterEntries(now)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(entries))
	}
}

func TestMockDB_GetWaterEntriesRange(t *testing.T) {
	db := NewMockDB()
	db.AddWaterEntry(models.WaterEntry{AmountML: 250})
	db.AddWaterEntry(models.WaterEntry{AmountML: 500})

	entries, err := db.GetWaterEntriesRange(7)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(entries))
	}
}

func TestMockDB_GetStatsRange(t *testing.T) {
	db := NewMockDB()
	now := time.Now()

	db.AddFoodEntry(models.FoodEntry{
		Calories:  100,
		Protein:   5,
		Carbs:     20,
		Fat:       2,
		Timestamp: now,
	})
	db.AddWaterEntry(models.WaterEntry{
		AmountML:  250,
		Timestamp: now,
	})

	stats, err := db.GetStatsRange(7)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(stats) == 0 {
		t.Fatal("Expected at least 1 day of stats")
	}

	dayStats := stats[0]
	if dayStats.Calories != 100 {
		t.Errorf("Expected Calories 100, got %f", dayStats.Calories)
	}
	if dayStats.WaterML != 250 {
		t.Errorf("Expected WaterML 250, got %f", dayStats.WaterML)
	}
}

func TestMockDB_SetGoal(t *testing.T) {
	db := NewMockDB()
	goal := models.Goal{Description: "Lose weight"}

	err := db.SetGoal(goal)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	retrieved, _ := db.GetLatestGoal()
	if retrieved == nil {
		t.Fatal("Expected goal to be set")
	}
	if retrieved.Description != "Lose weight" {
		t.Errorf("Expected 'Lose weight', got %s", retrieved.Description)
	}
}

func TestMockDB_GetLatestGoal_NoGoal(t *testing.T) {
	db := NewMockDB()

	goal, err := db.GetLatestGoal()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if goal != nil {
		t.Error("Expected nil goal")
	}
}

func TestMockDB_RemoveLastEntry(t *testing.T) {
	db := NewMockDB()
	db.AddFoodEntry(models.FoodEntry{Description: "Entry 1"})
	db.AddFoodEntry(models.FoodEntry{Description: "Entry 2"})

	err := db.RemoveLastEntry()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	entries := db.GetFoodEntries()
	if len(entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(entries))
	}
	if entries[0].Description != "Entry 1" {
		t.Errorf("Expected 'Entry 1', got %s", entries[0].Description)
	}
}

func TestMockDB_RemoveLastEntry_Empty(t *testing.T) {
	db := NewMockDB()

	err := db.RemoveLastEntry()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestMockDB_Close(t *testing.T) {
	db := NewMockDB()

	err := db.Close()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestMockDB_GetFoodEntries(t *testing.T) {
	db := NewMockDB()
	db.AddFoodEntry(models.FoodEntry{Description: "Entry 1"})
	db.AddFoodEntry(models.FoodEntry{Description: "Entry 2"})

	entries := db.GetFoodEntries()
	if len(entries) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(entries))
	}
}

func TestMockDB_GetWaterEntries(t *testing.T) {
	db := NewMockDB()
	db.AddWaterEntry(models.WaterEntry{AmountML: 250})
	db.AddWaterEntry(models.WaterEntry{AmountML: 500})

	entries := db.GetWaterEntries()
	if len(entries) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(entries))
	}
}

func TestMockDB_Clear(t *testing.T) {
	db := NewMockDB()
	db.AddFoodEntry(models.FoodEntry{Description: "Entry 1"})
	db.AddWaterEntry(models.WaterEntry{AmountML: 250})
	db.SetGoal(models.Goal{Description: "Test"})

	db.Clear()

	if len(db.GetFoodEntries()) != 0 {
		t.Error("Expected food entries to be cleared")
	}
	if len(db.GetWaterEntries()) != 0 {
		t.Error("Expected water entries to be cleared")
	}
	goal, _ := db.GetLatestGoal()
	if goal != nil {
		t.Error("Expected goal to be cleared")
	}
}

// testError is a simple error type for testing
type testError string

func (e testError) Error() string {
	return string(e)
}
