//go:build integration || integration_api
// +build integration integration_api

package api

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/Svirex/gofermart-loyality/internal/core/domain"
	"github.com/stretchr/testify/require"
)

func TestWithdrawNotAuth(t *testing.T) {
	defer setupTest(t)()

	testServer := NewTestServer(t)

	req, err := http.NewRequest(http.MethodPost, testServer.URL+"/api/user/balance/withdraw", nil)
	require.NoError(t, err)

	resp, err := testServer.Client().Do(req)
	require.NoError(t, err)
	resp.Body.Close()

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestWithdrawInvalidOrderNum(t *testing.T) {
	defer setupTest(t)()

	testServer := NewTestServer(t)

	client := RegisterTestUser(t, testServer, testServer.URL, "test")

	body := `
	{
		"order": "1",
		"sum": 751
	}
	`

	req, err := http.NewRequest(http.MethodPost, testServer.URL+"/api/user/balance/withdraw", strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err)

	require.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
}

func TestWithdrawDuplicate(t *testing.T) {
	defer setupTest(t)()

	testServer := NewTestServer(t)

	client := RegisterTestUser(t, testServer, testServer.URL, "test")

	body := `
	{
		"order": "67",
		"sum": 0
	}
	`

	req, err := http.NewRequest(http.MethodPost, testServer.URL+"/api/user/balance/withdraw", strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, resp.StatusCode)

	req, err = http.NewRequest(http.MethodPost, testServer.URL+"/api/user/balance/withdraw", strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err = client.Do(req)
	require.NoError(t, err)

	require.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
}

func TestWithdrawNotEnoughMoney(t *testing.T) {
	defer setupTest(t)()

	testServer := NewTestServer(t)

	client := RegisterTestUser(t, testServer, testServer.URL, "test")

	body := `
	{
		"order": "67",
		"sum": 100
	}
	`

	req, err := http.NewRequest(http.MethodPost, testServer.URL+"/api/user/balance/withdraw", strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err)

	require.Equal(t, http.StatusPaymentRequired, resp.StatusCode)
}

func TestWithdrawGood(t *testing.T) {
	defer setupTest(t)()

	testServer := NewTestServer(t)

	client := RegisterTestUser(t, testServer, testServer.URL, "test")

	req, err := http.NewRequest(http.MethodPost, testServer.URL+"/api/user/orders", strings.NewReader("67"))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "text/plain")

	resp, err := client.Do(req)
	require.NoError(t, err)
	resp.Body.Close()

	require.Equal(t, http.StatusAccepted, resp.StatusCode)

	time.Sleep(2 * time.Second)

	body := `
	{
		"order": "67",
		"sum": 100
	}
	`

	req, err = http.NewRequest(http.MethodPost, testServer.URL+"/api/user/balance/withdraw", strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err = client.Do(req)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetWithdrawalsNoContent(t *testing.T) {
	defer setupTest(t)()

	testServer := NewTestServer(t)

	client := RegisterTestUser(t, testServer, testServer.URL, "test")

	req, err := http.NewRequest(http.MethodGet, testServer.URL+"/api/user/withdrawals", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	resp.Body.Close()

	require.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestGetWithdrawalsGood(t *testing.T) {
	defer setupTest(t)()

	testServer := NewTestServer(t)

	client := RegisterTestUser(t, testServer, testServer.URL, "test")

	body := `
	{
		"order": "67",
		"sum": 0
	}
	`

	req, err := http.NewRequest(http.MethodPost, testServer.URL+"/api/user/balance/withdraw", strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, resp.StatusCode)

	body = `
	{
		"order": "18",
		"sum": 0
	}
	`

	req, err = http.NewRequest(http.MethodPost, testServer.URL+"/api/user/balance/withdraw", strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err = client.Do(req)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, resp.StatusCode)

	req, err = http.NewRequest(http.MethodGet, testServer.URL+"/api/user/withdrawals", nil)
	require.NoError(t, err)

	resp, err = client.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	d := make([]domain.WithdrawData, 0)

	err = json.Unmarshal(data, &d)
	require.NoError(t, err)

	resp.Body.Close()

}
