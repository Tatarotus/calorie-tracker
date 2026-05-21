package services

import "calorie-tracker/models"

// NewResolutionTrace initializes a new traceability trace for ingestion audit logs.
func NewResolutionTrace(parserUsed, canonicalKey string) *models.ResolutionTrace {
	return &models.ResolutionTrace{
		ParserUsed:         parserUsed,
		CanonicalKey:       canonicalKey,
		ValidationWarnings: make([]string, 0),
	}
}
