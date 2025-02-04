package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/Svirex/gofermart-loyality/internal/core/domain"
	"github.com/Svirex/gofermart-loyality/internal/core/ports"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	repo                   ports.AuthRepository
	minPasswordEntropyBits float64
	minPasswordLength      int
	bcryptCost             int
	jwtSecretKey           string
}

var _ ports.AuthService = (*AuthService)(nil)

func NewAuthService(repo ports.AuthRepository, minPasswordEntropyBits float64, minPasswordLength int, bcryptCost int, jwtSecretKey string) (*AuthService, error) {
	return &AuthService{
		// TODO убрать параметры в Config
		repo:                   repo,
		minPasswordEntropyBits: minPasswordEntropyBits,
		minPasswordLength:      minPasswordLength,
		bcryptCost:             bcryptCost,
		jwtSecretKey:           jwtSecretKey,
	}, nil
}

func (s *AuthService) Register(ctx context.Context, login, password string) (string, error) {
	if login == "" {
		return "", fmt.Errorf("auth service register, empty login: %w", ports.ErrEmptyLogin)
	}
	if password == "" {
		return "", fmt.Errorf("auth service register, empty password: %w", ports.ErrEmptyPassword)
	}
	if len(password) < s.minPasswordLength {
		return "", fmt.Errorf("auth service register, check min length: %w", ports.ErrPasswordTooShort)
	}
	hash, err := s.hashPassword(password)
	if err != nil {
		if errors.Is(err, bcrypt.ErrPasswordTooLong) {
			return "", fmt.Errorf("auth service register, password too long: %w", ports.ErrPasswordTooLong)
		}
		return "", fmt.Errorf("auth service register, hash password error: %w", err)
	}
	user := &domain.User{
		Login: login,
		Hash:  hash,
	}
	user, err = s.repo.CreateUser(ctx, user)
	if err != nil {
		if errors.Is(err, ports.ErrUserAlreadyExists) {
			return "", fmt.Errorf("auth service register, user already exists: %w", err)
		}
		return "", fmt.Errorf("auth service register, create user: %w", err)
	}
	token, err := buildJWTString(s.jwtSecretKey, user.ID)
	if err != nil {
		return "", fmt.Errorf("auth service register, build jwt token: %w", err)
	}

	return token, nil
}

func (s *AuthService) Login(ctx context.Context, login, password string) (string, error) {
	if login == "" {
		return "", fmt.Errorf("auth service login, empty login: %w", ports.ErrEmptyLogin)
	}
	if password == "" {
		return "", fmt.Errorf("auth service login, empty password: %w", ports.ErrEmptyPassword)
	}
	user, err := s.repo.GetUserByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, ports.ErrUserNotFound) {
			return "", fmt.Errorf("auth service login, user not found: %w", err)
		}
		return "", fmt.Errorf("auth service login, get user by login: %w", err)
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Hash), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return "", fmt.Errorf("%w: auth service login, invalid password: %v", ports.ErrInvalidPassword, err)
		}
		return "", fmt.Errorf("auth service login, compare hash and password: %w", err)
	}
	token, err := buildJWTString(s.jwtSecretKey, user.ID)
	if err != nil {
		return "", fmt.Errorf("auth service login, build jwt token: %w", err)
	}
	return token, nil
}

func (s *AuthService) GetUserGromJWT(ctx context.Context, jwt string) (int64, error) {
	return getUserIDFromJWT(s.jwtSecretKey, jwt)
}

// func (s *AuthService) validatePasswordSthregth(password string) error {
// 	if len(password) < s.minPasswordLength {
// 		return fmt.Errorf("validatePasswordSthregth, check length: %w", ports.ErrPasswordTooShort)
// 	}
// 	err := passwordvalidator.Validate(password, s.minPasswordEntropyBits)
// 	if err != nil {
// 		return fmt.Errorf("%w, validatePasswordSthregth, check password strength: %v", ports.ErrLowPasswordStrength, err)
// 	}
// 	return nil
// }

func (s *AuthService) hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), s.bcryptCost)
	return string(hash), err
}
