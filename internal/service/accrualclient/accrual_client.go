package accrualclient

import (
	"avgys-gophermat/internal/config"
	"avgys-gophermat/internal/model/responses"
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/rs/zerolog"
	"resty.dev/v3"
)

type ErrRetryAfter struct {
	error
	RetryAfter int64
}

var (
	ErrOrderNotExists = errors.New("order not exists")
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

func (s *AccrualService) Send(ctx context.Context, orderNum string, logger *zerolog.Logger) (*responses.AccrualOrder, error) {

	if s == nil || s.restClient == nil {
		return nil, fmt.Errorf("accrual service client is nil")
	}
	if s.accrualSystemAddr == "" {
		return nil, fmt.Errorf("accrual system address is empty")
	}

	url := fmt.Sprintf("%s/api/orders/%s", s.accrualSystemAddr, orderNum)

	var accrualResp responses.AccrualOrder

	logger.Debug().Str("order", orderNum).Str("url", url).Msg("sending request to accrual server")

	resp, err := s.restClient.R().
		SetContext(ctx).
		SetResult(&accrualResp).
		Get(url)

	if err != nil {
		logger.Error().Err(err).Str("order", orderNum).Msg("accrual request failed")
		return nil, err
	}

	defer resp.Body.Close()

	logger.Debug().Str("order", orderNum).Int("status", resp.StatusCode()).Msg("accrual server response")

	if resp.StatusCode() == http.StatusTooManyRequests {
		retryAfterSecs := resp.Header().Get("Retry-After")
		secs, _ := strconv.ParseInt(retryAfterSecs, 10, 32)
		logger.Warn().Str("order", orderNum).Int64("retry_after", secs).Msg("accrual server rate limited")
		return nil, ErrRetryAfter{error: errors.New("too many requests"), RetryAfter: secs}
	}

	if resp.StatusCode() == http.StatusNoContent {
		logger.Warn().Str("order", orderNum).Msg("order not found in accrual server")
		return nil, ErrOrderNotExists
	}

	if resp.StatusCode() == http.StatusOK {
		logger.Info().Str("order", orderNum).Str("status", accrualResp.Status).Float64("accrual", accrualResp.Accrual).Msg("accrual response received")
		return &accrualResp, nil
	}

	err = fmt.Errorf("unsupported error %s", resp.Status())
	logger.Error().Err(err).Str("order", orderNum).Msg("unexpected accrual server response")
	return nil, err
}
