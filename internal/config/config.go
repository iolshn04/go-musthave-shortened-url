package config

import (
	"flag"
	"os"
)

type AppConfig struct {
	ServerAddress   string
	BaseURL         string
	LogLevel        string
	FileStoragePath string
	SecretKey       string
	DSN             string
	AuditFile       string
	AuditURL        string
}

func NewAppConfig() *AppConfig {
	const (
		defaultServerAddress   = "localhost:8080"
		defaultBaseURL         = "http://localhost:8080"
		defaultLogLevel        = "info"
		defaultFileStoragePath = "stortened_urls.json"
		defaultSecretKey       = "super-secret-key"
	)

	fs := flag.NewFlagSet("shortener", flag.ContinueOnError)
	flagServer := fs.String("a", "", "server address")
	flagBase := fs.String("b", "", "base url")
	flagLogLevel := fs.String("l", "", "log level")
	flagFileStoragePath := fs.String("f", "", "file storage path")
	flagAuditFile := fs.String("audit-file", "", "audit file path")
	flagAuditURL := fs.String("audit-url", "", "audit remote url")
	flagDatabaseDSN := fs.String("d", "", "database DSN")

	_ = fs.Parse(os.Args[1:])

	cfg := &AppConfig{}

	if val, ok := os.LookupEnv("SERVER_ADDRESS"); ok {
		cfg.ServerAddress = val
	} else if *flagServer != "" {
		cfg.ServerAddress = *flagServer
	} else {
		cfg.ServerAddress = defaultServerAddress
	}

	if val, ok := os.LookupEnv("BASE_URL"); ok {
		cfg.BaseURL = val
	} else if *flagBase != "" {
		cfg.BaseURL = *flagBase
	} else {
		cfg.BaseURL = defaultBaseURL
	}

	if val, ok := os.LookupEnv("LOG_LEVEL"); ok {
		cfg.LogLevel = val
	} else if *flagLogLevel != "" {
		cfg.LogLevel = *flagLogLevel
	} else {
		cfg.LogLevel = defaultLogLevel
	}

	if val, ok := os.LookupEnv("FILE_STORAGE_PATH"); ok {
		cfg.FileStoragePath = val
	} else if *flagFileStoragePath != "" {
		cfg.FileStoragePath = *flagFileStoragePath
	} else {
		cfg.FileStoragePath = defaultFileStoragePath
	}
	if val, ok := os.LookupEnv("SECRET_KEY"); ok {
		cfg.SecretKey = val
	} else {
		cfg.SecretKey = defaultSecretKey
	}

	if val, ok := os.LookupEnv("DATABASE_DSN"); ok {
		cfg.DSN = val
	} else if *flagDatabaseDSN != "" {
		cfg.DSN = *flagDatabaseDSN
	} else {
		cfg.DSN = ""
	}

	if val, ok := os.LookupEnv("AUDIT_FILE"); ok {
		cfg.AuditFile = val
	} else {
		cfg.AuditFile = *flagAuditFile
	}

	if val, ok := os.LookupEnv("AUDIT_URL"); ok {
		cfg.AuditURL = val
	} else {
		cfg.AuditURL = *flagAuditURL
	}

	return cfg
}
