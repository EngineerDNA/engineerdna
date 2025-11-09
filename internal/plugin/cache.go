package plugin

import (
	"sync"
	"time"

	sdk "github.com/engineerdna/engineerdna/plugins/plugin-sdk"
)

// MetadataCache provides thread-safe caching of plugin metadata
type MetadataCache struct {
	cache map[string]*CachedMetadata
	mu    sync.RWMutex
	ttl   time.Duration
}

// CachedMetadata stores plugin metadata with expiration
type CachedMetadata struct {
	Info      sdk.PluginInfo
	LoadedAt  time.Time
	ExpiresAt time.Time
}

// NewMetadataCache creates a new plugin metadata cache with 5-minute TTL
func NewMetadataCache() *MetadataCache {
	return &MetadataCache{
		cache: make(map[string]*CachedMetadata),
		ttl:   5 * time.Minute,
	}
}

// NewMetadataCacheWithTTL creates a new cache with custom TTL
func NewMetadataCacheWithTTL(ttl time.Duration) *MetadataCache {
	return &MetadataCache{
		cache: make(map[string]*CachedMetadata),
		ttl:   ttl,
	}
}

// Get retrieves cached metadata if available and not expired
func (c *MetadataCache) Get(pluginName string) (*sdk.PluginInfo, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	cached, exists := c.cache[pluginName]
	if !exists {
		return nil, false
	}

	// Check if expired
	if time.Now().After(cached.ExpiresAt) {
		return nil, false
	}

	return &cached.Info, true
}

// Set stores plugin metadata in cache with expiration
func (c *MetadataCache) Set(pluginName string, info sdk.PluginInfo) {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	c.cache[pluginName] = &CachedMetadata{
		Info:      info,
		LoadedAt:  now,
		ExpiresAt: now.Add(c.ttl),
	}
}

// Invalidate removes a specific plugin from cache
func (c *MetadataCache) Invalidate(pluginName string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.cache, pluginName)
}

// InvalidateAll clears the entire cache
func (c *MetadataCache) InvalidateAll() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache = make(map[string]*CachedMetadata)
}

// CleanExpired removes expired entries from cache
func (c *MetadataCache) CleanExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for name, cached := range c.cache {
		if now.After(cached.ExpiresAt) {
			delete(c.cache, name)
		}
	}
}

// Size returns the current number of cached entries
func (c *MetadataCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.cache)
}
