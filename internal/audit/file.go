package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

type FileObserver struct {
	file *os.File
	mu   sync.Mutex
}

func NewFileObserver(path string) (*FileObserver, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("open audit file %s: %w", path, err)
	}

	return &FileObserver{
		file: f,
	}, nil
}

func (f *FileObserver) Notify(e Event) {
	data, err := json.Marshal(e)
	if err != nil {
		fmt.Printf("marshal audit event: %v\n", err)
		return
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	if _, err := f.file.Write(append(data, '\n')); err != nil {
		fmt.Printf("write audit event: %v\n", err)
	}
}

func (f *FileObserver) Close() error {
	return f.file.Close()
}
