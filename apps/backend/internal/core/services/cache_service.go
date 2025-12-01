package services

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// CacheService provides in-memory caching functionality
type CacheService struct {
	mu      sync.RWMutex
	data    map[string]*CacheEntry
	ttl     time.Duration
	maxSize int
}

// CacheEntry represents a single cache entry
type CacheEntry struct {
	Value      interface{}
	ExpiresAt  time.Time
	AccessedAt time.Time
	Size       int
}

// NewCacheService creates a new cache service
func NewCacheService(ttl time.Duration, maxSize int) *CacheService {
	cs := &CacheService{
		data:    make(map[string]*CacheEntry),
		ttl:     ttl,
		maxSize: maxSize,
	}
	
	// Start cleanup goroutine
	go cs.cleanupExpired()
	
	return cs
}

// Get retrieves a value from cache
func (cs *CacheService) Get(key string) (interface{}, bool) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	
	entry, exists := cs.data[key]
	if !exists {
		return nil, false
	}
	
	// Check if expired
	if time.Now().After(entry.ExpiresAt) {
		return nil, false
	}
	
	// Update access time
	entry.AccessedAt = time.Now()
	
	return entry.Value, true
}

// Set stores a value in cache
func (cs *CacheService) Set(key string, value interface{}) {
	cs.SetWithTTL(key, value, cs.ttl)
}

// SetWithTTL stores a value in cache with custom TTL
func (cs *CacheService) SetWithTTL(key string, value interface{}, ttl time.Duration) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	
	// Calculate size (simplified - in production would need better size calculation)
	size := 1
	if data, err := json.Marshal(value); err == nil {
		size = len(data)
	}
	
	// Check if we need to evict entries
	if len(cs.data) >= cs.maxSize {
		cs.evictLRU()
	}
	
	cs.data[key] = &CacheEntry{
		Value:      value,
		ExpiresAt:  time.Now().Add(ttl),
		AccessedAt: time.Now(),
		Size:       size,
	}
}

// Delete removes a value from cache
func (cs *CacheService) Delete(key string) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	
	delete(cs.data, key)
}

// Clear removes all values from cache
func (cs *CacheService) Clear() {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	
	cs.data = make(map[string]*CacheEntry)
}

// GetString retrieves a string value from cache
func (cs *CacheService) GetString(key string) (string, bool) {
	value, exists := cs.Get(key)
	if !exists {
		return "", false
	}
	
	str, ok := value.(string)
	return str, ok
}

// GetJSON retrieves and unmarshals a JSON value from cache
func (cs *CacheService) GetJSON(key string, target interface{}) error {
	value, exists := cs.Get(key)
	if !exists {
		return fmt.Errorf("key not found in cache")
	}
	
	// If value is already the correct type, return it
	if data, ok := value.([]byte); ok {
		return json.Unmarshal(data, target)
	}
	
	// Try to marshal and unmarshal to convert
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	
	return json.Unmarshal(data, target)
}

// SetJSON marshals and stores a JSON value in cache
func (cs *CacheService) SetJSON(key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	
	cs.Set(key, data)
	return nil
}

// GetOrSet retrieves a value from cache or sets it if not found
func (cs *CacheService) GetOrSet(key string, getter func() (interface{}, error)) (interface{}, error) {
	// Try to get from cache first
	if value, exists := cs.Get(key); exists {
		return value, nil
	}
	
	// Get fresh value
	value, err := getter()
	if err != nil {
		return nil, err
	}
	
	// Store in cache
	cs.Set(key, value)
	
	return value, nil
}

// GetStats returns cache statistics
func (cs *CacheService) GetStats() map[string]interface{} {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	
	totalSize := 0
	expiredCount := 0
	now := time.Now()
	
	for _, entry := range cs.data {
		totalSize += entry.Size
		if now.After(entry.ExpiresAt) {
			expiredCount++
		}
	}
	
	return map[string]interface{}{
		"entries":       len(cs.data),
		"total_size":    totalSize,
		"expired_count": expiredCount,
		"max_size":      cs.maxSize,
		"ttl_seconds":   cs.ttl.Seconds(),
	}
}

// evictLRU removes the least recently used entry
func (cs *CacheService) evictLRU() {
	var oldestKey string
	var oldestTime time.Time
	
	for key, entry := range cs.data {
		if oldestKey == "" || entry.AccessedAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.AccessedAt
		}
	}
	
	if oldestKey != "" {
		delete(cs.data, oldestKey)
	}
}

// cleanupExpired periodically removes expired entries
func (cs *CacheService) cleanupExpired() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		cs.mu.Lock()
		now := time.Now()
		
		for key, entry := range cs.data {
			if now.After(entry.ExpiresAt) {
				delete(cs.data, key)
			}
		}
		
		cs.mu.Unlock()
	}
}

// CacheKeyBuilder helps build consistent cache keys
type CacheKeyBuilder struct {
	parts []string
}

// NewCacheKeyBuilder creates a new cache key builder
func NewCacheKeyBuilder() *CacheKeyBuilder {
	return &CacheKeyBuilder{
		parts: []string{},
	}
}

// Add adds a part to the cache key
func (b *CacheKeyBuilder) Add(part string) *CacheKeyBuilder {
	b.parts = append(b.parts, part)
	return b
}

// AddUUID adds a UUID to the cache key
func (b *CacheKeyBuilder) AddUUID(id uuid.UUID) *CacheKeyBuilder {
	b.parts = append(b.parts, id.String())
	return b
}

// Build builds the final cache key
func (b *CacheKeyBuilder) Build() string {
	return fmt.Sprintf("%s", joinStrings(b.parts, ":"))
}

// Helper function to join strings
func joinStrings(parts []string, separator string) string {
	result := ""
	for i, part := range parts {
		if i > 0 {
			result += separator
		}
		result += part
	}
	return result
}

// Common cache key prefixes
const (
	CacheKeyPrefixUser        = "user"
	CacheKeyPrefixTenant      = "tenant"
	CacheKeyPrefixConsent     = "consent"
	CacheKeyPrefixDSR         = "dsr"
	CacheKeyPrefixAnalytics   = "analytics"
	CacheKeyPrefixTranslation = "translation"
	CacheKeyPrefixCookie      = "cookie"
	CacheKeyPrefixPurpose     = "purpose"
)

