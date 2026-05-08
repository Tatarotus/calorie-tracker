package services

import (
	"calorie-tracker/models"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type SerpAPIProvider struct {
	client *http.Client
	apiKey string
}

func NewSerpAPIProvider(apiKey string) *SerpAPIProvider {
	if apiKey == "" {
		return nil
	}
	return &SerpAPIProvider{
		client: &http.Client{Timeout: 15 * time.Second},
		apiKey: apiKey,
	}
}

type serpAPIResponse struct {
	NutritionInformation struct {
		Calories          string      `json:"calories"`
		TotalFat          interface{} `json:"total_fat"`
		TotalCarbohydrate interface{} `json:"total_carbohydrate"`
		Protein           interface{} `json:"protein"`
		AmountPer         string      `json:"amount_per"`
	} `json:"nutrition_information"`
	TextBlocks []struct {
		Type string `json:"type"`
		List []struct {
			Snippet string `json:"snippet"`
		} `json:"list"`
		Snippet string `json:"snippet"`
	} `json:"text_blocks"`
}

func (p *SerpAPIProvider) ResolveFood(item ParsedFood) (*models.ReferenceFood, error) {
	if p == nil || item.Name == "" {
		return nil, nil
	}

	// Try google_ai_mode first with "macros" query as it often provides curated info
	ref, err := p.resolveWithEngine(item, "google_ai_mode", item.Name+" macros")
	if err == nil && ref != nil {
		return ref, nil
	}

	// Fallback to standard google engine with "nutrition facts"
	return p.resolveWithEngine(item, "google", item.Name+" nutrition facts")
}

func (p *SerpAPIProvider) resolveWithEngine(item ParsedFood, engine, query string) (*models.ReferenceFood, error) {
	u, _ := url.Parse("https://serpapi.com/search")
	q := u.Query()
	q.Set("engine", engine)
	q.Set("q", query)
	q.Set("api_key", p.apiKey)
	u.RawQuery = q.Encode()

	resp, err := p.client.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, nil
	}

	body, _ := io.ReadAll(resp.Body)
	var result serpAPIResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	// Try extracting from NutritionInformation (standard google engine)
	if result.NutritionInformation.Calories != "" {
		ref := &models.ReferenceFood{
			Name: item.Name,
		}
		ref.Macros.Calories, _ = strconv.ParseFloat(result.NutritionInformation.Calories, 64)
		ref.Macros.Fat = p.extractValue(result.NutritionInformation.TotalFat)
		ref.Macros.Carbs = p.extractValue(result.NutritionInformation.TotalCarbohydrate)
		ref.Macros.Protein = p.extractValue(result.NutritionInformation.Protein)
		ref.BaseQuantity, ref.Unit = p.parseAmountPer(result.NutritionInformation.AmountPer)

		if ref.Macros.Calories > 0 || ref.Macros.Fat > 0 || ref.Macros.Carbs > 0 || ref.Macros.Protein > 0 {
			return ref, nil
		}
	}

	// Try extracting from TextBlocks (google_ai_mode)
	if len(result.TextBlocks) > 0 {
		ref := p.parseTextBlocks(result.TextBlocks, item.Name)
		if ref != nil {
			return ref, nil
		}
	}

	return nil, nil
}

func (p *SerpAPIProvider) parseTextBlocks(blocks []struct {
	Type string `json:"type"`
	List []struct {
		Snippet string `json:"snippet"`
	} `json:"list"`
	Snippet string `json:"snippet"`
}, name string) *models.ReferenceFood {
	var bestRef *models.ReferenceFood
	var bestScore int

	for i, block := range blocks {
		if block.Type == "list" && len(block.List) > 0 {
			macros, foundCount := p.extractMacrosFromList(block.List)

			if foundCount >= 2 && macros.Calories > 0 {
				score := 1
				if i > 0 && blocks[i-1].Type == "heading" {
					h := strings.ToLower(blocks[i-1].Snippet)
					if strings.Contains(h, "cozida") || strings.Contains(h, "refogada") || strings.Contains(h, "cooked") {
						score = 10
					} else if strings.Contains(h, "crua") || strings.Contains(h, "raw") {
						score = 5
					} else {
						score = 3
					}
				}

				if score > bestScore {
					bestScore = score
					bestRef = &models.ReferenceFood{
						Name:         name,
						BaseQuantity: 100,
						Unit:         "gram",
						Macros:       macros,
					}
				}
			}
		}
	}
	return bestRef
}

func (p *SerpAPIProvider) extractMacrosFromList(list []struct {
	Snippet string `json:"snippet"`
}) (models.Macros, int) {
	macros := models.Macros{}
	foundCount := 0
	for _, item := range list {
		s := strings.ToLower(item.Snippet)
		val := p.extractFloat(s)
		if strings.Contains(s, "caloria") || strings.Contains(s, "calories") || strings.Contains(s, "energia") {
			macros.Calories = val
			foundCount++
		} else if strings.Contains(s, "carboidrato") || strings.Contains(s, "carbohydrate") || strings.Contains(s, "carbs") {
			macros.Carbs = val
			foundCount++
		} else if strings.Contains(s, "proteína") || strings.Contains(s, "proteina") || strings.Contains(s, "protein") {
			macros.Protein = val
			foundCount++
		} else if strings.Contains(s, "gordura") || strings.Contains(s, "fat") || strings.Contains(s, "lipídeo") || strings.Contains(s, "lipideo") {
			macros.Fat = val
			foundCount++
		}
	}
	return macros, foundCount
}

func (p *SerpAPIProvider) extractFloat(s string) float64 {
	// First, try to find a number that might include dots and commas
	re := regexp.MustCompile(`(\d+([.,]\d+)*)`)
	matches := re.FindStringSubmatch(s)
	if len(matches) > 1 {
		valStr := matches[1]
		// If both comma and dot are present, e.g., 1.234,56
		if strings.Contains(valStr, ",") && strings.Contains(valStr, ".") {
			commaIdx := strings.LastIndex(valStr, ",")
			dotIdx := strings.LastIndex(valStr, ".")
			if commaIdx > dotIdx {
				// Comma is decimal, remove dots
				valStr = strings.ReplaceAll(valStr, ".", "")
				valStr = strings.Replace(valStr, ",", ".", 1)
			} else {
				// Dot is decimal, remove commas
				valStr = strings.ReplaceAll(valStr, ",", "")
			}
		} else if strings.Contains(valStr, ",") {
			// Only comma present, assume it's decimal (e.g., 2,7)
			// Unless it looks like a thousands separator (e.g., 1,000)
			// But for macros, 1,000 is more likely 1.0 than 1000.
			// However, TACO table uses 29,00 for 29.00.
			valStr = strings.Replace(valStr, ",", ".", 1)
		}
		// If only dot is present, it's already in the format ParseFloat expects (e.g., 12.5 or 1,000 as 1.000)

		val, _ := strconv.ParseFloat(valStr, 64)
		return val
	}
	return 0
}

func (p *SerpAPIProvider) extractValue(v interface{}) float64 {
	var s string
	switch val := v.(type) {
	case string:
		s = val
	case []interface{}:
		if len(val) > 0 {
			if str, ok := val[0].(string); ok {
				s = str
			}
		}
	}

	if s == "" {
		return 0
	}

	re := regexp.MustCompile(`^(\d+(?:\.\d+)?)\s*g`)
	matches := re.FindStringSubmatch(s)
	if len(matches) > 1 {
		val, _ := strconv.ParseFloat(matches[1], 64)
		return val
	}
	return 0
}

func (p *SerpAPIProvider) parseAmountPer(s string) (float64, string) {
	if s == "" {
		return 100, "gram"
	}

	// Example: "1 cup (248 g)" or "100 g"
	reGrams := regexp.MustCompile(`\((\d+)\s*g\)`)
	matches := reGrams.FindStringSubmatch(s)
	if len(matches) > 1 {
		val, _ := strconv.ParseFloat(matches[1], 64)
		return val, "gram"
	}

	reSimple := regexp.MustCompile(`^(\d+)\s*g`)
	matches = reSimple.FindStringSubmatch(s)
	if len(matches) > 1 {
		val, _ := strconv.ParseFloat(matches[1], 64)
		return val, "gram"
	}

	return 100, "gram"
}
