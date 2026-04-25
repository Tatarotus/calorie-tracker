package services

import (
	"errors"
	"fmt"
	"strings"
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
		"progress": "Consistent tracking",
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

	// No goal set - will return "No goal set"
	mockDB.SetGoal(models.Goal{
		Timestamp:   time.Now(),
		Description: "No specific goal",
	})

	reviewJSON := `{
		"summary": "Tracking without goal",
		"goal_progress": "No goal defined",
		"progress": "Consistent logging",
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

func TestTrackerService_RunReview_EmptyData(t *testing.T) {
	mockDB := db.NewMockDB()

	// No entries, no goal
	reviewJSON := `{
		"summary": "Just started",
		"goal_progress": "No data yet",
		"progress": "Account setup complete",
		"score": 2,
		"issues": ["No entries"],
		"suggestions": ["Log your first meal"],
		"patterns": []
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

	if result.Score != 2 {
		t.Errorf("Expected score 2, got %d", result.Score)
	}
}

func TestTrackerService_RunReview_DBError(t *testing.T) {
	mockDB := db.NewMockDB()

	// Setup error for GetStatsRange
	mockDB.SetError("GetStatsRange", errors.New("db error"))

	server := MockHTTPServer("{}")
	defer server.Close()

	cfg := &config.Config{
		SambaAPIKey:   "test-key",
		OpenAIBaseURL: server.URL,
	}
	llm := NewLLMServiceWithClient(cfg, server.Client())
	tracker := NewTrackerService(mockDB, llm)

	_, err := tracker.RunReview()
	if err == nil {
		t.Error("Expected error from DB, got nil")
	}
	if !strings.Contains(err.Error(), "db error") {
		t.Errorf("Expected 'db error' in message, got '%s'", err.Error())
	}
}

func TestTrackerService_RunReview_LLMError(t *testing.T) {
	mockDB := db.NewMockDB()

	// Setup data
	now := time.Now()
	mockDB.AddFoodEntry(models.FoodEntry{
		Timestamp:   now,
		Description: "Test food",
		Calories:    300,
	})

	// Setup LLM error server
	errorServer := MockHTTPServerError(500)
	defer errorServer.Close()

	cfg := &config.Config{
		SambaAPIKey:   "test-key",
		OpenAIBaseURL: errorServer.URL,
	}
	llm := NewLLMServiceWithClient(cfg, errorServer.Client())
	tracker := NewTrackerService(mockDB, llm)

	_, err := tracker.RunReview()
	if err == nil {
		t.Error("Expected LLM error, got nil")
	}
}

func TestTrackerService_RunReview_MissingDays(t *testing.T) {
	mockDB := db.NewMockDB()

	// Only add entries for 3 days
	now := time.Now()
	for i := 0; i < 3; i++ {
		date := now.AddDate(0, 0, -i)
		mockDB.AddFoodEntry(models.FoodEntry{
			Timestamp:   date,
			Description: fmt.Sprintf("Day %d", i),
			Calories:    400,
		})
	}

	reviewJSON := `{
		"summary": "Inconsistent tracking",
		"goal_progress": "3/7 days tracked",
		"progress": "Some progress made",
		"score": 4,
		"issues": ["Missing days"],
		"suggestions": ["Track every day"],
		"patterns": ["Weekend gaps"]
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

	if result.Score != 4 {
		t.Errorf("Expected score 4, got %d", result.Score)
	}
}

func TestTrackerService_RunReview_WithWaterOnly(t *testing.T) {
	mockDB := db.NewMockDB()

	// Only water entries, no food
	now := time.Now()
	for i := 0; i < 5; i++ {
		mockDB.AddWaterEntry(models.WaterEntry{
			Timestamp: now,
			AmountML:  500,
		})
	}

	reviewJSON := `{
		"summary": "Great hydration",
		"goal_progress": "Water goal met",
		"progress": "Consistent water intake",
		"score": 6,
		"issues": ["No food logged"],
		"suggestions": ["Log your meals"],
		"patterns": ["Good hydration habits"]
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

	if result.Score != 6 {
		t.Errorf("Expected score 6, got %d", result.Score)
	}
}

func TestTrackerService_RunReview_BoundaryScores(t *testing.T) {
	mockDB := db.NewMockDB()

	now := time.Now()
	mockDB.AddFoodEntry(models.FoodEntry{
		Timestamp: now,
		Calories:  2000,
	})

	// Test boundary scores
	testCases := []struct {
		score   int
		summary string
	}{
		{1, "Very poor"},
		{5, "Average"},
		{10, "Excellent"},
	}

	for _, tc := range testCases {
		reviewJSON := fmt.Sprintf(`{
			"summary": "%s",
			"goal_progress": "Score %d",
			"progress": "Test progress",
			"score": %d,
			"issues": [],
			"suggestions": [],
			"patterns": []
		}`, tc.summary, tc.score, tc.score)

		server := MockHTTPServer(reviewJSON)
		cfg := &config.Config{
			SambaAPIKey:   "test-key",
			OpenAIBaseURL: server.URL,
		}
		llm := NewLLMServiceWithClient(cfg, server.Client())
		tracker := NewTrackerService(mockDB, llm)
		result, err := tracker.RunReview()
		server.Close()

		if err != nil {
			t.Errorf("RunReview failed for score %d: %v", tc.score, err)
			continue
		}

		if result.Score != tc.score {
			t.Errorf("Expected score %d, got %d", tc.score, result.Score)
		}
	}
}
