package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// InitializeRedis initializes the Redis connection with retry logic
func InitializeRedis(redisURL string) (*redis.Client, error) {
	if redisURL == "" {
		return nil, fmt.Errorf("redis URL is empty - check REDIS_URL environment variable")
	}

	log.Println("  📡 Parsing Redis URL...")
	// Parse Redis URL
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}
	log.Printf("  ✅ Redis URL parsed (host: %s, db: %d)", opts.Addr, opts.DB)

	// Create Redis client
	log.Println("  🔧 Creating Redis client...")
	client := redis.NewClient(opts)

	// Retry connection with exponential backoff
	log.Println("  🔄 Attempting to connect to Redis (max 30 retries with exponential backoff)...")
	maxRetries := 30
	for i := 0; i < maxRetries; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		if err := client.Ping(ctx).Err(); err != nil {
			cancel()
			log.Printf("  ⏳ Redis connection attempt %d/%d failed: %v", i+1, maxRetries, err)
			if i < maxRetries-1 {
				backoff := time.Duration(i+1) * 2 * time.Second
				log.Printf("  ⏰ Retrying in %v...", backoff)
				time.Sleep(backoff)
			}
		} else {
			cancel()
			log.Printf("  ✅ Redis connected successfully on attempt %d/%d", i+1, maxRetries)
			return client, nil
		}
	}

	return nil, fmt.Errorf("❌ ERROR: failed to connect to Redis after %d attempts - giving up", maxRetries)
}
