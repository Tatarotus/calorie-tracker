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
var foodRegex = regexp.MustCompile(`^(\d+(?:\.\d+)?)\s*(tablespoons|tablespoon|teaspoons|teaspoon|ounces|ounce|pounds|pound|liters|liter|grams|gram|cups|cup|copos|copo|bowls|bowl|plates|plate|servings|serving|slices|slice|fatias|fatia|handfuls|handful|ml|gr|g|unidades|unidade|units|unit|unids|unid|u|un)?\s+(.+)$`)

var mealSplitter = regexp.MustCompile(`\s*(?:,|;|\+|\s+(?:and|e)\s+)\s*`)

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

	initialParts := mealSplitter.Split(desc, -1)
	var parts []string
	for _, part := range initialParts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		// Split further on with/com only if followed by a quantity or unit
		subParts := p.splitPartByConnectives(part)
		parts = append(parts, subParts...)
	}

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

func (p *FoodParser) ParseMealForDisplay(desc string) []ParsedFood {
	desc = strings.ToLower(strings.TrimSpace(desc))
	desc = strings.TrimPrefix(desc, "i had ")
	desc = strings.TrimPrefix(desc, "i ate ")
	desc = strings.TrimPrefix(desc, "eu comi ")
	desc = strings.TrimPrefix(desc, "comi ")

	initialParts := mealSplitter.Split(desc, -1)
	foods := make([]ParsedFood, 0, len(initialParts))
	for _, part := range initialParts {
		for _, subPart := range p.splitPartByConnectives(strings.TrimSpace(part)) {
			parsed := p.parseForDisplay(subPart)
			if parsed.Name != "" {
				foods = append(foods, parsed)
			}
		}
	}

	if len(foods) == 0 {
		parsed := p.parseForDisplay(desc)
		if parsed.Name != "" {
			foods = append(foods, parsed)
		}
	}

	return foods
}

func (p *FoodParser) parseForDisplay(desc string) ParsedFood {
	desc = strings.ToLower(strings.TrimSpace(desc))
	desc = p.normalizeLeadingNumberWord(desc)
	desc = strings.TrimPrefix(desc, "cerca de ")
	desc = strings.TrimPrefix(desc, "aproximadamente ")

	matches := foodRegex.FindStringSubmatch(desc)
	if len(matches) < 3 {
		return ParsedFood{Name: p.normalizeDisplayName(desc)}
	}

	amount, _ := strconv.ParseFloat(matches[1], 64)
	unit := ""
	if len(matches) > 2 && matches[2] != "" {
		unit = p.normalizeUnit(matches[2])
	}
	name := ""
	if len(matches) > 3 {
		name = p.normalizeDisplayName(matches[3])
	}

	return ParsedFood{
		Amount: amount,
		Unit:   unit,
		Name:   name,
	}
}

func (p *FoodParser) splitPartByConnectives(part string) []string {
	words := strings.Fields(part)
	if len(words) == 0 {
		return nil
	}

	var results []string
	var currentWords []string

	for i := 0; i < len(words); i++ {
		word := words[i]
		if (word == "with" || word == "com") && i+1 < len(words) {
			if isQuantityOrUnit(words[i+1]) {
				if len(currentWords) > 0 {
					results = append(results, strings.Join(currentWords, " "))
					currentWords = nil
				}
				continue
			}
		}
		currentWords = append(currentWords, word)
	}

	if len(currentWords) > 0 {
		results = append(results, strings.Join(currentWords, " "))
	}

	return results
}

func isQuantityOrUnit(word string) bool {
	if word == "" {
		return false
	}
	// Check if starts with a digit
	if unicode.IsDigit(rune(word[0])) {
		return true
	}
	// Check standard quantity words (articles, numbers)
	quantities := map[string]bool{
		"a": true, "an": true, "one": true, "two": true, "three": true, "four": true, "five": true,
		"um": true, "uma": true, "dois": true, "duas": true, "tres": true, "três": true,
	}
	if quantities[word] {
		return true
	}
	// Check standard units
	units := map[string]bool{
		"tablespoon": true, "tablespoons": true, "teaspoon": true, "teaspoons": true,
		"ounce": true, "ounces": true, "pound": true, "pounds": true,
		"liter": true, "liters": true, "gram": true, "grams": true,
		"cup": true, "cups": true, "copo": true, "copos": true,
		"bowl": true, "bowls": true, "plate": true, "plates": true,
		"serving": true, "servings": true, "slice": true, "slices": true,
		"handful": true, "handfuls": true, "ml": true, "gr": true, "g": true,
		"unidade": true, "unidades": true, "unit": true, "units": true,
		"unid": true, "unids": true, "u": true, "un": true,
		// Portuguese units
		"colher": true, "colheres": true, "xicara": true, "xicaras": true,
		"fatia": true, "fatias": true, "prato": true, "pratos": true,
	}
	return units[word]
}

func (p *FoodParser) normalizeUnit(unit string) string {
	unit = p.removeAccents(strings.ToLower(strings.TrimSpace(unit)))

	switch unit {
	case "cups", "copos", "copo":
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
	case "slices", "fatias", "fatia":
		return "slice"
	case "handfuls":
		return "handful"
	case "unidades", "unidade", "units", "unit", "unids", "unid", "u", "un":
		return "unit"
	default:
		return unit
	}
}

func (p *FoodParser) normalizeName(name string) string {
	name = strings.ReplaceAll(name, "_", " ")
	name = p.removeAccents(strings.ToLower(strings.TrimSpace(name)))

	// Remove common filler words that are NOT part of a food name.
	fillerWords := []string{
		"of", "the", "a", "an", "some", "and", "or", "but", "in", "on", "at", "to", "for",
		"unit", "units", "unidade", "unidades", "unid", "u", "un",
		"gram", "grams", "grama", "gramas", "g",
		"milliliter", "milliliters", "ml",
		"liter", "liters", "litro", "litros", "l",
		"ounce", "ounces", "oz",
		"pound", "pounds", "lb", "lbs",
	}
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

	// Trim leading and trailing connective words
	connectives := []string{"de", "da", "do", "com", "e"}
	for {
		changed := false
		if len(filtered) > 0 {
			first := filtered[0]
			for _, conn := range connectives {
				if first == conn {
					filtered = filtered[1:]
					changed = true
					break
				}
			}
		}
		if len(filtered) > 0 {
			last := filtered[len(filtered)-1]
			for _, conn := range connectives {
				if last == conn {
					filtered = filtered[:len(filtered)-1]
					changed = true
					break
				}
			}
		}
		if !changed {
			break
		}
	}

	return singularizeName(strings.Join(filtered, " "))
}

func (p *FoodParser) normalizeDisplayName(name string) string {
	name = strings.ReplaceAll(name, "_", " ")
	name = strings.ToLower(strings.TrimSpace(name))

	fillerWords := []string{
		"of", "the", "a", "an", "some", "and", "or", "but", "in", "on", "at", "to", "for",
		"unit", "units", "unidade", "unidades", "unid", "u", "un",
		"gram", "grams", "grama", "gramas", "g",
		"milliliter", "milliliters", "ml",
		"liter", "liters", "litro", "litros", "l",
		"ounce", "ounces", "oz",
		"pound", "pounds", "lb", "lbs",
	}
	words := strings.Fields(name)
	filtered := make([]string, 0, len(words))
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

	connectives := []string{"de", "da", "do", "com", "e"}
	for {
		changed := false
		if len(filtered) > 0 {
			first := filtered[0]
			for _, conn := range connectives {
				if first == conn {
					filtered = filtered[1:]
					changed = true
					break
				}
			}
		}
		if len(filtered) > 0 {
			last := filtered[len(filtered)-1]
			for _, conn := range connectives {
				if last == conn {
					filtered = filtered[:len(filtered)-1]
					changed = true
					break
				}
			}
		}
		if !changed {
			break
		}
	}

	return singularizeDisplayName(strings.Join(filtered, " "))
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

func singularizeDisplayName(name string) string {
	replacements := map[string]string{
		"eggs":    "egg",
		"ovos":    "ovo",
		"bananas": "banana",
		"maçãs":   "maçã",
		"macas":   "maca",
		"apples":  "apple",
	}

	words := strings.Fields(name)
	for i, word := range words {
		if replacement, ok := replacements[word]; ok {
			words[i] = replacement
		}
	}
	return strings.Join(words, " ")
}

func normalizeUnit(unit string) string {
	p := FoodParser{}
	return p.normalizeUnit(unit)
}
