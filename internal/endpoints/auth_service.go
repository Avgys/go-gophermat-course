package endpoints

import (
	"avgys-gophermat/internal/model/requests"
	"avgys-gophermat/internal/service/auth"
	"context"
)

type AuthService interface {
	Register(ctx context.Context, user *requests.UserRq) (*auth.TokenClaims, error)
	Login(ctx context.Context, user *requests.UserRq) (*auth.TokenClaims, error)
}

//go:generate mockgen -source=auth_service.go -destination=tests/mocks/auth_service_mock.go -package=mocks
