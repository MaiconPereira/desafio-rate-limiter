package limiter

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type SubjectType string

const (
	SubjectIP    SubjectType = "ip"
	SubjectToken SubjectType = "token"
)

type Subject struct {
	Type  SubjectType
	Value string
}

type Config struct {
	IPRequestsPerSecond    int64
	TokenRequestsPerSecond int64
	TokenOverrides         map[string]int64
	BlockDuration          time.Duration
	Window                 time.Duration
}

type RateLimiter struct {
	store  Store
	config Config
}

func New(store Store, config Config) *RateLimiter {
	if config.Window <= 0 {
		config.Window = time.Second
	}
	if config.BlockDuration <= 0 {
		config.BlockDuration = 5 * time.Minute
	}
	if config.TokenOverrides == nil {
		config.TokenOverrides = make(map[string]int64)
	}

	return &RateLimiter{
		store:  store,
		config: config,
	}
}

func (r *RateLimiter) Allow(ctx context.Context, subject Subject) (bool, error) {
	if strings.TrimSpace(subject.Value) == "" {
		return false, fmt.Errorf("subject value is required")
	}

	key := r.key(subject)
	blocked, err := r.store.IsBlocked(ctx, key)
	if err != nil {
		return false, err
	}
	if blocked {
		return false, nil
	}

	total, err := r.store.Increment(ctx, key, r.config.Window)
	if err != nil {
		return false, err
	}

	if total > r.limitFor(subject) {
		if err := r.store.Block(ctx, key, r.config.BlockDuration); err != nil {
			return false, err
		}
		return false, nil
	}

	return true, nil
}

func (r *RateLimiter) limitFor(subject Subject) int64 {
	if subject.Type == SubjectToken {
		if limit, ok := r.config.TokenOverrides[subject.Value]; ok {
			return limit
		}
		return r.config.TokenRequestsPerSecond
	}

	return r.config.IPRequestsPerSecond
}

func (r *RateLimiter) key(subject Subject) string {
	return fmt.Sprintf("rate_limit:%s:%s", subject.Type, subject.Value)
}
