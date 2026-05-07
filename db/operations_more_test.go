package db

import (
	"testing"
	"time"

	"calorie-tracker/models"
)

func TestDB_GetReferenceFood_PartialMatch(t *testing.T) {
	db, err := NewTestDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test DB: %v", err)
	}
	defer db.Close()

	// Seed a reference food - the LIKE query uses the input as the pattern
	// and matches against the stored name. So "arroz branco" should match
	// when searching for a substring of it.
	ref, err := db.GetReferenceFood("arroz branco")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if ref == nil {
		t.Error("Expected match for 'arroz branco'")
	}
}

func TestDB_GetReferenceFood_NotFound(t *testing.T) {
	db, err := NewTestDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test DB: %v", err)
	}
	defer db.Close()

	ref, err := db.GetReferenceFood("nonexistentfood123")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if ref != nil {
		t.Error("Expected nil for non-existent food")
	}
}

func TestDB_GetDailyFoodEntries_MultipleDays(t *testing.T) {
	db, err := NewTestDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test DB: %v", err)
	}
	defer db.Close()

	now := time.Now()
	db.AddFoodEntry(models.FoodEntry{Timestamp: now, Description: "today", Calories: 100})
	db.AddFoodEntry(models.FoodEntry{Timestamp: now.AddDate(0, 0, -1), Description: "yesterday", Calories: 200})

	entries, err := db.GetDailyFoodEntries(now)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("Expected 1 entry for today, got %d", len(entries))
	}
	if entries[0].Description != "today" {
		t.Errorf("Expected 'today', got %s", entries[0].Description)
	}
}

func TestDB_GetDailyWaterEntries(t *testing.T) {
	db, err := NewTestDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test DB: %v", err)
	}
	defer db.Close()

	now := time.Now()
	db.AddWaterEntry(models.WaterEntry{Timestamp: now, AmountML: 500})
	db.AddWaterEntry(models.WaterEntry{Timestamp: now.AddDate(0, 0, -1), AmountML: 300})

	entries, err := db.GetDailyWaterEntries(now)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("Expected 1 entry for today, got %d", len(entries))
	}
	if entries[0].AmountML != 500 {
		t.Errorf("Expected 500ml, got %f", entries[0].AmountML)
	}
}

func TestDB_GetStatsRange_MultipleDays(t *testing.T) {
	db, err := NewTestDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test DB: %v", err)
	}
	defer db.Close()

	now := time.Now()
	db.AddFoodEntry(models.FoodEntry{Timestamp: now, Description: "today", Calories: 100, Protein: 10, Carbs: 20, Fat: 5})
	db.AddFoodEntry(models.FoodEntry{Timestamp: now.AddDate(0, 0, -1), Description: "yesterday", Calories: 200, Protein: 20, Carbs: 30, Fat: 10})
	db.AddWaterEntry(models.WaterEntry{Timestamp: now, AmountML: 500})
	db.AddWaterEntry(models.WaterEntry{Timestamp: now.AddDate(0, 0, -1), AmountML: 300})

	stats, err := db.GetStatsRange(7)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(stats) != 2 {
		t.Errorf("Expected 2 stat entries, got %d", len(stats))
	}
}

func TestDB_CacheFood_Update(t *testing.T) {
	db, err := NewTestDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test DB: %v", err)
	}
	defer db.Close()

	food := models.ReferenceFood{
		Name:         "apple",
		BaseQuantity: 100,
		Unit:         "gram",
		Macros:       models.Macros{Calories: 52, Protein: 0.3, Carbs: 14, Fat: 0.2},
	}

	err = db.CacheFood(food)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Update the same food
	food.Macros.Calories = 55
	err = db.CacheFood(food)
	if err != nil {
		t.Errorf("Expected no error on update, got %v", err)
	}

	cached, err := db.GetCachedFood("apple")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if cached.Macros.Calories != 55 {
		t.Errorf("Expected updated calories 55, got %f", cached.Macros.Calories)
	}
}

func TestDB_RemoveLastEntry_WaterOnly(t *testing.T) {
	db, err := NewTestDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test DB: %v", err)
	}
	defer db.Close()

	now := time.Now()
	db.AddWaterEntry(models.WaterEntry{Timestamp: now, AmountML: 500})

	err = db.RemoveLastEntry()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	entries, _ := db.GetDailyWaterEntries(now)
	if len(entries) != 0 {
		t.Errorf("Expected 0 entries after removal, got %d", len(entries))
	}
}

func TestDB_RemoveLastEntry_Empty(t *testing.T) {
	db, err := NewTestDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test DB: %v", err)
	}
	defer db.Close()

	err = db.RemoveLastEntry()
	if err != nil {
		t.Errorf("Expected no error on empty DB, got %v", err)
	}
}

func TestDB_GetAllCacheEntries_Empty(t *testing.T) {
	db, err := NewTestDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test DB: %v", err)
	}
	defer db.Close()

	entries, err := db.GetAllCacheEntries()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("Expected 0 entries, got %d", len(entries))
	}
}
