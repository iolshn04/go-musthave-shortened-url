package repository_test

import (
	"context"
	"os"
	"testing"

	"github.com/iolshn04/go-musthave-shortened-url/internal/repository"
)

func TestFileStorage_SaveAndGet(t *testing.T) {
	file := "test_data.json"
	userID := "user1"
	_ = os.Remove(file)
	defer os.Remove(file)

	fs, err := repository.NewFileStorage(file)
	if err != nil {
		t.Fatalf("ошибка создания хранилища: %v", err)
	}

	err = fs.Save(context.Background(), userID, "abc123", "https://example.com")
	if err != nil {
		t.Fatalf("ошибка сохранения: %v", err)
	}

	url, err := fs.Get(context.Background(), "abc123")
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

	_, err := fs.Get(context.Background(), "unknown")
	if err == nil {
		t.Errorf("ожидалась ошибка ErrNotFound")
	}
}

func TestFileStorage_PersistLoad(t *testing.T) {
	file := "test_data.json"
	userID := "user1"
	_ = os.Remove(file)
	defer os.Remove(file)

	fs, err := repository.NewFileStorage(file)
	if err != nil {
		t.Fatalf("ошибка создания: %v", err)
	}

	fs.Save(context.Background(), userID, "a1", "https://google.com")
	fs.Save(context.Background(), userID, "b2", "https://yandex.ru")

	fs2, err := repository.NewFileStorage(file)
	if err != nil {
		t.Fatalf("ошибка загрузки: %v", err)
	}

	url1, err1 := fs2.Get(context.Background(), "a1")
	if err1 != nil || url1 != "https://google.com" {
		t.Errorf("неверное значение после загрузки: %v", url1)
	}

	url2, err2 := fs2.Get(context.Background(), "b2")
	if err2 != nil || url2 != "https://yandex.ru" {
		t.Errorf("неверное значение после загрузки: %v", url2)
	}
}

func TestFileStorage_SaveBatchAndGet(t *testing.T) {
	file := "test_data.json"
	userID := "user1"
	_ = os.Remove(file)
	defer os.Remove(file)

	fs, err := repository.NewFileStorage(file)
	if err != nil {
		t.Fatalf("ошибка создания хранилища: %v", err)
	}

	data := map[string]string{
		"id1": "https://google.com",
		"id2": "https://yandex.ru",
	}

	err = fs.SaveBatch(context.Background(), userID, data)
	if err != nil {
		t.Fatalf("ошибка пакетного сохранения: %v", err)
	}

	url1, err := fs.Get(context.Background(), "id1")
	if err != nil || url1 != "https://google.com" {
		t.Errorf("ожидалось https://google.com, получено %s, err=%v", url1, err)
	}

	url2, err := fs.Get(context.Background(), "id2")
	if err != nil || url2 != "https://yandex.ru" {
		t.Errorf("ожидалось https://yandex.ru, получено %s, err=%v", url2, err)
	}
}

func TestFileStorage_ContextCancelled(t *testing.T) {
	file := "test_data.json"
	userID := "user1"
	_ = os.Remove(file)
	defer os.Remove(file)

	fs, _ := repository.NewFileStorage(file)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := fs.Save(ctx, userID, "id", "url")
	if err != context.Canceled {
		t.Errorf("ожидался context.Canceled, получено %v", err)
	}

	err = fs.SaveBatch(ctx, userID, map[string]string{
		"id1": "url1",
	})
	if err != context.Canceled {
		t.Errorf("ожидался context.Canceled, получено %v", err)
	}

	_, err = fs.Get(ctx, "id")
	if err != context.Canceled {
		t.Errorf("ожидался context.Canceled, получено %v", err)
	}
}
