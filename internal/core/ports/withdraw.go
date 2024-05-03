package ports

import (
	"context"
	"errors"

	"github.com/Svirex/gofermart-loyality/internal/core/domain"
)

var ErrNotEnoughMoney = errors.New("not enough money")
var ErrDuplicateOrderNumber = errors.New("duplicate order number")
var ErrSumIsNegative = errors.New("sum is negative")

type WithdrawService interface {
	Withdraw(ctx context.Context, uid int64, data *domain.WithdrawData) error
	GetWithdrawals(ctx context.Context, uid int64) ([]*domain.WithdrawData, error)
}

type WithdrawRepository interface {
	Withdraw(ctx context.Context, uid int64, data *domain.WithdrawData) error
	GetWithdrawals(ctx context.Context, uid int64) ([]*domain.WithdrawData, error)
}
