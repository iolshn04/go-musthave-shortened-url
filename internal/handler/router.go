package handler

import (
	"net/http"

	"github.com/iolshn04/go-musthave-shortened-url/internal/service"
)

func NewRouter(s *service.ShortenerService) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /", func(w http.ResponseWriter, r *http.Request) {
		CreateHandler(w, r, s)
	})
	mux.HandleFunc("GET /{id}", func(w http.ResponseWriter, r *http.Request) {
		RedirectHandler(w, r, s)
	})
	return mux
}
