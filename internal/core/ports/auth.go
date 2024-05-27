package ports

import (
	"context"
	"errors"

	"github.com/Svirex/gofermart-loyality/internal/core/domain"
)

var ErrUserAlreadyExists = errors.New("user already exists")
var ErrEmptyLogin = errors.New("empty login")
var ErrEmptyPassword = errors.New("empty password")
var ErrLowPasswordStrength = errors.New("low password sthrength")
var ErrUserNotFound = errors.New("user not found")
var ErrPasswordTooShort = errors.New("the password is too short")
var ErrPasswordTooLong = errors.New("the password is too long")
var ErrInvalidPassword = errors.New("invalid password")

type AuthService interface {
	Register(ctx context.Context, login, password string) (string, error)
	Login(ctx context.Context, login, password string) (string, error)
	GetUserGromJWT(ctx context.Context, jwt string) (int64, error)
}

type AuthRepository interface {
	CreateUser(ctx context.Context, user *domain.User) (*domain.User, error)
	GetUserByLogin(ctx context.Context, login string) (*domain.User, error)
}
