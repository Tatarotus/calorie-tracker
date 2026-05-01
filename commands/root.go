package commands

import (
	"fmt"
	"log"
	"os"

	"calorie-tracker/config"
	"calorie-tracker/db"
	"calorie-tracker/services"
	"calorie-tracker/tui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "calorie-tracker",
	Short: "A smart CLI tool to track your daily nutrition and water intake",
	Run: func(cmd *cobra.Command, args []string) {
		// Run TUI by default
		runTUI()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runTUI() {
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

// initDBAndTracker is a helper for child commands
func initDBAndTracker() (*db.DB, *services.TrackerService) {
	cfg := config.Load()
	database, err := db.NewDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	var llm *services.LLMService
	if cfg.SambaAPIKey != "" {
		llm = services.NewLLMService(cfg)
	}
	
	tracker := services.NewTrackerService(database, llm)
	return database, tracker
}
