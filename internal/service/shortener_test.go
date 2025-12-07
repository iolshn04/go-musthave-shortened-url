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
	if m.saveErr != nil {
		return m.saveErr
	}
	m.data[id] = original
	return nil
}

func (m *mockRepo) Get(ctx context.Context, id string) (string, error) {
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

	id, err := s.Shorten(context.Background(), "https://yandex.ru")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := s.GetOriginal(context.Background(), id)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "https://yandex.ru" {
		t.Errorf("expected %q, got %q", "https://yandex.ru", got)
	}
}

func TestShortenerService_SaveError(t *testing.T) {
	repo := &mockRepo{data: make(map[string]string), saveErr: errors.New("boom")}
	s := NewShortenerService(repo)

	_, err := s.Shorten(context.Background(), "https://yandex.ru")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
