package balance

import (
	"avgys-gophermat/internal/model/requests"
	"avgys-gophermat/internal/model/responses"
	"avgys-gophermat/internal/repository"
	"avgys-gophermat/internal/service"
	"avgys-gophermat/internal/service/auth"
	httphelper "avgys-gophermat/internal/shared/http"
	balancerepository "avgys-gophermat/sqlc/balance"
	"context"
	"database/sql"
	"errors"
	"net/http"
	"testing"
	"time"

	"avgys-gophermat/internal/service/balance/tests/mocks"

	"github.com/golang/mock/gomock"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/suite"
)

type balanceServiceSuite struct {
	suite.Suite
	ctrl *gomock.Controller
	repo *mocks.MockBalanceRepository
	svc  *BalanceService
}

func (s *balanceServiceSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.repo = mocks.NewMockBalanceRepository(s.ctrl)
	s.svc = NewBalanceService(s.repo)
}

func (s *balanceServiceSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *balanceServiceSuite) TestGetBalanceOK() {
	claims := auth.NewToken(7, "alice")
	var amount pgtype.Numeric
	_ = amount.Scan("500.5")
	var withdrawn pgtype.Numeric
	_ = withdrawn.Scan("42")

	s.repo.EXPECT().
		GetBalance(gomock.Any(), claims.UserID).
		Return(balancerepository.GetBalanceRow{Amount: amount, Withdrawn: withdrawn, UserID: claims.UserID}, nil)

	got, err := s.svc.GetBalanceByUserID(context.Background(), claims)
	s.NoError(err)
	s.Equal(&responses.Balance{CurrentSum: service.NumericToFloat(amount), Withdrawn: service.NumericToFloat(withdrawn)}, got)
}

func (s *balanceServiceSuite) TestGetBalanceNoRows() {
	claims := auth.NewToken(7, "alice")

	s.repo.EXPECT().
		GetBalance(gomock.Any(), claims.UserID).
		Return(balancerepository.GetBalanceRow{}, sql.ErrNoRows)

	got, err := s.svc.GetBalanceByUserID(context.Background(), claims)
	s.NoError(err)
	s.Equal(&responses.Balance{CurrentSum: 0, Withdrawn: 0}, got)
}

func (s *balanceServiceSuite) TestGetBalanceError() {
	claims := auth.NewToken(7, "alice")
	repoErr := errors.New("db error")

	s.repo.EXPECT().
		GetBalance(gomock.Any(), claims.UserID).
		Return(balancerepository.GetBalanceRow{}, repoErr)

	_, err := s.svc.GetBalanceByUserID(context.Background(), claims)
	s.Error(err)
	s.True(errors.Is(err, repoErr))
}

func (s *balanceServiceSuite) TestGetWithdrawalsOK() {
	claims := auth.NewToken(7, "alice")
	created := time.Date(2020, 12, 9, 16, 9, 57, 0, time.UTC)

	var amount pgtype.Numeric
	_ = amount.Scan("500")

	s.repo.EXPECT().
		GetWithdrawals(gomock.Any(), claims.UserID).
		Return([]balancerepository.GetWithdrawalsRow{{
			OrderNum:       2377225624,
			WithdrawAmount: amount,
			UserID:         claims.UserID,
			CreatedAt:      pgtype.Timestamp{Time: created, Valid: true},
		}}, nil)

	got, err := s.svc.GetWithdrawals(context.Background(), claims)
	s.NoError(err)
	s.Require().Len(got, 1)
	s.Equal(responses.WithdrawRs{OrderNum: 2377225624, Sum: service.NumericToFloat(amount), ProcessedAt: created.Format(time.RFC3339)}, got[0])
}

func (s *balanceServiceSuite) TestGetWithdrawalsError() {
	claims := auth.NewToken(7, "alice")
	repoErr := errors.New("db error")

	s.repo.EXPECT().
		GetWithdrawals(gomock.Any(), claims.UserID).
		Return(nil, repoErr)

	_, err := s.svc.GetWithdrawals(context.Background(), claims)
	s.Error(err)
	s.True(errors.Is(err, repoErr))
}

func (s *balanceServiceSuite) TestWithdrawNegativeSum() {
	claims := auth.NewToken(7, "alice")
	_, err := s.svc.Withdraw(context.Background(), claims, &requests.WithdrawRq{Order: "2377225624", Sum: -1})
	var httpErr *httphelper.ShowHTTPError
	s.Require().True(errors.As(err, &httpErr))
	s.Equal(http.StatusBadRequest, httpErr.StatusCode)
}

func (s *balanceServiceSuite) TestWithdrawInvalidOrder() {
	claims := auth.NewToken(7, "alice")
	_, err := s.svc.Withdraw(context.Background(), claims, &requests.WithdrawRq{Order: "invalid", Sum: 10})
	var httpErr *httphelper.ShowHTTPError
	s.Require().True(errors.As(err, &httpErr))
	s.Equal(http.StatusUnprocessableEntity, httpErr.StatusCode)
}

func (s *balanceServiceSuite) TestWithdrawNoRows() {
	claims := auth.NewToken(7, "alice")
	withdraw := &requests.WithdrawRq{Order: "79927398713", Sum: 10}

	s.repo.EXPECT().
		Withdraw(gomock.Any(), claims.UserID, withdraw.Sum, int64(79927398713)).
		Return(nil, sql.ErrNoRows)

	got, err := s.svc.Withdraw(context.Background(), claims, withdraw)
	s.NoError(err)
	s.Equal(&responses.WithdrawDeltaRs{Modified: false}, got)
}

func (s *balanceServiceSuite) TestWithdrawInsufficientBalance() {
	claims := auth.NewToken(7, "alice")
	withdraw := &requests.WithdrawRq{Order: "79927398713", Sum: 10}

	s.repo.EXPECT().
		Withdraw(gomock.Any(), claims.UserID, withdraw.Sum, int64(79927398713)).
		Return(nil, repository.ErrInsufficientBalance)

	got, err := s.svc.Withdraw(context.Background(), claims, withdraw)
	var httpErr *httphelper.ShowHTTPError
	s.Require().True(errors.As(err, &httpErr))
	s.Equal(http.StatusPaymentRequired, httpErr.StatusCode)
	s.Equal(&responses.WithdrawDeltaRs{Modified: false}, got)
}

func (s *balanceServiceSuite) TestWithdrawRepoError() {
	claims := auth.NewToken(7, "alice")
	withdraw := &requests.WithdrawRq{Order: "79927398713", Sum: 10}
	repoErr := errors.New("db error")

	s.repo.EXPECT().
		Withdraw(gomock.Any(), claims.UserID, withdraw.Sum, int64(79927398713)).
		Return(nil, repoErr)

	_, err := s.svc.Withdraw(context.Background(), claims, withdraw)
	s.Error(err)
	s.True(errors.Is(err, repoErr))
}

func (s *balanceServiceSuite) TestWithdrawOK() {
	claims := auth.NewToken(7, "alice")
	withdraw := &requests.WithdrawRq{Order: "79927398713", Sum: 10}

	s.repo.EXPECT().
		Withdraw(gomock.Any(), claims.UserID, withdraw.Sum, int64(79927398713)).
		Return(&repository.TryAddDeltaRow{Modified: true, NewAmount: 990, OldAmount: 1000}, nil)

	got, err := s.svc.Withdraw(context.Background(), claims, withdraw)
	s.NoError(err)
	s.Equal(&responses.WithdrawDeltaRs{Modified: true, NewAmount: 990, OldAmount: 1000}, got)
}

func TestBalanceServiceSuite(t *testing.T) {
	suite.Run(t, new(balanceServiceSuite))
}
