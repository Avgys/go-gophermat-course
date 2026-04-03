package repository

import (
	"avgys-gophermat/internal/db"
	balancerepository "avgys-gophermat/sqlc/balance"
	"context"
	"errors"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	ErrInsufficientBalance = errors.New("insufficient balance")
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

	n := pgtype.Numeric{}
	_ = n.ScanScientific(strconv.FormatFloat(float64(amount), 'f', -1, 64))
	result, err := r.tryAddDeltaWithTx(ctx, tx, userID, n)

	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctxTimeout); err != nil {
		return nil, err
	}

	return result, nil
}

func (r *BalanceRepository) tryAddDeltaWithTx(ctx context.Context, tx pgx.Tx, userID int64, amount pgtype.Numeric) (*TryAddDeltaRow, error) {
	ctxTimeout, cancel := context.WithTimeout(ctx, operationTimeout)
	defer cancel()

	const sql = "SELECT * FROM public.try_add_delta($1, $2);"
	row := tx.QueryRow(ctxTimeout, sql, userID, amount)

	var result TryAddDeltaRow

	if err := row.Scan(&result.Modified, &result.NewAmount, &result.OldAmount); err != nil {

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "22003" {
			return nil, ErrInsufficientBalance
		}

		return nil, err
	}

	return &result, nil
}

func (r *BalanceRepository) Withdraw(ctx context.Context, userID int64, amount float32, orderNum int64) (*TryAddDeltaRow, error) {
	ctxTimeout, cancel := context.WithTimeout(ctx, operationTimeout)
	defer cancel()

	//Invert to be less than zero
	amount = -amount

	tx, err := r.db.Pool.BeginTx(ctxTimeout, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return nil, err
	}

	defer tx.Rollback(ctxTimeout)

	n := pgtype.Numeric{}
	_ = n.ScanScientific(strconv.FormatFloat(float64(amount), 'f', -1, 64))
	result, err := r.tryAddDeltaWithTx(ctx, tx, userID, n)

	if err != nil {
		return nil, err
	}

	txQuery := r.repository.WithTx(tx)

	n.Int.Neg(n.Int)
	err = txQuery.InsertWithdrawal(ctx, balancerepository.InsertWithdrawalParams{UserID: userID, WithdrawAmount: n, OrderNum: orderNum})

	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctxTimeout); err != nil {
		return nil, err
	}

	return result, nil
}
