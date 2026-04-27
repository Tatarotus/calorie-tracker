package services

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"calorie-tracker/config"
	"calorie-tracker/models"
)

func TestLLMService_ParseFood_Retry(t *testing.T) {
	callCount := 0

	cfg := &config.Config{
		SambaAPIKey:   "test-key",
		OpenAIBaseURL: "http://test.com",
		FoodModel:     "food-model",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		jsonResp := MockFoodPreviewResponse("Orange", 62, 1.2, 15, 0.2)
		escapedContent, _ := json.Marshal(jsonResp)
		response := `{"choices": [{"message": {"content": ` + string(escapedContent) + `}}]}`
		w.Write([]byte(response))
	}))
	defer server.Close()

	cfg.OpenAIBaseURL = server.URL

	llm := NewLLMService(cfg)

	result, err := llm.ParseFood("Orange")

	if err != nil {
		t.Errorf("Expected no error after retries, got %v", err)
	}
	if callCount != 3 {
		t.Errorf("Expected 3 calls (2 failures + 1 success), got %d", callCount)
	}
	if result.Macros.Calories != 62 {
		t.Errorf("Expected calories 62, got %f", result.Macros.Calories)
	}
}

func TestLLMService_ParseFood_PermanentFailure(t *testing.T) {
	cfg := &config.Config{
		SambaAPIKey:   "test-key",
		OpenAIBaseURL: "http://test.com",
		FoodModel:     "food-model",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	cfg.OpenAIBaseURL = server.URL

	llm := NewLLMService(cfg)

	_, err := llm.ParseFood("Test")

	if err == nil {
		t.Error("Expected error after all retries, got nil")
	}
}

func TestLLMService_ParseFood_TableDriven(t *testing.T) {
	tests := []struct {
		name            string
		response        string
		expectedCal     float64
		expectedProtein float64
		expectedError   bool
	}{
		{
			name:            "valid response",
			response:        `{"name": "Apple", "base_quantity": 100, "unit": "g", "macros": {"calories": 95, "protein": 0.5, "carbs": 25, "fat": 0.3}}`,
			expectedCal:     95,
			expectedProtein: 0.5,
			expectedError:   false,
		},
		{
			name:            "response with units",
			response:        `{"name": "test", "base_quantity": 100, "unit": "g", "macros": {"calories": "150 kcal", "protein": "20g", "carbs": "5g", "fat": "5g"}}`,
			expectedCal:     150,
			expectedProtein: 20,
			expectedError:   false,
		},
		{
			name:            "response with quoted numbers",
			response:        `{"name": "test", "base_quantity": 100, "unit": "g", "macros": {"calories": 200, "protein": 15, "carbs": 40, "fat": 10}}`,
			expectedCal:     200,
			expectedProtein: 15,
			expectedError:   false,
		},
		{
			name:            "invalid JSON",
			response:        `not json`,
			expectedCal:     0,
			expectedProtein: 0,
			expectedError:   true,
		},
		{
			name:            "missing fields",
			response:        `{"name": "test", "base_quantity": 100, "unit": "g", "macros": {"calories": 100}}`,
			expectedCal:     100,
			expectedProtein: 0,
			expectedError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				SambaAPIKey:   "test-key",
				OpenAIBaseURL: "http://test.com",
				FoodModel:     "food-model",
			}

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				escapedContent, _ := json.Marshal(tt.response)
				response := `{"choices": [{"message": {"content": ` + string(escapedContent) + `}}]}`
				w.Write([]byte(response))
			}))
			defer server.Close()

			cfg.OpenAIBaseURL = server.URL

			llm := NewLLMService(cfg)

			result, err := llm.ParseFood("Test Food")

			if tt.expectedError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Expected no error, got %v", err)
				return
			}

			if result.Macros.Calories != tt.expectedCal {
				t.Errorf("Expected calories %f, got %f", tt.expectedCal, result.Macros.Calories)
			}
			if result.Macros.Protein != tt.expectedProtein {
				t.Errorf("Expected protein %f, got %f", tt.expectedProtein, result.Macros.Protein)
			}
		})
	}
}

func TestLLMService_AnalyzeReview_TableDriven(t *testing.T) {
	tests := []struct {
		name          string
		response      string
		expectedScore int
		expectedError bool
	}{
		{
			name:          "valid response",
			response:      MockReviewResultResponse(),
			expectedScore: 85,
			expectedError: false,
		},
		{
			name:          "response with markdown",
			response:      "```json\n" + MockReviewResultResponse() + "\n```",
			expectedScore: 85,
			expectedError: false,
		},
		{
			name:          "invalid JSON",
			response:      "not valid json",
			expectedScore: 0,
			expectedError: true,
		},
		{
			name:          "score at boundaries",
			response:      `{"summary": "test", "goal_progress": "test", "progress": "stable", "score": 0, "issues": [], "suggestions": [], "patterns": []}`,
			expectedScore: 0,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				SambaAPIKey:   "test-key",
				OpenAIBaseURL: "http://test.com",
				ReviewModel:   "review-model",
			}

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				escapedContent, _ := json.Marshal(tt.response)
				response := `{"choices": [{"message": {"content": ` + string(escapedContent) + `}}]}`
				w.Write([]byte(response))
			}))
			defer server.Close()

			cfg.OpenAIBaseURL = server.URL

			llm := NewLLMService(cfg)

			data := models.ReviewData{
				Goal: "Test goal",
				Days: []models.DailyStats{},
			}

			result, err := llm.AnalyzeReview(data)

			if tt.expectedError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Expected no error, got %v", err)
				return
			}

			if result.Score != tt.expectedScore {
				t.Errorf("Expected score %d, got %d", tt.expectedScore, result.Score)
			}
		})
	}
}
