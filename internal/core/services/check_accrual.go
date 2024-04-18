package services

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type CheckAccrualService struct {
	dbpool *pgxpool.Pool
}

func NewCheckAccrualService(dbpool *pgxpool.Pool) (*CheckAccrualService, error) {
	return &CheckAccrualService{
		dbpool: dbpool,
	}, nil
}

func (service *CheckAccrualService) Start() {

}

func (service *CheckAccrualService) Process(ctx context.Context, orderNum string) {

}
