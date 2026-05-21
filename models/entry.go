package models

import "time"

type FoodEntry struct {
	ID              int64     `json:"id"`
	Timestamp       time.Time `json:"timestamp"`
	Description     string    `json:"description"`
	Calories        float64   `json:"calories"`
	Protein         float64   `json:"protein"`
	Carbs           float64   `json:"carbs"`
	Fat             float64   `json:"fat"`
	OriginalQuery   string    `json:"original_query,omitempty"`
	NormalizedQuery string    `json:"normalized_query,omitempty"`
	CanonicalKey    string    `json:"canonical_key,omitempty"`
	ResolutionTrace string    `json:"resolution_trace,omitempty"`
}

type FoodPreview struct {
	Name            string           `json:"name"`
	Unit            string           `json:"unit"`
	Description     string           `json:"description"`
	Calories        float64          `json:"calories"`
	Protein         float64          `json:"protein"`
	Carbs           float64          `json:"carbs"`
	Fat             float64          `json:"fat"`
	UserEdited      bool             `json:"user_edited,omitempty"`
	Confidence      float64          `json:"confidence,omitempty"`
	ResolutionTrace *ResolutionTrace `json:"resolution_trace,omitempty"`
}

type FoodType string

const (
	FoodTypeBeverage     FoodType = "beverage"
	FoodTypeGrain        FoodType = "grain"
	FoodTypeProtein      FoodType = "protein"
	FoodTypeDairy        FoodType = "dairy"
	FoodTypeComposite    FoodType = "composite"
	FoodTypeFruit        FoodType = "fruit"
	FoodTypeVegetable    FoodType = "vegetable"
	FoodTypePreparedMeal FoodType = "prepared_meal"
)

type ResolutionTrace struct {
	ParserUsed          string   `json:"parser_used"`
	CanonicalKey        string   `json:"canonical_key"`
	ResolutionMethod    string   `json:"resolution_method"`
	SourceType          string   `json:"source_type"`
	SourceConfidence    float64  `json:"source_confidence"`
	ValidationTriggered bool     `json:"validation_triggered"`
	ValidationWarnings  []string `json:"validation_warnings"`
	ValidationResult    string   `json:"validation_result"`
	StaleCacheUsed      bool     `json:"stale_cache_used"`
	CacheHit            bool     `json:"cache_hit"`
	FatSecretQueried    bool     `json:"fatsecret_queried"`
	SerpAPIFallback     bool     `json:"serpapi_fallback"`
}

type CanonicalFood struct {
	ID                   int64     `json:"id"`
	CanonicalName        string    `json:"canonical_name"`
	NormalizedName       string    `json:"normalized_name"`
	AliasesJSON          string    `json:"aliases_json"` // JSON string list of synonyms
	Language             string    `json:"language"`
	Category             string    `json:"category"`
	FoodType             string    `json:"food_type"`
	CompositionHints     string    `json:"composition_hints"`
	DefaultServingAmount float64   `json:"default_serving_amount"`
	DefaultServingUnit   string    `json:"default_serving_unit"`
	DensityMultiplier    float64   `json:"density_multiplier"`
	GramsPerML           float64   `json:"grams_per_ml"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

type NutritionCacheEntry struct {
	ID               int64     `json:"id"`
	CanonicalFoodID  int64     `json:"canonical_food_id"`
	ServingAmount    float64   `json:"serving_amount"`
	ServingUnit      string    `json:"serving_unit"`
	Calories         float64   `json:"calories"`
	Protein          float64   `json:"protein"`
	Carbs            float64   `json:"carbs"`
	Fat              float64   `json:"fat"`
	Fiber            float64   `json:"fiber"`
	SourceType       string    `json:"source_type"`
	SourceConfidence float64   `json:"source_confidence"`
	SourceReference  string    `json:"source_reference"`
	ResolutionMethod string    `json:"resolution_method"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type UserOverrideEntry struct {
	ID              int64     `json:"id"`
	CanonicalFoodID int64     `json:"canonical_food_id"`
	ServingAmount   float64   `json:"serving_amount"`
	ServingUnit     string    `json:"serving_unit"`
	Calories        float64   `json:"calories"`
	Protein         float64   `json:"protein"`
	Carbs           float64   `json:"carbs"`
	Fat             float64   `json:"fat"`
	Fiber           float64   `json:"fiber"`
	OverrideReason  string    `json:"override_reason"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
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

type ReferenceFood struct {
	Name         string  `json:"name"`
	BaseQuantity float64 `json:"base_quantity"`
	Unit         string  `json:"unit"`
	Macros       Macros  `json:"macros"`
}

type Macros struct {
	Calories float64 `json:"calories"`
	Protein  float64 `json:"protein"`
	Carbs    float64 `json:"carbs"`
	Fat      float64 `json:"fat"`
}

type WaterEntrySimple struct {
	Date     string  `json:"date"`
	AmountML float64 `json:"amount_ml"`
}

type ReviewData struct {
	Goal         string             `json:"goal"`
	Days         []DailyStats       `json:"days"`
	FoodEntries  []FoodEntrySimple  `json:"food_entries"`
	WaterEntries []WaterEntrySimple `json:"water_entries"`
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
