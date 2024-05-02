package services

import (
	"context"
	"testing"

	"github.com/Svirex/gofermart-loyality/internal/core/ports"
	"github.com/stretchr/testify/require"
)

func NewAuthTestService(t *testing.T) *AuthService {
	service, err := NewAuthService(nil, 80, 8, 10, "fake_secret")
	require.NoError(t, err)
	return service
}

func TestEmptyLogin(t *testing.T) {

	service := NewAuthTestService(t)

	_, err := service.Register(context.Background(), "", "password")
	require.ErrorIs(t, err, ports.ErrEmptyLogin)
}

func TestEmptyPassword(t *testing.T) {
	service := NewAuthTestService(t)

	_, err := service.Register(context.Background(), "login", "")
	require.ErrorIs(t, err, ports.ErrEmptyPassword)
}

func TestPasswordIsTooShort(t *testing.T) {
	service := NewAuthTestService(t)

	_, err := service.Register(context.Background(), "login", "pass")
	require.ErrorIs(t, err, ports.ErrPasswordTooShort)

}

func TestEmptyLoginLogin(t *testing.T) {

	service := NewAuthTestService(t)

	_, err := service.Login(context.Background(), "", "password")
	require.ErrorIs(t, err, ports.ErrEmptyLogin)
}

func TestEmptyPasswordLogin(t *testing.T) {
	service := NewAuthTestService(t)

	_, err := service.Login(context.Background(), "login", "")
	require.ErrorIs(t, err, ports.ErrEmptyPassword)
}

func TestPasswordTooLong(t *testing.T) {
	service := NewAuthTestService(t)

	_, err := service.Register(context.Background(), "svirex", "blablablablablablablablablablablablablablablablablablablablablablablablabla")
	require.ErrorIs(t, err, ports.ErrPasswordTooLong)
}
