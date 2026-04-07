package endpoint_tests

import (
	"avgys-gophermat/internal/endpoints"
	"avgys-gophermat/internal/endpoints/tests/mocks"
	"avgys-gophermat/internal/middlewares"
	"avgys-gophermat/internal/service/auth"
	httphelper "avgys-gophermat/internal/shared/http"
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
)

type ordersSuite struct {
	suite.Suite
	ctrl      *gomock.Controller
	orderMock *mocks.MockOrderService
}

func (s *ordersSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.orderMock = mocks.NewMockOrderService(s.ctrl)
}

func (s *ordersSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *ordersSuite) newEndpoints() *endpoints.Endpoints {
	return endpoints.New(nil, s.orderMock, nil)
}

func (s *ordersSuite) requestWithClaims(body []byte) *http.Request {
	claims := auth.NewToken(1, "alice")
	ctx := claims.WithContext(context.Background())
	return httptest.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewReader(body)).WithContext(ctx)
}

func (s *ordersSuite) TestLoadOrderOK() {
	orderNum := "12345678903"
	s.orderMock.EXPECT().
		Store(gomock.Any(), gomock.AssignableToTypeOf(&auth.TokenClaims{}), orderNum).
		DoAndReturn(func(ctx context.Context, userClaims *auth.TokenClaims, num string) error {
			s.Equal(int64(1), userClaims.UserID)
			s.Equal(orderNum, num)
			return nil
		})

	req := s.requestWithClaims([]byte(orderNum))
	resp := httptest.NewRecorder()

	s.newEndpoints().LoadOrder(resp, req)

	s.Equal(http.StatusOK, resp.Code)
}

func (s *ordersSuite) TestLoadOrderBadBody() {
	req := s.requestWithClaims(nil)
	resp := httptest.NewRecorder()

	s.newEndpoints().LoadOrder(resp, req)

	s.Equal(http.StatusBadRequest, resp.Code)
}

func (s *ordersSuite) TestLoadOrderAccepted() {
	orderNum := "12345678903"
	s.orderMock.EXPECT().
		Store(gomock.Any(), gomock.AssignableToTypeOf(&auth.TokenClaims{}), orderNum).
		Return(httphelper.NewError("order is already in processing", http.StatusAccepted))

	req := s.requestWithClaims([]byte(orderNum))
	resp := httptest.NewRecorder()

	s.newEndpoints().LoadOrder(resp, req)

	s.Equal(http.StatusAccepted, resp.Code)
}

func (s *ordersSuite) TestLoadOrderConflict() {
	orderNum := "12345678903"
	s.orderMock.EXPECT().
		Store(gomock.Any(), gomock.AssignableToTypeOf(&auth.TokenClaims{}), orderNum).
		Return(httphelper.NewError("order is already registered by another user", http.StatusConflict))

	req := s.requestWithClaims([]byte(orderNum))
	resp := httptest.NewRecorder()

	s.newEndpoints().LoadOrder(resp, req)

	s.Equal(http.StatusConflict, resp.Code)
}

func (s *ordersSuite) TestLoadOrderInvalidNumber() {
	orderNum := "invalid"
	s.orderMock.EXPECT().
		Store(gomock.Any(), gomock.AssignableToTypeOf(&auth.TokenClaims{}), orderNum).
		Return(httphelper.NewError("order number is invalid", http.StatusUnprocessableEntity))

	req := s.requestWithClaims([]byte(orderNum))
	resp := httptest.NewRecorder()

	s.newEndpoints().LoadOrder(resp, req)

	s.Equal(http.StatusUnprocessableEntity, resp.Code)
}

func (s *ordersSuite) TestLoadOrderInternalError() {
	orderNum := "12345678903"
	s.orderMock.EXPECT().
		Store(gomock.Any(), gomock.AssignableToTypeOf(&auth.TokenClaims{}), orderNum).
		Return(errors.New("boom"))

	req := s.requestWithClaims([]byte(orderNum))
	resp := httptest.NewRecorder()

	s.newEndpoints().LoadOrder(resp, req)

	s.Equal(http.StatusInternalServerError, resp.Code)
}

func (s *ordersSuite) TestLoadOrderUnauthorized() {
	h := middlewares.RequireCookie(http.HandlerFunc(s.newEndpoints().LoadOrder))
	req := httptest.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewReader([]byte("123")))
	resp := httptest.NewRecorder()

	h.ServeHTTP(resp, req)

	s.Equal(http.StatusUnauthorized, resp.Code)
}

func TestOrdersSuite(t *testing.T) {
	suite.Run(t, new(ordersSuite))
}
