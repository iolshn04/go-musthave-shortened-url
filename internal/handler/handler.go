package handler

import (
	"io"
	"net/http"
	"strings"

	"github.com/iolshn04/go-musthave-shortened-url/internal/repository"
	"github.com/iolshn04/go-musthave-shortened-url/internal/service"
)

func CreateHandler(w http.ResponseWriter, r *http.Request, s *service.ShortenerService) {
	if r.Method != http.MethodPost {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

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
	if r.Method != http.MethodGet {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/")
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

func NewRouter(s *service.ShortenerService) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			CreateHandler(w, r, s)
		} else if r.Method == http.MethodGet {
			RedirectHandler(w, r, s)
		} else {
			http.Error(w, "bad request", http.StatusBadRequest)
		}
	})
	return mux
}
