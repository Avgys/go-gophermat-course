package orders

import (
	"avgys-gophermat/internal/model"
	"avgys-gophermat/internal/model/order"
	"avgys-gophermat/internal/repository"
	"avgys-gophermat/internal/service/accrualclient"
	"avgys-gophermat/internal/service/auth"
	httphelper "avgys-gophermat/internal/shared/http"
	orderrepository "avgys-gophermat/sqlc/order"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type OrderService struct {
	orderRepository *repository.OrderRepository
	accrualService  *accrualclient.AccrualService
}

func NewOrderService(resository *repository.OrderRepository, accrualService *accrualclient.AccrualService) *OrderService {
	return &OrderService{orderRepository: resository, accrualService: accrualService}
}

func (a *OrderService) Store(ctx context.Context, userClaims *auth.TokenClaims, orderNum string) error {

	userId := userClaims.UserID

	if err := LuhnNumVerify(orderNum); err != nil {
		return fmt.Errorf("%w inner %w", httphelper.NewError("order number is invalid", http.StatusUnprocessableEntity), err)
	}

	orderNum64, err := strconv.ParseInt(orderNum, 10, 64)
	if err != nil {
		return err
	}

	order := &orderrepository.GetOrAddOrderParams{
		OrderNum: orderNum64,
		Status:   int32(order.StatusNew),
		UserID:   userId,
	}

	row, err := a.orderRepository.Store(ctx, order)

	if err != nil {
		return err
	}

	if row.UserID != userId {
		return httphelper.NewError("order is already registered by another user", http.StatusConflict)
	}

	if !row.IsNew {
		return httphelper.NewError("order is already in processing", http.StatusAccepted)
	}

	return nil
}

func (a *OrderService) Send(ctx context.Context, userClaims *auth.TokenClaims, orderNum string) (*model.AccrualResponse, error) {

	if err := LuhnNumVerify(orderNum); err != nil {
		return nil, httphelper.NewError("order number is invalid", http.StatusBadRequest)
	}

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

func LuhnNumVerify(num string) error {
	digits := strings.Split(num, "")

	mod := len(num) % 2
	sum := 0

	for i, f := range digits {

		digit, err := strconv.Atoi(f)

		if err != nil {
			return err
		}

		if i%2 == mod {
			sum += digit * 2 % 9
		} else {
			sum += digit
		}
	}

	if sum%10 != 0 {
		return fmt.Errorf("invalid code")
	}

	return nil
}
