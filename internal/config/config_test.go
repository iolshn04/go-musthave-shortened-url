package config

import (
	"os"
	"testing"
)

func TestNewAppConfig(t *testing.T) {
	tests := []struct {
		name              string
		envServer         string
		envBase           string
		envLog            string
		envFile           string
		flagServer        string
		flagBase          string
		flagLog           string
		flagFile          string
		wantServerAddr    string
		wantBaseURL       string
		wantLogLevel      string
		wantFileStorePath string
	}{
		{
			name:              "defaults",
			wantServerAddr:    "localhost:8080",
			wantBaseURL:       "http://localhost:8080",
			wantLogLevel:      "info",
			wantFileStorePath: "stortened_urls.json",
		},
		{
			name:              "flags only",
			flagServer:        "localhost:1111",
			flagBase:          "http://localhost:2222",
			flagLog:           "debug",
			flagFile:          "flags.json",
			wantServerAddr:    "localhost:1111",
			wantBaseURL:       "http://localhost:2222",
			wantLogLevel:      "debug",
			wantFileStorePath: "flags.json",
		},
		{
			name:              "env only",
			envServer:         "localhost:3333",
			envBase:           "http://localhost:4444",
			envLog:            "warn",
			envFile:           "env.json",
			wantServerAddr:    "localhost:3333",
			wantBaseURL:       "http://localhost:4444",
			wantLogLevel:      "warn",
			wantFileStorePath: "env.json",
		},
		{
			name:              "env overrides flags",
			envServer:         "localhost:5555",
			envBase:           "http://localhost:6666",
			envLog:            "error",
			envFile:           "env_override.json",
			flagServer:        "localhost:7777",
			flagBase:          "http://localhost:8888",
			flagLog:           "debug",
			flagFile:          "flags_override.json",
			wantServerAddr:    "localhost:5555",
			wantBaseURL:       "http://localhost:6666",
			wantLogLevel:      "error",
			wantFileStorePath: "env_override.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Clearenv()

			if tt.envServer != "" {
				_ = os.Setenv("SERVER_ADDRESS", tt.envServer)
			}
			if tt.envBase != "" {
				_ = os.Setenv("BASE_URL", tt.envBase)
			}
			if tt.envLog != "" {
				_ = os.Setenv("LOG_LEVEL", tt.envLog)
			}
			if tt.envFile != "" {
				_ = os.Setenv("FILE_STORAGE_PATH", tt.envFile)
			}

			os.Args = []string{"cmd"}
			if tt.flagServer != "" {
				os.Args = append(os.Args, "-a="+tt.flagServer)
			}
			if tt.flagBase != "" {
				os.Args = append(os.Args, "-b="+tt.flagBase)
			}
			if tt.flagLog != "" {
				os.Args = append(os.Args, "-l="+tt.flagLog)
			}
			if tt.flagFile != "" {
				os.Args = append(os.Args, "-f="+tt.flagFile)
			}

			cfg := NewAppConfig()

			if cfg.ServerAddress != tt.wantServerAddr {
				t.Errorf("ServerAddress = %s, want %s", cfg.ServerAddress, tt.wantServerAddr)
			}
			if cfg.BaseURL != tt.wantBaseURL {
				t.Errorf("BaseURL = %s, want %s", cfg.BaseURL, tt.wantBaseURL)
			}
			if cfg.LogLevel != tt.wantLogLevel {
				t.Errorf("LogLevel = %s, want %s", cfg.LogLevel, tt.wantLogLevel)
			}
			if cfg.FileStoragePath != tt.wantFileStorePath {
				t.Errorf("FileStoragePath = %s, want %s", cfg.FileStoragePath, tt.wantFileStorePath)
			}
		})
	}
}
