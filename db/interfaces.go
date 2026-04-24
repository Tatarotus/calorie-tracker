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
	CacheFood(entry models.FoodEntry) error
	GetCachedFood(name string) (*models.FoodEntry, error)

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
}
