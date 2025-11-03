package main

import (
	"net/http"

	"github.com/iolshn04/go-musthave-shortened-url/internal/handler"
	"github.com/iolshn04/go-musthave-shortened-url/internal/repository"
	"github.com/iolshn04/go-musthave-shortened-url/internal/service"
)

func main() {
	repo := repository.NewMemoryStorage()
	shortener := service.NewShortenerService(repo)
	router := handler.NewRouter(shortener)

	http.ListenAndServe(":8080", router)
}
