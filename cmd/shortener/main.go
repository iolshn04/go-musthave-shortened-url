package main

import (
	"fmt"
	"go.uber.org/zap"
	"net/http"
	"os"

	"github.com/iolshn04/go-musthave-shortened-url/internal/config"
	db "github.com/iolshn04/go-musthave-shortened-url/internal/config/db"
	"github.com/iolshn04/go-musthave-shortened-url/internal/handler"
	"github.com/iolshn04/go-musthave-shortened-url/internal/logger"
	"github.com/iolshn04/go-musthave-shortened-url/internal/repository"
	"github.com/iolshn04/go-musthave-shortened-url/internal/service"
)

func main() {
	appCfg := config.NewAppConfig()
	dbCfg := db.NewDBConfig()

	log, err := logger.Initialize(appCfg.LogLevel)
	if err != nil {
		fmt.Printf("failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	repo, err := repository.NewRepositoryFromConfig(dbCfg.DSN, appCfg.FileStoragePath, log)
	if err != nil {
		log.Fatal("failed to initialize repository", zap.Error(err))
	}
	shortener := service.NewShortenerService(repo)
	router := handler.NewRouter(shortener, appCfg.BaseURL, log, repo)

	log.Info("HTTP server listening", zap.String("address", appCfg.ServerAddress))
	if err := http.ListenAndServe(appCfg.ServerAddress, router); err != nil {
		log.Fatal("server stopped with error", zap.Error(err))
	}
}
