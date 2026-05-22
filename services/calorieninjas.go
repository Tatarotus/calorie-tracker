package services

import (
	"calorie-tracker/models"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

// CalorieNinjasProvider extracts nutrition facts from natural language queries
// using CalorieNinjas API and translates queries to/from Portuguese using MyMemory API.
type CalorieNinjasProvider struct {
	client *http.Client
}

// NewCalorieNinjasProvider creates a new CalorieNinjasProvider.
func NewCalorieNinjasProvider() *CalorieNinjasProvider {
	return &CalorieNinjasProvider{
		client: SharedHTTPClient,
	}
}

type calorieNinjasItem struct {
	Name                string  `json:"name"`
	Calories            float64 `json:"calories"`
	ServingSizeG        float64 `json:"serving_size_g"`
	FatTotalG           float64 `json:"fat_total_g"`
	FatSaturatedG       float64 `json:"fat_saturated_g"`
	ProteinG            float64 `json:"protein_g"`
	SodiumMg            float64 `json:"sodium_mg"`
	PotassiumMg         float64 `json:"potassium_mg"`
	CholesterolMg       float64 `json:"cholesterol_mg"`
	CarbohydratesTotalG float64 `json:"carbohydrates_total_g"`
	FiberG              float64 `json:"fiber_g"`
	SugarG              float64 `json:"sugar_g"`
}

type calorieNinjasResponse struct {
	Items []calorieNinjasItem `json:"items"`
}

// ResolveFood queries CalorieNinjas API and maps the parsed items to a models.ReferenceFood.
func (p *CalorieNinjasProvider) ResolveFood(item ParsedFood) (*models.ReferenceFood, error) {
	if p == nil || item.Name == "" {
		return nil, nil
	}

	cleanQuery := strings.TrimSpace(item.Name)

	// 1. Detect if the input is in Portuguese by translating pt|en
	englishQuery, err := p.translate(cleanQuery, "pt", "en")
	if err != nil {
		englishQuery = cleanQuery
	}

	// Normalization comparison to detect language
	normalizedOriginal := p.normalizeForLangDetection(cleanQuery)
	normalizedEnglish := p.normalizeForLangDetection(englishQuery)
	isPortugueseInput := normalizedOriginal != normalizedEnglish

	// Use translated query for API request if Portuguese is detected
	activeQuery := cleanQuery
	if isPortugueseInput {
		activeQuery = englishQuery
	}

	u, err := url.Parse("https://api.calorieninjas.com/v1/nutrition")
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("query", activeQuery)
	u.RawQuery = q.Encode()

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	// Set browser headers exactly like the original Node library
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Origin", "https://calorieninjas.com")
	req.Header.Set("Referer", "https://calorieninjas.com/")
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		var parsedError struct {
			Message string `json:"message"`
		}
		if unmarshalErr := json.Unmarshal(bodyBytes, &parsedError); unmarshalErr == nil && parsedError.Message != "" {
			return nil, fmt.Errorf("CalorieNinjas API responded with status %d: %s", resp.StatusCode, parsedError.Message)
		}
		return nil, fmt.Errorf("CalorieNinjas API responded with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data calorieNinjasResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	if len(data.Items) == 0 {
		return nil, nil
	}

	// 2. If input was in Portuguese, translate the parsed item names back into Portuguese in parallel
	if isPortugueseInput && len(data.Items) > 0 {
		type translationResult struct {
			index int
			name  string
			err   error
		}
		ch := make(chan translationResult, len(data.Items))
		for i, responseItem := range data.Items {
			go func(idx int, engName string) {
				ptName, tErr := p.translate(engName, "en", "pt")
				ch <- translationResult{index: idx, name: ptName, err: tErr}
			}(i, responseItem.Name)
		}
		for i := 0; i < len(data.Items); i++ {
			res := <-ch
			if res.err == nil && res.name != "" {
				data.Items[res.index].Name = res.name
			}
		}
	}

	// Map first item to ReferenceFood (this aligns with single-item provider resolutions)
	first := data.Items[0]
	ref := &models.ReferenceFood{
		Name:         first.Name,
		BaseQuantity: first.ServingSizeG,
		Unit:         "gram",
		Macros: models.Macros{
			Calories: first.Calories,
			Protein:  first.ProteinG,
			Carbs:    first.CarbohydratesTotalG,
			Fat:      first.FatTotalG,
		},
	}

	return ref, nil
}

// CalorieNinjasRecipe holds recipe info returned from CalorieNinjas API.
type CalorieNinjasRecipe struct {
	Title        string `json:"title"`
	Ingredients  string `json:"ingredients"`
	Servings     string `json:"servings"`
	Instructions string `json:"instructions"`
}

// QueryRecipes fetches recipes from CalorieNinjas API with auto-translation.
func (p *CalorieNinjasProvider) QueryRecipes(query string) ([]CalorieNinjasRecipe, error) {
	if query == "" {
		return nil, fmt.Errorf("query must be a non-empty string")
	}

	cleanQuery := strings.TrimSpace(query)
	englishQuery, err := p.translate(cleanQuery, "pt", "en")
	if err != nil {
		englishQuery = cleanQuery
	}

	normalizedOriginal := p.normalizeForLangDetection(cleanQuery)
	normalizedEnglish := p.normalizeForLangDetection(englishQuery)
	isPortugueseInput := normalizedOriginal != normalizedEnglish

	activeQuery := cleanQuery
	if isPortugueseInput {
		activeQuery = englishQuery
	}

	u, err := url.Parse("https://api.calorieninjas.com/v1/recipe")
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("query", activeQuery)
	u.RawQuery = q.Encode()

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("Origin", "https://calorieninjas.com")
	req.Header.Set("Referer", "https://calorieninjas.com/")
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("CalorieNinjas API responded with status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var recipes []CalorieNinjasRecipe
	if err := json.Unmarshal(body, &recipes); err != nil {
		return nil, err
	}

	if isPortugueseInput && len(recipes) > 0 {
		type translationResult struct {
			index       int
			title       string
			ingredients string
			instructs   string
			err         error
		}
		ch := make(chan translationResult, len(recipes))
		for i, r := range recipes {
			go func(idx int, rec CalorieNinjasRecipe) {
				t, err1 := p.translate(rec.Title, "en", "pt")
				ing, err2 := p.translate(rec.Ingredients, "en", "pt")
				ins, err3 := p.translate(rec.Instructions, "en", "pt")
				var tErr error
				if err1 != nil {
					tErr = err1
				} else if err2 != nil {
					tErr = err2
				} else if err3 != nil {
					tErr = err3
				}
				ch <- translationResult{index: idx, title: t, ingredients: ing, instructs: ins, err: tErr}
			}(i, r)
		}
		for i := 0; i < len(recipes); i++ {
			res := <-ch
			if res.err == nil {
				recipes[res.index].Title = res.title
				recipes[res.index].Ingredients = res.ingredients
				recipes[res.index].Instructions = res.instructs
			}
		}
	}

	return recipes, nil
}

func (p *CalorieNinjasProvider) translate(text, from, to string) (string, error) {
	if text == "" {
		return "", nil
	}
	u, err := url.Parse("https://api.mymemory.translated.net/get")
	if err != nil {
		return text, err
	}
	q := u.Query()
	q.Set("q", strings.TrimSpace(text))
	q.Set("langpair", from+"|"+to)
	q.Set("de", "apimacro@sampara.com")
	u.RawQuery = q.Encode()

	resp, err := p.client.Get(u.String())
	if err != nil {
		return text, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return text, fmt.Errorf("mymemory status error: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return text, err
	}

	var response struct {
		ResponseStatus int `json:"responseStatus"`
		ResponseData   struct {
			TranslatedText string `json:"translatedText"`
		} `json:"responseData"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return text, err
	}

	if response.ResponseStatus == 200 && response.ResponseData.TranslatedText != "" {
		return response.ResponseData.TranslatedText, nil
	}

	return text, nil
}

func (p *CalorieNinjasProvider) normalizeForLangDetection(text string) string {
	reg := regexp.MustCompile(`[\s\-_,.]`)
	return reg.ReplaceAllString(strings.ToLower(text), "")
}
