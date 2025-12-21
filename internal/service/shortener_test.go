package service

import (
	"context"
	"sync"
	"testing"

	"github.com/iolshn04/go-musthave-shortened-url/internal/model"
	"github.com/iolshn04/go-musthave-shortened-url/internal/repository"
)

type mockRepo struct {
	data    map[string]mockRecord // shortID -> record
	saveErr error
	mu      sync.Mutex
}

type mockRecord struct {
	UserID      string
	OriginalURL string
}

func newMockRepo() *mockRepo {
	return &mockRepo{data: make(map[string]mockRecord)}
}

func (m *mockRepo) Save(ctx context.Context, userID, shortID, original string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.saveErr != nil {
		return m.saveErr
	}

	m.data[shortID] = mockRecord{
		UserID:      userID,
		OriginalURL: original,
	}
	return nil
}

func (m *mockRepo) SaveBatch(ctx context.Context, userID string, data map[string]string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.saveErr != nil {
		return m.saveErr
	}

	for shortID, original := range data {
		m.data[shortID] = mockRecord{
			UserID:      userID,
			OriginalURL: original,
		}
	}

	return nil
}

func (m *mockRepo) Get(ctx context.Context, shortID string) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	rec, ok := m.data[shortID]
	if !ok {
		return "", repository.ErrNotFound
	}

	return rec.OriginalURL, nil
}

func (m *mockRepo) GetByUser(ctx context.Context, userID string) ([]model.UserURL, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	result := make([]model.UserURL, 0)
	for shortID, rec := range m.data {
		if rec.UserID == userID {
			result = append(result, model.UserURL{
				ShortURL:    shortID,
				OriginalURL: rec.OriginalURL,
			})
		}
	}

	if len(result) == 0 {
		return nil, repository.ErrNotFound
	}

	return result, nil
}

func (m *mockRepo) Ping(ctx context.Context) error {
	return nil
}

func TestShortenerService_ShortenAndGet(t *testing.T) {
	repo := newMockRepo()
	s := NewShortenerService(repo)
	ctx := context.Background()
	userID := "user123"

	id, err := s.Shorten(ctx, userID, "https://yandex.ru")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := s.GetOriginal(ctx, id)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != "https://yandex.ru" {
		t.Errorf("expected %q, got %q", "https://yandex.ru", got)
	}
}

func TestShortenerService_ShortenBatch(t *testing.T) {
	repo := newMockRepo()
	s := NewShortenerService(repo)
	ctx := context.Background()
	userID := "user123"

	input := map[string]string{
		"cid1": "https://google.com",
		"cid2": "https://yandex.ru",
	}

	result, err := s.ShortenBatch(ctx, userID, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != len(input) {
		t.Fatalf("expected %d results, got %d", len(input), len(result))
	}

	for cid, id := range result {
		if cid == "" || id == "" {
			t.Fatal("expected non-empty correlation_id and short id")
		}

		got, err := repo.Get(ctx, id)
		if err != nil {
			t.Fatalf("expected saved value, got error: %v", err)
		}
		if got != input[cid] {
			t.Errorf("expected %q, got %q", input[cid], got)
		}
	}
}

func TestShortenerService_GetByUser(t *testing.T) {
	repo := newMockRepo()
	s := NewShortenerService(repo)
	ctx := context.Background()
	userID := "user123"

	urls := []string{"https://google.com", "https://yandex.ru"}
	for _, u := range urls {
		_, err := s.Shorten(ctx, userID, u)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	result, err := repo.GetByUser(ctx, userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != len(urls) {
		t.Fatalf("expected %d results, got %d", len(urls), len(result))
	}

	seen := map[string]bool{}
	for _, r := range result {
		seen[r.OriginalURL] = true
	}
	for _, u := range urls {
		if !seen[u] {
			t.Errorf("expected URL %q to be present", u)
		}
	}
}

func TestShortenerService_GetByUser_NoContent(t *testing.T) {
	repo := newMockRepo()
	ctx := context.Background()
	userID := "empty_user"

	_, err := repo.GetByUser(ctx, userID)
	if err != repository.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
