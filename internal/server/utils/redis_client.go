package utils

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

// RedisClient is a package-level Redis client used by rate limiter.
var RedisClient *redis.Client

// InitRedis initializes Redis client from environment variables.
// It prefers REDIS_URL (redis:// or rediss://). If REDIS_URL is
// not provided, it falls back to REDIS_ADDR + REDIS_PASSWORD.
// If no configuration is found, RedisClient stays nil and Redis-based limiting is skipped.
func InitRedis() error {
	// Load .env for local development (ignore errors)
	_ = godotenv.Load()

	// Prefer full URL which may include TLS (rediss://)
	if u := os.Getenv("REDIS_URL"); u != "" {
		opt, err := redis.ParseURL(u)
		if err != nil {
			return fmt.Errorf("invalid REDIS_URL: %w", err)
		}
		client := redis.NewClient(opt)
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := client.Ping(ctx).Err(); err != nil {
			return fmt.Errorf("redis ping failed: %w", err)
		}
		RedisClient = client
		return nil
	}

	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		return nil
	}
	password := os.Getenv("REDIS_PASSWORD")
	opt := &redis.Options{Addr: addr, Password: password}
	client := redis.NewClient(opt)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis ping failed: %w", err)
	}
	RedisClient = client
	return nil
}
