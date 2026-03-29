package repository

import (
	"avgys-gophermat/internal/db"
	"avgys-gophermat/internal/model"
	userrepository "avgys-gophermat/sqlc/user"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
)

const operationTimeout = 2 * time.Second

type AuthRepository struct {
	repository *userrepository.Queries
}

func NewAuthRepository(db *db.DB) *AuthRepository {
	queries := userrepository.New(db.Pool)
	return &AuthRepository{queries}
}

func (s *AuthRepository) InsertUser(ctx context.Context, user *model.UserModel) (int64, error) {

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

func (s *AuthRepository) GetUserByLogin(ctx context.Context, userLogin string) (*model.UserModel, error) {

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
		ID:           user.ID,
		Login:        user.Login,
		HashSalt:     user.HashSalt,
		PasswordHash: user.PasswordHash,
		CreatedAtUTC: user.CreatedAt.Time}, nil
}
