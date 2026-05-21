package services

import (
	"calorie-tracker/db"
	"calorie-tracker/models"
)

// FuzzyResolver encapsulates the reference database fuzzy lookup stage.
type FuzzyResolver struct {
	db            db.DBProvider
	fuzzyMatcher  *FuzzyMatcher
	synonymMapper *SynonymMapper
}

// NewFuzzyResolver creates a new FuzzyResolver.
func NewFuzzyResolver(db db.DBProvider, fuzzyMatcher *FuzzyMatcher, mapper *SynonymMapper) *FuzzyResolver {
	if mapper == nil {
		mapper = NewSynonymMapper()
	}
	return &FuzzyResolver{
		db:            db,
		fuzzyMatcher:  fuzzyMatcher,
		synonymMapper: mapper,
	}
}

// Find fuzzy-matches a food name against the reference database entries.
func (r *FuzzyResolver) Find(name string) (*models.ReferenceFood, error) {
	// 1. Check direct match first
	refFood, err := r.db.GetReferenceFood(name)
	if err == nil && refFood != nil {
		return refFood, nil
	}

	// 2. Map using SynonymMapper to find closest synonym (e.g. "chiken" -> "chicken")
	mappedName := name
	keys := r.synonymMapper.GetKeys()
	bestSynonym, score := r.fuzzyMatcher.FindBestMatch(name, keys)
	if score >= r.fuzzyMatcher.threshold && bestSynonym != "" {
		mappedName = bestSynonym
	}

	// 3. Get the canonical key
	canonicalKey := r.synonymMapper.GetCanonical(mappedName)

	// 4. Try all synonyms of this canonical key against the DB
	synonyms := r.synonymMapper.GetSynonyms(canonicalKey)
	if len(synonyms) == 0 {
		// fallback to standard list
		synonyms = []string{mappedName, canonicalKey}
	}

	for _, syn := range synonyms {
		f, err := r.db.GetReferenceFood(syn)
		if err == nil && f != nil {
			return f, nil
		}
	}

	// 5. Fallback to older fuzzy matching against hardcoded foods
	var candidates []string
	refMap := make(map[string]models.ReferenceFood)

	foods := []string{
		"arroz branco", "white rice", "frango grelhado", "grilled chicken",
		"peito de frango", "chicken breast", "ovo", "egg", "banana",
		"olive oil", "azeite", "butter", "manteiga", "bread", "pao",
		"leite", "milk", "cafe", "coffee", "macaxeira", "cassava", "aipim",
	}
	for _, fName := range foods {
		f, err := r.db.GetReferenceFood(fName)
		if err == nil && f != nil {
			candidates = append(candidates, f.Name)
			refMap[f.Name] = *f
		}
	}

	bestMatch, score := r.fuzzyMatcher.FindBestMatch(name, candidates)
	if score >= r.fuzzyMatcher.threshold && bestMatch != "" {
		if entry, ok := refMap[bestMatch]; ok {
			return &entry, nil
		}
	}
	return nil, nil
}
