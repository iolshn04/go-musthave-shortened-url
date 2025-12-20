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

func (m *mockPostgresRepo) SaveBatch(ctx context.Context, data map[string]string) error {
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

func TestPostgresRepository_SaveBatch(t *testing.T) {
	repo := &mockPostgresRepo{data: make(map[string]string)}
	ctx := context.Background()

	data := map[string]string{
		"id1": "https://google.com",
		"id2": "https://yandex.ru",
	}

	err := repo.SaveBatch(ctx, data)
	assert.NoError(t, err)

	for id, expected := range data {
		got, err := repo.Get(ctx, id)
		assert.NoError(t, err)
		assert.Equal(t, expected, got)
	}
}

func TestPostgresRepository_SaveBatch_Error(t *testing.T) {
	repo := &mockPostgresRepo{
		data:    make(map[string]string),
		saveErr: errors.New("batch fail"),
	}
	ctx := context.Background()

	err := repo.SaveBatch(ctx, map[string]string{
		"id1": "https://yandex.ru",
	})
	assert.EqualError(t, err, "batch fail")
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

	err = repo.SaveBatch(ctx, map[string]string{
		"id1": "https://example.com",
	})
	assert.Equal(t, context.Canceled, err)
}

func TestPostgresRepository_SaveAlreadyExists(t *testing.T) {
	repo := &mockPostgresRepo{
		data: map[string]string{"existingID": "https://example.com"},
	}
	ctx := context.Background()

	repo.saveErr = ErrAlreadyExistsWithID{ExistingID: "existingID"}

	err := repo.Save(ctx, "newID", "https://example.com")

	var alreadyExistsErr ErrAlreadyExistsWithID
	ok := errors.As(err, &alreadyExistsErr)
	assert.True(t, ok, "expected ErrAlreadyExistsWithID")
	assert.Equal(t, "existingID", alreadyExistsErr.ExistingID)
}
