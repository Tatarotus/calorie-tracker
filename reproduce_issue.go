package main

import (
	"fmt"
	"os"

	"calorie-tracker/config"
	"calorie-tracker/services"
)

func main() {
	cfg := config.Load()
	if cfg.SambaAPIKey == "" {
		fmt.Println("Error: SAMBA_API_KEY environment variable is not set.")
		os.Exit(1)
	}

	llm := services.NewLLMService(cfg)
	
	descriptions := []string{
		"100g de feijão tropeiro",
		"100g feijão tropeiro",
	}

	for _, desc := range descriptions {
		fmt.Printf("\n--- Analyzing: %s ---\n", desc)
		preview, err := llm.ParseFood(desc)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		fmt.Printf("Parsed successfully:\n")
		fmt.Printf("Calories: %.2f, Protein: %.2f, Carbs: %.2f, Fat: %.2f\n", 
			preview.Calories, preview.Protein, preview.Carbs, preview.Fat)
	}
}
