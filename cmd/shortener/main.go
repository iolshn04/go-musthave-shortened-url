package main

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/iolshn04/go-musthave-shortened-url/internal/audit"
	"github.com/iolshn04/go-musthave-shortened-url/internal/config"
	"github.com/iolshn04/go-musthave-shortened-url/internal/handler"
	"github.com/iolshn04/go-musthave-shortened-url/internal/logger"
	"github.com/iolshn04/go-musthave-shortened-url/internal/repository"
	"github.com/iolshn04/go-musthave-shortened-url/internal/service"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func printBuildInfo() {
	fmt.Println("Build version:", buildVersion)
	fmt.Println("Build date:", buildDate)
	fmt.Println("Build commit:", buildCommit)
}

func main() {
	printBuildInfo()
	appCfg := config.NewAppConfig()

	log, err := logger.Initialize(appCfg.LogLevel)
	if err != nil {
		fmt.Printf("failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	repo, err := repository.NewRepositoryFromConfig(appCfg.DSN, appCfg.FileStoragePath, log)
	if err != nil {
		log.Fatal("failed to initialize repository", zap.Error(err))
	}

	auditor := audit.NewAuditor(log)

	if appCfg.AuditFile != "" {
		fileObs, err := audit.NewFileObserver(appCfg.AuditFile)
		if err != nil {
			log.Fatal("failed to init audit file", zap.Error(err))
		}
		defer fileObs.Close()
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

	shortener := service.NewShortenerService(repo)
	router := handler.NewRouter(shortener, appCfg.BaseURL, log, repo, appCfg.SecretKey, auditor)

	srv := &http.Server{
		Addr:    appCfg.ServerAddress,
		Handler: router,
	}

	certFile := "cert.pem"
	keyFile := "key.pem"

	go func() {
		if appCfg.EnableHTTPS {
			log.Info("HTTPS server listening", zap.String("address", appCfg.ServerAddress))

			if err := srv.ListenAndServeTLS(certFile, keyFile); err != nil && err != http.ErrServerClosed {
				log.Fatal("server error", zap.Error(err))
			}
		} else {
			log.Info("HTTP server listening", zap.String("address", appCfg.ServerAddress))

			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatal("server error", zap.Error(err))
			}
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)

	<-quit
	log.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("server shutdown failed", zap.Error(err))
	}

	log.Info("server exited properly")
}
