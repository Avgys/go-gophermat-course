package accrualclient

import (
	"avgys-gophermat/internal/model/responses"
	"context"
)

type AccrualClient interface {
	Send(ctx context.Context, orderNum string) (*responses.AccrualOrder, error)
}

//go:generate mockgen -source=accrual_service.go -destination=tests/mocks/accrual_service_mock.go -package=mocks
