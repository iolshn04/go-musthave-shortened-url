package handler

import (
	"encoding/json"
	"github.com/iolshn04/go-musthave-shortened-url/internal/logger"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/iolshn04/go-musthave-shortened-url/internal/model"
	"github.com/iolshn04/go-musthave-shortened-url/internal/repository"
	"github.com/iolshn04/go-musthave-shortened-url/internal/service"
)

func CreateHandler(w http.ResponseWriter, r *http.Request, s *service.ShortenerService, baseURL string) {
	body, err := io.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	id, err := s.Shorten(string(body))
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(baseURL + "/" + id))
}

func RedirectHandler(w http.ResponseWriter, r *http.Request, s *service.ShortenerService) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	original, err := s.GetOriginal(id)
	if err == repository.ErrNotFound {
		http.Error(w, "not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Location", original)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func JSONShortenHandler(w http.ResponseWriter, r *http.Request, s *service.ShortenerService, baseURL string) {
	var req model.ShortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.URL == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	id, err := s.Shorten(req.URL)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	resp := model.ShortenResponse{Result: baseURL + "/" + id}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}

func NewRouter(s *service.ShortenerService, baseURL string) *chi.Mux {
	r := chi.NewRouter()
	r.Use(logger.RequestLogger)
	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		CreateHandler(w, r, s, baseURL)
	})
	r.Post("/api/shorten", func(w http.ResponseWriter, r *http.Request) {
		JSONShortenHandler(w, r, s, baseURL)
	})
	r.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
		RedirectHandler(w, r, s)
	})
	return r
}
