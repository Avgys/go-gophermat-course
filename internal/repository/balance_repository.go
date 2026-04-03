package repository

import (
	"avgys-gophermat/internal/db"
	balancerepository "avgys-gophermat/sqlc/balance"
	"context"

	"github.com/jackc/pgx/v5"
)

type BalanceRepository struct {
	db         *db.DB
	repository *balancerepository.Queries
}

func NewBalanceRepository(db *db.DB) *BalanceRepository {
	queries := balancerepository.New(db.Pool)

	return &BalanceRepository{db: db, repository: queries}
}

func (r *BalanceRepository) GetBalance(ctx context.Context, userID int64) (balancerepository.GetBalanceRow, error) {
	ctxTimeout, cancel := context.WithTimeout(ctx, operationTimeout)
	defer cancel()

	return r.repository.GetBalance(ctxTimeout, userID)
}

func (r *BalanceRepository) GetWithdrawals(ctx context.Context, userID int64) ([]balancerepository.GetWithdrawalsRow, error) {
	ctxTimeout, cancel := context.WithTimeout(ctx, operationTimeout)
	defer cancel()

	return r.repository.GetWithdrawals(ctxTimeout, userID)
}

type TryAddDeltaRow struct {
	Modified  bool
	NewAmount float32
	OldAmount float32
}

func (r *BalanceRepository) TryAddDelta(ctx context.Context, userID int64, amount float32) (*TryAddDeltaRow, error) {
	ctxTimeout, cancel := context.WithTimeout(ctx, operationTimeout)
	defer cancel()

	tx, err := r.db.Pool.BeginTx(ctxTimeout, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return nil, err
	}

	defer tx.Rollback(ctxTimeout)

	// n := pgtype.Numeric{}
	// _ = n.ScanScientific(strconv.FormatFloat(float64(amount), 'f', -1, 64))

	const sql = "SELECT * FROM public.try_add_delta($1, $2);"
	row := tx.QueryRow(ctxTimeout, sql, struct {
		UserID int64
		Amount float32
	}{UserID: userID, Amount: amount})

	var result TryAddDeltaRow

	if err := row.Scan(&result.Modified, &result.NewAmount, &result.OldAmount); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctxTimeout); err != nil {
		return nil, err
	}

	return &result, nil
}
