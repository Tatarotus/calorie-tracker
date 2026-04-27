package db

import (
	"calorie-tracker/models"
	"testing"
	"time"
)

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

	ref := models.ReferenceFood{
		Name:         "Cached Apple",
		BaseQuantity: 100,
		Unit:         "gram",
		Macros: models.Macros{
			Calories: 100,
			Protein:  1,
			Carbs:    20,
			Fat:      1,
		},
	}

	err := db.CacheFood(ref)
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

	if cached.Macros.Calories != 100 {
		t.Errorf("Expected 100 calories, got %f", cached.Macros.Calories)
	}
}

// Test cache with special characters
func TestSQLite_CacheFood_SpecialCharacters(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	ref := models.ReferenceFood{
		Name:         "Café au lait",
		BaseQuantity: 200,
		Unit:         "ml",
		Macros: models.Macros{
			Calories: 150,
		},
	}

	err := db.CacheFood(ref)
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

func TestSQLite_CacheFood_Duplicate(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	ref1 := models.ReferenceFood{
		Name:         "Apple",
		BaseQuantity: 100,
		Unit:         "gram",
		Macros: models.Macros{
			Calories: 100,
		},
	}
	db.CacheFood(ref1)

	ref2 := models.ReferenceFood{
		Name:         "Apple",
		BaseQuantity: 100,
		Unit:         "gram",
		Macros: models.Macros{
			Calories: 150,
		},
	}
	err := db.CacheFood(ref2)
	if err != nil {
		t.Fatalf("Failed to cache duplicate: %v", err)
	}

	cached, _ := db.GetCachedFood("Apple")
	if cached.Macros.Calories != 150 {
		t.Errorf("Expected 150 calories (updated), got %f", cached.Macros.Calories)
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
