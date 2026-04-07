package accrualclient

import (
	"avgys-gophermat/internal/config"
	"avgys-gophermat/internal/model/order"
	"avgys-gophermat/internal/model/responses"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"

	"resty.dev/v3"
)

var (
	ErrTooManyRequests = errors.New("too many requests")
	ErrOrderNotExists  = errors.New("order not exists")
)

type AccrualService struct {
	restClient        *resty.Client
	accrualSystemAddr string
}

func NewAccrualService(ctx context.Context, cfg *config.Config) *AccrualService {
	client := resty.New()

	accrualAddr := ""
	if cfg != nil {
		accrualAddr = cfg.AccrualSystemAddr
	}

	service := &AccrualService{
		restClient:        client,
		accrualSystemAddr: accrualAddr,
	}

	if ctx != nil {
		go func() {
			<-ctx.Done()
			_ = client.Close()
		}()
	}

	return service
}

func (s *AccrualService) postToAccrual(ctx context.Context, orderNum string) (*resty.Response, error) {
	if s == nil || s.restClient == nil {
		return nil, fmt.Errorf("accrual service client is nil")
	}
	if s.accrualSystemAddr == "" {
		return nil, fmt.Errorf("accrual system address is empty")
	}

	url := fmt.Sprintf("http://%s/api/orders/%s", s.accrualSystemAddr, orderNum)

	return s.restClient.R().
		SetContext(ctx).
		Get(url)
}

func (s *AccrualService) Send(ctx context.Context, orderNum string) (*responses.AccrualOrder, error) {

	resp, err := s.postToAccrual(ctx, orderNum)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode() == http.StatusTooManyRequests {
		resp.Header().Get("Retry-After")
		return nil, ErrTooManyRequests
	}

	if resp.StatusCode() == http.StatusNoContent {
		return nil, ErrOrderNotExists
	}

	if resp.StatusCode() == http.StatusOK {

		buf, err := io.ReadAll(resp.Body)

		if err != nil {
			return nil, fmt.Errorf("read accrual response body: %w", err)
		}

		var accrualResp responses.AccrualOrder
		if err := json.Unmarshal(buf, &accrualResp); err != nil {
			return nil, fmt.Errorf("decode accrual response: %w", err)
		}

		return &accrualResp, nil

	}

	return nil, fmt.Errorf("unsupported error %s", resp.Status())

	statusName := order.StatusName[order.OrderStatus(rand.Intn(4))]

	result := &responses.AccrualOrder{OrderNum: orderNum, Accrual: rand.Float32() * 500, Status: statusName}

	if result.Accrual < 0 {
		return nil, errors.New("order accrual must be larger than zero")
	}

	return result, nil
}
