package main

import (
	"fmt"
	"go.uber.org/zap"
	"net/http"
	"os"

	"github.com/iolshn04/go-musthave-shortened-url/internal/config"
	"github.com/iolshn04/go-musthave-shortened-url/internal/handler"
	"github.com/iolshn04/go-musthave-shortened-url/internal/logger"
	"github.com/iolshn04/go-musthave-shortened-url/internal/repository"
	"github.com/iolshn04/go-musthave-shortened-url/internal/service"
)

func main() {
	cfg := config.NewConfig()

	log, err := logger.Initialize(cfg.LogLevel)
	if err != nil {
		fmt.Printf("failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	repo := repository.NewRepositoryFromConfig(cfg.FileStoragePath, log)
	shortener := service.NewShortenerService(repo)
	router := handler.NewRouter(shortener, cfg.BaseURL, log)

	log.Info("HTTP server listening", zap.String("address", cfg.ServerAddress))
	if err := http.ListenAndServe(cfg.ServerAddress, router); err != nil {
		log.Fatal("server stopped with error", zap.Error(err))
	}
}
