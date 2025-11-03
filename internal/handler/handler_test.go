package handler

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/iolshn04/go-musthave-shortened-url/internal/repository"
	"github.com/iolshn04/go-musthave-shortened-url/internal/service"
	"github.com/stretchr/testify/assert"
)

func TestCreateHandler(t *testing.T) {
	repo := repository.NewMemoryStorage()
	s := service.NewShortenerService(repo)
	baseURL := "http://localhost:8080"

	t.Run("valid url", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/", strings.NewReader("https://yandex.ru"))
		w := httptest.NewRecorder()
		CreateHandler(w, req, s, baseURL)

		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), "http://localhost:8080/")
	})

	t.Run("empty body", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/", strings.NewReader(""))
		w := httptest.NewRecorder()
		CreateHandler(w, req, s, baseURL)

		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestRedirectHandler(t *testing.T) {
	repo := repository.NewMemoryStorage()
	s := service.NewShortenerService(repo)

	id, _ := s.Shorten("https://yandex.ru")

	t.Run("redirect existing", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/"+id, nil)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", id)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		RedirectHandler(w, req, s)

		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusTemporaryRedirect, resp.StatusCode)
		assert.Equal(t, "https://yandex.ru", resp.Header.Get("Location"))
	})

	t.Run("not found", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/missing", nil)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "missing")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		RedirectHandler(w, req, s)

		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}
