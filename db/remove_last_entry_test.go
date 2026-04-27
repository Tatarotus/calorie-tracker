package db

import (
	"calorie-tracker/models"
	"testing"
	"time"
)

// TestMockDB_RemoveLastEntryEdgeCases tests edge cases for RemoveLastEntry function
func TestMockDB_RemoveLastEntryEdgeCases(t *testing.T) {
	db := NewMockDB()

	// Test the boundary condition: len(m.foodEntries) > 0
	// This specifically tests the > 0 condition that could be mutated to >= 0

	// Test with empty food entries (boundary condition)
	err := db.RemoveLastEntry()
	if err != nil {
		t.Errorf("Expected no error for empty entries, got %v", err)
	}

	// Test with exactly one entry (boundary condition)
	entry := models.FoodEntry{
		Timestamp:   time.Now(),
		Description: "Test entry",
		Calories:    100,
	}

	err = db.AddFoodEntry(entry)
	if err != nil {
		t.Errorf("Failed to add entry: %v", err)
	}

	// Now test removing when there's one entry
	err = db.RemoveLastEntry()
	if err != nil {
		t.Errorf("Expected no error for one entry, got %v", err)
	}

	// Test removing when there are no more entries (boundary condition)
	err = db.RemoveLastEntry()
	if err != nil {
		t.Errorf("Expected no error for removing from empty, got %v", err)
	}
}
