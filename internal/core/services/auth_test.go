package services_test

import (
	"context"
	"testing"

	"github.com/Svirex/gofermart-loyality/internal/core/domain"
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

}
