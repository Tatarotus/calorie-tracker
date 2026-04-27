package tui

import (
	"calorie-tracker/models"
	"calorie-tracker/services"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
)

type ViewMode int

const (
	DashboardView ViewMode = iota
	AddFoodView
	AddWaterView
	ReviewView
	ConfirmFoodView
	TodayLogView
	WeekLogView
	MonthLogView
	EditFoodPreviewView
	SetGoalView
)

type Model struct {
	Tracker         *services.TrackerService
	Mode            ViewMode
	Stats           models.DailyStats
	FoodInput       textinput.Model
	WaterInput      textinput.Model
	GoalInput       textinput.Model
	EditInput       textinput.Model
	Viewport        viewport.Model
	EditField       int // 0: Cal, 1: Pro, 2: Carb, 3: Fat
	Review          *models.ReviewResult
	PendingFood     *models.FoodPreview
	TodayLog        []models.FoodEntry
	WeekLog         []models.FoodEntry
	MonthLog        []models.FoodEntry
	RecentLog       []models.FoodEntry
	GoalDescription string
	Loading         bool
	Error           error
	Width           int
	Height          int
}

func NewModel(tracker *services.TrackerService) Model {
	fi := textinput.New()
	fi.Placeholder = "e.g. 2 eggs and a coffee"
	fi.Focus()

	wi := textinput.New()
	wi.Placeholder = "e.g. 500"

	gi := textinput.New()
	gi.Placeholder = "e.g. I want to reach 80kg in 8 months"

	ei := textinput.New()

	vp := viewport.New(0, 0)

	return Model{
		Tracker:    tracker,
		Mode:       DashboardView,
		FoodInput:  fi,
		WaterInput: wi,
		GoalInput:  gi,
		EditInput:  ei,
		Viewport:   vp,
	}
}

var (
	StyleTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1).
			MarginBottom(1)

	StyleHeader = lipgloss.NewStyle().
			Bold(true).
			Underline(true).
			MarginBottom(1)

	StyleSection = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#000000")).
			Padding(1).
			MarginBottom(1)

	StyleStats = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575")).
			Bold(true)

	StyleError = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true)

	StyleWarning = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFA500")).
			Bold(true)

	StyleSuccess = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00")).
			Bold(true)

	StyleHighlight = lipgloss.NewStyle().
			Background(lipgloss.Color("#3C3C3C")).
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true)

	StyleBold = lipgloss.NewStyle().Bold(true)
)
