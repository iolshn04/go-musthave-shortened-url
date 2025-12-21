package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/iolshn04/go-musthave-shortened-url/internal/model"
	"os"
	"sync"
)

type fileStorage struct {
	mem  Repository
	path string
	mu   sync.Mutex
}

type fileEntry struct {
	UserID   string `json:"user_id"`
	ShortURL string `json:"short_url"`
	Original string `json:"original_url"`
}

func NewFileStorage(path string) (Repository, error) {
	fs := &fileStorage{
		mem:  NewMemoryStorage(),
		path: path,
	}

	if _, err := os.Stat(path); err == nil {
		f, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("failed to open file %s: %w", path, err)
		}
		defer f.Close()

		var entries []fileEntry
		if err := json.NewDecoder(f).Decode(&entries); err != nil {
			return nil, fmt.Errorf("failed to decode file %s: %w", path, err)
		}

		for _, e := range entries {
			fs.mem.Save(context.Background(), e.UserID, e.ShortURL, e.Original)
		}
	}

	return fs, nil
}

func (f *fileStorage) Save(ctx context.Context, userID, shortURL, original string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	if err := f.mem.Save(ctx, userID, shortURL, original); err != nil {
		return err
	}

	return f.persist()
}

func (f *fileStorage) Get(ctx context.Context, shortID string) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	return f.mem.Get(ctx, shortID)
}

func (f *fileStorage) GetByUser(ctx context.Context, userID string) ([]model.UserURL, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	return f.mem.GetByUser(ctx, userID)
}

func (f *fileStorage) Ping(ctx context.Context) error {
	return nil
}

func (f *fileStorage) persist() error {
	memData, ok := f.mem.(*memoryStorage)
	if !ok {
		return nil
	}

	entries := make([]fileEntry, 0, len(memData.data))
	for short, rec := range memData.data {
		entries = append(entries, fileEntry{
			UserID:   rec.UserID,
			ShortURL: short,
			Original: rec.OriginalURL,
		})
	}

	fh, err := os.Create(f.path)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", f.path, err)
	}
	defer fh.Close()

	enc := json.NewEncoder(fh)
	enc.SetIndent("", "  ")
	return enc.Encode(entries)
}

func (f *fileStorage) SaveBatch(ctx context.Context, userID string, data map[string]string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	if err := f.mem.SaveBatch(ctx, userID, data); err != nil {
		return err
	}
	return f.persist()
}
