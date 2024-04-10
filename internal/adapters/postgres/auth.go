package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/Svirex/gofermart-loyality/internal/core/domain"
	"github.com/Svirex/gofermart-loyality/internal/core/ports"
	"github.com/jackc/pgerrcode"
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
	err := r.db.QueryRow(ctx, `INSERT INTO users (login, hash) VALUE ($1, $2) 
	ON CONFLICT (login) DO UPDATE SET login = excluded.login
	RETURNING id;`, user.Login, user.Hash).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return nil, fmt.Errorf("auth repository create user, user already exists: %w", ports.ErrUserAlreadyExists)
		}
		return nil, fmt.Errorf("auth repository create user: %w", err)
	}
	user.ID = id
	return user, nil
}

func (r *AuthRepository) GetUserByLogin(ctx context.Context, login string) (*domain.User, error) {

}
