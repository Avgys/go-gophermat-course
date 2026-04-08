package processor

import (
	"avgys-gophermat/internal/model/responses"
	orderrepository "avgys-gophermat/sqlc/order"
	"context"
)

type OrderService interface {
	GetOrderUnprocessedOrders(ctx context.Context, limit int) ([]orderrepository.Order, error)
	UpdateOrderStatus(ctx context.Context, accrualOrder *responses.AccrualOrder) error
}

//go:generate mockgen -source=order_service.go -destination=tests/mocks/order_service_mock.go -package=mocks
