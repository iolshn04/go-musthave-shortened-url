package config

import (
	"flag"
	"os"
)

type DBConfig struct {
	DSN string
}

func NewDBConfig() *DBConfig {
	fs := flag.NewFlagSet("shortener", flag.ContinueOnError)
	flagDatabaseDSN := fs.String("d", "", "database DSN")

	_ = fs.Parse(os.Args[1:])

	cfg := &DBConfig{}

	if val, ok := os.LookupEnv("DATABASE_DSN"); ok {
		cfg.DSN = val
	} else if *flagDatabaseDSN != "" {
		cfg.DSN = *flagDatabaseDSN
	} else {
		cfg.DSN = ""
	}

	return cfg
}
