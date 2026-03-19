package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	var s string

	s += StyleTitle.Render("CALORIE TRACKER PRO") + "\n\n"

	if m.Loading {
		s += "Loading... (LLM analysis in progress)\n"
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
		case EditFoodPreviewView:
			s += m.editFoodPreviewView()
		}
	}

	s += "\n" + m.helpView()
	return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, s)
}

func (m Model) dashboardView() string {
	content := fmt.Sprintf(
		"Daily Totals (%s)\n\n"+
			"Calories: %s\n"+
			"Protein:  %s g\n"+
			"Carbs:    %s g\n"+
			"Fat:      %s g\n"+
			"Water:    %s ml",
		m.Stats.Date,
		StyleStats.Render(fmt.Sprintf("%.0f", m.Stats.Calories)),
		StyleStats.Render(fmt.Sprintf("%.1f", m.Stats.Protein)),
		StyleStats.Render(fmt.Sprintf("%.1f", m.Stats.Carbs)),
		StyleStats.Render(fmt.Sprintf("%.1f", m.Stats.Fat)),
		StyleStats.Render(fmt.Sprintf("%.0f", m.Stats.WaterML)),
	)
	return StyleSection.Render(content)
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

func (m Model) todayLogView() string {
	var sb strings.Builder
	sb.WriteString(StyleHeader.Render("📅 Today's Food Log") + "\n\n")
	
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
	return StyleSection.Render(sb.String())
}

func (m Model) reviewView() string {
	if m.Review == nil {
		return "Starting AI Review..."
	}

	r := m.Review
	var sb strings.Builder
	sb.WriteString(StyleHeader.Render("AI Progress Review") + "\n\n")
	
	scoreStyle := StyleSuccess
	if r.Score < 50 {
		scoreStyle = StyleError
	} else if r.Score < 80 {
		scoreStyle = StyleWarning
	}

	sb.WriteString(fmt.Sprintf("Score: %s | Progress: %s\n\n", scoreStyle.Render(fmt.Sprintf("%d/100", r.Score)), r.Progress))
	sb.WriteString("Summary: " + r.Summary + "\n\n")
	
	sb.WriteString("Patterns Identified:\n")
	for _, p := range r.Patterns {
		sb.WriteString(" - " + p + "\n")
	}
	
	sb.WriteString("\nSuggestions:\n")
	for _, s := range r.Suggestions {
		sb.WriteString(" - " + s + "\n")
	}

	return StyleSection.Width(60).Render(sb.String())
}

func (m Model) helpView() string {
	var help string
	switch m.Mode {
	case ConfirmFoodView:
		help = "y: confirm • n: discard • e: edit • q: quit"
	case EditFoodPreviewView:
		help = "enter: next/save • esc: cancel • q: quit"
	case AddFoodView, AddWaterView:
		help = "enter: submit • esc: back • q: quit"
	default:
		help = "d: dashboard • a: add food • w: add water • t: today log • r: review • q: quit"
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#777777")).Render(help)
}
