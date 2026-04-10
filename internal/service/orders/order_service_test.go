package orders

import (
	"avgys-gophermat/internal/model/order"
	"avgys-gophermat/internal/model/responses"
	"avgys-gophermat/internal/service"
	"avgys-gophermat/internal/service/auth"
	httphelper "avgys-gophermat/internal/shared/http"
	orderrepository "avgys-gophermat/sqlc/order"
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"avgys-gophermat/internal/service/orders/tests/mocks"

	"github.com/golang/mock/gomock"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/suite"
)

type orderServiceSuite struct {
	suite.Suite
	ctrl *gomock.Controller
	repo *mocks.MockOrderRepository
	svc  *OrderService
}

func (s *orderServiceSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.repo = mocks.NewMockOrderRepository(s.ctrl)
	s.svc = NewOrderService(s.repo, nil)
}

func (s *orderServiceSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *orderServiceSuite) TestStoreValid() {
	claims := auth.NewToken(7, "alice")
	orderNum := "79927398713"
	orderNum64 := int64(79927398713)

	s.repo.EXPECT().
		GetOrAddEmptyOrder(gomock.Any(), gomock.AssignableToTypeOf(&orderrepository.GetOrAddOrderParams{})).
		DoAndReturn(func(ctx context.Context, params *orderrepository.GetOrAddOrderParams) (orderrepository.GetOrAddOrderRow, error) {
			s.Require().Equal(orderNum64, params.OrderNum)
			s.Require().Equal(claims.UserID, params.UserID)
			s.Require().Equal(int32(order.StatusNew), params.Status)
			return orderrepository.GetOrAddOrderRow{OrderNum: orderNum64, UserID: claims.UserID, Status: int32(order.StatusNew), IsNew: true}, nil
		})

	s.NoError(s.svc.Store(context.Background(), claims, orderNum))
}

func (s *orderServiceSuite) TestStoreInvalidLuhn() {
	claims := auth.NewToken(7, "alice")
	err := s.svc.Store(context.Background(), claims, "79927398710")

	s.Error(err)
	var httpErr *httphelper.ShowHTTPError
	s.Require().True(errors.As(err, &httpErr))
	s.Equal(http.StatusUnprocessableEntity, httpErr.StatusCode)
}

func (s *orderServiceSuite) TestStoreConflict() {
	claims := auth.NewToken(7, "alice")
	orderNum := "79927398713"

	s.repo.EXPECT().
		GetOrAddEmptyOrder(gomock.Any(), gomock.AssignableToTypeOf(&orderrepository.GetOrAddOrderParams{})).
		Return(orderrepository.GetOrAddOrderRow{OrderNum: 79927398713, UserID: claims.UserID + 1, Status: int32(order.StatusNew), IsNew: true}, nil)

	err := s.svc.Store(context.Background(), claims, orderNum)

	s.Error(err)
	var httpErr *httphelper.ShowHTTPError
	s.Require().True(errors.As(err, &httpErr))
	s.Equal(http.StatusConflict, httpErr.StatusCode)
}

func (s *orderServiceSuite) TestStoreAccepted() {
	claims := auth.NewToken(7, "alice")
	orderNum := "79927398713"

	s.repo.EXPECT().
		GetOrAddEmptyOrder(gomock.Any(), gomock.AssignableToTypeOf(&orderrepository.GetOrAddOrderParams{})).
		Return(orderrepository.GetOrAddOrderRow{OrderNum: 79927398713, UserID: claims.UserID, Status: int32(order.StatusNew), IsNew: false}, nil)

	err := s.svc.Store(context.Background(), claims, orderNum)

	s.Error(err)
	var httpErr *httphelper.ShowHTTPError
	s.Require().True(errors.As(err, &httpErr))
	s.Equal(http.StatusAccepted, httpErr.StatusCode)
}

func (s *orderServiceSuite) TestStoreRepoError() {
	claims := auth.NewToken(7, "alice")
	orderNum := "79927398713"
	repoErr := errors.New("db error")

	s.repo.EXPECT().
		GetOrAddEmptyOrder(gomock.Any(), gomock.AssignableToTypeOf(&orderrepository.GetOrAddOrderParams{})).
		Return(orderrepository.GetOrAddOrderRow{}, repoErr)

	err := s.svc.Store(context.Background(), claims, orderNum)
	s.Error(err)
	s.True(errors.Is(err, repoErr))
}

func (s *orderServiceSuite) TestGetOrderByUserID() {
	claims := auth.NewToken(7, "alice")
	created := time.Date(2020, 12, 10, 15, 15, 45, 0, time.UTC)

	var accrual pgtype.Numeric
	_ = accrual.Scan("500.5")

	s.repo.EXPECT().
		GetOrdersByUser(gomock.Any(), claims.UserID).
		Return([]orderrepository.Order{{
			OrderNum:  9278923470,
			Status:    int32(order.StatusProcessed),
			Accrual:   accrual,
			UserID:    claims.UserID,
			CreatedAt: pgtype.Timestamp{Time: created, Valid: true},
		}}, nil)

	orders, err := s.svc.GetOrderByUserID(context.Background(), claims)
	s.NoError(err)

	expected := []responses.Order{{
		OrderNum:     "9278923470",
		Status:       order.StatusProcessed.String(),
		Accrual:      service.NumericToFloat(accrual),
		CreatedAtUTC: created.Format(time.RFC3339),
	}}

	s.Require().Len(orders, 1)
	s.Equal(expected[0], orders[0])
}

func (s *orderServiceSuite) TestGetOrderByUserIDError() {
	repoErr := errors.New("db error")
	s.repo.EXPECT().
		GetOrdersByUser(gomock.Any(), int64(7)).
		Return(nil, repoErr)

	_, err := s.svc.GetOrderByUserID(context.Background(), &auth.TokenClaims{UserID: 7})
	s.Error(err)
	s.True(errors.Is(err, repoErr))
}

func (s *orderServiceSuite) TestGetOrderUnprocessedOrders() {
	orders, err := s.svc.GetOrderUnprocessedOrders(context.Background(), 0)
	s.NoError(err)
	s.Empty(orders)

	s.repo.EXPECT().
		GetUnproccessedOrders(gomock.Any(), int32(2)).
		Return([]orderrepository.Order{{OrderNum: 1}}, nil)

	orders, err = s.svc.GetOrderUnprocessedOrders(context.Background(), 2)
	s.NoError(err)
	s.Require().Len(orders, 1)
	s.Equal(int64(1), orders[0].OrderNum)
}

func (s *orderServiceSuite) TestUpdateOrderStatus() {
	accrual := &responses.AccrualOrder{OrderNum: "123", Status: "PROCESSED", Accrual: 12.34}

	s.repo.EXPECT().
		UpdateOrderAndIncreaseBalance(gomock.Any(), gomock.AssignableToTypeOf(&orderrepository.UpdateOrderParams{})).
		DoAndReturn(func(ctx context.Context, arg *orderrepository.UpdateOrderParams) error {
			s.Require().Equal(int64(123), arg.OrderNum)
			s.Require().Equal(int32(order.StatusProcessed), arg.Status)
			s.Require().Equal(accrual.Accrual, service.NumericToFloat(arg.Accrual))
			return nil
		})

	s.NoError(s.svc.UpdateOrderStatus(context.Background(), accrual))
}

func (s *orderServiceSuite) TestUpdateOrderStatusBadOrderNum() {
	err := s.svc.UpdateOrderStatus(context.Background(), &responses.AccrualOrder{OrderNum: "abc", Status: "PROCESSED", Accrual: 1})
	s.Error(err)
}

func TestOrderServiceSuite(t *testing.T) {
	suite.Run(t, new(orderServiceSuite))
}
