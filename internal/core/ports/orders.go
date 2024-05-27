package ports

import (
	"context"
	"errors"

	"github.com/Svirex/gofermart-loyality/internal/core/domain"
)

type Status int

const (
	AlreadyAdded Status = iota
	Ok
	NotOwnOrder
	Err
)

type UserOrder struct {
	ID  int64
	New bool
}

var ErrInvalidOrderNum = errors.New("invalid orders num")
var ErrInternalError = errors.New("internal error")

type OrdersService interface {
	CreateOrder(ctx context.Context, uid int64, orderNum string) (Status, error)
	GetOrders(ctx context.Context, uid int64) ([]domain.Order, error)
}

type OrdersRepository interface {
	CreateOrder(ctx context.Context, uid int64, orderNum string) (*UserOrder, error)
	GetOrders(ctx context.Context, uid int64) ([]domain.Order, error)
}
