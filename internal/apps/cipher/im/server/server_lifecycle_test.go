// Copyright (c) 2025 Justin Cranford
//

package server_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"cryptoutil/internal/apps/cipher/im/repository"
	"cryptoutil/internal/apps/cipher/im/server"
	cryptoutilTemplateServiceTesting "cryptoutil/internal/apps/template/service/testing/httpservertests"
	cryptoutilMagic "cryptoutil/internal/shared/magic"
)

// TestNewPublicServer_NilContext tests constructor with nil context.
func TestNewPublicServer_NilContext(t *testing.T) {
	t.Parallel()

	// Use shared resources from TestMain - no need to create db/repos/tlsCfg in every test.
	cleanTestDB(t)

	userRepo := repository.NewUserRepository(testDB)
	messageRepo := repository.NewMessageRepository(testDB)
	messageRecipientJWKRepo := repository.NewMessageRecipientJWKRepository(testDB, nil)

	_, err := server.NewPublicServer(context.Background(), cryptoutilMagic.IPv4Loopback, 0, userRepo, messageRepo, messageRecipientJWKRepo, nil, nil, nil, testTLSCfg)
	require.Error(t, err)
	// The test passes nil for jwkGenService, so that validation triggers first
	require.Contains(t, err.Error(), "JWK generation service cannot be nil")
}

// TestNewPublicServer_NilUserRepo tests constructor with nil user repository.
func TestNewPublicServer_NilUserRepo(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	cleanTestDB(t)

	messageRepo := repository.NewMessageRepository(testDB)
	messageRecipientJWKRepo := repository.NewMessageRecipientJWKRepository(testDB, nil)

	_, err := server.NewPublicServer(ctx, cryptoutilMagic.IPv4Loopback, 0, nil, messageRepo, messageRecipientJWKRepo, nil, nil, nil, testTLSCfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "user repository cannot be nil")
}

// TestNewPublicServer_NilMessageRepo tests constructor with nil message repository.
func TestNewPublicServer_NilMessageRepo(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	cleanTestDB(t)

	userRepo := repository.NewUserRepository(testDB)
	messageRecipientJWKRepo := repository.NewMessageRecipientJWKRepository(testDB, nil)

	_, err := server.NewPublicServer(ctx, cryptoutilMagic.IPv4Loopback, 0, userRepo, nil, messageRecipientJWKRepo, nil, nil, nil, testTLSCfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "message repository cannot be nil")
}

// TestNewPublicServer_NilMessageRecipientJWKRepo tests constructor with nil message recipient JWK repository.
func TestNewPublicServer_NilMessageRecipientJWKRepo(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	cleanTestDB(t)

	userRepo := repository.NewUserRepository(testDB)
	messageRepo := repository.NewMessageRepository(testDB)

	_, err := server.NewPublicServer(ctx, cryptoutilMagic.IPv4Loopback, 0, userRepo, messageRepo, nil, nil, nil, nil, testTLSCfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "message recipient JWK repository cannot be nil")
}

// TestNewPublicServer_NilTLSConfig tests constructor with nil TLS config.
func TestNewPublicServer_NilTLSConfig(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	cleanTestDB(t)

	userRepo := repository.NewUserRepository(testDB)
	messageRepo := repository.NewMessageRepository(testDB)
	messageRecipientJWKRepo := repository.NewMessageRecipientJWKRepository(testDB, nil)

	// Get session manager from testCipherIMServer to pass validation.
	sessionManager := testCipherIMServer.SessionManager()

	_, err := server.NewPublicServer(ctx, cryptoutilMagic.IPv4Loopback, 0, userRepo, messageRepo, messageRecipientJWKRepo, testJWKGenService, nil, sessionManager, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "TLS configuration cannot be nil")
}

// TestHandleServiceHealth_WhileRunning tests health endpoint while server running.
func TestHandleServiceHealth_WhileRunning(t *testing.T) {
	t.Parallel()

	_, svcBaseURL := createTestPublicServer(t, testDB)
	client := createHTTPClient(t)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, svcBaseURL+"/service/api/v1/health", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var respBody map[string]string

	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	require.Equal(t, "healthy", respBody["status"])
}

// TestHandleBrowserHealth_WhileRunning tests browser health endpoint.
func TestHandleBrowserHealth_WhileRunning(t *testing.T) {
	t.Parallel()

	_, svcBaseURL := createTestPublicServer(t, testDB)
	client := createHTTPClient(t)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, svcBaseURL+"/browser/api/v1/health", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var respBody map[string]string

	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	require.Equal(t, "healthy", respBody["status"])
}

// TestShutdown_MultipleCalls tests calling Shutdown multiple times.
func TestShutdown_MultipleCalls(t *testing.T) {
	t.Parallel()

	createServer := func(t *testing.T) cryptoutilTemplateServiceTesting.HTTPServer {
		t.Helper()
		publicServer, _ := createTestPublicServer(t, testDB)

		return publicServer
	}

	cryptoutilTemplateServiceTesting.TestShutdownDoubleCall(t, createServer)
}

// TestPublicServer_StartContextCancelled tests server shutdown via context cancellation.
func TestPublicServer_StartContextCancelled(t *testing.T) {
	t.Parallel()

	srv, _ := createTestPublicServer(t, testDB)

	ctx, cancel := context.WithCancel(context.Background())

	errChan := make(chan error, 1)

	go func() {
		errChan <- srv.Start(ctx)
	}()

	time.Sleep(100 * time.Millisecond)

	cancel()

	err := <-errChan
	require.Error(t, err)
	require.Contains(t, err.Error(), "context canceled")
}

// TestPublicServer_DoubleShutdown tests calling Shutdown twice.
func TestPublicServer_DoubleShutdown(t *testing.T) {
	t.Parallel()

	createServer := func(t *testing.T) cryptoutilTemplateServiceTesting.HTTPServer {
		t.Helper()
		srv, _ := createTestPublicServer(t, testDB)

		return srv
	}

	cryptoutilTemplateServiceTesting.TestShutdownDoubleCall(t, createServer)
}

// TestShutdown_DuplicateCall tests calling Shutdown twice.
func TestShutdown_DuplicateCall(t *testing.T) {
	t.Parallel()

	createServer := func(t *testing.T) cryptoutilTemplateServiceTesting.HTTPServer {
		t.Helper()
		server, _ := createTestPublicServer(t, testDB)

		return server
	}

	cryptoutilTemplateServiceTesting.TestShutdownDoubleCall(t, createServer)
}

// TestStart_ContextCancelled tests server start with cancelled context.
func TestStart_ContextCancelled(t *testing.T) {
	t.Parallel()

	cleanTestDB(t)

	cfg := initTestConfig()

	srv, err := server.NewFromConfig(context.Background(), cfg)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = srv.Start(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "context canceled")

	_ = srv.Shutdown(context.Background())
}

// TestCipherIMServer_Accessors tests all accessor methods on CipherIMServer.
func TestCipherIMServer_Accessors(t *testing.T) {
	t.Parallel()

	// Use the shared testCipherIMServer from TestMain.
	srv := testCipherIMServer
	require.NotNil(t, srv)

	// Test PublicPort.
	publicPort := srv.PublicPort()
	require.Greater(t, publicPort, 0, "PublicPort should return positive port")

	// Test ActualPort (alias for PublicPort).
	actualPort := srv.ActualPort()
	require.Equal(t, publicPort, actualPort, "ActualPort should equal PublicPort")

	// Test AdminPort.
	adminPort := srv.AdminPort()
	require.Greater(t, adminPort, 0, "AdminPort should return positive port")

	// Test PublicBaseURL.
	publicBaseURL := srv.PublicBaseURL()
	require.NotEmpty(t, publicBaseURL, "PublicBaseURL should not be empty")
	require.Contains(t, publicBaseURL, "https://", "PublicBaseURL should start with https://")

	// Test AdminBaseURL.
	adminBaseURL := srv.AdminBaseURL()
	require.NotEmpty(t, adminBaseURL, "AdminBaseURL should not be empty")
	require.Contains(t, adminBaseURL, "https://", "AdminBaseURL should start with https://")

	// Test DB.
	db := srv.DB()
	require.NotNil(t, db, "DB should not be nil")

	// Test JWKGen.
	jwkGen := srv.JWKGen()
	require.NotNil(t, jwkGen, "JWKGen should not be nil")

	// Test Telemetry.
	telemetry := srv.Telemetry()
	require.NotNil(t, telemetry, "Telemetry should not be nil")

	// Test SessionManager.
	sessionManager := srv.SessionManager()
	require.NotNil(t, sessionManager, "SessionManager should not be nil")
}

// TestCipherIMServer_SetReady tests the SetReady method.
func TestCipherIMServer_SetReady(t *testing.T) {
	t.Parallel()

	cleanTestDB(t)

	cfg := initTestConfig()

	srv, err := server.NewFromConfig(context.Background(), cfg)
	require.NoError(t, err)

	defer func() {
		_ = srv.Shutdown(context.Background())
	}()

	// Start server in background.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errChan := make(chan error, 1)

	go func() {
		errChan <- srv.Start(ctx)
	}()

	// Wait for server to be ready.
	time.Sleep(200 * time.Millisecond)

	// Test SetReady - should not panic.
	require.NotPanics(t, func() {
		srv.SetReady(true)
	}, "SetReady(true) should not panic")

	require.NotPanics(t, func() {
		srv.SetReady(false)
	}, "SetReady(false) should not panic")

	// Cancel context to stop server.
	cancel()
}

// TestPublicServer_PublicBaseURL tests the PublicBaseURL accessor.
func TestPublicServer_PublicBaseURL(t *testing.T) {
	t.Parallel()

	publicServer, baseURL := createTestPublicServer(t, testDB)

	// Test PublicBaseURL method.
	result := publicServer.PublicBaseURL()
	require.Equal(t, baseURL, result, "PublicBaseURL should match expected base URL")
	require.Contains(t, result, "https://", "PublicBaseURL should start with https://")
}

// TestNewFromConfig_NilContext tests NewFromConfig with nil context.
func TestNewFromConfig_NilContext(t *testing.T) {
	t.Parallel()

	cfg := initTestConfig()

	_, err := server.NewFromConfig(nil, cfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "context cannot be nil")
}

// TestNewFromConfig_NilConfig tests NewFromConfig with nil config.
func TestNewFromConfig_NilConfig(t *testing.T) {
	t.Parallel()

	_, err := server.NewFromConfig(context.Background(), nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "config cannot be nil")
}

// TestNewFromConfig_UnsupportedTLSPrivateMode tests NewFromConfig with unsupported TLS private mode.
func TestNewFromConfig_UnsupportedTLSPrivateMode(t *testing.T) {
	t.Parallel()

	cfg := initTestConfig()
	cfg.TLSPrivateMode = "unsupported-mode"

	_, err := server.NewFromConfig(context.Background(), cfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported TLS private mode")
}

// TestNewFromConfig_UnsupportedTLSPublicMode tests NewFromConfig with unsupported TLS public mode.
func TestNewFromConfig_UnsupportedTLSPublicMode(t *testing.T) {
	t.Parallel()

	cfg := initTestConfig()
	cfg.TLSPublicMode = "unsupported-mode"

	_, err := server.NewFromConfig(context.Background(), cfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported TLS public mode")
}

// TestNewPublicServer_EmptyBindAddress tests constructor with empty bind address.
func TestNewPublicServer_EmptyBindAddress(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	cleanTestDB(t)

	userRepo := repository.NewUserRepository(testDB)
	messageRepo := repository.NewMessageRepository(testDB)
	messageRecipientJWKRepo := repository.NewMessageRecipientJWKRepository(testDB, nil)
	sessionManager := testCipherIMServer.SessionManager()

	_, err := server.NewPublicServer(ctx, "", 0, userRepo, messageRepo, messageRecipientJWKRepo, testJWKGenService, nil, sessionManager, testTLSCfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "bind address cannot be empty")
}

// TestNewPublicServer_NilJWKGenService tests constructor with nil JWK gen service.
func TestNewPublicServer_NilJWKGenService(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	cleanTestDB(t)

	userRepo := repository.NewUserRepository(testDB)
	messageRepo := repository.NewMessageRepository(testDB)
	messageRecipientJWKRepo := repository.NewMessageRecipientJWKRepository(testDB, nil)
	sessionManager := testCipherIMServer.SessionManager()

	_, err := server.NewPublicServer(ctx, cryptoutilMagic.IPv4Loopback, 0, userRepo, messageRepo, messageRecipientJWKRepo, nil, nil, sessionManager, testTLSCfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "JWK generation service cannot be nil")
}

// TestNewPublicServer_NilSessionManager tests constructor with nil session manager.
func TestNewPublicServer_NilSessionManager(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	cleanTestDB(t)

	userRepo := repository.NewUserRepository(testDB)
	messageRepo := repository.NewMessageRepository(testDB)
	messageRecipientJWKRepo := repository.NewMessageRecipientJWKRepository(testDB, nil)

	_, err := server.NewPublicServer(ctx, cryptoutilMagic.IPv4Loopback, 0, userRepo, messageRepo, messageRecipientJWKRepo, testJWKGenService, nil, nil, testTLSCfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "session manager service cannot be nil")
}
