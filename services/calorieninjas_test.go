package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func setupMockCalorieNinjas(t *testing.T) (*CalorieNinjasProvider, func()) {
	mymemoryServer := httptest.NewServer(http.HandlerFunc(mockMyMemoryHandler))
	cnServer := httptest.NewServer(http.HandlerFunc(mockCalorieNinjasHandler))

	provider := NewCalorieNinjasProvider()
	provider.client = &http.Client{
		Transport: &redirectTransport{
			roundTripper: http.DefaultTransport,
			mymemoryURL:  mymemoryServer.URL,
			cnURL:        cnServer.URL,
		},
	}

	cleanup := func() {
		mymemoryServer.Close()
		cnServer.Close()
	}
	return provider, cleanup
}

func TestCalorieNinjasProvider_Translate(t *testing.T) {
	provider, cleanup := setupMockCalorieNinjas(t)
	defer cleanup()

	// Test case 1
	res, err := provider.translate("pão de forma", "pt", "en")
	if err != nil {
		t.Fatalf("translate failed: %v", err)
	}
	if res != "sliced bread" {
		t.Errorf("expected 'sliced bread', got %q", res)
	}

	// Test case 2
	res, err = provider.translate("", "pt", "en")
	if err != nil {
		t.Fatalf("translate failed: %v", err)
	}
	if res != "" {
		t.Errorf("expected empty string, got %q", res)
	}

	// Test case 3
	res, err = provider.translate("error_trigger", "pt", "en")
	if err == nil {
		t.Error("expected error but got nil")
	}
	if res != "error_trigger" {
		t.Errorf("expected fallback to input text, got %q", res)
	}

	// Test case 4
	res, err = provider.translate("bad_json_trigger", "pt", "en")
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if res != "bad_json_trigger" {
		t.Errorf("expected fallback to input text, got %q", res)
	}
}

func TestCalorieNinjasProvider_Resolve_NilOrEmpty(t *testing.T) {
	provider, cleanup := setupMockCalorieNinjas(t)
	defer cleanup()

	var nilProvider *CalorieNinjasProvider
	ref, err := nilProvider.ResolveFood(ParsedFood{Name: "bread"})
	if ref != nil || err != nil {
		t.Errorf("expected nil ref and nil err, got ref=%v, err=%v", ref, err)
	}

	ref, err = provider.ResolveFood(ParsedFood{Name: ""})
	if ref != nil || err != nil {
		t.Errorf("expected nil ref and nil err, got ref=%v, err=%v", ref, err)
	}
}

func TestCalorieNinjasProvider_Resolve_EnglishQuery(t *testing.T) {
	provider, cleanup := setupMockCalorieNinjas(t)
	defer cleanup()

	ref, err := provider.ResolveFood(ParsedFood{Name: "bread"})
	if err != nil {
		t.Fatalf("ResolveFood failed: %v", err)
	}
	if ref == nil {
		t.Fatal("expected ReferenceFood, got nil")
	}
	if ref.Name != "bread" {
		t.Errorf("expected name 'bread', got %q", ref.Name)
	}
	if ref.Macros.Calories != 264.8 {
		t.Errorf("expected 264.8 calories, got %f", ref.Macros.Calories)
	}
}

func TestCalorieNinjasProvider_Resolve_PortugueseQuery(t *testing.T) {
	provider, cleanup := setupMockCalorieNinjas(t)
	defer cleanup()

	ref, err := provider.ResolveFood(ParsedFood{Name: "pão de forma"})
	if err != nil {
		t.Fatalf("ResolveFood failed: %v", err)
	}
	if ref == nil {
		t.Fatal("expected ReferenceFood, got nil")
	}
	if ref.Name != "pão" {
		t.Errorf("expected translated name 'pão', got %q", ref.Name)
	}
	if ref.Macros.Calories != 264.8 {
		t.Errorf("expected 264.8 calories, got %f", ref.Macros.Calories)
	}
}

func TestCalorieNinjasProvider_Resolve_Errors(t *testing.T) {
	provider, cleanup := setupMockCalorieNinjas(t)
	defer cleanup()

	ref, err := provider.ResolveFood(ParsedFood{Name: "error_trigger"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if ref != nil {
		t.Fatalf("expected nil ref, got %v", ref)
	}
	expectedErr := "CalorieNinjas API responded with status 400: invalid query parameters"
	if err.Error() != expectedErr {
		t.Errorf("expected error %q, got %q", expectedErr, err.Error())
	}

	ref, err = provider.ResolveFood(ParsedFood{Name: "other_error_trigger"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if ref != nil {
		t.Fatalf("expected nil ref, got %v", ref)
	}
	expectedErr = "CalorieNinjas API responded with status 403: access denied"
	if err.Error() != expectedErr {
		t.Errorf("expected error %q, got %q", expectedErr, err.Error())
	}

	ref, err = provider.ResolveFood(ParsedFood{Name: "empty_trigger"})
	if err != nil {
		t.Fatalf("ResolveFood failed: %v", err)
	}
	if ref != nil {
		t.Errorf("expected nil ref for empty items response, got %v", ref)
	}
}

func TestCalorieNinjasProvider_Recipes(t *testing.T) {
	provider, cleanup := setupMockCalorieNinjas(t)
	defer cleanup()

	// Recipe query
	recipes, err := provider.QueryRecipes("pão de forma")
	if err != nil {
		t.Fatalf("QueryRecipes failed: %v", err)
	}
	if len(recipes) == 0 {
		t.Fatal("expected recipes, got none")
	}
	if recipes[0].Title != "Simple Bread" {
		t.Errorf("expected recipe title 'Simple Bread', got %q", recipes[0].Title)
	}

	// Empty query error
	recipes, err = provider.QueryRecipes("")
	if err == nil {
		t.Error("expected error for empty query, got nil")
	}
	if recipes != nil {
		t.Errorf("expected nil recipes, got %v", recipes)
	}

	// Error trigger
	recipes, err = provider.QueryRecipes("error_trigger")
	if err == nil {
		t.Error("expected error, got nil")
	}
	if recipes != nil {
		t.Errorf("expected nil recipes, got %v", recipes)
	}
}

func TestCalorieNinjasProvider_ConnectionErrors(t *testing.T) {
	provider, cleanup := setupMockCalorieNinjas(t)
	defer cleanup()

	// Translate conn error
	res, err := provider.translate("conn_error_trigger", "pt", "en")
	if err == nil {
		t.Error("expected error but got nil")
	}
	if res != "conn_error_trigger" {
		t.Errorf("expected fallback to input text, got %q", res)
	}

	// ResolveFood conn error
	ref, err := provider.ResolveFood(ParsedFood{Name: "conn_error_trigger"})
	if err == nil {
		t.Error("expected error, got nil")
	}
	if ref != nil {
		t.Errorf("expected nil ref, got %v", ref)
	}

	// QueryRecipes conn error
	recipes, err := provider.QueryRecipes("conn_error_trigger")
	if err == nil {
		t.Error("expected error, got nil")
	}
	if recipes != nil {
		t.Errorf("expected nil recipes, got %v", recipes)
	}
}

type redirectTransport struct {
	roundTripper http.RoundTripper
	mymemoryURL  string
	cnURL        string
}

func (t *redirectTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.Query().Get("q") == "conn_error_trigger" || req.URL.Query().Get("query") == "conn_error_trigger" {
		return nil, fmt.Errorf("simulated connection failure")
	}
	if req.URL.Host == "api.mymemory.translated.net" {
		target, _ := http.NewRequest(req.Method, t.mymemoryURL+req.URL.Path+"?"+req.URL.RawQuery, req.Body)
		target.Header = req.Header
		return t.roundTripper.RoundTrip(target)
	}
	if req.URL.Host == "api.calorieninjas.com" {
		target, _ := http.NewRequest(req.Method, t.cnURL+req.URL.Path+"?"+req.URL.RawQuery, req.Body)
		target.Header = req.Header
		return t.roundTripper.RoundTrip(target)
	}
	return t.roundTripper.RoundTrip(req)
}

func mockMyMemoryHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	langpair := r.URL.Query().Get("langpair")

	if q == "error_trigger" {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if q == "bad_json_trigger" {
		_, _ = w.Write([]byte(`{"responseStatus":400,"responseData":{"translatedText":""}}`))
		return
	}
	if q == "pão de forma" && langpair == "pt|en" {
		_, _ = w.Write([]byte(`{"responseStatus":200,"responseData":{"translatedText":"sliced bread"}}`))
		return
	}
	if q == "bread" && langpair == "en|pt" {
		_, _ = w.Write([]byte(`{"responseStatus":200,"responseData":{"translatedText":"pão"}}`))
		return
	}
	// Fallbacks
	_, _ = w.Write([]byte(`{"responseStatus":200,"responseData":{"translatedText":"` + q + `"}}`))
}

func mockCalorieNinjasHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	if r.URL.Path == "/v1/nutrition" {
		if query == "error_trigger" {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"message":"invalid query parameters"}`))
			return
		}
		if query == "other_error_trigger" {
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(`access denied`))
			return
		}
		if query == "empty_trigger" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"items":[]}`))
			return
		}
		if query == "sliced bread" || query == "bread" {
			response := calorieNinjasResponse{
				Items: []calorieNinjasItem{
					{
						Name:                "bread",
						Calories:            264.8,
						ServingSizeG:        100,
						ProteinG:            9.0,
						CarbohydratesTotalG: 49.0,
						FatTotalG:           3.2,
					},
				},
			}
			_ = json.NewEncoder(w).Encode(response)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"items":[]}`))
		return
	}

	if r.URL.Path == "/v1/recipe" {
		if query == "error_trigger" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if query == "sliced bread" || query == "bread" {
			response := []CalorieNinjasRecipe{
				{
					Title:        "Simple Bread",
					Ingredients:  "flour, water, yeast",
					Servings:     "4",
					Instructions: "Mix and bake.",
				},
			}
			_ = json.NewEncoder(w).Encode(response)
			return
		}
	}
	w.WriteHeader(http.StatusNotFound)
}
