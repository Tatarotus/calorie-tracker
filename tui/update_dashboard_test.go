package tui

import (
	"testing"

	"calorie-tracker/config"
	"calorie-tracker/db"
	"calorie-tracker/services"

	tea "github.com/charmbracelet/bubbletea"
)

func TestHandleDashboardKeys_Quit(t *testing.T) {
	mockDB := db.NewMockDB()
	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := services.NewLLMService(cfg)
	tracker := services.NewTrackerService(mockDB, llm)

	m := NewModel(tracker)
	msg := tea.KeyMsg{Type: tea.KeyCtrlC}

	_, cmd := m.handleDashboardKeys(msg)

	if cmd == nil {
		t.Error("Expected non-nil command for quit")
	}
}

func TestHandleDashboardKeys_AddFood(t *testing.T) {
	mockDB := db.NewMockDB()
	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := services.NewLLMService(cfg)
	tracker := services.NewTrackerService(mockDB, llm)

	m := NewModel(tracker)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}

	model, cmd := m.handleDashboardKeys(msg)
	m2 := model

	if m2.Mode != AddFoodView {
		t.Errorf("Expected AddFoodView, got %v", m2.Mode)
	}
	_ = cmd
}

func TestHandleDashboardKeys_AddWater(t *testing.T) {
	mockDB := db.NewMockDB()
	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := services.NewLLMService(cfg)
	tracker := services.NewTrackerService(mockDB, llm)

	m := NewModel(tracker)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'w'}}

	model, cmd := m.handleDashboardKeys(msg)
	m2 := model

	if m2.Mode != AddWaterView {
		t.Errorf("Expected AddWaterView, got %v", m2.Mode)
	}
	_ = cmd
}

func TestHandleDashboardKeys_Review(t *testing.T) {
	mockDB := db.NewMockDB()
	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := services.NewLLMService(cfg)
	tracker := services.NewTrackerService(mockDB, llm)

	m := NewModel(tracker)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}

	model, cmd := m.handleDashboardKeys(msg)
	m2 := model

	if !m2.Loading {
		t.Error("Expected Loading to be true")
	}
	if cmd == nil {
		t.Error("Expected non-nil command for review")
	}
}

func TestHandleDashboardKeys_Today(t *testing.T) {
	mockDB := db.NewMockDB()
	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := services.NewLLMService(cfg)
	tracker := services.NewTrackerService(mockDB, llm)

	m := NewModel(tracker)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}}

	model, cmd := m.handleDashboardKeys(msg)
	m2 := model

	if m2.Mode != TodayLogView {
		t.Errorf("Expected TodayLogView, got %v", m2.Mode)
	}
	_ = cmd
}

func TestHandleDashboardKeys_Week(t *testing.T) {
	mockDB := db.NewMockDB()
	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := services.NewLLMService(cfg)
	tracker := services.NewTrackerService(mockDB, llm)

	m := NewModel(tracker)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'7'}}

	model, cmd := m.handleDashboardKeys(msg)
	m2 := model

	if m2.Mode != WeekLogView {
		t.Errorf("Expected WeekLogView, got %v", m2.Mode)
	}
	_ = cmd
}

func TestHandleDashboardKeys_Month(t *testing.T) {
	mockDB := db.NewMockDB()
	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := services.NewLLMService(cfg)
	tracker := services.NewTrackerService(mockDB, llm)

	m := NewModel(tracker)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}}

	model, cmd := m.handleDashboardKeys(msg)
	m2 := model

	if m2.Mode != MonthLogView {
		t.Errorf("Expected MonthLogView, got %v", m2.Mode)
	}
	_ = cmd
}

func TestHandleDashboardKeys_Goal(t *testing.T) {
	mockDB := db.NewMockDB()
	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := services.NewLLMService(cfg)
	tracker := services.NewTrackerService(mockDB, llm)

	m := NewModel(tracker)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}}

	model, cmd := m.handleDashboardKeys(msg)
	m2 := model

	if m2.Mode != SetGoalView {
		t.Errorf("Expected SetGoalView, got %v", m2.Mode)
	}
	_ = cmd
}

func TestHandleDashboardKeys_Undo(t *testing.T) {
	mockDB := db.NewMockDB()
	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := services.NewLLMService(cfg)
	tracker := services.NewTrackerService(mockDB, llm)

	m := NewModel(tracker)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}}

	model, cmd := m.handleDashboardKeys(msg)
	m2 := model

	if !m2.Loading {
		t.Error("Expected Loading to be true")
	}
	if cmd == nil {
		t.Error("Expected non-nil command for undo")
	}
}
