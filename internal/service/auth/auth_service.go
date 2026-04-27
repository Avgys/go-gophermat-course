package auth

import (
	"avgys-gophermat/internal/model"
	"avgys-gophermat/internal/model/requests"
	"avgys-gophermat/internal/repository"
	"avgys-gophermat/internal/service/generator"
	httphelper "avgys-gophermat/internal/shared/http"
	"context"
	"encoding/hex"
	"errors"
	"net/http"
)

type AuthService struct {
	repository *repository.AuthRepository
}

func NewAuthService(resository *repository.AuthRepository) *AuthService {
	return &AuthService{resository}
}

func (a *AuthService) Register(ctx context.Context, user *requests.UserRq) (*TokenClaims, error) {

	dbUser := model.UserModel{Login: user.Login}
	var err error

	dbUser.PasswordHash, dbUser.HashSalt, err = generator.GetHashWithRandomSalt(user.Password)

	if err != nil {
		return nil, err
	}

	userID, err := a.repository.InsertUser(ctx, &dbUser)

	if err != nil {
		if errors.Is(err, repository.ErrUserAlreadyExists) {
			return nil, httphelper.NewError(err.Error(), http.StatusConflict)
		}

		return nil, err
	}

	return NewToken(userID, dbUser.Login), nil
}

func (a *AuthService) Login(ctx context.Context, user *requests.UserRq) (*TokenClaims, error) {
	dbUser, err := a.repository.GetUserByLogin(ctx, user.Login)

	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, httphelper.NewError(http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		}

		return nil, err
	}

	salt, _ := hex.DecodeString(dbUser.HashSalt)

	loginHash, _ := generator.GetHashWithSalt(user.Password, salt)

	if dbUser.PasswordHash != loginHash {
		return nil, httphelper.NewError(http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
	}

	return NewToken(dbUser.ID, dbUser.Login), nil
}
