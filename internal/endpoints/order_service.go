package endpoints

import (
	"avgys-gophermat/internal/model/responses"
	"avgys-gophermat/internal/service/auth"
	"context"
)

type OrderService interface {
	Store(ctx context.Context, userClaims *auth.TokenClaims, orderNum string) error
	GetOrderByUserID(ctx context.Context, userClaims *auth.TokenClaims) ([]responses.Order, error)
}

//go:generate mockgen -source=order_service.go -destination=tests/mocks/order_service_mock.go -package=mocks
