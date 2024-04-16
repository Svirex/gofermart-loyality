package services_test

import (
	"context"
	"testing"

	"github.com/Svirex/gofermart-loyality/internal/core/ports"
	"github.com/Svirex/gofermart-loyality/internal/core/services"
	"github.com/stretchr/testify/require"
)

func TestEmptyLogin(t *testing.T) {

	service, err := services.NewAuthService(nil, 80, 8, 10)
	require.NoError(t, err)

	_, err = service.Register(context.Background(), "", "password")
	require.ErrorIs(t, err, ports.ErrEmptyLogin)
}

func TestEmptyPassword(t *testing.T) {
	service, err := services.NewAuthService(nil, 80, 8, 10)
	require.NoError(t, err)

	_, err = service.Register(context.Background(), "login", "")
	require.ErrorIs(t, err, ports.ErrEmptyPassword)
}

func TestPasswordIsTooShort(t *testing.T) {
	service, err := services.NewAuthService(nil, 80, 8, 10)
	require.NoError(t, err)

	_, err = service.Register(context.Background(), "login", "pass")
	require.ErrorIs(t, err, ports.ErrPasswordToShort)

}

func TestLowPasswordStrength(t *testing.T) {
	service, err := services.NewAuthService(nil, 80, 8, 10)
	require.NoError(t, err)

	_, err = service.Register(context.Background(), "login", "password")
	require.ErrorIs(t, err, ports.ErrLowPasswordStrength)

}
