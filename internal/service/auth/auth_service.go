package auth

import (
	"avgys-gophermat/internal/model"
	"avgys-gophermat/internal/repository"
	"avgys-gophermat/internal/service/generator"
	"context"
	"encoding/hex"
	"errors"
)

type AuthService struct {
	repository *repository.AuthRepository
}

func NewAuthService(resository *repository.AuthRepository) *AuthService {
	return &AuthService{resository}
}

var (
	ErrUnauthorized      = errors.New("unauthorized")
	ErrUserAlreadyExists = errors.New("user already exists")
)

func (a *AuthService) Register(ctx context.Context, user *model.UserApi) (*TokenClaims, error) {

	dbUser := model.UserModel{Login: user.Login}
	var err error

	dbUser.PasswordHash, dbUser.HashSalt, err = generator.GetHashWithRandomSalt(user.Password)

	if err != nil {
		return nil, err
	}

	userID, err := a.repository.InsertUser(ctx, &dbUser)

	if err != nil {
		if errors.Is(err, repository.ErrUserAlreadyExists) {
			return nil, ErrUserAlreadyExists
		}

		return nil, err
	}

	return NewToken(userID, dbUser.Login), nil
}

func (a *AuthService) Login(ctx context.Context, user *model.UserApi) (*TokenClaims, error) {
	dbUser, err := a.repository.GetUserByLogin(ctx, user.Login)

	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrUnauthorized
		}

		return nil, err
	}

	salt, _ := hex.DecodeString(dbUser.HashSalt)

	loginHash, _ := generator.GetHashWithSalt(user.Password, salt)

	if dbUser.PasswordHash != loginHash {
		return nil, ErrUnauthorized
	}

	return NewToken(dbUser.ID, dbUser.Login), nil
}
