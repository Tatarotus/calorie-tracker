package services

import (
	"testing"

	"calorie-tracker/config"
)

func TestCleanJSON(t *testing.T) {
	llm := &LLMService{}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "remove units after numbers",
			input:    `{"calories": 100g, "protein": "20g", "carbs": 50kcal, "fat": 10mg}`,
			expected: `{"calories": 100, "protein": 20, "carbs": 50, "fat": 10}`,
		},
		{
			name:     "remove quotes around numbers",
			input:    `{"calories": "100", "protein": "20"}`,
			expected: `{"calories": 100, "protein": 20}`,
		},
		{
			name:     "remove various units",
			input:    `{"value": 25ml, "other": 30units, "fat": 5fatias}`,
			expected: `{"value": 25, "other": 30, "fat": 5}`,
		},
		{
			name:     "decimal numbers with units",
			input:    `{"calories": 100.5g, "protein": 20.3kcal}`,
			expected: `{"calories": 100.5, "protein": 20.3}`,
		},
		{
			name:     "no changes needed",
			input:    `{"calories": 100, "protein": 20}`,
			expected: `{"calories": 100, "protein": 20}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := llm.cleanJSON(tt.input)
			if result != tt.expected {
				t.Errorf("cleanJSON(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestExtractJSON(t *testing.T) {
	llm := &LLMService{}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "markdown json block",
			input:    "Here is the result:\n```json\n{\"calories\": 100}\n```\nEnd",
			expected: `{"calories": 100}`,
		},
		{
			name:     "markdown block without json tag",
			input:    "Result:\n```\n{\"protein\": 20}\n```\nDone",
			expected: `{"protein": 20}`,
		},
		{
			name:     "plain json object",
			input:    "The answer is {\"calories\": 150} here",
			expected: `{"calories": 150}`,
		},
		{
			name:     "multiple braces - take first to last",
			input:    "Start {\"a\": 1} middle {\"b\": 2} end",
			expected: `{"a": 1} middle {"b": 2}`,
		},
		{
			name:     "no braces - return as is",
			input:    "no json here",
			expected: "no json here",
		},
		{
			name:     "only opening brace",
			input:    "text { more text",
			expected: "text { more text",
		},
		{
			name:     "only closing brace",
			input:    "text } more text",
			expected: "text } more text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := llm.extractJSON(tt.input)
			if result != tt.expected {
				t.Errorf("extractJSON(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNewLLMService(t *testing.T) {
	cfg := &config.Config{
		SambaAPIKey: "test-key",
		OpenAIBaseURL: "https://test.url/v1",
		FoodModel: "test-model",
		ReviewModel: "test-model-2",
	}

	llm := NewLLMService(cfg)
	if llm == nil {
		t.Fatal("Expected non-nil LLMService")
	}
	if llm.config != cfg {
		t.Error("Expected config to be set")
	}
}
