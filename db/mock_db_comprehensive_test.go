package db

import (
	"testing"

	"calorie-tracker/models"
)

func TestMockDB_CanonicalFoodOperations(t *testing.T) {
	m := NewMockDB()

	food := &models.CanonicalFood{ID: 1, CanonicalName: "Test"}

	// Test Save
	err := m.SaveCanonicalFood(food)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Test Get by ID
	fetched, _ := m.GetCanonicalFood(1)
	if fetched == nil || fetched.CanonicalName != "Test" {
		t.Error("Failed to get canonical food from mock")
	}

	// Test Get by Name
	fetchedByName, _ := m.GetCanonicalFoodByName("Test")
	if fetchedByName == nil || fetchedByName.ID != 1 {
		t.Error("Failed to get canonical food by name from mock")
	}

	// Test Get All
	all, _ := m.GetAllCanonicalFoods()
	if len(all) != 1 {
		t.Errorf("Expected 1 food, got %d", len(all))
	}
}

func TestMockDB_NutritionCacheOperations(t *testing.T) {
	m := NewMockDB()

	entry := &models.NutritionCacheEntry{ID: 1, CanonicalFoodID: 10, ServingUnit: "gram", Calories: 100}

	// Test Save
	err := m.SaveNutritionCache(entry)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Test Get
	fetched, _ := m.GetNutritionCache(10, "gram")
	if fetched == nil || fetched.Calories != 100 {
		t.Error("Failed to get nutrition cache from mock")
	}
}

func TestMockDB_UserOverrideOperations(t *testing.T) {
	m := NewMockDB()

	override := &models.UserOverrideEntry{ID: 1, CanonicalFoodID: 10, Calories: 500}

	// Test Save
	err := m.SaveUserOverride(override)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Test Get
	fetched, _ := m.GetUserOverride(10)
	if fetched == nil || fetched.Calories != 500 {
		t.Error("Failed to get user override from mock")
	}
}

func TestMockDB_HelperMethods(t *testing.T) {
	m := NewMockDB()

	_ = m.AddFoodEntry(models.FoodEntry{Description: "Apple"})
	_ = m.AddWaterEntry(models.WaterEntry{AmountML: 250})

	if len(m.GetFoodEntries()) != 1 {
		t.Error("Expected 1 food entry")
	}
	if len(m.GetWaterEntries()) != 1 {
		t.Error("Expected 1 water entry")
	}

	m.Clear()
	if len(m.GetFoodEntries()) != 0 {
		t.Error("Expected 0 food entries after clear")
	}
}
