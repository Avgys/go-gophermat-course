package endpoint_tests

import (
	"avgys-gophermat/internal/endpoints"
	"avgys-gophermat/internal/endpoints/tests/mocks"
	"avgys-gophermat/internal/middlewares"
	"avgys-gophermat/internal/model/requests"
	"avgys-gophermat/internal/model/responses"
	"avgys-gophermat/internal/service/auth"
	httphelper "avgys-gophermat/internal/shared/http"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
)

type balanceSuite struct {
	suite.Suite
	ctrl        *gomock.Controller
	balanceMock *mocks.MockBalanceService
}

func (s *balanceSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.balanceMock = mocks.NewMockBalanceService(s.ctrl)
}

func (s *balanceSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *balanceSuite) newEndpoints() *endpoints.Endpoints {
	return endpoints.New(nil, nil, s.balanceMock)
}

func (s *balanceSuite) requestWithClaimsGet() *http.Request {
	claims := auth.NewToken(1, "alice")
	ctx := claims.WithContext(context.Background())
	return httptest.NewRequest(http.MethodGet, "/api/user/balance", nil).WithContext(ctx)
}

func (s *balanceSuite) requestWithClaimsWithdraw(body []byte) *http.Request {
	claims := auth.NewToken(1, "alice")
	ctx := claims.WithContext(context.Background())
	return httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", bytes.NewReader(body)).WithContext(ctx)
}

func (s *balanceSuite) requestWithClaimsWithdrawals() *http.Request {
	claims := auth.NewToken(1, "alice")
	ctx := claims.WithContext(context.Background())
	return httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil).WithContext(ctx)
}

func (s *balanceSuite) TestGetBalanceOK() {
	expected := &responses.Balance{CurrentSum: 500.5, Withdrawn: 42}

	s.balanceMock.EXPECT().
		GetBalanceByUserID(gomock.Any(), gomock.AssignableToTypeOf(&auth.TokenClaims{})).
		Return(expected, nil)

	req := s.requestWithClaimsGet()
	resp := httptest.NewRecorder()

	s.newEndpoints().GetBalanceByUserID(resp, req)

	s.Equal(http.StatusOK, resp.Code)
	s.Equal("application/json", resp.Header().Get("Content-type"))

	var got responses.Balance
	s.NoError(json.Unmarshal(resp.Body.Bytes(), &got))
	s.Equal(*expected, got)
}

func (s *balanceSuite) TestGetBalanceInternalError() {
	s.balanceMock.EXPECT().
		GetBalanceByUserID(gomock.Any(), gomock.AssignableToTypeOf(&auth.TokenClaims{})).
		Return(nil, errors.New("boom"))

	req := s.requestWithClaimsGet()
	resp := httptest.NewRecorder()

	s.newEndpoints().GetBalanceByUserID(resp, req)

	s.Equal(http.StatusInternalServerError, resp.Code)
}

func (s *balanceSuite) TestGetBalanceUnauthorized() {
	h := middlewares.RequireCookie(http.HandlerFunc(s.newEndpoints().GetBalanceByUserID))
	req := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)
	resp := httptest.NewRecorder()

	h.ServeHTTP(resp, req)

	s.Equal(http.StatusUnauthorized, resp.Code)
}

func (s *balanceSuite) TestWithdrawOK() {
	body := []byte(`{"order":"2377225624","sum":751}`)
	expected := &responses.WithdrawDeltaRs{Modified: true, NewAmount: 249, OldAmount: 1000}

	s.balanceMock.EXPECT().
		Withdraw(gomock.Any(), gomock.AssignableToTypeOf(&auth.TokenClaims{}), gomock.AssignableToTypeOf(&requests.WithdrawRq{})).
		DoAndReturn(func(ctx context.Context, userClaims *auth.TokenClaims, withdraw *requests.WithdrawRq) (*responses.WithdrawDeltaRs, error) {
			s.Equal("2377225624", withdraw.Order)
			s.Equal(float32(751), withdraw.Sum)
			return expected, nil
		})

	req := s.requestWithClaimsWithdraw(body)
	resp := httptest.NewRecorder()

	s.newEndpoints().Withdraw(resp, req)

	s.Equal(http.StatusOK, resp.Code)
	s.Equal("application/json", resp.Header().Get("Content-type"))

	var got responses.WithdrawDeltaRs
	s.NoError(json.Unmarshal(resp.Body.Bytes(), &got))
	s.Equal(*expected, got)
}

func (s *balanceSuite) TestWithdrawInsufficientFunds() {
	body := []byte(`{"order":"2377225624","sum":751}`)

	s.balanceMock.EXPECT().
		Withdraw(gomock.Any(), gomock.AssignableToTypeOf(&auth.TokenClaims{}), gomock.AssignableToTypeOf(&requests.WithdrawRq{})).
		Return(nil, httphelper.NewError("insufficient balance", http.StatusPaymentRequired))

	req := s.requestWithClaimsWithdraw(body)
	resp := httptest.NewRecorder()

	s.newEndpoints().Withdraw(resp, req)

	s.Equal(http.StatusPaymentRequired, resp.Code)
}

func (s *balanceSuite) TestWithdrawInvalidOrder() {
	body := []byte(`{"order":"invalid","sum":751}`)

	s.balanceMock.EXPECT().
		Withdraw(gomock.Any(), gomock.AssignableToTypeOf(&auth.TokenClaims{}), gomock.AssignableToTypeOf(&requests.WithdrawRq{})).
		Return(nil, httphelper.NewError("order number is invalid", http.StatusUnprocessableEntity))

	req := s.requestWithClaimsWithdraw(body)
	resp := httptest.NewRecorder()

	s.newEndpoints().Withdraw(resp, req)

	s.Equal(http.StatusUnprocessableEntity, resp.Code)
}

func (s *balanceSuite) TestWithdrawInternalError() {
	body := []byte(`{"order":"2377225624","sum":751}`)

	s.balanceMock.EXPECT().
		Withdraw(gomock.Any(), gomock.AssignableToTypeOf(&auth.TokenClaims{}), gomock.AssignableToTypeOf(&requests.WithdrawRq{})).
		Return(nil, errors.New("boom"))

	req := s.requestWithClaimsWithdraw(body)
	resp := httptest.NewRecorder()

	s.newEndpoints().Withdraw(resp, req)

	s.Equal(http.StatusInternalServerError, resp.Code)
}

func (s *balanceSuite) TestWithdrawUnauthorized() {
	h := middlewares.RequireCookie(http.HandlerFunc(s.newEndpoints().Withdraw))
	req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", bytes.NewReader([]byte("{}")))
	resp := httptest.NewRecorder()

	h.ServeHTTP(resp, req)

	s.Equal(http.StatusUnauthorized, resp.Code)
}

func (s *balanceSuite) TestGetWithdrawalsOK() {
	withdrawals := []responses.WithdrawRs{
		{OrderNum: int64(2377225624), Sum: 500, ProcessedAt: "2020-12-09T16:09:57+03:00"},
	}

	s.balanceMock.EXPECT().
		GetWithdrawals(gomock.Any(), gomock.AssignableToTypeOf(&auth.TokenClaims{})).
		Return(withdrawals, nil)

	req := s.requestWithClaimsWithdrawals()
	resp := httptest.NewRecorder()

	s.newEndpoints().GetWithdrawals(resp, req)

	s.Equal(http.StatusOK, resp.Code)
	s.Equal("application/json", resp.Header().Get("Content-type"))

	var got []responses.WithdrawRs
	s.NoError(json.Unmarshal(resp.Body.Bytes(), &got))
	s.Equal(withdrawals, got)
}

func (s *balanceSuite) TestGetWithdrawalsNoContent() {
	s.balanceMock.EXPECT().
		GetWithdrawals(gomock.Any(), gomock.AssignableToTypeOf(&auth.TokenClaims{})).
		Return([]responses.WithdrawRs{}, nil)

	req := s.requestWithClaimsWithdrawals()
	resp := httptest.NewRecorder()

	s.newEndpoints().GetWithdrawals(resp, req)

	s.Equal(http.StatusNoContent, resp.Code)
	s.Empty(resp.Body.Bytes())
}

func (s *balanceSuite) TestGetWithdrawalsInternalError() {
	s.balanceMock.EXPECT().
		GetWithdrawals(gomock.Any(), gomock.AssignableToTypeOf(&auth.TokenClaims{})).
		Return(nil, errors.New("boom"))

	req := s.requestWithClaimsWithdrawals()
	resp := httptest.NewRecorder()

	s.newEndpoints().GetWithdrawals(resp, req)

	s.Equal(http.StatusInternalServerError, resp.Code)
}

func (s *balanceSuite) TestGetWithdrawalsUnauthorized() {
	h := middlewares.RequireCookie(http.HandlerFunc(s.newEndpoints().GetWithdrawals))
	req := httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil)
	resp := httptest.NewRecorder()

	h.ServeHTTP(resp, req)

	s.Equal(http.StatusUnauthorized, resp.Code)
}

func TestBalanceSuite(t *testing.T) {
	suite.Run(t, new(balanceSuite))
}
