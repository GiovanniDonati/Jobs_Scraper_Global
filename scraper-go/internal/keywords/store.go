package keywords

import (
	"context"
	"encoding/json"
	"os"

	"github.com/redis/go-redis/v9"
)

type Store struct {
	redis *redis.Client
}

func NewStore(redis *redis.Client) *Store {
	return &Store{
		redis: redis,
	}
}

func (s *Store) redisKey() string {
	key := os.Getenv("KEYWORDS_REDIS_KEY")

	if key != "" {
		return key
	}

	prefix := os.Getenv("REDIS_KEY_PREFIX")

	if prefix == "" {
		prefix = "vagas-full"
	}

	return prefix + ":keywords"
}

func (s *Store) Load(
	ctx context.Context,
) ([]string, error) {

	fallback := LoadDefaultKeywords()

	raw, err := s.redis.Get(ctx, s.redisKey()).Result()

	if err == redis.Nil {
		return fallback, nil
	}

	if err != nil {
		return fallback, err
	}

	var keywords []string

	if err := json.Unmarshal([]byte(raw), &keywords); err != nil {
		return fallback, nil
	}

	return NormalizeKeywords(keywords), nil
}

func (s *Store) Save(
	ctx context.Context,
	keywords []string,
) error {

	normalized := NormalizeKeywords(keywords)

	payload, err := json.Marshal(normalized)

	if err != nil {
		return err
	}

	return s.redis.Set(
		ctx,
		s.redisKey(),
		payload,
		0,
	).Err()
}
