package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/iolshn04/go-musthave-shortened-url/internal/audit"
	"github.com/iolshn04/go-musthave-shortened-url/internal/middlewares"
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

type DummyObserver struct{}

func (d *DummyObserver) Notify(e audit.Event) {}

func TestCreateHandler(t *testing.T) {
	repo := repository.NewMemoryStorage()
	s := service.NewShortenerService(repo)
	baseURL := "http://localhost:8080"
	log := zap.NewNop()
	auditor := audit.NewAuditor()
	auditor.Register(&DummyObserver{})

	t.Run("valid url", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/", strings.NewReader("https://yandex.ru"))
		ctx := context.WithValue(req.Context(), middlewares.UserIDKey, "test-user")
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()
		CreateHandler(w, req, s, baseURL, log, auditor)

		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), "http://localhost:8080/")
	})

	t.Run("empty body", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/", strings.NewReader(""))
		ctx := context.WithValue(req.Context(), middlewares.UserIDKey, "test-user")
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()
		CreateHandler(w, req, s, baseURL, log, auditor)

		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestRedirectHandler(t *testing.T) {
	repo := repository.NewMemoryStorage()
	s := service.NewShortenerService(repo)
	userID := "user123"
	auditor := audit.NewAuditor()
	auditor.Register(&DummyObserver{})

	id, _ := s.Shorten(context.Background(), userID, "https://yandex.ru")

	t.Run("redirect existing", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/"+id, nil)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", id)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		RedirectHandler(w, req, s, auditor)

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
		RedirectHandler(w, req, s, auditor)

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
	auditor := audit.NewAuditor()
	auditor.Register(&DummyObserver{})

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
			ctx := context.WithValue(req.Context(), middlewares.UserIDKey, "test-user")
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()

			JSONShortenHandler(w, req, s, baseURL, log, auditor)

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

func TestPingHandler(t *testing.T) {
	repo := repository.NewMemoryStorage()

	req := httptest.NewRequest("GET", "/ping", nil)
	w := httptest.NewRecorder()

	PingHandler(w, req, repo)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestBatchShortenHandler(t *testing.T) {
	repo := repository.NewMemoryStorage()
	s := service.NewShortenerService(repo)
	baseURL := "http://localhost:8080"
	log := zap.NewNop()

	batchReq := []map[string]string{
		{"correlation_id": "cid1", "original_url": "https://google.com"},
		{"correlation_id": "cid2", "original_url": "https://yandex.ru"},
	}

	body, _ := json.Marshal(batchReq)
	req := httptest.NewRequest("POST", "/api/shorten/batch", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	ctx := context.WithValue(req.Context(), middlewares.UserIDKey, "test-user")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	BatchShortenHandler(w, req, s, baseURL, log)
	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var respBody []struct {
		CorrelationID string `json:"correlation_id"`
		ShortURL      string `json:"short_url"`
	}
	err := json.NewDecoder(resp.Body).Decode(&respBody)
	assert.NoError(t, err)
	assert.Len(t, respBody, 2)
	for _, item := range respBody {
		assert.True(t, item.ShortURL != "")
	}
}

func BenchmarkCreateHandler(b *testing.B) {
	repo := repository.NewMemoryStorage()
	s := service.NewShortenerService(repo)
	auditor := audit.NewAuditor()
	auditor.Register(&DummyObserver{})

	baseURL := "http://localhost:8080"
	log := zap.NewNop()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/", strings.NewReader("https://example.com"))

		ctx := context.WithValue(req.Context(), middlewares.UserIDKey, "bench-user")
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()

		CreateHandler(w, req, s, baseURL, log, auditor)
	}
}
