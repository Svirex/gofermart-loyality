package services

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/Svirex/gofermart-loyality/internal/common"
	"github.com/Svirex/gofermart-loyality/internal/core/domain"
	"github.com/Svirex/gofermart-loyality/internal/core/ports"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderService struct {
	repo                ports.OrdersRepository
	checkAccrualService *CheckAccrualService
	logger              common.Logger
}

func NewOrderService(dbpool *pgxpool.Pool, repo ports.OrdersRepository, logger common.Logger, queueSize int, accrualAddr string, pauseBetweenRequests time.Duration, maxRunnedGenerators int32, dbLoaderPause time.Duration) (*OrderService, error) {
	cas, err := NewCheckAccrualService(dbpool, logger, queueSize, accrualAddr, pauseBetweenRequests, maxRunnedGenerators, dbLoaderPause)
	if err != nil {
		logger.Errorf("new order service, create check accrual service: %v", err)
		return nil, fmt.Errorf("new order service, create check accrual service: %w", err)
	}
	cas.Start()
	return &OrderService{
		repo:                repo,
		checkAccrualService: cas,
		logger:              logger,
	}, nil
}

var _ ports.OrdersService = (*OrderService)(nil)

func (service *OrderService) CreateOrder(ctx context.Context, uid int64, orderNum string) (ports.Status, error) {
	ok, err := checkLuhn(orderNum)
	if err != nil {
		service.logger.Errorf("order service, create order: %v", err)
		return ports.Err, fmt.Errorf("%w: order service, create order: %v", ports.ErrInvalidOrderNum, err)
	}
	if !ok {
		return ports.Err, ports.ErrInvalidOrderNum
	}
	userOrder, err := service.repo.CreateOrder(ctx, uid, orderNum)
	if err != nil {
		service.logger.Errorf("order service, create order, repo answer: %v", err)
		return ports.Err, fmt.Errorf("order service, create order, repo answer: %w", err)
	}
	if userOrder.New {
		service.logger.Debugln("SERVICE CREATE ORDER WITH NUM", orderNum)
		service.checkAccrualService.Process(orderNum)
		return ports.Ok, nil
	} else {
		if userOrder.ID == uid {
			return ports.AlreadyAdded, nil
		}
		return ports.NotOwnOrder, nil
	}
}

func (service *OrderService) GetOrders(ctx context.Context, uid int64) ([]domain.Order, error) {
	return service.repo.GetOrders(ctx, uid)
}

func (service *OrderService) Shutdown() {
	service.checkAccrualService.Shutdown()
}

func checkLuhn(orderNum string) (bool, error) {
	sum := 0
	digitsQnt := len(orderNum)
	parity := digitsQnt % 2
	for i := 0; i < digitsQnt; i++ {
		digit, err := strconv.Atoi(string(orderNum[i]))
		if err != nil {
			return false, fmt.Errorf("checkLunh, atoi number: %w", err)
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
