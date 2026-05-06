package services

import (
	"calorie-tracker/models"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
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
}

func (p *SerpAPIProvider) ResolveFood(item ParsedFood) (*models.ReferenceFood, error) {
	if p == nil || item.Name == "" {
		return nil, nil
	}

	query := item.Name + " nutrition facts"
	u, _ := url.Parse("https://serpapi.com/search")
	q := u.Query()
	q.Set("engine", "google")
	q.Set("q", query)
	q.Set("api_key", p.apiKey)
	u.RawQuery = q.Encode()

	resp, err := p.client.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, nil
	}

	body, _ := io.ReadAll(resp.Body)
	var result serpAPIResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if result.NutritionInformation.Calories == "" {
		return nil, nil
	}

	ref := &models.ReferenceFood{
		Name: item.Name,
	}

	ref.Macros.Calories, _ = strconv.ParseFloat(result.NutritionInformation.Calories, 64)
	ref.Macros.Fat = p.extractValue(result.NutritionInformation.TotalFat)
	ref.Macros.Carbs = p.extractValue(result.NutritionInformation.TotalCarbohydrate)
	ref.Macros.Protein = p.extractValue(result.NutritionInformation.Protein)

	ref.BaseQuantity, ref.Unit = p.parseAmountPer(result.NutritionInformation.AmountPer)

	if ref.Macros.Calories == 0 && ref.Macros.Fat == 0 && ref.Macros.Carbs == 0 && ref.Macros.Protein == 0 {
		return nil, nil
	}

	return ref, nil
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
