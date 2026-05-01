package commands

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/cobra"
)

var reviewCmd = &cobra.Command{
	Use:   "review",
	Short: "Get an AI-powered review of your recent progress",
	Run: func(cmd *cobra.Command, args []string) {
		database, tracker := initDBAndTracker()
		defer database.Close()

		fmt.Println("Analyzing your recent progress... This may take a moment.")
		res, err := tracker.RunReview()
		if err != nil {
			log.Fatalf("Error running review: %v", err)
		}

		fmt.Printf("\n=== AI PROGRESS REVIEW ===\n")
		fmt.Printf("Score: %d/100 | Progress: %s\n", res.Score, strings.ToUpper(res.Progress))
		fmt.Println(strings.Repeat("-", 40))
		
		if res.GoalProgress != "" {
			fmt.Printf("\n🎯 Progress Towards Goal\n%s\n", res.GoalProgress)
		}
		
		fmt.Printf("\n📊 Summary\n%s\n", res.Summary)
		
		if len(res.Issues) > 0 {
			fmt.Printf("\n⚠️ Issues Found\n")
			for _, i := range res.Issues {
				fmt.Printf(" • %s\n", i)
			}
		}
		
		if len(res.Patterns) > 0 {
			fmt.Printf("\n🔍 Patterns Identified\n")
			for _, p := range res.Patterns {
				fmt.Printf(" • %s\n", p)
			}
		}
		
		if len(res.Suggestions) > 0 {
			fmt.Printf("\n💡 Suggestions\n")
			for _, s := range res.Suggestions {
				fmt.Printf(" • %s\n", s)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(reviewCmd)
}
