package postgres

import (
	"context"
	"fmt"

	"github.com/Svirex/gofermart-loyality/internal/common"
	"github.com/Svirex/gofermart-loyality/internal/core/domain"
	"github.com/Svirex/gofermart-loyality/internal/core/ports"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BalanceRepository struct {
	db     *pgxpool.Pool
	logger common.Logger
}

func NewBalanceRepository(db *pgxpool.Pool, logger common.Logger) *BalanceRepository {
	return &BalanceRepository{
		db:     db,
		logger: logger,
	}
}

var _ ports.BalanceRepository = (*BalanceRepository)(nil)

func (repo *BalanceRepository) GetBalance(ctx context.Context, uid int64) (*domain.Balance, error) {
	data := &domain.Balance{}
	err := repo.db.QueryRow(ctx, "SELECT current, withdrawn FROM balance WHERE uid=$1", uid).Scan(&data.Current, &data.Withdrawn)
	if err != nil {
		repo.logger.Errorf("balance repo, get balance, select: %v", err)
		return nil, fmt.Errorf("balance repo, get balance, select: %w", err)
	}
	return data, nil
}
