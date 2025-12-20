package config

import (
	"os"
	"testing"
)

func TestNewDBConfig(t *testing.T) {
	tests := []struct {
		name    string
		envDSN  string
		flagDSN string
		wantDSN string
	}{
		{
			name:    "default empty",
			wantDSN: "",
		},
		{
			name:    "flag only",
			flagDSN: "postgres://flag",
			wantDSN: "postgres://flag",
		},
		{
			name:    "env only",
			envDSN:  "postgres://env",
			wantDSN: "postgres://env",
		},
		{
			name:    "env overrides flag",
			envDSN:  "postgres://env-priority",
			flagDSN: "postgres://flag",
			wantDSN: "postgres://env-priority",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Clearenv()

			if tt.envDSN != "" {
				_ = os.Setenv("DATABASE_DSN", tt.envDSN)
			}

			os.Args = []string{"cmd"}
			if tt.flagDSN != "" {
				os.Args = append(os.Args, "-d="+tt.flagDSN)
			}

			cfg := NewDBConfig()

			if cfg.DSN != tt.wantDSN {
				t.Errorf("DSN = %s, want %s", cfg.DSN, tt.wantDSN)
			}
		})
	}
}
