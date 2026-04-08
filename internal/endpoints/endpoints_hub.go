package endpoints

type Endpoints struct {
	AuthService
	OrderService
	BalanceService
}

func New(authservice AuthService, orderService OrderService, balanceService BalanceService) *Endpoints {
	return &Endpoints{AuthService: authservice, OrderService: orderService, BalanceService: balanceService}
}
