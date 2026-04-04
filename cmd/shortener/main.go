package main

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"sync"
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

	var wg sync.WaitGroup

	pprofSrv := &http.Server{
		Addr:    "localhost:6060",
		Handler: nil,
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := pprofSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("pprof server error", zap.Error(err))
		}
	}()

	shortener := service.NewShortenerService(repo)
	router := handler.NewRouter(shortener, appCfg.BaseURL, log, repo, appCfg.SecretKey, auditor, appCfg.TrustedSubnet)

	srv := &http.Server{
		Addr:    appCfg.ServerAddress,
		Handler: router,
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		if appCfg.EnableHTTPS {
			log.Info("HTTPS server listening", zap.String("address", appCfg.ServerAddress))
			if err := srv.ListenAndServeTLS(appCfg.CertFile, appCfg.KeyFile); err != nil && err != http.ErrServerClosed {
				log.Fatal("server error", zap.Error(err))
			}
		} else {
			log.Info("HTTP server listening", zap.String("address", appCfg.ServerAddress))
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatal("server error", zap.Error(err))
			}
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	defer stop()

	<-ctx.Done()
	log.Info("shutting down servers...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("main server shutdown failed", zap.Error(err))
	}

	if err := pprofSrv.Shutdown(shutdownCtx); err != nil {
		log.Error("pprof server shutdown failed", zap.Error(err))
	}

	wg.Wait()
	log.Info("all servers exited properly")
}
