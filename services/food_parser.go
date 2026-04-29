package services

import (
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

type FoodParser struct{}

func NewFoodParser() *FoodParser {
	return &FoodParser{}
}

type ParsedFood struct {
	Amount float64
	Unit   string
	Name   string
}

// Regex to capture [amount][unit] [name]
// Requires whitespace before the name to avoid matching "100g" as amount+unit
var foodRegex = regexp.MustCompile(`^(\d+(?:\.\d+)?)\s*(tablespoons|tablespoon|teaspoons|teaspoon|ounces|ounce|pounds|pound|liters|liter|grams|gram|cups|cup|bowls|bowl|plates|plate|servings|serving|slices|slice|handfuls|handful|ml|gr|g|unidades|unidade|units|unit|unids|unid|u)?\s+(.+)$`)

var mealSplitter = regexp.MustCompile(`\s*(?:,|;|\+|\s+(?:and|with|e|com)\s+)\s*`)

func (p *FoodParser) removeAccents(s string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	result, _, _ := transform.String(t, s)
	return result
}

func (p *FoodParser) Parse(desc string) ParsedFood {
	desc = strings.ToLower(strings.TrimSpace(desc))
	desc = p.normalizeLeadingNumberWord(desc)
	// Pre-normalize: remove some common words that might confuse the regex
	desc = strings.TrimPrefix(desc, "cerca de ")
	desc = strings.TrimPrefix(desc, "aproximadamente ")
	desc = strings.TrimPrefix(desc, "i had ")
	desc = strings.TrimPrefix(desc, "i ate ")
	desc = strings.TrimPrefix(desc, "eu comi ")
	desc = strings.TrimPrefix(desc, "comi ")

	matches := foodRegex.FindStringSubmatch(desc)
	if len(matches) < 3 {
		// No amount found, try to parse as just a name
		return ParsedFood{
			Amount: 0,
			Unit:   "",
			Name:   p.normalizeName(desc),
		}
	}

	amount, _ := strconv.ParseFloat(matches[1], 64)
	unit := ""
	if len(matches) > 2 && matches[2] != "" {
		unit = p.normalizeUnit(matches[2])
	}
	name := ""
	if len(matches) > 3 {
		name = p.normalizeName(matches[3])
	}

	// If we matched a unit but no name, treat the whole input as a name
	if unit != "" && name == "" {
		return ParsedFood{
			Amount: 0,
			Unit:   "",
			Name:   p.normalizeName(desc),
		}
	}

	return ParsedFood{
		Amount: amount,
		Unit:   unit,
		Name:   name,
	}
}

func (p *FoodParser) ParseMeal(desc string) []ParsedFood {
	desc = strings.ToLower(strings.TrimSpace(desc))
	desc = strings.TrimPrefix(desc, "i had ")
	desc = strings.TrimPrefix(desc, "i ate ")
	desc = strings.TrimPrefix(desc, "eu comi ")
	desc = strings.TrimPrefix(desc, "comi ")

	parts := mealSplitter.Split(desc, -1)
	foods := make([]ParsedFood, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		parsed := p.Parse(part)
		if parsed.Name != "" {
			foods = append(foods, parsed)
		}
	}

	if len(foods) == 0 {
		parsed := p.Parse(desc)
		if parsed.Name != "" {
			foods = append(foods, parsed)
		}
	}

	return foods
}

func (p *FoodParser) normalizeUnit(unit string) string {
	unit = p.removeAccents(strings.ToLower(strings.TrimSpace(unit)))

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
	case "bowls":
		return "bowl"
	case "plates":
		return "plate"
	case "servings":
		return "serving"
	case "slices":
		return "slice"
	case "handfuls":
		return "handful"
	case "unidades", "unidade", "units", "unit", "unids", "unid", "u":
		return "unit"
	default:
		return unit
	}
}

func (p *FoodParser) normalizeName(name string) string {
	name = p.removeAccents(strings.ToLower(strings.TrimSpace(name)))

	// Remove common filler words
	fillerWords := []string{"of", "the", "a", "an", "some", "and", "or", "but", "in", "on", "at", "to", "for", "de", "da", "do"}
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

	return singularizeName(strings.Join(filtered, " "))
}

func (p *FoodParser) normalizeLeadingNumberWord(desc string) string {
	words := strings.Fields(desc)
	if len(words) == 0 {
		return desc
	}

	numberWords := map[string]string{
		"one": "1", "two": "2", "three": "3", "four": "4", "five": "5",
		"a": "1", "an": "1",
		"um": "1", "uma": "1", "dois": "2", "duas": "2", "tres": "3", "três": "3",
	}
	if value, ok := numberWords[words[0]]; ok {
		words[0] = value
		return strings.Join(words, " ")
	}

	return desc
}

func singularizeName(name string) string {
	replacements := map[string]string{
		"eggs":     "egg",
		"ovos":     "ovo",
		"bananas":  "banana",
		"slices":   "slice",
		"fatias":   "fatia",
		"tomatoes": "tomato",
		"potatoes": "potato",
		"macas":    "maca",
		"apples":   "apple",
	}

	words := strings.Fields(name)
	for i, word := range words {
		if replacement, ok := replacements[word]; ok {
			words[i] = replacement
		}
	}
	return strings.Join(words, " ")
}
