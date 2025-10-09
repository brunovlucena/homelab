package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// CacheManager handles Redis caching operations
type CacheManager struct {
	client *redis.Client
	ctx    context.Context
}

// NewCacheManager creates a new cache manager instance
func NewCacheManager(client *redis.Client) *CacheManager {
	return &CacheManager{
		client: client,
		ctx:    context.Background(),
	}
}

// CacheConfig defines cache configuration
type CacheConfig struct {
	TTL time.Duration
}

// Default cache TTLs
const (
	SkillsCacheTTL     = 1 * time.Hour      // Skills change rarely
	ProjectsCacheTTL   = 30 * time.Minute   // Projects change occasionally
	ExperienceCacheTTL = 24 * time.Hour     // Experience changes very rarely
	ContentCacheTTL    = 1 * time.Hour      // Content changes occasionally
	AboutCacheTTL      = 24 * time.Hour     // About changes very rarely
	ContactCacheTTL    = 24 * time.Hour     // Contact changes very rarely
)

// Cache keys
const (
	SkillsKey     = "api:skills"
	ProjectsKey   = "api:projects"
	ExperienceKey = "api:experience"
	AboutKey      = "api:about"
	ContactKey    = "api:contact"
	ContentKey    = "api:content:%s" // Format with content key
)

// Set stores data in cache with TTL
func (c *CacheManager) Set(key string, data interface{}, ttl time.Duration) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data for cache: %w", err)
	}

	return c.client.Set(c.ctx, key, jsonData, ttl).Err()
}

// Get retrieves data from cache
func (c *CacheManager) Get(key string, dest interface{}) error {
	val, err := c.client.Get(c.ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("cache miss for key: %s", key)
		}
		return fmt.Errorf("failed to get from cache: %w", err)
	}

	return json.Unmarshal([]byte(val), dest)
}

// Delete removes data from cache
func (c *CacheManager) Delete(key string) error {
	return c.client.Del(c.ctx, key).Err()
}

// DeletePattern removes all keys matching a pattern
func (c *CacheManager) DeletePattern(pattern string) error {
	keys, err := c.client.Keys(c.ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get keys for pattern: %w", err)
	}

	if len(keys) > 0 {
		return c.client.Del(c.ctx, keys...).Err()
	}

	return nil
}

// Exists checks if key exists in cache
func (c *CacheManager) Exists(key string) bool {
	return c.client.Exists(c.ctx, key).Val() > 0
}

// GetOrSet retrieves from cache or sets if not found
func (c *CacheManager) GetOrSet(key string, dest interface{}, ttl time.Duration, fetchFunc func() (interface{}, error)) error {
	// Try to get from cache first
	err := c.Get(key, dest)
	if err == nil {
		return nil // Cache hit
	}

	// Cache miss - fetch data
	data, err := fetchFunc()
	if err != nil {
		return fmt.Errorf("failed to fetch data: %w", err)
	}

	// Set in cache
	if err := c.Set(key, data, ttl); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Warning: failed to set cache for key %s: %v\n", key, err)
	}

	// Set dest to fetched data
	return json.Unmarshal([]byte(fmt.Sprintf("%v", data)), dest)
}

// InvalidateSkillsCache invalidates skills-related cache
func (c *CacheManager) InvalidateSkillsCache() error {
	return c.Delete(SkillsKey)
}

// InvalidateProjectsCache invalidates projects-related cache
func (c *CacheManager) InvalidateProjectsCache() error {
	return c.Delete(ProjectsKey)
}

// InvalidateExperienceCache invalidates experience-related cache
func (c *CacheManager) InvalidateExperienceCache() error {
	return c.Delete(ExperienceKey)
}

// InvalidateContentCache invalidates all content cache
func (c *CacheManager) InvalidateContentCache() error {
	return c.DeletePattern("api:content:*")
}

// InvalidateAllCache invalidates all API cache
func (c *CacheManager) InvalidateAllCache() error {
	return c.DeletePattern("api:*")
}

// Health checks if Redis is available
func (c *CacheManager) Health() error {
	return c.client.Ping(c.ctx).Err()
}

// Stats returns cache statistics
func (c *CacheManager) Stats() (map[string]interface{}, error) {
	info, err := c.client.Info(c.ctx, "memory", "stats").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get Redis info: %w", err)
	}

	// Parse basic stats
	stats := map[string]interface{}{
		"connected": true,
		"info":      info,
	}

	return stats, nil
}
