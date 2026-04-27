package services

import (
	"calorie-tracker/config"
	"calorie-tracker/db"
	"calorie-tracker/models"
	"testing"
)

func TestTrackerService_ParseFood_Hybrid(t *testing.T) {
	mockDB := db.NewMockDB()
	
	// Seed reference data
	mockDB.SeedReferenceFood(models.ReferenceFood{
		Name: "arroz branco",
		BaseQuantity: 100,
		Unit: "gram",
		Calories: 130,
		Protein: 2.7,
		Carbs: 28,
		Fat: 0.3,
	})

	tracker := NewTrackerService(mockDB, nil)

	t.Run("Reference match with scaling", func(t *testing.T) {
		preview, err := tracker.ParseFood("200g arroz branco")
		if err != nil {
			t.Fatalf("ParseFood failed: %v", err)
		}

		// 200g is 2x base quantity (100g)
		expectedCal := 130.0 * 2
		if preview.Calories != expectedCal {
			t.Errorf("Expected %f calories, got %f", expectedCal, preview.Calories)
		}
		if preview.Protein != 2.7*2 {
			t.Errorf("Expected %f protein, got %f", 2.7*2, preview.Protein)
		}
	})

	t.Run("Reference match without quantity (use base)", func(t *testing.T) {
		preview, err := tracker.ParseFood("arroz branco")
		if err != nil {
			t.Fatalf("ParseFood failed: %v", err)
		}

		if preview.Calories != 130.0 {
			t.Errorf("Expected 130 calories, got %f", preview.Calories)
		}
	})

	t.Run("Reference match with fuzzy name", func(t *testing.T) {
		// MockDB implementation of GetReferenceFood uses strings.Contains
		preview, err := tracker.ParseFood("100g de arroz branco cozido")
		if err != nil {
			t.Fatalf("ParseFood failed: %v", err)
		}

		if preview.Calories != 130.0 {
			t.Errorf("Expected 130 calories, got %f", preview.Calories)
		}
	})

	t.Run("LLM Fallback for unknown food", func(t *testing.T) {
		server := MockHTTPServer(`{"calories": 50, "protein": 0, "carbs": 12, "fat": 0}`)
		defer server.Close()

		cfg := &config.Config{
			SambaAPIKey:   "test",
			OpenAIBaseURL: server.URL,
			FoodModel:     "test",
		}
		llm := NewLLMServiceWithClient(cfg, server.Client())
		tracker := NewTrackerService(mockDB, llm)

		preview, err := tracker.ParseFood("pêra")
		if err != nil {
			t.Fatalf("ParseFood failed: %v", err)
		}

		if preview.Calories != 50 {
			t.Errorf("Expected 50 calories from LLM, got %f", preview.Calories)
		}
	})

	t.Run("Reject unrealistic LLM values", func(t *testing.T) {
		server := MockHTTPServer(`{"calories": 10000, "protein": 0, "carbs": 0, "fat": 0}`)
		defer server.Close()

		cfg := &config.Config{
			SambaAPIKey:   "test",
			OpenAIBaseURL: server.URL,
			FoodModel:     "test",
		}
		llm := NewLLMServiceWithClient(cfg, server.Client())
		tracker := NewTrackerService(mockDB, llm)

		_, err := tracker.ParseFood("magic beans")
		if err == nil {
			t.Error("Expected error for unrealistic calorie value, got nil")
		}
	})
}
