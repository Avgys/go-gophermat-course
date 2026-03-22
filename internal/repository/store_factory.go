package repository

import (
	"avgys-gophermat/internal/config"
	"context"
	"errors"

	"avgys-gophermat/internal/repository/db"

	"github.com/rs/zerolog"
)

type Full2ShortBatch map[string]string

type Repository interface {
	TestConnection(ctx context.Context) error
	Close() error
}

func NewRepository(ctx context.Context, cfg *config.Config, logger *zerolog.Logger) (Repository, error) {

	if cfg.DBConnectionString == "" {
		return nil, errors.New("no db connection string provided")
	}

	return NewDBStore(ctx, &db.Config{ConnectionString: cfg.DBConnectionString}, logger)
}
