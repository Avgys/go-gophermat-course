package orders

import (
	orderrepository "avgys-gophermat/sqlc/order"
	"context"
)

type OrderRepository interface {
	GetOrAddEmptyOrder(ctx context.Context, params *orderrepository.GetOrAddOrderParams) (orderrepository.GetOrAddOrderRow, error)
	GetOrdersByUser(ctx context.Context, userID int64) ([]orderrepository.Order, error)
	GetUnproccessedOrders(ctx context.Context, limit int32) ([]orderrepository.Order, error)
	UpdateOrderAndIncreaseBalance(ctx context.Context, arg *orderrepository.UpdateOrderParams) error
}

//go:generate mockgen -source=order_repository.go -destination=tests/mocks/order_repository_mock.go -package=mocks
