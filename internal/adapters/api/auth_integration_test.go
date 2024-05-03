//go:build integration || integration_api
// +build integration integration_api

package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Svirex/gofermart-loyality/internal/adapters/postgres"
	"github.com/Svirex/gofermart-loyality/internal/core/services"
	"github.com/Svirex/gofermart-loyality/test/testdb"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	testdb.Init()
	defer testdb.Close()

	testdb.MigrateUp()
	code := m.Run()
	testdb.MigrateDown()
	os.Exit(code)
}

func setupTest(t *testing.T) func() {
	// Setup code here

	// tear down later
	return func() {
		err := testdb.Truncate()
		require.NoError(t, err)
	}
}

func NewTestServer(t *testing.T) *httptest.Server {
	authRepo := postgres.NewAuthRepository(testdb.GetPool())
	auth, err := services.NewAuthService(authRepo, 80, 8, 10, "fake_secret")
	require.NoError(t, err)

	ordersRepo := postgres.NewOrdersRepository(testdb.GetPool(), testdb.GetLogger())
	orders, err := services.NewOrderService(testdb.GetPool(), ordersRepo, testdb.GetLogger(), 20, "http://mock_accrual:3000", 1*time.Second, 20, 2*time.Second)

	balanceRepo := postgres.NewBalanceRepository(testdb.GetPool(), testdb.GetLogger())
	balance := services.NewBalanceService(balanceRepo)

	withdrawRepo := postgres.NewWithdrawRepository(testdb.GetPool(), testdb.GetLogger())
	withdraw := services.NewWithdrawService(withdrawRepo)

	api := NewAPI(auth, orders, balance, withdraw, testdb.GetLogger())

	return httptest.NewServer(api.Routes())
}

func TestRegisterInvalidContentType(t *testing.T) {
	defer setupTest(t)()

	testServer := NewTestServer(t)

	req, err := http.NewRequest(http.MethodPost, testServer.URL+"/api/user/register", nil)
	require.NoError(t, err)

	resp, err := testServer.Client().Do(req)
	require.NoError(t, err)
	resp.Body.Close()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestRegisterNilBody(t *testing.T) {
	defer setupTest(t)()

	testServer := NewTestServer(t)

	req, err := http.NewRequest(http.MethodPost, testServer.URL+"/api/user/register", nil)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := testServer.Client().Do(req)
	require.NoError(t, err)
	resp.Body.Close()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestRegisterEmptyBody(t *testing.T) {
	defer setupTest(t)()

	testServer := NewTestServer(t)

	body := ``

	req, err := http.NewRequest(http.MethodPost, testServer.URL+"/api/user/register", strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := testServer.Client().Do(req)
	require.NoError(t, err)
	resp.Body.Close()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestRegisterUnmarshalError(t *testing.T) {
	defer setupTest(t)()

	testServer := NewTestServer(t)

	body := `sgasdfsdvgzfvzxcvzxcv`

	req, err := http.NewRequest(http.MethodPost, testServer.URL+"/api/user/register", strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := testServer.Client().Do(req)
	require.NoError(t, err)
	resp.Body.Close()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestRegisterGood(t *testing.T) {
	defer setupTest(t)()

	testServer := NewTestServer(t)

	body := `
	{
		"login": "svirex",
		"password": "password"
	}
	`

	req, err := http.NewRequest(http.MethodPost, testServer.URL+"/api/user/register", strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := testServer.Client().Do(req)
	require.NoError(t, err)
	resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	foundJWTCookie := false
	for _, v := range resp.Cookies() {
		if v.Name == "jwt" {
			foundJWTCookie = true
			break
		}
	}
	require.True(t, foundJWTCookie)
}

func TestLoginInvalidContentType(t *testing.T) {
	defer setupTest(t)()

	testServer := NewTestServer(t)

	req, err := http.NewRequest(http.MethodPost, testServer.URL+"/api/user/login", nil)
	require.NoError(t, err)

	resp, err := testServer.Client().Do(req)
	require.NoError(t, err)
	resp.Body.Close()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestLoginNilBody(t *testing.T) {
	defer setupTest(t)()

	testServer := NewTestServer(t)

	req, err := http.NewRequest(http.MethodPost, testServer.URL+"/api/user/login", nil)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := testServer.Client().Do(req)
	require.NoError(t, err)
	resp.Body.Close()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestLoginEmptyBody(t *testing.T) {
	defer setupTest(t)()

	testServer := NewTestServer(t)

	body := ``

	req, err := http.NewRequest(http.MethodPost, testServer.URL+"/api/user/login", strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := testServer.Client().Do(req)
	require.NoError(t, err)
	resp.Body.Close()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestLoginUnmarshalError(t *testing.T) {
	defer setupTest(t)()

	testServer := NewTestServer(t)

	body := `sgasdfsdvgzfvzxcvzxcv`

	req, err := http.NewRequest(http.MethodPost, testServer.URL+"/api/user/login", strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := testServer.Client().Do(req)
	require.NoError(t, err)
	resp.Body.Close()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestLoginGood(t *testing.T) {
	defer setupTest(t)()

	testServer := NewTestServer(t)

	body := `
	{
		"login": "svirex",
		"password": "password"
	}
	`

	req, err := http.NewRequest(http.MethodPost, testServer.URL+"/api/user/register", strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := testServer.Client().Do(req)
	require.NoError(t, err)
	resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	req, err = http.NewRequest(http.MethodPost, testServer.URL+"/api/user/login", strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err = testServer.Client().Do(req)
	require.NoError(t, err)
	resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	foundJWTCookie := false
	for _, v := range resp.Cookies() {
		if v.Name == "jwt" {
			foundJWTCookie = true
			break
		}
	}
	require.True(t, foundJWTCookie)
}

func TestLoginUserNotFound(t *testing.T) {
	defer setupTest(t)()

	testServer := NewTestServer(t)

	body := `
	{
		"login": "svirex",
		"password": "password"
	}
	`

	req, err := http.NewRequest(http.MethodPost, testServer.URL+"/api/user/register", strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := testServer.Client().Do(req)
	require.NoError(t, err)
	resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	body = `
	{
		"login": "tytyrtyrt",
		"password": "password"
	}
	`

	req, err = http.NewRequest(http.MethodPost, testServer.URL+"/api/user/login", strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err = testServer.Client().Do(req)
	require.NoError(t, err)
	resp.Body.Close()

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)

}

func TestLoginInvalidPassword(t *testing.T) {
	defer setupTest(t)()

	testServer := NewTestServer(t)

	body := `
	{
		"login": "svirex",
		"password": "password"
	}
	`

	req, err := http.NewRequest(http.MethodPost, testServer.URL+"/api/user/register", strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := testServer.Client().Do(req)
	require.NoError(t, err)
	resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	body = `
	{
		"login": "svirex",
		"password": "passwordblalblabla"
	}
	`

	req, err = http.NewRequest(http.MethodPost, testServer.URL+"/api/user/login", strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err = testServer.Client().Do(req)
	require.NoError(t, err)
	resp.Body.Close()

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)

}
