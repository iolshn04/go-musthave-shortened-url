package audit

import (
	"encoding/json"
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
		return nil, err
	}

	return &FileObserver{file: f}, nil
}

func (f *FileObserver) Notify(e Event) {
	data, err := json.Marshal(e)
	if err != nil {
		return
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	_, _ = f.file.Write(append(data, '\n'))
}
