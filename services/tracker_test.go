package services

import (
	"fmt"
	"testing"
	"time"

	"calorie-tracker/config"
	"calorie-tracker/db"
	"calorie-tracker/models"
)

// RunReview tests
func TestTrackerService_RunReview_Success(t *testing.T) {
	mockDB := db.NewMockDB()

	// Setup mock DB data
	now := time.Now()
	for i := 0; i < 7; i++ {
		date := now.AddDate(0, 0, -i)
		mockDB.AddFoodEntry(models.FoodEntry{
			Timestamp:   date,
			Description: fmt.Sprintf("Day %d food", i),
			Calories:    500 * float64(i+1),
		})
		mockDB.AddWaterEntry(models.WaterEntry{
			Timestamp: date,
			AmountML:  250 * float64(i+1),
		})
	}

	// Setup goal
	mockDB.SetGoal(models.Goal{
		Timestamp:   now,
		Description: "Lose 5kg in 3 months",
	})

	// Setup LLM mock HTTP server
	reviewJSON := `{
		"summary": "Good progress overall",
		"goal_progress": "60% of weekly goal achieved",
		"progress": "improving",
		"score": 7,
		"issues": ["Evening snacks"],
		"suggestions": ["Add more protein"],
		"patterns": ["Higher calories on weekends"]
	}`
	server := MockHTTPServer(reviewJSON)
	defer server.Close()

	cfg := &config.Config{
		SambaAPIKey:   "test-key",
		OpenAIBaseURL: server.URL,
	}
	llm := NewLLMServiceWithClient(cfg, server.Client())
	tracker := NewTrackerService(mockDB, llm)

	result, err := tracker.RunReview()
	if err != nil {
		t.Fatalf("RunReview failed: %v", err)
	}

	if result.Score != 7 {
		t.Errorf("Expected score 7, got %d", result.Score)
	}
	if result.Summary != "Good progress overall" {
		t.Errorf("Expected summary, got '%s'", result.Summary)
	}
	if len(result.Issues) != 1 {
		t.Errorf("Expected 1 issue, got %d", len(result.Issues))
	}
}

func TestTrackerService_RunReview_NoGoal(t *testing.T) {
	mockDB := db.NewMockDB()

	// No goal set - GetGoal returns "No goal set"
	reviewJSON := `{
		"summary": "Tracking without goal",
		"goal_progress": "No goal defined",
		"progress": "stable",
		"score": 5,
		"issues": ["No target"],
		"suggestions": ["Set a specific goal"],
		"patterns": ["Regular tracking"]
	}`
	server := MockHTTPServer(reviewJSON)
	defer server.Close()

	cfg := &config.Config{
		SambaAPIKey:   "test-key",
		OpenAIBaseURL: server.URL,
	}
	llm := NewLLMServiceWithClient(cfg, server.Client())
	tracker := NewTrackerService(mockDB, llm)

	result, err := tracker.RunReview()
	if err != nil {
		t.Fatalf("RunReview failed: %v", err)
	}

	if result.Score != 5 {
		t.Errorf("Expected score 5, got %d", result.Score)
	}
}

func TestTrackerService_SaveFood_WithCaching(t *testing.T) {
	mockDB := db.NewMockDB()
	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := NewLLMService(cfg)
	svc := NewTrackerService(mockDB, llm)

	preview := &models.FoodPreview{
		Description: "Super Food",
		Calories:    500,
	}

	err := svc.SaveFood(preview)
	if err != nil {
		t.Fatalf("SaveFood failed: %v", err)
	}

	// Verify it's in the main log
	entries := mockDB.GetFoodEntries()
	found := false
	for _, e := range entries {
		if e.Description == "Super Food" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Food entry not found in database")
	}

	// Verify it's in the cache
	cached, err := mockDB.GetCachedFood("super food") // it normalizes to lowercase
	if err != nil {
		t.Fatalf("GetCachedFood failed: %v", err)
	}
	if cached == nil {
		t.Fatal("Food entry not found in cache")
	}
	if cached.Calories != 500 {
		t.Errorf("Expected 500 calories in cache, got %f", cached.Calories)
	}
}

func TestTrackerService_ParseFood_CacheFirst(t *testing.T) {
	mockDB := db.NewMockDB()
	// Add to cache
	mockDB.CacheFood(models.FoodEntry{
		Description: "apple",
		Calories:    95,
	})

	// Setup LLM that would fail if called
	server := MockHTTPServerError(500)
	defer server.Close()

	cfg := &config.Config{
		SambaAPIKey:   "test-key",
		OpenAIBaseURL: server.URL,
	}
	llm := NewLLMServiceWithClient(cfg, server.Client())
	tracker := NewTrackerService(mockDB, llm)

	// Should match "apple" in cache and NOT call LLM
	preview, err := tracker.ParseFood("apple")
	if err != nil {
		t.Fatalf("ParseFood failed: %v", err)
	}
	if preview.Calories != 95 {
		t.Errorf("Expected 95 calories from cache, got %f", preview.Calories)
	}
}

func TestTrackerService_ParseFood_LLMFallback(t *testing.T) {
	mockDB := db.NewMockDB()

	// Mock LLM success
	server := MockHTTPServer(`{"calories": 100, "protein": 1, "carbs": 25, "fat": 0}`)
	defer server.Close()

	cfg := &config.Config{
		SambaAPIKey:   "test-key",
		OpenAIBaseURL: server.URL,
		FoodModel:     "test",
	}
	llm := NewLLMServiceWithClient(cfg, server.Client())
	tracker := NewTrackerService(mockDB, llm)

	preview, err := tracker.ParseFood("banana")
	if err != nil {
		t.Fatalf("ParseFood failed: %v", err)
	}
	if preview.Calories != 100 {
		t.Errorf("Expected 100 calories from LLM, got %f", preview.Calories)
	}
}
