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
// e.g., "100g de arroz" -> "100", "g", "arroz"
// e.g., "1 pão" -> "1", "", "pão"
var foodRegex = regexp.MustCompile(`^(\d+(?:\.\d+)?)\s*([a-zA-Záàâãéèêíïóôõöúçñ\-]*)?\s+(?:de\s+)?(.*)$`)

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
	// Remove common connectors and filler words
	name = strings.TrimPrefix(name, "de ")
	name = strings.TrimSuffix(name, ".")
	
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

	// Handle parts separated by spaces, hyphens, or other common separators
	re := regexp.MustCompile(`([a-zA-Z0-9]+)`)
	parts := re.FindAllString(name, -1)
	for i, part := range parts {
		parts[i] = processWord(part)
	}
	
	return strings.Join(parts, " ")
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

	// First pass: look for exact name and unit match (standard scaling)
	for _, entry := range entries {
		c := m.Parse(entry.Description)
		
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

	// Second pass: if query has a unit (like "g") but cache entry has NO unit (just name)
	// we assume the cache entry is for 100g if unit is "g", or just 1 portion if query has no unit.
	// This handles "200g rice" matching "rice" entry.
	for _, entry := range entries {
		c := m.Parse(entry.Description)
		if q.Name == c.Name {
			// If cache has just "name" and query is "amount unit name"
			if c.Unit == "" && q.Unit == "g" {
				// We assume the cache entry was a standard 100g portion if it's "arroz",
				// but LLM usually returns per portion. 
				// Actually, many users expect "arroz" to mean 100g when matched with "g".
				// Let's assume the cache entry "arroz" represents 100g for scaling.
				ratio := q.Amount / 100.0
				return &models.FoodPreview{
					Description: query,
					Calories:    entry.Calories * ratio,
					Protein:     entry.Protein * ratio,
					Carbs:       entry.Carbs * ratio,
					Fat:         entry.Fat * ratio,
				}, nil
			}
			
			// If both have no unit, it's just a different amount of the same thing (e.g., 1 pão vs 2 pães)
			if c.Unit == "" && q.Unit == "" {
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
	}

	return nil, nil
}
