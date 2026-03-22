package services

import (
	"calorie-tracker/db"
	"calorie-tracker/models"
	"time"
)

type TrackerService struct {
	db      *db.DB
	llm     *LLMService
	matcher *FoodMatcher
}

func NewTrackerService(db *db.DB, llm *LLMService) *TrackerService {
	return &TrackerService{
		db:      db,
		llm:     llm,
		matcher: NewFoodMatcher(db),
	}
}

func (s *TrackerService) ParseFood(description string) (*models.FoodPreview, error) {
	if matched, err := s.matcher.Match(description); err == nil && matched != nil {
		return matched, nil
	}

	preview, err := s.llm.ParseFood(description)
	if err != nil {
		return nil, err
	}
	return preview, nil
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

func (s *TrackerService) RunReview() (*models.ReviewResult, error) {
	goal, err := s.GetGoal()
	if err != nil {
		goal = "No goal set"
	}

	stats, err := s.db.GetStatsRange(7)
	if err != nil {
		return nil, err
	}
	
	entries, err := s.db.GetFoodEntriesRange(7)
	if err != nil {
		return nil, err
	}
	
	simpleEntries := make([]models.FoodEntrySimple, len(entries))
	for i, e := range entries {
		simpleEntries[i] = models.FoodEntrySimple{
			Date:        e.Timestamp.Format("2006-01-02"),
			Description: e.Description,
			Calories:    e.Calories,
			Protein:     e.Protein,
			Carbs:       e.Carbs,
			Fat:         e.Fat,
		}
	}

	data := models.ReviewData{
		Goal:    goal,
		Days:    stats,
		Entries: simpleEntries,
	}
	
	return s.llm.AnalyzeReview(data)
}
