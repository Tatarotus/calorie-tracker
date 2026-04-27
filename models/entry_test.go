package models

import (
	"testing"
	"time"
)

func TestFoodEntryCreation(t *testing.T) {
	entry := FoodEntry{
		ID:          1,
		Timestamp:   time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
		Description: "Apple",
		Calories:    95,
		Protein:     0.5,
		Carbs:       25,
		Fat:         0.3,
	}

	if entry.ID != 1 {
		t.Errorf("Expected ID 1, got %d", entry.ID)
	}
	if entry.Description != "Apple" {
		t.Errorf("Expected Description 'Apple', got %s", entry.Description)
	}
	if entry.Calories != 95 {
		t.Errorf("Expected Calories 95, got %f", entry.Calories)
	}
}

func TestWaterEntryCreation(t *testing.T) {
	entry := WaterEntry{
		ID:        1,
		Timestamp: time.Date(2024, 1, 15, 8, 0, 0, 0, time.UTC),
		AmountML:  250,
	}

	if entry.AmountML != 250 {
		t.Errorf("Expected AmountML 250, got %f", entry.AmountML)
	}
}

func TestDailyStatsAggregation(t *testing.T) {
	stats := DailyStats{
		Date:     "2024-01-15",
		Calories: 2000,
		Protein:  150,
		Carbs:    200,
		Fat:      70,
		WaterML:  2000,
	}

	if stats.Calories != 2000 {
		t.Errorf("Expected Calories 2000, got %f", stats.Calories)
	}
	if stats.WaterML != 2000 {
		t.Errorf("Expected WaterML 2000, got %f", stats.WaterML)
	}
}

func TestFoodPreviewCreation(t *testing.T) {
	preview := FoodPreview{
		Description: "Banana",
		Calories:    105,
		Protein:     1.3,
		Carbs:       27,
		Fat:         0.4,
	}

	if preview.Calories != 105 {
		t.Errorf("Expected Calories 105, got %f", preview.Calories)
	}
}

func TestReviewDataCreation(t *testing.T) {
	data := ReviewData{
		Goal: "Lose weight",
		Days: []DailyStats{
			{Date: "2024-01-15", Calories: 1800},
			{Date: "2024-01-14", Calories: 2000},
		},
		FoodEntries: []FoodEntrySimple{
			{Date: "2024-01-15", Description: "Breakfast", Calories: 400},
		},
		WaterEntries: []WaterEntrySimple{
			{Date: "2024-01-15", AmountML: 1500},
		},
	}

	if len(data.Days) != 2 {
		t.Errorf("Expected 2 days, got %d", len(data.Days))
	}
	if len(data.FoodEntries) != 1 {
		t.Errorf("Expected 1 food entry, got %d", len(data.FoodEntries))
	}
}

func TestReviewResultCreation(t *testing.T) {
	result := ReviewResult{
		Summary:      "Good progress",
		GoalProgress: "On track",
		Progress:     "75%",
		Score:        8,
		Issues:       []string{"Low protein on Monday"},
		Suggestions:  []string{"Add more chicken"},
		Patterns:     []string{"Consistent breakfast"},
	}

	if result.Score != 8 {
		t.Errorf("Expected Score 8, got %d", result.Score)
	}
	if len(result.Issues) != 1 {
		t.Errorf("Expected 1 issue, got %d", len(result.Issues))
	}
}
