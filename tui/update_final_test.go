package tui

import (
	"testing"

	"calorie-tracker/config"
	"calorie-tracker/db"
	"calorie-tracker/models"
	"calorie-tracker/services"

	tea "github.com/charmbracelet/bubbletea"
)

func TestUpdate_HandleKeyMsg_AddWater(t *testing.T) {
	mockDB := db.NewMockDB()
	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := services.NewLLMService(cfg)
	tracker := services.NewTrackerService(mockDB, llm)

	m := NewModel(tracker)
	m.Mode = DashboardView
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'w'}}

	model, _ := m.Update(msg)
	m2 := model.(Model)

	if m2.Mode != AddWaterView {
		t.Errorf("Expected AddWaterView, got %v", m2.Mode)
	}
}

func TestUpdate_HandleKeyMsg_Review(t *testing.T) {
	mockDB := db.NewMockDB()
	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := services.NewLLMService(cfg)
	tracker := services.NewTrackerService(mockDB, llm)

	m := NewModel(tracker)
	m.Mode = DashboardView
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}

	model, _ := m.Update(msg)
	m2 := model.(Model)

	if m2.Mode != ReviewView {
		t.Errorf("Expected ReviewView, got %v", m2.Mode)
	}
	if !m2.Loading {
		t.Error("Expected Loading to be true")
	}
}

func TestUpdate_HandleKeyMsg_Goal(t *testing.T) {
	mockDB := db.NewMockDB()
	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := services.NewLLMService(cfg)
	tracker := services.NewTrackerService(mockDB, llm)

	m := NewModel(tracker)
	m.Mode = DashboardView
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}}

	model, _ := m.Update(msg)
	m2 := model.(Model)

	if m2.Mode != SetGoalView {
		t.Errorf("Expected SetGoalView, got %v", m2.Mode)
	}
}

func TestUpdate_HandleKeyMsg_Undo(t *testing.T) {
	mockDB := db.NewMockDB()
	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := services.NewLLMService(cfg)
	tracker := services.NewTrackerService(mockDB, llm)

	m := NewModel(tracker)
	m.Mode = DashboardView
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}}

	model, _ := m.Update(msg)
	m2 := model.(Model)

	if !m2.Loading {
		t.Error("Expected Loading to be true")
	}
}

func TestUpdate_HandleKeyMsg_Week(t *testing.T) {
	mockDB := db.NewMockDB()
	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := services.NewLLMService(cfg)
	tracker := services.NewTrackerService(mockDB, llm)

	m := NewModel(tracker)
	m.Mode = DashboardView
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'7'}}

	model, _ := m.Update(msg)
	m2 := model.(Model)

	if m2.Mode != WeekLogView {
		t.Errorf("Expected WeekLogView, got %v", m2.Mode)
	}
	if !m2.Loading {
		t.Error("Expected Loading to be true")
	}
}

func TestUpdate_HandleKeyMsg_Month(t *testing.T) {
	mockDB := db.NewMockDB()
	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := services.NewLLMService(cfg)
	tracker := services.NewTrackerService(mockDB, llm)

	m := NewModel(tracker)
	m.Mode = DashboardView
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}}

	model, _ := m.Update(msg)
	m2 := model.(Model)

	if m2.Mode != MonthLogView {
		t.Errorf("Expected MonthLogView, got %v", m2.Mode)
	}
	if !m2.Loading {
		t.Error("Expected Loading to be true")
	}
}

func TestUpdate_HandleKeyMsg_Unknown(t *testing.T) {
	mockDB := db.NewMockDB()
	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := services.NewLLMService(cfg)
	tracker := services.NewTrackerService(mockDB, llm)

	m := NewModel(tracker)
	m.Mode = DashboardView
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}

	model, _ := m.Update(msg)
	m2 := model.(Model)

	if m2.Mode != DashboardView {
		t.Errorf("Expected DashboardView, got %v", m2.Mode)
	}
}

func TestUpdate_HandleKeyMsg_Quit(t *testing.T) {
	mockDB := db.NewMockDB()
	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := services.NewLLMService(cfg)
	tracker := services.NewTrackerService(mockDB, llm)

	m := NewModel(tracker)
	m.Mode = DashboardView
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}

	model, cmd := m.Update(msg)
	m2 := model.(Model)

	if m2.Mode != DashboardView {
		t.Errorf("Expected DashboardView, got %v", m2.Mode)
	}
	_ = cmd
}

func TestUpdate_HandleKeyMsg_Today(t *testing.T) {
	mockDB := db.NewMockDB()
	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := services.NewLLMService(cfg)
	tracker := services.NewTrackerService(mockDB, llm)

	m := NewModel(tracker)
	m.Mode = DashboardView
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}}

	model, _ := m.Update(msg)
	m2 := model.(Model)

	if m2.Mode != TodayLogView {
		t.Errorf("Expected TodayLogView, got %v", m2.Mode)
	}
	if !m2.Loading {
		t.Error("Expected Loading to be true")
	}
}

func TestUpdate_HandleKeyMsg_EditPreviewEnter(t *testing.T) {
	mockDB := db.NewMockDB()
	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := services.NewLLMService(cfg)
	tracker := services.NewTrackerService(mockDB, llm)

	m := NewModel(tracker)
	m.Mode = EditFoodPreviewView
	m.EditField = 3
	m.PendingFood = &models.FoodPreview{
		Calories: 100,
		Protein:  5,
		Carbs:    20,
		Fat:      2,
	}
	m.EditInput.SetValue("3.5")
	msg := tea.KeyMsg{Type: tea.KeyEnter}

	model, _ := m.Update(msg)
	m2 := model.(Model)

	if m2.Mode != ConfirmFoodView {
		t.Errorf("Expected ConfirmFoodView, got %v", m2.Mode)
	}
	if m2.PendingFood.Fat != 3.5 {
		t.Errorf("Expected Fat 3.5, got %f", m2.PendingFood.Fat)
	}
}
