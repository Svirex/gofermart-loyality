//go:build integration || integration_api
// +build integration integration_api

package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Svirex/gofermart-loyality/internal/core/domain"
	"github.com/stretchr/testify/require"
)

func TestCreateOrderNotAuth(t *testing.T) {
	defer setupTest(t)()

	testServer := NewTestServer(t)

	req, err := http.NewRequest(http.MethodPost, testServer.URL+"/api/user/orders", nil)
	require.NoError(t, err)

	resp, err := testServer.Client().Do(req)
	require.NoError(t, err)
	resp.Body.Close()

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)

}

func RegisterTestUser(t *testing.T, server *httptest.Server, addr string, login string) *http.Client {
	body := fmt.Sprintf(`
	{
		"login": "%s",
		"password": "testpassword"
	}
	`, login)
	req, err := http.NewRequest(http.MethodPost, addr+"/api/user/register", strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	jar, err := cookiejar.New(nil)
	require.NoError(t, err)

	client := server.Client()
	client.Jar = jar

	resp, err := client.Do(req)
	require.NoError(t, err)
	resp.Body.Close()

	return client
}

func TestCreateOrderContentType(t *testing.T) {
	defer setupTest(t)()

	testServer := NewTestServer(t)

	client := RegisterTestUser(t, testServer, testServer.URL, "test")

	req, err := http.NewRequest(http.MethodPost, testServer.URL+"/api/user/orders", nil)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err)
	resp.Body.Close()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestCreateOrderNilBody(t *testing.T) {
	defer setupTest(t)()

	testServer := NewTestServer(t)

	client := RegisterTestUser(t, testServer, testServer.URL, "test")

	req, err := http.NewRequest(http.MethodPost, testServer.URL+"/api/user/orders", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	resp.Body.Close()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestCreateOrderEmptyBody(t *testing.T) {
	defer setupTest(t)()

	testServer := NewTestServer(t)

	client := RegisterTestUser(t, testServer, testServer.URL, "test")

	body := ""

	req, err := http.NewRequest(http.MethodPost, testServer.URL+"/api/user/orders", strings.NewReader(body))
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	resp.Body.Close()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestCreateOrderInvalidOrderNum(t *testing.T) {
	defer setupTest(t)()

	testServer := NewTestServer(t)

	client := RegisterTestUser(t, testServer, testServer.URL, "test")

	body := "11"

	req, err := http.NewRequest(http.MethodPost, testServer.URL+"/api/user/orders", strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "text/plain")

	resp, err := client.Do(req)
	require.NoError(t, err)
	resp.Body.Close()

	require.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
}

func TestCreateOrderAlreadyAdded(t *testing.T) {
	defer setupTest(t)()

	testServer := NewTestServer(t)

	client := RegisterTestUser(t, testServer, testServer.URL, "test")

	body := "67"

	req, err := http.NewRequest(http.MethodPost, testServer.URL+"/api/user/orders", strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "text/plain")

	resp, err := client.Do(req)
	require.NoError(t, err)
	resp.Body.Close()

	require.Equal(t, http.StatusAccepted, resp.StatusCode)

	req, err = http.NewRequest(http.MethodPost, testServer.URL+"/api/user/orders", strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "text/plain")

	resp, err = client.Do(req)
	require.NoError(t, err)
	resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestCreateOrderNotOwnOrder(t *testing.T) {
	defer setupTest(t)()

	testServer := NewTestServer(t)

	client := RegisterTestUser(t, testServer, testServer.URL, "test")

	body := "67"

	req, err := http.NewRequest(http.MethodPost, testServer.URL+"/api/user/orders", strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "text/plain")

	resp, err := client.Do(req)
	require.NoError(t, err)
	resp.Body.Close()

	require.Equal(t, http.StatusAccepted, resp.StatusCode)

	client = RegisterTestUser(t, testServer, testServer.URL, "svirex")

	req, err = http.NewRequest(http.MethodPost, testServer.URL+"/api/user/orders", strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "text/plain")

	resp, err = client.Do(req)
	require.NoError(t, err)
	resp.Body.Close()

	require.Equal(t, http.StatusConflict, resp.StatusCode)
}

func TestGetOrdersNoContent(t *testing.T) {
	defer setupTest(t)()

	testServer := NewTestServer(t)

	client := RegisterTestUser(t, testServer, testServer.URL, "test")

	req, err := http.NewRequest(http.MethodGet, testServer.URL+"/api/user/orders", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	resp.Body.Close()

	require.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestGetOrdersGood(t *testing.T) {
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

	req, err = http.NewRequest(http.MethodPost, testServer.URL+"/api/user/orders", strings.NewReader("26"))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "text/plain")

	resp, err = client.Do(req)
	require.NoError(t, err)
	resp.Body.Close()

	require.Equal(t, http.StatusAccepted, resp.StatusCode)

	req, err = http.NewRequest(http.MethodPost, testServer.URL+"/api/user/orders", strings.NewReader("18"))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "text/plain")

	resp, err = client.Do(req)
	require.NoError(t, err)
	resp.Body.Close()

	require.Equal(t, http.StatusAccepted, resp.StatusCode)

	req, err = http.NewRequest(http.MethodGet, testServer.URL+"/api/user/orders", nil)
	require.NoError(t, err)

	resp, err = client.Do(req)
	require.NoError(t, err)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()

	orders := make([]domain.Order, 3)
	err = json.Unmarshal(body, &orders)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, resp.StatusCode)

}
