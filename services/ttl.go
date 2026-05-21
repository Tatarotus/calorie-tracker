package services

import (
	"time"

	"calorie-tracker/models"
)

// TTLPolicy manages the dynamic cache confidence decay strategy.
type TTLPolicy struct{}

// NewTTLPolicy creates a new TTLPolicy.
func NewTTLPolicy() *TTLPolicy {
	return &TTLPolicy{}
}

// IsExpired determines if a cache entry is expired based on its source type and confidence.
func (p *TTLPolicy) IsExpired(entry *models.NutritionCacheEntry) bool {
	if entry == nil {
		return true
	}

	var ttl time.Duration

	switch entry.SourceType {
	case "user_override":
		// Permanent override
		return false
	case "fatsecret":
		if entry.SourceConfidence >= 1.0 {
			// fatsecret_exact: 30d
			ttl = 30 * 24 * time.Hour
		} else {
			// fatsecret_fuzzy: 14d
			ttl = 14 * 24 * time.Hour
		}
	case "serpapi_fallback", "serpapi":
		// serpapi_fallback: 3d
		ttl = 3 * 24 * time.Hour
	case "fuzzy_fallback", "reference_db":
		// fuzzy_fallback: 3d
		ttl = 3 * 24 * time.Hour
	default:
		// Default conservative TTL for low-confidence data
		if entry.SourceConfidence >= 0.9 {
			ttl = 14 * 24 * time.Hour
		} else {
			ttl = 3 * 24 * time.Hour
		}
	}

	return time.Since(entry.UpdatedAt) > ttl
}
