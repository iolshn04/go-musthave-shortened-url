package config

import (
	"flag"
	"os"
)

type Config struct {
	ServerAddress string
	BaseURL       string
	LogLevel      string
}

func NewConfig() *Config {
	defaultServer := "localhost:8080"
	defaultBaseURL := "http://localhost:8080"

	fs := flag.NewFlagSet("shortener", flag.ContinueOnError)

	flagServer := fs.String("a", "", "server address")
	flagBase := fs.String("b", "", "base url")
	flagLogLevel := fs.String("l", "", "log level")

	_ = fs.Parse(os.Args[1:])

	cfg := &Config{}

	if env := os.Getenv("SERVER_ADDRESS"); env != "" {
		cfg.ServerAddress = env
	} else if *flagServer != "" {
		cfg.ServerAddress = *flagServer
	} else {
		cfg.ServerAddress = defaultServer
	}

	if env := os.Getenv("BASE_URL"); env != "" {
		cfg.BaseURL = env
	} else if *flagBase != "" {
		cfg.BaseURL = *flagBase
	} else {
		cfg.BaseURL = defaultBaseURL
	}

	if env := os.Getenv("LOG_LEVEL"); env != "" {
		cfg.LogLevel = env
	} else if *flagLogLevel != "" {
		cfg.LogLevel = *flagLogLevel
	} else {
		cfg.LogLevel = "info"
	}

	return cfg
}
