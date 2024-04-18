//go:build integration
// +build integration

package postgres

import (
	"context"
	"testing"

	"github.com/Svirex/gofermart-loyality/internal/core/domain"
	"github.com/Svirex/gofermart-loyality/test/testdb"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func NewAuthRepo() *AuthRepository {
	return NewAuthRepository(testdb.GetPool())
}

func TestCreateUser(t *testing.T) {
	repo := NewAuthRepo()
	hash, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	require.NoError(t, err)
	user := &domain.User{
		Login: "svirex",
		Hash:  string(hash),
	}
	user, err = repo.CreateUser(context.Background(), user)
	require.NoError(t, err)
	require.NotEqual(t, int64(0), user)
}
