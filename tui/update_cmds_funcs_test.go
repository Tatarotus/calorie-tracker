package tui

import (
	"testing"

	"calorie-tracker/config"
	"calorie-tracker/db"
	"calorie-tracker/models"
	"calorie-tracker/services"

	tea "github.com/charmbracelet/bubbletea"
)

func TestUpdateCmdFuncs_Part1(t *testing.T) {
	mockDB := db.NewMockDB()
	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := services.NewLLMService(cfg)
	tracker := services.NewTrackerService(mockDB, llm)
	m := NewModel(tracker)

	t.Run("getStatsCmd", func(t *testing.T) {
		cmd := m.getStatsCmd()
		msg := cmd()

		if batch, ok := msg.(tea.BatchMsg); ok {
			for _, c := range batch {
				m := c()
				switch m.(type) {
				case StatsMsg, RecentLogMsg, ErrMsg:
					// Expected
				default:
					t.Errorf("Expected StatsMsg, RecentLogMsg, or ErrMsg, got %T", m)
				}
			}
		} else {
			t.Errorf("Expected tea.BatchMsg, got %T", msg)
		}
	})

	t.Run("getGoalCmd", func(t *testing.T) {
		cmd := m.getGoalCmd()
		msg := cmd()
		if _, ok := msg.(GoalDescriptionMsg); !ok {
			if _, isErr := msg.(ErrMsg); !isErr {
				t.Errorf("Expected GoalDescriptionMsg or ErrMsg, got %T", msg)
			}
		}
	})

	t.Run("parseFoodCmd", func(t *testing.T) {
		cmd := m.parseFoodCmd("apple")
		msg := cmd()
		switch msg.(type) {
		case FoodParsedMsg, ErrMsg:
			// Expected
		default:
			t.Errorf("Expected FoodParsedMsg or ErrMsg, got %T", msg)
		}
	})

	t.Run("saveFoodCmd", func(t *testing.T) {
		preview := &models.FoodPreview{Description: "Apple", Calories: 95}
		cmd := m.saveFoodCmd(preview)
		msg := cmd()
		if _, ok := msg.(FoodSavedMsg); !ok {
			if _, isErr := msg.(ErrMsg); !isErr {
				t.Errorf("Expected FoodSavedMsg or ErrMsg, got %T", msg)
			}
		}
	})

	t.Run("addWaterCmd", func(t *testing.T) {
		cmd := m.addWaterCmd(500)
		msg := cmd()
		if _, ok := msg.(WaterMsg); !ok {
			if _, isErr := msg.(ErrMsg); !isErr {
				t.Errorf("Expected WaterMsg or ErrMsg, got %T", msg)
			}
		}
	})

	t.Run("saveGoalCmd", func(t *testing.T) {
		cmd := m.saveGoalCmd("Lose weight")
		msg := cmd()
		if _, ok := msg.(GoalSavedMsg); !ok {
			if _, isErr := msg.(ErrMsg); !isErr {
				t.Errorf("Expected GoalSavedMsg or ErrMsg, got %T", msg)
			}
		}
	})
}

func TestUpdateCmdFuncs_Part2(t *testing.T) {
	mockDB := db.NewMockDB()
	cfg := &config.Config{
		SambaAPIKey:   "test",
		OpenAIBaseURL: "https://test.com/v1",
	}
	llm := services.NewLLMService(cfg)
	tracker := services.NewTrackerService(mockDB, llm)
	m := NewModel(tracker)

	t.Run("removeLastEntryCmd", func(t *testing.T) {
		cmd := m.removeLastEntryCmd()
		msg := cmd()
		if _, ok := msg.(UndoMsg); !ok {
			if _, isErr := msg.(ErrMsg); !isErr {
				t.Errorf("Expected UndoMsg or ErrMsg, got %T", msg)
			}
		}
	})

	t.Run("runReviewCmd", func(t *testing.T) {
		cmd := m.runReviewCmd()
		msg := cmd()
		switch msg.(type) {
		case ReviewMsg, ErrMsg:
			// Expected
		default:
			t.Errorf("Expected ReviewMsg or ErrMsg, got %T", msg)
		}
	})

	t.Run("getTodayLogCmd", func(t *testing.T) {
		cmd := m.getTodayLogCmd()
		msg := cmd()
		if _, ok := msg.(TodayLogMsg); !ok {
			if _, isErr := msg.(ErrMsg); !isErr {
				t.Errorf("Expected TodayLogMsg or ErrMsg, got %T", msg)
			}
		}
	})

	t.Run("getWeekLogCmd", func(t *testing.T) {
		cmd := m.getWeekLogCmd()
		msg := cmd()
		if _, ok := msg.(WeekLogMsg); !ok {
			if _, isErr := msg.(ErrMsg); !isErr {
				t.Errorf("Expected WeekLogMsg or ErrMsg, got %T", msg)
			}
		}
	})

	t.Run("getMonthLogCmd", func(t *testing.T) {
		cmd := m.getMonthLogCmd()
		msg := cmd()
		if _, ok := msg.(MonthLogMsg); !ok {
			if _, isErr := msg.(ErrMsg); !isErr {
				t.Errorf("Expected MonthLogMsg or ErrMsg, got %T", msg)
			}
		}
	})
}
