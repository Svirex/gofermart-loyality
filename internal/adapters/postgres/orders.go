package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/Svirex/gofermart-loyality/internal/common"
	"github.com/Svirex/gofermart-loyality/internal/core/domain"
	"github.com/Svirex/gofermart-loyality/internal/core/ports"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrdersRepository struct {
	db     *pgxpool.Pool
	logger common.Logger
}

func NewOrdersRepository(db *pgxpool.Pool, logger common.Logger) *OrdersRepository {
	return &OrdersRepository{
		db:     db,
		logger: logger,
	}
}

var _ ports.OrdersRepository = (*OrdersRepository)(nil)

func (repo *OrdersRepository) CreateOrder(ctx context.Context, uid int64, orderNum string) (*ports.UserOrder, error) {
	_, err := repo.db.Exec(ctx, "INSERT INTO orders (uid, order_num) VALUES ($1, $2);", uid, orderNum)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			var uid int64
			err := repo.db.QueryRow(ctx, "SELECT uid FROM orders WHERE order_num=$1;", orderNum).Scan(&uid)
			if err != nil {
				repo.logger.Errorln("orders repo, create order, select uid: ", err)
				return nil, fmt.Errorf("orders repo, create order, select uid: %w", err)
			}
			return &ports.UserOrder{
				ID:  uid,
				New: false,
			}, nil
		}
		repo.logger.Errorln("orders repo, create order, insert: ", err)
		return nil, fmt.Errorf("orders repo, create order, insert: %w", err)
	}
	return &ports.UserOrder{
		ID:  uid,
		New: true,
	}, nil
}

func (repo *OrdersRepository) GetOrders(ctx context.Context, uid int64) ([]domain.Order, error) {
	rows, _ := repo.db.Query(ctx, "SELECT order_num, status, accrual, uploaded_at FROM orders WHERE uid=$1 ORDER BY uploaded_at DESC;", uid)
	if err := rows.Err(); err != nil {
		repo.logger.Errorln("orders repo, get orders, select: ", err)
		return nil, fmt.Errorf("orders repo, get orders, select: %w", err)
	}
	orders := make([]domain.Order, 0)
	for rows.Next() {
		if err := rows.Err(); err != nil {
			repo.logger.Errorln("orders repo, get orders, next row: ", err)
			return nil, fmt.Errorf("orders repo, get orders, next row: %w", err)
		}
		order := domain.Order{}
		if err := rows.Scan(&order.Number, &order.Status, &order.Accrual, &order.UploadedAt); err != nil {
			repo.logger.Errorln("orders repo, get orders, next row, scan: ", err)
			return nil, fmt.Errorf("orders repo, get orders, next row, scan: %w", err)
		}
		orders = append(orders, order)

	}
	return orders, nil
}
