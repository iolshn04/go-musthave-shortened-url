package middlewares_test

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iolshn04/go-musthave-shortened-url/internal/audit"
	"go.uber.org/zap"

	"github.com/go-chi/chi/v5"
	"github.com/iolshn04/go-musthave-shortened-url/internal/handler"
	"github.com/iolshn04/go-musthave-shortened-url/internal/middlewares"
	"github.com/iolshn04/go-musthave-shortened-url/internal/model"
	"github.com/iolshn04/go-musthave-shortened-url/internal/repository"
	"github.com/iolshn04/go-musthave-shortened-url/internal/service"
	"github.com/stretchr/testify/require"
)

type DummyObserver struct{}

func (d *DummyObserver) Notify(e audit.Event) {}

func TestJSONShortenHandler_Gzip(t *testing.T) {
	repo := repository.NewMemoryStorage()
	s := service.NewShortenerService(repo)
	baseURL := "http://localhost:8080"
	log := zap.NewNop()
	secretKey := "secret-key"
	auditor := audit.NewAuditor(log)
	auditor.Register(&DummyObserver{})

	r := chi.NewRouter()
	r.Use(middlewares.GzipRequestMiddleware)
	r.Use(middlewares.GzipMiddleware)
	r.Use(middlewares.AuthMiddleware(secretKey))
	r.Post("/api/shorten", func(w http.ResponseWriter, r *http.Request) {
		handler.JSONShortenHandler(w, r, s, baseURL, log, auditor)
	})

	srv := httptest.NewServer(r)
	defer srv.Close()

	reqObj := model.ShortenRequest{URL: "https://yandex.ru"}
	reqBody, _ := json.Marshal(reqObj)
	expectedPrefix := baseURL + "/"

	t.Run("handles gzipped request", func(t *testing.T) {
		var buf bytes.Buffer
		zw := gzip.NewWriter(&buf)
		_, err := zw.Write(reqBody)
		require.NoError(t, err)
		require.NoError(t, zw.Close())

		req, err := http.NewRequest("POST", srv.URL+"/api/shorten", &buf)
		require.NoError(t, err)
		req.Header.Set("Content-Encoding", "gzip")
		req.Header.Set("Content-Type", "application/json")

		req = req.WithContext(context.WithValue(req.Context(), middlewares.UserIDKey, "testuser"))

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		bodyBytes, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		if resp.Header.Get("Content-Encoding") == "gzip" {
			zr, err := gzip.NewReader(bytes.NewReader(bodyBytes))
			require.NoError(t, err)
			defer zr.Close()
			bodyBytes, err = io.ReadAll(zr)
			require.NoError(t, err)
		}

		require.Equal(t, http.StatusCreated, resp.StatusCode)
		require.Contains(t, string(bodyBytes), expectedPrefix)
	})

	t.Run("sends gzipped response", func(t *testing.T) {
		req, err := http.NewRequest("POST", srv.URL+"/api/shorten", bytes.NewReader(reqBody))
		require.NoError(t, err)
		req.Header.Set("Accept-Encoding", "gzip")
		req.Header.Set("Content-Type", "application/json")

		req = req.WithContext(context.WithValue(req.Context(), middlewares.UserIDKey, "testuser"))

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, "gzip", resp.Header.Get("Content-Encoding"))

		zr, err := gzip.NewReader(resp.Body)
		require.NoError(t, err)
		defer zr.Close()

		body, err := io.ReadAll(zr)
		require.NoError(t, err)

		var respObj model.ShortenResponse
		require.NoError(t, json.Unmarshal(body, &respObj))
		require.Contains(t, respObj.Result, expectedPrefix)
	})
}
