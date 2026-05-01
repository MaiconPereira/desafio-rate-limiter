package limiter

import (
	"context"
	"testing"
	"time"
)

func TestLimiterBlocksAfterIPLimit(t *testing.T) {
	store := NewMemoryStore()
	rl := New(store, Config{
		IPRequestsPerSecond:    2,
		TokenRequestsPerSecond: 10,
		BlockDuration:          time.Minute,
	})

	subject := Subject{Type: SubjectIP, Value: "127.0.0.1"}

	assertAllowed(t, rl, subject, true)
	assertAllowed(t, rl, subject, true)
	assertAllowed(t, rl, subject, false)
	assertAllowed(t, rl, subject, false)
}

func TestLimiterTokenHasPrecedenceOverIP(t *testing.T) {
	store := NewMemoryStore()
	rl := New(store, Config{
		IPRequestsPerSecond:    1,
		TokenRequestsPerSecond: 3,
		BlockDuration:          time.Minute,
	})

	token := Subject{Type: SubjectToken, Value: "premium-token"}

	assertAllowed(t, rl, token, true)
	assertAllowed(t, rl, token, true)
	assertAllowed(t, rl, token, true)
	assertAllowed(t, rl, token, false)
}

func TestLimiterSpecificTokenOverride(t *testing.T) {
	store := NewMemoryStore()
	rl := New(store, Config{
		IPRequestsPerSecond:    10,
		TokenRequestsPerSecond: 2,
		TokenOverrides: map[string]int64{
			"vip": 4,
		},
		BlockDuration: time.Minute,
	})

	token := Subject{Type: SubjectToken, Value: "vip"}

	assertAllowed(t, rl, token, true)
	assertAllowed(t, rl, token, true)
	assertAllowed(t, rl, token, true)
	assertAllowed(t, rl, token, true)
	assertAllowed(t, rl, token, false)
}

func assertAllowed(t *testing.T, rl *RateLimiter, subject Subject, expected bool) {
	t.Helper()

	allowed, err := rl.Allow(context.Background(), subject)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if allowed != expected {
		t.Fatalf("expected allowed=%v, got %v", expected, allowed)
	}
}
