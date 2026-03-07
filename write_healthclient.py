import os

healthclient_go = r"""// Copyright (c) 2025 Justin Cranford
//

// Package healthclient provides an HTTP client for testing service health endpoints.
package healthclient

import (
"io"
"net/http"

cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
cryptoutilSharedCryptoTls "cryptoutil/internal/shared/crypto/tls"
)

// HealthClient is a test-only HTTPS client for hitting service health endpoints.
type HealthClient struct {
publicBaseURL string
adminBaseURL  string
client        *http.Client
}

// NewHealthClient creates a new HealthClient using TLS skip-verify (safe for auto-generated test certs).
func NewHealthClient(publicBaseURL, adminBaseURL string) *HealthClient {
tlsConfig, err := cryptoutilSharedCryptoTls.NewClientConfig(&cryptoutilSharedCryptoTls.ClientConfigOptions{
SkipVerify: true, //nolint:gosec // test-only: auto-generated self-signed test certificates
})

transport := &http.Transport{}

if err == nil {
transport.TLSClientConfig = tlsConfig.TLSConfig
}

return &HealthClient{
publicBaseURL: publicBaseURL,
adminBaseURL:  adminBaseURL,
client: &http.Client{
Timeout:   cryptoutilSharedMagic.IMDefaultTimeout,
Transport: transport,
},
}
}

// Livez calls the admin livez endpoint.
func (h *HealthClient) Livez() (*http.Response, error) {
url := h.adminBaseURL + cryptoutilSharedMagic.DefaultPrivateAdminAPIContextPath + cryptoutilSharedMagic.PrivateAdminLivezRequestPath
resp, err := h.client.Get(url) //nolint:noctx // test helper: no context needed for health polling
if err != nil {
return nil, err
}

return resp, nil
}

// Readyz calls the admin readyz endpoint.
func (h *HealthClient) Readyz() (*http.Response, error) {
url := h.adminBaseURL + cryptoutilSharedMagic.DefaultPrivateAdminAPIContextPath + cryptoutilSharedMagic.PrivateAdminReadyzRequestPath
resp, err := h.client.Get(url) //nolint:noctx // test helper: no context needed for health polling
if err != nil {
return nil, err
}

return resp, nil
}

// ServiceHealth calls the public service-path health endpoint.
func (h *HealthClient) ServiceHealth() (*http.Response, error) {
url := h.publicBaseURL + cryptoutilSharedMagic.DefaultPublicServiceAPIContextPath + "/health"
resp, err := h.client.Get(url) //nolint:noctx // test helper: no context needed for health polling
if err != nil {
return nil, err
}

return resp, nil
}

// BrowserHealth calls the public browser-path health endpoint.
func (h *HealthClient) BrowserHealth() (*http.Response, error) {
url := h.publicBaseURL + cryptoutilSharedMagic.DefaultPublicBrowserAPIContextPath + "/health"
resp, err := h.client.Get(url) //nolint:noctx // test helper: no context needed for health polling
if err != nil {
return nil, err
}

return resp, nil
}

// PublicHealth calls the public browser-path health endpoint (alias for BrowserHealth).
func (h *HealthClient) PublicHealth() (*http.Response, error) {
return h.BrowserHealth()
}

// drainAndClose reads and discards the response body then closes it.
func drainAndClose(resp *http.Response) {
if resp != nil && resp.Body != nil {
_, _ = io.Copy(io.Discard, resp.Body)
_ = resp.Body.Close()
}
}

// DrainAndClose reads and discards all bytes from resp.Body then closes it.
// Use this when you want the connection returned to the pool but don't need the body.
func DrainAndClose(resp *http.Response) {
drainAndClose(resp)
}
"""

healthclient_test_go = r"""// Copyright (c) 2025 Justin Cranford
//

package healthclient_test

import (
"net/http"
"net/http/httptest"
"testing"

"github.com/stretchr/testify/assert"
"github.com/stretchr/testify/require"

cryptoutilTestingHealthclient "cryptoutil/internal/apps/template/service/testing/healthclient"
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

assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHealthClient_Readyz_Success(t *testing.T) {
t.Parallel()

srv := startTestServer(t, http.HandlerFunc(alwaysOK))
hc := cryptoutilTestingHealthclient.NewHealthClient(srv.URL, srv.URL)

resp, err := hc.Readyz()
require.NoError(t, err)
require.NotNil(t, resp)
t.Cleanup(func() { _ = resp.Body.Close() })

assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHealthClient_ServiceHealth_Success(t *testing.T) {
t.Parallel()

srv := startTestServer(t, http.HandlerFunc(alwaysOK))
hc := cryptoutilTestingHealthclient.NewHealthClient(srv.URL, srv.URL)

resp, err := hc.ServiceHealth()
require.NoError(t, err)
require.NotNil(t, resp)
t.Cleanup(func() { _ = resp.Body.Close() })

assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHealthClient_BrowserHealth_Success(t *testing.T) {
t.Parallel()

srv := startTestServer(t, http.HandlerFunc(alwaysOK))
hc := cryptoutilTestingHealthclient.NewHealthClient(srv.URL, srv.URL)

resp, err := hc.BrowserHealth()
require.NoError(t, err)
require.NotNil(t, resp)
t.Cleanup(func() { _ = resp.Body.Close() })

assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHealthClient_PublicHealth_Success(t *testing.T) {
t.Parallel()

srv := startTestServer(t, http.HandlerFunc(alwaysOK))
hc := cryptoutilTestingHealthclient.NewHealthClient(srv.URL, srv.URL)

resp, err := hc.PublicHealth()
require.NoError(t, err)
require.NotNil(t, resp)
t.Cleanup(func() { _ = resp.Body.Close() })

assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHealthClient_Livez_ServerError(t *testing.T) {
t.Parallel()

srv := startTestServer(t, http.HandlerFunc(alwaysErr))
hc := cryptoutilTestingHealthclient.NewHealthClient(srv.URL, srv.URL)

resp, err := hc.Livez()
require.NoError(t, err)
require.NotNil(t, resp)
t.Cleanup(func() { _ = resp.Body.Close() })

assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
}

func TestHealthClient_Readyz_ServerError(t *testing.T) {
t.Parallel()

srv := startTestServer(t, http.HandlerFunc(alwaysErr))
hc := cryptoutilTestingHealthclient.NewHealthClient(srv.URL, srv.URL)

resp, err := hc.Readyz()
require.NoError(t, err)
require.NotNil(t, resp)
t.Cleanup(func() { _ = resp.Body.Close() })

assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
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

// DrainAndClose must not panic and must consume the body.
cryptoutilTestingHealthclient.DrainAndClose(resp)
}

func TestHealthClient_ConnectionError(t *testing.T) {
t.Parallel()

hc := cryptoutilTestingHealthclient.NewHealthClient("https://127.0.0.1:1", "https://127.0.0.1:1")

resp, err := hc.Livez()
assert.Error(t, err)
assert.Nil(t, resp)
}
"""

base = "internal/apps/template/service/testing/healthclient"
os.makedirs(base, exist_ok=True)

with open(f"{base}/healthclient.go", "w", encoding="utf-8", newline="\n") as f:
    f.write(healthclient_go.lstrip("\n"))

with open(f"{base}/healthclient_test.go", "w", encoding="utf-8", newline="\n") as f:
    f.write(healthclient_test_go.lstrip("\n"))

print("Written healthclient.go and healthclient_test.go")
