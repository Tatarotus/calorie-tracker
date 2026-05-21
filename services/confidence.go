package services

import (
	"calorie-tracker/models"
	"strings"
)

// ConfidenceCalculator aggregates multi-factor confidence scoring deterministically.
type ConfidenceCalculator struct{}

// NewConfidenceCalculator creates a new ConfidenceCalculator.
func NewConfidenceCalculator() *ConfidenceCalculator {
	return &ConfidenceCalculator{}
}

// Calculate computes the final unified confidence metric based on parser, canonicalisation,
// source reliability, and semantic validation outcomes.
func (c *ConfidenceCalculator) Calculate(trace *models.ResolutionTrace) float64 {
	if trace == nil {
		return 0.0
	}

	parserConfidence := 1.0
	if strings.Contains(trace.ParserUsed, "llm") {
		parserConfidence = 0.9
	}

	canonicalConfidence := 1.0
	if trace.ResolutionMethod == "fuzzy_fallback" {
		canonicalConfidence = 0.8
	}

	validationConfidence := 1.0
	if trace.ValidationTriggered && len(trace.ValidationWarnings) > 0 {
		validationConfidence = 0.8
	}
	if trace.ResolutionMethod == "unresolved_rejected" || trace.ResolutionMethod == "unresolved_fallback" {
		return 0.0
	}

	sourceConfidence := trace.SourceConfidence
	if sourceConfidence == 0 {
		sourceConfidence = 1.0 // default fallback
	}

	return parserConfidence * canonicalConfidence * sourceConfidence * validationConfidence
}
