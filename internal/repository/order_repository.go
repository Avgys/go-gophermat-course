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

func (r *OrderRepository) Store(ctx context.Context, f *orderrepository.GetOrAddOrderParams) (orderrepository.GetOrAddOrderRow, error) {
	ctxTimeout, cancel := context.WithTimeout(ctx, operationTimeout)
	defer cancel()

	userID, err := r.repository.GetOrAddOrder(ctxTimeout, *f)

	return userID, err
}
