package endpoints

import (
	"avgys-gophermat/internal/service/auth"
	"avgys-gophermat/internal/service/balance"
	"avgys-gophermat/internal/service/orders"
)

type Endpoints struct {
	*auth.AuthService
	*orders.OrderService
	*balance.BalanceService
}

func New(authservice *auth.AuthService, orderService *orders.OrderService, balanceService *balance.BalanceService) *Endpoints {
	return &Endpoints{AuthService: authservice, OrderService: orderService, BalanceService: balanceService}
}
