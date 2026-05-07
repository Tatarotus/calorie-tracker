package services

import (
	"testing"

	"calorie-tracker/config"
)

func TestDesiredMetricAmount(t *testing.T) {
	tests := []struct {
		name     string
		item     ParsedFood
		expected float64
	}{
		{"zero amount", ParsedFood{Amount: 0, Unit: "gram"}, 0},
		{"gram", ParsedFood{Amount: 100, Unit: "gram"}, 100},
		{"ml", ParsedFood{Amount: 250, Unit: "ml"}, 250},
		{"tablespoon", ParsedFood{Amount: 2, Unit: "tablespoon"}, 30},
		{"teaspoon", ParsedFood{Amount: 3, Unit: "teaspoon"}, 15},
		{"cup", ParsedFood{Amount: 1, Unit: "cup"}, 240},
		{"bowl", ParsedFood{Amount: 1, Unit: "bowl"}, 250},
		{"plate", ParsedFood{Amount: 1, Unit: "plate"}, 350},
		{"serving", ParsedFood{Amount: 1, Unit: "serving"}, 100},
		{"slice", ParsedFood{Amount: 2, Unit: "slice"}, 60},
		{"handful", ParsedFood{Amount: 1, Unit: "handful"}, 28},
		{"unknown unit", ParsedFood{Amount: 100, Unit: "unknown"}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := desiredMetricAmount(tt.item)
			if got != tt.expected {
				t.Errorf("desiredMetricAmount() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestSelectFatSecretServing(t *testing.T) {
	tests := []struct {
		name        string
		servings    []fatSecretServing
		item        ParsedFood
		wantFound   bool
		wantServing string
	}{
		{
			name:        "empty servings",
			servings:    []fatSecretServing{},
			item:        ParsedFood{Name: "apple"},
			wantFound:   false,
			wantServing: "",
		},
		{
			name: "default serving",
			servings: []fatSecretServing{
				{ServingDescription: "100g", MetricServingAmount: "100", MetricServingUnit: "g", IsDefault: "1"},
			},
			item:        ParsedFood{Name: "apple", Amount: 0},
			wantFound:   true,
			wantServing: "100g",
		},
		{
			name: "closest metric match",
			servings: []fatSecretServing{
				{ServingDescription: "50g", MetricServingAmount: "50", MetricServingUnit: "g"},
				{ServingDescription: "100g", MetricServingAmount: "100", MetricServingUnit: "g"},
				{ServingDescription: "200g", MetricServingAmount: "200", MetricServingUnit: "g"},
			},
			item:        ParsedFood{Name: "apple", Amount: 90, Unit: "gram"},
			wantFound:   true,
			wantServing: "100g",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serving, found := selectFatSecretServing(tt.servings, tt.item)
			if found != tt.wantFound {
				t.Errorf("selectFatSecretServing() found = %v, want %v", found, tt.wantFound)
			}
			if tt.wantServing != "" && serving.ServingDescription != tt.wantServing {
				t.Errorf("selectFatSecretServing() serving = %v, want %v", serving.ServingDescription, tt.wantServing)
			}
		})
	}
}

func TestFatSecretServingReferenceFood(t *testing.T) {
	tests := []struct {
		name        string
		serving     fatSecretServing
		wantOK      bool
		wantName    string
		wantQty     float64
		wantUnit    string
		wantCal     float64
		wantProtein float64
	}{
		{
			name: "valid serving",
			serving: fatSecretServing{
				MetricServingAmount: "100",
				MetricServingUnit:   "g",
				Calories:            "52",
				Protein:             "0.3",
				Carbohydrate:        "14",
				Fat:                 "0.2",
			},
			wantOK:      true,
			wantName:    "apple",
			wantQty:     100,
			wantUnit:    "gram",
			wantCal:     52,
			wantProtein: 0.3,
		},
		{
			name: "invalid serving - all zeros",
			serving: fatSecretServing{
				MetricServingAmount: "0",
				MetricServingUnit:   "g",
				Calories:            "0",
				Protein:             "0",
				Carbohydrate:        "0",
				Fat:                 "0",
			},
			wantOK: false,
		},
		{
			name: "ml unit",
			serving: fatSecretServing{
				MetricServingAmount: "250",
				MetricServingUnit:   "ml",
				Calories:            "100",
				Protein:             "0",
				Carbohydrate:        "0",
				Fat:                 "0",
			},
			wantOK:   true,
			wantName: "milk",
			wantQty:  250,
			wantUnit: "ml",
			wantCal:  100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ref, ok := tt.serving.referenceFood(tt.wantName)
			if ok != tt.wantOK {
				t.Errorf("referenceFood() ok = %v, want %v", ok, tt.wantOK)
				return
			}
			if !tt.wantOK {
				return
			}
			if ref.Name != tt.wantName {
				t.Errorf("referenceFood() name = %v, want %v", ref.Name, tt.wantName)
			}
			if ref.BaseQuantity != tt.wantQty {
				t.Errorf("referenceFood() baseQuantity = %v, want %v", ref.BaseQuantity, tt.wantQty)
			}
			if ref.Unit != tt.wantUnit {
				t.Errorf("referenceFood() unit = %v, want %v", ref.Unit, tt.wantUnit)
			}
			if ref.Macros.Calories != tt.wantCal {
				t.Errorf("referenceFood() calories = %v, want %v", ref.Macros.Calories, tt.wantCal)
			}
			if ref.Macros.Protein != tt.wantProtein {
				t.Errorf("referenceFood() protein = %v, want %v", ref.Macros.Protein, tt.wantProtein)
			}
		})
	}
}

func TestParseFatSecretFloat(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"10", 10},
		{"10.5", 10.5},
		{"  10  ", 10},
		{"", 0},
		{"invalid", 0},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := parseFatSecretFloat(tt.input)
			if got != tt.expected {
				t.Errorf("parseFatSecretFloat(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestNewFatSecretProviderFromConfig(t *testing.T) {
	// Empty config should return nil
	p := NewFatSecretProviderFromConfig(nil)
	if p != nil {
		t.Error("expected nil provider for nil config")
	}

	// Config without credentials should return nil
	p = NewFatSecretProviderFromConfig(&config.Config{})
	if p != nil {
		t.Error("expected nil provider for config without credentials")
	}

	// Valid config should return provider
	cfg := &config.Config{
		FatSecretClientID:     "test-id",
		FatSecretClientSecret: "test-secret",
		FatSecretScope:        "basic",
		FatSecretRegion:       "BR",
		FatSecretLanguage:     "pt",
		FatSecretTokenURL:     "https://test.com/token",
		FatSecretAPIURL:       "https://test.com/api",
	}
	p = NewFatSecretProviderFromConfig(cfg)
	if p == nil {
		t.Error("expected non-nil provider for valid config")
	}
	if p.clientID != "test-id" {
		t.Errorf("expected clientID 'test-id', got %s", p.clientID)
	}
}
