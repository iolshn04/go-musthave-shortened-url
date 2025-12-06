package handler

import (
	"context"
	"encoding/json"
	"go.uber.org/zap"
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
	log := zap.NewNop()

	t.Run("valid url", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/", strings.NewReader("https://yandex.ru"))
		w := httptest.NewRecorder()
		CreateHandler(w, req, s, baseURL, log)

		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), "http://localhost:8080/")
	})

	t.Run("empty body", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/", strings.NewReader(""))
		w := httptest.NewRecorder()
		CreateHandler(w, req, s, baseURL, log)

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

func TestJSONShortenHandler(t *testing.T) {
	repo := repository.NewMemoryStorage()
	s := service.NewShortenerService(repo)
	baseURL := "http://localhost:8080"
	log := zap.NewNop()

	tests := []struct {
		name       string
		body       string
		wantStatus int
		wantPrefix string
	}{
		{
			name:       "valid JSON",
			body:       `{"url":"https://yandex.ru"}`,
			wantStatus: http.StatusCreated,
			wantPrefix: baseURL + "/",
		},
		{
			name:       "empty JSON",
			body:       `{}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid JSON",
			body:       `{"url":`,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/api/shorten", strings.NewReader(tt.body))
			w := httptest.NewRecorder()

			JSONShortenHandler(w, req, s, baseURL, log)

			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tt.wantStatus, resp.StatusCode)

			if tt.wantStatus == http.StatusCreated {
				var respBody struct {
					Result string `json:"result"`
				}
				err := json.NewDecoder(resp.Body).Decode(&respBody)
				assert.NoError(t, err)
				assert.True(t, strings.HasPrefix(respBody.Result, tt.wantPrefix))
			}
		})
	}
}
