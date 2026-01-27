// Copyright (c) 2025 Justin Cranford

package listener

import (
	"context"
	"crypto/tls"
	json "encoding/json"
	"fmt"
	"io"
	http "net/http"
	"sync"
	"testing"
	"time"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilAppsTemplateServiceConfigTlsGenerator "cryptoutil/internal/apps/template/service/config/tls_generator"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testUnknownTLSMode = "unknown"

func TestNewHTTPServers_AutoMode_HappyPath(t *testing.T) {
	settings := cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)

	ctx := context.Background()
	h, err := NewHTTPServers(ctx, settings)
	require.NoError(t, err)
	require.NotNil(t, h)
	require.NotNil(t, h.PublicServer)
	require.NotNil(t, h.AdminServer)
}

func TestNewHTTPServers_NilContext(t *testing.T) {
	settings := cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)

	h, err := NewHTTPServers(nil, settings) //nolint:staticcheck // Testing nil context handling.
	require.Error(t, err)
	assert.Contains(t, err.Error(), "context cannot be nil")
	assert.Nil(t, h)
}

func TestNewHTTPServers_NilSettings(t *testing.T) {
	ctx := context.Background()

	h, err := NewHTTPServers(ctx, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "settings cannot be nil")
	assert.Nil(t, h)
}

func TestNewHTTPServers_StaticMode_HappyPath(t *testing.T) {
	// Generate static certs first using auto mode.
	staticTLS, err := cryptoutilAppsTemplateServiceConfigTlsGenerator.GenerateAutoTLSGeneratedSettings(
		[]string{"localhost"},
		[]string{cryptoutilSharedMagic.IPv4Loopback},
		cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year,
	)
	require.NoError(t, err)

	// Create settings with static TLS mode.
	settings := cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)
	settings.TLSPublicMode = cryptoutilAppsTemplateServiceConfig.TLSModeStatic
	settings.TLSPrivateMode = cryptoutilAppsTemplateServiceConfig.TLSModeStatic
	settings.TLSStaticCertPEM = staticTLS.StaticCertPEM
	settings.TLSStaticKeyPEM = staticTLS.StaticKeyPEM

	ctx := context.Background()
	h, err := NewHTTPServers(ctx, settings)
	require.NoError(t, err)
	require.NotNil(t, h)
	require.NotNil(t, h.PublicServer)
	require.NotNil(t, h.AdminServer)
}

func TestNewHTTPServers_UnknownPublicTLSMode(t *testing.T) {
	settings := cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)
	settings.TLSPublicMode = testUnknownTLSMode

	ctx := context.Background()
	h, err := NewHTTPServers(ctx, settings)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown public TLS mode")
	assert.Nil(t, h)
}

func TestNewHTTPServers_UnknownPrivateTLSMode(t *testing.T) {
	settings := cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)
	settings.TLSPrivateMode = testUnknownTLSMode

	ctx := context.Background()
	h, err := NewHTTPServers(ctx, settings)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown admin TLS mode")
	assert.Nil(t, h)
}

func TestNewHTTPServers_MixedMode_HappyPath(t *testing.T) {
	// First, generate a CA to use for mixed mode.
	caCertPEM, caKeyPEM, err := cryptoutilAppsTemplateServiceConfigTlsGenerator.GenerateTestCA()
	require.NoError(t, err)

	// Create settings with mixed TLS mode using the generated CA.
	settings := cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)
	settings.TLSPublicMode = cryptoutilAppsTemplateServiceConfig.TLSModeMixed
	settings.TLSPrivateMode = cryptoutilAppsTemplateServiceConfig.TLSModeMixed
	settings.TLSMixedCACertPEM = caCertPEM
	settings.TLSMixedCAKeyPEM = caKeyPEM
	settings.TLSPublicDNSNames = []string{"localhost"}
	settings.TLSPublicIPAddresses = []string{cryptoutilSharedMagic.IPv4Loopback}
	settings.TLSPrivateDNSNames = []string{"localhost"}
	settings.TLSPrivateIPAddresses = []string{cryptoutilSharedMagic.IPv4Loopback}

	ctx := context.Background()
	h, err := NewHTTPServers(ctx, settings)
	require.NoError(t, err)
	require.NotNil(t, h)
	require.NotNil(t, h.PublicServer)
	require.NotNil(t, h.AdminServer)
}

func TestNewHTTPServers_MixedMode_InvalidPublicCA(t *testing.T) {
	// Create settings with mixed TLS mode but invalid CA cert.
	settings := cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)
	settings.TLSPublicMode = cryptoutilAppsTemplateServiceConfig.TLSModeMixed
	settings.TLSMixedCACertPEM = []byte("invalid-ca-cert")
	settings.TLSMixedCAKeyPEM = []byte("invalid-ca-key")
	settings.TLSPublicDNSNames = []string{"localhost"}
	settings.TLSPublicIPAddresses = []string{cryptoutilSharedMagic.IPv4Loopback}

	ctx := context.Background()
	h, err := NewHTTPServers(ctx, settings)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to generate public server cert from CA")
	assert.Nil(t, h)
}

func TestNewHTTPServers_MixedMode_InvalidPrivateCA(t *testing.T) {
	// Create settings with auto for public, mixed for private with invalid CA.
	settings := cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)
	settings.TLSPublicMode = cryptoutilAppsTemplateServiceConfig.TLSModeAuto
	settings.TLSPrivateMode = cryptoutilAppsTemplateServiceConfig.TLSModeMixed
	settings.TLSMixedCACertPEM = []byte("invalid-ca-cert")
	settings.TLSMixedCAKeyPEM = []byte("invalid-ca-key")
	settings.TLSPrivateDNSNames = []string{"localhost"}
	settings.TLSPrivateIPAddresses = []string{cryptoutilSharedMagic.IPv4Loopback}

	ctx := context.Background()
	h, err := NewHTTPServers(ctx, settings)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to generate admin server cert from CA")
	assert.Nil(t, h)
}

// ========================
// Dual HTTPS Server Integration Tests
// ========================

// TestDualServers_StartBothServers tests that both public and admin servers can start concurrently.
func TestDualServers_StartBothServers(t *testing.T) {
	// NOT parallel - tests port binding.
	ctx := context.Background()
	settings := cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)

	h, err := NewHTTPServers(ctx, settings)
	require.NoError(t, err)
	require.NotNil(t, h)

	// Start both servers in background.
	var wg sync.WaitGroup

	serverCtx, cancelServers := context.WithCancel(ctx)
	defer cancelServers()

	// Start public server.
	wg.Add(1)

	publicErrCh := make(chan error, 1)

	go func() {
		defer wg.Done()

		if err := h.PublicServer.Start(serverCtx); err != nil {
			publicErrCh <- err
		}
	}()

	// Start admin server.
	wg.Add(1)

	adminErrCh := make(chan error, 1)

	go func() {
		defer wg.Done()

		if err := h.AdminServer.Start(serverCtx); err != nil {
			adminErrCh <- err
		}
	}()

	// Wait for servers to start.
	time.Sleep(200 * time.Millisecond)

	// Verify both servers allocated ports.
	publicPort := h.PublicServer.ActualPort()
	adminPort := h.AdminServer.ActualPort()

	require.Greater(t, publicPort, 0, "Public server should have allocated a port")
	require.Greater(t, adminPort, 0, "Admin server should have allocated a port")
	require.NotEqual(t, publicPort, adminPort, "Public and admin servers should use different ports")

	// Shutdown both servers.
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	_ = h.PublicServer.Shutdown(shutdownCtx)
	_ = h.AdminServer.Shutdown(shutdownCtx)

	cancelServers()
	wg.Wait()

	// Check for startup errors.
	select {
	case err := <-publicErrCh:
		require.Fail(t, "Public server error", err.Error())
	default:
	}

	select {
	case err := <-adminErrCh:
		require.Fail(t, "Admin server error", err.Error())
	default:
	}
}

// TestDualServers_HealthEndpoints tests health check endpoints on both servers.
func TestDualServers_HealthEndpoints(t *testing.T) {
	// NOT parallel - tests port binding.
	ctx := context.Background()
	settings := cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)

	h, err := NewHTTPServers(ctx, settings)
	require.NoError(t, err)

	// Start both servers.
	var wg sync.WaitGroup

	serverCtx, cancelServers := context.WithCancel(ctx)
	defer cancelServers()

	wg.Add(2)

	go func() {
		defer wg.Done()

		_ = h.PublicServer.Start(serverCtx)
	}()

	go func() {
		defer wg.Done()

		_ = h.AdminServer.Start(serverCtx)
	}()

	// Wait for servers to start.
	time.Sleep(200 * time.Millisecond)

	publicPort := h.PublicServer.ActualPort()
	adminPort := h.AdminServer.ActualPort()

	require.Greater(t, publicPort, 0)
	require.Greater(t, adminPort, 0)

	// Create HTTPS client.
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, //nolint:gosec // Test environment with self-signed certs.
			},
		},
		Timeout: 5 * time.Second,
	}

	// Test public server health endpoints.
	publicServiceHealthURL := fmt.Sprintf("https://%s:%d/service/api/v1/health", cryptoutilSharedMagic.IPv4Loopback, publicPort)
	publicBrowserHealthURL := fmt.Sprintf("https://%s:%d/browser/api/v1/health", cryptoutilSharedMagic.IPv4Loopback, publicPort)

	reqCtx, reqCancel := context.WithTimeout(ctx, 5*time.Second)
	defer reqCancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, publicServiceHealthURL, nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	_ = resp.Body.Close()

	req, err = http.NewRequestWithContext(reqCtx, http.MethodGet, publicBrowserHealthURL, nil)
	require.NoError(t, err)

	resp, err = client.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	_ = resp.Body.Close()

	// Test admin server health endpoints.
	adminLivezURL := fmt.Sprintf("https://%s:%d/admin/api/v1/livez", cryptoutilSharedMagic.IPv4Loopback, adminPort)
	adminReadyzURL := fmt.Sprintf("https://%s:%d/admin/api/v1/readyz", cryptoutilSharedMagic.IPv4Loopback, adminPort)

	req, err = http.NewRequestWithContext(reqCtx, http.MethodGet, adminLivezURL, nil)
	require.NoError(t, err)

	resp, err = client.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	bodyBytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	_ = resp.Body.Close()

	var livezResponse map[string]any

	err = json.Unmarshal(bodyBytes, &livezResponse)
	require.NoError(t, err)
	require.Equal(t, "alive", livezResponse["status"])

	// Readyz should return 503 initially (not ready).
	req, err = http.NewRequestWithContext(reqCtx, http.MethodGet, adminReadyzURL, nil)
	require.NoError(t, err)

	resp, err = client.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
	_ = resp.Body.Close()

	// Set admin server to ready.
	h.AdminServer.SetReady(true)

	// Now readyz should return 200.
	req, err = http.NewRequestWithContext(reqCtx, http.MethodGet, adminReadyzURL, nil)
	require.NoError(t, err)

	resp, err = client.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	bodyBytes, err = io.ReadAll(resp.Body)
	require.NoError(t, err)

	_ = resp.Body.Close()

	var readyzResponse map[string]any

	err = json.Unmarshal(bodyBytes, &readyzResponse)
	require.NoError(t, err)
	require.Equal(t, "ready", readyzResponse["status"])

	// Shutdown.
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	_ = h.PublicServer.Shutdown(shutdownCtx)
	_ = h.AdminServer.Shutdown(shutdownCtx)

	cancelServers()
	wg.Wait()
}

// TestDualServers_GracefulShutdown tests that both servers shut down gracefully.
func TestDualServers_GracefulShutdown(t *testing.T) {
	// NOT parallel - tests port binding.
	ctx := context.Background()
	settings := cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)

	h, err := NewHTTPServers(ctx, settings)
	require.NoError(t, err)

	// Start both servers.
	var wg sync.WaitGroup

	serverCtx, cancelServers := context.WithCancel(ctx)

	wg.Add(2)

	go func() {
		defer wg.Done()

		_ = h.PublicServer.Start(serverCtx)
	}()

	go func() {
		defer wg.Done()

		_ = h.AdminServer.Start(serverCtx)
	}()

	// Wait for servers to start.
	time.Sleep(200 * time.Millisecond)

	publicPort := h.PublicServer.ActualPort()
	adminPort := h.AdminServer.ActualPort()

	require.Greater(t, publicPort, 0)
	require.Greater(t, adminPort, 0)

	// Create HTTPS client.
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, //nolint:gosec // Test environment with self-signed certs.
			},
		},
		Timeout: 5 * time.Second,
	}

	// Verify both servers are responding.
	reqCtx, reqCancel := context.WithTimeout(ctx, 5*time.Second)

	adminLivezURL := fmt.Sprintf("https://%s:%d/admin/api/v1/livez", cryptoutilSharedMagic.IPv4Loopback, adminPort)
	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, adminLivezURL, nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	_ = resp.Body.Close()

	reqCancel()

	// Initiate graceful shutdown.
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	publicShutdownErr := h.PublicServer.Shutdown(shutdownCtx)
	adminShutdownErr := h.AdminServer.Shutdown(shutdownCtx)

	require.NoError(t, publicShutdownErr, "Public server should shut down without error")
	require.NoError(t, adminShutdownErr, "Admin server should shut down without error")

	cancelServers()
	wg.Wait()

	// Verify servers are no longer responding.
	reqCtx2, reqCancel2 := context.WithTimeout(ctx, 2*time.Second)
	defer reqCancel2()

	req, err = http.NewRequestWithContext(reqCtx2, http.MethodGet, adminLivezURL, nil)
	require.NoError(t, err)

	resp, err = client.Do(req)
	if resp != nil {
		_ = resp.Body.Close()
	}

	require.Error(t, err, "Admin server should no longer respond after shutdown")
}

// TestDualServers_BothServersAccessibleSimultaneously tests that both servers can handle requests at the same time.
func TestDualServers_BothServersAccessibleSimultaneously(t *testing.T) {
	// NOT parallel - tests port binding.
	ctx := context.Background()
	settings := cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)

	h, err := NewHTTPServers(ctx, settings)
	require.NoError(t, err)

	// Start both servers.
	var wg sync.WaitGroup

	serverCtx, cancelServers := context.WithCancel(ctx)
	defer cancelServers()

	wg.Add(2)

	go func() {
		defer wg.Done()

		_ = h.PublicServer.Start(serverCtx)
	}()

	go func() {
		defer wg.Done()

		_ = h.AdminServer.Start(serverCtx)
	}()

	// Wait for servers to start.
	time.Sleep(200 * time.Millisecond)

	publicPort := h.PublicServer.ActualPort()
	adminPort := h.AdminServer.ActualPort()

	require.Greater(t, publicPort, 0)
	require.Greater(t, adminPort, 0)

	// Create HTTPS client.
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, //nolint:gosec // Test environment with self-signed certs.
			},
		},
		Timeout: 5 * time.Second,
	}

	// Make concurrent requests to both servers.
	var requestWg sync.WaitGroup

	var publicErr, adminErr error

	requestWg.Add(2)

	go func() {
		defer requestWg.Done()

		reqCtx, reqCancel := context.WithTimeout(ctx, 5*time.Second)
		defer reqCancel()

		url := fmt.Sprintf("https://%s:%d/service/api/v1/health", cryptoutilSharedMagic.IPv4Loopback, publicPort)

		req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, url, nil)
		if err != nil {
			publicErr = err

			return
		}

		resp, err := client.Do(req)
		if err != nil {
			publicErr = err

			return
		}

		if resp.StatusCode != http.StatusOK {
			publicErr = fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}

		_ = resp.Body.Close()
	}()

	go func() {
		defer requestWg.Done()

		reqCtx, reqCancel := context.WithTimeout(ctx, 5*time.Second)
		defer reqCancel()

		url := fmt.Sprintf("https://%s:%d/admin/api/v1/livez", cryptoutilSharedMagic.IPv4Loopback, adminPort)

		req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, url, nil)
		if err != nil {
			adminErr = err

			return
		}

		resp, err := client.Do(req)
		if err != nil {
			adminErr = err

			return
		}

		if resp.StatusCode != http.StatusOK {
			adminErr = fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}

		_ = resp.Body.Close()
	}()

	requestWg.Wait()

	require.NoError(t, publicErr, "Public server request should succeed")
	require.NoError(t, adminErr, "Admin server request should succeed")

	// Shutdown.
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	_ = h.PublicServer.Shutdown(shutdownCtx)
	_ = h.AdminServer.Shutdown(shutdownCtx)

	cancelServers()
	wg.Wait()
}
