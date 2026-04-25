package db

import (
	"calorie-tracker/models"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

// Test helpers
func setupTestDB(t *testing.T) *DB {
	t.Helper()
	db, err := NewTestDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test DB: %v", err)
	}
	return db
}

func cleanupTestDB(t *testing.T, db *DB) {
	t.Helper()
	if err := db.Close(); err != nil {
		t.Errorf("Failed to close test DB: %v", err)
	}
}

// Test table creation and migration
func TestSQLite_Migration(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	// Verify tables exist by trying to insert
	entry := models.FoodEntry{
		Timestamp:   time.Now(),
		Description: "Test",
		Calories:    100,
		Protein:     5,
		Carbs:       10,
		Fat:         2,
	}

	err := db.AddFoodEntry(entry)
	if err != nil {
		t.Errorf("Migration failed - cannot insert: %v", err)
	}
}

// Test food entry operations
func TestSQLite_AddAndGetFoodEntry(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	entry := models.FoodEntry{
		Timestamp:   time.Now(),
		Description: "Apple",
		Calories:    95,
		Protein:     0.5,
		Carbs:       25,
		Fat:         0.3,
	}

	err := db.AddFoodEntry(entry)
	if err != nil {
		t.Fatalf("Failed to add food entry: %v", err)
	}

	// Retrieve entries for today
	entries, err := db.GetDailyFoodEntries(entry.Timestamp)
	if err != nil {
		t.Fatalf("Failed to get daily entries: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(entries))
	}

	retrieved := entries[0]
	if retrieved.Description != "Apple" {
		t.Errorf("Expected 'Apple', got '%s'", retrieved.Description)
	}
	if retrieved.Calories != 95 {
		t.Errorf("Expected 95 calories, got %f", retrieved.Calories)
	}
}

// Test multiple food entries
func TestSQLite_MultipleFoodEntries(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	now := time.Now()
	entries := []models.FoodEntry{
		{Timestamp: now, Description: "Breakfast", Calories: 300},
		{Timestamp: now, Description: "Lunch", Calories: 500},
		{Timestamp: now, Description: "Dinner", Calories: 600},
	}

	for _, entry := range entries {
		if err := db.AddFoodEntry(entry); err != nil {
			t.Fatalf("Failed to add entry: %v", err)
		}
	}

	retrieved, err := db.GetDailyFoodEntries(now)
	if err != nil {
		t.Fatalf("Failed to retrieve entries: %v", err)
	}

	if len(retrieved) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(retrieved))
	}
}

// Test food entries across different dates
func TestSQLite_FoodEntriesDifferentDates(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	today := time.Now()
	yesterday := today.AddDate(0, 0, -1)

	// Add entries for today
	todayEntry := models.FoodEntry{
		Timestamp:   today,
		Description: "Today's Food",
		Calories:    200,
	}
	db.AddFoodEntry(todayEntry)

	// Add entries for yesterday
	yesterdayEntry := models.FoodEntry{
		Timestamp:   yesterday,
		Description: "Yesterday's Food",
		Calories:    300,
	}
	db.AddFoodEntry(yesterdayEntry)

	// Get today's entries
	todayEntries, err := db.GetDailyFoodEntries(today)
	if err != nil {
		t.Fatalf("Failed to get today's entries: %v", err)
	}
	if len(todayEntries) != 1 {
		t.Errorf("Expected 1 today entry, got %d", len(todayEntries))
	}

	// Get yesterday's entries
	yesterdayEntries, err := db.GetDailyFoodEntries(yesterday)
	if err != nil {
		t.Fatalf("Failed to get yesterday's entries: %v", err)
	}
	if len(yesterdayEntries) != 1 {
		t.Errorf("Expected 1 yesterday entry, got %d", len(yesterdayEntries))
	}
}

// Test food cache operations
func TestSQLite_CacheFood(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	entry := models.FoodEntry{
		Description: "Cached Apple",
		Calories:    100,
		Protein:     1,
		Carbs:       20,
		Fat:         1,
	}

	err := db.CacheFood(entry)
	if err != nil {
		t.Fatalf("Failed to cache food: %v", err)
	}

	cached, err := db.GetCachedFood("Cached Apple")
	if err != nil {
		t.Fatalf("Failed to get cached food: %v", err)
	}

	if cached == nil {
		t.Fatal("Expected cached entry, got nil")
	}

	if cached.Calories != 100 {
		t.Errorf("Expected 100 calories, got %f", cached.Calories)
	}
}

// Test cache with special characters
func TestSQLite_CacheFood_SpecialCharacters(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	entry := models.FoodEntry{
		Description: "Café au lait",
		Calories:    150,
	}

	err := db.CacheFood(entry)
	if err != nil {
		t.Fatalf("Failed to cache food with special chars: %v", err)
	}

	cached, err := db.GetCachedFood("Café au lait")
	if err != nil {
		t.Fatalf("Failed to get cached food: %v", err)
	}

	if cached == nil {
		t.Fatal("Expected cached entry")
	}
}

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

func TestSQLite_GetFoodEntriesRange(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	now := time.Now()
	for i := 0; i < 5; i++ {
		date := now.AddDate(0, 0, -i)
		db.AddFoodEntry(models.FoodEntry{
			Timestamp:   date,
			Description: "Day " + string(rune('0'+i)),
			Calories:    100 * float64(i+1),
		})
	}

	entries, err := db.GetFoodEntriesRange(7)
	if err != nil {
		t.Fatalf("Failed to get entries range: %v", err)
	}

	if len(entries) != 5 {
		t.Errorf("Expected 5 entries, got %d", len(entries))
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

func TestSQLite_CacheFood_Duplicate(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	entry1 := models.FoodEntry{
		Description: "Apple",
		Calories:    100,
	}
	db.CacheFood(entry1)

	entry2 := models.FoodEntry{
		Description: "Apple",
		Calories:    150,
	}
	err := db.CacheFood(entry2)
	if err != nil {
		t.Fatalf("Failed to cache duplicate: %v", err)
	}

	cached, _ := db.GetCachedFood("Apple")
	if cached.Calories != 150 {
		t.Errorf("Expected 150 calories (updated), got %f", cached.Calories)
	}
}

func TestSQLite_EmptyDescription(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	entry := models.FoodEntry{
		Timestamp: time.Now(),
		Calories:  100,
	}

	err := db.AddFoodEntry(entry)
	if err != nil {
		t.Fatalf("Failed to add entry with empty description: %v", err)
	}

	entries, err := db.GetDailyFoodEntries(entry.Timestamp)
	if err != nil {
		t.Fatalf("Failed to get entries: %v", err)
	}

	if len(entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(entries))
	}
}

func TestSQLite_ZeroValues(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	entry := models.FoodEntry{
		Timestamp:   time.Now(),
		Description: "Zero entry",
		Calories:    0,
		Protein:     0,
		Carbs:       0,
		Fat:         0,
	}

	err := db.AddFoodEntry(entry)
	if err != nil {
		t.Fatalf("Failed to add zero-value entry: %v", err)
	}

	entries, err := db.GetDailyFoodEntries(entry.Timestamp)
	if err != nil {
		t.Fatalf("Failed to get entries: %v", err)
	}

	if len(entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(entries))
	}

	if entries[0].Calories != 0 {
		t.Errorf("Expected 0 calories, got %f", entries[0].Calories)
	}
}

func TestSQLite_LargeValues(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	entry := models.FoodEntry{
		Timestamp:   time.Now(),
		Description: "Large entry",
		Calories:    10000,
		Protein:     500,
		Carbs:       1000,
		Fat:         200,
	}

	err := db.AddFoodEntry(entry)
	if err != nil {
		t.Fatalf("Failed to add large-value entry: %v", err)
	}

	entries, err := db.GetDailyFoodEntries(entry.Timestamp)
	if err != nil {
		t.Fatalf("Failed to get entries: %v", err)
	}

	if entries[0].Calories != 10000 {
		t.Errorf("Expected 10000 calories, got %f", entries[0].Calories)
	}
}
