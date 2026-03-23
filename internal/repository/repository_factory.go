package repository

import (
	"avgys-gophermat/internal/config"
	"avgys-gophermat/internal/db"
	"avgys-gophermat/internal/model"
	userrepository "avgys-gophermat/sqlc/user"
	"context"
	"errors"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
)

type Repository interface {
	InsertUser(ctx context.Context, user *model.UserModel) (int64, error)
	GetUserByLogin(ctx context.Context, userLogin string) (*model.UserModel, error)

	// TestConnection(ctx context.Context) error
}

func CreateRepository(ctx context.Context, cfg *config.Config) (Repository, error) {

	if cfg.DBConnectionString == "" {
		return nil, errors.New("no db connection string provided")
	}

	dbConnection, err := db.NewDB(ctx, &db.Config{ConnectionString: cfg.DBConnectionString})

	if err != nil {
		return nil, err
	}

	queries := userrepository.New(dbConnection.Pool)

	return NewRepository(queries), nil
}
