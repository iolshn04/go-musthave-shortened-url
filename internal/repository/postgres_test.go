package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockPostgresRepo struct {
	data    map[string]string
	saveErr error
	getErr  error
	pingErr error
}

func (m *mockPostgresRepo) Save(ctx context.Context, id, original string) error {
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

func (m *mockPostgresRepo) Get(ctx context.Context, id string) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}
	if m.getErr != nil {
		return "", m.getErr
	}
	v, ok := m.data[id]
	if !ok {
		return "", ErrNotFound
	}
	return v, nil
}

func (m *mockPostgresRepo) Ping(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	return m.pingErr
}

func TestPostgresRepository_SaveAndGet(t *testing.T) {
	repo := &mockPostgresRepo{data: make(map[string]string)}
	ctx := context.Background()

	id := "abc123"
	url := "https://example.com"

	err := repo.Save(ctx, id, url)
	assert.NoError(t, err)

	got, err := repo.Get(ctx, id)
	assert.NoError(t, err)
	assert.Equal(t, url, got)
}

func TestPostgresRepository_GetNotFound(t *testing.T) {
	repo := &mockPostgresRepo{data: make(map[string]string)}
	ctx := context.Background()

	_, err := repo.Get(ctx, "missing")
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestPostgresRepository_SaveError(t *testing.T) {
	repo := &mockPostgresRepo{saveErr: errors.New("save fail"), data: make(map[string]string)}
	ctx := context.Background()

	err := repo.Save(ctx, "id1", "url1")
	assert.EqualError(t, err, "save fail")
}

func TestPostgresRepository_Ping(t *testing.T) {
	repo := &mockPostgresRepo{}
	ctx := context.Background()

	err := repo.Ping(ctx)
	assert.NoError(t, err)

	repo.pingErr = errors.New("ping fail")
	err = repo.Ping(ctx)
	assert.EqualError(t, err, "ping fail")
}

func TestPostgresRepository_ContextCancelled(t *testing.T) {
	repo := &mockPostgresRepo{data: make(map[string]string)}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := repo.Save(ctx, "id", "url")
	assert.Equal(t, context.Canceled, err)

	_, err = repo.Get(ctx, "id")
	assert.Equal(t, context.Canceled, err)

	err = repo.Ping(ctx)
	assert.Equal(t, context.Canceled, err)
}
