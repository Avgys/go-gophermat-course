package repository

import (
	"avgys-gophermat/internal/db"
	orderrepository "avgys-gophermat/sqlc/order"
	"context"

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
	repository *orderrepository.Queries
}

func NewOrderRepository(db *db.DB) *OrderRepository {
	queries := orderrepository.New(db.Pool)
	return &OrderRepository{queries}
}

func (r *OrderRepository) Store(ctx context.Context, params *orderrepository.GetOrAddOrderParams) (orderrepository.GetOrAddOrderRow, error) {
	ctxTimeout, cancel := context.WithTimeout(ctx, operationTimeout)
	defer cancel()

	return r.repository.GetOrAddOrder(ctxTimeout, *params)
}

func (r *OrderRepository) GetOrdersByUser(ctx context.Context, userID int64) ([]orderrepository.Order, error) {
	ctxTimeout, cancel := context.WithTimeout(ctx, operationTimeout)
	defer cancel()

	return r.repository.GetOrdersByUser(ctxTimeout, userID)
}

func (r *OrderRepository) GetUnproccessedOrders(ctx context.Context, limit int32) ([]orderrepository.Order, error) {
	ctxTimeout, cancel := context.WithTimeout(ctx, operationTimeout)
	defer cancel()

	return r.repository.GetUnproccessedOrders(ctxTimeout, limit)
}

func (r *OrderRepository) UpdateOrder(ctx context.Context, arg *orderrepository.UpdateOrderParams) error {
	ctxTimeout, cancel := context.WithTimeout(ctx, operationTimeout)
	defer cancel()

	return r.repository.UpdateOrder(ctxTimeout, *arg)
}
