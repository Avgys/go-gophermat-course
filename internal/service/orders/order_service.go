package orders

import (
	"avgys-gophermat/internal/model"
	"avgys-gophermat/internal/model/order"
	"avgys-gophermat/internal/model/response"
	"avgys-gophermat/internal/repository"
	"avgys-gophermat/internal/service"
	"avgys-gophermat/internal/service/accrualclient"
	"avgys-gophermat/internal/service/auth"
	httphelper "avgys-gophermat/internal/shared/http"
	orderrepository "avgys-gophermat/sqlc/order"
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/samber/lo"
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

	row, err := a.orderRepository.GetOrAddEmptyOrder(ctx, order)

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

func (a *OrderService) GetOrderByUserID(ctx context.Context, userClaims *auth.TokenClaims) ([]response.Order, error) {

	userId := userClaims.UserID

	rows, err := a.orderRepository.GetOrdersByUser(ctx, userId)

	if err != nil {
		return nil, err
	}

	orders := lo.Map(rows, func(row orderrepository.Order, _ int) response.Order {

		return response.Order{
			OrderNum:     row.OrderNum,
			Status:       order.OrderStatus(row.Status).String(),
			Accrual:      service.NumericToStr(row.Accrual),
			CreatedAtUTC: row.CreatedAt.Time.Format(time.RFC3339),
		}
	})

	return orders, nil
}

func (a *OrderService) GetOrderUnprocessedOrders(ctx context.Context, limit int) ([]orderrepository.Order, error) {
	if limit == 0 {
		return []orderrepository.Order{}, nil
	}

	return a.orderRepository.GetUnproccessedOrders(ctx, int32(limit))
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

func (a *OrderService) UpdateStatus(ctx context.Context, accrualOrder *model.AccrualOrder) error {

	orderNum64, err := strconv.ParseInt(accrualOrder.OrderNum, 10, 64)
	if err != nil {
		return err
	}

	var orderStatus order.OrderStatus
	orderStatus.Parse(accrualOrder.Status)

	n := pgtype.Numeric{}
	_ = n.ScanScientific(strconv.FormatFloat(float64(accrualOrder.Accrual), 'f', -1, 64))

	order := &orderrepository.UpdateOrderParams{
		OrderNum: orderNum64,
		Status:   int32(orderStatus),
		Accrual:  n,
	}

	return a.orderRepository.UpdateOrder(ctx, order)
}
