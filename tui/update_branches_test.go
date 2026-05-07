package tui

import (
	"testing"

	"calorie-tracker/config"
	"calorie-tracker/db"
	"calorie-tracker/models"
	"calorie-tracker/services"

	tea "github.com/charmbracelet/bubbletea"
)

func TestUpdate_HandleKeyMsg_Dashboard(t *testing.T) {
	mockDB := db.NewMockDB()
	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := services.NewLLMService(cfg)
	tracker := services.NewTrackerService(mockDB, llm)

	m := NewModel(tracker)
	m.Mode = DashboardView
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}

	model, _ := m.Update(msg)
	m2 := model.(Model)

	if m2.Mode != AddFoodView {
		t.Errorf("Expected AddFoodView, got %v", m2.Mode)
	}
}

func TestUpdate_HandleKeyMsg_InputMode(t *testing.T) {
	mockDB := db.NewMockDB()
	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := services.NewLLMService(cfg)
	tracker := services.NewTrackerService(mockDB, llm)

	m := NewModel(tracker)
	m.Mode = AddFoodView
	msg := tea.KeyMsg{Type: tea.KeyEsc}

	model, _ := m.Update(msg)
	m2 := model.(Model)

	if m2.Mode != DashboardView {
		t.Errorf("Expected DashboardView, got %v", m2.Mode)
	}
}

func TestUpdate_HandleKeyMsg_ConfirmFood(t *testing.T) {
	mockDB := db.NewMockDB()
	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := services.NewLLMService(cfg)
	tracker := services.NewTrackerService(mockDB, llm)

	m := NewModel(tracker)
	m.Mode = ConfirmFoodView
	m.PendingFood = &models.FoodPreview{Description: "Apple"}
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}

	model, _ := m.Update(msg)
	m2 := model.(Model)

	if m2.Mode != DashboardView {
		t.Errorf("Expected DashboardView, got %v", m2.Mode)
	}
}

func TestUpdate_HandleKeyMsg_EditPreview(t *testing.T) {
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
	m.PendingFood = &models.FoodPreview{Calories: 100}
	msg := tea.KeyMsg{Type: tea.KeyEsc}

	model, _ := m.Update(msg)
	m2 := model.(Model)

	if m2.Mode != ConfirmFoodView {
		t.Errorf("Expected ConfirmFoodView, got %v", m2.Mode)
	}
}

func TestUpdate_HandleKeyMsg_LogView(t *testing.T) {
	mockDB := db.NewMockDB()
	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := services.NewLLMService(cfg)
	tracker := services.NewTrackerService(mockDB, llm)

	m := NewModel(tracker)
	m.Mode = ReviewView
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}

	model, _ := m.Update(msg)
	m2 := model.(Model)

	if m2.Mode != DashboardView {
		t.Errorf("Expected DashboardView, got %v", m2.Mode)
	}
}

func TestUpdate_DefaultCase(t *testing.T) {
	mockDB := db.NewMockDB()
	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := services.NewLLMService(cfg)
	tracker := services.NewTrackerService(mockDB, llm)

	m := NewModel(tracker)
	// Send an unknown message type
	model, _ := m.Update("unknown message")
	m2 := model.(Model)

	// Should just return the model unchanged
	if m2.Mode != DashboardView {
		t.Errorf("Expected DashboardView, got %v", m2.Mode)
	}
}
