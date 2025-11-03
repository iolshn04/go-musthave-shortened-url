package handler

import (
	"net/http"

	"github.com/iolshn04/go-musthave-shortened-url/internal/repository"
	"github.com/iolshn04/go-musthave-shortened-url/internal/service"
)

func RedirectHandler(w http.ResponseWriter, r *http.Request, s *service.ShortenerService) {
	id := r.PathValue("id")
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
