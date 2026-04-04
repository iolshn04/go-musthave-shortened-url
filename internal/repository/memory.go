package repository

import (
	"context"
	"sync"

	"github.com/iolshn04/go-musthave-shortened-url/internal/model"
)

type memoryRecord struct {
	UserID      string
	OriginalURL string
	Deleted     bool
}
type memoryStorage struct {
	data map[string]memoryRecord
	mu   sync.Mutex
}

// NewMemoryStorage создаёт реализацию репозитория
// с хранением данных в оперативной памяти.
func NewMemoryStorage() Repository {
	return &memoryStorage{data: make(map[string]memoryRecord)}
}

func (m *memoryStorage) Save(ctx context.Context, userID, shortID, original string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[shortID] = memoryRecord{UserID: userID, OriginalURL: original}
	return nil
}

func (m *memoryStorage) SaveBatch(ctx context.Context, userID string, data map[string]string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for short, original := range data {
		m.data[short] = memoryRecord{UserID: userID, OriginalURL: original}
	}
	return nil
}

func (m *memoryStorage) Get(ctx context.Context, shortID string) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	url, ok := m.data[shortID]
	if !ok {
		return "", ErrNotFound
	}
	if url.Deleted {
		return "", ErrDeleted
	}
	return url.OriginalURL, nil
}

func (m *memoryStorage) GetByUser(
	ctx context.Context,
	userID string,
) ([]model.UserURL, error) {

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	result := make([]model.UserURL, 0)
	for short, record := range m.data {
		if record.UserID == userID {
			result = append(result, model.UserURL{
				ShortURL:    short,
				OriginalURL: record.OriginalURL,
			})
		}
	}

	return result, nil
}

func (m *memoryStorage) Ping(ctx context.Context) error {
	return nil
}

func (m *memoryStorage) MarkDeleted(ctx context.Context, userID string, shortIDs []string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for _, id := range shortIDs {
		rec, ok := m.data[id]
		if ok && rec.UserID == userID {
			rec.Deleted = true
			m.data[id] = rec
		}
	}
	return nil
}

func (m *memoryStorage) GetStats(ctx context.Context) (int, int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	urls := len(m.data)

	usersMap := make(map[string]struct{})
	for _, rec := range m.data {
		usersMap[rec.UserID] = struct{}{}
	}

	return urls, len(usersMap), nil
}
