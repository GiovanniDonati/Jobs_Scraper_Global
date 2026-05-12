package cache

import (
	"encoding/json"
	"sync"
	"time"
)

type memoryEntry struct {
	Value     []byte
	ExpiresAt time.Time
}

type MemoryCache struct {
	mu    sync.RWMutex
	store map[string]memoryEntry
}

func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		store: make(map[string]memoryEntry),
	}
}

func (m *MemoryCache) Get(
	key string,
	target any,
) (bool, error) {
	m.mu.RLock()
	entry, exists := m.store[key]
	m.mu.RUnlock()

	if !exists {
		return false, nil
	}

	if time.Now().After(entry.ExpiresAt) {
		_ = m.Delete(key)
		return false, nil
	}

	if err := json.Unmarshal(entry.Value, target); err != nil {
		return false, err
	}

	return true, nil
}

func (m *MemoryCache) Set(
	key string,
	value any,
	ttl time.Duration,
) error {
	payload, err := json.Marshal(value)
	if err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.store[key] = memoryEntry{
		Value:     payload,
		ExpiresAt: time.Now().Add(ttl),
	}

	return nil
}

func (m *MemoryCache) Delete(key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.store, key)

	return nil
}

func (m *MemoryCache) Clear() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.store = make(map[string]memoryEntry)

	return nil
}
