package services

import (
	"calorie-tracker/models"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"
)

// LLMCacheEntry stores a cached LLM response with TTL
type LLMCacheEntry struct {
	Food        models.ReferenceFood
	CreatedAt   time.Time
	ExpiresAt   time.Time
	AccessCount int
}

// LLMCache provides in-memory caching for LLM responses with TTL
// This reduces API calls and improves response times for repeated queries
type LLMCache struct {
	mu         sync.RWMutex
	entries    map[string]*LLMCacheEntry
	defaultTTL time.Duration
	maxSize    int
}

// NewLLMCache creates a new LLM cache with the given default TTL and max size
func NewLLMCache(defaultTTL time.Duration, maxSize int) *LLMCache {
	if defaultTTL <= 0 {
		defaultTTL = 24 * time.Hour
	}
	if maxSize <= 0 {
		maxSize = 1000
	}

	cache := &LLMCache{
		entries:    make(map[string]*LLMCacheEntry),
		defaultTTL: defaultTTL,
		maxSize:    maxSize,
	}

	// Start background cleanup goroutine
	go cache.cleanupLoop()

	return cache
}

// Get retrieves a cached entry if it exists and hasn't expired
func (c *LLMCache) Get(description string) (*models.ReferenceFood, bool) {
	c.mu.RLock()
	entry, exists := c.entries[c.hashKey(description)]
	c.mu.RUnlock()

	if !exists {
		return nil, false
	}

	if time.Now().After(entry.ExpiresAt) {
		c.Delete(description)
		return nil, false
	}

	// Update access count
	c.mu.Lock()
	entry.AccessCount++
	c.mu.Unlock()

	// Return a copy to prevent mutation
	foodCopy := entry.Food
	return &foodCopy, true
}

// Set stores a food in the cache with the default TTL
func (c *LLMCache) Set(description string, food models.ReferenceFood) {
	c.SetWithTTL(description, food, c.defaultTTL)
}

// SetWithTTL stores a food in the cache with a custom TTL
func (c *LLMCache) SetWithTTL(description string, food models.ReferenceFood, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Evict oldest entries if we're at capacity
	if len(c.entries) >= c.maxSize {
		c.evictOldest(1)
	}

	key := c.hashKey(description)
	c.entries[key] = &LLMCacheEntry{
		Food:        food,
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(ttl),
		AccessCount: 0,
	}
}

// Delete removes an entry from the cache
func (c *LLMCache) Delete(description string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.entries, c.hashKey(description))
}

// Clear removes all entries from the cache
func (c *LLMCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]*LLMCacheEntry)
}

// Size returns the current number of cached entries
func (c *LLMCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.entries)
}

// Stats returns cache statistics
func (c *LLMCache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var totalAccesses int
	var expired int
	now := time.Now()

	for _, entry := range c.entries {
		totalAccesses += entry.AccessCount
		if now.After(entry.ExpiresAt) {
			expired++
		}
	}

	return CacheStats{
		TotalEntries:   len(c.entries),
		ExpiredEntries: expired,
		TotalAccesses:  totalAccesses,
	}
}

// CacheStats holds cache statistics
type CacheStats struct {
	TotalEntries   int
	ExpiredEntries int
	TotalAccesses  int
}

// hashKey creates a deterministic hash key from a description
func (c *LLMCache) hashKey(description string) string {
	// Normalize the description for consistent hashing
	normalized := strings.ToLower(strings.TrimSpace(description))
	hash := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(hash[:])
}

// evictOldest removes the oldest entries from the cache
func (c *LLMCache) evictOldest(count int) {
	if len(c.entries) == 0 {
		return
	}

	// Simple eviction: remove entries with lowest access count and oldest creation
	type entryWithKey struct {
		key   string
		entry *LLMCacheEntry
	}

	var entries []entryWithKey
	for k, v := range c.entries {
		entries = append(entries, entryWithKey{key: k, entry: v})
	}

	// Sort by access count (ascending), then by creation time (oldest first)
	for i := 0; i < len(entries)-1; i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[i].entry.AccessCount > entries[j].entry.AccessCount ||
				(entries[i].entry.AccessCount == entries[j].entry.AccessCount &&
					entries[i].entry.CreatedAt.After(entries[j].entry.CreatedAt)) {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}

	for i := 0; i < count && i < len(entries); i++ {
		delete(c.entries, entries[i].key)
	}
}

// cleanupLoop periodically removes expired entries
func (c *LLMCache) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.cleanup()
	}
}

// cleanup removes all expired entries
func (c *LLMCache) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, entry := range c.entries {
		if now.After(entry.ExpiresAt) {
			delete(c.entries, key)
		}
	}
}

// GetOrCompute retrieves from cache or computes and stores the result
func (c *LLMCache) GetOrCompute(description string, compute func() (*models.ReferenceFood, error)) (*models.ReferenceFood, error) {
	// Try cache first
	if cached, found := c.Get(description); found {
		return cached, nil
	}

	// Compute the result
	result, err := compute()
	if err != nil {
		return nil, err
	}

	// Store in cache
	c.Set(description, *result)

	return result, nil
}

// String returns a string representation of the cache stats
func (s CacheStats) String() string {
	return fmt.Sprintf("CacheStats{entries=%d, expired=%d, accesses=%d}",
		s.TotalEntries, s.ExpiredEntries, s.TotalAccesses)
}
