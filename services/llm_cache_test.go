package services

import (
	"calorie-tracker/models"
	"fmt"
	"testing"
	"time"
)

func TestLLMCacheBasicOperations(t *testing.T) {
	cache := NewLLMCache(1*time.Hour, 100)

	// Test Set and Get
	food := models.ReferenceFood{
		Name:         "test food",
		BaseQuantity: 100,
		Unit:         "gram",
		Macros:       models.Macros{Calories: 100, Protein: 10, Carbs: 20, Fat: 5},
	}

	cache.Set("test food", food)

	retrieved, found := cache.Get("test food")
	if !found {
		t.Fatal("Expected to find cached food")
	}
	if retrieved.Name != "test food" {
		t.Errorf("Expected name 'test food', got %q", retrieved.Name)
	}
	if retrieved.Macros.Calories != 100 {
		t.Errorf("Expected 100 calories, got %f", retrieved.Macros.Calories)
	}
}

func TestLLMCacheCaseInsensitive(t *testing.T) {
	cache := NewLLMCache(1*time.Hour, 100)

	food := models.ReferenceFood{
		Name:         "Apple",
		BaseQuantity: 100,
		Unit:         "gram",
		Macros:       models.Macros{Calories: 52},
	}

	cache.Set("Apple", food)

	// Should find with different case
	retrieved, found := cache.Get("apple")
	if !found {
		t.Fatal("Expected to find cached food with different case")
	}
	if retrieved.Name != "Apple" {
		t.Errorf("Expected name 'Apple', got %q", retrieved.Name)
	}
}

func TestLLMCacheTTLExpiration(t *testing.T) {
	cache := NewLLMCache(50*time.Millisecond, 100)

	food := models.ReferenceFood{
		Name:         "expiring food",
		BaseQuantity: 100,
		Unit:         "gram",
		Macros:       models.Macros{Calories: 100},
	}

	cache.Set("expiring food", food)

	// Should find immediately
	_, found := cache.Get("expiring food")
	if !found {
		t.Fatal("Expected to find food before expiration")
	}

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	_, found = cache.Get("expiring food")
	if found {
		t.Error("Expected food to be expired")
	}
}

func TestLLMCacheMaxSize(t *testing.T) {
	cache := NewLLMCache(1*time.Hour, 3)

	// Add 3 items
	for i := 1; i <= 3; i++ {
		food := models.ReferenceFood{
			Name:         fmt.Sprintf("food%d", i),
			BaseQuantity: 100,
			Unit:         "gram",
			Macros:       models.Macros{Calories: float64(i * 100)},
		}
		cache.Set(fmt.Sprintf("food%d", i), food)
	}

	if cache.Size() != 3 {
		t.Fatalf("Expected 3 entries, got %d", cache.Size())
	}

	// Add a 4th item - should evict the oldest
	food4 := models.ReferenceFood{
		Name:         "food4",
		BaseQuantity: 100,
		Unit:         "gram",
		Macros:       models.Macros{Calories: 400},
	}
	cache.Set("food4", food4)

	if cache.Size() != 3 {
		t.Fatalf("Expected 3 entries after eviction, got %d", cache.Size())
	}
}

func TestLLMCacheDelete(t *testing.T) {
	cache := NewLLMCache(1*time.Hour, 100)

	food := models.ReferenceFood{
		Name:         "delete me",
		BaseQuantity: 100,
		Unit:         "gram",
		Macros:       models.Macros{Calories: 100},
	}

	cache.Set("delete me", food)
	cache.Delete("delete me")

	_, found := cache.Get("delete me")
	if found {
		t.Error("Expected food to be deleted")
	}
}

func TestLLMCacheClear(t *testing.T) {
	cache := NewLLMCache(1*time.Hour, 100)

	for i := 1; i <= 5; i++ {
		food := models.ReferenceFood{
			Name:         fmt.Sprintf("food%d", i),
			BaseQuantity: 100,
			Unit:         "gram",
			Macros:       models.Macros{Calories: float64(i * 100)},
		}
		cache.Set(fmt.Sprintf("food%d", i), food)
	}

	cache.Clear()

	if cache.Size() != 0 {
		t.Fatalf("Expected 0 entries after clear, got %d", cache.Size())
	}
}

func TestLLMCacheStats(t *testing.T) {
	cache := NewLLMCache(1*time.Hour, 100)

	food := models.ReferenceFood{
		Name:         "stats food",
		BaseQuantity: 100,
		Unit:         "gram",
		Macros:       models.Macros{Calories: 100},
	}

	cache.Set("stats food", food)

	// Access multiple times
	cache.Get("stats food")
	cache.Get("stats food")
	cache.Get("stats food")

	stats := cache.Stats()
	if stats.TotalEntries != 1 {
		t.Errorf("Expected 1 entry, got %d", stats.TotalEntries)
	}
	if stats.TotalAccesses != 3 {
		t.Errorf("Expected 3 accesses, got %d", stats.TotalAccesses)
	}
}

func TestLLMCacheGetOrCompute(t *testing.T) {
	cache := NewLLMCache(1*time.Hour, 100)

	computeCalled := 0
	compute := func() (*models.ReferenceFood, error) {
		computeCalled++
		return &models.ReferenceFood{
			Name:         "computed food",
			BaseQuantity: 100,
			Unit:         "gram",
			Macros:       models.Macros{Calories: 200},
		}, nil
	}

	// First call should compute
	result1, err := cache.GetOrCompute("computed food", compute)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if computeCalled != 1 {
		t.Errorf("Expected compute to be called once, got %d", computeCalled)
	}

	// Second call should use cache
	result2, err := cache.GetOrCompute("computed food", compute)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if computeCalled != 1 {
		t.Errorf("Expected compute to still be called once, got %d", computeCalled)
	}

	if result1.Macros.Calories != result2.Macros.Calories {
		t.Error("Expected cached result to match computed result")
	}
}

func TestLLMCacheAccessCount(t *testing.T) {
	cache := NewLLMCache(1*time.Hour, 100)

	food := models.ReferenceFood{
		Name:         "access test",
		BaseQuantity: 100,
		Unit:         "gram",
		Macros:       models.Macros{Calories: 100},
	}

	cache.Set("access test", food)

	// Access multiple times
	for i := 0; i < 5; i++ {
		cache.Get("access test")
	}

	stats := cache.Stats()
	if stats.TotalAccesses != 5 {
		t.Errorf("Expected 5 accesses, got %d", stats.TotalAccesses)
	}
}
