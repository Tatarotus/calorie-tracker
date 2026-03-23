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

type LLMService struct {
	config *config.Config
}

func NewLLMService(cfg *config.Config) *LLMService {
	return &LLMService{config: cfg}
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func (s *LLMService) ParseFood(description string) (*models.FoodPreview, error) {
	prompt := fmt.Sprintf(`You are a nutrition expert. Analyze the following food description and return a JSON block with nutritional estimates for the given portion.
{
  "calories": number,
  "protein": number,
  "carbs": number,
  "fat": number
}
Food: "%s"

Rules:
1. Return ONLY the JSON block, no other text.
2. Use numbers ONLY for all fields.
3. DO NOT include units (like "g", "kcal", etc.) in the JSON values.
4. Estimate accurately for traditional and regional dishes.`, description)

	var content string
	var err error
	for i := 0; i < 3; i++ {
		content, err = s.callLLM(s.config.FoodModel, prompt)
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

func (s *LLMService) AnalyzeReview(data models.ReviewData) (*models.ReviewResult, error) {
	jsonData, _ := json.MarshalIndent(data, "", "  ")
	prompt := fmt.Sprintf(`You are a nutrition and performance analyst.
Analyze the following user data against their current goal.

Goal: %s

Return a JSON response with EXACTLY this structure:
{
  "summary": "string",
  "goal_progress": "string (detailed evaluation of progress towards the specific goal)",
  "progress": "improving" | "stable" | "regressing",
  "score": number,
  "issues": ["string"],
  "suggestions": ["string"],
  "patterns": ["string"]
}

Data:
%s

Rules:
1. Base ONLY on the provided data and evaluate specifically against the Goal.
2. Be specific, no generic advice.
3. Return ONLY a valid JSON block.
4. Use lowercase keys as shown above.`, data.Goal, string(jsonData))

	var content string
	var err error
	for i := 0; i < 3; i++ {
		content, err = s.callLLM(s.config.ReviewModel, prompt)
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

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
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
