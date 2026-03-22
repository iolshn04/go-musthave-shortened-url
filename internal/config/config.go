package config

import (
	"encoding/json"
	"flag"
	"os"
)

type AppConfig struct {
	ServerAddress   string `json:"server_address"`
	BaseURL         string `json:"base_url"`
	LogLevel        string `json:"log_level"`
	FileStoragePath string `json:"file_storage_path"`
	SecretKey       string `json:"secret_key"`
	DSN             string `json:"database_dsn"`
	AuditFile       string `json:"audit_file"`
	AuditURL        string `json:"audit_url"`
	EnableHTTPS     bool   `json:"enable_https"`
}

func loadFromFile(path string) (*AppConfig, error) {
	if path == "" {
		return &AppConfig{}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg AppConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
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

	// flags
	flagServer := fs.String("a", "", "server address")
	flagBase := fs.String("b", "", "base url")
	flagLogLevel := fs.String("l", "", "log level")
	flagFileStoragePath := fs.String("f", "", "file storage path")
	flagAuditFile := fs.String("audit-file", "", "audit file path")
	flagAuditURL := fs.String("audit-url", "", "audit remote url")
	flagDatabaseDSN := fs.String("d", "", "database DSN")
	flagHTTPS := fs.Bool("s", false, "enable HTTPS")

	var configPath string
	fs.StringVar(&configPath, "c", "", "config file path")
	fs.StringVar(&configPath, "config", "", "config file path")

	_ = fs.Parse(os.Args[1:])

	// ENV для config
	if envConfig := os.Getenv("CONFIG"); envConfig != "" {
		configPath = envConfig
	}

	// загрузка файла
	fileCfg, _ := loadFromFile(configPath)

	cfg := &AppConfig{}

	// ========================
	// ServerAddress
	cfg.ServerAddress = defaultServerAddress
	if fileCfg.ServerAddress != "" {
		cfg.ServerAddress = fileCfg.ServerAddress
	}
	if *flagServer != "" {
		cfg.ServerAddress = *flagServer
	}
	if val, ok := os.LookupEnv("SERVER_ADDRESS"); ok {
		cfg.ServerAddress = val
	}

	// ========================
	// BaseURL
	cfg.BaseURL = defaultBaseURL
	if fileCfg.BaseURL != "" {
		cfg.BaseURL = fileCfg.BaseURL
	}
	if *flagBase != "" {
		cfg.BaseURL = *flagBase
	}
	if val, ok := os.LookupEnv("BASE_URL"); ok {
		cfg.BaseURL = val
	}

	// ========================
	// LogLevel
	cfg.LogLevel = defaultLogLevel
	if fileCfg.LogLevel != "" {
		cfg.LogLevel = fileCfg.LogLevel
	}
	if *flagLogLevel != "" {
		cfg.LogLevel = *flagLogLevel
	}
	if val, ok := os.LookupEnv("LOG_LEVEL"); ok {
		cfg.LogLevel = val
	}

	// ========================
	// FileStoragePath
	cfg.FileStoragePath = defaultFileStoragePath
	if fileCfg.FileStoragePath != "" {
		cfg.FileStoragePath = fileCfg.FileStoragePath
	}
	if *flagFileStoragePath != "" {
		cfg.FileStoragePath = *flagFileStoragePath
	}
	if val, ok := os.LookupEnv("FILE_STORAGE_PATH"); ok {
		cfg.FileStoragePath = val
	}

	// ========================
	// SecretKey (без файла)
	if val, ok := os.LookupEnv("SECRET_KEY"); ok {
		cfg.SecretKey = val
	} else {
		cfg.SecretKey = defaultSecretKey
	}

	// ========================
	// Database DSN
	if fileCfg.DSN != "" {
		cfg.DSN = fileCfg.DSN
	}
	if *flagDatabaseDSN != "" {
		cfg.DSN = *flagDatabaseDSN
	}
	if val, ok := os.LookupEnv("DATABASE_DSN"); ok {
		cfg.DSN = val
	}

	// ========================
	// AuditFile
	if fileCfg.AuditFile != "" {
		cfg.AuditFile = fileCfg.AuditFile
	}
	if *flagAuditFile != "" {
		cfg.AuditFile = *flagAuditFile
	}
	if val, ok := os.LookupEnv("AUDIT_FILE"); ok {
		cfg.AuditFile = val
	}

	// ========================
	// AuditURL
	if fileCfg.AuditURL != "" {
		cfg.AuditURL = fileCfg.AuditURL
	}
	if *flagAuditURL != "" {
		cfg.AuditURL = *flagAuditURL
	}
	if val, ok := os.LookupEnv("AUDIT_URL"); ok {
		cfg.AuditURL = val
	}

	// ========================
	// EnableHTTPS
	if fileCfg.EnableHTTPS {
		cfg.EnableHTTPS = true
	}
	if *flagHTTPS {
		cfg.EnableHTTPS = true
	}
	if val, ok := os.LookupEnv("ENABLE_HTTPS"); ok && val == "true" {
		cfg.EnableHTTPS = true
	}

	return cfg
}
