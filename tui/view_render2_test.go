package tui

import (
	"testing"

	"calorie-tracker/config"
	"calorie-tracker/db"
	"calorie-tracker/models"
	"calorie-tracker/services"
)

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
