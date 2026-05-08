package db

import (
	"testing"
	"time"

	"calorie-tracker/models"
)

func TestDB_AddFoodEntry(t *testing.T) {
	db, err := NewTestDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test DB: %v", err)
	}
	defer func() { _ = db.Close() }()

	entry := models.FoodEntry{
		Timestamp:   time.Now(),
		Description: "apple",
		Calories:    95,
		Protein:     0.5,
		Carbs:       25,
		Fat:         0.3,
	}

	err = db.AddFoodEntry(entry)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	entries, err := db.GetDailyFoodEntries(time.Now())
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(entries))
	}
}

func TestDB_AddWaterEntry(t *testing.T) {
	db, err := NewTestDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test DB: %v", err)
	}
	defer func() { _ = db.Close() }()

	entry := models.WaterEntry{
		Timestamp: time.Now(),
		AmountML:  500,
	}

	err = db.AddWaterEntry(entry)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	entries, err := db.GetDailyWaterEntries(time.Now())
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(entries))
	}
}

func TestDB_GetFoodEntriesRange(t *testing.T) {
	db, err := NewTestDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test DB: %v", err)
	}
	defer func() { _ = db.Close() }()

	now := time.Now()
	_ = db.AddFoodEntry(models.FoodEntry{Timestamp: now, Description: "today", Calories: 100})
	_ = db.AddFoodEntry(models.FoodEntry{Timestamp: now.AddDate(0, 0, -1), Description: "yesterday", Calories: 200})
	_ = db.AddFoodEntry(models.FoodEntry{Timestamp: now.AddDate(0, 0, -8), Description: "last week", Calories: 300})

	entries, err := db.GetFoodEntriesRange(7)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("Expected 2 entries (last 7 days), got %d", len(entries))
	}
}

func TestDB_GetWaterEntriesRange(t *testing.T) {
	db, err := NewTestDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test DB: %v", err)
	}
	defer func() { _ = db.Close() }()

	now := time.Now()
	_ = db.AddWaterEntry(models.WaterEntry{Timestamp: now, AmountML: 500})
	_ = db.AddWaterEntry(models.WaterEntry{Timestamp: now.AddDate(0, 0, -1), AmountML: 300})
	_ = db.AddWaterEntry(models.WaterEntry{Timestamp: now.AddDate(0, 0, -8), AmountML: 200})

	entries, err := db.GetWaterEntriesRange(7)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("Expected 2 entries (last 7 days), got %d", len(entries))
	}
}

func TestDB_GetStatsRange(t *testing.T) {
	db, err := NewTestDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test DB: %v", err)
	}
	defer func() { _ = db.Close() }()

	now := time.Now()
	_ = db.AddFoodEntry(models.FoodEntry{Timestamp: now, Description: "today", Calories: 100, Protein: 10, Carbs: 20, Fat: 5})
	_ = db.AddWaterEntry(models.WaterEntry{Timestamp: now, AmountML: 500})

	stats, err := db.GetStatsRange(7)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(stats) != 1 {
		t.Errorf("Expected 1 stat entry, got %d", len(stats))
	}
	if stats[0].Calories != 100 {
		t.Errorf("Expected 100 calories, got %f", stats[0].Calories)
	}
	if stats[0].WaterML != 500 {
		t.Errorf("Expected 500ml water, got %f", stats[0].WaterML)
	}
}

func TestDB_CacheFood(t *testing.T) {
	db, err := NewTestDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test DB: %v", err)
	}
	defer func() { _ = db.Close() }()

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

	cached, err := db.GetCachedFood("apple")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if cached == nil {
		t.Fatal("Expected cached food, got nil")
	}
	if cached.Macros.Calories != 52 {
		t.Errorf("Expected 52 calories, got %f", cached.Macros.Calories)
	}
}

func TestDB_GetAllCacheEntries(t *testing.T) {
	db, err := NewTestDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test DB: %v", err)
	}
	defer func() { _ = db.Close() }()

	_ = db.CacheFood(models.ReferenceFood{Name: "apple", BaseQuantity: 100, Unit: "gram", Macros: models.Macros{Calories: 52}})
	_ = db.CacheFood(models.ReferenceFood{Name: "banana", BaseQuantity: 100, Unit: "gram", Macros: models.Macros{Calories: 89}})

	entries, err := db.GetAllCacheEntries()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(entries))
	}
}

func TestDB_GetReferenceFood(t *testing.T) {
	db, err := NewTestDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test DB: %v", err)
	}
	defer func() { _ = db.Close() }()

	// Reference foods are seeded on DB creation
	ref, err := db.GetReferenceFood("arroz branco")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if ref == nil {
		t.Fatal("Expected reference food, got nil")
	}
	if ref.Macros.Calories != 130 {
		t.Errorf("Expected 130 calories, got %f", ref.Macros.Calories)
	}
}

func TestDB_SetGoal(t *testing.T) {
	db, err := NewTestDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test DB: %v", err)
	}
	defer func() { _ = db.Close() }()

	goal := models.Goal{
		Timestamp:   time.Now(),
		Description: "Lose weight",
	}

	err = db.SetGoal(goal)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	latest, err := db.GetLatestGoal()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if latest == nil {
		t.Fatal("Expected goal, got nil")
	}
	if latest.Description != "Lose weight" {
		t.Errorf("Expected 'Lose weight', got %s", latest.Description)
	}
}

func TestDB_RemoveLastEntry(t *testing.T) {
	db, err := NewTestDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test DB: %v", err)
	}
	defer func() { _ = db.Close() }()

	now := time.Now()
	_ = db.AddFoodEntry(models.FoodEntry{Timestamp: now, Description: "apple", Calories: 95})

	err = db.RemoveLastEntry()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	entries, _ := db.GetDailyFoodEntries(now)
	if len(entries) != 0 {
		t.Errorf("Expected 0 entries after removal, got %d", len(entries))
	}
}

func TestDB_GetCachedFood_NotFound(t *testing.T) {
	db, err := NewTestDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test DB: %v", err)
	}
	defer func() { _ = db.Close() }()

	cached, err := db.GetCachedFood("nonexistent")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if cached != nil {
		t.Error("Expected nil for non-existent food")
	}
}

func TestDB_GetLatestGoal_NoGoal(t *testing.T) {
	db, err := NewTestDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test DB: %v", err)
	}
	defer func() { _ = db.Close() }()

	goal, err := db.GetLatestGoal()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if goal != nil {
		t.Error("Expected nil goal when no goals exist")
	}
}
