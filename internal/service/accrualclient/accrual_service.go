package accrualclient

import (
	"avgys-gophermat/internal/model/responses"
	"context"

	"github.com/rs/zerolog"
)

type AccrualClient interface {
	Send(ctx context.Context, orderNum string, logger *zerolog.Logger) (*responses.AccrualOrder, error)
}

//go:generate mockgen -source=accrual_service.go -destination=tests/mocks/accrual_service_mock.go -package=mocks
