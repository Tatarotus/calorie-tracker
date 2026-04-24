package tui

import (
	"testing"

	"calorie-tracker/config"
	"calorie-tracker/db"
	"calorie-tracker/models"
	"calorie-tracker/services"
)

func TestNewModel(t *testing.T) {
	mockDB := db.NewMockDB()
	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := services.NewLLMService(cfg)
	tracker := services.NewTrackerService(mockDB, llm)

	m := NewModel(tracker)

	if m.Tracker != tracker {
		t.Error("Expected Tracker to be set")
	}
	if m.Mode != DashboardView {
		t.Errorf("Expected Mode to be DashboardView, got %v", m.Mode)
	}
	if m.FoodInput.Placeholder != "e.g. 2 eggs and a coffee" {
		t.Errorf("Expected FoodInput placeholder, got %s", m.FoodInput.Placeholder)
	}
	if m.WaterInput.Placeholder != "e.g. 500" {
		t.Errorf("Expected WaterInput placeholder, got %s", m.WaterInput.Placeholder)
	}
	if m.GoalInput.Placeholder != "e.g. I want to reach 80kg in 8 months" {
		t.Errorf("Expected GoalInput placeholder, got %s", m.GoalInput.Placeholder)
	}
}

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

func TestRenderGoalComparison_WithCalorieGoal(t *testing.T) {
	mockDB := db.NewMockDB()
	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := services.NewLLMService(cfg)
	tracker := services.NewTrackerService(mockDB, llm)

	m := NewModel(tracker)
	m.GoalDescription = "Eat 2000 kcal per day"
	m.Stats.Calories = 1800

	view := m.renderGoalComparison()

	if view == "" {
		t.Error("Expected non-empty goal comparison")
	}
}

func TestRenderGoalComparison_WithWeightGoal(t *testing.T) {
	mockDB := db.NewMockDB()
	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := services.NewLLMService(cfg)
	tracker := services.NewTrackerService(mockDB, llm)

	m := NewModel(tracker)
	m.GoalDescription = "Reach 70kg"
	m.Stats.Calories = 1800

	view := m.renderGoalComparison()

	if view == "" {
		t.Error("Expected non-empty goal comparison")
	}
}

func TestRenderGoalComparison_NoMatch(t *testing.T) {
	mockDB := db.NewMockDB()
	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := services.NewLLMService(cfg)
	tracker := services.NewTrackerService(mockDB, llm)

	m := NewModel(tracker)
	m.GoalDescription = "Just a general goal"
	m.Stats.Calories = 1800

	view := m.renderGoalComparison()

	if view == "" {
		t.Error("Expected non-empty goal comparison")
	}
}

func TestRenderTodayLogString_WithEntries(t *testing.T) {
	mockDB := db.NewMockDB()
	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := services.NewLLMService(cfg)
	tracker := services.NewTrackerService(mockDB, llm)

	m := NewModel(tracker)
	m.TodayLog = []models.FoodEntry{
		{Description: "Breakfast", Calories: 300},
		{Description: "Lunch", Calories: 500},
	}

	view := m.renderTodayLogString()

	if view == "" {
		t.Error("Expected non-empty today log")
	}
}

func TestRenderTodayLogString_Empty(t *testing.T) {
	mockDB := db.NewMockDB()
	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := services.NewLLMService(cfg)
	tracker := services.NewTrackerService(mockDB, llm)

	m := NewModel(tracker)
	m.TodayLog = []models.FoodEntry{}

	view := m.renderTodayLogString()

	if view == "" {
		t.Error("Expected non-empty today log")
	}
}

func TestRenderWeekLogString_WithEntries(t *testing.T) {
	mockDB := db.NewMockDB()
	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := services.NewLLMService(cfg)
	tracker := services.NewTrackerService(mockDB, llm)

	m := NewModel(tracker)
	m.WeekLog = []models.FoodEntry{
		{Description: "Day 1", Calories: 1500},
		{Description: "Day 2", Calories: 1600},
	}

	view := m.renderWeekLogString()

	if view == "" {
		t.Error("Expected non-empty week log")
	}
}

func TestRenderWeekLogString_Empty(t *testing.T) {
	mockDB := db.NewMockDB()
	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := services.NewLLMService(cfg)
	tracker := services.NewTrackerService(mockDB, llm)

	m := NewModel(tracker)
	m.WeekLog = []models.FoodEntry{}

	view := m.renderWeekLogString()

	if view == "" {
		t.Error("Expected non-empty week log")
	}
}

func TestRenderReviewString_WithReview(t *testing.T) {
	mockDB := db.NewMockDB()
	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := services.NewLLMService(cfg)
	tracker := services.NewTrackerService(mockDB, llm)

	m := NewModel(tracker)
	m.Review = &models.ReviewResult{
		GoalProgress: "On track",
		Summary:      "Good job",
		Issues:       []string{"High sodium"},
		Patterns:     []string{"Consistent protein"},
		Suggestions:  []string{"Drink more water"},
	}

	view := m.renderReviewString()

	if view == "" {
		t.Error("Expected non-empty review string")
	}
}

func TestRenderReviewString_NoReview(t *testing.T) {
	mockDB := db.NewMockDB()
	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := services.NewLLMService(cfg)
	tracker := services.NewTrackerService(mockDB, llm)

	m := NewModel(tracker)
	m.Review = nil

	view := m.renderReviewString()

	if view == "" {
		t.Error("Expected non-empty review string")
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

	// Test log views
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

	// Test non-log views
	m.Mode = DashboardView
	if m.isLogOrReviewView() {
		t.Error("Expected DashboardView to not be a log/review view")
	}

	m.Mode = AddFoodView
	if m.isLogOrReviewView() {
		t.Error("Expected AddFoodView to not be a log/review view")
	}
}

// testError is a simple error type for testing
type testError string

func (e testError) Error() string {
	return string(e)
}
