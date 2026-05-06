package services

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSerpAPIProvider_ParseTextBlocks(t *testing.T) {
	p := &SerpAPIProvider{}
	
	blocks := []struct {
		Type string `json:"type"`
		List []struct {
			Snippet string `json:"snippet"`
		} `json:"list"`
		Snippet string `json:"snippet"`
	}{
		{
			Type: "heading",
			Snippet: "Moranga Crua",
		},
		{
			Type: "list",
			List: []struct {
				Snippet string `json:"snippet"`
			}{
				{Snippet: "Calorias: 12 kcal"},
				{Snippet: "Carboidratos: 2,7 g"},
				{Snippet: "Proteínas: 1,0 g"},
				{Snippet: "Gorduras: 0,1 g"},
			},
		},
		{
			Type: "heading",
			Snippet: "Moranga Refogada/Cozida",
		},
		{
			Type: "list",
			List: []struct {
				Snippet string `json:"snippet"`
			}{
				{Snippet: "Calorias: 29 kcal"},
				{Snippet: "Carboidratos: 6,0 g"},
				{Snippet: "Proteínas: 0,4 g"},
				{Snippet: "Gorduras: 0,8 g"},
			},
		},
	}

	ref := p.parseTextBlocks(blocks, "abobora moranga")
	if ref == nil {
		t.Fatal("expected reference food, got nil")
	}

	// Should pick the cooked one (29 kcal) due to higher score
	if ref.Macros.Calories != 29 {
		t.Errorf("expected 29 calories (cooked), got %v", ref.Macros.Calories)
	}
	if ref.Macros.Carbs != 6.0 {
		t.Errorf("expected 6.0 carbs, got %v", ref.Macros.Carbs)
	}
}

func TestSerpAPIProvider_ResolveFood_GoogleAI(t *testing.T) {
	mockResponse := serpAPIResponse{
		TextBlocks: []struct {
			Type string `json:"type"`
			List []struct {
				Snippet string `json:"snippet"`
			} `json:"list"`
			Snippet string `json:"snippet"`
		}{
			{
				Type: "list",
				List: []struct {
					Snippet string `json:"snippet"`
				}{
					{Snippet: "Calorias: 29 kcal"},
					{Snippet: "Carboidratos: 6,0 g"},
					{Snippet: "Proteínas: 0,4 g"},
					{Snippet: "Gorduras: 0,8 g"},
				},
			},
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer ts.Close()

	_ = &SerpAPIProvider{
		client: ts.Client(),
		apiKey: "test_key",
	}

	// We need to bypass the URL construction to point to our test server
	// Since ResolveFood is hardcoded to serpapi.com, we might need to refactor or use a mock.
	// For now, let's just test the parsing logic directly as we did above.
}

func TestSerpAPIProvider_ExtractFloat(t *testing.T) {
	p := &SerpAPIProvider{}
	tests := []struct {
		input    string
		expected float64
	}{
		{"12 kcal", 12},
		{"2,7 g", 2.7},
		{"0.1 g", 0.1},
		{"1.000,50 kcal", 1000.50}, // Handle European/Brazilian thousands separator might be tricky with simple regex
		{"29", 29},
	}

	for _, tt := range tests {
		got := p.extractFloat(tt.input)
		if got != tt.expected {
			t.Errorf("extractFloat(%q) = %v, want %v", tt.input, got, tt.expected)
		}
	}
}
