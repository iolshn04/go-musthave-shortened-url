package handler

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/iolshn04/go-musthave-shortened-url/internal/service"
)

func CreateHandler(w http.ResponseWriter, r *http.Request, s *service.ShortenerService) {
	body, err := io.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	original := strings.TrimSpace(string(body))
	if original == "" {
		http.Error(w, "empty url", http.StatusBadRequest)
		return
	}

	id, err := s.Shorten(original)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	shortURL := fmt.Sprintf("http://%s/%s", r.Host, id)
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write([]byte(shortURL))
}
