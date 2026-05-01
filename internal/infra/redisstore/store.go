package redisstore

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Store struct {
	client *redis.Client
}

func New(client *redis.Client) *Store {
	return &Store{client: client}
}

func (s *Store) Increment(ctx context.Context, key string, expiration time.Duration) (int64, error) {
	count, err := s.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	if count == 1 {
		if err := s.client.Expire(ctx, key, expiration).Err(); err != nil {
			return 0, err
		}
	}

	return count, nil
}

func (s *Store) Block(ctx context.Context, key string, expiration time.Duration) error {
	return s.client.Set(ctx, s.blockKey(key), "1", expiration).Err()
}

func (s *Store) IsBlocked(ctx context.Context, key string) (bool, error) {
	exists, err := s.client.Exists(ctx, s.blockKey(key)).Result()
	if err != nil {
		return false, err
	}

	return exists > 0, nil
}

func (s *Store) blockKey(key string) string {
	return key + ":blocked"
}
