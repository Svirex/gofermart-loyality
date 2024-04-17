package ports

import "context"

type OrdersService interface {
	CreateOrder(ctx context.Context, orderNum string) (bool, error)
	GetOrders(ctx context.Context)
}
