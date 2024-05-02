package ports

import (
	"context"
	"time"
)

type WithdrawData struct {
	OrderNum    string    `json:"order"`
	Sum         float64   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}

type WithdrawService interface {
	Withdraw(ctx context.Context, uid int64, data *WithdrawData) error
}

type WithdrawRepository interface {
	Withdraw(ctx context.Context, uid int64, data *WithdrawData) error
}
