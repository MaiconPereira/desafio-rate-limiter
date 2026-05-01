package limiter

import (
	"context"
	"sync"
	"time"
)

type MemoryStore struct {
	mu      sync.Mutex
	now     func() time.Time
	items   map[string]memoryItem
	blocked map[string]time.Time
}

type memoryItem struct {
	count     int64
	expiresAt time.Time
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		now:     time.Now,
		items:   make(map[string]memoryItem),
		blocked: make(map[string]time.Time),
	}
}

func (m *MemoryStore) Increment(_ context.Context, key string, expiration time.Duration) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := m.now()
	item, ok := m.items[key]
	if !ok || !item.expiresAt.After(now) {
		item = memoryItem{expiresAt: now.Add(expiration)}
	}
	item.count++
	m.items[key] = item

	return item.count, nil
}

func (m *MemoryStore) Block(_ context.Context, key string, expiration time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.blocked[key] = m.now().Add(expiration)
	return nil
}

func (m *MemoryStore) IsBlocked(_ context.Context, key string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	expiresAt, ok := m.blocked[key]
	if !ok {
		return false, nil
	}
	if !expiresAt.After(m.now()) {
		delete(m.blocked, key)
		return false, nil
	}

	return true, nil
}
