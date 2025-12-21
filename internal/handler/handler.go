package handler

import (
	"encoding/json"
	"errors"
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
	userID, ok := middlewares.UserIDFromContext(r.Context())
	if !ok || userID == "" {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	id, err := s.Shorten(r.Context(), userID, string(body))
	if err != nil {
		var eErr repository.ErrAlreadyExistsWithID
		if errors.As(err, &eErr) {
			fullURL, _ := url.JoinPath(baseURL, eErr.ExistingID)
			w.WriteHeader(http.StatusConflict)
			_, _ = w.Write([]byte(fullURL))
			return
		}

		log.Error("failed to shorten URL", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	fullURL, err := url.JoinPath(baseURL, id)
	if err != nil {
		log.Error("failed to join URL path", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write([]byte(fullURL))
}

func RedirectHandler(w http.ResponseWriter, r *http.Request, s *service.ShortenerService) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	original, err := s.GetOriginal(r.Context(), id)
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
	userID, ok := middlewares.UserIDFromContext(r.Context())
	if !ok || userID == "" {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	var req model.ShortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.URL == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	id, err := s.Shorten(r.Context(), userID, req.URL)
	if err != nil {
		var eErr repository.ErrAlreadyExistsWithID
		if errors.As(err, &eErr) {
			fullURL, _ := url.JoinPath(baseURL, eErr.ExistingID)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			_ = json.NewEncoder(w).Encode(model.ShortenResponse{Result: fullURL})
			return
		}

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

func PingHandler(w http.ResponseWriter, r *http.Request, repo repository.Repository) {
	if err := repo.Ping(r.Context()); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func BatchShortenHandler(
	w http.ResponseWriter,
	r *http.Request,
	s *service.ShortenerService,
	baseURL string,
	log *zap.Logger,
) {
	userID, ok := middlewares.UserIDFromContext(r.Context())
	if !ok || userID == "" {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	var req []model.BatchRequestItem

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if len(req) == 0 {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	input := make(map[string]string, len(req))
	for _, item := range req {
		input[item.CorrelationID] = item.OriginalURL
	}

	result, err := s.ShortenBatch(r.Context(), userID, input)
	if err != nil {
		log.Error("batch failed", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	resp := make([]model.BatchResponseItem, 0, len(req))
	for _, item := range req {
		id := result[item.CorrelationID]
		full, _ := url.JoinPath(baseURL, id)
		resp = append(resp, model.BatchResponseItem{
			CorrelationID: item.CorrelationID,
			ShortURL:      full,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}

func UserURLsHandler(
	w http.ResponseWriter,
	r *http.Request,
	repo repository.Repository,
	baseURL string,
) {
	userID, ok := middlewares.UserIDFromContext(r.Context())
	if !ok || userID == "" {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	urls, err := repo.GetByUser(r.Context(), userID)
	if err != nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	for i := range urls {
		full, err := url.JoinPath(baseURL, urls[i].ShortURL)
		if err == nil {
			urls[i].ShortURL = full
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(urls)
}

func NewRouter(s *service.ShortenerService, baseURL string, log *zap.Logger, repo repository.Repository) *chi.Mux {
	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler { return logger.RequestLogger(log, next) })
	r.Use(middlewares.GzipRequestMiddleware)
	r.Use(middlewares.GzipMiddleware)
	r.Use(middlewares.AuthMiddleware)
	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		CreateHandler(w, r, s, baseURL, log)
	})
	r.Post("/api/shorten", func(w http.ResponseWriter, r *http.Request) {
		JSONShortenHandler(w, r, s, baseURL, log)
	})
	r.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
		RedirectHandler(w, r, s)
	})
	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		PingHandler(w, r, repo)
	})
	r.Post("/api/shorten/batch", func(w http.ResponseWriter, r *http.Request) {
		BatchShortenHandler(w, r, s, baseURL, log)
	})
	r.Get("/api/user/urls", func(w http.ResponseWriter, r *http.Request) {
		UserURLsHandler(w, r, repo, baseURL)
	})
	return r
}
