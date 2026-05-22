package services

import (
	"calorie-tracker/models"
	"net/http"
	"time"
)

// SharedHTTPClient is a globally optimized, concurrency-safe HTTP client
// configured with connection pooling and keep-alives to eliminate TCP/TLS handshake latency.
var SharedHTTPClient = &http.Client{
	Timeout: 15 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 20,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 5 * time.Second,
	},
}

// LLMProvider defines the interface for LLM interactions
type LLMProvider interface {
	Call(model, prompt string) (string, error)
}

// FoodAnalyzer defines the interface for analyzing food descriptions
type FoodAnalyzer interface {
	ParseFood(description string) (*models.ReferenceFood, error)
}

type NutritionProvider interface {
	ResolveFood(item ParsedFood) (*models.ReferenceFood, error)
}

// ReviewAnalyzer defines the interface for analyzing nutrition reviews
type ReviewAnalyzer interface {
	AnalyzeReview(data models.ReviewData) (*models.ReviewResult, error)
}
