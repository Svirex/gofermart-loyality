//go:build integration
// +build integration

package postgres

import (
	"context"
	"os"
	"testing"

	"github.com/Svirex/gofermart-loyality/internal/core/domain"
	"github.com/Svirex/gofermart-loyality/internal/core/ports"
	"github.com/Svirex/gofermart-loyality/test/testdb"
	_ "github.com/golang-migrate/migrate/source/file"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

var logger *zap.SugaredLogger

func NewAuthRepo() *AuthRepository {
	return NewAuthRepository(testdb.GetPool())
}

func TestMain(m *testing.M) {
	testdb.Init()
	defer testdb.Close()

	testdb.MigrateUp()
	code := m.Run()
	testdb.MigrateDown()
	os.Exit(code)
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
	require.NotEqual(t, int64(0), user.ID)

	var id int64

	err = testdb.GetPool().QueryRow(context.Background(), "SELECT id FROM users WHERE login='svirex'").Scan(&id)
	require.NoError(t, err)

	err = testdb.Truncate()
	require.NoError(t, err)
}

func TestErrorCreateUserExistsLogin(t *testing.T) {
	repo := NewAuthRepo()
	hash, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	require.NoError(t, err)
	user := &domain.User{
		Login: "svirex",
		Hash:  string(hash),
	}
	user, err = repo.CreateUser(context.Background(), user)
	require.NoError(t, err)
	require.NotEqual(t, int64(0), user.ID)

	user.ID = 0
	user, err = repo.CreateUser(context.Background(), user)
	require.ErrorIs(t, err, ports.ErrUserAlreadyExists)
	require.Nil(t, user)

	err = testdb.Truncate()
	require.NoError(t, err)
}

func TestGetUserNotFound(t *testing.T) {
	repo := NewAuthRepo()
	_, err := repo.GetUserByLogin(context.Background(), "svirex")
	require.ErrorIs(t, err, ports.ErrUserNotFound)

	hash, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	require.NoError(t, err)
	user := &domain.User{
		Login: "svirex",
		Hash:  string(hash),
	}
	user, err = repo.CreateUser(context.Background(), user)
	require.NoError(t, err)

	_, err = repo.GetUserByLogin(context.Background(), "nobody")
	require.ErrorIs(t, err, ports.ErrUserNotFound)

	err = testdb.Truncate()
	require.NoError(t, err)
}

func TestGetUserByLogin(t *testing.T) {
	repo := NewAuthRepo()
	hash, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	require.NoError(t, err)
	user := &domain.User{
		Login: "svirex",
		Hash:  string(hash),
	}
	user, err = repo.CreateUser(context.Background(), user)
	require.NoError(t, err)

	user2, err := repo.GetUserByLogin(context.Background(), "svirex")
	require.NoError(t, err)
	require.Equal(t, user.ID, user2.ID)
	require.Equal(t, user.Login, user2.Login)
	require.Equal(t, user.Hash, user2.Hash)

	err = testdb.Truncate()
	require.NoError(t, err)
}
