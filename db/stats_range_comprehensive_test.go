package db

import (
	"calorie-tracker/models"
	"testing"
	"time"
)

func TestGetStatsRangeComprehensive(t *testing.T) {
	mockDB := NewMockDB()

	now := time.Now()
	// Add entries for different days
	mockDB.AddFoodEntry(models.FoodEntry{
		Timestamp: now,
		Calories:  1000,
	})
	mockDB.AddFoodEntry(models.FoodEntry{
		Timestamp: now.AddDate(0, 0, -5),
		Calories:  500,
	})
	mockDB.AddWaterEntry(models.WaterEntry{
		Timestamp: now.AddDate(0, 0, -2),
		AmountML:  1000,
	})

	testCases := []struct {
		name          string
		days          int
		expectedCount int
	}{
		{"Today only", 0, 1},
		{"Last 2 days", 2, 2}, // Today and 2 days ago (actually 2 days would be today, yesterday, 2 days ago)
		// Wait, MockDB.GetStatsRange implementation:
		// rangeStartUTC := todayStart.AddDate(0, 0, -days).UTC()
		// WHERE timestamp >= ?
		{"Last 7 days", 7, 3},
		{"Large range", 100, 3},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			stats, err := mockDB.GetStatsRange(tc.days)
			if err != nil {
				t.Fatalf("GetStatsRange failed: %v", err)
			}
			// Note: stats count depends on how many unique dates have entries within the range
			if len(stats) != tc.expectedCount {
				t.Errorf("Expected %d days with stats, got %d", tc.expectedCount, len(stats))
			}
		})
	}
}
