package services

import "calorie-tracker/models"

// LLMProvider defines the interface for LLM interactions
type LLMProvider interface {
	Call(model, prompt string) (string, error)
}

// FoodAnalyzer defines the interface for analyzing food descriptions
type FoodAnalyzer interface {
	ParseFood(description string) (*models.ReferenceFood, error)
}

// ReviewAnalyzer defines the interface for analyzing nutrition reviews
type ReviewAnalyzer interface {
	AnalyzeReview(data models.ReviewData) (*models.ReviewResult, error)
}
