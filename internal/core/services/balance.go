package services

import (
	"context"

	"github.com/Svirex/gofermart-loyality/internal/core/domain"
	"github.com/Svirex/gofermart-loyality/internal/core/ports"
)

type BalanceService struct {
	repository ports.BalanceRepository
}

func NewBalanceService(repository ports.BalanceRepository) *BalanceService {
	return &BalanceService{
		repository: repository,
	}
}

var _ ports.BalanceService = (*BalanceService)(nil)

func (service *BalanceService) GetBalance(ctx context.Context, uid int64) (*domain.Balance, error) {
	return service.repository.GetBalance(ctx, uid)
}
