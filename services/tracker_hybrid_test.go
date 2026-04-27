package services

import (
	"calorie-tracker/config"
	"calorie-tracker/db"
	"calorie-tracker/models"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNutritionEngine_HybridFlow(t *testing.T) {
	mockDB := db.NewMockDB()
	
	// Seed reference data
	mockDB.SeedReferenceFood(models.ReferenceFood{
		Name: "arroz branco",
		BaseQuantity: 100,
		Unit: "gram",
		Macros: models.Macros{
			Calories: 130,
			Protein: 2.7,
		},
	})

	t.Run("Priority 1: Reference DB match with scaling", func(t *testing.T) {
		engine := NewNutritionEngine(mockDB, nil)
		preview, err := engine.Analyze("200g arroz branco")
		if err != nil {
			t.Fatalf("Analyze failed: %v", err)
		}

		if preview.Calories != 260 {
			t.Errorf("Expected 260 calories (scaled), got %f", preview.Calories)
		}
	})

	t.Run("Priority 2: Cache match (Previous LLM result)", func(t *testing.T) {
		// Mock cache entry
		mockDB.CacheFood(models.ReferenceFood{
			Name: "pao integral",
			BaseQuantity: 50,
			Unit: "gram",
			Macros: models.Macros{
				Calories: 120,
			},
		})

		engine := NewNutritionEngine(mockDB, nil)
		preview, err := engine.Analyze("100g pao integral")
		if err != nil {
			t.Fatalf("Analyze failed: %v", err)
		}

		// 100g is 2x base of 50g
		if preview.Calories != 240 {
			t.Errorf("Expected 240 calories (scaled from cache), got %f", preview.Calories)
		}
	})

	t.Run("Priority 3: LLM Fallback and caching", func(t *testing.T) {
		// Mock LLM server
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, `{"choices":[{"message":{"content":"{\"name\":\"fruta magica\",\"base_quantity\":100,\"unit\":\"g\",\"macros\":{\"calories\":50,\"protein\":1,\"carbs\":10,\"fat\":0}}"}}]}`)
		}))
		defer ts.Close()

		cfg := &config.Config{
			SambaAPIKey: "test",
			OpenAIBaseURL: ts.URL,
			FoodModel: "test",
		}
		llm := NewLLMServiceWithClient(cfg, ts.Client())
		engine := NewNutritionEngine(mockDB, llm)

		preview, err := engine.Analyze("200g fruta magica")
		if err != nil {
			t.Fatalf("Analyze failed: %v", err)
		}

		if preview.Calories != 100 {
			t.Errorf("Expected 100 calories (scaled from LLM base), got %f", preview.Calories)
		}

		// Verify it was cached as base ReferenceFood
		cached, _ := mockDB.GetCachedFood("fruta magica")
		if cached == nil {
			t.Fatal("Expected LLM result to be cached")
		}
		if cached.Macros.Calories != 50 {
			t.Errorf("Expected cached base calories to be 50, got %f", cached.Macros.Calories)
		}
	})

	t.Run("Reject unrealistic LLM data", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, `{"choices":[{"message":{"content":"{\"name\":\"bad\",\"base_quantity\":100,\"unit\":\"g\",\"macros\":{\"calories\":10000,\"protein\":0,\"carbs\":0,\"fat\":0}}"}}]}`)
		}))
		defer ts.Close()

		cfg := &config.Config{OpenAIBaseURL: ts.URL}
		llm := NewLLMServiceWithClient(cfg, ts.Client())
		engine := NewNutritionEngine(mockDB, llm)

		_, err := engine.Analyze("something")
		if err == nil {
			t.Error("Expected error for unrealistic calories")
		}
	})
}
