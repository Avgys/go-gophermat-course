package repository

import (
	"avgys-gophermat/internal/db"
	balancerepository "avgys-gophermat/sqlc/balance"
	"context"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type BalanceRepository struct {
	db         *db.DB
	repository *balancerepository.Queries
}

func NewBalanceRepository(db *db.DB) *BalanceRepository {
	queries := balancerepository.New(db.Pool)

	return &BalanceRepository{db: db, repository: queries}
}

func (r *BalanceRepository) GetBalance(ctx context.Context, userID int64) (balancerepository.Balance, error) {
	ctxTimeout, cancel := context.WithTimeout(ctx, operationTimeout)
	defer cancel()

	return r.repository.GetBalance(ctxTimeout, userID)
}

func (r *BalanceRepository) GetWithdrawals(ctx context.Context, userID int64) ([]balancerepository.GetWithdrawalsRow, error) {
	ctxTimeout, cancel := context.WithTimeout(ctx, operationTimeout)
	defer cancel()

	return r.repository.GetWithdrawals(ctxTimeout, userID)
}

func (r *BalanceRepository) TryDecreaseBalance(ctx context.Context, userID int64, amount float32) (bool, error) {
	ctxTimeout, cancel := context.WithTimeout(ctx, operationTimeout)
	defer cancel()

	tx, err := r.db.Pool.BeginTx(ctxTimeout, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return false, err
	}

	defer tx.Rollback(ctxTimeout)

	n := pgtype.Numeric{}
	_ = n.ScanScientific(strconv.FormatFloat(float64(amount), 'f', -1, 64))

	result, err := r.repository.WithTx(tx).TryDecreaseBalance(ctxTimeout, balancerepository.TryDecreaseBalanceParams{UserID: userID, Balance: n})
	if err != nil {
		return false, err
	}

	if err := tx.Commit(ctxTimeout); err != nil {
		return false, err
	}

	return result, nil
}
