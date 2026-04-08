package balance

import (
	"avgys-gophermat/internal/repository"
	balancerepository "avgys-gophermat/sqlc/balance"
	"context"
)

type BalanceRepository interface {
	GetBalance(ctx context.Context, userID int64) (balancerepository.GetBalanceRow, error)
	GetWithdrawals(ctx context.Context, userID int64) ([]balancerepository.GetWithdrawalsRow, error)
	Withdraw(ctx context.Context, userID int64, amount float32, orderNum int64) (*repository.TryAddDeltaRow, error)
}

//go:generate mockgen -source=balance_repository.go -destination=tests/mocks/balance_repository_mock.go -package=mocks
