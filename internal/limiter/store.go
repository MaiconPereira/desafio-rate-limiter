package limiter

import (
	"context"
	"time"
)

type Store interface {
	Increment(ctx context.Context, key string, expiration time.Duration) (int64, error)
	Block(ctx context.Context, key string, expiration time.Duration) error
	IsBlocked(ctx context.Context, key string) (bool, error)
}
