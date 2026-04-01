package balance

import (
	"avgys-gophermat/internal/model/response"
	"avgys-gophermat/internal/repository"
	"avgys-gophermat/internal/service"
	"avgys-gophermat/internal/service/auth"
	"context"
	"database/sql"
	"errors"
)

type BalanceService struct {
	repository *repository.BalanceRepository
}

func NewBalanceService(resository *repository.BalanceRepository) *BalanceService {
	return &BalanceService{resository}
}

func (b *BalanceService) GetBalanceByUserID(ctx context.Context, userClaims *auth.TokenClaims) (*response.Balance, error) {

	userID := userClaims.UserID

	row, err := b.repository.GetBalance(ctx, userID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &response.Balance{CurrentSum: "0", Withdrawn: "0"}, nil
		}
		return nil, err
	}

	return &response.Balance{CurrentSum: service.NumericToStr(row.Balance), Withdrawn: service.NumericToStr(row.Withdrawn)}, nil
}
