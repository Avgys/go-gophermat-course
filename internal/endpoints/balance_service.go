package endpoints

import (
	"avgys-gophermat/internal/model/requests"
	"avgys-gophermat/internal/model/responses"
	"avgys-gophermat/internal/service/auth"
	"context"
)

type BalanceService interface {
	GetBalanceByUserID(ctx context.Context, userClaims *auth.TokenClaims) (*responses.Balance, error)
	Withdraw(ctx context.Context, userClaims *auth.TokenClaims, withdraw *requests.WithdrawRq) (*responses.WithdrawDeltaRs, error)
	GetWithdrawals(ctx context.Context, userClaims *auth.TokenClaims) ([]responses.WithdrawRs, error)
}

//go:generate mockgen -source=balance_service.go -destination=tests/mocks/balance_service_mock.go -package=mocks
