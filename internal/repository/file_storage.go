package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

type fileStorage struct {
	mem  Repository
	path string
	mu   sync.Mutex
}

type fileEntry struct {
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
			fs.mem.Save(context.Background(), e.ShortURL, e.Original)
		}
	}

	return fs, nil
}

func (f *fileStorage) Save(ctx context.Context, shortURL, original string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	if err := f.mem.Save(ctx, shortURL, original); err != nil {
		return err
	}

	return f.persist()
}

func (f *fileStorage) Get(ctx context.Context, shortURL string) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	return f.mem.Get(ctx, shortURL)
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
	for k, v := range memData.data {
		entries = append(entries, fileEntry{
			ShortURL: k,
			Original: v,
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

func (f *fileStorage) SaveBatch(ctx context.Context, data map[string]string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	for k, v := range data {
		if err := f.mem.Save(ctx, k, v); err != nil {
			return err
		}
	}
	return f.persist()
}
