package db

import (
	"testing"

	"calorie-tracker/models"
)

func TestDB_CanonicalFoodOperations(t *testing.T) {
	db, err := NewTestDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test DB: %v", err)
	}
	defer func() { _ = db.Close() }()

	food := &models.CanonicalFood{
		CanonicalName:  "test food",
		NormalizedName: "test food",
		Language:       "en",
		Category:       "Test",
	}

	// Test Save
	err = db.SaveCanonicalFood(food)
	if err != nil {
		t.Fatalf("Failed to save canonical food: %v", err)
	}
	if food.ID == 0 {
		t.Error("Expected food ID to be set after save")
	}

	// Test Get by ID
	fetched, err := db.GetCanonicalFood(food.ID)
	if err != nil {
		t.Fatalf("Failed to get canonical food by ID: %v", err)
	}
	if fetched == nil {
		t.Fatal("Expected fetched food, got nil")
	}
	if fetched.CanonicalName != food.CanonicalName {
		t.Errorf("Expected name %s, got %s", food.CanonicalName, fetched.CanonicalName)
	}

	// Test Get by Name
	fetchedByName, err := db.GetCanonicalFoodByName("test food")
	if err != nil {
		t.Fatalf("Failed to get canonical food by name: %v", err)
	}
	if fetchedByName == nil {
		t.Fatal("Expected fetched food by name, got nil")
	}
	if fetchedByName.ID != food.ID {
		t.Errorf("Expected ID %d, got %d", food.ID, fetchedByName.ID)
	}

	// Test Update
	food.Category = "Updated Category"
	err = db.SaveCanonicalFood(food)
	if err != nil {
		t.Fatalf("Failed to update canonical food: %v", err)
	}
	fetchedUpdated, _ := db.GetCanonicalFood(food.ID)
	if fetchedUpdated.Category != "Updated Category" {
		t.Errorf("Expected category 'Updated Category', got %s", fetchedUpdated.Category)
	}

	// Test Get All
	all, err := db.GetAllCanonicalFoods()
	if err != nil {
		t.Fatalf("Failed to get all canonical foods: %v", err)
	}
	found := false
	for _, f := range all {
		if f.ID == food.ID {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected to find the saved food in GetAllCanonicalFoods")
	}
}

func TestDB_NutritionCacheOperations(t *testing.T) {
	db, err := NewTestDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test DB: %v", err)
	}
	defer func() { _ = db.Close() }()

	// Need a canonical food first
	food := &models.CanonicalFood{CanonicalName: "Cache Test"}
	_ = db.SaveCanonicalFood(food)

	entry := &models.NutritionCacheEntry{
		CanonicalFoodID: food.ID,
		ServingAmount:   100,
		ServingUnit:     "gram",
		Calories:        100,
		Protein:         10,
		Carbs:           20,
		Fat:             5,
		SourceType:      "test",
	}

	// Test Save
	err = db.SaveNutritionCache(entry)
	if err != nil {
		t.Fatalf("Failed to save nutrition cache: %v", err)
	}
	if entry.ID == 0 {
		t.Error("Expected entry ID to be set after save")
	}

	// Test Get
	fetched, err := db.GetNutritionCache(food.ID, "gram")
	if err != nil {
		t.Fatalf("Failed to get nutrition cache: %v", err)
	}
	if fetched == nil {
		t.Fatal("Expected fetched entry, got nil")
	}
	if fetched.Calories != 100 {
		t.Errorf("Expected 100 calories, got %f", fetched.Calories)
	}

	// Test Get Fallback (any unit)
	fetchedAny, err := db.GetNutritionCache(food.ID, "ounce")
	if err != nil {
		t.Fatalf("Failed to get nutrition cache with fallback: %v", err)
	}
	if fetchedAny == nil {
		t.Error("Expected fallback to return some entry for the food")
	}

	// Test Update
	entry.Calories = 120
	err = db.SaveNutritionCache(entry)
	if err != nil {
		t.Fatalf("Failed to update nutrition cache: %v", err)
	}
	fetchedUpdated, _ := db.GetNutritionCache(food.ID, "gram")
	if fetchedUpdated.Calories != 120 {
		t.Errorf("Expected updated calories 120, got %f", fetchedUpdated.Calories)
	}
}

func TestDB_UserOverrideOperations(t *testing.T) {
	db, err := NewTestDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test DB: %v", err)
	}
	defer func() { _ = db.Close() }()

	// Need a canonical food first
	food := &models.CanonicalFood{CanonicalName: "Override Test"}
	_ = db.SaveCanonicalFood(food)

	override := &models.UserOverrideEntry{
		CanonicalFoodID: food.ID,
		ServingAmount:   1,
		ServingUnit:     "unit",
		Calories:        500,
		OverrideReason:  "User choice",
	}

	// Test Save
	err = db.SaveUserOverride(override)
	if err != nil {
		t.Fatalf("Failed to save user override: %v", err)
	}
	if override.ID == 0 {
		t.Error("Expected override ID to be set after save")
	}

	// Test Get
	fetched, err := db.GetUserOverride(food.ID)
	if err != nil {
		t.Fatalf("Failed to get user override: %v", err)
	}
	if fetched == nil {
		t.Fatal("Expected fetched override, got nil")
	}
	if fetched.Calories != 500 {
		t.Errorf("Expected 500 calories, got %f", fetched.Calories)
	}

	// Test Update
	override.Calories = 600
	err = db.SaveUserOverride(override)
	if err != nil {
		t.Fatalf("Failed to update user override: %v", err)
	}
	fetchedUpdated, _ := db.GetUserOverride(food.ID)
	if fetchedUpdated.Calories != 600 {
		t.Errorf("Expected updated calories 600, got %f", fetchedUpdated.Calories)
	}
}

func TestDB_ErrorPaths(t *testing.T) {
	db, err := NewTestDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test DB: %v", err)
	}
	_ = db.Close() // Close immediately to trigger errors

	// Test errors on closed DB
	_, err = db.GetCanonicalFood(1)
	if err == nil {
		t.Error("Expected error on closed DB for GetCanonicalFood")
	}

	_, err = db.GetCanonicalFoodByName("test")
	if err == nil {
		t.Error("Expected error on closed DB for GetCanonicalFoodByName")
	}

	err = db.SaveCanonicalFood(&models.CanonicalFood{CanonicalName: "test"})
	if err == nil {
		t.Error("Expected error on closed DB for SaveCanonicalFood")
	}

	_, err = db.GetAllCanonicalFoods()
	if err == nil {
		t.Error("Expected error on closed DB for GetAllCanonicalFoods")
	}

	_, err = db.GetNutritionCache(1, "gram")
	if err == nil {
		t.Error("Expected error on closed DB for GetNutritionCache")
	}

	err = db.SaveNutritionCache(&models.NutritionCacheEntry{CanonicalFoodID: 1})
	if err == nil {
		t.Error("Expected error on closed DB for SaveNutritionCache")
	}

	_, err = db.GetUserOverride(1)
	if err == nil {
		t.Error("Expected error on closed DB for GetUserOverride")
	}

	err = db.SaveUserOverride(&models.UserOverrideEntry{CanonicalFoodID: 1})
	if err == nil {
		t.Error("Expected error on closed DB for SaveUserOverride")
	}
}
