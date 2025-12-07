package repository

import (
	"context"
	"testing"
)

func TestMemoryStorage_SaveAndGet(t *testing.T) {
	repo := NewMemoryStorage()
	id := "abc123"
	url := "https://example.com"

	if err := repo.Save(context.Background(), id, url); err != nil {
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
