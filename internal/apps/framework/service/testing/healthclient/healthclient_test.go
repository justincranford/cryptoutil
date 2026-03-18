// Copyright (c) 2025 Justin Cranford
//

package healthclient_test

import (
	http "net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilTestingHealthclient "cryptoutil/internal/apps/framework/service/testing/healthclient"
)

// startTestServer starts a plain HTTP test server that handles GET requests with a fixed 200 response.
func startTestServer(t *testing.T, handler http.Handler) *httptest.Server {
	t.Helper()

	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	return srv
}

// alwaysOK is an http.HandlerFunc that always returns 200 OK with an empty body.
func alwaysOK(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// alwaysErr is an http.HandlerFunc that always returns 503 Service Unavailable.
func alwaysErr(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusServiceUnavailable)
}

func TestNewHealthClient_Constructor(t *testing.T) {
	t.Parallel()

	srv := startTestServer(t, http.HandlerFunc(alwaysOK))
	hc := cryptoutilTestingHealthclient.NewHealthClient(srv.URL, srv.URL)

	require.NotNil(t, hc)
}

func TestHealthClient_Livez_Success(t *testing.T) {
	t.Parallel()

	srv := startTestServer(t, http.HandlerFunc(alwaysOK))
	hc := cryptoutilTestingHealthclient.NewHealthClient(srv.URL, srv.URL)

	resp, err := hc.Livez()
	require.NoError(t, err)
	require.NotNil(t, resp)
	t.Cleanup(func() { _ = resp.Body.Close() })

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHealthClient_Readyz_Success(t *testing.T) {
	t.Parallel()

	srv := startTestServer(t, http.HandlerFunc(alwaysOK))
	hc := cryptoutilTestingHealthclient.NewHealthClient(srv.URL, srv.URL)

	resp, err := hc.Readyz()
	require.NoError(t, err)
	require.NotNil(t, resp)
	t.Cleanup(func() { _ = resp.Body.Close() })

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHealthClient_ServiceHealth_Success(t *testing.T) {
	t.Parallel()

	srv := startTestServer(t, http.HandlerFunc(alwaysOK))
	hc := cryptoutilTestingHealthclient.NewHealthClient(srv.URL, srv.URL)

	resp, err := hc.ServiceHealth()
	require.NoError(t, err)
	require.NotNil(t, resp)
	t.Cleanup(func() { _ = resp.Body.Close() })

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHealthClient_BrowserHealth_Success(t *testing.T) {
	t.Parallel()

	srv := startTestServer(t, http.HandlerFunc(alwaysOK))
	hc := cryptoutilTestingHealthclient.NewHealthClient(srv.URL, srv.URL)

	resp, err := hc.BrowserHealth()
	require.NoError(t, err)
	require.NotNil(t, resp)
	t.Cleanup(func() { _ = resp.Body.Close() })

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHealthClient_PublicHealth_Success(t *testing.T) {
	t.Parallel()

	srv := startTestServer(t, http.HandlerFunc(alwaysOK))
	hc := cryptoutilTestingHealthclient.NewHealthClient(srv.URL, srv.URL)

	resp, err := hc.PublicHealth()
	require.NoError(t, err)
	require.NotNil(t, resp)
	t.Cleanup(func() { _ = resp.Body.Close() })

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHealthClient_Livez_ServerError(t *testing.T) {
	t.Parallel()

	srv := startTestServer(t, http.HandlerFunc(alwaysErr))
	hc := cryptoutilTestingHealthclient.NewHealthClient(srv.URL, srv.URL)

	resp, err := hc.Livez()
	require.NoError(t, err)
	require.NotNil(t, resp)
	t.Cleanup(func() { _ = resp.Body.Close() })

	require.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
}

func TestHealthClient_Readyz_ServerError(t *testing.T) {
	t.Parallel()

	srv := startTestServer(t, http.HandlerFunc(alwaysErr))
	hc := cryptoutilTestingHealthclient.NewHealthClient(srv.URL, srv.URL)

	resp, err := hc.Readyz()
	require.NoError(t, err)
	require.NotNil(t, resp)
	t.Cleanup(func() { _ = resp.Body.Close() })

	require.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
}

func TestHealthClient_DrainAndClose_NilResp(t *testing.T) {
	t.Parallel()

	// Should not panic on nil response.
	cryptoutilTestingHealthclient.DrainAndClose(nil)
}

func TestHealthClient_DrainAndClose_WithBody(t *testing.T) {
	t.Parallel()

	srv := startTestServer(t, http.HandlerFunc(alwaysOK))
	hc := cryptoutilTestingHealthclient.NewHealthClient(srv.URL, srv.URL)

	resp, err := hc.Livez()
	require.NoError(t, err)
	t.Cleanup(func() { _ = resp.Body.Close() })

	// DrainAndClose must not panic and must consume the body.
	cryptoutilTestingHealthclient.DrainAndClose(resp)
}

func TestHealthClient_ConnectionError(t *testing.T) {
	t.Parallel()

	hc := cryptoutilTestingHealthclient.NewHealthClient("https://127.0.0.1:1", "https://127.0.0.1:1")

	resp, err := hc.Livez()
	if resp != nil {
		t.Cleanup(func() { _ = resp.Body.Close() })
	}

	require.Error(t, err)
	require.Nil(t, resp)
}

func TestHealthClient_Readyz_ConnectionError(t *testing.T) {
	t.Parallel()

	hc := cryptoutilTestingHealthclient.NewHealthClient("https://127.0.0.1:1", "https://127.0.0.1:1")

	resp, err := hc.Readyz()
	if resp != nil {
		t.Cleanup(func() { _ = resp.Body.Close() })
	}

	require.Error(t, err)
	require.Nil(t, resp)
}

func TestHealthClient_ServiceHealth_ConnectionError(t *testing.T) {
	t.Parallel()

	hc := cryptoutilTestingHealthclient.NewHealthClient("https://127.0.0.1:1", "https://127.0.0.1:1")

	resp, err := hc.ServiceHealth()
	if resp != nil {
		t.Cleanup(func() { _ = resp.Body.Close() })
	}

	require.Error(t, err)
	require.Nil(t, resp)
}

func TestHealthClient_BrowserHealth_ConnectionError(t *testing.T) {
	t.Parallel()

	hc := cryptoutilTestingHealthclient.NewHealthClient("https://127.0.0.1:1", "https://127.0.0.1:1")

	resp, err := hc.BrowserHealth()
	if resp != nil {
		t.Cleanup(func() { _ = resp.Body.Close() })
	}

	require.Error(t, err)
	require.Nil(t, resp)
}
