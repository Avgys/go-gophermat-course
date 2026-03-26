package accrualclient

import (
	"avgys-gophermat/internal/config"
	"context"
	"fmt"

	"resty.dev/v3"
)

type AccrualService struct {
	*resty.Client
	AccrualSystemAddr string
}

func NewAccrualService(ctx context.Context, cfg *config.Config) *AccrualService {
	client := resty.New()

	accrualAddr := ""
	if cfg != nil {
		accrualAddr = cfg.AccrualSystemAddr
	}
	service := &AccrualService{
		Client:            client,
		AccrualSystemAddr: accrualAddr,
	}

	if ctx != nil {
		go func() {
			<-ctx.Done()
			_ = client.Close()
		}()
	}

	return service
}

func (s *AccrualService) PostToAccrual(ctx context.Context, orderNum int) (*resty.Response, error) {
	if s == nil || s.Client == nil {
		return nil, fmt.Errorf("accrual service client is nil")
	}
	if s.AccrualSystemAddr == "" {
		return nil, fmt.Errorf("accrual system address is empty")
	}

	url := fmt.Sprintf("http://%s/api/orders/%v", s.AccrualSystemAddr, orderNum)

	return s.R().
		SetContext(ctx).
		Get(url)
}
