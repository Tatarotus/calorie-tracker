package services

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"calorie-tracker/models"
)

func TestLLMCache_SetWithTTL(t *testing.T) {
	cache := NewLLMCache(1*time.Hour, 100)

	food := models.ReferenceFood{
		Name:         "ttl food",
		BaseQuantity: 100,
		Unit:         "gram",
		Macros:       models.Macros{Calories: 100},
	}

	cache.SetWithTTL("ttl food", food, 30*time.Minute)

	retrieved, found := cache.Get("ttl food")
	if !found {
		t.Fatal("Expected to find cached food")
	}
	if retrieved.Name != "ttl food" {
		t.Errorf("Expected name 'ttl food', got %q", retrieved.Name)
	}
}

func TestLLMCache_EvictOldest(t *testing.T) {
	cache := NewLLMCache(1*time.Hour, 2)

	// Add 2 items
	for i := 1; i <= 2; i++ {
		food := models.ReferenceFood{
			Name:         fmt.Sprintf("food%d", i),
			BaseQuantity: 100,
			Unit:         "gram",
			Macros:       models.Macros{Calories: float64(i * 100)},
		}
		cache.Set(fmt.Sprintf("food%d", i), food)
	}

	// Access food1 multiple times to increase its access count
	for i := 0; i < 5; i++ {
		cache.Get("food1")
	}

	// Add a 3rd item - should evict food2 (lower access count)
	food3 := models.ReferenceFood{
		Name:         "food3",
		BaseQuantity: 100,
		Unit:         "gram",
		Macros:       models.Macros{Calories: 300},
	}
	cache.Set("food3", food3)

	// food1 should still be there (higher access count)
	_, found := cache.Get("food1")
	if !found {
		t.Error("Expected food1 to still be in cache (higher access count)")
	}

	// food2 should be evicted
	_, found = cache.Get("food2")
	if found {
		t.Error("Expected food2 to be evicted (lower access count)")
	}
}

func TestLLMCache_Cleanup(t *testing.T) {
	cache := NewLLMCache(50*time.Millisecond, 100)

	food := models.ReferenceFood{
		Name:         "cleanup food",
		BaseQuantity: 100,
		Unit:         "gram",
		Macros:       models.Macros{Calories: 100},
	}

	cache.Set("cleanup food", food)

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	// Run cleanup
	cache.cleanup()

	// Should be removed
	if cache.Size() != 0 {
		t.Errorf("Expected 0 entries after cleanup, got %d", cache.Size())
	}
}

func TestLLMCache_GetOrCompute_Error(t *testing.T) {
	cache := NewLLMCache(1*time.Hour, 100)

	compute := func() (*models.ReferenceFood, error) {
		return nil, errors.New("compute error")
	}

	_, err := cache.GetOrCompute("error food", compute)
	if err == nil {
		t.Error("Expected error from compute function")
	}
	if err.Error() != "compute error" {
		t.Errorf("Expected 'compute error', got %v", err)
	}
}

func TestLLMCache_StatsString(t *testing.T) {
	stats := CacheStats{
		TotalEntries:   5,
		ExpiredEntries: 1,
		TotalAccesses:  10,
	}

	str := stats.String()
	expected := "CacheStats{entries=5, expired=1, accesses=10}"
	if str != expected {
		t.Errorf("Expected %q, got %q", expected, str)
	}
}

func TestLLMCache_NewLLMCache_Defaults(t *testing.T) {
	// Test with zero values - should use defaults
	cache := NewLLMCache(0, 0)

	if cache.defaultTTL != 24*time.Hour {
		t.Errorf("Expected default TTL 24h, got %v", cache.defaultTTL)
	}
	if cache.maxSize != 1000 {
		t.Errorf("Expected default maxSize 1000, got %d", cache.maxSize)
	}
}
