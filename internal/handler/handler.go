package handler

import (
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/iolshn04/go-musthave-shortened-url/internal/repository"
	"github.com/iolshn04/go-musthave-shortened-url/internal/service"
)

func CreateHandler(w http.ResponseWriter, r *http.Request, s *service.ShortenerService) {
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
	w.Write([]byte("http://localhost:8080/" + id))
}

func RedirectHandler(w http.ResponseWriter, r *http.Request, s *service.ShortenerService) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	original, err := s.GetOriginal(id)
	if err == repository.ErrNotFound {
		http.Error(w, "not found", http.StatusBadRequest)
		return
	} else if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Location", original)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func NewRouter(s *service.ShortenerService) *chi.Mux {
	r := chi.NewRouter()
	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		CreateHandler(w, r, s)
	})
	r.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
		RedirectHandler(w, r, s)
	})
	return r
}
