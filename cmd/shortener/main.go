package main

import (
	"github.com/iolshn04/go-musthave-shortened-url/internal/config"
	"github.com/iolshn04/go-musthave-shortened-url/internal/logger"
	"net/http"

	"github.com/iolshn04/go-musthave-shortened-url/internal/handler"
	"github.com/iolshn04/go-musthave-shortened-url/internal/repository"
	"github.com/iolshn04/go-musthave-shortened-url/internal/service"
)

func main() {
	cfg := config.NewConfig()
	_ = logger.Initialize(cfg.LogLevel)
	repo := repository.NewMemoryStorage()
	shortener := service.NewShortenerService(repo)
	router := handler.NewRouter(shortener, cfg.BaseURL)

	err := http.ListenAndServe(cfg.ServerAddress, router)
	if err != nil {
		return
	}
}
