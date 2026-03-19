package models

import "time"

type FoodEntry struct {
	ID          int64     `json:"id"`
	Timestamp   time.Time `json:"timestamp"`
	Description string    `json:"description"`
	Calories    float64   `json:"calories"`
	Protein     float64   `json:"protein"`
	Carbs       float64   `json:"carbs"`
	Fat         float64   `json:"fat"`
}

type FoodPreview struct {
	Description string  `json:"description"`
	Calories    float64 `json:"calories"`
	Protein     float64 `json:"protein"`
	Carbs       float64 `json:"carbs"`
	Fat         float64 `json:"fat"`
}

type WaterEntry struct {
	ID        int64     `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	AmountML  float64   `json:"amount_ml"`
}

type DailyStats struct {
	Date     string  `json:"date"`
	Calories float64 `json:"calories"`
	Protein  float64 `json:"protein"`
	Carbs    float64 `json:"carbs"`
	Fat      float64 `json:"fat"`
	WaterML  float64 `json:"water_ml"`
}

type Goal struct {
	ID          int64     `json:"id"`
	Timestamp   time.Time `json:"timestamp"`
	Description string    `json:"description"`
}

type FoodEntrySimple struct {
	Date        string  `json:"date"`
	Description string  `json:"description"`
	Calories    float64 `json:"calories"`
	Protein     float64 `json:"protein"`
	Carbs       float64 `json:"carbs"`
	Fat         float64 `json:"fat"`
}

type ReviewData struct {
	Goal    string            `json:"goal"`
	Days    []DailyStats      `json:"days"`
	Entries []FoodEntrySimple `json:"entries"`
}

type ReviewResult struct {
	Summary      string   `json:"summary"`
	GoalProgress string   `json:"goal_progress"`
	Progress     string   `json:"progress"`
	Score        int      `json:"score"`
	Issues       []string `json:"issues"`
	Suggestions  []string `json:"suggestions"`
	Patterns     []string `json:"patterns"`
}
