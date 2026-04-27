package tui

import (
	"calorie-tracker/models"
	"fmt"
	"strconv"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type (
	StatsMsg           models.DailyStats
	FoodParsedMsg      *models.FoodPreview
	FoodSavedMsg       struct{}
	WaterMsg           struct{}
	GoalSavedMsg       struct{}
	UndoMsg            struct{}
	GoalDescriptionMsg string
	ReviewMsg          *models.ReviewResult
	TodayLogMsg        []models.FoodEntry
	WeekLogMsg         []models.FoodEntry
	MonthLogMsg        []models.FoodEntry
	RecentLogMsg       []models.FoodEntry
	ErrMsg             error
)

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.getStatsCmd(), m.getGoalCmd())
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.handleWindowSize(msg)
	case tea.KeyMsg:
		model, cmd := m.handleKeyMsg(msg)
		if cmd != nil {
			return model, cmd
		}
		m = model
	case StatsMsg:
		m.Stats = models.DailyStats(msg)
		m.Loading = false
	case GoalDescriptionMsg:
		m.GoalDescription = string(msg)
	case FoodParsedMsg:
		m.Loading = false
		m.PendingFood = (*models.FoodPreview)(msg)
		m.Mode = ConfirmFoodView
	case FoodSavedMsg:
		m.Loading = false
		m.Mode = DashboardView
		return m, m.getStatsCmd()
	case UndoMsg:
		m.Loading = false
		return m, m.getStatsCmd()
	case GoalSavedMsg:
		m.Loading = false
		m.Mode = DashboardView
		return m, tea.Batch(m.getStatsCmd(), m.getGoalCmd())
	case TodayLogMsg:
		m.Loading = false
		m.TodayLog = []models.FoodEntry(msg)
		m.updateViewportContent(m.renderTodayLogString())
	case WeekLogMsg:
		m.Loading = false
		m.WeekLog = []models.FoodEntry(msg)
		m.updateViewportContent(m.renderWeekLogString())
	case MonthLogMsg:
		m.Loading = false
		m.MonthLog = []models.FoodEntry(msg)
		m.updateViewportContent(m.renderMonthLogString())
	case RecentLogMsg:
		m.RecentLog = []models.FoodEntry(msg)
	case WaterMsg:
		m.Loading = false
		m.Mode = DashboardView
		return m, m.getStatsCmd()
	case ReviewMsg:
		m.Loading = false
		m.Review = (*models.ReviewResult)(msg)
		m.updateViewportContent(m.renderReviewString())
	case ErrMsg:
		m.Loading = false
		m.Error = msg
	}

	return m.updateInputs(msg)
}

func (m Model) handleWindowSize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.Width = msg.Width
	m.Height = msg.Height

	vpWidth := m.Width - 4
	if vpWidth > 60 {
		vpWidth = 60
	}
	m.Viewport.Width = vpWidth
	m.Viewport.Height = m.Height - 16
	if m.Viewport.Height < 5 {
		m.Viewport.Height = 5
	}
	return m, nil
}

func (m Model) handleKeyMsg(msg tea.KeyMsg) (Model, tea.Cmd) {
	if m.isLogOrReviewView() {
		return m.handleLogViewKeys(msg)
	}

	switch m.Mode {
	case AddFoodView, AddWaterView, SetGoalView:
		return m.handleInputModeKeys(msg)
	case ConfirmFoodView:
		return m.handleConfirmFoodKeys(msg)
	case EditFoodPreviewView:
		return m.handleEditPreviewKeys(msg)
	default:
		return m.handleDashboardKeys(msg)
	}
}

func (m Model) isLogOrReviewView() bool {
	return m.Mode == ReviewView || m.Mode == TodayLogView ||
		m.Mode == WeekLogView || m.Mode == MonthLogView
}

func (m Model) handleLogViewKeys(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "d":
		m.Mode = DashboardView
		return m, m.getStatsCmd()
	case "t":
		m.Mode = TodayLogView
		m.Loading = true
		return m, m.getTodayLogCmd()
	case "7":
		m.Mode = WeekLogView
		m.Loading = true
		return m, m.getWeekLogCmd()
	case "m":
		m.Mode = MonthLogView
		m.Loading = true
		return m, m.getMonthLogCmd()
	}

	var cmd tea.Cmd
	m.Viewport, cmd = m.Viewport.Update(msg)
	return m, cmd
}

func (m Model) handleInputModeKeys(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.Mode = DashboardView
		return m, m.getStatsCmd()
	case "enter":
		return m.handleInputModeEnter()
	}
	return m, nil
}

func (m Model) handleInputModeEnter() (Model, tea.Cmd) {
	switch m.Mode {
	case AddFoodView:
		m.Loading = true
		m.Error = nil
		return m, m.parseFoodCmd(m.FoodInput.Value())
	case AddWaterView:
		m.Loading = true
		m.Error = nil
		amount, _ := strconv.ParseFloat(m.WaterInput.Value(), 64)
		return m, m.addWaterCmd(amount)
	case SetGoalView:
		m.Loading = true
		m.Error = nil
		return m, m.saveGoalCmd(m.GoalInput.Value())
	}
	return m, nil
}

func (m Model) handleConfirmFoodKeys(msg tea.KeyMsg) (Model, tea.Cmd) {
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
		return m, noOpCmd()
	case "ctrl+c", "q":
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) handleEditPreviewKeys(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		m.updatePendingFoodFromEdit()
		if m.EditField < 3 {
			m.EditField++
			m.setupEditInput()
			return m, nil
		}
		m.Mode = ConfirmFoodView
		return m, noOpCmd()
	case "esc":
		m.Mode = ConfirmFoodView
		return m, noOpCmd()
	case "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) handleDashboardKeys(msg tea.KeyMsg) (Model, tea.Cmd) {
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
		return m, noOpCmd()
	case "w":
		m.Mode = AddWaterView
		m.WaterInput.Focus()
		m.WaterInput.SetValue("")
		return m, noOpCmd()
	case "r":
		m.Mode = ReviewView
		m.Loading = true
		return m, m.runReviewCmd()
	case "t":
		m.Mode = TodayLogView
		m.Loading = true
		return m, m.getTodayLogCmd()
	case "7":
		m.Mode = WeekLogView
		m.Loading = true
		return m, m.getWeekLogCmd()
	case "m":
		m.Mode = MonthLogView
		m.Loading = true
		return m, m.getMonthLogCmd()
	case "g":
		m.Mode = SetGoalView
		m.GoalInput.Focus()
		m.GoalInput.SetValue("")
		return m, noOpCmd()
	case "u":
		m.Loading = true
		return m, m.removeLastEntryCmd()
	}
	return m, nil
}

// noOpCmd returns a command that does nothing.
// Used to signal that a key was handled but no action is needed.
func noOpCmd() tea.Cmd {
	return func() tea.Msg { return nil }
}

func (m Model) updateInputs(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch m.Mode {
	case AddFoodView:
		m.FoodInput, cmd = m.FoodInput.Update(msg)
		cmds = append(cmds, cmd)
	case AddWaterView:
		m.WaterInput, cmd = m.WaterInput.Update(msg)
		cmds = append(cmds, cmd)
	case SetGoalView:
		m.GoalInput, cmd = m.GoalInput.Update(msg)
		cmds = append(cmds, cmd)
	case EditFoodPreviewView:
		m.EditInput, cmd = m.EditInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}
func (m *Model) updateViewportContent(content string) {
	m.Viewport.SetContent(content)
	m.Viewport.GotoTop()
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
	return tea.Batch(
		func() tea.Msg {
			stats, err := m.Tracker.GetDailyStats(time.Now())
			if err != nil {
				return ErrMsg(err)
			}
			return StatsMsg(stats)
		},
		func() tea.Msg {
			recent, _ := m.Tracker.GetFoodEntriesRange(1) // last 24h
			return RecentLogMsg(recent)
		},
	)
}

func (m Model) getGoalCmd() tea.Cmd {
	return func() tea.Msg {
		goal, err := m.Tracker.GetGoal()
		if err != nil {
			return ErrMsg(err)
		}
		return GoalDescriptionMsg(goal)
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

func (m Model) saveGoalCmd(desc string) tea.Cmd {
	return func() tea.Msg {
		err := m.Tracker.SetGoal(desc)
		if err != nil {
			return ErrMsg(err)
		}
		return GoalSavedMsg{}
	}
}

func (m Model) removeLastEntryCmd() tea.Cmd {
	return func() tea.Msg {
		err := m.Tracker.RemoveLastEntry()
		if err != nil {
			return ErrMsg(err)
		}
		return UndoMsg{}
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

func (m Model) getWeekLogCmd() tea.Cmd {
	return func() tea.Msg {
		entries, err := m.Tracker.GetFoodEntriesRange(7)
		if err != nil {
			return ErrMsg(err)
		}
		return WeekLogMsg(entries)
	}
}

func (m Model) getMonthLogCmd() tea.Cmd {
	return func() tea.Msg {
		entries, err := m.Tracker.GetFoodEntriesRange(30)
		if err != nil {
			return ErrMsg(err)
		}
		return MonthLogMsg(entries)
	}
}
