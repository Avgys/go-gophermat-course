package endpoints

import "avgys-gophermat/internal/service/auth"

type Endpoints struct {
	*auth.AuthService
}

func New(authservice *auth.AuthService) *Endpoints {
	return &Endpoints{AuthService: authservice}
}
