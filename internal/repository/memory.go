package repository

import "sync"

type memoryStorage struct {
	data map[string]string
	mu   sync.RWMutex
}

func NewMemoryStorage() Repository {
	return &memoryStorage{data: make(map[string]string)}
}

func (m *memoryStorage) Save(id, original string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[id] = original
	return nil
}

func (m *memoryStorage) Get(id string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	url, ok := m.data[id]
	if !ok {
		return "", ErrNotFound
	}
	return url, nil
}
