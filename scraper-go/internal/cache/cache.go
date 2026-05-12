package cache

import "time"

type Cache interface {
	Get(key string, target any) (bool, error)
	Set(key string, value any, ttl time.Duration) error
	Delete(key string) error
	Clear() error
}
