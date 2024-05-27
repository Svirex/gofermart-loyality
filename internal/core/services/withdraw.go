package services

import (
	"context"
	"fmt"

	"github.com/Svirex/gofermart-loyality/internal/core/domain"
	"github.com/Svirex/gofermart-loyality/internal/core/ports"
)

type WithdrawService struct {
	repository ports.WithdrawRepository
}

func NewWithdrawService(repository ports.WithdrawRepository) *WithdrawService {
	return &WithdrawService{
		repository: repository,
	}
}

var _ ports.WithdrawService = (*WithdrawService)(nil)

func (service *WithdrawService) Withdraw(ctx context.Context, uid int64, data *domain.WithdrawData) error {
	ok, err := checkLuhn(data.OrderNum)
	if err != nil {
		return fmt.Errorf("%w: withdraw service, withdraw, check luhn: %v", ports.ErrInvalidOrderNum, err)
	}
	if !ok {
		return ports.ErrInvalidOrderNum
	}
	return service.repository.Withdraw(ctx, uid, data)
}

func (service *WithdrawService) GetWithdrawals(ctx context.Context, uid int64) ([]*domain.WithdrawData, error) {
	return service.repository.GetWithdrawals(ctx, uid)
}
