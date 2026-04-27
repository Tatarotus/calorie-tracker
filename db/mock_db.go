package db

import (
	"calorie-tracker/models"
	"errors"
	"strings"
	"sync"
	"time"
)

// MockDB is an in-memory implementation of DBProvider for testing
type MockDB struct {
	mu           sync.RWMutex
	foodEntries  []models.FoodEntry
	waterEntries []models.WaterEntry
	goal         *models.Goal
	cache        map[string]models.FoodEntry
	reference    map[string]models.ReferenceFood
	lastRemoved  *models.FoodEntry
	errorOnCall  map[string]error // Track which operations should return errors
}

// NewMockDB creates a new mock database
func NewMockDB() *MockDB {
	return &MockDB{
		cache:       make(map[string]models.FoodEntry),
		reference:   make(map[string]models.ReferenceFood),
		errorOnCall: make(map[string]error),
	}
}

// SetError sets an error to be returned for a specific operation
func (m *MockDB) SetError(operation string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errorOnCall[operation] = err
}

// ClearError clears errors for a specific operation
func (m *MockDB) ClearError(operation string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.errorOnCall, operation)
}

// AddFoodEntry implements DBProvider
func (m *MockDB) AddFoodEntry(entry models.FoodEntry) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.errorOnCall["AddFoodEntry"]; err != nil {
		return err
	}

	m.foodEntries = append(m.foodEntries, entry)
	return nil
}

// GetDailyFoodEntries implements DBProvider
func (m *MockDB) GetDailyFoodEntries(t time.Time) ([]models.FoodEntry, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if err := m.errorOnCall["GetDailyFoodEntries"]; err != nil {
		return nil, err
	}

	dateStr := t.Format("2006-01-02")
	var result []models.FoodEntry
	for _, entry := range m.foodEntries {
		if entry.Timestamp.Format("2006-01-02") == dateStr {
			result = append(result, entry)
		}
	}
	return result, nil
}

// GetFoodEntriesRange implements DBProvider
func (m *MockDB) GetFoodEntriesRange(days int) ([]models.FoodEntry, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if err := m.errorOnCall["GetFoodEntriesRange"]; err != nil {
		return nil, err
	}

	return m.foodEntries, nil
}

// CacheFood implements DBProvider
func (m *MockDB) CacheFood(entry models.FoodEntry) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.errorOnCall["CacheFood"]; err != nil {
		return err
	}

	description := strings.ToLower(strings.TrimSpace(entry.Description))
	m.cache[description] = entry
	return nil
}

// GetCachedFood implements DBProvider
func (m *MockDB) GetCachedFood(name string) (*models.FoodEntry, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if err := m.errorOnCall["GetCachedFood"]; err != nil {
		return nil, err
	}

	description := strings.ToLower(strings.TrimSpace(name))
	entry, ok := m.cache[description]
	if !ok {
		return nil, nil
	}

	return &entry, nil
}

// GetReferenceFood implements DBProvider
func (m *MockDB) GetReferenceFood(name string) (*models.ReferenceFood, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if err := m.errorOnCall["GetReferenceFood"]; err != nil {
		return nil, err
	}

	name = strings.ToLower(strings.TrimSpace(name))
	for refName, ref := range m.reference {
		if refName == name || strings.Contains(name, refName) {
			return &ref, nil
		}
	}

	return nil, nil
}

// SeedReferenceFood adds a reference food to the mock
func (m *MockDB) SeedReferenceFood(f models.ReferenceFood) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.reference[strings.ToLower(f.Name)] = f
}

// AddWaterEntry implements DBProvider
func (m *MockDB) AddWaterEntry(entry models.WaterEntry) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.errorOnCall["AddWaterEntry"]; err != nil {
		return err
	}

	m.waterEntries = append(m.waterEntries, entry)
	return nil
}

// GetDailyWaterEntries implements DBProvider
func (m *MockDB) GetDailyWaterEntries(t time.Time) ([]models.WaterEntry, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if err := m.errorOnCall["GetDailyWaterEntries"]; err != nil {
		return nil, err
	}

	dateStr := t.Format("2006-01-02")
	var result []models.WaterEntry
	for _, entry := range m.waterEntries {
		if entry.Timestamp.Format("2006-01-02") == dateStr {
			result = append(result, entry)
		}
	}
	return result, nil
}

// GetWaterEntriesRange implements DBProvider
func (m *MockDB) GetWaterEntriesRange(days int) ([]models.WaterEntry, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if err := m.errorOnCall["GetWaterEntriesRange"]; err != nil {
		return nil, err
	}

	return m.waterEntries, nil
}

// GetStatsRange implements DBProvider
func (m *MockDB) GetStatsRange(days int) ([]models.DailyStats, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if err := m.errorOnCall["GetStatsRange"]; err != nil {
		return nil, err
	}

	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	rangeStart := todayStart.AddDate(0, 0, -days)

	// Calculate daily stats from food and water entries
	statsMap := make(map[string]*models.DailyStats)

	for _, entry := range m.foodEntries {
		if entry.Timestamp.Before(rangeStart) {
			continue
		}
		dateStr := entry.Timestamp.Format("2006-01-02")
		if _, ok := statsMap[dateStr]; !ok {
			statsMap[dateStr] = &models.DailyStats{Date: dateStr}
		}
		statsMap[dateStr].Calories += entry.Calories
		statsMap[dateStr].Protein += entry.Protein
		statsMap[dateStr].Carbs += entry.Carbs
		statsMap[dateStr].Fat += entry.Fat
	}

	for _, entry := range m.waterEntries {
		if entry.Timestamp.Before(rangeStart) {
			continue
		}
		dateStr := entry.Timestamp.Format("2006-01-02")
		if _, ok := statsMap[dateStr]; !ok {
			statsMap[dateStr] = &models.DailyStats{Date: dateStr}
		}
		statsMap[dateStr].WaterML += entry.AmountML
	}

	var result []models.DailyStats
	for _, stats := range statsMap {
		result = append(result, *stats)
	}

	return result, nil
}

// SetGoal implements DBProvider
func (m *MockDB) SetGoal(goal models.Goal) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.errorOnCall["SetGoal"]; err != nil {
		return err
	}

	m.goal = &goal
	return nil
}

// GetLatestGoal implements DBProvider
func (m *MockDB) GetLatestGoal() (*models.Goal, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if err := m.errorOnCall["GetLatestGoal"]; err != nil {
		return nil, err
	}

	return m.goal, nil
}

// RemoveLastEntry implements DBProvider
func (m *MockDB) RemoveLastEntry() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.errorOnCall["RemoveLastEntry"]; err != nil {
		return err
	}

	if len(m.foodEntries) > 0 {
		lastIdx := len(m.foodEntries) - 1
		m.lastRemoved = &m.foodEntries[lastIdx]
		m.foodEntries = m.foodEntries[:lastIdx]
	}

	return nil
}

// Close implements DBProvider
func (m *MockDB) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.errorOnCall["Close"]; err != nil {
		return err
	}

	return nil
}

// GetFoodEntries returns all food entries (for testing)
func (m *MockDB) GetFoodEntries() []models.FoodEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.foodEntries
}

// GetWaterEntries returns all water entries (for testing)
func (m *MockDB) GetWaterEntries() []models.WaterEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.waterEntries
}

// Clear clears all data (for testing)
func (m *MockDB) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.foodEntries = nil
	m.waterEntries = nil
	m.goal = nil
	m.cache = make(map[string]models.FoodEntry)
	m.lastRemoved = nil
}

// Error for no goal found
var ErrNoGoal = errors.New("no goal found")
