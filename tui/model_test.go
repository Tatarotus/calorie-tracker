package tui

import (
	"testing"

	"calorie-tracker/config"
	"calorie-tracker/db"
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

type testError string

func (e testError) Error() string {
	return string(e)
}
