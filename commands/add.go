package commands

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add [food description]",
	Short: "Add a new food entry",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		description := strings.Join(args, " ")
		
		database, tracker := initDBAndTracker()
		defer database.Close()

		fmt.Printf("Analyzing: %s...\n", description)
		preview, err := tracker.ParseFood(description)
		if err != nil {
			log.Fatalf("Error analyzing food: %v", err)
		}

		err = tracker.SaveFood(preview)
		if err != nil {
			log.Fatalf("Error saving food: %v", err)
		}

		fmt.Printf("Successfully added: %s\n", preview.Description)
		fmt.Printf("Calories: %.0f kcal | Protein: %.1fg | Carbs: %.1fg | Fat: %.1fg\n", 
			preview.Calories, preview.Protein, preview.Carbs, preview.Fat)
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}
