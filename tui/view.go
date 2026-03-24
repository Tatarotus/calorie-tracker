package tui

import (
	"calorie-tracker/models"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	var s string

	s += StyleTitle.Render("CALORIE TRACKER PRO") + "\n\n"

	if m.Loading {
		msg := "⏳ Loading..."
		switch m.Mode {
		case AddFoodView:
			msg = "🤖 Analyzing your meal..."
		case ReviewView:
			msg = "📊 Generating your review..."
		case ConfirmFoodView:
			msg = "💾 Saving to database..."
		case SetGoalView:
			msg = "🎯 Setting your goal..."
		}
		s += msg + "\n"
	} else if m.Error != nil {
		s += StyleError.Render(fmt.Sprintf("Error: %v", m.Error)) + "\n"
	} else {
		switch m.Mode {
		case DashboardView:
			s += m.dashboardView()
		case AddFoodView:
			s += m.addFoodView()
		case AddWaterView:
			s += m.addWaterView()
		case ReviewView:
			s += m.reviewView()
		case ConfirmFoodView:
			s += m.confirmFoodView()
		case TodayLogView:
			s += m.todayLogView()
		case WeekLogView:
			s += m.weekLogView()
		case MonthLogView:
			s += m.monthLogView()
		case EditFoodPreviewView:
			s += m.editFoodPreviewView()
		case SetGoalView:
			s += m.setGoalView()
		}
	}

	s += "\n" + m.helpView()
	return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, s)
}

func (m Model) setGoalView() string {
	return StyleSection.Render(
		StyleHeader.Render("🎯 Set Your Health Goal") + "\n\n" +
			m.GoalInput.View(),
	)
}

func (m Model) dashboardView() string {
	var goalInfo string
	if m.GoalDescription != "" && m.GoalDescription != "No goal set" {
		goalInfo = m.renderGoalComparison() + "\n\n"
	}

	statsContent := fmt.Sprintf(
		"%sDaily Totals (%s)\n\n"+
			"Calories: %s\n"+
			"Protein:  %s g\n"+
			"Carbs:    %s g\n"+
			"Fat:      %s g\n"+
			"Water:    %s ml",
		goalInfo,
		m.Stats.Date,
		StyleStats.Render(fmt.Sprintf("%.0f", m.Stats.Calories)),
		StyleStats.Render(fmt.Sprintf("%.1f", m.Stats.Protein)),
		StyleStats.Render(fmt.Sprintf("%.1f", m.Stats.Carbs)),
		StyleStats.Render(fmt.Sprintf("%.1f", m.Stats.Fat)),
		StyleStats.Render(fmt.Sprintf("%.0f", m.Stats.WaterML)),
	)

	var recentContent string
	if len(m.RecentLog) > 0 {
		recentContent = "\n\n" + StyleHeader.Render("🕒 Recent Entries") + "\n"
		// Show last 3 entries
		max := 3
		if len(m.RecentLog) < max {
			max = len(m.RecentLog)
		}
		for i := 0; i < max; i++ {
			e := m.RecentLog[i]
			recentContent += fmt.Sprintf("• %s (%s kcal)\n", e.Description, StyleStats.Render(fmt.Sprintf("%.0f", e.Calories)))
		}
	}

	return StyleSection.Render(statsContent + recentContent)
}

func (m Model) renderGoalComparison() string {
	// Simple heuristic for calorie goal extraction
	re := regexp.MustCompile(`(\d+)\s*kcal`)
	match := re.FindStringSubmatch(strings.ToLower(m.GoalDescription))
	if len(match) > 1 {
		goalKcal, _ := strconv.ParseFloat(match[1], 64)
		diff := m.Stats.Calories - goalKcal
		diffStr := fmt.Sprintf("%.0f", diff)
		if diff > 0 {
			diffStr = "+" + diffStr
		}
		return fmt.Sprintf("🎯 Goal: %.0f kcal/day\n📊 Today: %.0f kcal (%s)", goalKcal, m.Stats.Calories, diffStr)
	}

	// Simple weight goal extraction
	reWeight := regexp.MustCompile(`reach\s*(\d+)\s*kg`)
	matchW := reWeight.FindStringSubmatch(strings.ToLower(m.GoalDescription))
	if len(matchW) > 1 {
		goalKg := matchW[1]
		return fmt.Sprintf("🎯 Goal: Reach %s kg\n📊 Current progress in AI Review (r)", goalKg)
	}

	return fmt.Sprintf("🎯 Goal: %s", m.GoalDescription)
}

func (m Model) addFoodView() string {
	return StyleSection.Render(
		StyleHeader.Render("Log Food (Describe what you ate)") + "\n\n" +
			m.FoodInput.View(),
	)
}

func (m Model) addWaterView() string {
	return StyleSection.Render(
		StyleHeader.Render("Log Water (ml)") + "\n\n" +
			m.WaterInput.View(),
	)
}

func (m Model) confirmFoodView() string {
	p := m.PendingFood
	content := fmt.Sprintf(
		"🍽 Parsed Entry:\n\n"+
			"Description: %s\n"+
			"Calories:    %s kcal\n"+
			"Protein:     %.1fg\n"+
			"Carbs:       %.1fg\n"+
			"Fat:         %.1fg",
		StyleHighlight.Render(p.Description),
		StyleSuccess.Render(fmt.Sprintf("%.0f", p.Calories)),
		p.Protein, p.Carbs, p.Fat,
	)
	return StyleSection.Render(content)
}

func (m Model) editFoodPreviewView() string {
	fields := []string{"Calories", "Protein", "Carbs", "Fat"}
	var sb strings.Builder
	sb.WriteString(StyleHeader.Render("Edit Macros") + "\n\n")
	for i, f := range fields {
		if i == m.EditField {
			sb.WriteString(StyleHighlight.Render(f+":") + " " + m.EditInput.View() + "\n")
		} else {
			sb.WriteString(f + ": ...\n")
		}
	}
	return StyleSection.Render(sb.String())
}

func (m Model) renderTodayLogString() string {
	var sb strings.Builder
	
	var total float64
	if len(m.TodayLog) == 0 {
		sb.WriteString("No entries yet today.")
	} else {
		for _, e := range m.TodayLog {
			sb.WriteString(fmt.Sprintf("- %s → %s kcal\n", e.Description, StyleStats.Render(fmt.Sprintf("%.0f", e.Calories))))
			total += e.Calories
		}
		sb.WriteString("\n" + StyleHeader.Render(fmt.Sprintf("Total: %.0f kcal", total)))
	}
	return lipgloss.NewStyle().Width(58).Render(sb.String())
}

func (m Model) todayLogView() string {
	header := StyleHeader.Render("📅 Today's Food Log") + "\n\n"
	return StyleSection.Render(header + m.Viewport.View())
}

func (m Model) renderWeekLogString() string {
	return m.renderRangeLogString(m.WeekLog, "Last 7 Days")
}

func (m Model) weekLogView() string {
	header := StyleHeader.Render("📅 Weekly Food Log (Last 7 Days)") + "\n\n"
	return StyleSection.Render(header + m.Viewport.View())
}

func (m Model) renderMonthLogString() string {
	return m.renderRangeLogString(m.MonthLog, "Last 30 Days")
}

func (m Model) monthLogView() string {
	header := StyleHeader.Render("📅 Monthly Food Log (Last 30 Days)") + "\n\n"
	return StyleSection.Render(header + m.Viewport.View())
}

func (m Model) renderRangeLogString(entries []models.FoodEntry, title string) string {
	var sb strings.Builder
	
	if len(entries) == 0 {
		sb.WriteString("No entries found in this range.")
	} else {
		currentDate := ""
		var dayTotal float64
		for _, e := range entries {
			date := e.Timestamp.Local().Format("2006-01-02")
			if date != currentDate {
				if currentDate != "" {
					sb.WriteString(fmt.Sprintf("  "+StyleBold.Render("Subtotal: %.0f kcal")+"\n\n", dayTotal))
				}
				sb.WriteString(StyleHeader.Render(date) + "\n")
				currentDate = date
				dayTotal = 0
			}
			sb.WriteString(fmt.Sprintf("• %s (%s kcal)\n", e.Description, StyleStats.Render(fmt.Sprintf("%.0f", e.Calories))))
			dayTotal += e.Calories
		}
		sb.WriteString(fmt.Sprintf("  "+StyleBold.Render("Subtotal: %.0f kcal")+"\n", dayTotal))
	}
	return lipgloss.NewStyle().Width(58).Render(sb.String())
}

func (m Model) renderReviewString() string {
	if m.Review == nil {
		return "Starting AI Review..."
	}

	r := m.Review
	var sb strings.Builder
	
	if r.GoalProgress != "" {
		sb.WriteString(StyleHeader.Render("🎯 Progress Towards Goal") + "\n")
		sb.WriteString(r.GoalProgress + "\n\n")
	}

	sb.WriteString(StyleHeader.Render("Summary") + "\n")
	sb.WriteString(r.Summary + "\n\n")
	
	if len(r.Issues) > 0 {
		sb.WriteString(StyleHeader.Render("Issues Found") + "\n")
		for _, i := range r.Issues {
			sb.WriteString(StyleError.Render(" • ") + i + "\n")
		}
		sb.WriteString("\n")
	}
	
	if len(r.Patterns) > 0 {
		sb.WriteString(StyleHeader.Render("Patterns Identified") + "\n")
		for _, p := range r.Patterns {
			sb.WriteString(" • " + p + "\n")
		}
		sb.WriteString("\n")
	}
	
	if len(r.Suggestions) > 0 {
		sb.WriteString(StyleHeader.Render("Suggestions") + "\n")
		for _, s := range r.Suggestions {
			sb.WriteString(StyleSuccess.Render(" • ") + s + "\n")
		}
	}

	return lipgloss.NewStyle().Width(58).Render(sb.String())
}

func (m Model) reviewView() string {
	if m.Review == nil {
		return StyleSection.Render("📊 Generating your review...")
	}

	r := m.Review
	scoreStyle := StyleSuccess
	if r.Score < 50 {
		scoreStyle = StyleError
	} else if r.Score < 80 {
		scoreStyle = StyleWarning
	}

	stickyHeader := fmt.Sprintf(
		"%s\nScore: %s | Progress: %s\n%s\n",
		StyleHeader.Render("AI PROGRESS REVIEW"),
		scoreStyle.Render(fmt.Sprintf("%d/100", r.Score)),
		StyleHighlight.Render(strings.ToUpper(r.Progress)),
		strings.Repeat("─", 58),
	)

	return StyleSection.Render(stickyHeader + m.Viewport.View())
}

func (m Model) helpView() string {
	var help string
	switch m.Mode {
	case ConfirmFoodView:
		help = "y: confirm • n: discard • e: edit • q: quit"
	case EditFoodPreviewView:
		help = "enter: next/save • esc: cancel • q: quit"
	case AddFoodView, AddWaterView, SetGoalView:
		help = "enter: submit • esc: back • q: quit"
	case ReviewView, TodayLogView, WeekLogView, MonthLogView:
		help = "↑/↓: scroll • d: dashboard • t: today • 7: week • m: month • q: quit"
	default:
		help = "d: dashboard • a: add food • w: add water • g: goal • t: today • 7: week • m: month • r: review • q: quit"
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#777777")).Render(help)
}
