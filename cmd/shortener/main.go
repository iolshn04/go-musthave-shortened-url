package main

import (
	"fmt"
	"go.uber.org/zap"
	"net/http"
	_ "net/http/pprof"
	"os"

	"github.com/iolshn04/go-musthave-shortened-url/internal/audit"
	"github.com/iolshn04/go-musthave-shortened-url/internal/config"
	"github.com/iolshn04/go-musthave-shortened-url/internal/handler"
	"github.com/iolshn04/go-musthave-shortened-url/internal/logger"
	"github.com/iolshn04/go-musthave-shortened-url/internal/repository"
	"github.com/iolshn04/go-musthave-shortened-url/internal/service"
)

func main() {
	appCfg := config.NewAppConfig()

	log, err := logger.Initialize(appCfg.LogLevel)
	if err != nil {
		fmt.Printf("failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	repo, err := repository.NewRepositoryFromConfig(appCfg.DSN, appCfg.FileStoragePath, log)
	auditor := audit.NewAuditor()

	if appCfg.AuditFile != "" {
		fileObs, err := audit.NewFileObserver(appCfg.AuditFile)
		if err != nil {
			log.Fatal("failed to init audit file", zap.Error(err))
		}
		auditor.Register(fileObs)
	}

	if appCfg.AuditURL != "" {
		httpObs := audit.NewHTTPObserver(appCfg.AuditURL)
		auditor.Register(httpObs)
	}
	go func() {
		if err := http.ListenAndServe("localhost:6060", nil); err != nil {
			log.Error("pprof server error", zap.Error(err))
		}
	}()

	if err != nil {
		log.Fatal("failed to initialize repository", zap.Error(err))
	}
	shortener := service.NewShortenerService(repo)
	router := handler.NewRouter(shortener, appCfg.BaseURL, log, repo, appCfg.SecretKey, auditor)

	log.Info("HTTP server listening", zap.String("address", appCfg.ServerAddress))
	if err := http.ListenAndServe(appCfg.ServerAddress, router); err != nil {
		log.Fatal("server stopped with error", zap.Error(err))
	}
}
