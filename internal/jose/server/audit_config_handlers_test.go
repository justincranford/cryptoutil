// Copyright (c) 2025 Justin Cranford
//

package server

import (
	"bytes"
	"context"
	"crypto/tls"
	json "encoding/json"
	"io"
	http "net/http"
	"sync"
	"testing"
	"time"

	cryptoutilJoseConfig "cryptoutil/internal/jose/config"
	cryptoutilJoseDomain "cryptoutil/internal/jose/domain"

	"github.com/stretchr/testify/require"
)

var (
	auditTestServer     *JoseServer
	auditTestHTTPClient *http.Client
	auditSetupOnce      sync.Once
	auditSetupErr       error
)

// setupAuditTestServer initializes the test server once for all audit tests.
func setupAuditTestServer() (*JoseServer, *http.Client, error) {
	auditSetupOnce.Do(func() {
		ctx := context.Background()

		// Create test server using NewTestSettings (bypasses pflag global state).
		cfg := cryptoutilJoseConfig.NewTestSettings()

		auditTestServer, auditSetupErr = NewFromConfig(ctx, cfg)
		if auditSetupErr != nil {
			return
		}

		// Start server in background.
		go func() {
			_ = auditTestServer.Start(ctx)
		}()

		// Wait for server to be ready and have actual port assigned.
		// Use a longer wait time to ensure proper initialization.
		time.Sleep(500 * time.Millisecond)
		auditTestServer.SetReady(true)

		// Create HTTP client.
		auditTestHTTPClient = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true, //nolint:gosec // Test environment only
				},
			},
		}
	})

	return auditTestServer, auditTestHTTPClient, auditSetupErr
}

func TestAuditConfigHandlers_GetAuditConfig(t *testing.T) {
	ctx := context.Background()
	server, httpClient, err := setupAuditTestServer()
	require.NoError(t, err)
	require.NotNil(t, server)

	// Make request to get all audit configs.
	url := server.PublicBaseURL() + "/browser/api/v1/admin/audit-config"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	require.NoError(t, err)

	resp, err := httpClient.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	// Parse response.
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var result AuditConfigListResponse

	err = json.Unmarshal(body, &result)
	require.NoError(t, err)

	// Should have default configs for all operations.
	require.NotEmpty(t, result.Configs)
	require.Len(t, result.Configs, 9) // 9 audit operations.

	// Verify each config has expected defaults.
	for _, cfg := range result.Configs {
		require.NotEmpty(t, cfg.TenantID)
		require.NotEmpty(t, cfg.Operation)
		require.True(t, cfg.Enabled)                      // Default enabled.
		require.InDelta(t, 0.01, cfg.SamplingRate, 0.001) // Default 1% sampling.
	}
}

func TestAuditConfigHandlers_GetAuditConfigByOperation(t *testing.T) {
	ctx := context.Background()
	server, httpClient, err := setupAuditTestServer()
	require.NoError(t, err)
	require.NotNil(t, server)

	// Test getting specific operation config.
	url := server.PublicBaseURL() + "/browser/api/v1/admin/audit-config/encrypt"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	require.NoError(t, err)

	resp, err := httpClient.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var result AuditConfigResponse

	err = json.Unmarshal(body, &result)
	require.NoError(t, err)

	require.Equal(t, "encrypt", result.Operation)
	require.True(t, result.Enabled)
	require.InDelta(t, 0.01, result.SamplingRate, 0.001)
}

func TestAuditConfigHandlers_GetAuditConfigByOperation_InvalidOperation(t *testing.T) {
	ctx := context.Background()
	server, httpClient, err := setupAuditTestServer()
	require.NoError(t, err)
	require.NotNil(t, server)

	// Test getting invalid operation config.
	url := server.PublicBaseURL() + "/browser/api/v1/admin/audit-config/invalid_operation"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	require.NoError(t, err)

	resp, err := httpClient.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestAuditConfigHandlers_SetAuditConfig(t *testing.T) {
	ctx := context.Background()
	server, httpClient, err := setupAuditTestServer()
	require.NoError(t, err)
	require.NotNil(t, server)

	// Set config for sign operation (use a different operation to avoid conflicts).
	setReq := AuditConfigRequest{
		Operation:    "sign",
		Enabled:      false,
		SamplingRate: 0.5,
	}

	reqBody, err := json.Marshal(setReq)
	require.NoError(t, err)

	url := server.PublicBaseURL() + "/browser/api/v1/admin/audit-config"

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(reqBody))
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var result AuditConfigResponse

	err = json.Unmarshal(body, &result)
	require.NoError(t, err)

	require.Equal(t, "sign", result.Operation)
	require.False(t, result.Enabled)
	require.InDelta(t, 0.5, result.SamplingRate, 0.001)

	// Verify the config was persisted by getting it again.
	getURL := server.PublicBaseURL() + "/browser/api/v1/admin/audit-config/sign"

	getReq, err := http.NewRequestWithContext(ctx, http.MethodGet, getURL, nil)
	require.NoError(t, err)

	getResp, err := httpClient.Do(getReq)
	require.NoError(t, err)

	defer func() { _ = getResp.Body.Close() }()

	require.Equal(t, http.StatusOK, getResp.StatusCode)

	getBody, err := io.ReadAll(getResp.Body)
	require.NoError(t, err)

	var getResult AuditConfigResponse

	err = json.Unmarshal(getBody, &getResult)
	require.NoError(t, err)

	require.Equal(t, "sign", getResult.Operation)
	require.False(t, getResult.Enabled)
	require.InDelta(t, 0.5, getResult.SamplingRate, 0.001)
}

func TestAuditConfigHandlers_SetAuditConfig_InvalidSamplingRate(t *testing.T) {
	ctx := context.Background()
	server, httpClient, err := setupAuditTestServer()
	require.NoError(t, err)
	require.NotNil(t, server)

	// Set config with invalid sampling rate (> 1).
	setReq := AuditConfigRequest{
		Operation:    "decrypt",
		Enabled:      true,
		SamplingRate: 1.5, // Invalid - must be between 0 and 1.
	}

	reqBody, err := json.Marshal(setReq)
	require.NoError(t, err)

	url := server.PublicBaseURL() + "/browser/api/v1/admin/audit-config"

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(reqBody))
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestAuditConfigHandlers_SetAuditConfig_MissingOperation(t *testing.T) {
	ctx := context.Background()
	server, httpClient, err := setupAuditTestServer()
	require.NoError(t, err)
	require.NotNil(t, server)

	// Set config without operation.
	setReq := AuditConfigRequest{
		Operation:    "", // Missing operation.
		Enabled:      true,
		SamplingRate: 0.5,
	}

	reqBody, err := json.Marshal(setReq)
	require.NoError(t, err)

	url := server.PublicBaseURL() + "/browser/api/v1/admin/audit-config"

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(reqBody))
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// Ensure domain is imported.
var _ = cryptoutilJoseDomain.AuditConfig{}
