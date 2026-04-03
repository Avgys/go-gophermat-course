package repository

import (
	"avgys-gophermat/internal/db"
	"avgys-gophermat/internal/model/order"
	orderrepository "avgys-gophermat/sqlc/order"
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type Order struct {
	OrderNum  int64
	Status    int32
	Accrual   pgtype.Numeric
	UserID    int64
	CreatedAt pgtype.Timestamp
}

type OrderRepository struct {
	orderRepository   *orderrepository.Queries
	balanceRepository *BalanceRepository
}

func NewOrderRepository(db *db.DB) *OrderRepository {
	queries := orderrepository.New(db.Pool)
	balancerepository := NewBalanceRepository(db)

	return &OrderRepository{queries, balancerepository}
}

func (r *OrderRepository) GetOrAddEmptyOrder(ctx context.Context, params *orderrepository.GetOrAddOrderParams) (orderrepository.GetOrAddOrderRow, error) {
	ctxTimeout, cancel := context.WithTimeout(ctx, operationTimeout)
	defer cancel()

	return r.orderRepository.GetOrAddOrder(ctxTimeout, *params)
}

func (r *OrderRepository) GetOrdersByUser(ctx context.Context, userID int64) ([]orderrepository.Order, error) {
	ctxTimeout, cancel := context.WithTimeout(ctx, operationTimeout)
	defer cancel()

	return r.orderRepository.GetOrdersByUser(ctxTimeout, userID)
}

func (r *OrderRepository) GetUnproccessedOrders(ctx context.Context, limit int32) ([]orderrepository.Order, error) {
	ctxTimeout, cancel := context.WithTimeout(ctx, operationTimeout)
	defer cancel()

	return r.orderRepository.GetUnproccessedOrders(ctxTimeout, limit)
}

func (r *OrderRepository) UpdateOrder(ctx context.Context, arg *orderrepository.UpdateOrderParams) error {
	ctxTimeout, cancel := context.WithTimeout(ctx, operationTimeout)
	defer cancel()

	tx, err := r.balanceRepository.db.Pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	row, err := r.orderRepository.UpdateOrder(ctxTimeout, *arg)

	if err != nil {
		return err
	}

	if row.Status == int32(order.StatusProcessed) {
		_, err := r.balanceRepository.tryAddDeltaWithTx(ctx, tx, row.UserID, row.Accrual)

		if err != nil {
			return err
		}
	}

	if err := tx.Commit(ctxTimeout); err != nil {
		return err
	}

	return nil
}
