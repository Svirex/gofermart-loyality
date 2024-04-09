package services

import (
	"context"
	"fmt"

	"github.com/Svirex/gofermart-loyality/internal/core/domain"
	"github.com/Svirex/gofermart-loyality/internal/core/ports"
)

type AuthService struct {
	repo ports.AuthRepository
}

var _ ports.AuthService = (*AuthService)(nil)

func NewAuthService(repo ports.AuthRepository) (*AuthService, error) {
	return &AuthService{
		repo: repo,
	}, nil
}

func (s *AuthService) Register(ctx context.Context, data *domain.AuthData) error {
	if data.Login == "" {
		return fmt.Errorf("auth service registr, empty login: %w", ports.ErrEmptyLogin)
	}
	if data.Password == "" {
		return fmt.Errorf("auth service registr, empty password: %w", ports.ErrEmptyPassword)
	}
	s.repo.CreateUser(ctx, &domain.User{
		Login: data.Login,
		Hash:  data.Password,
	})
	return nil
}

func (s *AuthService) Login(context.Context, *domain.AuthData) (string, error) {
	return "", nil
}
