package services

import (
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var decimalCommaRegex = regexp.MustCompile(`(\d+),(\d+)`)
var multipleSpacesRegex = regexp.MustCompile(`\s+`)

// Normalizer provides a dedicated layer to unify input pre-processing.
type Normalizer struct{}

// NewNormalizer creates a new Normalizer instance.
func NewNormalizer() *Normalizer {
	return &Normalizer{}
}

// Normalize processes a raw query string into a standardized, clean format.
func (n *Normalizer) Normalize(s string) string {
	// 1. Lowercase
	s = strings.ToLower(s)

	// 2. Normalize decimal separators: e.g. "2,5g" -> "2.5g"
	s = decimalCommaRegex.ReplaceAllString(s, "$1.$2")

	// 3. Remove accents
	s = n.RemoveAccents(s)

	// 4. Remove punctuation, keeping only letters, digits, dots, and spaces.
	// For decimals, we keep the dot. We also keep spaces to separate words.
	var sb strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '.' || unicode.IsSpace(r) {
			sb.WriteRune(r)
		} else {
			sb.WriteRune(' ') // Replace other punctuation with a space
		}
	}
	s = sb.String()

	// 5. Normalize whitespace
	s = multipleSpacesRegex.ReplaceAllString(s, " ")
	s = strings.TrimSpace(s)

	// 6. Normalize pluralization (singularization)
	s = n.Singularize(s)

	return s
}

// RemoveAccents removes diacritics (accents) from characters.
func (n *Normalizer) RemoveAccents(s string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	result, _, _ := transform.String(t, s)
	return result
}

// Singularize maps common plural terms back to their singular counterparts.
func (n *Normalizer) Singularize(s string) string {
	replacements := map[string]string{
		"eggs":        "egg",
		"ovos":        "ovo",
		"bananas":     "banana",
		"slices":      "slice",
		"fatias":      "fatia",
		"tomatoes":    "tomato",
		"potatoes":    "potato",
		"macas":       "maca",
		"apples":      "apple",
		"unidades":    "unidade",
		"units":       "unit",
		"servings":    "serving",
		"gramas":      "grama",
		"milliliters": "milliliter",
		"litros":      "litro",
		"copos":       "copo",
	}

	words := strings.Fields(s)
	for i, word := range words {
		if replacement, ok := replacements[word]; ok {
			words[i] = replacement
		}
	}
	return strings.Join(words, " ")
}
