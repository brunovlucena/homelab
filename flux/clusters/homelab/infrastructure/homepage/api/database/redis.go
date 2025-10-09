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
		return nil, fmt.Errorf("redis URL is empty")
	}

	// Parse Redis URL
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	// Create Redis client
	client := redis.NewClient(opts)

	// Retry connection with exponential backoff
	maxRetries := 30
	for i := 0; i < maxRetries; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		if err := client.Ping(ctx).Err(); err != nil {
			cancel()
			log.Printf("⏳ Redis connection attempt %d/%d failed: %v", i+1, maxRetries, err)
			if i < maxRetries-1 {
				time.Sleep(time.Duration(i+1) * 2 * time.Second)
			}
		} else {
			cancel()
			log.Println("✅ Redis connected successfully")
			return client, nil
		}
	}

	return nil, fmt.Errorf("failed to connect to Redis after %d attempts", maxRetries)
}
