package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
	ctx    context.Context
}

func NewRedisCache(client *redis.Client) *RedisCache {
	return &RedisCache{
		client: client,
		ctx:    context.Background(),
	}
}

func (r *RedisCache) Get(
	key string,
	target any,
) (bool, error) {
	value, err := r.client.Get(r.ctx, key).Result()

	if err == redis.Nil {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	if err := json.Unmarshal([]byte(value), target); err != nil {
		return false, err
	}

	return true, nil
}

func (r *RedisCache) Set(
	key string,
	value any,
	ttl time.Duration,
) error {
	payload, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return r.client.Set(
		r.ctx,
		key,
		payload,
		ttl,
	).Err()
}

func (r *RedisCache) Delete(key string) error {
	return r.client.Del(r.ctx, key).Err()
}

func (r *RedisCache) Clear() error {
	return r.client.FlushDB(r.ctx).Err()
}
