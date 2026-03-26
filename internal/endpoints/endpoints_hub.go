package endpoints

import (
	"avgys-gophermat/internal/service/auth"
	"avgys-gophermat/internal/service/orders"
)

type Endpoints struct {
	*auth.AuthService
	*orders.OrderService
}

func New(authservice *auth.AuthService, orderService *orders.OrderService) *Endpoints {
	return &Endpoints{AuthService: authservice, OrderService: orderService}
}
