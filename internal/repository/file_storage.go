package repository

import (
	"encoding/json"
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
			return nil, err
		}
		defer f.Close()

		var entries []fileEntry
		if err := json.NewDecoder(f).Decode(&entries); err != nil {
			return nil, err
		}

		for _, e := range entries {
			fs.mem.Save(e.ShortURL, e.Original)
		}
	}

	return fs, nil
}

func (f *fileStorage) Save(shortURL, original string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if err := f.mem.Save(shortURL, original); err != nil {
		return err
	}

	return f.persist()
}

func (f *fileStorage) Get(shortURL string) (string, error) {
	return f.mem.Get(shortURL)
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
		return err
	}
	defer fh.Close()

	enc := json.NewEncoder(fh)
	enc.SetIndent("", "  ")
	return enc.Encode(entries)
}
