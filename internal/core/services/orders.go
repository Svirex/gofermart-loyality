package services

import (
	"context"
	"fmt"
	"strconv"

	"github.com/Svirex/gofermart-loyality/internal/core/domain"
	"github.com/Svirex/gofermart-loyality/internal/core/ports"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderService struct {
	repo                ports.OrderRepository
	checkAccrualService *CheckAccrualService
}

func NewOrderService(dbpool *pgxpool.Pool, repo ports.OrderRepository) (*OrderService, error) {
	cas, err := NewCheckAccrualService(dbpool)
	if err != nil {
		return nil, fmt.Errorf("new order service, create check accrual service: %w", err)
	}
	cas.Start()
	return &OrderService{
		repo:                repo,
		checkAccrualService: cas,
	}, nil
}

var _ ports.OrdersService = (*OrderService)(nil)

func (service *OrderService) CreateOrder(ctx context.Context, uid int64, orderNum string) (ports.Status, error) {
	ok, err := checkLuhn(orderNum)
	if err != nil {
		return ports.Err, fmt.Errorf("order service, create order: %w", err)
	}
	if !ok {
		return ports.Err, ports.ErrInvalidOrderNum
	}
	userOrder, err := service.repo.CreateOrder(ctx, uid, orderNum)
	if err != nil {
		return ports.Err, fmt.Errorf("order service, create order, repo answer: %w", err)
	}
	if userOrder.New {
		service.checkAccrualService.Process(ctx, orderNum)
		return ports.Ok, nil
	} else {
		if userOrder.ID == uid {
			return ports.AlreadyAdded, nil
		}
		return ports.NotOwnOrder, nil
	}
}

func (service *OrderService) GetOrders(ctx context.Context, uid int64) ([]domain.Order, error) {
	return nil, nil
}

func checkLuhn(orderNum string) (bool, error) {
	sum := 0
	digitsQnt := len(orderNum)
	parity := digitsQnt % 2
	for i := 0; i < digitsQnt; i++ {
		digit, err := strconv.Atoi(string(orderNum[i]))
		if err != nil {
			return false, fmt.Errorf("checkLunh, atoi numbder: %w", err)
		}
		if i%2 == parity {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
	}
	return (sum % 10) == 0, nil
}
