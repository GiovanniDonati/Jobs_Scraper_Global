package cache

import (
	"os"

	"github.com/redis/go-redis/v9"
)

func NewCache() Cache {
	redisURL := os.Getenv("REDIS_URL")

	if redisURL == "" {
		return NewMemoryCache()
	}

	client := redis.NewClient(&redis.Options{
		Addr: redisURL,
	})

	return NewRedisCache(client)
}
