package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/Svirex/gofermart-loyality/internal/core/domain"
	"github.com/Svirex/gofermart-loyality/internal/core/ports"
	passwordvalidator "github.com/wagslane/go-password-validator"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	repo                   ports.AuthRepository
	minPasswordEntropyBits float64
	minPasswordLength      int
	bcryptCost             int
}

var _ ports.AuthService = (*AuthService)(nil)

func NewAuthService(repo ports.AuthRepository, minPasswordEntropyBits float64, minPasswordLength int, bcryptCost int) (*AuthService, error) {
	return &AuthService{
		repo:                   repo,
		minPasswordEntropyBits: minPasswordEntropyBits,
		minPasswordLength:      minPasswordLength,
		bcryptCost:             bcryptCost,
	}, nil
}

func (s *AuthService) Register(ctx context.Context, login, password string) (string, error) {
	if login == "" {
		return "", fmt.Errorf("auth service register, empty login: %w", ports.ErrEmptyLogin)
	}
	if password == "" {
		return "", fmt.Errorf("auth service register, empty password: %w", ports.ErrEmptyPassword)
	}
	if err := s.validatePasswordSthregth(password); err != nil {
		return "", fmt.Errorf("auth service register, low password strength: %w", err)
	}
	hash, err := s.hashPassword(password)
	if err != nil {
		if errors.Is(err, bcrypt.ErrPasswordTooLong) {
			return "", fmt.Errorf("auth service register, password too long: %w", bcrypt.ErrPasswordTooLong)
		}
		return "", fmt.Errorf("auth service register, hash password error: %w", err)
	}
	user := &domain.User{
		Login: login,
		Hash:  hash,
	}

	return "", nil
}

func (s *AuthService) Login(ctx context.Context, login, password string) (string, error) {
	return "", nil
}

func (s *AuthService) validatePasswordSthregth(password string) error {
	if len(password) < s.minPasswordLength {
		return fmt.Errorf("validatePasswordSthregth, check length: %w", ports.ErrPasswordToShort)
	}
	err := passwordvalidator.Validate(password, s.minPasswordEntropyBits)
	if err != nil {
		return fmt.Errorf("%w, validatePasswordSthregth, check password strength: %v", ports.ErrLowPasswordStrength, err)
	}
	return nil
}

func (s *AuthService) hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), s.bcryptCost)
	return string(hash), err
}
