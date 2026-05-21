package services

import (
	"calorie-tracker/db"
	"calorie-tracker/models"
	"strings"
	"time"
)

// CanonicalResolverService handles alias mapping, multilingual resolution,
// and canonical entity database persistence.
type CanonicalResolverService struct {
	db            db.DBProvider
	synonymMapper *SynonymMapper
	normalizer    *Normalizer
}

// NewCanonicalResolverService creates a new CanonicalResolverService instance.
func NewCanonicalResolverService(db db.DBProvider, mapper *SynonymMapper, norm *Normalizer) *CanonicalResolverService {
	return &CanonicalResolverService{
		db:            db,
		synonymMapper: mapper,
		normalizer:    norm,
	}
}

// Resolve maps a raw name to a canonical database entity record, creating it if needed.
func (r *CanonicalResolverService) Resolve(name string) (*models.CanonicalFood, error) {
	// 1. Normalize the query
	normalizedName := r.normalizer.Normalize(name)

	// 2. Map using synonym mapper to a canonical key (e.g. "aipim cozido" -> "macaxeira")
	canonicalKey := r.synonymMapper.GetCanonical(normalizedName)

	// 3. Retrieve from SQLite
	f, err := r.db.GetCanonicalFoodByName(canonicalKey)
	if err != nil {
		return nil, err
	}
	if f != nil {
		return f, nil
	}

	// 4. If not found, persist a clean canonical entry
	now := time.Now()
	newFood := &models.CanonicalFood{
		CanonicalName:        canonicalKey,
		NormalizedName:       normalizedName,
		AliasesJSON:          `["` + name + `"]`,
		Language:             "en",
		Category:             "auto",
		DefaultServingAmount: 100,
		DefaultServingUnit:   "gram",
		DensityMultiplier:    1.0,
		GramsPerML:           1.0,
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	// Handle standard serving units/portions
	if strings.Contains(canonicalKey, "ovo") || strings.Contains(canonicalKey, "egg") || strings.Contains(canonicalKey, "banana") {
		newFood.DefaultServingAmount = 1.0
		newFood.DefaultServingUnit = "unit"
	}

	err = r.db.SaveCanonicalFood(newFood)
	if err != nil {
		return nil, err
	}

	return newFood, nil
}
