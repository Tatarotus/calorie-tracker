package services

import (
	"calorie-tracker/db"
	"calorie-tracker/models"
	"fmt"
	"time"
)

type TrackerService struct {
	db      db.DBProvider
	llm     *LLMService
	matcher *FoodMatcher
}

func NewTrackerService(db db.DBProvider, llm *LLMService) *TrackerService {
	return &TrackerService{
		db:      db,
		llm:     llm,
		matcher: NewFoodMatcher(db),
	}
}

func (s *TrackerService) ParseFood(description string) (*models.FoodPreview, error) {
	parsed := s.matcher.Parse(description)
	if parsed.Name == "" {
		return nil, fmt.Errorf("could not parse food name from: %s", description)
	}

	// 1. Check local reference database first (Source of Truth)
	ref, err := s.db.GetReferenceFood(parsed.Name)
	if err != nil {
		return nil, fmt.Errorf("reference lookup error: %w", err)
	}

	if ref != nil {
		// Use deterministic scaling
		return s.scaleMacros(ref, parsed), nil
	}

	// 2. Check cache for previous LLM results
	matched, err := s.db.GetCachedFood(parsed.Name)
	if err != nil {
		return nil, fmt.Errorf("cache match error: %w", err)
	}
	if matched != nil {
		return &models.FoodPreview{
			Description: matched.Description,
			Calories:    matched.Calories,
			Protein:     matched.Protein,
			Carbs:       matched.Carbs,
			Fat:         matched.Fat,
		}, nil
	}

	// 3. Fallback to LLM
	preview, err := s.llm.ParseFood(description)
	if err != nil {
		return nil, err
	}

	// Validate LLM response
	if preview.Calories < 0 || preview.Protein < 0 || preview.Carbs < 0 || preview.Fat < 0 {
		return nil, fmt.Errorf("LLM returned unrealistic negative values")
	}
	if preview.Calories > 5000 {
		return nil, fmt.Errorf("LLM returned unrealistic high calorie value: %.0f", preview.Calories)
	}

	return preview, nil
}

func (s *TrackerService) scaleMacros(ref *models.ReferenceFood, parsed ParsedFood) *models.FoodPreview {
	amount := parsed.Amount
	if amount == 0 {
		amount = ref.BaseQuantity
	}

	scale := amount / ref.BaseQuantity

	desc := fmt.Sprintf("%.1f%s %s", amount, ref.Unit, ref.Name)
	if ref.Unit == "unit" {
		desc = fmt.Sprintf("%.0f %s", amount, ref.Name)
	}

	return &models.FoodPreview{
		Description: desc,
		Calories:    ref.Calories * scale,
		Protein:     ref.Protein * scale,
		Carbs:       ref.Carbs * scale,
		Fat:         ref.Fat * scale,
	}
}

func (s *TrackerService) SaveFood(preview *models.FoodPreview) error {
	entry := models.FoodEntry{
		Timestamp:   time.Now(),
		Description: preview.Description,
		Calories:    preview.Calories,
		Protein:     preview.Protein,
		Carbs:       preview.Carbs,
		Fat:         preview.Fat,
	}
	if err := s.db.AddFoodEntry(entry); err != nil {
		return err
	}
	return s.db.CacheFood(entry)
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
