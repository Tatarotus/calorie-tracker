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
var foodRegex = regexp.MustCompile(`^(\d+(?:\.\d+)?)\s*(tablespoons|tablespoon|teaspoons|teaspoon|ounces|ounce|pounds|pound|liters|liter|grams|gram|cups|cup|ml|gr|g|unidades|unidade|unids|unid|u)?\s+(.+)$`)

func (p *FoodParser) removeAccents(s string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	result, _, _ := transform.String(t, s)
	return result
}

func (p *FoodParser) Parse(desc string) ParsedFood {
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
	case "unidades", "unids", "unid", "u":
		return "unit"
	default:
		return unit
	}
}

func (p *FoodParser) normalizeName(name string) string {
	name = p.removeAccents(strings.ToLower(strings.TrimSpace(name)))

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
