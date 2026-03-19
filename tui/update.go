package tui

import (
	"calorie-tracker/models"
	"fmt"
	"strconv"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type (
	StatsMsg      models.DailyStats
	FoodParsedMsg *models.FoodPreview
	FoodSavedMsg  struct{}
	WaterMsg      struct{}
	ReviewMsg     *models.ReviewResult
	TodayLogMsg   []models.FoodEntry
	ErrMsg        error
)

func (m Model) Init() tea.Cmd {
	return m.getStatsCmd()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height

	case tea.KeyMsg:
		// Handle global keys and mode-specific keys
		switch m.Mode {
		case AddFoodView, AddWaterView:
			switch msg.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "esc":
				m.Mode = DashboardView
				return m, m.getStatsCmd()
			case "enter":
				if m.Mode == AddFoodView {
					m.Loading = true
					m.Error = nil
					return m, m.parseFoodCmd(m.FoodInput.Value())
				}
				if m.Mode == AddWaterView {
					m.Loading = true
					m.Error = nil
					amount, _ := strconv.ParseFloat(m.WaterInput.Value(), 64)
					return m, m.addWaterCmd(amount)
				}
			}

		case ConfirmFoodView:
			switch msg.String() {
			case "y":
				m.Loading = true
				return m, m.saveFoodCmd(m.PendingFood)
			case "n":
				m.Mode = DashboardView
				return m, m.getStatsCmd()
			case "e":
				m.Mode = EditFoodPreviewView
				m.EditField = 0
				m.setupEditInput()
			case "ctrl+c", "q":
				return m, tea.Quit
			}

		case EditFoodPreviewView:
			switch msg.String() {
			case "enter":
				m.updatePendingFoodFromEdit()
				if m.EditField < 3 {
					m.EditField++
					m.setupEditInput()
				} else {
					m.Mode = ConfirmFoodView
				}
			case "esc":
				m.Mode = ConfirmFoodView
			case "ctrl+c":
				return m, tea.Quit
			}

		default: // Dashboard, Review, TodayLog
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "d":
				m.Mode = DashboardView
				return m, m.getStatsCmd()
			case "a":
				m.Mode = AddFoodView
				m.FoodInput.Focus()
				m.FoodInput.SetValue("")
			case "w":
				m.Mode = AddWaterView
				m.WaterInput.Focus()
				m.WaterInput.SetValue("")
			case "r":
				m.Mode = ReviewView
				m.Loading = true
				return m, m.runReviewCmd()
			case "t":
				m.Mode = TodayLogView
				m.Loading = true
				return m, m.getTodayLogCmd()
			}
		}

	case StatsMsg:
		m.Stats = models.DailyStats(msg)
		m.Loading = false

	case FoodParsedMsg:
		m.Loading = false
		m.PendingFood = (*models.FoodPreview)(msg)
		m.Mode = ConfirmFoodView

	case FoodSavedMsg:
		m.Loading = false
		m.Mode = DashboardView
		return m, m.getStatsCmd()

	case TodayLogMsg:
		m.Loading = false
		m.TodayLog = []models.FoodEntry(msg)

	case WaterMsg:
		m.Loading = false
		m.Mode = DashboardView
		return m, m.getStatsCmd()

	case ReviewMsg:
		m.Loading = false
		m.Review = (*models.ReviewResult)(msg)

	case ErrMsg:
		m.Loading = false
		m.Error = msg
	}

	// Update inputs based on mode
	if m.Mode == AddFoodView {
		m.FoodInput, cmd = m.FoodInput.Update(msg)
	} else if m.Mode == AddWaterView {
		m.WaterInput, cmd = m.WaterInput.Update(msg)
	} else if m.Mode == EditFoodPreviewView {
		m.EditInput, cmd = m.EditInput.Update(msg)
	}

	return m, cmd
}

func (m *Model) setupEditInput() {
	m.EditInput.Focus()
	switch m.EditField {
	case 0:
		m.EditInput.SetValue(fmt.Sprintf("%.0f", m.PendingFood.Calories))
	case 1:
		m.EditInput.SetValue(fmt.Sprintf("%.1f", m.PendingFood.Protein))
	case 2:
		m.EditInput.SetValue(fmt.Sprintf("%.1f", m.PendingFood.Carbs))
	case 3:
		m.EditInput.SetValue(fmt.Sprintf("%.1f", m.PendingFood.Fat))
	}
}

func (m *Model) updatePendingFoodFromEdit() {
	val, _ := strconv.ParseFloat(m.EditInput.Value(), 64)
	switch m.EditField {
	case 0:
		m.PendingFood.Calories = val
	case 1:
		m.PendingFood.Protein = val
	case 2:
		m.PendingFood.Carbs = val
	case 3:
		m.PendingFood.Fat = val
	}
}

func (m Model) getStatsCmd() tea.Cmd {
	return func() tea.Msg {
		stats, err := m.Tracker.GetDailyStats(time.Now())
		if err != nil {
			return ErrMsg(err)
		}
		return StatsMsg(stats)
	}
}

func (m Model) parseFoodCmd(desc string) tea.Cmd {
	return func() tea.Msg {
		preview, err := m.Tracker.ParseFood(desc)
		if err != nil {
			return ErrMsg(err)
		}
		return FoodParsedMsg(preview)
	}
}

func (m Model) saveFoodCmd(preview *models.FoodPreview) tea.Cmd {
	return func() tea.Msg {
		err := m.Tracker.SaveFood(preview)
		if err != nil {
			return ErrMsg(err)
		}
		return FoodSavedMsg{}
	}
}

func (m Model) addWaterCmd(amount float64) tea.Cmd {
	return func() tea.Msg {
		err := m.Tracker.AddWater(amount)
		if err != nil {
			return ErrMsg(err)
		}
		return WaterMsg{}
	}
}

func (m Model) runReviewCmd() tea.Cmd {
	return func() tea.Msg {
		res, err := m.Tracker.RunReview()
		if err != nil {
			return ErrMsg(err)
		}
		return ReviewMsg(res)
	}
}

func (m Model) getTodayLogCmd() tea.Cmd {
	return func() tea.Msg {
		entries, err := m.Tracker.GetTodayFoodEntries()
		if err != nil {
			return ErrMsg(err)
		}
		return TodayLogMsg(entries)
	}
}
