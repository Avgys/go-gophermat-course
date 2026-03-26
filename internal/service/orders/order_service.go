package orders

import (
	"avgys-gophermat/internal/model"
	"avgys-gophermat/internal/repository"
	"avgys-gophermat/internal/service/accrualclient"
	"avgys-gophermat/internal/service/auth"
	httphelper "avgys-gophermat/internal/shared/http"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type OrderService struct {
	repository     repository.Repository
	accrualService *accrualclient.AccrualService
}

func NewOrderService(resository repository.Repository, accrualService *accrualclient.AccrualService) *OrderService {
	return &OrderService{repository: resository, accrualService: accrualService}
}

func (a *OrderService) Load(ctx context.Context, userClaims *auth.TokenClaims, orderNum int) (*model.AccrualResponse, error) {
	resp, err := a.accrualService.PostToAccrual(ctx, orderNum)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode() == http.StatusNoContent {
		return nil, httphelper.NewError("no order found", http.StatusNoContent)
	}

	if resp.StatusCode() == http.StatusOK {

		buf, err := io.ReadAll(resp.Body)

		if err != nil {
			return nil, fmt.Errorf("read accrual response body: %w", err)
		}

		var accrualResp model.AccrualResponse
		if err := json.Unmarshal(buf, &accrualResp); err != nil {
			return nil, fmt.Errorf("decode accrual response: %w", err)
		}

		return &accrualResp, nil
	}

	return nil, fmt.Errorf("unsupported error %s", resp.Status())
}
