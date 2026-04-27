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

	err := parseEnv(cfg)

	if err != nil {
		return nil, fmt.Errorf("error parsing env variables for config, %w", err)
	}

	err = parseFlags(cfg, args)

	if err != nil {
		return nil, fmt.Errorf("error parsing flags for config, %w", err)
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

	flagCfg := &Config{}

	fs.StringVar(&flagCfg.AppAddr, "a", "", "address of HTTP server")
	fs.StringVar(&flagCfg.DBConnectionString, "d", "", "db connection string url")
	fs.StringVar(&flagCfg.AccrualSystemAddr, "r", "", "accrual system address")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if flagCfg.AppAddr != "" {
		cfg.AppAddr = flagCfg.AppAddr
	}
	if flagCfg.DBConnectionString != "" {
		cfg.DBConnectionString = flagCfg.DBConnectionString
	}
	if flagCfg.AccrualSystemAddr != "" {
		cfg.AccrualSystemAddr = flagCfg.AccrualSystemAddr
	}

	return nil
}
