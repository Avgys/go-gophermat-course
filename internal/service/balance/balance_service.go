package balance

import (
	"avgys-gophermat/internal/model/requests"
	"avgys-gophermat/internal/model/responses"
	"avgys-gophermat/internal/repository"
	"avgys-gophermat/internal/service"
	"avgys-gophermat/internal/service/auth"
	"avgys-gophermat/internal/service/validation"
	httphelper "avgys-gophermat/internal/shared/http"
	balancerepository "avgys-gophermat/sqlc/balance"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/samber/lo"
)

type BalanceService struct {
	repository BalanceRepository
}

func NewBalanceService(resository BalanceRepository) *BalanceService {
	return &BalanceService{resository}
}

func (b *BalanceService) GetBalanceByUserID(ctx context.Context, userClaims *auth.TokenClaims) (*responses.Balance, error) {

	userID := userClaims.UserID

	row, err := b.repository.GetBalance(ctx, userID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &responses.Balance{CurrentSum: 0, Withdrawn: 0}, nil
		}
		return nil, err
	}

	return &responses.Balance{CurrentSum: service.NumericToFloat(row.Amount), Withdrawn: service.NumericToFloat(row.Withdrawn)}, nil
}

func (b *BalanceService) GetWithdrawals(ctx context.Context, userClaims *auth.TokenClaims) ([]responses.WithdrawRs, error) {

	userID := userClaims.UserID

	rows, err := b.repository.GetWithdrawals(ctx, userID)

	if err != nil {
		return nil, err
	}

	withdrawals := lo.Map(rows, func(row balancerepository.GetWithdrawalsRow, _ int) responses.WithdrawRs {
		return responses.WithdrawRs{
			OrderNum:    row.OrderNum,
			Sum:         service.NumericToFloat(row.WithdrawAmount),
			ProcessedAt: row.CreatedAt.Time.Format(time.RFC3339),
		}
	})

	return withdrawals, nil
}

func (b *BalanceService) Withdraw(ctx context.Context, userClaims *auth.TokenClaims, withdraw *requests.WithdrawRq) (*responses.WithdrawDeltaRs, error) {

	if withdraw.Sum < 0 {
		return nil, httphelper.NewError("withdraw amount must be more than 0", http.StatusBadRequest)
	}

	if err := validation.LuhnNumVerify(withdraw.Order); err != nil {
		return nil, fmt.Errorf("%w inner %w", httphelper.NewError("order number is invalid", http.StatusUnprocessableEntity), err)
	}

	userID := userClaims.UserID

	n, err := strconv.ParseInt(withdraw.Order, 10, 64)

	if err != nil {
		return nil, err
	}

	row, err := b.repository.Withdraw(ctx, userID, withdraw.Sum, n)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &responses.WithdrawDeltaRs{Modified: false}, nil
		}

		if errors.Is(err, repository.ErrInsufficientBalance) {
			showErr := httphelper.NewError(err.Error(), http.StatusPaymentRequired)
			return &responses.WithdrawDeltaRs{Modified: false}, fmt.Errorf("%w: inner error %w", showErr, err)
		}

		return nil, err
	}

	return &responses.WithdrawDeltaRs{Modified: row.Modified, NewAmount: row.NewAmount, OldAmount: row.OldAmount}, nil
}
