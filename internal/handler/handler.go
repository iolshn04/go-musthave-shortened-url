package handler

import (
	"encoding/json"
	"github.com/iolshn04/go-musthave-shortened-url/internal/logger"
	"github.com/iolshn04/go-musthave-shortened-url/internal/middlewares"
	"go.uber.org/zap"
	"io"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"github.com/iolshn04/go-musthave-shortened-url/internal/model"
	"github.com/iolshn04/go-musthave-shortened-url/internal/repository"
	"github.com/iolshn04/go-musthave-shortened-url/internal/service"
)

func CreateHandler(w http.ResponseWriter, r *http.Request, s *service.ShortenerService, baseURL string, log *zap.Logger) {
	body, err := io.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	id, err := s.Shorten(string(body))
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusInternalServerError)
		return
	}

	fullURL, err := url.JoinPath(baseURL, id)
	if err != nil {
		log.Error("failed to join URL path", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fullURL))
}

func RedirectHandler(w http.ResponseWriter, r *http.Request, s *service.ShortenerService) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	original, err := s.GetOriginal(id)
	if err == repository.ErrNotFound {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Location", original)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func JSONShortenHandler(w http.ResponseWriter, r *http.Request, s *service.ShortenerService, baseURL string, log *zap.Logger) {
	var req model.ShortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.URL == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	id, err := s.Shorten(req.URL)
	if err != nil {
		log.Error("failed to shorten URL", zap.String("url", req.URL), zap.Error(err))

		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	fullURL, err := url.JoinPath(baseURL, id)
	if err != nil {
		log.Error("failed to join URL path", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	resp := model.ShortenResponse{Result: fullURL}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}

func NewRouter(s *service.ShortenerService, baseURL string, log *zap.Logger) *chi.Mux {
	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler { return logger.RequestLogger(log, next) })
	r.Use(middlewares.GzipRequestMiddleware)
	r.Use(middlewares.GzipMiddleware)
	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		CreateHandler(w, r, s, baseURL, log)
	})
	r.Post("/api/shorten", func(w http.ResponseWriter, r *http.Request) {
		JSONShortenHandler(w, r, s, baseURL, log)
	})
	r.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
		RedirectHandler(w, r, s)
	})
	return r
}
