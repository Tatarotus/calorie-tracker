package tui

import (
	"testing"

	"calorie-tracker/config"
	"calorie-tracker/db"
	"calorie-tracker/models"
	"calorie-tracker/services"
)

func TestView_DashboardView(t *testing.T) {
	mockDB := db.NewMockDB()
	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := services.NewLLMService(cfg)
	tracker := services.NewTrackerService(mockDB, llm)

	m := NewModel(tracker)
	m.Mode = DashboardView
	m.Stats = models.DailyStats{
		Date:     "2024-01-15",
		Calories: 1500,
		Protein:  50,
		Carbs:    200,
		Fat:      60,
		WaterML:  1500,
	}
	m.GoalDescription = "Lose weight"
	m.RecentLog = []models.FoodEntry{
		{Description: "Breakfast", Calories: 300},
		{Description: "Lunch", Calories: 500},
	}
	m.Width = 80
	m.Height = 24

	view := m.View()

	if view == "" {
		t.Error("Expected non-empty view")
	}
	if len(view) < 100 {
		t.Errorf("Expected longer view, got %d chars", len(view))
	}
}

func TestView_AddFoodView(t *testing.T) {
	mockDB := db.NewMockDB()
	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := services.NewLLMService(cfg)
	tracker := services.NewTrackerService(mockDB, llm)

	m := NewModel(tracker)
	m.Mode = AddFoodView
	m.Width = 80
	m.Height = 24

	view := m.View()

	if view == "" {
		t.Error("Expected non-empty view")
	}
}

func TestView_AddWaterView(t *testing.T) {
	mockDB := db.NewMockDB()
	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := services.NewLLMService(cfg)
	tracker := services.NewTrackerService(mockDB, llm)

	m := NewModel(tracker)
	m.Mode = AddWaterView
	m.Width = 80
	m.Height = 24

	view := m.View()

	if view == "" {
		t.Error("Expected non-empty view")
	}
}

func TestView_ReviewView_WithReview(t *testing.T) {
	mockDB := db.NewMockDB()
	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := services.NewLLMService(cfg)
	tracker := services.NewTrackerService(mockDB, llm)

	m := NewModel(tracker)
	m.Mode = ReviewView
	m.Review = &models.ReviewResult{
		Score:       85,
		Progress:    "Good",
		Summary:     "You're doing well!",
		Issues:      []string{"High sodium"},
		Patterns:    []string{"Consistent protein intake"},
		Suggestions: []string{"Drink more water"},
	}
	m.Width = 80
	m.Height = 24

	view := m.View()

	if view == "" {
		t.Error("Expected non-empty view")
	}
}

func TestView_ReviewView_NoReview(t *testing.T) {
	mockDB := db.NewMockDB()
	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := services.NewLLMService(cfg)
	tracker := services.NewTrackerService(mockDB, llm)

	m := NewModel(tracker)
	m.Mode = ReviewView
	m.Review = nil
	m.Width = 80
	m.Height = 24

	view := m.View()

	if view == "" {
		t.Error("Expected non-empty view")
	}
}

func TestView_ConfirmFoodView(t *testing.T) {
	mockDB := db.NewMockDB()
	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := services.NewLLMService(cfg)
	tracker := services.NewTrackerService(mockDB, llm)

	m := NewModel(tracker)
	m.Mode = ConfirmFoodView
	m.PendingFood = &models.FoodPreview{
		Description: "Apple",
		Calories:    95,
		Protein:     0.5,
		Carbs:       25,
		Fat:         0.3,
	}
	m.Width = 80
	m.Height = 24

	view := m.View()

	if view == "" {
		t.Error("Expected non-empty view")
	}
}

func TestView_TodayLogView(t *testing.T) {
	mockDB := db.NewMockDB()
	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := services.NewLLMService(cfg)
	tracker := services.NewTrackerService(mockDB, llm)

	m := NewModel(tracker)
	m.Mode = TodayLogView
	m.TodayLog = []models.FoodEntry{
		{Description: "Breakfast", Calories: 300},
		{Description: "Lunch", Calories: 500},
	}
	m.Width = 80
	m.Height = 24

	view := m.View()

	if view == "" {
		t.Error("Expected non-empty view")
	}
}

func TestView_WeekLogView(t *testing.T) {
	mockDB := db.NewMockDB()
	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := services.NewLLMService(cfg)
	tracker := services.NewTrackerService(mockDB, llm)

	m := NewModel(tracker)
	m.Mode = WeekLogView
	m.WeekLog = []models.FoodEntry{
		{Description: "Day 1", Calories: 1500},
		{Description: "Day 2", Calories: 1600},
	}
	m.Width = 80
	m.Height = 24

	view := m.View()

	if view == "" {
		t.Error("Expected non-empty view")
	}
}

func TestView_MonthLogView(t *testing.T) {
	mockDB := db.NewMockDB()
	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := services.NewLLMService(cfg)
	tracker := services.NewTrackerService(mockDB, llm)

	m := NewModel(tracker)
	m.Mode = MonthLogView
	m.MonthLog = []models.FoodEntry{
		{Description: "Entry 1", Calories: 1500},
	}
	m.Width = 80
	m.Height = 24

	view := m.View()

	if view == "" {
		t.Error("Expected non-empty view")
	}
}

func TestView_SetGoalView(t *testing.T) {
	mockDB := db.NewMockDB()
	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := services.NewLLMService(cfg)
	tracker := services.NewTrackerService(mockDB, llm)

	m := NewModel(tracker)
	m.Mode = SetGoalView
	m.Width = 80
	m.Height = 24

	view := m.View()

	if view == "" {
		t.Error("Expected non-empty view")
	}
}

func TestView_EditFoodPreviewView(t *testing.T) {
	mockDB := db.NewMockDB()
	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := services.NewLLMService(cfg)
	tracker := services.NewTrackerService(mockDB, llm)

	m := NewModel(tracker)
	m.Mode = EditFoodPreviewView
	m.EditField = 0
	m.PendingFood = &models.FoodPreview{
		Calories: 100,
		Protein:  5,
		Carbs:    20,
		Fat:      2,
	}
	m.Width = 80
	m.Height = 24

	view := m.View()

	if view == "" {
		t.Error("Expected non-empty view")
	}
}

func TestView_Loading(t *testing.T) {
	mockDB := db.NewMockDB()
	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := services.NewLLMService(cfg)
	tracker := services.NewTrackerService(mockDB, llm)

	m := NewModel(tracker)
	m.Loading = true
	m.Mode = AddFoodView
	m.Width = 80
	m.Height = 24

	view := m.View()

	if view == "" {
		t.Error("Expected non-empty view")
	}
}

func TestView_Error(t *testing.T) {
	mockDB := db.NewMockDB()
	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := services.NewLLMService(cfg)
	tracker := services.NewTrackerService(mockDB, llm)

	m := NewModel(tracker)
	m.Error = testError("test error")
	m.Width = 80
	m.Height = 24

	view := m.View()

	if view == "" {
		t.Error("Expected non-empty view")
	}
}

func TestIsLogOrReviewView(t *testing.T) {
	mockDB := db.NewMockDB()
	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := services.NewLLMService(cfg)
	tracker := services.NewTrackerService(mockDB, llm)

	m := NewModel(tracker)

	m.Mode = ReviewView
	if !m.isLogOrReviewView() {
		t.Error("Expected ReviewView to be a log/review view")
	}

	m.Mode = TodayLogView
	if !m.isLogOrReviewView() {
		t.Error("Expected TodayLogView to be a log/review view")
	}

	m.Mode = WeekLogView
	if !m.isLogOrReviewView() {
		t.Error("Expected WeekLogView to be a log/review view")
	}

	m.Mode = MonthLogView
	if !m.isLogOrReviewView() {
		t.Error("Expected MonthLogView to be a log/review view")
	}

	m.Mode = DashboardView
	if m.isLogOrReviewView() {
		t.Error("Expected DashboardView to not be a log/review view")
	}

	m.Mode = AddFoodView
	if m.isLogOrReviewView() {
		t.Error("Expected AddFoodView to not be a log/review view")
	}
}
