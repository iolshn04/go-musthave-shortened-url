package repository

import (
	"context"
	"sync"
)

type memoryStorage struct {
	data map[string]string
	mu   sync.Mutex
}

func NewMemoryStorage() Repository {
	return &memoryStorage{data: make(map[string]string)}
}

func (m *memoryStorage) Save(ctx context.Context, id, original string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[id] = original
	return nil
}

func (m *memoryStorage) Get(ctx context.Context, id string) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	url, ok := m.data[id]
	if !ok {
		return "", ErrNotFound
	}
	return url, nil
}

func (m *memoryStorage) Ping(ctx context.Context) error {
	return nil
}
