package ports

import (
	"context"
	"errors"

	"github.com/Svirex/gofermart-loyality/internal/core/domain"
)

var ErrUserAlreadyExists = errors.New("user already exists")
var ErrEmptyPassword = errors.New("empty password")
var ErrEmptyLogin = errors.New("empty login")
var ErrUserNotFound = errors.New("user not found")

type AuthService interface {
	Register(context.Context, *domain.AuthData) error
	Login(context.Context, *domain.AuthData) (string, error)
}

type AuthRepository interface {
	CreateUser(context.Context, *domain.User) (*domain.User, error)
	GetUserByLogin(context.Context, string) (*domain.User, error)
}
