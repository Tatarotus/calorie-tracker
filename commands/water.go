package commands

import (
	"fmt"
	"log"
	"strconv"

	"github.com/spf13/cobra"
)

var waterCmd = &cobra.Command{
	Use:   "water [amount_in_ml]",
	Short: "Add water intake in ml",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		amountStr := args[0]
		amount, err := strconv.ParseFloat(amountStr, 64)
		if err != nil {
			log.Fatalf("Invalid amount: %s. Must be a number.", amountStr)
		}

		database, tracker := initDBAndTracker()
		defer database.Close()

		err = tracker.AddWater(amount)
		if err != nil {
			log.Fatalf("Error saving water: %v", err)
		}

		fmt.Printf("Successfully added %.0f ml of water.\n", amount)
	},
}

func init() {
	rootCmd.AddCommand(waterCmd)
}
