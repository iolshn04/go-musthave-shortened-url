package config

import (
	"os"
	"testing"
)

func TestNewConfig(t *testing.T) {
	tests := []struct {
		name           string
		envServer      string
		envBase        string
		flagServer     string
		flagBase       string
		wantServerAddr string
		wantBaseURL    string
	}{
		{
			name:           "defaults",
			envServer:      "",
			envBase:        "",
			flagServer:     "",
			flagBase:       "",
			wantServerAddr: "localhost:8080",
			wantBaseURL:    "http://localhost:8080",
		},
		{
			name:           "flags only",
			flagServer:     "localhost:1111",
			flagBase:       "http://localhost:2222",
			wantServerAddr: "localhost:1111",
			wantBaseURL:    "http://localhost:2222",
		},
		{
			name:           "env only",
			envServer:      "localhost:3333",
			envBase:        "http://localhost:4444",
			wantServerAddr: "localhost:3333",
			wantBaseURL:    "http://localhost:4444",
		},
		{
			name:           "env overrides flags",
			envServer:      "localhost:5555",
			envBase:        "http://localhost:6666",
			flagServer:     "localhost:7777",
			flagBase:       "http://localhost:8888",
			wantServerAddr: "localhost:5555",
			wantBaseURL:    "http://localhost:6666",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			os.Clearenv()

			if tt.envServer != "" {
				os.Setenv("SERVER_ADDRESS", tt.envServer)
			}
			if tt.envBase != "" {
				os.Setenv("BASE_URL", tt.envBase)
			}

			os.Args = []string{"cmd"}
			if tt.flagServer != "" {
				os.Args = append(os.Args, "-a="+tt.flagServer)
			}
			if tt.flagBase != "" {
				os.Args = append(os.Args, "-b="+tt.flagBase)
			}

			cfg := NewConfig()

			if cfg.ServerAddress != tt.wantServerAddr {
				t.Errorf("ServerAddress = %s, want %s",
					cfg.ServerAddress, tt.wantServerAddr)
			}

			if cfg.BaseURL != tt.wantBaseURL {
				t.Errorf("BaseURL = %s, want %s",
					cfg.BaseURL, tt.wantBaseURL)
			}
		})
	}
}
