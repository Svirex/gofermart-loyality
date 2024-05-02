//go:build integration
// +build integration

package postgres

import (
	"context"
	"math"
	"os"
	"testing"
	"time"

	"github.com/Svirex/gofermart-loyality/internal/core/domain"
	"github.com/Svirex/gofermart-loyality/internal/core/ports"
	"github.com/Svirex/gofermart-loyality/test/testdb"
	_ "github.com/golang-migrate/migrate/source/file"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

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

func NewOrdersTestRepo() *OrdersRepository {
	return NewOrdersRepository(testdb.GetPool(), testdb.GetLogger())
}

func TestCreateOrderGood(t *testing.T) {
	userRepo := NewAuthRepo()

	hash, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	require.NoError(t, err)
	user := &domain.User{
		Login: "svirex",
		Hash:  string(hash),
	}
	user, err = userRepo.CreateUser(context.Background(), user)
	require.NoError(t, err)

	repo := NewOrdersTestRepo()

	result, err := repo.CreateOrder(context.Background(), int64(user.ID), "4525634534")
	require.NotNil(t, result)
	require.NoError(t, err)

	require.Equal(t, int64(user.ID), result.ID)
	require.True(t, result.New)

	var accrual float64
	var uploadedAt time.Time
	var status string
	var uid int64
	err = testdb.GetPool().QueryRow(context.Background(),
		"SELECT uid, status, accrual, uploaded_at FROM orders WHERE order_num=$1", "4525634534").Scan(&uid, &status, &accrual, &uploadedAt)
	require.NoError(t, err)
	require.Equal(t, int64(user.ID), uid)
	require.Equal(t, float64(0.00), accrual)
	require.Equal(t, "NEW", status)

	err = testdb.Truncate()
	require.NoError(t, err)
}

func TestCreateOrderAlreadyExistsSameUser(t *testing.T) {
	userRepo := NewAuthRepo()

	hash, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	require.NoError(t, err)
	user := &domain.User{
		Login: "svirex",
		Hash:  string(hash),
	}
	user, err = userRepo.CreateUser(context.Background(), user)
	require.NoError(t, err)

	repo := NewOrdersTestRepo()

	result, err := repo.CreateOrder(context.Background(), int64(user.ID), "4525634534")
	require.NotNil(t, result)
	require.NoError(t, err)

	require.Equal(t, int64(user.ID), result.ID)
	require.True(t, result.New)

	result, err = repo.CreateOrder(context.Background(), int64(user.ID), "4525634534")
	require.NotNil(t, result)
	require.NoError(t, err)

	require.Equal(t, int64(user.ID), result.ID)
	require.False(t, result.New)

	err = testdb.Truncate()
	require.NoError(t, err)
}

func TestCreateOrdersAlreadyExistsAnotherUser(t *testing.T) {
	userRepo := NewAuthRepo()

	hash, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	require.NoError(t, err)
	user := &domain.User{
		Login: "svirex",
		Hash:  string(hash),
	}
	user, err = userRepo.CreateUser(context.Background(), user)
	require.NoError(t, err)

	anotherUser := &domain.User{
		Login: "another",
		Hash:  string(hash),
	}
	anotherUser, err = userRepo.CreateUser(context.Background(), anotherUser)
	require.NoError(t, err)

	repo := NewOrdersTestRepo()

	result, err := repo.CreateOrder(context.Background(), int64(anotherUser.ID), "4525634534")
	require.NotNil(t, result)
	require.NoError(t, err)

	result, err = repo.CreateOrder(context.Background(), int64(user.ID), "4525634534")
	require.NotNil(t, result)
	require.NoError(t, err)

	require.Equal(t, int64(anotherUser.ID), result.ID)
	require.False(t, result.New)

	err = testdb.Truncate()
	require.NoError(t, err)
}

func TestGetOrdersNotFound(t *testing.T) {
	repo := NewOrdersTestRepo()

	orders, err := repo.GetOrders(context.Background(), 1)
	require.NoError(t, err)
	require.NotNil(t, orders)
	require.Empty(t, orders)

	err = testdb.Truncate()
	require.NoError(t, err)
}

func TestGetOrdersOneOrder(t *testing.T) {
	userRepo := NewAuthRepo()

	hash, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	require.NoError(t, err)
	user := &domain.User{
		Login: "svirex",
		Hash:  string(hash),
	}
	user, err = userRepo.CreateUser(context.Background(), user)
	require.NoError(t, err)

	repo := NewOrdersTestRepo()

	result, err := repo.CreateOrder(context.Background(), int64(user.ID), "4525634534")
	require.NotNil(t, result)
	require.NoError(t, err)

	orders, err := repo.GetOrders(context.Background(), user.ID)
	require.NoError(t, err)
	require.NotNil(t, orders)
	require.NotEmpty(t, orders)

	require.Len(t, orders, 1)

	require.Equal(t, "4525634534", orders[0].Number)

	err = testdb.Truncate()
	require.NoError(t, err)
}

func TestGetOrdersOrderBy(t *testing.T) {
	userRepo := NewAuthRepo()

	hash, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	require.NoError(t, err)
	user := &domain.User{
		Login: "svirex",
		Hash:  string(hash),
	}
	user, err = userRepo.CreateUser(context.Background(), user)
	require.NoError(t, err)

	repo := NewOrdersTestRepo()

	result, err := repo.CreateOrder(context.Background(), int64(user.ID), "4525634534")
	require.NotNil(t, result)
	require.NoError(t, err)

	result, err = repo.CreateOrder(context.Background(), int64(user.ID), "4525634535")
	require.NotNil(t, result)
	require.NoError(t, err)

	result, err = repo.CreateOrder(context.Background(), int64(user.ID), "4525634536")
	require.NotNil(t, result)
	require.NoError(t, err)

	orders, err := repo.GetOrders(context.Background(), user.ID)
	require.NoError(t, err)
	require.NotNil(t, orders)
	require.NotEmpty(t, orders)

	require.Len(t, orders, 3)

	require.Equal(t, "4525634536", orders[0].Number)
	require.Equal(t, "4525634535", orders[1].Number)
	require.Equal(t, "4525634534", orders[2].Number)

	err = testdb.Truncate()
	require.NoError(t, err)
}

func NewTestBalanceRepository() *BalanceRepository {
	return NewBalanceRepository(testdb.GetPool(), testdb.GetLogger())
}

func TestBalanceGetBalanceForNewUser(t *testing.T) {
	userRepo := NewAuthRepo()

	hash, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	require.NoError(t, err)
	user := &domain.User{
		Login: "svirex",
		Hash:  string(hash),
	}
	user, err = userRepo.CreateUser(context.Background(), user)
	require.NoError(t, err)

	repo := NewTestBalanceRepository()

	data, err := repo.GetBalance(context.Background(), user.ID)
	require.NoError(t, err)

	require.True(t, math.Abs(0.0-data.Current) < 0.000001)
	require.True(t, math.Abs(0.0-data.Withdrawn) < 0.000001)

	data, err = repo.GetBalance(context.Background(), user.ID)
	require.NoError(t, err)

	require.True(t, math.Abs(0.0-data.Current) < 0.000001)
	require.True(t, math.Abs(0.0-data.Withdrawn) < 0.000001)

	err = testdb.Truncate()
	require.NoError(t, err)
}
