//go:build integration || integration_api
// +build integration integration_api

package api

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/Svirex/gofermart-loyality/internal/core/domain"
	"github.com/stretchr/testify/require"
)

func TestGetBalanceNotAuth(t *testing.T) {
	defer setupTest(t)()

	testServer := NewTestServer(t)

	req, err := http.NewRequest(http.MethodGet, testServer.URL+"/api/user/balance", nil)
	require.NoError(t, err)

	resp, err := testServer.Client().Do(req)
	require.NoError(t, err)
	resp.Body.Close()

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestGetBalanceGood(t *testing.T) {
	defer setupTest(t)()

	testServer := NewTestServer(t)

	client := RegisterTestUser(t, testServer, testServer.URL, "test")

	req, err := http.NewRequest(http.MethodGet, testServer.URL+"/api/user/balance", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, resp.StatusCode)

	require.Equal(t, resp.Header.Get("Content-Type"), "application/json")

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	resp.Body.Close()

	balance := &domain.Balance{}

	err = json.Unmarshal(body, &balance)
	require.NoError(t, err)
}
