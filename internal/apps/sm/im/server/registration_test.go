// Copyright (c) 2025 Justin Cranford
//

package server_test

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	http "net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	googleUuid "github.com/google/uuid"

	cryptoutilAppsSmImServer "cryptoutil/internal/apps/sm/im/server"
	cryptoutilAppsSmImServerConfig "cryptoutil/internal/apps/sm/im/server/config"
)

// newTestHTTPClient creates an HTTPS client for testing with self-signed certificates.
func newTestHTTPClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, //nolint:gosec // Test environment only.
			},
		},
		Timeout: cryptoutilSharedMagic.JoseJADefaultMaxMaterials * time.Second,
	}
}

// TestUserRegistration_TriggerUserFactory tests user registration via HTTP to cover
// the userFactory closure in public_server.go.
// Two sequential registrations cover both branches:
//  1. First registration: creates demo tenant (demoTenantID == nil path).
//  2. Second registration: reuses existing demo tenant (demoTenantID != nil path).
func TestUserRegistration_TriggerUserFactory(t *testing.T) {
	t.Parallel()

	client := newTestHTTPClient()
	registerURL := baseURL + "/service/api/v1/users/register"

	tests := []struct {
		name string
	}{
		{name: "first registration creates demo tenant"},
		{name: "second registration reuses demo tenant"},
	}

	// Sequential execution required: second test depends on first setting demoTenantID.
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// NOT parallel — order matters for userFactory state.
			username := fmt.Sprintf("reg-user-%d-%s", i, googleUuid.New().String()[:cryptoutilSharedMagic.IMMinPasswordLength])
			body := fmt.Sprintf(`{"username":"%s","password":"testpassword123"}`, username)

			ctx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.JoseJADefaultMaxMaterials*time.Second)
			defer cancel()

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, registerURL, strings.NewReader(body))
			require.NoError(t, err)

			req.Header.Set("Content-Type", "application/json")

			resp, err := client.Do(req)
			require.NoError(t, err)

			defer func() { _ = resp.Body.Close() }()

			responseBody, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			require.Equal(t, http.StatusCreated, resp.StatusCode, "Response body: %s", string(responseBody))
		})
	}
}

// TestUserRegistration_DBClosedError covers the error path in userFactory's
// RegisterUserWithTenant call by closing the database before registration.
// This triggers public_server.go block L117-124 (error handling in userFactory).
func TestUserRegistration_DBClosedError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cfg := cryptoutilAppsSmImServerConfig.DefaultTestConfig()

	// Create and start a fresh server (separate from TestMain's testSmIMServer).
	server, err := cryptoutilAppsSmImServer.NewFromConfig(ctx, cfg)
	require.NoError(t, err)
	require.NotNil(t, server)

	startCtx, startCancel := context.WithCancel(ctx)
	defer startCancel()

	go func() { _ = server.Start(startCtx) }()

	// Wait for server to bind to ports.
	require.Eventually(t, func() bool {
		return server.PublicPort() > 0
	}, cryptoutilSharedMagic.JoseJADefaultMaxMaterials*time.Second, cryptoutilSharedMagic.JoseJAMaxMaterials*time.Millisecond, "Server should start within 10 seconds")

	// Close DB to force RegisterUserWithTenant failure in userFactory.
	sqlDB, err := server.DB().DB()
	require.NoError(t, err)

	err = sqlDB.Close()
	require.NoError(t, err)

	// Send registration request — userFactory will call RegisterUserWithTenant which fails.
	client := newTestHTTPClient()
	registerURL := server.PublicBaseURL() + "/service/api/v1/users/register"

	username := fmt.Sprintf("db-closed-%s", googleUuid.New().String()[:cryptoutilSharedMagic.IMMinPasswordLength])
	body := fmt.Sprintf(`{"username":"%s","password":"testpassword123"}`, username)

	reqCtx, reqCancel := context.WithTimeout(ctx, cryptoutilSharedMagic.JoseJADefaultMaxMaterials*time.Second)
	defer reqCancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, registerURL, strings.NewReader(body))
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	// Expect error response since DB is closed. The important thing is:
	// the code path through userFactory error handling (block 4) executed.
	require.GreaterOrEqual(t, resp.StatusCode, http.StatusBadRequest, "Expected error status code, got %d", resp.StatusCode)

	// Cleanup: shutdown server.
	startCancel()

	_ = server.Shutdown(context.Background())
}
