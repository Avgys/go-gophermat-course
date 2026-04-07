package endpoints

import (
	"avgys-gophermat/internal/service/balance"
	"avgys-gophermat/internal/service/orders"
)

type Endpoints struct {
	AuthService
	*orders.OrderService
	*balance.BalanceService
}

func New(authservice AuthService, orderService *orders.OrderService, balanceService *balance.BalanceService) *Endpoints {
	return &Endpoints{AuthService: authservice, OrderService: orderService, BalanceService: balanceService}
}
