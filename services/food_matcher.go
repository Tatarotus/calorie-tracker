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
	db *db.DB
}

func NewFoodMatcher(db *db.DB) *FoodMatcher {
	return &FoodMatcher{db: db}
}

type ParsedFood struct {
	Amount float64
	Unit   string
	Name   string
}

// Regex to capture [amount][unit] [name]
// e.g., "100g de arroz" -> "100", "g", "de arroz"
// e.g., "1 pão" -> "1", "", "pão"
var foodRegex = regexp.MustCompile(`^(\d+(?:\.\d+)?)\s*([a-zA-Záàâãéèêíïóôõöúçñ\-]*)?(?:\s+de\s+|\s+)(.*)$`)

func (m *FoodMatcher) removeAccents(s string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	result, _, _ := transform.String(t, s)
	return result
}

func (m *FoodMatcher) Parse(desc string) ParsedFood {
	desc = strings.ToLower(strings.TrimSpace(desc))
	matches := foodRegex.FindStringSubmatch(desc)
	
	if len(matches) < 4 {
		// Try to see if it's just name
		return ParsedFood{
			Amount: 1,
			Unit:   "",
			Name:   m.normalizeName(desc),
		}
	}

	amount, _ := strconv.ParseFloat(matches[1], 64)
	unit := m.normalizeUnit(matches[2])
	name := m.normalizeName(matches[3])

	// If it looks like it was "1 ovo frito", unit="ovo", name="frito"
	// and we want to match "5 ovos fritos", unit="ovo", name="frito"
	// The current logic does this because normalizeUnit and normalizeName
	// both remove plurals and accents.

	return ParsedFood{
		Amount: amount,
		Unit:   unit,
		Name:   name,
	}
}

func (m *FoodMatcher) normalizeUnit(unit string) string {
	unit = m.removeAccents(strings.ToLower(strings.TrimSpace(unit)))
	switch unit {
	case "fatia", "fatias":
		return "fatia"
	case "unidade", "unidades", "un":
		return "unidade"
	case "ovo", "ovos":
		return "ovo"
	case "pao", "paes":
		return "pao"
	case "grama", "gramas", "g":
		return "g"
	case "mililitro", "mililitros", "ml":
		return "ml"
	case "copo", "copos":
		return "copo"
	case "colher", "colheres":
		return "colher"
	}
	
	// Default: if it ends with 's', try removing it
	if len(unit) > 1 && strings.HasSuffix(unit, "s") {
		return strings.TrimSuffix(unit, "s")
	}
	return unit
}

func (m *FoodMatcher) normalizeName(name string) string {
	name = m.removeAccents(strings.ToLower(strings.TrimSpace(name)))
	name = strings.TrimPrefix(name, "de ")
	
	// Split by whitespace and hyphens to normalize each part
	// We want to preserve hyphens, so we'll use a regex split or manual handling
	
	processWord := func(word string) string {
		if len(word) > 3 {
			if strings.HasSuffix(word, "os") {
				return strings.TrimSuffix(word, "s")
			} else if strings.HasSuffix(word, "as") {
				return strings.TrimSuffix(word, "s")
			} else if strings.HasSuffix(word, "oes") {
				return strings.TrimSuffix(word, "oes") + "ao"
			} else if strings.HasSuffix(word, "aes") {
				return strings.TrimSuffix(word, "aes") + "ao"
			}
		}
		return word
	}

	// Handle parts separated by spaces or hyphens
	re := regexp.MustCompile(`([a-zA-Z0-9]+)|([^a-zA-Z0-9]+)`)
	matches := re.FindAllString(name, -1)
	for i, part := range matches {
		if regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString(part) {
			matches[i] = processWord(part)
		}
	}
	
	return strings.Join(matches, "")
}

func (m *FoodMatcher) Match(query string) (*models.FoodPreview, error) {
	// 1. Try exact match first
	exact, err := m.db.GetCachedFood(query)
	if err == nil && exact != nil {
		return &models.FoodPreview{
			Description: exact.Description,
			Calories:    exact.Calories,
			Protein:     exact.Protein,
			Carbs:       exact.Carbs,
			Fat:         exact.Fat,
		}, nil
	}

	// 2. Parse query
	q := m.Parse(query)
	
	// 3. Get all cache entries and try to find a match
	entries, err := m.db.GetAllCacheEntries()
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		c := m.Parse(entry.Description)
		
		// If names match (already normalized in Parse) and units match, we can calculate
		if q.Name == c.Name && q.Unit == c.Unit {
			ratio := q.Amount / c.Amount
			return &models.FoodPreview{
				Description: query,
				Calories:    entry.Calories * ratio,
				Protein:     entry.Protein * ratio,
				Carbs:       entry.Carbs * ratio,
				Fat:         entry.Fat * ratio,
			}, nil
		}
	}

	return nil, nil
}
