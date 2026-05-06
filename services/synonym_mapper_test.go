package services

import (
	"testing"
)

func TestSynonymMapperGetCanonical(t *testing.T) {
	sm := NewSynonymMapper()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"exact match", "arroz branco", "arroz branco"},
		{"synonym", "white rice", "arroz branco"},
		{"case insensitive", "WHITE RICE", "arroz branco"},
		{"with spaces", " white rice ", "arroz branco"},
		{"unknown food", "unknown food xyz", "unknown food xyz"},
		{"portuguese egg", "ovo", "ovo"},
		{"english egg", "egg", "ovo"},
		{"portuguese plural egg", "ovos", "ovo"},
		{"chicken breast", "chicken breast", "frango"},
		{"peito de frango", "peito de frango", "frango"},
		{"coffee", "coffee", "cafe"},
		{"cafe", "cafe", "cafe"},
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

func TestSynonymMapperIsSynonym(t *testing.T) {
	sm := NewSynonymMapper()

	tests := []struct {
		name1    string
		name2    string
		expected bool
	}{
		{"arroz branco", "white rice", true},
		{"arroz", "rice", true},
		{"arroz branco", "arroz", true}, // Same group (arroz is in the arroz branco group)
		{"ovo", "egg", true},
		{"ovos", "eggs", true},
		{"banana", "banana", true},
		{"apple", "maca", true},
	}

	for _, tt := range tests {
		t.Run(tt.name1+"_"+tt.name2, func(t *testing.T) {
			got := sm.IsSynonym(tt.name1, tt.name2)
			if got != tt.expected {
				t.Errorf("IsSynonym(%q, %q) = %v, want %v", tt.name1, tt.name2, got, tt.expected)
			}
		})
	}
}

func TestSynonymMapperGetSynonyms(t *testing.T) {
	sm := NewSynonymMapper()

	syns := sm.GetSynonyms("arroz branco")
	if len(syns) == 0 {
		t.Error("Expected synonyms for 'arroz branco', got none")
	}

	found := false
	for _, s := range syns {
		if s == "white rice" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected 'white rice' in synonyms, got %v", syns)
	}
}

func TestSynonymMapperAddCustomSynonym(t *testing.T) {
	sm := NewSynonymMapper()

	// Add a custom synonym
	sm.AddCustomSynonym("quinoa", "quinua", "quinwa")

	if got := sm.GetCanonical("quinua"); got != "quinoa" {
		t.Errorf("GetCanonical('quinua') = %q, want 'quinoa'", got)
	}

	if got := sm.GetCanonical("quinwa"); got != "quinoa" {
		t.Errorf("GetCanonical('quinwa') = %q, want 'quinoa'", got)
	}

	if !sm.IsSynonym("quinua", "quinwa") {
		t.Error("Expected 'quinua' and 'quinwa' to be synonyms")
	}
}

func TestRemoveAccentsBasic(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"café", "cafe"},
		{"mãçã", "maca"},
		{"pão", "pao"},
		{"açaí", "acai"},
		{"normal", "normal"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := removeAccentsBasic(tt.input)
			if got != tt.expected {
				t.Errorf("removeAccentsBasic(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
