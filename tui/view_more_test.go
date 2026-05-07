package tui

import (
	"testing"

	"calorie-tracker/config"
	"calorie-tracker/db"
	"calorie-tracker/models"
	"calorie-tracker/services"
)

func TestView_LoadingStates(t *testing.T) {
	mockDB := db.NewMockDB()
	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := services.NewLLMService(cfg)
	tracker := services.NewTrackerService(mockDB, llm)

	modes := []ViewMode{AddFoodView, ReviewView, ConfirmFoodView, SetGoalView}
	for _, mode := range modes {
		m := NewModel(tracker)
		m.Loading = true
		m.Mode = mode
		m.Width = 80
		m.Height = 24

		view := m.View()
		if view == "" {
			t.Errorf("Expected non-empty view for mode %v", mode)
		}
	}
}

func TestRenderMonthLogString_WithEntries(t *testing.T) {
	mockDB := db.NewMockDB()
	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := services.NewLLMService(cfg)
	tracker := services.NewTrackerService(mockDB, llm)

	m := NewModel(tracker)
	m.MonthLog = []models.FoodEntry{
		{Description: "Entry 1", Calories: 1500},
	}

	view := m.renderMonthLogString()
	if view == "" {
		t.Error("Expected non-empty month log")
	}
}

func TestRenderMonthLogString_Empty(t *testing.T) {
	mockDB := db.NewMockDB()
	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := services.NewLLMService(cfg)
	tracker := services.NewTrackerService(mockDB, llm)

	m := NewModel(tracker)
	m.MonthLog = []models.FoodEntry{}

	view := m.renderMonthLogString()
	if view == "" {
		t.Error("Expected non-empty month log")
	}
}
