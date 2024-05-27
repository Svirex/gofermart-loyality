package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/Svirex/gofermart-loyality/internal/common"
	"github.com/Svirex/gofermart-loyality/internal/core/domain"
	"github.com/Svirex/gofermart-loyality/internal/core/ports"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WithdrawRepository struct {
	db     *pgxpool.Pool
	logger common.Logger
}

func NewWithdrawRepository(db *pgxpool.Pool, logger common.Logger) *WithdrawRepository {
	return &WithdrawRepository{
		db:     db,
		logger: logger,
	}
}

var _ ports.WithdrawRepository = (*WithdrawRepository)(nil)

func (repo *WithdrawRepository) Withdraw(ctx context.Context, uid int64, data *domain.WithdrawData) error {
	trx, err := repo.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("withdraw repo, Withdraw, start trx: %w", err)
	}
	defer trx.Rollback(ctx)
	_, err = trx.Exec(ctx, "UPDATE balance SET current=current-$1, withdrawn=withdrawn+$1 WHERE uid=$2;", data.Sum, uid)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.CheckViolation {
			return fmt.Errorf("%w: withdraw repo, Withdraw, update balance, not enough: %v", ports.ErrNotEnoughMoney, err)
		}
		return fmt.Errorf("withdraw repo, Withdraw, update balance: %w", err)
	}
	_, err = trx.Exec(ctx, "INSERT INTO withdraws (uid, order_num, sum) VALUES ($1, $2, $3);", uid, data.OrderNum, data.Sum)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return fmt.Errorf("%w: withdraw repo, Withdraw, insert withdraw record, duplicate order number: %v", ports.ErrDuplicateOrderNumber, err)
		}
		return fmt.Errorf("withdraw repo, Withdraw, insert withdraw record: %w", err)
	}
	err = trx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("withdraw repo, Withdraw, commit: %w", err)
	}
	return nil
}

func (repo *WithdrawRepository) GetWithdrawals(ctx context.Context, uid int64) ([]*domain.WithdrawData, error) {
	rows, _ := repo.db.Query(ctx, "SELECT order_num, sum, processed_at FROM withdraws WHERE uid=$1 ORDER BY processed_at DESC;", uid)
	if err := rows.Err(); err != nil {
		repo.logger.Errorln("withdraw repo, get withdrawals, select: %v", err)
		return nil, fmt.Errorf("withdraw repo, get withdrawals, select: %w", err)
	}
	data := make([]*domain.WithdrawData, 0)
	for rows.Next() {
		if err := rows.Err(); err != nil {
			repo.logger.Errorln("withdraw repo, get withdrawals, next row: %v", err)
			return nil, fmt.Errorf("withdraw repo, get withdrawals, next row: %w", err)
		}
		order := &domain.WithdrawData{}
		if err := rows.Scan(&order.OrderNum, &order.Sum, &order.ProcessedAt); err != nil {
			repo.logger.Errorln("withdraw repo, get withdrawals, scan: %v", err)
			return nil, fmt.Errorf("withdraw repo, get withdrawals, scan: %w", err)
		}
		data = append(data, order)

	}
	return data, nil
}
