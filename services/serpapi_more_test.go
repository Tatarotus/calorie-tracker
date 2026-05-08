package services

import (
	"testing"
)

func TestSerpAPIProvider_ExtractValue(t *testing.T) {
	p := &SerpAPIProvider{}

	tests := []struct {
		input    interface{}
		expected float64
	}{
		{"10 g", 10},
		{"5.5g", 5.5},
		{"", 0},
		{[]interface{}{"15 g"}, 15},
		{[]interface{}{}, 0},
		{nil, 0},
	}

	for _, tt := range tests {
		got := p.extractValue(tt.input)
		if got != tt.expected {
			t.Errorf("extractValue(%v) = %v, want %v", tt.input, got, tt.expected)
		}
	}
}

func TestSerpAPIProvider_ParseAmountPer(t *testing.T) {
	p := &SerpAPIProvider{}

	tests := []struct {
		input        string
		expectedQty  float64
		expectedUnit string
	}{
		{"", 100, "gram"},
		{"1 cup (248 g)", 248, "gram"},
		{"100 g", 100, "gram"},
		{"50g", 50, "gram"},
	}

	for _, tt := range tests {
		qty, unit := p.parseAmountPer(tt.input)
		if qty != tt.expectedQty || unit != tt.expectedUnit {
			t.Errorf("parseAmountPer(%q) = (%v, %v), want (%v, %v)",
				tt.input, qty, unit, tt.expectedQty, tt.expectedUnit)
		}
	}
}

func TestSerpAPIProvider_ExtractMacrosFromList(t *testing.T) {
	p := &SerpAPIProvider{}

	list := []struct {
		Snippet string `json:"snippet"`
	}{
		{Snippet: "Calorias: 100 kcal"},
		{Snippet: "Carboidratos: 20 g"},
		{Snippet: "Proteínas: 5 g"},
		{Snippet: "Gorduras: 2 g"},
	}

	macros, count := p.extractMacrosFromList(list)
	if count != 4 {
		t.Errorf("expected count 4, got %d", count)
	}
	if macros.Calories != 100 {
		t.Errorf("expected calories 100, got %v", macros.Calories)
	}
	if macros.Carbs != 20 {
		t.Errorf("expected carbs 20, got %v", macros.Carbs)
	}
	if macros.Protein != 5 {
		t.Errorf("expected protein 5, got %v", macros.Protein)
	}
	if macros.Fat != 2 {
		t.Errorf("expected fat 2, got %v", macros.Fat)
	}
}

func TestSerpAPIProvider_ExtractMacrosFromList_Empty(t *testing.T) {
	p := &SerpAPIProvider{}

	list := []struct {
		Snippet string `json:"snippet"`
	}{}

	macros, count := p.extractMacrosFromList(list)
	if count != 0 {
		t.Errorf("expected count 0, got %d", count)
	}
	if macros.Calories != 0 {
		t.Errorf("expected calories 0, got %v", macros.Calories)
	}
}

func TestSerpAPIProvider_NewSerpAPIProvider(t *testing.T) {
	// Empty API key should return nil
	p := NewSerpAPIProvider("")
	if p != nil {
		t.Error("expected nil provider for empty API key")
	}

	// Valid API key should return provider
	p = NewSerpAPIProvider("test-key")
	if p == nil {
		t.Fatalf("expected non-nil provider for valid API key")
	}
	if p.apiKey != "test-key" {
		t.Errorf("expected apiKey 'test-key', got %s", p.apiKey)
	}
}

func TestSerpAPIProvider_ResolveFood_NilProvider(t *testing.T) {
	var p *SerpAPIProvider
	result, err := p.ResolveFood(ParsedFood{Name: "apple"})
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result != nil {
		t.Error("expected nil result for nil provider")
	}
}

func TestSerpAPIProvider_ResolveFood_EmptyName(t *testing.T) {
	p := &SerpAPIProvider{apiKey: "test"}
	result, err := p.ResolveFood(ParsedFood{Name: ""})
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result != nil {
		t.Error("expected nil result for empty name")
	}
}
