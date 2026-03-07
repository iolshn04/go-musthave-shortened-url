package handler

import (
	"context"
	"fmt"
	"net/http/httptest"
	"strings"

	"github.com/iolshn04/go-musthave-shortened-url/internal/audit"

	"github.com/iolshn04/go-musthave-shortened-url/internal/middlewares"
	"github.com/iolshn04/go-musthave-shortened-url/internal/repository"
	"github.com/iolshn04/go-musthave-shortened-url/internal/service"
	"go.uber.org/zap"
)

// ExampleCreateHandler демонстрирует создание короткой ссылки
// через HTTP-обработчик CreateHandler.
func ExampleCreateHandler() {
	repo := repository.NewMemoryStorage()

	// Инициализируем логгер
	log := zap.NewNop()

	// Инициализируем сервис
	svc := service.NewShortenerService(repo)

	// Формируем HTTP-запрос
	req := httptest.NewRequest(
		"POST",
		"/",
		strings.NewReader("https://example.com"),
	)

	auditor := audit.NewAuditor(log)
	auditor.Register(&DummyObserver{})

	// Добавляем пользователя в context
	ctx := context.WithValue(req.Context(), middlewares.UserIDKey, "example-user")
	req = req.WithContext(ctx)

	// Рекордер для ответа
	rec := httptest.NewRecorder()

	// Вызываем handler
	CreateHandler(
		rec,
		req,
		svc,
		"http://localhost:8080",
		zap.NewNop(),
		auditor,
	)

	// Проверяем HTTP-статус
	fmt.Println(rec.Code)

	// Проверяем, что ссылка была создана
	fmt.Println(rec.Body.Len() > 0)

	// Output:
	// 201
	// true
}
