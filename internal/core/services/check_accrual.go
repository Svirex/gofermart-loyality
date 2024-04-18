package services

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type CheckAccrualService struct {
	dbpool      *pgxpool.Pool
	stopCh      chan struct{}
	orderNumsCh chan string
}

func NewCheckAccrualService(dbpool *pgxpool.Pool) (*CheckAccrualService, error) {
	return &CheckAccrualService{
		dbpool:      dbpool,
		stopCh:      make(chan struct{}),
		orderNumsCh: make(chan string, 10),
	}, nil
}

func (service *CheckAccrualService) Start() {

}

func (service *CheckAccrualService) Process(ctx context.Context, orderNum string) {
	go service.generator(orderNum)
}

func (service *CheckAccrualService) generator(orderNum string) {
	for {
		select {
		case <-service.stopCh:
			return
		default:
			service.orderNumsCh <- orderNum
		}
	}
}

func (service *CheckAccrualService) queueUpdater() {

}

func (service *CheckAccrualService) Shutdown() {
	close(service.stopCh)
}
