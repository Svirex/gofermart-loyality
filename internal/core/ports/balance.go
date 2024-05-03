package ports

import (
	"context"

	"github.com/Svirex/gofermart-loyality/internal/core/domain"
)

type BalanceService interface {
	GetBalance(ctx context.Context, uid int64) (*domain.Balance, error)
}

type BalanceRepository interface {
	GetBalance(ctx context.Context, uid int64) (*domain.Balance, error)
}
