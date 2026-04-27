package endpoints

type Endpoints struct {
	AuthService    AuthService
	OrderService   OrderService
	BalanceService BalanceService
}

func New(authservice AuthService, orderService OrderService, balanceService BalanceService) *Endpoints {
	return &Endpoints{AuthService: authservice, OrderService: orderService, BalanceService: balanceService}
}
