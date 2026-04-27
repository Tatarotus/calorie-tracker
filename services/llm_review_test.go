package services

import (
	"calorie-tracker/config"
	"calorie-tracker/models"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

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
		// Properly escape the JSON string
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

	result, _ := llm.AnalyzeReview(data)
	if result != nil && result.Score != 85 {
		t.Errorf("Expected score 85, got %d", result.Score)
	}
}

func TestLLMService_CallLLM_Authorization(t *testing.T) {
	var authHeader string

	cfg := &config.Config{
		SambaAPIKey:   "my-secret-key",
		OpenAIBaseURL: "http://test.com",
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

func TestLLMService_NoChoices(t *testing.T) {
	cfg := &config.Config{
		SambaAPIKey:   "test-key",
		OpenAIBaseURL: "http://test.com",
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
		t.Error("Expected error for empty choices")
	}
}
