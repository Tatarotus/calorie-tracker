package services

import (
	"bytes"
	"calorie-tracker/config"
	"calorie-tracker/models"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// LLMService implements LLMProvider interface
type LLMService struct {
	config *config.Config
	client *http.Client
}

// NewLLMService creates a new LLMService with default HTTP client
func NewLLMService(cfg *config.Config) *LLMService {
	return &LLMService{
		config: cfg,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// NewLLMServiceWithClient creates a new LLMService with a custom HTTP client
// This is useful for testing with a mock client
func NewLLMServiceWithClient(cfg *config.Config, client *http.Client) *LLMService {
	return &LLMService{
		config: cfg,
		client: client,
	}
}

// Call implements the LLMProvider interface
func (s *LLMService) Call(model, prompt string) (string, error) {
	return s.callLLM(model, prompt)
}

// ParseFood implements FoodParser interface
func (s *LLMService) ParseFood(description string) (*models.FoodPreview, error) {
	prompt := fmt.Sprintf(`You are a nutrition expert. Analyze the following food description and return a JSON block with nutritional estimates for the given portion. { "calories": number, "protein": number, "carbs": number, "fat": number } Food: "%s" Rules: 1. Return ONLY the JSON block, no other text. 2. Use numbers ONLY for all fields. 3. DO NOT include units (like "g", "kcal", etc.) in the JSON values. 4. Estimate accurately for traditional and regional dishes.`, description)

	var content string
	var err error
	for i := 0; i < 3; i++ {
		content, err = s.Call(s.config.FoodModel, prompt)
		if err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}

	if err != nil {
		return nil, err
	}

	jsonStr := s.extractJSON(content)
	jsonStr = s.cleanJSON(jsonStr)

	var result models.FoodPreview
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("failed to parse LLM response: %w, content: %s", err, content)
	}

	result.Description = description
	return &result, nil
}

// AnalyzeReview implements ReviewAnalyzer interface
func (s *LLMService) AnalyzeReview(data models.ReviewData) (*models.ReviewResult, error) {
	jsonData, _ := json.MarshalIndent(data, "", " ")
	prompt := fmt.Sprintf(`You are a nutrition and performance analyst. Analyze the following user data against their current goal. Goal: %s Data includes: - Daily summarized stats (calories, protein, carbs, fat, water) - Individual food entries - Individual water entries Return a JSON response with EXACTLY this structure: { "summary": "string (concise overall evaluation)", "goal_progress": "string (detailed evaluation of progress towards the specific goal)", "progress": "improving" | "stable" | "regressing", "score": number (0-100 based on consistency and goal alignment), "issues": ["string (specific concerns about nutrition, macros, or hydration)"], "suggestions": ["string (actionable advice)"], "patterns": ["string (identified habits or trends)"] } Data: %s Rules: 1. Base ONLY on the provided data and evaluate specifically against the Goal. 2. Analyze macro-nutrient balance (protein, carbs, fat) and hydration levels. 3. Be specific, no generic advice. 4. Return ONLY a valid JSON block. 5. Use lowercase keys as shown above.`, data.Goal, string(jsonData))

	var content string
	var err error
	for i := 0; i < 3; i++ {
		content, err = s.Call(s.config.ReviewModel, prompt)
		if err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}

	if err != nil {
		return nil, err
	}

	jsonStr := s.extractJSON(content)

	var result models.ReviewResult
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("failed to parse LLM review: %w, content: %s", err, content)
	}

	return &result, nil
}

// callLLM makes the actual HTTP request to the LLM API
func (s *LLMService) callLLM(model, prompt string) (string, error) {
	reqBody, _ := json.Marshal(chatRequest{
		Model: model,
		Messages: []chatMessage{
			{Role: "user", Content: prompt},
		},
	})

	url := s.config.OpenAIBaseURL
	if !strings.HasSuffix(url, "/chat/completions") {
		url = strings.TrimSuffix(url, "/") + "/chat/completions"
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.config.SambaAPIKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("LLM API error (status %d): %s", resp.StatusCode, string(body))
	}

	var chatResp chatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return "", err
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no choices returned from LLM")
	}

	return chatResp.Choices[0].Message.Content, nil
}

// cleanJSON removes units and quotes from JSON strings
func (s *LLMService) cleanJSON(jsonStr string) string {
	// 1. Remove units like "g", "kcal", etc. when they follow a number
	reUnits := regexp.MustCompile(`(\d+(?:\.\d+)?)\s*(g|kcal|mg|ml|units|unidades|fatias|fatia)`)
	jsonStr = reUnits.ReplaceAllString(jsonStr, "$1")

	// 2. Remove quotes around numbers (e.g., "calories": "100" -> "calories": 100)
	// This ensures json.Unmarshal can handle them as float64
	reQuotes := regexp.MustCompile(`"(\d+(?:\.\d+)?)"`)
	jsonStr = reQuotes.ReplaceAllString(jsonStr, "$1")

	return jsonStr
}

// extractJSON extracts JSON from content that may contain markdown or other text
func (s *LLMService) extractJSON(content string) string {
	// First, try to find a markdown block
	if start := strings.Index(content, "```json"); start != -1 {
		rest := content[start+7:]
		if end := strings.Index(rest, "```"); end != -1 {
			return strings.TrimSpace(rest[:end])
		}
	}

	// Fallback to first '{' and last '}'
	start := strings.Index(content, "{")
	end := strings.LastIndex(content, "}")
	if start == -1 || end == -1 || end < start {
		return content
	}

	return content[start : end+1]
}

// chatMessage represents a message in the chat request
type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// chatRequest represents the request to the chat API
type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
}

// chatResponse represents the response from the chat API
type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}
