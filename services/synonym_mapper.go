package services

import (
	"strings"
	"sync"
)

// SynonymMapper provides bidirectional synonym mapping for food names
type SynonymMapper struct {
	mu        sync.RWMutex
	canonical map[string]string   // maps normalized name -> canonical name
	groups    map[string][]string // canonical name -> list of synonyms
}

// NewSynonymMapper creates a new synonym mapper with default mappings
func NewSynonymMapper() *SynonymMapper {
	sm := &SynonymMapper{
		canonical: make(map[string]string),
		groups:    make(map[string][]string),
	}
	sm.loadDefaults()
	return sm
}

// loadDefaults loads the built-in synonym mappings
func (sm *SynonymMapper) loadDefaults() {
	// Grains & Starches
	sm.addGroup([]string{
		"arroz branco", "white rice",
		"arroz", "rice",
	})
	sm.addGroup([]string{
		"arroz integral", "brown rice",
	})
	sm.addGroup([]string{
		"macarrao", "pasta", "espaguete", "spaghetti",
		"noodle", "noodles",
	})
	sm.addGroup([]string{
		"pao", "bread",
		"pao frances", "french bread", "baguette",
	})
	sm.addGroup([]string{
		"batata", "potato",
		"batata frita", "french fries", "fries",
	})
	sm.addGroup([]string{
		"batata doce", "sweet potato",
	})

	// Proteins
	sm.addGroup([]string{
		"frango", "chicken",
		"frango grelhado", "grilled chicken",
		"peito de frango", "chicken breast",
	})
	sm.addGroup([]string{
		"carne", "beef", "meat",
		"carne bovina", "ground beef",
		"carne moida", "minced beef",
	})
	sm.addGroup([]string{
		"porco", "pork",
		"carne de porco", "pork meat",
	})
	sm.addGroup([]string{
		"peixe", "fish",
	})
	sm.addGroup([]string{
		"ovo", "egg",
		"ovos", "eggs",
	})
	sm.addGroup([]string{
		"atum", "tuna",
	})

	// Dairy
	sm.addGroup([]string{
		"leite", "milk",
	})
	sm.addGroup([]string{
		"queijo", "cheese",
	})
	sm.addGroup([]string{
		"manteiga", "butter",
	})
	sm.addGroup([]string{
		"iogurte", "yogurt", "yoghurt",
	})

	// Fruits
	sm.addGroup([]string{
		"banana", "banana",
	})
	sm.addGroup([]string{
		"maca", "apple",
	})
	sm.addGroup([]string{
		"laranja", "orange",
	})
	sm.addGroup([]string{
		"uva", "grape", "grapes",
	})
	sm.addGroup([]string{
		"morango", "strawberry", "strawberries",
	})
	sm.addGroup([]string{
		"abacate", "avocado",
	})
	sm.addGroup([]string{
		"manga", "mango",
	})

	// Vegetables
	sm.addGroup([]string{
		"tomate", "tomato",
	})
	sm.addGroup([]string{
		"cebola", "onion",
	})
	sm.addGroup([]string{
		"alho", "garlic",
	})
	sm.addGroup([]string{
		"cenoura", "carrot",
	})
	sm.addGroup([]string{
		"brocolis", "broccoli",
	})
	sm.addGroup([]string{
		"espinafre", "spinach",
	})
	sm.addGroup([]string{
		"alface", "lettuce",
	})
	sm.addGroup([]string{
		"pepino", "cucumber",
	})
	sm.addGroup([]string{
		"pimentao", "bell pepper", "pepper",
	})

	// Legumes
	sm.addGroup([]string{
		"feijao", "beans", "bean",
	})
	sm.addGroup([]string{
		"lentilha", "lentil", "lentils",
	})
	sm.addGroup([]string{
		"grao de bico", "chickpea", "chickpeas", "garbanzo",
	})

	// Fats & Oils
	sm.addGroup([]string{
		"azeite", "olive oil",
	})
	sm.addGroup([]string{
		"oleo de coco", "coconut oil",
	})

	// Beverages
	sm.addGroup([]string{
		"cafe", "coffee",
	})
	sm.addGroup([]string{
		"cha", "tea",
	})
	sm.addGroup([]string{
		"suco", "juice",
	})
	sm.addGroup([]string{
		"agua", "water",
	})
	sm.addGroup([]string{
		"refrigerante", "soda", "soft drink",
	})

	// Sweets & Desserts
	sm.addGroup([]string{
		"chocolate", "chocolate",
	})
	sm.addGroup([]string{
		"acucar", "sugar",
	})
	sm.addGroup([]string{
		"mel", "honey",
	})
	sm.addGroup([]string{
		"doce", "candy", "sweet",
	})

	// Nuts & Seeds
	sm.addGroup([]string{
		"amendoim", "peanut", "peanuts",
	})
	sm.addGroup([]string{
		"castanha", "cashew", "cashews",
	})
	sm.addGroup([]string{
		"noz", "walnut", "walnuts",
	})
	sm.addGroup([]string{
		"aveia", "oats", "oatmeal",
	})

	// Common dishes
	sm.addGroup([]string{
		"salada", "salad",
	})
	sm.addGroup([]string{
		"sopa", "soup",
	})
	sm.addGroup([]string{
		"sanduiche", "sandwich",
	})
	sm.addGroup([]string{
		"hamburguer", "hamburger", "burger",
	})
	sm.addGroup([]string{
		"pizza", "pizza",
	})
	sm.addGroup([]string{
		"taco", "taco",
	})
	sm.addGroup([]string{
		"sushi", "sushi",
	})
}

// addGroup adds a group of synonyms
func (sm *SynonymMapper) addGroup(synonyms []string) {
	if len(synonyms) == 0 {
		return
	}

	// Use the first synonym as the canonical form
	canonical := sm.normalize(synonyms[0])

	sm.mu.Lock()
	defer sm.mu.Unlock()

	for _, syn := range synonyms {
		normalized := sm.normalize(syn)
		sm.canonical[normalized] = canonical
	}
	sm.groups[canonical] = synonyms
}

// normalize normalizes a string for lookup
func (sm *SynonymMapper) normalize(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

// GetCanonical returns the canonical form of a food name, or the name itself if not found
func (sm *SynonymMapper) GetCanonical(name string) string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	normalized := sm.normalize(name)
	if canonical, ok := sm.canonical[normalized]; ok {
		return canonical
	}

	// Try without accents (basic normalization)
	deaccented := removeAccentsBasic(normalized)
	if canonical, ok := sm.canonical[deaccented]; ok {
		return canonical
	}

	return normalized
}

// IsSynonym returns true if two food names are synonyms
func (sm *SynonymMapper) IsSynonym(name1, name2 string) bool {
	return sm.GetCanonical(name1) == sm.GetCanonical(name2)
}

// GetSynonyms returns all synonyms for a given canonical name
func (sm *SynonymMapper) GetSynonyms(canonical string) []string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	normalized := sm.normalize(canonical)
	if group, ok := sm.groups[normalized]; ok {
		return append([]string(nil), group...)
	}

	// Try looking up via canonical mapping
	if actualCanonical, ok := sm.canonical[normalized]; ok {
		if group, ok := sm.groups[actualCanonical]; ok {
			return append([]string(nil), group...)
		}
	}

	return nil
}

// AddCustomSynonym adds a custom synonym mapping at runtime
func (sm *SynonymMapper) AddCustomSynonym(canonical string, synonyms ...string) {
	all := append([]string{canonical}, synonyms...)
	sm.addGroup(all)
}

// removeAccentsBasic removes common accents from a string (simplified version)
func removeAccentsBasic(s string) string {
	replacements := map[rune]rune{
		'á': 'a', 'à': 'a', 'â': 'a', 'ã': 'a', 'ä': 'a',
		'é': 'e', 'è': 'e', 'ê': 'e', 'ë': 'e',
		'í': 'i', 'ì': 'i', 'î': 'i', 'ï': 'i',
		'ó': 'o', 'ò': 'o', 'ô': 'o', 'õ': 'o', 'ö': 'o',
		'ú': 'u', 'ù': 'u', 'û': 'u', 'ü': 'u',
		'ç': 'c', 'ñ': 'n',
	}

	result := []rune(s)
	for i, r := range result {
		if replacement, ok := replacements[r]; ok {
			result[i] = replacement
		}
	}
	return string(result)
}
