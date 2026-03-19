package main

import (
	"fmt"
	"log"
	"os"

	"calorie-tracker/config"
	"calorie-tracker/db"
	"calorie-tracker/services"
	"calorie-tracker/tui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	cfg := config.Load()
	if cfg.SambaAPIKey == "" {
		fmt.Println("Error: SAMBA_API_KEY environment variable is not set.")
		fmt.Println("Please set it to your SambaNova API key.")
		os.Exit(1)
	}

	database, err := db.NewDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	llm := services.NewLLMService(cfg)
	tracker := services.NewTrackerService(database, llm)

	p := tea.NewProgram(tui.NewModel(tracker), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running TUI: %v", err)
		os.Exit(1)
	}
}
