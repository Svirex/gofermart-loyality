//go:build integration
// +build integration

package services

import (
	"context"
	"fmt"
	"math"
	"os"
	"testing"
	"time"

	"github.com/Svirex/gofermart-loyality/internal/adapters/postgres"
	"github.com/Svirex/gofermart-loyality/internal/core/domain"
	"github.com/Svirex/gofermart-loyality/internal/core/ports"
	"github.com/Svirex/gofermart-loyality/test/testdb"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func NewAuthIntegrationTestService(t *testing.T) *AuthService {
	repo := postgres.NewAuthRepository(testdb.GetPool())
	service, err := NewAuthService(repo, 80, 8, 10, "fake_secret")
	require.NoError(t, err)
	return service
}

func TestMain(m *testing.M) {
	testdb.Init()
	defer testdb.Close()

	testdb.MigrateUp()
	code := m.Run()
	testdb.MigrateDown()
	os.Exit(code)
}

func TestUserAlreadyExists(t *testing.T) {
	service := NewAuthIntegrationTestService(t)

	_, err := service.Register(context.Background(), "svirex", "SuperPuperPassword")
	require.NoError(t, err)

	key, err := service.Register(context.Background(), "svirex", "SuperPuperPassword")
	require.Empty(t, key)
	require.ErrorIs(t, err, ports.ErrUserAlreadyExists)

	err = testdb.Truncate()
	require.NoError(t, err)
}

func TestGoodRegister(t *testing.T) {
	service := NewAuthIntegrationTestService(t)

	key, err := service.Register(context.Background(), "svirex", "SuperPuperPassword")
	require.NotEmpty(t, key)
	require.NoError(t, err)

	uid, err := getUserIDFromJWT("fake_secret", key)
	require.NotEqual(t, int64(-1), uid)
	require.NoError(t, err)

	err = testdb.Truncate()
	require.NoError(t, err)
}

func NewOrdersTestService(t *testing.T) *OrderService {
	repo := postgres.NewOrdersRepository(testdb.GetPool(), testdb.GetLogger())
	service, err := NewOrderService(testdb.GetPool(), repo, testdb.GetLogger(), 20, "http://mock_accrual:3000", 1*time.Second, 20, 2*time.Second)
	require.NoError(t, err)
	return service
}

func TestCreateOrderNotNum(t *testing.T) {
	service := NewOrdersTestService(t)

	status, err := service.CreateOrder(context.Background(), 1, "йо-хо-хо")
	require.Error(t, err)
	require.Equal(t, ports.Err, status)

	service.Shutdown()
	err = testdb.Truncate()
	require.NoError(t, err)
}

func TestCreateOrderInvalidNum(t *testing.T) {
	service := NewOrdersTestService(t)

	status, err := service.CreateOrder(context.Background(), 1, "4561261212345464")
	require.ErrorIs(t, err, ports.ErrInvalidOrderNum)
	require.Equal(t, ports.Err, status)

	service.Shutdown()
	err = testdb.Truncate()
	require.NoError(t, err)
}

func TestCreateOrderValidNewNum(t *testing.T) {
	userRepo := postgres.NewAuthRepository(testdb.GetPool())

	hash, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	require.NoError(t, err)
	user := &domain.User{
		Login: "svirex",
		Hash:  string(hash),
	}
	user, err = userRepo.CreateUser(context.Background(), user)
	require.NoError(t, err)

	service := NewOrdersTestService(t)

	status, err := service.CreateOrder(context.Background(), user.ID, "4561261212345467")
	require.NoError(t, err)
	require.Equal(t, ports.Ok, status)

	service.Shutdown()
	err = testdb.Truncate()
	require.NoError(t, err)
}

func TestCreateOrderAlreadyAdded(t *testing.T) {
	userRepo := postgres.NewAuthRepository(testdb.GetPool())

	hash, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	require.NoError(t, err)
	user := &domain.User{
		Login: "svirex",
		Hash:  string(hash),
	}
	user, err = userRepo.CreateUser(context.Background(), user)
	require.NoError(t, err)

	service := NewOrdersTestService(t)

	status, err := service.CreateOrder(context.Background(), user.ID, "4561261212345467")
	require.NoError(t, err)
	require.Equal(t, ports.Ok, status)

	status, err = service.CreateOrder(context.Background(), user.ID, "4561261212345467")
	require.NoError(t, err)
	require.Equal(t, ports.AlreadyAdded, status)

	service.Shutdown()
	err = testdb.Truncate()
	require.NoError(t, err)
}

func TestCreateOrderAddedByAnotherUser(t *testing.T) {
	userRepo := postgres.NewAuthRepository(testdb.GetPool())

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

	service := NewOrdersTestService(t)

	status, err := service.CreateOrder(context.Background(), anotherUser.ID, "4561261212345467")
	require.NoError(t, err)
	require.Equal(t, ports.Ok, status)

	status, err = service.CreateOrder(context.Background(), user.ID, "4561261212345467")
	require.NoError(t, err)
	require.Equal(t, ports.NotOwnOrder, status)

	service.Shutdown()
	err = testdb.Truncate()
	require.NoError(t, err)
}

func TestGetOrdersNotFound(t *testing.T) {
	service := NewOrdersTestService(t)

	orders, err := service.GetOrders(context.Background(), 1)
	require.NoError(t, err)
	require.NotNil(t, orders)
	require.Empty(t, orders)

	service.Shutdown()
	err = testdb.Truncate()
	require.NoError(t, err)
}

func TestGetOrdersOneOrder(t *testing.T) {
	userRepo := postgres.NewAuthRepository(testdb.GetPool())

	hash, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	require.NoError(t, err)
	user := &domain.User{
		Login: "svirex",
		Hash:  string(hash),
	}
	user, err = userRepo.CreateUser(context.Background(), user)
	require.NoError(t, err)

	service := NewOrdersTestService(t)

	result, err := service.CreateOrder(context.Background(), int64(user.ID), "4561261212345467")
	require.NotNil(t, result)
	require.NoError(t, err)

	orders, err := service.GetOrders(context.Background(), user.ID)
	require.NoError(t, err)
	require.NotNil(t, orders)
	require.NotEmpty(t, orders)

	require.Len(t, orders, 1)

	require.Equal(t, "4561261212345467", orders[0].Number)

	service.Shutdown()
	err = testdb.Truncate()
	require.NoError(t, err)
}

func TestGetOrdersOrderBy(t *testing.T) {
	userRepo := postgres.NewAuthRepository(testdb.GetPool())

	hash, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	require.NoError(t, err)
	user := &domain.User{
		Login: "svirex",
		Hash:  string(hash),
	}
	user, err = userRepo.CreateUser(context.Background(), user)
	require.NoError(t, err)

	service := NewOrdersTestService(t)

	result, err := service.CreateOrder(context.Background(), int64(user.ID), "4561261212345467")
	require.NotNil(t, result)
	require.NoError(t, err)

	result, err = service.CreateOrder(context.Background(), int64(user.ID), "2634")
	require.NotNil(t, result)
	require.NoError(t, err)

	result, err = service.CreateOrder(context.Background(), int64(user.ID), "8334")
	require.NotNil(t, result)
	require.NoError(t, err)

	orders, err := service.GetOrders(context.Background(), user.ID)
	require.NoError(t, err)
	require.NotNil(t, orders)
	require.NotEmpty(t, orders)

	require.Len(t, orders, 3)

	require.Equal(t, "8334", orders[0].Number)
	require.Equal(t, "2634", orders[1].Number)
	require.Equal(t, "4561261212345467", orders[2].Number)

	service.Shutdown()
	err = testdb.Truncate()
	require.NoError(t, err)
}

func TestCheckAccrualInvalid(t *testing.T) {
	userRepo := postgres.NewAuthRepository(testdb.GetPool())

	hash, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	require.NoError(t, err)
	user := &domain.User{
		Login: "svirex",
		Hash:  string(hash),
	}
	user, err = userRepo.CreateUser(context.Background(), user)
	require.NoError(t, err)

	service := NewOrdersTestService(t)

	result, err := service.CreateOrder(context.Background(), int64(user.ID), "18")
	require.NotNil(t, result)
	require.NoError(t, err)

	time.Sleep(2 * time.Second)

	orders, err := service.GetOrders(context.Background(), user.ID)
	require.NoError(t, err)
	require.NotNil(t, orders)
	require.NotEmpty(t, orders)

	require.Len(t, orders, 1)

	require.Equal(t, "18", orders[0].Number)
	require.Equal(t, domain.Invalid, orders[0].Status)

	service.Shutdown()
	err = testdb.Truncate()
	require.NoError(t, err)
}

func TestCheckAccrualProcessing(t *testing.T) {
	userRepo := postgres.NewAuthRepository(testdb.GetPool())

	hash, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	require.NoError(t, err)
	user := &domain.User{
		Login: "svirex",
		Hash:  string(hash),
	}
	user, err = userRepo.CreateUser(context.Background(), user)
	require.NoError(t, err)

	service := NewOrdersTestService(t)

	result, err := service.CreateOrder(context.Background(), int64(user.ID), "26")
	require.NotNil(t, result)
	require.NoError(t, err)

	time.Sleep(2 * time.Second)

	orders, err := service.GetOrders(context.Background(), user.ID)
	require.NoError(t, err)
	require.NotNil(t, orders)
	require.NotEmpty(t, orders)

	require.Len(t, orders, 1)

	require.Equal(t, "26", orders[0].Number)
	require.Equal(t, domain.Processing, orders[0].Status)

	service.Shutdown()
	err = testdb.Truncate()
	require.NoError(t, err)
}

func TestCheckAccrualChangePauseBetweenRequests(t *testing.T) {
	userRepo := postgres.NewAuthRepository(testdb.GetPool())

	hash, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	require.NoError(t, err)
	user := &domain.User{
		Login: "svirex",
		Hash:  string(hash),
	}
	user, err = userRepo.CreateUser(context.Background(), user)
	require.NoError(t, err)

	service := NewOrdersTestService(t)

	result, err := service.CreateOrder(context.Background(), int64(user.ID), "59")
	require.NotNil(t, result)
	require.NoError(t, err)

	time.Sleep(2 * time.Second)

	require.Equal(t, 30*time.Second, service.checkAccrualService.pauseBetweenRequests)

	service.Shutdown()
	err = testdb.Truncate()
	require.NoError(t, err)
}

func TestCheckAccrualDBLoader(t *testing.T) {
	userRepo := postgres.NewAuthRepository(testdb.GetPool())

	hash, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	require.NoError(t, err)
	user := &domain.User{
		Login: "svirex",
		Hash:  string(hash),
	}
	user, err = userRepo.CreateUser(context.Background(), user)
	require.NoError(t, err)

	_, err = testdb.GetPool().Exec(context.Background(), "INSERT INTO orders (uid, order_num) VALUES ($1, $2);", user.ID, "1")
	require.NoError(t, err)

	_, err = testdb.GetPool().Exec(context.Background(), "INSERT INTO orders (uid, order_num) VALUES ($1, $2);", user.ID, "2")
	require.NoError(t, err)

	_, err = testdb.GetPool().Exec(context.Background(), "INSERT INTO orders (uid, order_num) VALUES ($1, $2);", user.ID, "3")
	require.NoError(t, err)

	service := NewOrdersTestService(t)

	time.Sleep(5 * time.Second)

	orders, err := service.GetOrders(context.Background(), user.ID)
	require.NoError(t, err)
	require.NotNil(t, orders)
	require.NotEmpty(t, orders)

	require.Len(t, orders, 3)

	fmt.Println(orders)

	require.Equal(t, "3", orders[0].Number)
	require.Equal(t, domain.Invalid, orders[0].Status)

	require.Equal(t, "2", orders[1].Number)
	require.Equal(t, domain.Invalid, orders[1].Status)

	require.Equal(t, "1", orders[2].Number)
	require.Equal(t, domain.Invalid, orders[2].Status)

	service.Shutdown()
	err = testdb.Truncate()
	require.NoError(t, err)
}

func NewTestBalanceService() *BalanceService {
	repo := postgres.NewBalanceRepository(testdb.GetPool(), testdb.GetLogger())
	return NewBalanceService(repo)
}

func TestCheckAccrualProcessed(t *testing.T) {
	userRepo := postgres.NewAuthRepository(testdb.GetPool())

	hash, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	require.NoError(t, err)
	user := &domain.User{
		Login: "svirex",
		Hash:  string(hash),
	}
	user, err = userRepo.CreateUser(context.Background(), user)
	require.NoError(t, err)

	service := NewOrdersTestService(t)

	result, err := service.CreateOrder(context.Background(), int64(user.ID), "67")
	require.NotNil(t, result)
	require.NoError(t, err)

	time.Sleep(8 * time.Second)

	bs := NewTestBalanceService()
	balance, err := bs.GetBalance(context.Background(), user.ID)
	require.NoError(t, err)

	fmt.Println(balance)

	require.True(t, math.Abs(500.0-balance.Current) < 1e-9)
	require.True(t, math.Abs(0.0-balance.Withdrawn) < 1e-9)

	service.Shutdown()
	err = testdb.Truncate()
	require.NoError(t, err)
}

func TestCheckAccrualProcessed2(t *testing.T) {
	userRepo := postgres.NewAuthRepository(testdb.GetPool())

	hash, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	require.NoError(t, err)
	user := &domain.User{
		Login: "svirex",
		Hash:  string(hash),
	}
	user, err = userRepo.CreateUser(context.Background(), user)
	require.NoError(t, err)

	service := NewOrdersTestService(t)

	result, err := service.CreateOrder(context.Background(), int64(user.ID), "3511871356")
	require.NotNil(t, result)
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	bs := NewTestBalanceService()
	balance, err := bs.GetBalance(context.Background(), user.ID)
	require.NoError(t, err)

	fmt.Println(balance)

	require.True(t, math.Abs(729.98-balance.Current) < 1e-9)
	require.True(t, math.Abs(0.0-balance.Withdrawn) < 1e-9)

	service.Shutdown()
	err = testdb.Truncate()
	require.NoError(t, err)
}

func TestGetOrders(t *testing.T) {
	userRepo := postgres.NewAuthRepository(testdb.GetPool())

	hash, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	require.NoError(t, err)
	user := &domain.User{
		Login: "svirex",
		Hash:  string(hash),
	}
	user, err = userRepo.CreateUser(context.Background(), user)
	require.NoError(t, err)

	service := NewOrdersTestService(t)

	result, err := service.CreateOrder(context.Background(), int64(user.ID), "3511871356")
	require.NotNil(t, result)
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	orders, err := service.GetOrders(context.Background(), user.ID)
	require.NoError(t, err)
	require.Len(t, orders, 1)

	require.True(t, math.Abs(729.98-orders[0].Accrual) < 1e-9)

	service.Shutdown()
	err = testdb.Truncate()
	require.NoError(t, err)
}
