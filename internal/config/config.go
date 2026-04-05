package config

import (
	"flag"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type AppConfig struct {
	ServerAddress   string `mapstructure:"server_address"`
	BaseURL         string `mapstructure:"base_url"`
	LogLevel        string `mapstructure:"log_level"`
	FileStoragePath string `mapstructure:"file_storage_path"`
	SecretKey       string `mapstructure:"secret_key"`
	DSN             string `mapstructure:"database_dsn"`
	AuditFile       string `mapstructure:"audit_file"`
	AuditURL        string `mapstructure:"audit_url"`
	EnableHTTPS     bool   `mapstructure:"enable_https"`
	CertFile        string `mapstructure:"cert_file"`
	KeyFile         string `mapstructure:"key_file"`
	TrustedSubnet   string `mapstructure:"trusted_subnet"`
	GRPCAddress     string `mapstructure:"grpc_address"`
}

// NewAppConfig загружает конфиг с приоритетом:
// defaults < config file < flags < ENV
func NewAppConfig() *AppConfig {
	const (
		defaultServerAddress   = "localhost:8080"
		defaultBaseURL         = "http://localhost:8080"
		defaultLogLevel        = "info"
		defaultFileStoragePath = "stortened_urls.json"
		defaultSecretKey       = "super-secret-key"
		defaultCertFile        = "cert.pem"
		defaultKeyFile         = "key.pem"
	)

	fs := flag.NewFlagSet("shortener", flag.ContinueOnError)

	flagServer := fs.String("a", "", "server address")
	flagBase := fs.String("b", "", "base url")
	flagLog := fs.String("l", "", "log level")
	flagFile := fs.String("f", "", "file storage path")
	flagAuditFile := fs.String("audit-file", "", "audit file path")
	flagAuditURL := fs.String("audit-url", "", "audit remote url")
	flagDatabaseDSN := fs.String("d", "", "database DSN")
	flagHTTPS := fs.Bool("s", false, "enable HTTPS")
	flagCert := fs.String("cert", "", "path to cert file")
	flagKey := fs.String("key", "", "path to key file")
	flagTrustedSubnet := fs.String("t", "", "trusted subnet in CIDR")
	flagGRPC := fs.String("g", "", "grpc address")

	var configPath string
	fs.StringVar(&configPath, "c", "", "config file path")
	fs.StringVar(&configPath, "config", "", "config file path")

	_ = fs.Parse(os.Args[1:])

	if envConfig := os.Getenv("CONFIG"); envConfig != "" {
		configPath = envConfig
	}

	v := viper.New()

	// defaults
	v.SetDefault("server_address", defaultServerAddress)
	v.SetDefault("base_url", defaultBaseURL)
	v.SetDefault("log_level", defaultLogLevel)
	v.SetDefault("file_storage_path", defaultFileStoragePath)
	v.SetDefault("secret_key", defaultSecretKey)
	v.SetDefault("enable_https", false)
	v.SetDefault("cert_file", defaultCertFile)
	v.SetDefault("key_file", defaultKeyFile)

	// config file
	if configPath != "" {
		v.SetConfigFile(configPath)
		_ = v.ReadInConfig()
	}

	// ENV (bind before flags)
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// flags override config file
	if *flagServer != "" {
		v.Set("server_address", *flagServer)
	}
	if *flagBase != "" {
		v.Set("base_url", *flagBase)
	}
	if *flagLog != "" {
		v.Set("log_level", *flagLog)
	}
	if *flagFile != "" {
		v.Set("file_storage_path", *flagFile)
	}
	if *flagDatabaseDSN != "" {
		v.Set("database_dsn", *flagDatabaseDSN)
	}
	if *flagAuditFile != "" {
		v.Set("audit_file", *flagAuditFile)
	}
	if *flagAuditURL != "" {
		v.Set("audit_url", *flagAuditURL)
	}
	if *flagHTTPS {
		v.Set("enable_https", true)
	}
	if *flagCert != "" {
		v.Set("cert_file", *flagCert)
	}
	if *flagKey != "" {
		v.Set("key_file", *flagKey)
	}
	if *flagTrustedSubnet != "" {
		v.Set("trusted_subnet", *flagTrustedSubnet)
	}
	if *flagGRPC != "" {
		v.Set("grpc_address", *flagGRPC)
	}

	// повторно применяем ENV, чтобы он имел последний приоритет
	for _, key := range []string{
		"SERVER_ADDRESS", "BASE_URL", "LOG_LEVEL", "FILE_STORAGE_PATH",
		"SECRET_KEY", "DATABASE_DSN", "AUDIT_FILE", "AUDIT_URL",
		"ENABLE_HTTPS", "CERT_FILE", "KEY_FILE", "TRUSTED_SUBNET", "GRPC_ADDRESS",
	} {
		if val, ok := os.LookupEnv(key); ok {
			k := strings.ToLower(strings.ReplaceAll(key, "_", "_"))
			v.Set(k, val)
		}
	}

	cfg := &AppConfig{}
	if err := v.Unmarshal(cfg); err != nil {
		panic(err)
	}

	return cfg
}
