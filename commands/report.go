package commands

import (
	"fmt"
	"log"
	"time"

	"github.com/spf13/cobra"
)

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "View your daily stats",
	Run: func(cmd *cobra.Command, args []string) {
		database, tracker := initDBAndTracker()
		defer database.Close()

		stats, err := tracker.GetDailyStats(time.Now())
		if err != nil {
			log.Fatalf("Error getting daily stats: %v", err)
		}

		fmt.Printf("--- Daily Report (%s) ---\n", stats.Date)
		fmt.Printf("Calories: %.0f kcal\n", stats.Calories)
		fmt.Printf("Protein:  %.1f g\n", stats.Protein)
		fmt.Printf("Carbs:    %.1f g\n", stats.Carbs)
		fmt.Printf("Fat:      %.1f g\n", stats.Fat)
		fmt.Printf("Water:    %.0f ml\n", stats.WaterML)
	},
}

func init() {
	rootCmd.AddCommand(reportCmd)
}
