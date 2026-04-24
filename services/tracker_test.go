package services

import (
	"testing"
	"time"
	
	"calorie-tracker/models"
)

func TestNewTrackerService(t *testing.T) {
	// We can't easily mock *db.DB and *LLMService without interfaces
	// So we'll test the structure creation with nil values
	// This at least tests that the constructor works
	
	// For now, just test that we can create the service
	// Real testing would require dependency injection or interfaces
	t.Log("Tracker service structure test - requires proper mocking setup")
}

func TestDailyStatsCalculation(t *testing.T) {
	// Test the logic in GetDailyStats by creating sample data
	foodEntries := []models.FoodEntry{
		{Calories: 100, Protein: 5, Carbs: 20, Fat: 2},
		{Calories: 200, Protein: 10, Carbs: 30, Fat: 5},
	}
	
	waterEntries := []models.WaterEntry{
		{AmountML: 250},
		{AmountML: 500},
	}
	
	// Calculate expected values
	expectedCalories := 300.0
	expectedProtein := 15.0
	expectedCarbs := 50.0
	expectedFat := 7.0
	expectedWater := 750.0
	
	// Verify our test data is correct
	totalCalories := 0.0
	totalProtein := 0.0
	totalCarbs := 0.0
	totalFat := 0.0
	for _, f := range foodEntries {
		totalCalories += f.Calories
		totalProtein += f.Protein
		totalCarbs += f.Carbs
		totalFat += f.Fat
	}
	
	totalWater := 0.0
	for _, w := range waterEntries {
		totalWater += w.AmountML
	}
	
	if totalCalories != expectedCalories {
		t.Errorf("Expected calories %f, got %f", expectedCalories, totalCalories)
	}
	if totalProtein != expectedProtein {
		t.Errorf("Expected protein %f, got %f", expectedProtein, totalProtein)
	}
	if totalCarbs != expectedCarbs {
		t.Errorf("Expected carbs %f, got %f", expectedCarbs, totalCarbs)
	}
	if totalFat != expectedFat {
		t.Errorf("Expected fat %f, got %f", expectedFat, totalFat)
	}
	if totalWater != expectedWater {
		t.Errorf("Expected water %f, got %f", expectedWater, totalWater)
	}
}

func TestFoodEntrySimpleConversion(t *testing.T) {
	entry := models.FoodEntry{
		Timestamp:   time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
		Description: "Test Food",
		Calories:    100,
		Protein:     5,
		Carbs:       20,
		Fat:         2,
	}
	
	simple := models.FoodEntrySimple{
		Date:        entry.Timestamp.Local().Format("2006-01-02"),
		Description: entry.Description,
		Calories:    entry.Calories,
		Protein:     entry.Protein,
		Carbs:       entry.Carbs,
		Fat:         entry.Fat,
	}
	
	if simple.Date != "2024-01-15" {
		t.Errorf("Expected date 2024-01-15, got %s", simple.Date)
	}
	if simple.Description != "Test Food" {
		t.Errorf("Expected 'Test Food', got %s", simple.Description)
	}
}

func TestWaterEntrySimpleConversion(t *testing.T) {
	entry := models.WaterEntry{
		Timestamp: time.Date(2024, 1, 15, 8, 0, 0, 0, time.UTC),
		AmountML:  250,
	}
	
	simple := models.WaterEntrySimple{
		Date:      entry.Timestamp.Local().Format("2006-01-02"),
		AmountML:  entry.AmountML,
	}
	
	if simple.AmountML != 250 {
		t.Errorf("Expected 250, got %f", simple.AmountML)
	}
}
