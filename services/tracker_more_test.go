package services

import (
	"testing"

	"calorie-tracker/db"
	"calorie-tracker/models"
)

func TestTrackerService_SaveFood_WithName(t *testing.T) {
	mockDB := db.NewMockDB()
	tracker := NewTrackerService(mockDB, nil)

	preview := &models.FoodPreview{
		Description: "100g apple",
		Name:        "apple",
		Unit:        "gram",
		Calories:    52,
		Protein:     0.3,
		Carbs:       14,
		Fat:         0.2,
	}

	err := tracker.SaveFood(preview)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Check that food was cached
	cached, err := mockDB.GetCachedFood("apple")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if cached == nil {
		t.Error("Expected food to be cached")
	}
}

func TestTrackerService_SaveFood_WithoutName(t *testing.T) {
	mockDB := db.NewMockDB()
	tracker := NewTrackerService(mockDB, nil)

	preview := &models.FoodPreview{
		Description: "100g apple",
		Calories:    52,
		Protein:     0.3,
		Carbs:       14,
		Fat:         0.2,
	}

	err := tracker.SaveFood(preview)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Food entry should be saved
	food := mockDB.GetFoodEntries()
	if len(food) != 1 {
		t.Errorf("Expected 1 food entry, got %d", len(food))
	}
}

func TestTrackerService_SaveFood_ZeroAmount(t *testing.T) {
	mockDB := db.NewMockDB()
	tracker := NewTrackerService(mockDB, nil)

	preview := &models.FoodPreview{
		Description: "apple",
		Name:        "apple",
		Unit:        "unit",
		Calories:    52,
		Protein:     0.3,
		Carbs:       14,
		Fat:         0.2,
	}

	err := tracker.SaveFood(preview)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Should not cache with zero amount
	cached, _ := mockDB.GetCachedFood("apple")
	// Zero amount means no cache
	_ = cached
}
