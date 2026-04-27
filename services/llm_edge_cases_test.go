package services

import (
	"calorie-tracker/config"
	"calorie-tracker/models"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLLMService_CallLLM_EmptyChoices(t *testing.T) {
	// Mock server that returns 0 choices
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"choices": []}`))
	}))
	defer server.Close()

	cfg := &config.Config{
		SambaAPIKey:   "test-key",
		OpenAIBaseURL: server.URL,
	}
	llm := NewLLMServiceWithClient(cfg, server.Client())

	_, err := llm.ParseFood("apple")
	if err == nil {
		t.Fatal("Expected error when no choices are returned, got nil")
	}
	if !strings.Contains(err.Error(), "no choices returned") {
		t.Errorf("Expected 'no choices returned' in error, got '%v'", err)
	}
}

func TestLLMService_CallLLM_NonOKStatus(t *testing.T) {
	// Mock server that returns 401 Unauthorized
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "Unauthorized"}`))
	}))
	defer server.Close()

	cfg := &config.Config{
		SambaAPIKey:   "test-key",
		OpenAIBaseURL: server.URL,
	}
	llm := NewLLMServiceWithClient(cfg, server.Client())

	_, err := llm.callLLM("model", "prompt")
	if err == nil {
		t.Fatal("Expected error for non-OK status, got nil")
	}
	if !strings.Contains(err.Error(), "status 401") {
		t.Errorf("Expected status 401 in error, got '%v'", err)
	}
}

func TestLLMService_ParseFood_RetrySuccess(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 2 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		resp := chatResponse{
			Choices: []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			}{
				{Message: struct {
					Content string `json:"content"`
				}{Content: `{"name": "test", "base_quantity": 100, "unit": "g", "macros": {"calories": 100, "protein": 1, "carbs": 25, "fat": 0}}`}},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	cfg := &config.Config{
		SambaAPIKey:   "test-key",
		OpenAIBaseURL: server.URL,
		FoodModel:     "test-model",
	}
	llm := NewLLMServiceWithClient(cfg, server.Client())

	preview, err := llm.ParseFood("apple")
	if err != nil {
		t.Fatalf("Expected success after retry, got error: %v", err)
	}
	if attempts != 2 {
		t.Errorf("Expected 2 attempts, got %d", attempts)
	}
	if preview.Macros.Calories != 100 {
		t.Errorf("Expected 100 calories, got %f", preview.Macros.Calories)
	}
}

func TestLLMService_AnalyzeReview_RetrySuccess(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 2 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		resp := chatResponse{
			Choices: []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			}{
				{Message: struct {
					Content string `json:"content"`
				}{Content: `{"summary": "Good", "score": 80}`}},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	cfg := &config.Config{
		SambaAPIKey:   "test-key",
		OpenAIBaseURL: server.URL,
		ReviewModel:   "test-model",
	}
	llm := NewLLMServiceWithClient(cfg, server.Client())

	result, err := llm.AnalyzeReview(models.ReviewData{Goal: "Lose weight"})
	if err != nil {
		t.Fatalf("Expected success after retry, got error: %v", err)
	}
	if attempts != 2 {
		t.Errorf("Expected 2 attempts, got %d", attempts)
	}
	if result.Score != 80 {
		t.Errorf("Expected score 80, got %d", result.Score)
	}
}

func TestLLMService_CleanJSON_EdgeCases(t *testing.T) {
	llm := &LLMService{}

	testCases := []struct {
		input    string
		expected string
	}{
		{`{"calories": "100.5 kcal"}`, `{"calories": 100.5}`},
		{`{"protein": "10g"}`, `{"protein": 10}`},
		{`{"carbs": "25.0"}`, `{"carbs": 25.0}`},
		{`{"fat": "2.5mg"}`, `{"fat": 2.5}`},
		{`{"water": "500 ml"}`, `{"water": 500}`},
	}

	for _, tc := range testCases {
		got := llm.cleanJSON(tc.input)
		if got != tc.expected {
			t.Errorf("cleanJSON(%s) = %s, want %s", tc.input, got, tc.expected)
		}
	}
}
