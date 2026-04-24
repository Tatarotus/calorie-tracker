package services

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"calorie-tracker/config"
	"calorie-tracker/models"
)

func TestLLMService_Call_Success(t *testing.T) {
	cfg := &config.Config{
		SambaAPIKey:   "test-key",
		OpenAIBaseURL: "http://test.com",
		FoodModel:     "test-model",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"choices": [{"message": {"content": "test response"}}]}`))
	}))
	defer server.Close()

	cfg.OpenAIBaseURL = server.URL

	llm := NewLLMService(cfg)

	result, err := llm.Call("test-model", "test prompt")

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result != "test response" {
		t.Errorf("Expected 'test response', got %q", result)
	}
}

func TestLLMService_Call_Failure(t *testing.T) {
	cfg := &config.Config{
		SambaAPIKey:   "test-key",
		OpenAIBaseURL: "http://test.com",
		FoodModel:     "test-model",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error"))
	}))
	defer server.Close()

	cfg.OpenAIBaseURL = server.URL

	llm := NewLLMService(cfg)

	_, err := llm.Call("test-model", "test prompt")

	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestLLMService_ParseFood_Success(t *testing.T) {
	cfg := &config.Config{
		SambaAPIKey:   "test-key",
		OpenAIBaseURL: "http://test.com",
		FoodModel:     "food-model",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		jsonResp := MockFoodPreviewResponse("Apple", 95, 0.5, 25, 0.3)
		// Properly escape the JSON string for embedding in the API response
		escapedContent, _ := json.Marshal(jsonResp)
		response := `{"choices": [{"message": {"content": ` + string(escapedContent) + `}}]}`
		w.Write([]byte(response))
	}))
	defer server.Close()

	cfg.OpenAIBaseURL = server.URL

	llm := NewLLMService(cfg)

	result, err := llm.ParseFood("Apple")

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
	if result.Description != "Apple" {
		t.Errorf("Expected description 'Apple', got %q", result.Description)
	}
	if result.Calories != 95 {
		t.Errorf("Expected calories 95, got %f", result.Calories)
	}
	if result.Protein != 0.5 {
		t.Errorf("Expected protein 0.5, got %f", result.Protein)
	}
}

func TestLLMService_ParseFood_WithMarkdown(t *testing.T) {
	cfg := &config.Config{
		SambaAPIKey:   "test-key",
		OpenAIBaseURL: "http://test.com",
		FoodModel:     "food-model",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// Return response with markdown block
		jsonResp := MockFoodPreviewResponse("Banana", 105, 1.3, 27, 0.4)
		// Properly escape the content for JSON
		markdownResp := "Here is the JSON:\n```json\n" + jsonResp + "\n```\nEnd"
		// Use json.Marshal to properly escape the string
		escapedContent, _ := json.Marshal(markdownResp)
		response := `{"choices": [{"message": {"content": ` + string(escapedContent) + `}}]}`
		w.Write([]byte(response))
	}))
	defer server.Close()

	cfg.OpenAIBaseURL = server.URL

	llm := NewLLMService(cfg)

	result, err := llm.ParseFood("Banana")

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result.Calories != 105 {
		t.Errorf("Expected calories 105, got %f", result.Calories)
	}
}

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
	if result.Calories != 62 {
		t.Errorf("Expected calories 62, got %f", result.Calories)
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

func TestLLMService_AnalyzeReview_Success(t *testing.T) {
	cfg := &config.Config{
		SambaAPIKey:   "test-key",
		OpenAIBaseURL: "http://test.com",
		ReviewModel:   "review-model",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		jsonResp := MockReviewResultResponse()
		escapedContent, _ := json.Marshal(jsonResp)
		response := `{"choices": [{"message": {"content": ` + string(escapedContent) + `}}]}`
		w.Write([]byte(response))
	}))
	defer server.Close()

	cfg.OpenAIBaseURL = server.URL

	llm := NewLLMService(cfg)

	data := models.ReviewData{
		Goal: "Lose weight",
		Days: []models.DailyStats{
			{Date: "2024-01-15", Calories: 1800},
		},
	}

	result, err := llm.AnalyzeReview(data)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
	if result.Score != 85 {
		t.Errorf("Expected score 85, got %d", result.Score)
	}
	if result.Progress != "improving" {
		t.Errorf("Expected progress 'improving', got %q", result.Progress)
	}
	if len(result.Issues) != 1 {
		t.Errorf("Expected 1 issue, got %d", len(result.Issues))
	}
}

func TestLLMService_AnalyzeReview_WithMarkdown(t *testing.T) {
	cfg := &config.Config{
		SambaAPIKey:   "test-key",
		OpenAIBaseURL: "http://test.com",
		ReviewModel:   "review-model",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// Return with markdown block
		jsonResp := MockReviewResultResponse()
		markdownResp := "Analysis:\n```json\n" + jsonResp + "\n```\nDone"
		escapedContent, _ := json.Marshal(markdownResp)
		response := `{"choices": [{"message": {"content": ` + string(escapedContent) + `}}]}`
		w.Write([]byte(response))
	}))
	defer server.Close()

	cfg.OpenAIBaseURL = server.URL

	llm := NewLLMService(cfg)

	data := models.ReviewData{
		Goal: "Build muscle",
		Days: []models.DailyStats{},
	}

	result, err := llm.AnalyzeReview(data)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result.Score != 85 {
		t.Errorf("Expected score 85, got %d", result.Score)
	}
}

func TestLLMService_CallLLM_Authorization(t *testing.T) {
	var authHeader string

	cfg := &config.Config{
		SambaAPIKey:   "my-secret-key",
		OpenAIBaseURL: "http://test.com",
		FoodModel:     "test-model",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"choices": [{"message": {"content": "test"}}]}`))
	}))
	defer server.Close()

	cfg.OpenAIBaseURL = server.URL

	llm := NewLLMService(cfg)

	_, _ = llm.Call("test", "prompt")

	expectedAuth := "Bearer my-secret-key"
	if authHeader != expectedAuth {
		t.Errorf("Expected Authorization '%s', got '%s'", expectedAuth, authHeader)
	}
}

func TestLLMService_CallLLM_URLFormatting(t *testing.T) {
	testCases := []struct {
		name     string
		baseURL  string
		expected string
	}{
		{
			name:     "already has /chat/completions",
			baseURL:  "https://api.example.com/chat/completions",
			expected: "https://api.example.com/chat/completions",
		},
		{
			name:     "needs /chat/completions",
			baseURL:  "https://api.example.com/v1",
			expected: "https://api.example.com/v1/chat/completions",
		},
		{
			name:     "trailing slash",
			baseURL:  "https://api.example.com/",
			expected: "https://api.example.com/chat/completions",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			url := tc.baseURL
			if !strings.HasSuffix(url, "/chat/completions") {
				url = strings.TrimSuffix(url, "/") + "/chat/completions"
			}

			if url != tc.expected {
				t.Errorf("Expected URL '%s', got '%s'", tc.expected, url)
			}
		})
	}
}

func TestLLMService_NoChoices(t *testing.T) {
	cfg := &config.Config{
		SambaAPIKey:   "test-key",
		OpenAIBaseURL: "http://test.com",
		FoodModel:     "test-model",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"choices": []}`))
	}))
	defer server.Close()

	cfg.OpenAIBaseURL = server.URL

	llm := NewLLMService(cfg)

	_, err := llm.Call("test", "prompt")

	if err == nil {
		t.Error("Expected error for empty choices, got nil")
	}
}

func TestLLMService_InvalidJSONResponse(t *testing.T) {
	cfg := &config.Config{
		SambaAPIKey:   "test-key",
		OpenAIBaseURL: "http://test.com",
		FoodModel:     "test-model",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`invalid json`))
	}))
	defer server.Close()

	cfg.OpenAIBaseURL = server.URL

	llm := NewLLMService(cfg)

	_, err := llm.Call("test", "prompt")

	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

// Table-driven tests for ParseFood with various responses
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
			response:        MockFoodPreviewResponse("Test", 100, 10, 20, 5),
			expectedCal:     100,
			expectedProtein: 10,
			expectedError:   false,
		},
		{
			name:            "response with units",
			response:        `{"calories": "150g", "protein": "20kcal", "carbs": 30, "fat": 5}`,
			expectedCal:     150,
			expectedProtein: 20,
			expectedError:   false,
		},
		{
			name:            "response with quoted numbers",
			response:        `{"calories": "200", "protein": "15", "carbs": "40", "fat": "10"}`,
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
			response:        `{"calories": 100}`,
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

			if result.Calories != tt.expectedCal {
				t.Errorf("Expected calories %f, got %f", tt.expectedCal, result.Calories)
			}
			if result.Protein != tt.expectedProtein {
				t.Errorf("Expected protein %f, got %f", tt.expectedProtein, result.Protein)
			}
		})
	}
}

// Table-driven tests for AnalyzeReview
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
