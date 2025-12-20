package service

import (
	"context"
	"errors"
	"testing"

	"github.com/iolshn04/go-musthave-shortened-url/internal/repository"
)

type mockRepo struct {
	data    map[string]string
	saveErr error
}

func (m *mockRepo) Save(ctx context.Context, id, original string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if m.saveErr != nil {
		return m.saveErr
	}
	m.data[id] = original
	return nil
}

func (m *mockRepo) SaveBatch(ctx context.Context, data map[string]string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if m.saveErr != nil {
		return m.saveErr
	}
	for k, v := range data {
		m.data[k] = v
	}
	return nil
}

func (m *mockRepo) Get(ctx context.Context, id string) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}
	url, ok := m.data[id]
	if !ok {
		return "", repository.ErrNotFound
	}
	return url, nil
}

func (m *mockRepo) Ping(ctx context.Context) error {
	return nil
}

func TestShortenerService_ShortenAndGet(t *testing.T) {
	repo := &mockRepo{data: make(map[string]string)}
	s := NewShortenerService(repo)
	ctx := context.Background()

	id, err := s.Shorten(ctx, "https://yandex.ru")
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

func TestShortenerService_SaveError(t *testing.T) {
	repo := &mockRepo{
		data:    make(map[string]string),
		saveErr: errors.New("boom"),
	}
	s := NewShortenerService(repo)
	ctx := context.Background()

	_, err := s.Shorten(ctx, "https://yandex.ru")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestShortenerService_ShortenBatch(t *testing.T) {
	repo := &mockRepo{data: make(map[string]string)}
	s := NewShortenerService(repo)
	ctx := context.Background()

	input := map[string]string{
		"cid1": "https://google.com",
		"cid2": "https://yandex.ru",
	}

	result, err := s.ShortenBatch(ctx, input)
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

		// Проверяем, что реально сохранилось в repo
		_, err := repo.Get(ctx, id)
		if err != nil {
			t.Fatalf("expected saved value, got error: %v", err)
		}
	}
}

func TestShortenerService_ShortenBatch_SaveError(t *testing.T) {
	repo := &mockRepo{
		data:    make(map[string]string),
		saveErr: errors.New("boom"),
	}
	s := NewShortenerService(repo)
	ctx := context.Background()

	input := map[string]string{
		"cid1": "https://yandex.ru",
	}

	_, err := s.ShortenBatch(ctx, input)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
