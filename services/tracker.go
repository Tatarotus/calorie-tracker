package services

import (
	"encoding/json"
	"strings"
	"time"

	"calorie-tracker/config"
	"calorie-tracker/db"
	"calorie-tracker/models"
)

type TrackerService struct {
	db     db.DBProvider
	llm    *LLMService
	engine *NutritionEngine
}

func NewTrackerService(db db.DBProvider, llm *LLMService) *TrackerService {
	cfg := config.Load()
	priority := strings.Split(cfg.NutritionPriority, ",")
	providers := make([]NutritionProvider, 0)

	for _, p := range priority {
		p = strings.TrimSpace(p)
		switch p {
		case "fatsecret":
			if fs := NewFatSecretProviderFromConfig(cfg); fs != nil {
				providers = append(providers, fs)
			}
		case "serpapi":
			if serp := NewSerpAPIProvider(cfg.SerpAPIKey); serp != nil {
				providers = append(providers, serp)
			}
		}
	}

	return &TrackerService{
		db:     db,
		llm:    llm,
		engine: NewNutritionEngineWithProviders(db, llm, providers),
	}
}

func (s *TrackerService) ParseFood(description string) (*models.FoodPreview, error) {
	return s.engine.Analyze(description)
}

func (s *TrackerService) SaveFood(preview *models.FoodPreview) error {
	var originalQuery string
	var normalizedQuery string
	var canonicalKey string
	var traceJSON string

	if preview.ResolutionTrace != nil {
		canonicalKey = preview.ResolutionTrace.CanonicalKey
		if b, err := json.Marshal(preview.ResolutionTrace); err == nil {
			traceJSON = string(b)
		}
	}

	name, unit, amount := s.extractNameUnitAmount(preview)

	// Lineage logging parameters
	originalQuery = preview.Description
	normalizedQuery = strings.ToLower(strings.TrimSpace(name))
	if normalizedQuery == "" {
		normalizedQuery = strings.ToLower(strings.TrimSpace(preview.Description))
	}

	entry := models.FoodEntry{
		Timestamp:       time.Now(),
		Description:     preview.Description,
		Calories:        preview.Calories,
		Protein:         preview.Protein,
		Carbs:           preview.Carbs,
		Fat:             preview.Fat,
		OriginalQuery:   originalQuery,
		NormalizedQuery: normalizedQuery,
		CanonicalKey:    canonicalKey,
		ResolutionTrace: traceJSON,
	}

	if err := s.db.AddFoodEntry(entry); err != nil {
		return err
	}

	// Resolve the CanonicalFood entry for the override
	var canonicalID int64
	if name != "" {
		cf, err := s.engine.canonicalResolver.Resolve(name)
		if err == nil && cf != nil {
			canonicalID = cf.ID
		}
	}

	if name != "" && amount > 0 {
		baseQuantity := 100.0
		if unit != "gram" && unit != "" {
			baseQuantity = 1.0
		}

		factor := baseQuantity / amount
		refMacros := models.Macros{
			Calories: preview.Calories * factor,
			Protein:  preview.Protein * factor,
			Carbs:    preview.Carbs * factor,
			Fat:      preview.Fat * factor,
		}

		cacheEntry := models.ReferenceFood{
			Name:         name,
			BaseQuantity: baseQuantity,
			Unit:         unit,
			Macros:       refMacros,
		}
		if shouldCacheSavedPreview(preview) {
			_ = s.db.CacheFood(cacheEntry)
		}

		// If the preview was edited, persist a new entry in the user_overrides database table
		if preview.UserEdited && canonicalID != 0 {
			override := &models.UserOverrideEntry{
				CanonicalFoodID: canonicalID,
				ServingAmount:   baseQuantity,
				ServingUnit:     unit,
				Calories:        refMacros.Calories,
				Protein:         refMacros.Protein,
				Carbs:           refMacros.Carbs,
				Fat:             refMacros.Fat,
				OverrideReason:  "user_edited",
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			}
			_ = s.db.SaveUserOverride(override)
		}
	}

	return nil
}

func shouldCacheSavedPreview(preview *models.FoodPreview) bool {
	if preview == nil {
		return false
	}
	if preview.UserEdited {
		return true
	}
	if preview.ResolutionTrace == nil {
		return true
	}
	switch preview.ResolutionTrace.SourceType {
	case "fatsecret", "serpapi_fallback", "user_override":
		return true
	default:
		return false
	}
}

func (s *TrackerService) extractNameUnitAmount(preview *models.FoodPreview) (string, string, float64) {
	name := preview.Name
	unit := preview.Unit
	amount := 0.0

	var parsedItems []ParsedFoodItem
	var parseErr error
	if preview.Description != "" {
		parsedItems, parseErr = s.engine.parser.Parse(preview.Description)
	}

	if parseErr == nil && len(parsedItems) > 0 {
		amount = parsedItems[0].Quantity
		if name == "" {
			name = parsedItems[0].FoodName
		}
		if unit == "" {
			unit = parsedItems[0].Unit
		}
	} else {
		// Fallback to basic regex parser
		bp := NewFoodParser()
		parsed := bp.Parse(preview.Description)
		amount = parsed.Amount
		if name == "" {
			name = parsed.Name
		}
		if unit == "" {
			unit = parsed.Unit
		}
	}
	return name, unit, amount
}

func (s *TrackerService) AddWater(amountML float64) error {
	entry := models.WaterEntry{
		Timestamp: time.Now(),
		AmountML:  amountML,
	}
	return s.db.AddWaterEntry(entry)
}

func (s *TrackerService) GetDailyStats(t time.Time) (models.DailyStats, error) {
	food, err := s.db.GetDailyFoodEntries(t)
	if err != nil {
		return models.DailyStats{}, err
	}
	water, err := s.db.GetDailyWaterEntries(t)
	if err != nil {
		return models.DailyStats{}, err
	}
	stats := models.DailyStats{
		Date: t.Format("2006-01-02"),
	}
	for _, f := range food {
		stats.Calories += f.Calories
		stats.Protein += f.Protein
		stats.Carbs += f.Carbs
		stats.Fat += f.Fat
	}
	for _, w := range water {
		stats.WaterML += w.AmountML
	}
	return stats, nil
}

func (s *TrackerService) GetTodayFoodEntries() ([]models.FoodEntry, error) {
	return s.db.GetDailyFoodEntries(time.Now())
}

func (s *TrackerService) GetFoodEntriesRange(days int) ([]models.FoodEntry, error) {
	return s.db.GetFoodEntriesRange(days)
}

func (s *TrackerService) SetGoal(description string) error {
	goal := models.Goal{
		Timestamp:   time.Now(),
		Description: description,
	}
	return s.db.SetGoal(goal)
}

func (s *TrackerService) GetGoal() (string, error) {
	goal, err := s.db.GetLatestGoal()
	if err != nil {
		return "", err
	}
	if goal == nil {
		return "No goal set", nil
	}
	return goal.Description, nil
}

func (s *TrackerService) RemoveLastEntry() error {
	return s.db.RemoveLastEntry()
}

func (s *TrackerService) RunReview() (*models.ReviewResult, error) {
	goal, err := s.GetGoal()
	if err != nil {
		goal = "No goal set"
	}

	stats, err := s.db.GetStatsRange(7)
	if err != nil {
		return nil, err
	}

	// Create a map for easy lookup and ensure we have all 7 days (including today)
	statsMap := make(map[string]models.DailyStats)
	for _, st := range stats {
		statsMap[st.Date] = st
	}

	now := time.Now()
	allDays := make([]models.DailyStats, 0, 7)
	for i := 6; i >= 0; i-- {
		dateStr := now.AddDate(0, 0, -i).Format("2006-01-02")
		if st, ok := statsMap[dateStr]; ok {
			allDays = append(allDays, st)
		} else {
			allDays = append(allDays, models.DailyStats{Date: dateStr})
		}
	}

	foodEntries, err := s.db.GetFoodEntriesRange(7)
	if err != nil {
		return nil, err
	}
	simpleFoodEntries := make([]models.FoodEntrySimple, len(foodEntries))
	for i, e := range foodEntries {
		simpleFoodEntries[i] = models.FoodEntrySimple{
			Date:        e.Timestamp.Local().Format("2006-01-02"),
			Description: e.Description,
			Calories:    e.Calories,
			Protein:     e.Protein,
			Carbs:       e.Carbs,
			Fat:         e.Fat,
		}
	}

	waterEntries, err := s.db.GetWaterEntriesRange(7)
	if err != nil {
		return nil, err
	}
	simpleWaterEntries := make([]models.WaterEntrySimple, len(waterEntries))
	for i, e := range waterEntries {
		simpleWaterEntries[i] = models.WaterEntrySimple{
			Date:     e.Timestamp.Local().Format("2006-01-02"),
			AmountML: e.AmountML,
		}
	}

	data := models.ReviewData{
		Goal:         goal,
		Days:         allDays,
		FoodEntries:  simpleFoodEntries,
		WaterEntries: simpleWaterEntries,
	}

	return s.llm.AnalyzeReview(data)
}
