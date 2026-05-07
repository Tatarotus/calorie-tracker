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

func TestLLMService_SanitizeJSON_NewlinesInStrings(t *testing.T) {
	llm := &LLMService{}

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "newline in string value",
			input:    `{"name": "some food\nwith a newline", "amount": 100}`,
			expected: `{"name": "some food\nwith a newline", "amount": 100}`,
		},
		{
			name:     "tab in string value",
			input:    `{"name": "some food\twith a tab", "amount": 100}`,
			expected: `{"name": "some food\twith a tab", "amount": 100}`,
		},
		{
			name:     "carriage return in string value",
			input:    `{"name": "some food\rwith a cr", "amount": 100}`,
			expected: `{"name": "some food\rwith a cr", "amount": 100}`,
		},
		{
			name:     "multiple control chars in string",
			input:    `{"name": "line1\nline2\tcol", "amount": 100}`,
			expected: `{"name": "line1\nline2\tcol", "amount": 100}`,
		},
		{
			name:  "newline outside string (preserved as valid JSON whitespace)",
			input: "{\"name\": \"test\",\n\"amount\": 100}",
			expected: `{"name": "test",
"amount": 100}`,
		},
		{
			name:     "trailing comma before closing brace",
			input:    `{"name": "test", "amount": 100,}`,
			expected: `{"name": "test", "amount": 100}`,
		},
		{
			name:     "trailing comma before closing bracket",
			input:    `{"items": [1, 2, 3,]}`,
			expected: `{"items": [1, 2, 3]}`,
		},
		{
			name:     "markdown code block markers",
			input:    "```json\n{\"name\": \"test\"}\n```",
			expected: `{"name": "test"}`,
		},
		{
			name:     "complex malformed JSON with newlines and trailing comma",
			input:    "{\"items\": [{\"name\": \"food\nname\", \"amount\": 100,}]}",
			expected: `{"items": [{"name": "food\nname", "amount": 100}]}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := llm.sanitizeJSON(tc.input)
			if got != tc.expected {
				t.Errorf("sanitizeJSON() = %q, want %q", got, tc.expected)
			}
		})
	}
}

func TestLLMService_ParseFoodItems_MalformedJSONWithNewlines(t *testing.T) {
	// Simulate an LLM response with unescaped newlines inside a string value
	malformedResponse := `{
  "items": [
    {
      "name": "some food
with a newline",
      "amount": 100,
      "unit": "gram",
      "confidence": 0.95
    }
  ]
}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		escapedContent, _ := json.Marshal(malformedResponse)
		response := `{"choices": [{"message": {"content": ` + string(escapedContent) + `}}]}`
		w.Write([]byte(response))
	}))
	defer server.Close()

	cfg := &config.Config{
		SambaAPIKey:   "test-key",
		OpenAIBaseURL: server.URL,
		FoodModel:     "test-model",
	}
	llm := NewLLMServiceWithClient(cfg, server.Client())

	items, err := llm.ParseFoodItems("some food with a newline")
	if err != nil {
		t.Fatalf("Expected ParseFoodItems to handle malformed JSON, got error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(items))
	}
	if items[0].Name != "food with newline" {
		t.Errorf("Expected name 'food with newline', got %q", items[0].Name)
	}
	if items[0].Amount != 100 {
		t.Errorf("Expected amount 100, got %f", items[0].Amount)
	}
	if items[0].Unit != "gram" {
		t.Errorf("Expected unit 'gram', got %q", items[0].Unit)
	}
}

func TestLLMService_ParseFood_MalformedJSONWithNewlines(t *testing.T) {
	// Simulate an LLM response with unescaped newlines inside a string value
	malformedResponse := `{
  "name": "test food
with newline",
  "base_quantity": 100,
  "unit": "g",
  "macros": {
    "calories": 150,
    "protein": 5,
    "carbs": 20,
    "fat": 3
  }
}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		escapedContent, _ := json.Marshal(malformedResponse)
		response := `{"choices": [{"message": {"content": ` + string(escapedContent) + `}}]}`
		w.Write([]byte(response))
	}))
	defer server.Close()

	cfg := &config.Config{
		SambaAPIKey:   "test-key",
		OpenAIBaseURL: server.URL,
		FoodModel:     "test-model",
	}
	llm := NewLLMServiceWithClient(cfg, server.Client())

	result, err := llm.ParseFood("test food with newline")
	if err != nil {
		t.Fatalf("Expected ParseFood to handle malformed JSON, got error: %v", err)
	}
	if result.Macros.Calories != 150 {
		t.Errorf("Expected calories 150, got %f", result.Macros.Calories)
	}
	if result.Macros.Protein != 5 {
		t.Errorf("Expected protein 5, got %f", result.Macros.Protein)
	}
}
