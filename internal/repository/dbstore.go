package repository

import (
	"avgys-gophermat/internal/repository/db"
	"context"
	"time"

	"github.com/rs/zerolog"
)

type DBStore struct {
	db *db.DB
}

const dbOpTimeout = 1 * time.Second

func NewDBStore(ctx context.Context, dbConfig *db.Config, logger *zerolog.Logger) (*DBStore, error) {
	dbConnection, err := db.NewDB(ctx, dbConfig)

	if err != nil {
		return nil, err
	}

	return &DBStore{db: dbConnection}, nil
}

func (s *DBStore) TestConnection(ctx context.Context) error {
	return s.db.Ping(ctx)
}

func (s *DBStore) Close() error {
	s.db.Close()
	return nil
}
