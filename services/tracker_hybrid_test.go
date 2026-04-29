package services

import (
	"calorie-tracker/config"
	"calorie-tracker/db"
	"calorie-tracker/models"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestNutritionEngine_HybridFlow(t *testing.T) {
	mockDB := db.NewMockDB()

	// Seed reference data
	mockDB.SeedReferenceFood(models.ReferenceFood{
		Name:         "arroz branco",
		BaseQuantity: 100,
		Unit:         "gram",
		Macros: models.Macros{
			Calories: 130,
			Protein:  2.7,
		},
	})

	t.Run("Priority 1: Reference DB match with scaling", func(t *testing.T) {
		engine := NewNutritionEngine(mockDB, nil)
		preview, err := engine.Analyze("200g arroz branco")
		if err != nil {
			t.Fatalf("Analyze failed: %v", err)
		}

		if preview.Calories != 260 {
			t.Errorf("Expected 260 calories (scaled), got %f", preview.Calories)
		}
	})

	t.Run("Natural language meal resolves without LLM", func(t *testing.T) {
		mockDB.SeedReferenceFood(models.ReferenceFood{
			Name:         "egg",
			BaseQuantity: 1,
			Unit:         "unit",
			Macros: models.Macros{
				Calories: 70,
				Protein:  6,
				Carbs:    0.6,
				Fat:      5,
			},
		})
		mockDB.SeedReferenceFood(models.ReferenceFood{
			Name:         "rice",
			BaseQuantity: 100,
			Unit:         "gram",
			Macros: models.Macros{
				Calories: 130,
				Protein:  2.7,
				Carbs:    28,
				Fat:      0.3,
			},
		})
		mockDB.SeedReferenceFood(models.ReferenceFood{
			Name:         "olive oil",
			BaseQuantity: 100,
			Unit:         "gram",
			Macros: models.Macros{
				Calories: 884,
				Fat:      100,
			},
		})

		engine := NewNutritionEngine(mockDB, nil)
		preview, err := engine.Analyze("I had two eggs and a bowl of rice with 1 tablespoon olive oil")
		if err != nil {
			t.Fatalf("Analyze failed: %v", err)
		}

		expectedCalories := 140.0 + 234.0 + 119.34
		if math.Abs(preview.Calories-expectedCalories) > 0.01 {
			t.Errorf("Expected %.2f calories, got %.2f", expectedCalories, preview.Calories)
		}
		if preview.Description != "2.0 egg + 1.0bowl rice + 1.0tablespoon olive oil" {
			t.Errorf("Expected combined description, got %q", preview.Description)
		}
	})

	t.Run("Priority 2: Cache match (Previous LLM result)", func(t *testing.T) {
		// Mock cache entry
		mockDB.CacheFood(models.ReferenceFood{
			Name:         "pao integral",
			BaseQuantity: 50,
			Unit:         "gram",
			Macros: models.Macros{
				Calories: 120,
			},
		})

		engine := NewNutritionEngine(mockDB, nil)
		preview, err := engine.Analyze("100g pao integral")
		if err != nil {
			t.Fatalf("Analyze failed: %v", err)
		}

		// 100g is 2x base of 50g
		if preview.Calories != 240 {
			t.Errorf("Expected 240 calories (scaled from cache), got %f", preview.Calories)
		}
	})

	t.Run("Explicit unit request skips incompatible gram cache", func(t *testing.T) {
		mockDB.CacheFood(models.ReferenceFood{
			Name:         "pao frances",
			BaseQuantity: 100,
			Unit:         "gram",
			Macros: models.Macros{
				Calories: 300,
			},
		})

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, `{"choices":[{"message":{"content":"{\"name\":\"pao frances\",\"base_quantity\":1,\"unit\":\"unit\",\"macros\":{\"calories\":135,\"protein\":4.5,\"carbs\":28,\"fat\":1.5}}"}}]}`)
		}))
		defer ts.Close()

		cfg := &config.Config{
			SambaAPIKey:   "test",
			OpenAIBaseURL: ts.URL,
			FoodModel:     "test",
		}
		llm := NewLLMServiceWithClient(cfg, ts.Client())
		engine := NewNutritionEngine(mockDB, llm)

		preview, err := engine.Analyze("1u pão francês")
		if err != nil {
			t.Fatalf("Analyze failed: %v", err)
		}

		if preview.Description != "1 pao frances" {
			t.Errorf("Expected unit description, got %q", preview.Description)
		}

		if preview.Calories != 135 {
			t.Errorf("Expected 135 calories from unit LLM data, got %f", preview.Calories)
		}
	})

	t.Run("Priority 3: LLM Fallback and caching", func(t *testing.T) {
		// Mock LLM server
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, `{"choices":[{"message":{"content":"{\"name\":\"fruta magica\",\"base_quantity\":100,\"unit\":\"g\",\"macros\":{\"calories\":50,\"protein\":1,\"carbs\":10,\"fat\":0}}"}}]}`)
		}))
		defer ts.Close()

		cfg := &config.Config{
			SambaAPIKey:   "test",
			OpenAIBaseURL: ts.URL,
			FoodModel:     "test",
		}
		llm := NewLLMServiceWithClient(cfg, ts.Client())
		engine := NewNutritionEngine(mockDB, llm)

		preview, err := engine.Analyze("200g fruta magica")
		if err != nil {
			t.Fatalf("Analyze failed: %v", err)
		}

		if preview.Calories != 100 {
			t.Errorf("Expected 100 calories (scaled from LLM base), got %f", preview.Calories)
		}

		// Verify it was cached as base ReferenceFood
		cached, _ := mockDB.GetCachedFood("fruta magica")
		if cached == nil {
			t.Fatal("Expected LLM result to be cached")
		}
		if cached.Macros.Calories != 50 {
			t.Errorf("Expected cached base calories to be 50, got %f", cached.Macros.Calories)
		}
	})

	t.Run("Priority 3: FatSecret provider resolves and caches before LLM", func(t *testing.T) {
		provider := stubNutritionProvider{
			ref: &models.ReferenceFood{
				Name:         "feijao carioca",
				BaseQuantity: 100,
				Unit:         "gram",
				Macros: models.Macros{
					Calories: 76,
					Protein:  4.8,
					Carbs:    13.6,
					Fat:      0.5,
				},
			},
		}

		engine := NewNutritionEngineWithProvider(mockDB, nil, provider)
		preview, err := engine.Analyze("200g feijao carioca")
		if err != nil {
			t.Fatalf("Analyze failed: %v", err)
		}

		if preview.Calories != 152 {
			t.Errorf("Expected 152 calories from provider, got %f", preview.Calories)
		}

		cached, _ := mockDB.GetCachedFood("feijao carioca")
		if cached == nil {
			t.Fatal("Expected provider result to be cached")
		}
	})

	t.Run("Reject unrealistic LLM data", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, `{"choices":[{"message":{"content":"{\"name\":\"bad\",\"base_quantity\":100,\"unit\":\"g\",\"macros\":{\"calories\":10000,\"protein\":0,\"carbs\":0,\"fat\":0}}"}}]}`)
		}))
		defer ts.Close()

		cfg := &config.Config{OpenAIBaseURL: ts.URL}
		llm := NewLLMServiceWithClient(cfg, ts.Client())
		engine := NewNutritionEngine(mockDB, llm)

		_, err := engine.Analyze("something")
		if err == nil {
			t.Error("Expected error for unrealistic calories")
		}
	})
}

type stubNutritionProvider struct {
	ref *models.ReferenceFood
	err error
}

func (s stubNutritionProvider) ResolveFood(ParsedFood) (*models.ReferenceFood, error) {
	return s.ref, s.err
}

func TestFatSecretProviderResolveFood(t *testing.T) {
	var tokenCalls int
	var apiCalls []url.Values

	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenCalls++
		if r.Method != http.MethodPost {
			t.Errorf("expected token POST, got %s", r.Method)
		}
		id, secret, ok := r.BasicAuth()
		if !ok || id != "client-id" || secret != "client-secret" {
			t.Errorf("unexpected basic auth id=%q ok=%v", id, ok)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}
		if r.Form.Get("scope") != "basic" {
			t.Errorf("expected basic scope, got %q", r.Form.Get("scope"))
		}
		fmt.Fprintln(w, `{"access_token":"test-token","token_type":"Bearer","expires_in":3600}`)
	}))
	defer tokenServer.Close()

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("unexpected authorization header %q", r.Header.Get("Authorization"))
		}
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}
		values := url.Values{}
		for key, value := range r.Form {
			values[key] = append([]string(nil), value...)
		}
		apiCalls = append(apiCalls, values)

		switch r.Form.Get("method") {
		case "foods.search":
			writeJSON(t, w, map[string]any{
				"foods": map[string]any{
					"food": map[string]string{"food_id": "123"},
				},
			})
		case "food.get.v2":
			writeJSON(t, w, map[string]any{
				"food": map[string]any{
					"food_id":   "123",
					"food_name": "Feijao Carioca",
					"servings": map[string]any{
						"serving": map[string]string{
							"metric_serving_amount": "100.000",
							"metric_serving_unit":   "g",
							"calories":              "76",
							"protein":               "4.8",
							"carbohydrate":          "13.6",
							"fat":                   "0.5",
							"is_default":            "1",
						},
					},
				},
			})
		default:
			t.Fatalf("unexpected method %q", r.Form.Get("method"))
		}
	}))
	defer apiServer.Close()

	provider := &FatSecretProvider{
		client:       apiServer.Client(),
		clientID:     "client-id",
		clientSecret: "client-secret",
		scope:        "basic",
		tokenURL:     tokenServer.URL,
		apiURL:       apiServer.URL,
	}

	ref, err := provider.ResolveFood(ParsedFood{Name: "feijao carioca", Amount: 200, Unit: "gram"})
	if err != nil {
		t.Fatalf("ResolveFood failed: %v", err)
	}
	if ref == nil {
		t.Fatal("expected resolved reference food")
	}
	if ref.Name != "feijao carioca" || ref.BaseQuantity != 100 || ref.Unit != "gram" {
		t.Fatalf("unexpected reference food: %#v", ref)
	}
	if ref.Macros.Calories != 76 || ref.Macros.Protein != 4.8 || ref.Macros.Carbs != 13.6 || ref.Macros.Fat != 0.5 {
		t.Fatalf("unexpected macros: %#v", ref.Macros)
	}
	if tokenCalls != 1 {
		t.Errorf("expected one token call, got %d", tokenCalls)
	}
	if len(apiCalls) != 2 {
		t.Fatalf("expected two api calls, got %d", len(apiCalls))
	}
	if apiCalls[0].Get("region") != "" || apiCalls[0].Get("language") != "" {
		t.Errorf("basic provider should not send localization by default: %v", apiCalls[0])
	}
}

func writeJSON(t *testing.T, w http.ResponseWriter, value any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(value); err != nil {
		t.Fatal(err)
	}
}
