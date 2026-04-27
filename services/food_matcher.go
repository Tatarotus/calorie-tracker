package services

import (
	"calorie-tracker/db"
	"calorie-tracker/models"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

type FoodMatcher struct {
	db db.DBProvider
}

func NewFoodMatcher(db db.DBProvider) *FoodMatcher {
	return &FoodMatcher{db: db}
}

type ParsedFood struct {
	Amount float64
	Unit   string
	Name   string
}

// Regex to capture [amount][unit] [name]
// Requires whitespace before the name to avoid matching "100g" as amount+unit
var foodRegex = regexp.MustCompile(`^(\d+(?:\.\d+)?)\s*(tablespoons|tablespoon|teaspoons|teaspoon|ounces|ounce|pounds|pound|liters|liter|grams|gram|cups|cup|ml|gr|g)?\s+(.+)$`)

func (m *FoodMatcher) removeAccents(s string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	result, _, _ := transform.String(t, s)
	return result
}

func (m *FoodMatcher) Parse(desc string) ParsedFood {
	desc = strings.ToLower(strings.TrimSpace(desc))
	// Pre-normalize: remove some common words that might confuse the regex
	desc = strings.TrimPrefix(desc, "cerca de ")
	desc = strings.TrimPrefix(desc, "aproximadamente ")

	matches := foodRegex.FindStringSubmatch(desc)
	if len(matches) < 3 {
		// No amount found, try to parse as just a name
		return ParsedFood{
			Amount: 0,
			Unit:   "",
			Name:   desc,
		}
	}

	amount, _ := strconv.ParseFloat(matches[1], 64)
	unit := ""
	if len(matches) > 2 && matches[2] != "" {
		unit = m.normalizeUnit(matches[2])
	}
	name := ""
	if len(matches) > 3 {
		name = m.normalizeName(matches[3])
	}

	// If we matched a unit but no name, treat the whole input as a name
	// (e.g., "100g" should be treated as a name, not amount=100, unit=gram)
	if unit != "" && name == "" {
		return ParsedFood{
			Amount: 0,
			Unit:   "",
			Name:   desc,
		}
	}

	return ParsedFood{
		Amount: amount,
		Unit:   unit,
		Name:   name,
	}
}

func (m *FoodMatcher) normalizeUnit(unit string) string {
	unit = m.removeAccents(strings.ToLower(strings.TrimSpace(unit)))

	switch unit {
	case "cups":
		return "cup"
	case "tablespoons":
		return "tablespoon"
	case "teaspoons":
		return "teaspoon"
	case "grams", "g", "gr":
		return "gram"
	case "ounces":
		return "ounce"
	case "pounds":
		return "pound"
	case "liters":
		return "liter"
	default:
		return unit
	}
}

func (m *FoodMatcher) normalizeName(name string) string {
	name = m.removeAccents(strings.ToLower(strings.TrimSpace(name)))

	// Remove common filler words
	fillerWords := []string{"of", "the", "a", "an", "and", "or", "but", "in", "on", "at", "to", "for", "de"}
	words := strings.Fields(name)
	var filtered []string
	for _, word := range words {
		isFiller := false
		for _, filler := range fillerWords {
			if word == filler {
				isFiller = true
				break
			}
		}
		if !isFiller {
			filtered = append(filtered, word)
		}
	}

	return strings.Join(filtered, " ")
}

func (m *FoodMatcher) Match(description string) (*models.FoodPreview, error) {
	parsed := m.Parse(description)
	if parsed.Name == "" {
		return nil, nil
	}

	// Try to find in cache
	cached, err := m.db.GetCachedFood(parsed.Name)
	if err != nil {
		return nil, err
	}

	if cached != nil {
		return &models.FoodPreview{
			Description: cached.Description,
			Calories:    cached.Calories,
			Protein:     cached.Protein,
			Carbs:       cached.Carbs,
			Fat:         cached.Fat,
		}, nil
	}

	return nil, nil
}
