package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
)

// MockLLMProvider is a mock implementation of LLMProvider for testing
type MockLLMProvider struct {
	Response       string
	Error          error
	CallCount      int
	LastModel      string
	LastPrompt     string
	FixedResponses map[string]string // model -> response mapping
}

// Call implements LLMProvider interface
func (m *MockLLMProvider) Call(model, prompt string) (string, error) {
	m.CallCount++
	m.LastModel = model
	m.LastPrompt = prompt

	if m.Error != nil {
		return "", m.Error
	}

	// Check for model-specific response
	if resp, ok := m.FixedResponses[model]; ok {
		return resp, nil
	}

	return m.Response, nil
}

// NewMockLLMProvider creates a new mock LLM provider
func NewMockLLMProvider() *MockLLMProvider {
	return &MockLLMProvider{
		FixedResponses: make(map[string]string),
	}
}

// SetResponse sets the response for all calls
func (m *MockLLMProvider) SetResponse(response string) {
	m.Response = response
}

// SetError sets an error to be returned
func (m *MockLLMProvider) SetError(err error) {
	m.Error = err
}

// SetFixedResponse sets a response for a specific model
func (m *MockLLMProvider) SetFixedResponse(model, response string) {
	m.FixedResponses[model] = response
}

// Reset clears all state
func (m *MockLLMProvider) Reset() {
	m.Response = ""
	m.Error = nil
	m.CallCount = 0
	m.LastModel = ""
	m.LastPrompt = ""
}

// MockHTTPServer creates a test HTTP server that returns mock LLM responses
func MockHTTPServer(response string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		resp := fmt.Sprintf(`{
			"choices": [{
				"message": {
					"content": %q
				}
			}]
		}`, response)

		w.Write([]byte(resp))
	}))
}

// MockHTTPServerError creates a test HTTP server that returns an error
func MockHTTPServerError(statusCode int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
		w.Write([]byte("Error"))
	}))
}

// MockFoodPreviewResponse returns a valid JSON response for food parsing
// The return value is the JSON object as a string (what the LLM would return)
func MockFoodPreviewResponse(desc string, calories, protein, carbs, fat float64) string {
	return fmt.Sprintf(`{"calories": %.2f, "protein": %.2f, "carbs": %.2f, "fat": %.2f}`, calories, protein, carbs, fat)
}

// MockReviewResult returns a valid JSON response for review analysis (properly escaped for embedding)
func MockReviewResultResponse() string {
	return `{"summary": "Good progress overall", "goal_progress": "On track to meet goals", "progress": "improving", "score": 85, "issues": ["Low protein on Monday"], "suggestions": ["Add more protein-rich foods"], "patterns": ["Consistent breakfast habits"]}`
}

// MockHTTPServerWithJSON creates a test server that returns parsed JSON
func MockHTTPServerWithJSON(model, jsonContent string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Parse the request to verify it's correct
		var req chatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		resp := fmt.Sprintf(`{
			"choices": [{
				"message": {
					"content": %q
				}
			}]
		}`, jsonContent)

		w.Write([]byte(resp))
	}))
}
