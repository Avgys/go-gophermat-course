package config

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v11"
	"github.com/rs/zerolog"
)

type Config struct {
	AppAddr            string `env:"RUN_ADDRESS"`
	AccrualSystemAddr  string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	DBConnectionString string `env:"DATABASE_URI"`
}

func GetConfig(args []string, traceLogger *zerolog.Logger) (*Config, error) {
	cfg := &Config{}

	err := parseFlags(cfg, args)

	if err != nil {
		return nil, fmt.Errorf("error parsing flags for config, %w", err)
	}

	err = parseEnv(cfg)

	if err != nil {
		return nil, fmt.Errorf("error parsing env variables for config, %w", err)
	}

	traceLogger.Info().
		Str("ServerAddr", cfg.AppAddr).
		Str("DBConnectionString", cfg.DBConnectionString).
		Str("AccrualSystemAddr", cfg.AccrualSystemAddr).
		Send()

	return cfg, nil
}

func parseEnv(cfg *Config) error {
	err := env.ParseWithOptions(cfg, env.Options{})

	return err
}

func parseFlags(cfg *Config, args []string) error {
	fs := flag.NewFlagSet("shortener", flag.ContinueOnError)

	fs.StringVar(&cfg.AppAddr, "a", "", "address of HTTP server")
	fs.StringVar(&cfg.DBConnectionString, "d", "", "db connection string url")
	fs.StringVar(&cfg.AccrualSystemAddr, "r", "", "accrual system address")

	return fs.Parse(args)
}
