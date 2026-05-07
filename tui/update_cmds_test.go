package tui

import (
	"testing"

	"calorie-tracker/config"
	"calorie-tracker/db"
	"calorie-tracker/models"
	"calorie-tracker/services"

	tea "github.com/charmbracelet/bubbletea"
)

func TestUpdate_StatsMsg(t *testing.T) {
	mockDB := db.NewMockDB()
	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := services.NewLLMService(cfg)
	tracker := services.NewTrackerService(mockDB, llm)

	m := NewModel(tracker)
	m.Loading = true

	stats := models.DailyStats{
		Date:     "2024-01-15",
		Calories: 1500,
		Protein:  50,
		Carbs:    200,
		Fat:      60,
		WaterML:  1500,
	}

	model, _ := m.Update(StatsMsg(stats))
	m2 := model.(Model)

	if m2.Stats.Date != "2024-01-15" {
		t.Errorf("Expected stats to be updated")
	}
	if m2.Loading {
		t.Error("Expected Loading to be false after StatsMsg")
	}
}

func TestUpdate_GoalDescriptionMsg(t *testing.T) {
	mockDB := db.NewMockDB()
	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := services.NewLLMService(cfg)
	tracker := services.NewTrackerService(mockDB, llm)

	m := NewModel(tracker)
	model, _ := m.Update(GoalDescriptionMsg("Lose weight"))
	m2 := model.(Model)

	if m2.GoalDescription != "Lose weight" {
		t.Errorf("Expected goal description, got %s", m2.GoalDescription)
	}
}

func TestUpdate_FoodParsedMsg(t *testing.T) {
	mockDB := db.NewMockDB()
	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := services.NewLLMService(cfg)
	tracker := services.NewTrackerService(mockDB, llm)

	m := NewModel(tracker)
	m.Loading = true

	preview := &models.FoodPreview{
		Description: "Apple",
		Calories:    95,
	}

	model, _ := m.Update(FoodParsedMsg(preview))
	m2 := model.(Model)

	if m2.PendingFood == nil {
		t.Error("Expected PendingFood to be set")
	}
	if m2.Mode != ConfirmFoodView {
		t.Errorf("Expected ConfirmFoodView, got %v", m2.Mode)
	}
	if m2.Loading {
		t.Error("Expected Loading to be false")
	}
}

func TestUpdate_FoodSavedMsg(t *testing.T) {
	mockDB := db.NewMockDB()
	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := services.NewLLMService(cfg)
	tracker := services.NewTrackerService(mockDB, llm)

	m := NewModel(tracker)
	m.Loading = true
	m.Mode = ConfirmFoodView

	model, _ := m.Update(FoodSavedMsg{})
	m2 := model.(Model)

	if m2.Mode != DashboardView {
		t.Errorf("Expected DashboardView, got %v", m2.Mode)
	}
	if m2.Loading {
		t.Error("Expected Loading to be false")
	}
}

func TestUpdate_UndoMsg(t *testing.T) {
	mockDB := db.NewMockDB()
	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := services.NewLLMService(cfg)
	tracker := services.NewTrackerService(mockDB, llm)

	m := NewModel(tracker)
	m.Loading = true

	model, _ := m.Update(UndoMsg{})
	m2 := model.(Model)

	if m2.Loading {
		t.Error("Expected Loading to be false after UndoMsg")
	}
}

func TestUpdate_GoalSavedMsg(t *testing.T) {
	mockDB := db.NewMockDB()
	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := services.NewLLMService(cfg)
	tracker := services.NewTrackerService(mockDB, llm)

	m := NewModel(tracker)
	m.Loading = true
	m.Mode = SetGoalView

	model, _ := m.Update(GoalSavedMsg{})
	m2 := model.(Model)

	if m2.Mode != DashboardView {
		t.Errorf("Expected DashboardView, got %v", m2.Mode)
	}
	if m2.Loading {
		t.Error("Expected Loading to be false")
	}
}

func TestUpdate_TodayLogMsg(t *testing.T) {
	mockDB := db.NewMockDB()
	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := services.NewLLMService(cfg)
	tracker := services.NewTrackerService(mockDB, llm)

	m := NewModel(tracker)
	m.Loading = true

	entries := []models.FoodEntry{
		{Description: "Breakfast", Calories: 300},
	}

	model, _ := m.Update(TodayLogMsg(entries))
	m2 := model.(Model)

	if len(m2.TodayLog) != 1 {
		t.Errorf("Expected 1 today log entry, got %d", len(m2.TodayLog))
	}
	if m2.Loading {
		t.Error("Expected Loading to be false")
	}
}

func TestUpdate_WeekLogMsg(t *testing.T) {
	mockDB := db.NewMockDB()
	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := services.NewLLMService(cfg)
	tracker := services.NewTrackerService(mockDB, llm)

	m := NewModel(tracker)
	m.Loading = true

	entries := []models.FoodEntry{
		{Description: "Day 1", Calories: 1500},
	}

	model, _ := m.Update(WeekLogMsg(entries))
	m2 := model.(Model)

	if len(m2.WeekLog) != 1 {
		t.Errorf("Expected 1 week log entry, got %d", len(m2.WeekLog))
	}
	if m2.Loading {
		t.Error("Expected Loading to be false")
	}
}

func TestUpdate_MonthLogMsg(t *testing.T) {
	mockDB := db.NewMockDB()
	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := services.NewLLMService(cfg)
	tracker := services.NewTrackerService(mockDB, llm)

	m := NewModel(tracker)
	m.Loading = true

	entries := []models.FoodEntry{
		{Description: "Entry 1", Calories: 1500},
	}

	model, _ := m.Update(MonthLogMsg(entries))
	m2 := model.(Model)

	if len(m2.MonthLog) != 1 {
		t.Errorf("Expected 1 month log entry, got %d", len(m2.MonthLog))
	}
	if m2.Loading {
		t.Error("Expected Loading to be false")
	}
}

func TestUpdate_RecentLogMsg(t *testing.T) {
	mockDB := db.NewMockDB()
	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := services.NewLLMService(cfg)
	tracker := services.NewTrackerService(mockDB, llm)

	m := NewModel(tracker)
	entries := []models.FoodEntry{
		{Description: "Recent", Calories: 300},
	}

	model, _ := m.Update(RecentLogMsg(entries))
	m2 := model.(Model)

	if len(m2.RecentLog) != 1 {
		t.Errorf("Expected 1 recent log entry, got %d", len(m2.RecentLog))
	}
}

func TestUpdate_WaterMsg(t *testing.T) {
	mockDB := db.NewMockDB()
	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := services.NewLLMService(cfg)
	tracker := services.NewTrackerService(mockDB, llm)

	m := NewModel(tracker)
	m.Loading = true
	m.Mode = AddWaterView

	model, _ := m.Update(WaterMsg{})
	m2 := model.(Model)

	if m2.Mode != DashboardView {
		t.Errorf("Expected DashboardView, got %v", m2.Mode)
	}
	if m2.Loading {
		t.Error("Expected Loading to be false")
	}
}

func TestUpdate_ReviewMsg(t *testing.T) {
	mockDB := db.NewMockDB()
	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := services.NewLLMService(cfg)
	tracker := services.NewTrackerService(mockDB, llm)

	m := NewModel(tracker)
	m.Loading = true

	review := &models.ReviewResult{
		Score:    85,
		Progress: "Good",
		Summary:  "You're doing well!",
	}

	model, _ := m.Update(ReviewMsg(review))
	m2 := model.(Model)

	if m2.Review == nil {
		t.Error("Expected Review to be set")
	}
	if m2.Loading {
		t.Error("Expected Loading to be false")
	}
}

func TestUpdate_ErrMsg(t *testing.T) {
	mockDB := db.NewMockDB()
	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := services.NewLLMService(cfg)
	tracker := services.NewTrackerService(mockDB, llm)

	m := NewModel(tracker)
	m.Loading = true

	err := testError("test error")
	model, _ := m.Update(ErrMsg(err))
	m2 := model.(Model)

	if m2.Error == nil {
		t.Error("Expected Error to be set")
	}
	if m2.Loading {
		t.Error("Expected Loading to be false after ErrMsg")
	}
}

func TestUpdate_WindowSizeMsg(t *testing.T) {
	mockDB := db.NewMockDB()
	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := services.NewLLMService(cfg)
	tracker := services.NewTrackerService(mockDB, llm)

	m := NewModel(tracker)
	msg := tea.WindowSizeMsg{Width: 100, Height: 30}

	model, _ := m.Update(msg)
	m2 := model.(Model)

	if m2.Width != 100 {
		t.Errorf("Expected Width 100, got %d", m2.Width)
	}
	if m2.Height != 30 {
		t.Errorf("Expected Height 30, got %d", m2.Height)
	}
}

func TestUpdate_KeyMsg(t *testing.T) {
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
