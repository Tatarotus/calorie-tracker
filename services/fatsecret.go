package services

import (
	"calorie-tracker/config"
	"calorie-tracker/models"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type FatSecretProvider struct {
	client          *http.Client
	clientID        string
	clientSecret    string
	scope           string
	region          string
	language        string
	useLocalization bool
	tokenURL        string
	apiURL          string
	accessToken     string
	tokenExpiresAt  time.Time
}

func NewFatSecretProviderFromConfig(cfg *config.Config) *FatSecretProvider {
	if cfg == nil || cfg.FatSecretClientID == "" || cfg.FatSecretClientSecret == "" {
		return nil
	}

	return &FatSecretProvider{
		client:          &http.Client{Timeout: 20 * time.Second},
		clientID:        cfg.FatSecretClientID,
		clientSecret:    cfg.FatSecretClientSecret,
		scope:           cfg.FatSecretScope,
		region:          cfg.FatSecretRegion,
		language:        cfg.FatSecretLanguage,
		useLocalization: cfg.FatSecretUseLocalization,
		tokenURL:        cfg.FatSecretTokenURL,
		apiURL:          cfg.FatSecretAPIURL,
	}
}

func (p *FatSecretProvider) ResolveFood(item ParsedFood) (*models.ReferenceFood, error) {
	if p == nil || item.Name == "" {
		return nil, nil
	}

	token, err := p.token()
	if err != nil {
		return nil, err
	}

	foodID, err := p.searchFoodID(token, item.Name)
	if err != nil || foodID == "" {
		return nil, err
	}

	food, err := p.getFood(token, foodID)
	if err != nil {
		return nil, err
	}

	serving, ok := selectFatSecretServing(food.Servings.Serving, item)
	if !ok {
		return nil, nil
	}

	ref, ok := serving.referenceFood(item.Name)
	if !ok {
		return nil, nil
	}

	return ref, nil
}

func (p *FatSecretProvider) token() (string, error) {
	if p.accessToken != "" && time.Now().Before(p.tokenExpiresAt) {
		return p.accessToken, nil
	}

	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	if p.scope != "" {
		form.Set("scope", p.scope)
	}

	req, err := http.NewRequest(http.MethodPost, p.tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.SetBasicAuth(p.clientID, p.clientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := p.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("fatsecret token error (status %d): %s", resp.StatusCode, string(body))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", err
	}
	if tokenResp.AccessToken == "" {
		return "", fmt.Errorf("fatsecret token response did not include access_token")
	}

	p.accessToken = tokenResp.AccessToken
	expiresIn := tokenResp.ExpiresIn
	if expiresIn <= 0 {
		expiresIn = 3600
	}
	p.tokenExpiresAt = time.Now().Add(time.Duration(expiresIn-60) * time.Second)
	return p.accessToken, nil
}

func (p *FatSecretProvider) searchFoodID(token, name string) (string, error) {
	values := url.Values{}
	values.Set("method", "foods.search")
	values.Set("search_expression", name)
	values.Set("max_results", "5")
	values.Set("format", "json")
	p.addLocalization(values)

	body, err := p.callAPI(token, values)
	if err != nil {
		return "", err
	}

	var result struct {
		Foods struct {
			Food json.RawMessage `json:"food"`
		} `json:"foods"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	foods := parseFatSecretFoodList(result.Foods.Food)
	if len(foods) == 0 {
		return "", nil
	}
	return foods[0].FoodID, nil
}

func (p *FatSecretProvider) getFood(token, foodID string) (*fatSecretFood, error) {
	values := url.Values{}
	values.Set("method", "food.get.v2")
	values.Set("food_id", foodID)
	values.Set("format", "json")
	p.addLocalization(values)

	body, err := p.callAPI(token, values)
	if err != nil {
		return nil, err
	}

	var result struct {
		Food fatSecretFood `json:"food"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	if result.Food.FoodID == "" {
		return nil, nil
	}
	return &result.Food, nil
}

func (p *FatSecretProvider) callAPI(token string, values url.Values) ([]byte, error) {
	req, err := http.NewRequest(http.MethodPost, p.apiURL, strings.NewReader(values.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fatsecret api error (status %d): %s", resp.StatusCode, string(body))
	}
	return body, nil
}

func (p *FatSecretProvider) addLocalization(values url.Values) {
	if !p.useLocalization {
		return
	}
	if p.region != "" {
		values.Set("region", p.region)
	}
	if p.language != "" {
		values.Set("language", p.language)
	}
}

type fatSecretFoodSummary struct {
	FoodID string `json:"food_id"`
}

type fatSecretFood struct {
	FoodID   string            `json:"food_id"`
	FoodName string            `json:"food_name"`
	Servings fatSecretServings `json:"servings"`
}

type fatSecretServings struct {
	Serving fatSecretServingList `json:"serving"`
}

type fatSecretServingList []fatSecretServing

type fatSecretServing struct {
	ServingDescription  string `json:"serving_description"`
	MetricServingAmount string `json:"metric_serving_amount"`
	MetricServingUnit   string `json:"metric_serving_unit"`
	Calories            string `json:"calories"`
	Carbohydrate        string `json:"carbohydrate"`
	Protein             string `json:"protein"`
	Fat                 string `json:"fat"`
	IsDefault           string `json:"is_default"`
}

func (l *fatSecretServingList) UnmarshalJSON(data []byte) error {
	var one fatSecretServing
	if err := json.Unmarshal(data, &one); err == nil && one.Calories != "" {
		*l = []fatSecretServing{one}
		return nil
	}

	var many []fatSecretServing
	if err := json.Unmarshal(data, &many); err != nil {
		return err
	}
	*l = many
	return nil
}

func parseFatSecretFoodList(data json.RawMessage) []fatSecretFoodSummary {
	if len(data) == 0 {
		return nil
	}

	var one fatSecretFoodSummary
	if err := json.Unmarshal(data, &one); err == nil && one.FoodID != "" {
		return []fatSecretFoodSummary{one}
	}

	var many []fatSecretFoodSummary
	if err := json.Unmarshal(data, &many); err != nil {
		return nil
	}
	return many
}

func selectFatSecretServing(servings []fatSecretServing, item ParsedFood) (fatSecretServing, bool) {
	if len(servings) == 0 {
		return fatSecretServing{}, false
	}

	desired := desiredMetricAmount(item)
	if desired <= 0 {
		for _, serving := range servings {
			if serving.IsDefault == "1" {
				return serving, true
			}
		}
		return servings[0], true
	}

	var best fatSecretServing
	bestDistance := math.MaxFloat64
	for _, serving := range servings {
		amount := parseFatSecretFloat(serving.MetricServingAmount)
		unit := strings.ToLower(serving.MetricServingUnit)
		if amount <= 0 || (unit != "g" && unit != "ml") {
			continue
		}
		distance := math.Abs(amount - desired)
		if distance < bestDistance {
			best = serving
			bestDistance = distance
		}
	}

	return best, bestDistance < math.MaxFloat64
}

func desiredMetricAmount(item ParsedFood) float64 {
	if item.Amount <= 0 {
		return 0
	}

	switch item.Unit {
	case "gram", "ml":
		return item.Amount
	case "tablespoon":
		return item.Amount * 15
	case "teaspoon":
		return item.Amount * 5
	case "cup":
		return item.Amount * 240
	case "bowl":
		return item.Amount * 250
	case "plate":
		return item.Amount * 350
	case "serving":
		return item.Amount * 100
	case "slice":
		return item.Amount * 30
	case "handful":
		return item.Amount * 28
	default:
		return 0
	}
}

func (s fatSecretServing) referenceFood(name string) (*models.ReferenceFood, bool) {
	baseQuantity := parseFatSecretFloat(s.MetricServingAmount)
	unit := strings.ToLower(s.MetricServingUnit)
	if baseQuantity <= 0 || (unit != "g" && unit != "ml") {
		baseQuantity = 1
		unit = "unit"
	}
	if unit == "g" {
		unit = "gram"
	}

	calories := parseFatSecretFloat(s.Calories)
	protein := parseFatSecretFloat(s.Protein)
	carbs := parseFatSecretFloat(s.Carbohydrate)
	fat := parseFatSecretFloat(s.Fat)
	if calories <= 0 && protein <= 0 && carbs <= 0 && fat <= 0 {
		return nil, false
	}

	return &models.ReferenceFood{
		Name:         name,
		BaseQuantity: baseQuantity,
		Unit:         unit,
		Macros: models.Macros{
			Calories: calories,
			Protein:  protein,
			Carbs:    carbs,
			Fat:      fat,
		},
	}, true
}

func parseFatSecretFloat(value string) float64 {
	n, _ := strconv.ParseFloat(strings.TrimSpace(value), 64)
	return n
}
