package services_test

import (
	"context"
	"testing"

	"github.com/Svirex/gofermart-loyality/internal/core/domain"
	"github.com/Svirex/gofermart-loyality/internal/core/ports"
	"github.com/Svirex/gofermart-loyality/internal/core/services"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestGoodRegister(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockRepo := NewMockAuthRepository(ctrl)

	mockRepo.EXPECT().CreateUser(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, user *domain.User) (*domain.User, error) {
		user.ID = 1
		return user, nil
	})

	service, err := services.NewAuthService(mockRepo)
	require.NoError(t, err)

	user := &domain.AuthData{
		Login:    "login",
		Password: "password",
	}

	err = service.Register(context.Background(), user)
	require.NoError(t, err)
}

func TestEmptyLogin(t *testing.T) {

	service, err := services.NewAuthService(nil)
	require.NoError(t, err)

	user := &domain.AuthData{
		Login:    "",
		Password: "password",
	}

	err = service.Register(context.Background(), user)
	require.ErrorIs(t, err, ports.ErrEmptyLogin)
}

func TestEmptyPassword(t *testing.T) {
	service, err := services.NewAuthService(nil)
	require.NoError(t, err)

	user := &domain.AuthData{
		Login:    "login",
		Password: "",
	}

	err = service.Register(context.Background(), user)
	require.ErrorIs(t, err, ports.ErrEmptyPassword)
}

func TestLowPasswordStrength(t *testing.T) {
	service, err := services.NewAuthService(nil)
	require.NoError(t, err)

	user := &domain.AuthData{
		Login:    "login",
		Password: "password",
	}

	err = service.Register(context.Background(), user)
	require.ErrorIs(t, err, ports.ErrLowPasswordStrength)

}

func TestUserAlreadyExists(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockRepo := NewMockAuthRepository(ctrl)

	var prevLogin string

	mockRepo.EXPECT().CreateUser(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, user *domain.User) (*domain.User, error) {
		if user.Login == prevLogin {
			return nil, ports.ErrUserAlreadyExists
		}
		prevLogin = user.Login
		user.ID = 1
		return user, nil
	})

	service, err := services.NewAuthService(mockRepo)
	require.NoError(t, err)

	user := &domain.AuthData{
		Login:    "login",
		Password: "345VZDFw5Q@^sdf*(*$)",
	}

	err = service.Register(context.Background(), user)
	require.NoError(t, err)

	err = service.Register(context.Background(), user)
	require.ErrorIs(t, err, ports.ErrUserAlreadyExists)
}
