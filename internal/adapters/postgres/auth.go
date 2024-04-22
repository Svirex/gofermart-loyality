package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/Svirex/gofermart-loyality/internal/core/domain"
	"github.com/Svirex/gofermart-loyality/internal/core/ports"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthRepository struct {
	db *pgxpool.Pool
}

func NewAuthRepository(db *pgxpool.Pool) *AuthRepository {
	return &AuthRepository{
		db: db,
	}
}

var _ ports.AuthRepository = (*AuthRepository)(nil)

func (r *AuthRepository) CreateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	var id int64
	err := r.db.QueryRow(ctx, `INSERT INTO users (login, hash) VALUES ($1, $2) RETURNING id;`, user.Login, user.Hash).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			// TODO переделать на просто возврат ошибки. Определение, что делать с этой ошибкой - задача сервиса
			return nil, fmt.Errorf("%w: auth repository create user, user already exists: %v", ports.ErrUserAlreadyExists, err)
		}
		return nil, fmt.Errorf("auth repository create user: %w", err)
	}
	user.ID = id
	return user, nil
}

func (r *AuthRepository) GetUserByLogin(ctx context.Context, login string) (*domain.User, error) {
	user := &domain.User{}
	err := r.db.QueryRow(ctx, `SELECT id, login, hash FROM users WHERE login=$1`, login).Scan(&user.ID, &user.Login, &user.Hash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// TODO переделать на просто возврат ошибки. Определение, что делать с этой ошибкой - задача сервиса
			return nil, fmt.Errorf("%w: auth repository, get user by login, user not found: %v", ports.ErrUserNotFound, err)
		}
		return nil, fmt.Errorf("auth repository, get user by login: %w", err)
	}
	return user, nil
}
