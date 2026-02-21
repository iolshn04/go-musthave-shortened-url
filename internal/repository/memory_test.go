package repository

import (
	"context"
	"testing"
)

func TestMemoryStorage_SaveAndGet(t *testing.T) {
	repo := NewMemoryStorage()
	id := "abc123"
	url := "https://example.com"
	userID := "user123"

	if err := repo.Save(context.Background(), userID, id, url); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := repo.Get(context.Background(), id)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != url {
		t.Errorf("expected %q, got %q", url, got)
	}
}

func TestMemoryStorage_SaveBatch(t *testing.T) {
	repo := NewMemoryStorage()
	userID := "user123"
	data := map[string]string{
		"id1": "https://google.com",
		"id2": "https://yandex.ru",
	}

	err := repo.SaveBatch(context.Background(), userID, data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for id, url := range data {
		got, err := repo.Get(context.Background(), id)
		if err != nil {
			t.Fatalf("unexpected error for %s: %v", id, err)
		}

		if got != url {
			t.Errorf("for %s expected %q, got %q", id, url, got)
		}
	}
}

func TestMemoryStorage_NotFound(t *testing.T) {
	repo := NewMemoryStorage()
	_, err := repo.Get(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestMemoryStorage_GetByUser(t *testing.T) {
	repo := NewMemoryStorage()
	ctx := context.Background()
	user1 := "user1"
	user2 := "user2"

	_ = repo.Save(ctx, user1, "id1", "https://google.com")
	_ = repo.Save(ctx, user1, "id2", "https://yandex.ru")
	_ = repo.Save(ctx, user2, "id3", "https://example.com")

	urls, err := repo.GetByUser(ctx, user1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(urls) != 2 {
		t.Fatalf("expected 2 urls, got %d", len(urls))
	}

	shorts := map[string]bool{}
	for _, u := range urls {
		shorts[u.ShortURL] = true
	}

	if !shorts["id1"] || !shorts["id2"] {
		t.Errorf("expected user1 urls id1 and id2")
	}
}
