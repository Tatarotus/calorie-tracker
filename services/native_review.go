package services

import (
	"calorie-tracker/models"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// AnalyzeNativeReview computes a comprehensive, native nutrition review locally without calling an LLM API.
func AnalyzeNativeReview(data models.ReviewData) *models.ReviewResult {
	var totalCalories, totalProtein, totalCarbs, totalFat, totalWater float64
	activeDays := 0

	for _, day := range data.Days {
		if day.Calories > 0 || day.WaterML > 0 {
			activeDays++
		}
		totalCalories += day.Calories
		totalProtein += day.Protein
		totalCarbs += day.Carbs
		totalFat += day.Fat
		totalWater += day.WaterML
	}

	divDays := float64(activeDays)
	if divDays == 0 {
		divDays = 1
	}

	avgCalories := totalCalories / 7.0

	avgCaloriesActive := totalCalories / divDays
	avgProteinActive := totalProtein / divDays
	avgCarbsActive := totalCarbs / divDays
	avgFatActive := totalFat / divDays
	avgWaterActive := totalWater / divDays

	goalType := "maintenance"
	targetKcal := 2200.0
	targetProtein := 100.0

	goalLower := strings.ToLower(data.Goal)

	// Extract calorie target (e.g. "2500 kcal")
	kcalRe := regexp.MustCompile(`(\d+)\s*kcal`)
	if matches := kcalRe.FindStringSubmatch(goalLower); len(matches) > 1 {
		if val, err := strconv.ParseFloat(matches[1], 64); err == nil {
			targetKcal = val
		}
	}

	// Extract weight range target (e.g. "72kg -> 83kg" or "72 to 83")
	weightRangeRe := regexp.MustCompile(`(\d+)\s*(?:kg)?\s*(?:->|to)\s*(\d+)\s*(?:kg)?`)
	if matches := weightRangeRe.FindStringSubmatch(goalLower); len(matches) > 2 {
		w1, _ := strconv.ParseFloat(matches[1], 64)
		w2, _ := strconv.ParseFloat(matches[2], 64)
		if w1 < w2 {
			goalType = "weight_gain"
			if !strings.Contains(goalLower, "kcal") {
				targetKcal = 2800.0
				targetProtein = w1 * 2.0
				if targetProtein < 120.0 {
					targetProtein = 120.0
				}
			}
		} else if w1 > w2 {
			goalType = "weight_loss"
			if !strings.Contains(goalLower, "kcal") {
				targetKcal = 1800.0
				targetProtein = w1 * 2.0
				if targetProtein < 110.0 {
					targetProtein = 110.0
				}
			}
		}
	} else {
		// Single weight target
		weightSingleRe := regexp.MustCompile(`(?:reach|lose|gain|perder|ganhar)\s*(\d+)`)
		if matches := weightSingleRe.FindStringSubmatch(goalLower); len(matches) > 1 {
			if strings.Contains(goalLower, "gain") || strings.Contains(goalLower, "ganhar") || strings.Contains(goalLower, "bulk") {
				goalType = "weight_gain"
				targetKcal = 2800.0
				targetProtein = 130.0
			} else if strings.Contains(goalLower, "lose") || strings.Contains(goalLower, "perder") || strings.Contains(goalLower, "cut") {
				goalType = "weight_loss"
				targetKcal = 1800.0
				targetProtein = 120.0
			}
		}
	}

	var issues []string
	if activeDays < 7 {
		issues = append(issues, fmt.Sprintf("Zero intake recorded on %d out of 7 days", 7-activeDays))
	}

	if goalType == "weight_gain" {
		if avgCaloriesActive < targetKcal-150 {
			issues = append(issues, fmt.Sprintf("Caloric intake on active days is %.0f%% of requirement (avg %.0f vs %.0f kcal target)", (avgCaloriesActive/targetKcal)*100.0, avgCaloriesActive, targetKcal))
		}
	} else if goalType == "weight_loss" {
		if avgCaloriesActive > targetKcal+150 {
			issues = append(issues, fmt.Sprintf("Caloric intake on active days is above deficit target (avg %.0f vs %.0f kcal target)", avgCaloriesActive, targetKcal))
		}
	}

	if avgProteinActive < targetProtein {
		issues = append(issues, fmt.Sprintf("Protein intake is low (avg %.1fg/day on active days vs %.0fg target)", avgProteinActive, targetProtein))
	}

	if avgWaterActive < 2000 {
		issues = append(issues, fmt.Sprintf("Inconsistent hydration (averaging %.0f ml daily on active days)", avgWaterActive))
	}

	if avgFatActive < 45 {
		issues = append(issues, fmt.Sprintf("Very low fat intake: averaging %.1fg/day (fats are essential for hormone levels)", avgFatActive))
	} else if avgFatActive > 110 {
		issues = append(issues, fmt.Sprintf("High fat intake: averaging %.1fg/day (could lead to unwanted fat gain)", avgFatActive))
	}

	junkCount := 0
	bananaCount := 0
	eggCount := 0
	coffeeCount := 0
	for _, e := range data.FoodEntries {
		desc := strings.ToLower(e.Description)
		if strings.Contains(desc, "nescau") || strings.Contains(desc, "bolo") || strings.Contains(desc, "pastel") || strings.Contains(desc, "coxinha") || strings.Contains(desc, "doce") || strings.Contains(desc, "chocolate") || strings.Contains(desc, "frito") || strings.Contains(desc, "frita") || strings.Contains(desc, "broa") {
			junkCount++
		}
		if strings.Contains(desc, "banana") {
			bananaCount++
		}
		if strings.Contains(desc, "egg") || strings.Contains(desc, "ovo") {
			eggCount++
		}
		if strings.Contains(desc, "cafe") || strings.Contains(desc, "café") {
			coffeeCount++
		}
	}

	if junkCount > 3 {
		issues = append(issues, fmt.Sprintf("High frequency of processed/fried foods or sweet drinks (%d entries found)", junkCount))
	}

	var patterns []string
	if activeDays >= 5 {
		patterns = append(patterns, fmt.Sprintf("Solid tracking consistency: logged intake on %d/7 days", activeDays))
	} else {
		patterns = append(patterns, fmt.Sprintf("Inconsistent tracking habits: logged intake on only %d/7 days", activeDays))
	}

	if avgProteinActive > 0 && avgCarbsActive/avgProteinActive > 3.0 {
		patterns = append(patterns, fmt.Sprintf("High carbohydrate-to-protein ratio (%.1f:1)", avgCarbsActive/avgProteinActive))
	}

	if bananaCount > 3 {
		patterns = append(patterns, "Bananas are a frequent carbohydrate and potassium source in your diet")
	}
	if eggCount > 3 {
		patterns = append(patterns, "Eggs are a regular protein staple in your breakfast/snacks")
	}
	if coffeeCount > 3 {
		patterns = append(patterns, "Coffee (with or without milk) is a daily dietary constant")
	}

	var suggestions []string
	if activeDays < 7 {
		suggestions = append(suggestions, "Establish a daily habit of logging every meal to avoid missing logs")
	}
	if goalType == "weight_gain" {
		suggestions = append(suggestions, fmt.Sprintf("Increase daily calories to target surplus (~%.0f kcal) by adding healthy calorie-dense foods like peanut butter, avocados, and nuts", targetKcal))
		suggestions = append(suggestions, fmt.Sprintf("Prioritize protein to reach your %.1fg daily target to support muscle gain (currently %.1fg)", targetProtein, avgProteinActive))
	} else if goalType == "weight_loss" {
		suggestions = append(suggestions, fmt.Sprintf("Ensure you maintain a consistent caloric deficit (target ~%.0f kcal) while keeping protein high", targetKcal))
	}
	if avgWaterActive < 2500 {
		suggestions = append(suggestions, "Drink at least 2.5–3 liters of water daily, especially around workouts")
	}
	if avgFatActive < 50 {
		suggestions = append(suggestions, "Incorporate healthy fats (extra virgin olive oil, nuts, avocados) for hormone regulation")
	}
	if junkCount > 3 {
		suggestions = append(suggestions, "Reduce processed/fried foods and sweet drinks, substituting them with whole complex carbs (sweet potatoes, oats)")
	}

	score := 100
	score -= (7 - activeDays) * 10

	if goalType == "weight_gain" && avgCaloriesActive < targetKcal-100 {
		diff := targetKcal - avgCaloriesActive
		score -= int(diff / 25)
	} else if goalType == "weight_loss" && avgCaloriesActive > targetKcal+100 {
		diff := avgCaloriesActive - targetKcal
		score -= int(diff / 25)
	}

	if avgProteinActive < targetProtein {
		diff := targetProtein - avgProteinActive
		score -= int(diff * 0.5)
	}

	if avgWaterActive < 2000 {
		score -= 10
		if avgWaterActive < 1000 {
			score -= 10
		}
	}

	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	progress := "stable"
	if activeDays < 4 {
		progress = "stable (insufficient data)"
	} else if goalType == "weight_gain" {
		if avgCaloriesActive >= targetKcal-150 && avgProteinActive >= targetProtein-15 && activeDays >= 5 {
			progress = "improving"
		} else if avgCaloriesActive < targetKcal-300 || avgProteinActive < targetProtein-30 {
			progress = "regressing"
		}
	} else if goalType == "weight_loss" {
		if avgCaloriesActive <= targetKcal+150 && activeDays >= 5 {
			progress = "improving"
		} else if avgCaloriesActive > targetKcal+300 {
			progress = "regressing"
		}
	}

	summary := fmt.Sprintf("Your average daily intake is %.0f kcal (%.0f kcal on active days), with %.1fg protein, %.1fg carbs, and %.1fg fat. You logged intake on %d out of 7 days and averaged %.0f ml of water.", avgCalories, avgCaloriesActive, avgProteinActive, avgCarbsActive, avgFatActive, activeDays, avgWaterActive)

	var gp string
	if data.Goal == "No goal set" || data.Goal == "" {
		gp = "No specific fitness goal is set in the app. Set a goal (e.g. '72kg -> 83kg' or '2800 kcal') to get customized progress tracking."
	} else if goalType == "weight_gain" {
		gp = fmt.Sprintf("Your goal is to gain weight (%s). Your average daily intake of %.0f kcal on logged days is %.0f kcal below your target of %.0f kcal. To build muscle, prioritize increasing calorie and protein intake consistently.", data.Goal, avgCaloriesActive, targetKcal-avgCaloriesActive, targetKcal)
	} else if goalType == "weight_loss" {
		gp = fmt.Sprintf("Your goal is to lose weight (%s). Your average daily intake of %.0f kcal is currently evaluated against a target of %.0f kcal to maintain a fat-loss calorie deficit.", data.Goal, avgCaloriesActive, targetKcal)
	} else {
		gp = fmt.Sprintf("Your goal is set to: %s. Your average calorie intake is %.0f kcal.", data.Goal, avgCaloriesActive)
	}

	return &models.ReviewResult{
		Summary:      summary,
		GoalProgress: gp,
		Progress:     progress,
		Score:        score,
		Issues:       issues,
		Suggestions:  suggestions,
		Patterns:     patterns,
	}
}
