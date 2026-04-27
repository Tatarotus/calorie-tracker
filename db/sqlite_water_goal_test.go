package db

import (
	"calorie-tracker/models"
	"testing"
	"time"
)

// Test water entry operations
func TestSQLite_AddAndGetWaterEntry(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	entry := models.WaterEntry{
		Timestamp: time.Now(),
		AmountML:  250,
	}

	err := db.AddWaterEntry(entry)
	if err != nil {
		t.Fatalf("Failed to add water entry: %v", err)
	}

	entries, err := db.GetDailyWaterEntries(entry.Timestamp)
	if err != nil {
		t.Fatalf("Failed to get daily water entries: %v", err)
	}

	if len(entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(entries))
	}

	if entries[0].AmountML != 250 {
		t.Errorf("Expected 250ml, got %f", entries[0].AmountML)
	}
}

// Test multiple water entries
func TestSQLite_MultipleWaterEntries(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	now := time.Now()
	amounts := []float64{250, 500, 300}

	for _, amount := range amounts {
		entry := models.WaterEntry{
			Timestamp: now,
			AmountML:  amount,
		}
		if err := db.AddWaterEntry(entry); err != nil {
			t.Fatalf("Failed to add water entry: %v", err)
		}
	}

	entries, err := db.GetDailyWaterEntries(now)
	if err != nil {
		t.Fatalf("Failed to get water entries: %v", err)
	}

	if len(entries) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(entries))
	}
}

// Test stats calculation
func TestSQLite_GetStatsRange(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	now := time.Now()

	// Add food entries
	db.AddFoodEntry(models.FoodEntry{
		Timestamp: now,
		Calories:  100,
		Protein:   5,
		Carbs:     20,
		Fat:       2,
	})

	db.AddFoodEntry(models.FoodEntry{
		Timestamp: now,
		Calories:  200,
		Protein:   10,
		Carbs:     30,
		Fat:       5,
	})

	// Add water entries
	db.AddWaterEntry(models.WaterEntry{
		Timestamp: now,
		AmountML:  250,
	})

	db.AddWaterEntry(models.WaterEntry{
		Timestamp: now,
		AmountML:  500,
	})

	stats, err := db.GetStatsRange(7)
	if err != nil {
		t.Fatalf("Failed to get stats: %v", err)
	}

	if len(stats) == 0 {
		t.Fatal("Expected at least one day of stats")
	}

	dayStats := stats[0]
	expectedCalories := 300.0
	expectedWater := 750.0

	if dayStats.Calories != expectedCalories {
		t.Errorf("Expected %.0f calories, got %.0f", expectedCalories, dayStats.Calories)
	}
	if dayStats.WaterML != expectedWater {
		t.Errorf("Expected %.0f water, got %.0f", expectedWater, dayStats.WaterML)
	}
}

// Test goal operations
func TestSQLite_SetAndGetGoal(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	goal := models.Goal{
		Timestamp:   time.Now(),
		Description: "Lose 5kg in 3 months",
	}

	err := db.SetGoal(goal)
	if err != nil {
		t.Fatalf("Failed to set goal: %v", err)
	}

	retrieved, err := db.GetLatestGoal()
	if err != nil {
		t.Fatalf("Failed to get goal: %v", err)
	}

	if retrieved == nil {
		t.Fatal("Expected goal, got nil")
	}

	if retrieved.Description != "Lose 5kg in 3 months" {
		t.Errorf("Expected 'Lose 5kg in 3 months', got '%s'", retrieved.Description)
	}
}

// Test goal update
func TestSQLite_UpdateGoal(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	// Set initial goal
	goal1 := models.Goal{
		Timestamp:   time.Now(),
		Description: "Initial goal",
	}
	db.SetGoal(goal1)

	// Update goal
	goal2 := models.Goal{
		Timestamp:   time.Now(),
		Description: "Updated goal",
	}
	db.SetGoal(goal2)

	retrieved, _ := db.GetLatestGoal()
	if retrieved.Description != "Updated goal" {
		t.Errorf("Expected 'Updated goal', got '%s'", retrieved.Description)
	}
}

// Test no goal exists
func TestSQLite_NoGoal(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	goal, err := db.GetLatestGoal()
	if err != nil {
		t.Fatalf("Failed to get goal: %v", err)
	}

	if goal != nil {
		t.Error("Expected nil goal, got non-nil")
	}
}

// Test RemoveLastEntry

func TestSQLite_RemoveLastEntry(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	now := time.Now()
	// Use different timestamps to ensure deterministic removal
	db.AddFoodEntry(models.FoodEntry{
		Timestamp:   now.Add(-2 * time.Second),
		Description: "Entry 1",
		Calories:    100,
	})
	db.AddFoodEntry(models.FoodEntry{
		Timestamp:   now.Add(-1 * time.Second),
		Description: "Entry 2",
		Calories:    200,
	})

	err := db.RemoveLastEntry()
	if err != nil {
		t.Fatalf("Failed to remove last entry: %v", err)
	}

	entries, _ := db.GetFoodEntriesRange(7)
	if len(entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(entries))
	}

	// Entry 1 should remain (Entry 2 was the last by timestamp)
	if len(entries) > 0 && entries[0].Description != "Entry 1" {
		t.Errorf("Expected Entry 1 to remain, got %s", entries[0].Description)
	}
}

func TestSQLite_RemoveLastEntry_Empty(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	err := db.RemoveLastEntry()
	if err != nil {
		t.Errorf("Expected no error on empty remove, got %v", err)
	}
}

func TestSQLite_GetWaterEntriesRange(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	now := time.Now()
	for i := 0; i < 3; i++ {
		db.AddWaterEntry(models.WaterEntry{
			Timestamp: now,
			AmountML:  250 * float64(i+1),
		})
	}

	entries, err := db.GetWaterEntriesRange(7)
	if err != nil {
		t.Fatalf("Failed to get water entries range: %v", err)
	}

	if len(entries) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(entries))
	}
}
