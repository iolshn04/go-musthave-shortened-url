package repository_test

import (
	"os"
	"testing"

	"github.com/iolshn04/go-musthave-shortened-url/internal/repository"
)

func TestFileStorage_SaveAndGet(t *testing.T) {
	file := "test_data.json"
	_ = os.Remove(file)
	defer os.Remove(file)

	fs, err := repository.NewFileStorage(file)
	if err != nil {
		t.Fatalf("ошибка создания хранилища: %v", err)
	}

	err = fs.Save("abc123", "https://example.com")
	if err != nil {
		t.Fatalf("ошибка сохранения: %v", err)
	}

	url, err := fs.Get("abc123")
	if err != nil {
		t.Fatalf("ошибка получения: %v", err)
	}

	if url != "https://example.com" {
		t.Errorf("ожидалось https://example.com, получено %s", url)
	}
}

func TestFileStorage_GetNotFound(t *testing.T) {
	file := "test_data.json"
	_ = os.Remove(file)
	defer os.Remove(file)

	fs, _ := repository.NewFileStorage(file)

	_, err := fs.Get("unknown")
	if err == nil {
		t.Errorf("ожидалась ошибка ErrNotFound")
	}
}

func TestFileStorage_PersistLoad(t *testing.T) {
	file := "test_data.json"
	_ = os.Remove(file)
	defer os.Remove(file)

	fs, err := repository.NewFileStorage(file)
	if err != nil {
		t.Fatalf("ошибка создания: %v", err)
	}

	fs.Save("a1", "https://google.com")
	fs.Save("b2", "https://yandex.ru")

	fs2, err := repository.NewFileStorage(file)
	if err != nil {
		t.Fatalf("ошибка загрузки: %v", err)
	}

	url1, err1 := fs2.Get("a1")
	if err1 != nil || url1 != "https://google.com" {
		t.Errorf("неверное значение после загрузки: %v", url1)
	}

	url2, err2 := fs2.Get("b2")
	if err2 != nil || url2 != "https://yandex.ru" {
		t.Errorf("неверное значение после загрузки: %v", url2)
	}
}
