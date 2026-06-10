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

// LLMService implements LLMProvider interface and Validator, Parser
type LLMService struct {
	config *config.Config
	client *http.Client
}

type ParsedFoodItemsResponse struct {
	Items []ParsedFoodItem `json:"items"`
}

type ParsedFoodItem struct {
	Quantity     float64 `json:"quantity"`
	Unit         string  `json:"unit"`
	FoodName     string  `json:"food_name"`
	CanonicalKey string  `json:"canonical_key"`
	DisplayName  string  `json:"-"`
}

// NewLLMService creates a new LLMService with default HTTP client
func NewLLMService(cfg *config.Config) *LLMService {
	return &LLMService{
		config: cfg,
		client: SharedHTTPClient,
	}
}

// NewLLMServiceWithClient creates a new LLMService with a custom HTTP client
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

// ParseFood analyzes food description and returns deterministic reference data
func (s *LLMService) ParseFood(description string) (*models.ReferenceFood, error) {
	prompt := fmt.Sprintf(`Analyze the following food description and return nutritional information.
Description: %s

You MUST return a STRICT JSON object with this exact structure:
{
  "name": "standardized food name",
  "base_quantity": 100,
  "unit": "g",
  "macros": {
    "calories": number,
    "protein": number,
    "carbs": number,
    "fat": number
  }
}

Rules:
1. "base_quantity" should usually be 100 for grams/ml or 1 for units.
2. "unit" should be "g", "ml", or "unit".
3. "macros" are for the "base_quantity".
4. If multiple items are mentioned, return estimated macros for the entire described meal, using "base_quantity": 1 and "unit": "unit".
5. If the description uses "u", "unit", "units", "unid", "unidade", or "unidades", treat it as a whole item and return "base_quantity": 1 and "unit": "unit".
6. NO prose, NO markdown blocks, only the raw JSON.`, description)

	var content string
	var err error
	for i := 0; i < 3; i++ {
		content, err = s.Call(s.config.FoodModel, prompt)
		if err == nil && content != "" {
			break
		}
		if err == nil && content == "" {
			err = fmt.Errorf("empty response from LLM")
		}
		time.Sleep(1 * time.Second)
	}

	if err != nil {
		return nil, err
	}

	var result models.ReferenceFood
	if err := s.parseLLMResponse(content, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// ParseFoodItems extracts food items from a meal description without macro values.
func (s *LLMService) ParseFoodItems(description string) ([]ParsedFoodItem, error) {
	prompt := fmt.Sprintf(`Extract food items from this meal description.
Description: %s

Return ONLY strict JSON with this exact shape:
{
  "items": [
    {
      "food_name": "name of food in lowercase (e.g., 'banana', 'cassava', 'coffee with milk')",
      "quantity": number,
      "unit": "gram|ml|unit|cup|tablespoon|teaspoon|bowl|plate|serving|slice|handful",
      "canonical_key": "snake_case_canonical_key (e.g., 'banana', 'macaxeira', 'cafe_com_leite', 'leite')"
    }
  ]
}

Rules:
1. Extract every food item, including oils, butter, sauces, and sides.
2. Use 'unit' for whole countable foods like eggs and bananas; use 'slice' for slices/fatias.
3. Use quantity 1 for vague singular portions.
4. Convert regional terms to clean canonical snake_case keys (e.g. 'macaxeira' -> 'macaxeira', 'aipim' -> 'macaxeira', 'pão francês' -> 'pao_frances', 'café com leite' -> 'cafe_com_leite').
5. Do not return calories or macros.
6. NO prose, NO markdown blocks, only raw JSON.`, description)

	var content string
	var err error
	for i := 0; i < 3; i++ {
		content, err = s.Call(s.config.FoodModel, prompt)
		if err == nil && content != "" {
			break
		}
		if err == nil && content == "" {
			err = fmt.Errorf("empty response from LLM")
		}
		time.Sleep(1 * time.Second)
	}

	if err != nil {
		return nil, err
	}

	var result ParsedFoodItemsResponse
	if err := s.parseLLMResponse(content, &result); err != nil {
		return nil, err
	}

	parser := NewFoodParser()
	items := make([]ParsedFoodItem, 0, len(result.Items))
	for _, item := range result.Items {
		name := parser.normalizeName(item.FoodName)
		if name == "" {
			continue
		}
		unit := parser.normalizeUnit(item.Unit)
		key := strings.ToLower(strings.TrimSpace(item.CanonicalKey))
		if key == "" {
			key = strings.ReplaceAll(name, " ", "_")
		}
		items = append(items, ParsedFoodItem{
			Quantity:     item.Quantity,
			Unit:         unit,
			FoodName:     name,
			CanonicalKey: key,
		})
	}

	return items, nil
}

// Validate checks if the resolved food is semantically equivalent or a valid match for the original query.
func (s *LLMService) Validate(query string, resolved *models.ReferenceFood) (bool, []string, error) {
	prompt := fmt.Sprintf(`Compare the user's original food description with the resolved database reference food.
Original description: "%s"
Resolved reference food name: "%s"
Resolved base serving: %g %s

Determine if this is a dangerous or incorrect semantic mismatch (e.g. coffee with milk matched to black coffee, or cassava matched to potato, or a completely different food).
Return ONLY strict JSON with this exact shape:
{
  "valid": boolean,
  "warnings": ["string warning messages describing any discrepancies, if any"]
}

Rules:
1. "valid" should be false only if there is a real semantic discrepancy or incorrect food category match.
2. Moderate differences in portion/preparation can be handled with warnings but marked valid.
3. NO prose, NO markdown blocks, only raw JSON.`, query, resolved.Name, resolved.BaseQuantity, resolved.Unit)

	var content string
	var err error
	for i := 0; i < 3; i++ {
		content, err = s.Call(s.config.FoodModel, prompt)
		if err == nil && content != "" {
			break
		}
		if err == nil && content == "" {
			err = fmt.Errorf("empty response from LLM")
		}
		time.Sleep(1 * time.Second)
	}

	if err != nil {
		return false, nil, err
	}

	type ValidateResponse struct {
		Valid    bool     `json:"valid"`
		Warnings []string `json:"warnings"`
	}
	var res ValidateResponse
	if err := s.parseLLMResponse(content, &res); err != nil {
		return false, nil, fmt.Errorf("failed to parse validation response: %w", err)
	}

	return res.Valid, res.Warnings, nil
}

// AnalyzeReview implements ReviewAnalyzer interface
func (s *LLMService) AnalyzeReview(data models.ReviewData) (*models.ReviewResult, error) {
	jsonData, _ := json.MarshalIndent(data, "", " ")
	prompt := fmt.Sprintf(`You are a nutrition and performance analyst. Analyze the following user data against their current goal. Goal: %s Data includes: - Daily summarized stats (calories, protein, carbs, fat, water) - Individual food entries - Individual water entries Return a JSON response with EXACTLY this structure: { "summary": "string (concise overall evaluation)", "goal_progress": "string (detailed evaluation of progress towards the specific goal)", "progress": "improving" | "stable" | "regressing", "score": number (0-100 based on consistency and goal alignment), "issues": ["string (specific concerns about nutrition, macros, or hydration)"], "suggestions": ["string (actionable advice)"], "patterns": ["string (identified habits or trends)"] } Data: %s Rules: 1. Base ONLY on the provided data and evaluate specifically against the Goal. 2. Analyze macro-nutrient balance (protein, carbs, fat) and hydration levels. 3. Be specific, no generic advice. 4. Return ONLY a valid JSON block. 5. Use lowercase keys as shown above.`, data.Goal, string(jsonData))

	var content string
	var err error
	for i := 0; i < 3; i++ {
		content, err = s.Call(s.config.ReviewModel, prompt)
		if err == nil && content != "" {
			break
		}
		if err == nil && content == "" {
			err = fmt.Errorf("empty response from LLM")
		}
		time.Sleep(1 * time.Second)
	}

	if err != nil {
		return nil, err
	}

	var result models.ReviewResult
	if err := s.parseLLMResponse(content, &result); err != nil {
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
		MaxTokens: 1024,
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
	defer func() { _ = resp.Body.Close() }()

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

// sanitizeJSON fixes common LLM JSON malformations:
// - Unescaped newlines, tabs, carriage returns inside string values
// - Trailing commas before closing brackets/braces
// - Single-quoted strings (rare but possible)
func (s *LLMService) sanitizeJSON(jsonStr string) string {
	// Remove markdown code block markers if present
	jsonStr = strings.TrimSpace(jsonStr)
	jsonStr = strings.TrimPrefix(jsonStr, "```json")
	jsonStr = strings.TrimPrefix(jsonStr, "```")
	jsonStr = strings.TrimSuffix(jsonStr, "```")
	jsonStr = strings.TrimSpace(jsonStr)

	// Fix unescaped control characters inside string values.
	// We scan character by character, tracking whether we're inside a string.
	var sb strings.Builder
	inString := false
	escaped := false

	for _, r := range jsonStr {
		if escaped {
			// If the previous char was a backslash, just write this char as-is
			sb.WriteRune(r)
			escaped = false
			continue
		}

		if r == '\\' {
			sb.WriteRune(r)
			escaped = true
			continue
		}

		if r == '"' {
			inString = !inString
			sb.WriteRune(r)
			continue
		}

		if inString {
			// Inside a string: escape control characters
			switch r {
			case '\n':
				sb.WriteString("\\n")
			case '\t':
				sb.WriteString("\\t")
			case '\r':
				sb.WriteString("\\r")
			case '\b':
				sb.WriteString("\\b")
			case '\f':
				sb.WriteString("\\f")
			default:
				// Also escape any other control characters (0x00-0x1F except those handled above)
				if r >= 0x00 && r <= 0x1F {
					fmt.Fprintf(&sb, "\\u%04x", r)
				} else {
					sb.WriteRune(r)
				}
			}
		} else {
			// Outside a string: write as-is
			sb.WriteRune(r)
		}
	}

	result := sb.String()

	// Fix trailing commas before } or ]
	// Pattern: ,\s*([}\]]) -> $1
	reTrailingComma := regexp.MustCompile(`,\s*([}\]])`)
	result = reTrailingComma.ReplaceAllString(result, "$1")

	return result
}

// chatMessage represents a message in the chat request
type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// chatRequest represents the request to the chat API
type chatRequest struct {
	Model     string        `json:"model"`
	Messages  []chatMessage `json:"messages"`
	MaxTokens int           `json:"max_tokens,omitempty"`
}

// chatResponse represents the response from the chat API
type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// parseLLMResponse robustly extracts and parses JSON candidates from LLM response.
func (s *LLMService) parseLLMResponse(content string, dest interface{}) error {
	var candidates []string

	// 1. Try to find all markdown blocks (both ```json and ```)
	reCodeBlock := regexp.MustCompile("(?s)```(?:json)?\n*(.*?)\n*```")
	matches := reCodeBlock.FindAllStringSubmatch(content, -1)
	for _, m := range matches {
		if len(m) > 1 {
			candidates = append(candidates, strings.TrimSpace(m[1]))
		}
	}

	// 2. Also try brace-delimited blocks (matching '{' and '}')
	startIdx := 0
	for {
		start := strings.Index(content[startIdx:], "{")
		if start == -1 {
			break
		}
		start += startIdx

		depth := 0
		end := -1
		for i := start; i < len(content); i++ {
			if content[i] == '{' {
				depth++
			} else if content[i] == '}' {
				depth--
				if depth == 0 {
					end = i
					break
				}
			}
		}

		if end != -1 {
			candidates = append(candidates, content[start:end+1])
			startIdx = start + 1
		} else {
			break
		}
	}

	// 3. Fallback: try the whole content
	candidates = append(candidates, content)

	// Try parsing each candidate in order
	var lastErr error
	for _, cand := range candidates {
		cand = s.sanitizeJSON(cand)
		cand = s.cleanJSON(cand)

		if err := json.Unmarshal([]byte(cand), dest); err == nil {
			return nil
		} else {
			lastErr = err
		}
	}

	if lastErr != nil {
		return fmt.Errorf("failed to parse JSON from candidates (last error: %w), content: %s", lastErr, content)
	}
	return fmt.Errorf("no JSON candidates found in content: %s", content)
}
