package orders

import (
	"avgys-gophermat/internal/model/order"
	"avgys-gophermat/internal/model/responses"
	"avgys-gophermat/internal/service"
	"avgys-gophermat/internal/service/accrualclient"
	"avgys-gophermat/internal/service/auth"
	"avgys-gophermat/internal/service/validation"
	httphelper "avgys-gophermat/internal/shared/http"
	orderrepository "avgys-gophermat/sqlc/order"
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/samber/lo"
)

type OrderService struct {
	orderRepository OrderRepository
	accrualService  accrualclient.AccrualClient
}

func NewOrderService(resository OrderRepository, accrualService accrualclient.AccrualClient) *OrderService {
	return &OrderService{orderRepository: resository, accrualService: accrualService}
}

func (a *OrderService) Store(ctx context.Context, userClaims *auth.TokenClaims, orderNum string) error {

	userId := userClaims.UserID

	if err := validation.LuhnNumVerify(orderNum); err != nil {
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
		return httphelper.NewError("order is already in processing", http.StatusOK)
	}

	return nil
}

func (a *OrderService) GetOrderByUserID(ctx context.Context, userClaims *auth.TokenClaims) ([]responses.Order, error) {

	userId := userClaims.UserID

	rows, err := a.orderRepository.GetOrdersByUser(ctx, userId)

	if err != nil {
		return nil, err
	}

	orders := lo.Map(rows, func(row orderrepository.Order, _ int) responses.Order {
		return responses.Order{
			OrderNum:     strconv.FormatInt(row.OrderNum, 10),
			Status:       order.OrderStatus(row.Status).String(),
			Accrual:      service.NumericToFloat(row.Accrual),
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

func (a *OrderService) UpdateOrderStatus(ctx context.Context, accrualOrder *responses.AccrualOrder) error {

	orderNum64, err := strconv.ParseInt(accrualOrder.OrderNum, 10, 64)
	if err != nil {
		return err
	}

	var orderStatus order.OrderStatus
	orderStatus.Parse(accrualOrder.Status)

	n := pgtype.Numeric{}
	_ = n.ScanScientific(strconv.FormatFloat(accrualOrder.Accrual, 'f', -1, 64))

	order := &orderrepository.UpdateOrderParams{
		OrderNum: orderNum64,
		Status:   int32(orderStatus),
		Accrual:  n,
	}

	return a.orderRepository.UpdateOrderAndIncreaseBalance(ctx, order)
}
