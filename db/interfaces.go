package db

import (
	"calorie-tracker/models"
	"time"
)

// DBProvider defines the interface for database operations
// This allows us to mock the database for testing
type DBProvider interface {
	// Food operations
	AddFoodEntry(entry models.FoodEntry) error
	GetDailyFoodEntries(t time.Time) ([]models.FoodEntry, error)
	GetFoodEntriesRange(days int) ([]models.FoodEntry, error)
	CacheFood(f models.ReferenceFood) error
	GetCachedFood(name string) (*models.ReferenceFood, error)
	GetAllCacheEntries() ([]models.ReferenceFood, error)
	GetReferenceFood(name string) (*models.ReferenceFood, error)

	// Water operations
	AddWaterEntry(entry models.WaterEntry) error
	GetDailyWaterEntries(t time.Time) ([]models.WaterEntry, error)
	GetWaterEntriesRange(days int) ([]models.WaterEntry, error)

	// Stats operations
	GetStatsRange(days int) ([]models.DailyStats, error)

	// Goal operations
	SetGoal(goal models.Goal) error
	GetLatestGoal() (*models.Goal, error)

	// Other operations
	RemoveLastEntry() error
	Close() error

	// Canonical Food operations
	GetCanonicalFood(id int64) (*models.CanonicalFood, error)
	GetCanonicalFoodByName(name string) (*models.CanonicalFood, error)
	SaveCanonicalFood(food *models.CanonicalFood) error
	GetAllCanonicalFoods() ([]models.CanonicalFood, error)

	// Nutrition Cache operations
	GetNutritionCache(canonicalFoodID int64, unit string) (*models.NutritionCacheEntry, error)
	SaveNutritionCache(entry *models.NutritionCacheEntry) error

	// User Override operations
	GetUserOverride(canonicalFoodID int64) (*models.UserOverrideEntry, error)
	SaveUserOverride(override *models.UserOverrideEntry) error
}
