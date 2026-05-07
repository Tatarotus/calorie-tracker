package services

import (
	"testing"
)

func TestSynonymMapper_GetCanonical_Accented(t *testing.T) {
	sm := NewSynonymMapper()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"accented cafe", "café", "cafe"},
		{"accented pao", "pão", "pao"},
		{"accented maca", "maçã", "maca"},
		{"accented acai", "açaí", "açaí"}, // açaí is not in the default mappings, so it returns itself
		{"accented frango", "frangô", "frango"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sm.GetCanonical(tt.input)
			if got != tt.expected {
				t.Errorf("GetCanonical(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestSynonymMapper_GetSynonyms_Unknown(t *testing.T) {
	sm := NewSynonymMapper()

	// Unknown food should return nil
	syns := sm.GetSynonyms("unknown food xyz")
	if syns != nil {
		t.Errorf("Expected nil for unknown food, got %v", syns)
	}
}

func TestSynonymMapper_GetSynonyms_ViaCanonical(t *testing.T) {
	sm := NewSynonymMapper()

	// Get synonyms using a synonym name (not the canonical)
	syns := sm.GetSynonyms("white rice")
	if len(syns) == 0 {
		t.Error("Expected synonyms for 'white rice', got none")
	}

	found := false
	for _, s := range syns {
		if s == "arroz branco" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected 'arroz branco' in synonyms, got %v", syns)
	}
}

func TestSynonymMapper_AddCustomSynonym_Empty(t *testing.T) {
	sm := NewSynonymMapper()

	// Add empty group should not panic
	sm.AddCustomSynonym("test")

	if got := sm.GetCanonical("test"); got != "test" {
		t.Errorf("GetCanonical('test') = %q, want 'test'", got)
	}
}

func TestSynonymMapper_Normalize(t *testing.T) {
	sm := NewSynonymMapper()

	tests := []struct {
		input    string
		expected string
	}{
		{"  Test  ", "test"},
		{"UPPER", "upper"},
		{" Mixed Case ", "mixed case"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := sm.normalize(tt.input)
			if got != tt.expected {
				t.Errorf("normalize(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
