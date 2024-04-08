package services

import (
	"context"

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

func (s *AuthService) Register(context.Context, *domain.AuthData) error

func (s *AuthService) Login(context.Context, *domain.AuthData) (string, error)
