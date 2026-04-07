package endpoints

import (
	"avgys-gophermat/internal/service/balance"
)

type Endpoints struct {
	AuthService
	OrderService
	*balance.BalanceService
}

func New(authservice AuthService, orderService OrderService, balanceService *balance.BalanceService) *Endpoints {
	return &Endpoints{AuthService: authservice, OrderService: orderService, BalanceService: balanceService}
}
