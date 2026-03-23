package repository

import (
	"avgys-gophermat/internal/model"
	userrepository "avgys-gophermat/sqlc/user"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func (s *DBStore) InsertUser(ctx context.Context, user *model.UserModel) (int64, error) {

	ctxTimeout, cancel := context.WithTimeout(ctx, operationTimeout)
	defer cancel()

	userToCreate := userrepository.CreateUserParams{
		Login:        user.Login,
		HashSalt:     user.HashSalt,
		PasswordHash: user.PasswordHash}

	userID, err := s.repository.CreateUser(ctxTimeout, userToCreate)

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return 0, ErrUserAlreadyExists
	}

	return userID, err
}

func (s *DBStore) GetUserByLogin(ctx context.Context, userLogin string) (*model.UserModel, error) {

	ctxTimeout, cancel := context.WithTimeout(ctx, operationTimeout)
	defer cancel()

	user, err := s.repository.GetUserByLogin(ctxTimeout, userLogin)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}

		return nil, fmt.Errorf("scan error: %w", err)
	}

	return &model.UserModel{
		Login:        user.Login,
		HashSalt:     user.HashSalt,
		PasswordHash: user.PasswordHash,
		CreatedAtUTC: user.CreatedAt.Time}, nil
}
