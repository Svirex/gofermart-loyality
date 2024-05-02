package ports

import "context"

type Balance struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

type BalanceService interface {
	GetBalance(ctx context.Context, uid int64) (*Balance, error)
}

type BalanceRepository interface {
	GetBalance(ctx context.Context, uid int64) (*Balance, error)
}
