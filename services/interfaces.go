package services

import "calorie-tracker/models"

// LLMProvider defines the interface for LLM interactions
// This allows us to mock the LLM for testing
type LLMProvider interface {
	Call(model, prompt string) (string, error)
}

// FoodParser defines the interface for parsing food descriptions
type FoodParser interface {
	ParseFood(description string) (*models.FoodPreview, error)
}

// ReviewAnalyzer defines the interface for analyzing nutrition reviews
type ReviewAnalyzer interface {
	AnalyzeReview(data models.ReviewData) (*models.ReviewResult, error)
}

// DBProvider defines the interface for database operations
// This will be used by TrackerService and FoodMatcher
type DBProvider interface {
	GetCachedFood(name string) (*models.FoodPreview, error)
	CacheFood(entry models.FoodEntry) error
}
