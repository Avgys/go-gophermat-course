package processor

import (
	"avgys-gophermat/internal/model/responses"
	accrualmocks "avgys-gophermat/internal/service/accrualclient/tests/mocks"
	orderrepository "avgys-gophermat/sqlc/order"
	"context"
	"testing"
	"time"

	"avgys-gophermat/internal/processor/tests/mocks"

	"github.com/golang/mock/gomock"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/suite"
)

type accrualProcessorSuite struct {
	suite.Suite
	ctrl        *gomock.Controller
	orderMock   *mocks.MockOrderService
	accrualMock *accrualmocks.MockAccrualClient
	pool        *Pool
	logger      *zerolog.Logger
}

func (s *accrualProcessorSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.orderMock = mocks.NewMockOrderService(s.ctrl)
	s.accrualMock = accrualmocks.NewMockAccrualClient(s.ctrl)
	s.pool = NewPool(1, 1, 1)
	logger := zerolog.Nop()
	s.logger = &logger
}

func (s *accrualProcessorSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *accrualProcessorSuite) newProcessor() *AcrrualProcessor {
	p := &AcrrualProcessor{
		workersLimit:   1,
		pool:           s.pool,
		orderService:   s.orderMock,
		accrualService: s.accrualMock,
		logger:         s.logger,
	}
	p.accrualPoolSleepTime.Store(0)
	return p
}

func (s *accrualProcessorSuite) TestStartPollingSuccess() {
	p := s.newProcessor()
	done, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		_ = p.pool.Run(done)
	}()

	orderCh := make(chan int64, 1)
	resultCh := p.startPolling(done, orderCh)

	orderNum := int64(123)
	orderNumStr := "123"
	expected := &responses.AccrualOrder{OrderNum: orderNumStr, Status: "PROCESSED", Accrual: 500}

	s.accrualMock.EXPECT().
		Send(gomock.Any(), orderNumStr).
		Return(expected, nil)

	orderCh <- orderNum

	select {
	case result := <-resultCh:
		s.Equal(orderNumStr, result.orderNum)
		s.Equal(expected, result.response)
		s.NoError(result.err)
	case <-time.After(2 * time.Second):
		s.Fail("timeout waiting for accrual result")
	}
}

func (s *accrualProcessorSuite) TestInitPollingSuccess() {
	p := s.newProcessor()
	done, cancel := context.WithCancel(context.Background())
	defer cancel()

	updatedCh := make(chan *responses.AccrualOrder, 1)

	s.orderMock.EXPECT().
		GetOrderUnprocessedOrders(gomock.Any(), 1).
		Return([]orderrepository.Order{{OrderNum: 123}}, nil).
		MinTimes(1)

	expected := &responses.AccrualOrder{OrderNum: "123", Status: "PROCESSED", Accrual: 500}
	s.accrualMock.EXPECT().
		Send(gomock.Any(), "123").
		Return(expected, nil).
		MinTimes(1)

	s.orderMock.EXPECT().
		UpdateOrderStatus(gomock.Any(), expected).
		DoAndReturn(func(ctx context.Context, accrualOrder *responses.AccrualOrder) error {
			updatedCh <- accrualOrder
			cancel()
			return nil
		})

	p.InitPolling(done)

	select {
	case got := <-updatedCh:
		s.Equal(expected, got)
	case <-time.After(50 * time.Second):
		s.Fail("timeout waiting for UpdateOrderStatus")
	}
}

func TestAcrrualProcessorSuite(t *testing.T) {
	suite.Run(t, new(accrualProcessorSuite))
}
